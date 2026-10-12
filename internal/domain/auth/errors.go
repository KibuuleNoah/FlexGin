package auth

import "flexgin/internal/apperror"

var (
	ErrMissingToken = &apperror.Error{Kind: apperror.KindUnauthorized, Code: "missing_token", Message: "authentication required"}
	ErrInvalidToken = &apperror.Error{Kind: apperror.KindUnauthorized, Code: "invalid_token", Message: "invalid or expired token"}
	ErrInvalidOTP   = &apperror.Error{Kind: apperror.KindUnauthorized, Code: "invalid_otp", Message: "invalid or expired code"}

	ErrInvalidResendKey = &apperror.Error{Kind: apperror.KindValidation, Code: "invalid_resend_key", Message: "invalid or expired resend key"}
	ErrOTPCooldown      = apperror.TooManyRequests("otp_cooldown", "please wait before requesting another code")
	ErrUserNotFound     = apperror.NotFound("user_not_found", "user not found")
)
