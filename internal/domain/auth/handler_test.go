package auth_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"flexgin/internal/config"
	"flexgin/internal/domain/auth"
	"flexgin/internal/httpx"
	"flexgin/internal/middleware"
)

type envelope struct {
	Data  map[string]any `json:"data"`
	Error *struct {
		Code   string            `json:"code"`
		Fields map[string]string `json:"fields"`
	} `json:"error"`
}

func newRouter(otp config.OTP) (http.Handler, *captureSender) {
	gin.SetMode(gin.TestMode)
	httpx.UseJSONFieldNames()

	svc, _, sender := newService(otp)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.ErrorHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
	auth.RegisterAuthRoutes(r.Group("/api/v1"), svc, middleware.Authenticate(svc))
	return r, sender
}

func do(r http.Handler, method, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) envelope {
	t.Helper()
	var e envelope
	if w.Body.Len() == 0 {
		return e
	}
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode: %v, body = %s", err, w.Body)
	}
	return e
}

func requestOTP(t *testing.T, r http.Handler) envelope {
	t.Helper()
	w := do(r, http.MethodPost, "/api/v1/auth/request-otp", fmt.Sprintf(`{"identifier":%q}`, testPhone), "")
	if w.Code != http.StatusOK {
		t.Fatalf("request-otp status = %d, body = %s", w.Code, w.Body)
	}
	return decode(t, w)
}

func signIn(t *testing.T, r http.Handler, s *captureSender) map[string]any {
	t.Helper()
	ch := requestOTP(t, r)
	body := fmt.Sprintf(`{"user_id":%q,"otp":%q}`, ch.Data["user_id"], s.code)
	w := do(r, http.MethodPost, "/api/v1/auth/with-otp", body, "")
	if w.Code != http.StatusOK {
		t.Fatalf("with-otp status = %d, body = %s", w.Code, w.Body)
	}
	return decode(t, w).Data
}

func TestHandler_RequestOTP(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		r, sender := newRouter(defaultOTP())
		e := requestOTP(t, r)
		if e.Data["user_id"] == "" || e.Data["otp_resend_key"] == "" || e.Data["expires_at"] == nil {
			t.Fatalf("bad data: %+v", e.Data)
		}
		if sender.calls != 1 {
			t.Fatalf("sender calls = %d", sender.calls)
		}
	})

	t.Run("missing identifier", func(t *testing.T) {
		r, _ := newRouter(defaultOTP())
		w := do(r, http.MethodPost, "/api/v1/auth/request-otp", `{}`, "")
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", w.Code)
		}
		if e := decode(t, w); e.Error == nil || e.Error.Fields["identifier"] == "" {
			t.Fatalf("missing field error: %+v", e.Error)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		r, _ := newRouter(defaultOTP())
		w := do(r, http.MethodPost, "/api/v1/auth/request-otp", `{`, "")
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", w.Code)
		}
	})

	t.Run("invalid identifier", func(t *testing.T) {
		r, _ := newRouter(defaultOTP())
		w := do(r, http.MethodPost, "/api/v1/auth/request-otp", `{"identifier":"abc"}`, "")
		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d", w.Code)
		}
	})

	t.Run("cooldown returns 429", func(t *testing.T) {
		r, _ := newRouter(config.OTP{TTL: 5 * time.Minute, Cooldown: time.Minute, MaxAttempts: 3})
		requestOTP(t, r)
		w := do(r, http.MethodPost, "/api/v1/auth/request-otp", fmt.Sprintf(`{"identifier":%q}`, testPhone), "")
		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("status = %d, body = %s", w.Code, w.Body)
		}
		if e := decode(t, w); e.Error == nil || e.Error.Code != "otp_cooldown" {
			t.Fatalf("unexpected error: %+v", e.Error)
		}
	})
}

func TestHandler_ResendOTP(t *testing.T) {
	r, sender := newRouter(defaultOTP())
	first := requestOTP(t, r)
	w := do(r, http.MethodPost, "/api/v1/auth/resend-otp", fmt.Sprintf(`{"otp_resend_key":%q}`, first.Data["otp_resend_key"]), "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	if second := decode(t, w); second.Data["otp_resend_key"] == first.Data["otp_resend_key"] {
		t.Fatal("resend key must change")
	}
	if sender.calls != 2 {
		t.Fatalf("sender calls = %d", sender.calls)
	}

	w = do(r, http.MethodPost, "/api/v1/auth/resend-otp", `{"otp_resend_key":"nope"}`, "")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad uuid status = %d", w.Code)
	}
	w = do(r, http.MethodPost, "/api/v1/auth/resend-otp", `{"otp_resend_key":"6f1b1c0e-7f3a-4a53-9f43-0b6b7f1e2a11"}`, "")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown key status = %d", w.Code)
	}
}

func TestHandler_VerifyOTP(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		r, sender := newRouter(defaultOTP())
		data := signIn(t, r, sender)
		if data["access_token"] == "" || data["refresh_token"] == "" || data["token_type"] != "Bearer" {
			t.Fatalf("bad auth response: %+v", data)
		}
		if _, ok := data["user"].(map[string]any); !ok {
			t.Fatalf("missing user: %+v", data)
		}
	})

	t.Run("wrong code is 401", func(t *testing.T) {
		r, sender := newRouter(defaultOTP())
		ch := requestOTP(t, r)
		body := fmt.Sprintf(`{"user_id":%q,"otp":%q}`, ch.Data["user_id"], wrongCode(sender.code))
		w := do(r, http.MethodPost, "/api/v1/auth/with-otp", body, "")
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", w.Code)
		}
		if e := decode(t, w); e.Error == nil || e.Error.Code != "invalid_otp" {
			t.Fatalf("unexpected error: %+v", e.Error)
		}
	})

	t.Run("validation", func(t *testing.T) {
		r, _ := newRouter(defaultOTP())
		for name, body := range map[string]string{
			"short otp":      `{"user_id":"u1","otp":"12"}`,
			"non numeric":    `{"user_id":"u1","otp":"abcdef"}`,
			"missing user":   `{"otp":"123456"}`,
			"missing fields": `{}`,
		} {
			if w := do(r, http.MethodPost, "/api/v1/auth/with-otp", body, ""); w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("%s: status = %d, body = %s", name, w.Code, w.Body)
			}
		}
	})
}

func TestHandler_Refresh(t *testing.T) {
	r, sender := newRouter(defaultOTP())
	first := signIn(t, r, sender)
	body := fmt.Sprintf(`{"refresh_token":%q}`, first["refresh_token"])

	w := do(r, http.MethodPost, "/api/v1/auth/token/refresh", body, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	if next := decode(t, w).Data; next["refresh_token"] == first["refresh_token"] {
		t.Fatal("refresh token must rotate")
	}

	if w := do(r, http.MethodPost, "/api/v1/auth/token/refresh", body, ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("reuse status = %d", w.Code)
	}
	if w := do(r, http.MethodPost, "/api/v1/auth/token/refresh", `{}`, ""); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty body status = %d", w.Code)
	}
}

func TestHandler_Me(t *testing.T) {
	r, sender := newRouter(defaultOTP())

	w := do(r, http.MethodGet, "/api/v1/auth/me", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("no token status = %d", w.Code)
	}
	if e := decode(t, w); e.Error == nil || e.Error.Code != "missing_token" {
		t.Fatalf("unexpected error: %+v", e.Error)
	}

	data := signIn(t, r, sender)
	user := data["user"].(map[string]any)

	w = do(r, http.MethodGet, "/api/v1/auth/me", "", data["access_token"].(string))
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	if got := decode(t, w).Data["id"]; got != user["id"] {
		t.Fatalf("id = %v, want %v", got, user["id"])
	}
}

func TestHandler_Logout(t *testing.T) {
	r, sender := newRouter(defaultOTP())
	data := signIn(t, r, sender)
	access := data["access_token"].(string)

	if w := do(r, http.MethodPost, "/api/v1/auth/logout", "", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("no token status = %d", w.Code)
	}
	if w := do(r, http.MethodPost, "/api/v1/auth/logout", "", access); w.Code != http.StatusNoContent {
		t.Fatalf("logout status = %d", w.Code)
	}

	w := do(r, http.MethodGet, "/api/v1/auth/me", "", access)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("token after logout status = %d", w.Code)
	}
	if e := decode(t, w); e.Error == nil || e.Error.Code != "invalid_token" {
		t.Fatalf("unexpected error: %+v", e.Error)
	}
}
