package auth

import (
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, authSvc *AuthService, authn gin.HandlerFunc) {
	h := NewAuthHandler(authSvc)
	g := rg.Group("/auth")
	g.POST("/pwd/signup", h.PasswordAuthSignUp)
	g.POST("/pwd/signin", h.PasswordAuthSignIn)
	g.POST("/request-otp", h.RequestOTP)
	g.POST("/resend-otp", h.ResendOTP)
	g.POST("/with-otp", h.VerifyOTP)
	g.POST("/token/refresh", h.Refresh)
	g.POST("/logout", authn, h.Logout)
	g.GET("/me", authn, h.Me)
	g.GET("/pwd/verified", h.IsUserVerified)
}
