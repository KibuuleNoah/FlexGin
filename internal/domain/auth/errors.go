package auth

import "flexgin/internal/apperror"

var (
	ErrMissingToken = &apperror.Error{Kind: apperror.KindUnauthorized, Code: "missing_token", Message: "authentication required"}
	ErrInvalidToken = &apperror.Error{Kind: apperror.KindUnauthorized, Code: "invalid_token", Message: "invalid or expired token"}
	ErrInvalidOTP   = &apperror.Error{Kind: apperror.KindUnauthorized, Code: "invalid_otp", Message: "invalid or expired code"}

	ErrInvalidResendKey = &apperror.Error{Kind: apperror.KindValidation, Code: "invalid_resend_key", Message: "invalid or expired resend key"}
	ErrOTPCooldown      = apperror.TooManyRequests("otp_cooldown", "please wait before requesting another code")
	ErrUserNotFound     = apperror.NotFound("user_not_found", "user not found")

	ErrInvalidCredentials = apperror.Unauthorized("Email or Password is wrong")
	ErrUserBanned         = apperror.Forbidden("account is banned")
	ErrUsernameTaken      = apperror.Conflict("USERNAME_TAKEN", "username already taken")
	ErrIdentifierTaken    = apperror.Conflict("IDENTIFIER_TAKEN", "email address or phone number already registered")

	ErrAccountPending = apperror.Forbidden("account is still waiting in approval")
)
