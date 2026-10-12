package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const ClaimsKey = "auth.claims"

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (c *Claims) UserID() uuid.UUID {
	id, _ := uuid.Parse(c.Subject)
	return id
}

func ClaimsFrom(c *gin.Context) (*Claims, bool) {
	v, ok := c.Get(ClaimsKey)
	if !ok {
		return nil, false
	}
	cl, ok := v.(*Claims)
	return cl, ok
}
