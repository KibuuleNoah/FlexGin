package auth

import (
	"time"

	"flexgin/internal/store"
)

type RequestOTPRequest struct {
	Identifier string `json:"identifier" binding:"required,max=254"`
}

type ResendOTPRequest struct {
	ResendKey string `json:"otp_resend_key" binding:"required,uuid"`
}

type VerifyOTPRequest struct {
	UserID string `json:"user_id" binding:"required,gt=0"`
	OTP    string `json:"otp"     binding:"required,len=6,numeric"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChallengeResponse struct {
	UserID    string    `json:"user_id"`
	ResendKey string    `json:"otp_resend_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

type UserResponse struct {
	ID         string  `json:"id"`
	Identifier string  `json:"identifier"`
	Name       *string `json:"name"`
	Role       string  `json:"role"`
	CreatedAt  int64   `json:"created_at"`
}

type AuthResponse struct {
	AccessToken      string       `json:"access_token"`
	RefreshToken     string       `json:"refresh_token"`
	TokenType        string       `json:"token_type"`
	AccessExpiresAt  time.Time    `json:"access_expires_at"`
	RefreshExpiresAt time.Time    `json:"refresh_expires_at"`
	User             UserResponse `json:"user"`
}

func ToChallengeResponse(c Challenge) ChallengeResponse {
	return ChallengeResponse{UserID: c.UserID, ResendKey: c.ResendKey.String(), ExpiresAt: c.ExpiresAt}
}

func ToUserResponse(u store.User) UserResponse {
	return UserResponse{ID: u.ID, Identifier: u.Identifier, Name: u.Username, Role: u.Role, CreatedAt: u.CreatedAt}
}

func ToAuthResponse(s Session) AuthResponse {
	return AuthResponse{
		AccessToken:      s.AccessToken,
		RefreshToken:     s.RefreshToken,
		TokenType:        "Bearer",
		AccessExpiresAt:  s.AccessExpiresAt,
		RefreshExpiresAt: s.RefreshExpiresAt,
		User:             ToUserResponse(s.User),
	}
}
