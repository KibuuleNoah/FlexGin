package auth

import (
	"time"

	"flexgin/internal/store"
)

type Challenge struct {
	UserID    string
	ResendKey string
	ExpiresAt time.Time
}

type Session struct {
	User             store.User `json:"user"`
	AccessToken      string     `json:"access_token"`
	RefreshToken     string     `json:"refresh_token"`
	AccessExpiresAt  time.Time  `json:"access_expires_at"`
	RefreshExpiresAt time.Time  `json:"refresh_expires_at"`
}

type isVerifiedRequest struct {
	Identifier string `form:"identifier" binding:"required,max=254"`
}

type RequestOTPRequest struct {
	Identifier string `json:"identifier" binding:"required,max=254"`
}

type ResendOTPRequest struct {
	ResendKey string `json:"otp_resend_key" binding:"required,len=6"`
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

func ptr[T any](v T) *T { return &v }

type PasswordSignupReq struct {
	Username   string `json:"username" binding:"required,min=3,max=20,alphanumunicode|contains=_"`
	Identifier string `json:"identifier" binding:"required,max=254"`
	Password   string `json:"password" binding:"required,min=8,max=72"`
}

type PasswordSigninReq struct {
	UsernameOrIdentifier string `json:"username_or_identifier" binding:"required,max=254"`
	Password             string `json:"password" binding:"required,max=72"`
}

//
// type changePasswordReq struct {
// 	OldPassword string `json:"old_password" binding:"max=72"`
// 	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
// }

func ToChallengeResponse(c Challenge) ChallengeResponse {
	return ChallengeResponse{UserID: c.UserID, ResendKey: c.ResendKey, ExpiresAt: c.ExpiresAt}
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
