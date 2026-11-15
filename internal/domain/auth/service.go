package auth

import (
	"context"
	"crypto/hmac"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"flexgin/internal/config"
	"flexgin/internal/store"
	"flexgin/internal/utils"
)

// compared against when the user doesn't exist, so response time doesn't leak which logins exist
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)

type AuthService struct {
	q      store.Querier
	tokens *TokenManager
	sender Sender
	otp    config.OTP
}

func NewAuthService(q store.Querier, tokens *TokenManager, sender Sender, otp config.OTP) *AuthService {
	return &AuthService{q: q, tokens: tokens, sender: sender, otp: otp}
}

// RequestOTP signs up the identifier if new, then sends a code
func (s *AuthService) RequestOTP(ctx context.Context, req RequestOTPRequest) (Challenge, error) {
	identifier, err := utils.NormalizeIdentifier(req.Identifier)
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
	ch, err := s.q.GetOTPChallenge(ctx, resendKey)
	utils.PrintPretty(s.q)
	fmt.Println(s.q)
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

	want := hashOTP(s.tokens.secret, req.OTP, req.OTP)
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

// Refresh rotates the presented refresh token is revoked and a new pair is issued
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
	if _, err := s.q.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// Authenticate validates the JWT and that its session is still active (immediate logout)
func (s *AuthService) Authenticate(ctx context.Context, raw string) (Claims, error) {
	cl, err := s.tokens.ParseAccess(raw)
	if err != nil {
		return Claims{}, err
	}

	if len(cl.ID) != 6 {
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

func (s *AuthService) PasswordAuthSignUp(ctx context.Context, username, identifier, password string) error {
	username = strings.TrimSpace(username)
	identifier = strings.ToLower(strings.TrimSpace(identifier))

	if user, err := s.q.GetUserByIdentifier(ctx, identifier); err == nil {

		if user.Username == &username {
			return ErrUsernameTaken
		}
		return ErrIdentifierTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get user by identifier: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	h := string(hash)

	_, err = s.q.CreateUser(ctx, store.CreateUserParams{
		Username:     &username,
		Identifier:   identifier,
		PasswordHash: &h,
	})

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *AuthService) PasswordAuthSignIn(ctx context.Context, usernameOrIdentifier, password string) (Session, error) {
	usernameOrIdentifier = strings.TrimSpace(usernameOrIdentifier)

	user, err := s.q.GetUserByUsernameOrIdentifier(ctx, &usernameOrIdentifier)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return Session{}, ErrInvalidCredentials
		}
		return Session{}, fmt.Errorf("get user by info: %w", err)
	}

	// OTP-only accounts have no password set
	if user.PasswordHash == nil {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return Session{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return Session{}, ErrInvalidCredentials
	}

	if !user.Verified {
		return Session{}, ErrAccountPending
	}

	if user.IsBanned {
		return Session{}, ErrUserBanned
	}

	_ = s.q.TouchUserLastLogin(ctx, user.ID)
	return s.newSession(ctx, user)
}

func (s *AuthService) IsUserVerified(ctx context.Context, identifier *string) (bool, error) {
	verified, err := s.q.IsUserVerifiedByIdentifierAlt(ctx, identifier)
	if err != nil {
		return false, fmt.Errorf("is user verified: %w", err)
	}
	return verified, nil
}

// func (s *PasswordAuthService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
// 	user, err := s.q.GetUserByID(ctx, userID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return ErrInvalidCredentials
// 		}
// 		return fmt.Errorf("get user: %w", err)
// 	}
// 	// users who signed up via OTP may set a first password without an old one
// 	if user.PasswordHash != nil {
// 		if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(oldPassword)); err != nil {
// 			return ErrInvalidCredentials
// 		}
// 	}
// 	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
// 	if err != nil {
// 		return fmt.Errorf("hash password: %w", err)
// 	}
// 	if err := s.q.SetUserPassword(ctx, store.SetUserPasswordParams{ID: userID, PasswordHash: ptr(string(hash))}); err != nil {
// 		return fmt.Errorf("set password: %w", err)
// 	}
// 	return nil
// }

// Register mounts public routes on r; mount ChangePassword on an authenticated group.
// func (h *PasswordAuthHandler) Register(r gin.IRouter) {
// 	g := r.Group("/auth/password")
// 	g.POST("/register", h.SignUp)
// 	g.POST("/login", h.SignIn)
// }

// // userIDFromCtx: replace with however your auth middleware stores the user id.
// func (h *PasswordAuthHandler) ChangePassword(c *gin.Context) {
// 	var req changePasswordReq
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		_ = c.Error(err)
// 		return
// 	}
// 	if err := h.svc.ChangePassword(c.Request.Context(), c.GetString("user_id"), req.OldPassword, req.NewPassword); err != nil {
// 		_ = c.Error(err)
// 		return
// 	}
// 	httpx.NoContent(c)
// }

func (h *AuthService) RevokeToken() {}

func (s *AuthService) newSession(ctx context.Context, u store.User) (Session, error) {
	refresh, hash, err := newRefreshToken()
	fmt.Println()
	fmt.Println(refresh)
	fmt.Println(hash)
	fmt.Println()

	if err != nil {
		return Session{}, fmt.Errorf("refresh token: %w", err)
	}
	refreshExp := time.Now().Add(s.tokens.refreshTTL)

	session, err := s.q.CreateSession(ctx, store.CreateSessionParams{
		UserID:      u.ID,
		RefreshHash: hash,
		ExpiresAt:   refreshExp,
	})
	if err != nil {
		fmt.Println()
		fmt.Println(err)
		fmt.Println()
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	access, accessExp, err := s.tokens.IssueAccess(session.ID, u.ID, u.Role)
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
	exp := time.Now().Add(s.otp.TTL)

	id, err := utils.GenerateID(6)
	if err != nil {
		return Challenge{}, fmt.Errorf("generate id: %w", err)
	}
	ch, err := s.q.CreateOTPChallenge(ctx, store.CreateOTPChallengeParams{
		ID:        id,
		UserID:    u.ID,
		CodeHash:  hashOTP(s.tokens.secret, code, code),
		ExpiresAt: exp,
	})

	if err != nil {
		return Challenge{}, fmt.Errorf("create challenge: %w", err)
	}
	if err := s.sender.Send(ctx, u.Identifier, code); err != nil {
		return Challenge{}, fmt.Errorf("send otp: %w", err)
	}
	return Challenge{UserID: u.ID, ResendKey: ch.ID, ExpiresAt: exp}, nil
}
