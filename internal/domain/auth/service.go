package auth

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"flexgin/internal/apperror"
	"flexgin/internal/config"
	"flexgin/internal/store"
)

type Challenge struct {
	UserID    string
	ResendKey uuid.UUID
	ExpiresAt time.Time
}

type Session struct {
	User             store.User
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type AuthService struct {
	q      store.Querier
	tokens *TokenManager
	sender Sender
	otp    config.OTP
}

func NewAuthService(q store.Querier, tokens *TokenManager, sender Sender, otp config.OTP) *AuthService {
	return &AuthService{q: q, tokens: tokens, sender: sender, otp: otp}
}

var phoneRe = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)

func normalizeIdentifier(raw string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(raw))
	if a, err := mail.ParseAddress(v); err == nil && a.Address == v {
		return v, nil
	}
	if phoneRe.MatchString(v) {
		return v, nil
	}
	return "", apperror.Validation("validation failed", map[string]string{
		"identifier": "must be a valid email or E.164 phone number",
	})
}

// RequestOTP signs up the identifier if new, then sends a code.
func (s *AuthService) RequestOTP(ctx context.Context, req RequestOTPRequest) (Challenge, error) {
	identifier, err := normalizeIdentifier(req.Identifier)
	if err != nil {
		return Challenge{}, err
	}
	u, err := s.q.UpsertUser(ctx, identifier)
	if err != nil {
		return Challenge{}, fmt.Errorf("upsert user: %w", err)
	}
	return s.issueOTP(ctx, u)
}

func (s *AuthService) ResendOTP(ctx context.Context, resendKey string) (Challenge, error) {
	_, err := uuid.Parse(resendKey)
	if err != nil {
		return Challenge{}, ErrInvalidResendKey
	}
	ch, err := s.q.GetOTPChallenge(ctx, resendKey)
	if errors.Is(err, sql.ErrNoRows) {
		return Challenge{}, ErrInvalidResendKey
	}
	if err != nil {
		return Challenge{}, fmt.Errorf("get challenge: %w", err)
	}
	u, err := s.q.GetUserByID(ctx, ch.UserID)
	if err != nil {
		return Challenge{}, fmt.Errorf("get user: %w", err)
	}
	return s.issueOTP(ctx, u)
}

func (s *AuthService) issueOTP(ctx context.Context, u store.User) (Challenge, error) {
	last, err := s.q.GetLatestOTPChallenge(ctx, u.ID)
	switch {
	case err == nil:
		if time.Since(time.Unix(last.CreatedAt, 0)) < s.otp.Cooldown {
			return Challenge{}, ErrOTPCooldown
		}
	case !errors.Is(err, sql.ErrNoRows):
		return Challenge{}, fmt.Errorf("latest challenge: %w", err)
	}

	code, err := newOTPCode()
	if err != nil {
		return Challenge{}, fmt.Errorf("generate otp: %w", err)
	}
	id := uuid.New()
	exp := time.Now().Add(s.otp.TTL)

	if _, err := s.q.CreateOTPChallenge(ctx, store.CreateOTPChallengeParams{
		ID:        fmt.Sprintf("%d", id),
		UserID:    u.ID,
		CodeHash:  hashOTP(s.tokens.secret, id.String(), code),
		ExpiresAt: exp,
	}); err != nil {
		return Challenge{}, fmt.Errorf("create challenge: %w", err)
	}
	if err := s.sender.Send(ctx, u.Identifier, code); err != nil {
		return Challenge{}, fmt.Errorf("send otp: %w", err)
	}
	return Challenge{UserID: u.ID, ResendKey: id, ExpiresAt: exp}, nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) (Session, error) {
	ch, err := s.q.GetActiveOTPChallenge(ctx, req.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrInvalidOTP
	}
	if err != nil {
		return Session{}, fmt.Errorf("get challenge: %w", err)
	}
	if int(ch.Attempts) >= s.otp.MaxAttempts {
		return Session{}, ErrInvalidOTP
	}

	want := hashOTP(s.tokens.secret, ch.ID, req.OTP)
	if !hmac.Equal([]byte(ch.CodeHash), []byte(want)) {
		if err := s.q.IncrementOTPAttempts(ctx, ch.ID); err != nil {
			return Session{}, fmt.Errorf("increment attempts: %w", err)
		}
		return Session{}, ErrInvalidOTP
	}

	n, err := s.q.ConsumeOTPChallenge(ctx, ch.ID)
	if err != nil {
		return Session{}, fmt.Errorf("consume challenge: %w", err)
	}
	if n == 0 {
		return Session{}, ErrInvalidOTP
	}

	u, err := s.q.GetUserByID(ctx, req.UserID)
	if err != nil {
		return Session{}, fmt.Errorf("get user: %w", err)
	}
	return s.newSession(ctx, u)
}

// Refresh rotates: the presented refresh token is revoked and a new pair is issued.
func (s *AuthService) Refresh(ctx context.Context, raw string) (Session, error) {
	old, err := s.q.ClaimRefreshSession(ctx, hashRefresh(raw))
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, ErrInvalidToken
	}
	if err != nil {
		return Session{}, fmt.Errorf("claim session: %w", err)
	}
	u, err := s.q.GetUserByID(ctx, old.UserID)
	if err != nil {
		return Session{}, fmt.Errorf("get user: %w", err)
	}
	return s.newSession(ctx, u)
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	_, err := uuid.Parse(sessionID)
	if err != nil {
		return ErrInvalidToken
	}
	if _, err := s.q.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// Authenticate validates the JWT and that its session is still active (immediate logout).
func (s *AuthService) Authenticate(ctx context.Context, raw string) (Claims, error) {
	cl, err := s.tokens.ParseAccess(raw)
	if err != nil {
		return Claims{}, err
	}
	_, err = uuid.Parse(cl.ID)
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	active, err := s.q.SessionIsActive(ctx, cl.ID)
	if err != nil {
		return Claims{}, fmt.Errorf("check session: %w", err)
	}
	if !active {
		return Claims{}, ErrInvalidToken
	}
	return cl, nil
}

func (s *AuthService) Me(ctx context.Context, id string) (store.User, error) {
	u, err := s.q.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return store.User{}, ErrUserNotFound
	}
	if err != nil {
		return store.User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (s *AuthService) newSession(ctx context.Context, u store.User) (Session, error) {
	refresh, hash, err := newRefreshToken()
	if err != nil {
		return Session{}, fmt.Errorf("refresh token: %w", err)
	}
	sid := uuid.New()
	refreshExp := time.Now().Add(s.tokens.refreshTTL)

	if _, err := s.q.CreateSession(ctx, store.CreateSessionParams{
		ID:          sid.String(),
		UserID:      u.ID,
		RefreshHash: hash,
		ExpiresAt:   refreshExp,
	}); err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	access, accessExp, err := s.tokens.IssueAccess(sid.String(), u.ID, u.Role)
	if err != nil {
		return Session{}, fmt.Errorf("issue access token: %w", err)
	}
	return Session{
		User:             u,
		AccessToken:      access,
		RefreshToken:     refresh,
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}
