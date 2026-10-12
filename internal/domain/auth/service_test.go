package auth_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"

	"flexgin/internal/apperror"
	"flexgin/internal/config"
	"flexgin/internal/domain/auth"
)

var sixDigits = regexp.MustCompile(`^[0-9]{6}$`)

func TestService_RequestOTP(t *testing.T) {
	ctx := context.Background()

	t.Run("creates guest and sends code", func(t *testing.T) {
		svc, q, sender := newService(defaultOTP())

		ch, err := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		if err != nil {
			t.Fatal(err)
		}
		if ch.UserID == "" || ch.ResendKey == "" || !ch.ExpiresAt.After(time.Now()) {
			t.Fatalf("bad challenge: %+v", ch)
		}
		if sender.calls != 1 || sender.phone != testPhone || !sixDigits.MatchString(sender.code) {
			t.Fatalf("bad send: %+v", sender)
		}
		if u := q.users[ch.UserID]; u.Role != "guest" || u.Verified {
			t.Fatalf("new user should be unverified guest: %+v", u)
		}
	})

	t.Run("same phone reuses user", func(t *testing.T) {
		svc, q, _ := newService(defaultOTP())
		a, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		b, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		if a.UserID != b.UserID || len(q.users) != 1 {
			t.Fatalf("expected one user, got %d", len(q.users))
		}
	})

	t.Run("invalid identifier", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		_, err := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: "not-a-phone"})
		var ae *apperror.Error
		if !errors.As(err, &ae) || ae.Kind != apperror.KindValidation {
			t.Fatalf("want validation error, got %v", err)
		}
		if sender.calls != 0 {
			t.Fatal("must not send for invalid identifier")
		}
	})

	t.Run("sender failure surfaces", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		sender.err = errors.New("sms down")
		if _, err := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone}); err == nil {
			t.Fatal("want error")
		}
	})

	t.Run("cooldown", func(t *testing.T) {
		svc, _, sender := newService(config.OTP{TTL: 5 * time.Minute, Cooldown: time.Minute, MaxAttempts: 3})
		if _, err := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone}); err != nil {
			t.Fatal(err)
		}
		_, err := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		if !errors.Is(err, auth.ErrOTPCooldown) {
			t.Fatalf("want ErrOTPCooldown, got %v", err)
		}
		if sender.calls != 1 {
			t.Fatalf("sender calls = %d, want 1", sender.calls)
		}
	})
}

func TestService_ResendOTP(t *testing.T) {
	ctx := context.Background()

	t.Run("ok issues a new challenge", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		first, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		oldCode := sender.code

		second, err := svc.ResendOTP(ctx, first.ResendKey)
		if err != nil {
			t.Fatal(err)
		}
		if second.ResendKey == first.ResendKey || second.UserID != first.UserID || sender.calls != 2 {
			t.Fatalf("bad resend: %+v calls=%d", second, sender.calls)
		}

		if sender.code != oldCode { // the old code must stop working once a new one is issued
			_, err = svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: first.UserID, OTP: oldCode})
			if !errors.Is(err, auth.ErrInvalidOTP) {
				t.Fatalf("old code must be rejected, got %v", err)
			}
		}
	})

	t.Run("invalid keys", func(t *testing.T) {
		svc, _, _ := newService(defaultOTP())
		for name, key := range map[string]string{"not a uuid": "abc", "unknown": uuid.NewString()} {
			if _, err := svc.ResendOTP(ctx, key); !errors.Is(err, auth.ErrInvalidResendKey) {
				t.Fatalf("%s: want ErrInvalidResendKey, got %v", name, err)
			}
		}
	})

	t.Run("expired key", func(t *testing.T) {
		svc, q, _ := newService(defaultOTP())
		ch, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		q.expireChallenges()
		if _, err := svc.ResendOTP(ctx, ch.ResendKey); !errors.Is(err, auth.ErrInvalidResendKey) {
			t.Fatalf("want ErrInvalidResendKey, got %v", err)
		}
	})
}

func TestService_VerifyOTP(t *testing.T) {
	ctx := context.Background()

	t.Run("success promotes guest to verified user", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		sess := login(t, svc, sender)

		if sess.AccessToken == "" || sess.RefreshToken == "" {
			t.Fatalf("missing tokens: %+v", sess)
		}
		if !sess.AccessExpiresAt.Before(sess.RefreshExpiresAt) {
			t.Fatal("refresh must outlive access")
		}
		if !sess.User.Verified || sess.User.Role != "user" {
			t.Fatalf("user not promoted: %+v", sess.User)
		}
	})

	t.Run("wrong code counts an attempt", func(t *testing.T) {
		svc, q, sender := newService(defaultOTP())
		ch, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})

		_, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: ch.UserID, OTP: wrongCode(sender.code)})
		if !errors.Is(err, auth.ErrInvalidOTP) {
			t.Fatalf("want ErrInvalidOTP, got %v", err)
		}
		if q.challenges[0].Attempts != 1 {
			t.Fatalf("attempts = %d, want 1", q.challenges[0].Attempts)
		}
	})

	t.Run("max attempts locks even the correct code", func(t *testing.T) {
		svc, _, sender := newService(config.OTP{TTL: 5 * time.Minute, MaxAttempts: 2})
		ch, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		bad := wrongCode(sender.code)

		for i := 0; i < 2; i++ {
			_, _ = svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: ch.UserID, OTP: bad})
		}
		_, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: ch.UserID, OTP: sender.code})
		if !errors.Is(err, auth.ErrInvalidOTP) {
			t.Fatalf("want ErrInvalidOTP, got %v", err)
		}
	})

	t.Run("expired challenge", func(t *testing.T) {
		svc, q, sender := newService(defaultOTP())
		ch, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		q.expireChallenges()

		_, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: ch.UserID, OTP: sender.code})
		if !errors.Is(err, auth.ErrInvalidOTP) {
			t.Fatalf("want ErrInvalidOTP, got %v", err)
		}
	})

	t.Run("code is single use", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		ch, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		req := auth.VerifyOTPRequest{UserID: ch.UserID, OTP: sender.code}

		if _, err := svc.VerifyOTP(ctx, req); err != nil {
			t.Fatal(err)
		}
		if _, err := svc.VerifyOTP(ctx, req); !errors.Is(err, auth.ErrInvalidOTP) {
			t.Fatalf("replay: want ErrInvalidOTP, got %v", err)
		}
	})

	t.Run("no challenge for user", func(t *testing.T) {
		svc, _, _ := newService(defaultOTP())
		_, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: "nobody", OTP: "123456"})
		if !errors.Is(err, auth.ErrInvalidOTP) {
			t.Fatalf("want ErrInvalidOTP, got %v", err)
		}
	})

	t.Run("banned user gets no session", func(t *testing.T) {
		svc, q, sender := newService(defaultOTP())
		ch, _ := svc.RequestOTP(ctx, auth.RequestOTPRequest{Identifier: testPhone})
		u := q.users[ch.UserID]
		u.IsBanned = true
		q.users[ch.UserID] = u

		sess, err := svc.VerifyOTP(ctx, auth.VerifyOTPRequest{UserID: ch.UserID, OTP: sender.code})
		if err == nil || sess.AccessToken != "" {
			t.Fatalf("banned user must not receive tokens, err=%v", err)
		}
		if len(q.sessions) != 0 {
			t.Fatal("no session row may be created for a banned user")
		}
	})
}

func TestService_Authenticate(t *testing.T) {
	ctx := context.Background()

	t.Run("valid token", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		sess := login(t, svc, sender)

		cl, err := svc.Authenticate(ctx, sess.AccessToken)
		if err != nil {
			t.Fatal(err)
		}
		if cl.UserID != sess.User.ID || cl.Role != "user" || cl.ID == "" {
			t.Fatalf("bad claims: %+v", cl)
		}
	})

	t.Run("garbage, wrong secret and expired tokens", func(t *testing.T) {
		svc, _, _ := newService(defaultOTP())

		otherCfg := jwtCfg()
		otherCfg.Secret = "another-secret-another-secret-another-1"
		wrongSecret, _, _ := auth.NewTokenManager(otherCfg).IssueAccess(uuid.New().String(), "u1", "user")

		expCfg := jwtCfg()
		expCfg.AccessTTL = -time.Minute
		expired, _, _ := auth.NewTokenManager(expCfg).IssueAccess(uuid.New().String(), "u1", "user")

		for name, tok := range map[string]string{"garbage": "abc.def.ghi", "wrong secret": wrongSecret, "expired": expired} {
			if _, err := svc.Authenticate(ctx, tok); !errors.Is(err, auth.ErrInvalidToken) {
				t.Fatalf("%s: want ErrInvalidToken, got %v", name, err)
			}
		}
	})

	t.Run("valid signature but unknown session", func(t *testing.T) {
		svc, _, _ := newService(defaultOTP())
		tok, _, _ := auth.NewTokenManager(jwtCfg()).IssueAccess(uuid.New().String(), "u1", "user")
		if _, err := svc.Authenticate(ctx, tok); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("want ErrInvalidToken, got %v", err)
		}
	})
}

func TestService_Refresh(t *testing.T) {
	ctx := context.Background()

	t.Run("rotates the pair and invalidates the old one", func(t *testing.T) {
		svc, _, sender := newService(defaultOTP())
		old := login(t, svc, sender)

		next, err := svc.Refresh(ctx, old.RefreshToken)
		if err != nil {
			t.Fatal(err)
		}
		if next.RefreshToken == old.RefreshToken || next.AccessToken == old.AccessToken {
			t.Fatal("tokens must rotate")
		}
		if _, err := svc.Authenticate(ctx, next.AccessToken); err != nil {
			t.Fatalf("new access token rejected: %v", err)
		}
		if _, err := svc.Authenticate(ctx, old.AccessToken); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("old access token must die with its session, got %v", err)
		}
		if _, err := svc.Refresh(ctx, old.RefreshToken); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("refresh reuse: want ErrInvalidToken, got %v", err)
		}
	})

	t.Run("unknown token", func(t *testing.T) {
		svc, _, _ := newService(defaultOTP())
		if _, err := svc.Refresh(ctx, "does-not-exist"); !errors.Is(err, auth.ErrInvalidToken) {
			t.Fatalf("want ErrInvalidToken, got %v", err)
		}
	})
}

func TestService_Logout(t *testing.T) {
	ctx := context.Background()
	svc, _, sender := newService(defaultOTP())
	sess := login(t, svc, sender)
	cl, err := svc.Authenticate(ctx, sess.AccessToken)
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.Logout(ctx, cl.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Authenticate(ctx, sess.AccessToken); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("token must be dead after logout, got %v", err)
	}
	if _, err := svc.Refresh(ctx, sess.RefreshToken); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("refresh must be dead after logout, got %v", err)
	}
	if err := svc.Logout(ctx, "not-a-uuid"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestService_Me(t *testing.T) {
	ctx := context.Background()
	svc, _, sender := newService(defaultOTP())
	sess := login(t, svc, sender)

	u, err := svc.Me(ctx, sess.User.ID)
	if err != nil || u.ID != sess.User.ID {
		t.Fatalf("Me: %+v, %v", u, err)
	}
	if _, err := svc.Me(ctx, "missing"); !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}
