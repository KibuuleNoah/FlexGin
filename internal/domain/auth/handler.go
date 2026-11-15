package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"flexgin/internal/httpx"
)

type AuthHandler struct{ svc *AuthService }

func NewAuthHandler(svc *AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req RequestOTPRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	ch, err := h.svc.RequestOTP(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, ToChallengeResponse(ch))
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req ResendOTPRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	ch, err := h.svc.ResendOTP(c.Request.Context(), req.ResendKey)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, ToChallengeResponse(ch))
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	s, err := h.svc.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, ToAuthResponse(s))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		fmt.Println()
		fmt.Println(err)
		fmt.Println()
		_ = c.Error(err)
		return
	}
	s, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, ToAuthResponse(s))
}

func (h *AuthHandler) PasswordAuthSignUp(c *gin.Context) {
	var req PasswordSignupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	err := h.svc.PasswordAuthSignUp(c.Request.Context(), req.Username, req.Identifier, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	httpx.Created(c, "")
}

func (h *AuthHandler) PasswordAuthSignIn(c *gin.Context) {
	var req PasswordSigninReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	res, err := h.svc.PasswordAuthSignIn(c.Request.Context(), req.UsernameOrIdentifier, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	claims, ok := ClaimsFrom(c)
	if !ok {
		_ = c.Error(ErrMissingToken)
		return
	}
	if err := h.svc.Logout(c.Request.Context(), claims.ID); err != nil {
		_ = c.Error(err)
		return
	}
	httpx.NoContent(c)
}

func (h *AuthHandler) IsUserVerified(c *gin.Context) {
	var req isVerifiedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(err)
		return
	}

	verified, err := h.svc.IsUserVerified(c.Request.Context(), &req.Identifier)
	if err != nil {
		_ = c.Error(err)
		return
	}

	httpx.OK(c, map[string]bool{"verified": verified})
}

func (h *AuthHandler) Me(c *gin.Context) {
	claims, ok := ClaimsFrom(c)
	if !ok {
		_ = c.Error(ErrMissingToken)
		return
	}
	u, err := h.svc.Me(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, ToUserResponse(u))
}

func (h *AuthHandler) RevokeToken(c *gin.Context) {}
