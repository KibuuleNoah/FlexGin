package auth_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"flexgin/internal/config"
	"flexgin/internal/domain/auth"
	"flexgin/internal/store"
)

const (
	testPhone  = "+256700000001"
	testSecret = "test-secret-test-secret-test-secret-123"
)

// fakeQ is an in-memory store.Querier covering only what the auth service calls.
// Unimplemented methods hit the nil embedded interface and panic, so unexpected calls fail loudly.
type fakeQ struct {
	store.Querier
	users      map[string]store.User
	challenges []store.OtpChallenge
	sessions   map[string]store.Session
}

func newFakeQ() *fakeQ {
	return &fakeQ{users: map[string]store.User{}, sessions: map[string]store.Session{}}
}

func (f *fakeQ) UpsertUser(_ context.Context, identifier string) (store.User, error) {
	for _, u := range f.users {
		if u.Identifier == identifier {
			return u, nil
		}
	}
	u := store.User{
		ID: fmt.Sprintf("abcde%d", len(f.users)+1), Identifier: identifier,
		Role: "guest", CreatedAt: time.Now().Unix(),
	}
	f.users[u.ID] = u
	return u, nil
}

func (f *fakeQ) GetUserByID(_ context.Context, id string) (store.User, error) {
	u, ok := f.users[id]
	if !ok {
		return store.User{}, sql.ErrNoRows
	}
	return u, nil
}

func (f *fakeQ) MarkUserVerified(_ context.Context, id string) (store.User, error) {
	u, ok := f.users[id]
	if !ok {
		return store.User{}, sql.ErrNoRows
	}
	u.Verified = true
	if u.Role == "guest" {
		u.Role = "user"
	}
	f.users[id] = u
	return u, nil
}

func (f *fakeQ) TouchUserLastLogin(context.Context, string) error { return nil }

func (f *fakeQ) CreateOTPChallenge(_ context.Context, p store.CreateOTPChallengeParams) (store.OtpChallenge, error) {
	c := store.OtpChallenge{
		ID: p.ID, UserID: p.UserID, CodeHash: p.CodeHash,
		ExpiresAt: p.ExpiresAt, CreatedAt: time.Now().Unix(),
	}
	f.challenges = append(f.challenges, c)
	return c, nil
}

func (f *fakeQ) GetLatestOTPChallenge(_ context.Context, userID string) (store.OtpChallenge, error) {
	for i := len(f.challenges) - 1; i >= 0; i-- {
		if f.challenges[i].UserID == userID {
			return f.challenges[i], nil
		}
	}
	return store.OtpChallenge{}, sql.ErrNoRows
}

func (f *fakeQ) GetActiveOTPChallenge(_ context.Context, userID string) (store.OtpChallenge, error) {
	for i := len(f.challenges) - 1; i >= 0; i-- {
		c := f.challenges[i]
		if c.UserID == userID && !c.ConsumedAt.Valid && c.ExpiresAt.After(time.Now()) {
			return c, nil
		}
	}
	return store.OtpChallenge{}, sql.ErrNoRows
}

func (f *fakeQ) GetOTPChallenge(_ context.Context, id string) (store.OtpChallenge, error) {
	for _, c := range f.challenges {
		if c.ID == id && !c.ConsumedAt.Valid && c.ExpiresAt.After(time.Now()) {
			return c, nil
		}
	}
	return store.OtpChallenge{}, sql.ErrNoRows
}

func (f *fakeQ) IncrementOTPAttempts(_ context.Context, id string) error {
	for i := range f.challenges {
		if f.challenges[i].ID == id {
			f.challenges[i].Attempts++
		}
	}
	return nil
}

func (f *fakeQ) ConsumeOTPChallenge(_ context.Context, id string) (int64, error) {
	for i := range f.challenges {
		c := &f.challenges[i]
		if c.ID == id && !c.ConsumedAt.Valid && c.ExpiresAt.After(time.Now()) {
			c.ConsumedAt = sql.NullTime{Time: time.Now(), Valid: true}
			return 1, nil
		}
	}
	return 0, nil
}

func (f *fakeQ) CreateSession(_ context.Context, p store.CreateSessionParams) (store.Session, error) {
	s := store.Session{ID: p.ID, UserID: p.UserID, RefreshHash: p.RefreshHash, ExpiresAt: p.ExpiresAt, CreatedAt: time.Now().Unix()}
	f.sessions[s.ID] = s
	return s, nil
}

func (f *fakeQ) SessionIsActive(_ context.Context, id string) (bool, error) {
	s, ok := f.sessions[id]
	return ok && !s.RevokedAt.Valid && s.ExpiresAt.After(time.Now()), nil
}

func (f *fakeQ) ClaimRefreshSession(_ context.Context, hash string) (store.Session, error) {
	for id, s := range f.sessions {
		if s.RefreshHash == hash && !s.RevokedAt.Valid && s.ExpiresAt.After(time.Now()) {
			s.RevokedAt = sql.NullTime{Time: time.Now(), Valid: true}
			f.sessions[id] = s
			return s, nil
		}
	}
	return store.Session{}, sql.ErrNoRows
}

func (f *fakeQ) RevokeSession(_ context.Context, id string) (int64, error) {
	s, ok := f.sessions[id]
	if !ok || s.RevokedAt.Valid {
		return 0, nil
	}
	s.RevokedAt = sql.NullTime{Time: time.Now(), Valid: true}
	f.sessions[id] = s
	return 1, nil
}

func (f *fakeQ) expireChallenges() {
	for i := range f.challenges {
		f.challenges[i].ExpiresAt = time.Now().Add(-time.Minute)
	}
}

type captureSender struct {
	phone, code string
	calls       int
	err         error
}

func (s *captureSender) Send(_ context.Context, identifier, code string) error {
	s.calls++
	s.phone, s.code = identifier, code
	return s.err
}

func jwtCfg() config.JWT {
	return config.JWT{Secret: testSecret, Issuer: "flex-test", AccessTTL: 15 * time.Minute, RefreshTTL: 24 * time.Hour}
}

func defaultOTP() config.OTP {
	return config.OTP{TTL: 5 * time.Minute, Cooldown: 0, MaxAttempts: 3}
}

func newService(otp config.OTP) (*auth.AuthService, *fakeQ, *captureSender) {
	q, sender := newFakeQ(), &captureSender{}
	return auth.NewAuthService(q, auth.NewTokenManager(jwtCfg()), sender, otp), q, sender
}

func wrongCode(code string) string {
	if code == "000000" {
		return "111111"
	}
	return "000000"
}

func login(t *testing.T, svc *auth.AuthService, s *captureSender) auth.Session {
	t.Helper()
	ctx := context.Background()
	ch, err := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
	if err != nil {
		t.Fatalf("RequestOTP: %v", err)
	}
	sess, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: ch.UserID, OTP: s.code})
	if err != nil {
		t.Fatalf("VerifyOTP: %v", err)
	}
	return sess
}
