package auth

import (
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
