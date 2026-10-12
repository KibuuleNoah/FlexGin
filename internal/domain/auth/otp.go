package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
)

// Sender delivers the code (SMS, email, ...). Implement it and pass it in server.go.
type Sender interface {
	Send(ctx context.Context, identifier, code string) error
}

var ErrSenderNotConfigured = errors.New("otp sender not configured")

// LogSender prints codes to the log. Fails closed when Enabled is false (prod).
type LogSender struct {
	Log     *slog.Logger
	Enabled bool
}

func (l LogSender) Send(ctx context.Context, identifier, code string) error {
	if !l.Enabled {
		return ErrSenderNotConfigured
	}
	l.Log.InfoContext(ctx, "otp (dev only)", "identifier", identifier, "code", code)
	return nil
}

func newOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashOTP(secret []byte, id string, code string) string {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(id + ":" + code))
	return hex.EncodeToString(m.Sum(nil))
}
