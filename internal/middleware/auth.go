package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"pajo/services"
)

// TokenAuth validates the JWT access token and sets userID in the context.
func TokenAuth() gin.HandlerFunc {
	Asvc := services.NewAuthService()
	return func(c *gin.Context) {
		accessDetails, err := Asvc.ExtractTokenMetadata(c.Request)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication Required"})
			return
		}

		// log.Println("\n\n\n*****", tokenAuth)
		// userID, err := authModel.FetchAuth(tokenAuth)
		// if err != nil {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication Required"})
		// 	return
		// }

		c.Set("userID", accessDetails.UserID)
		c.Next()
	}
}

// AuthRequired handles dual authentication: Guest ID first, JWT fallback second.
func GuestOrUserRequired() gin.HandlerFunc {

	return func(c *gin.Context) {
		guestID := c.GetHeader("X-GUEST-ID")

		if guestID != "" {
			// #####################################
			// GUESTID VERIFICATION LOGIC HERE
			// #####################################

			c.Set("guestID", guestID)
			c.Set("authType", "guest")
			c.Next()
			return
		}

		userID := c.GetHeader("X-USER-ID")
		if len(userID) != 6 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication Required (No valid Guest ID or Token found)"})
			return
		}

		// accessDetails, err := Asvc.ExtractTokenMetadata(c.Request)
		// if err != nil {
		// 	log.Println(err)
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication Required (No valid Guest ID or Token found)"})
		// 	return
		// }

		// c.Set("userID", accessDetails.UserID)
		c.Set("userID", userID)
		c.Set("authType", "user")
		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	// Asvc := services.NewAuthService()

	return func(c *gin.Context) {
		userID := c.GetHeader("X-USER-ID")

		if len(userID) != 6 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication Required (No valid Guest ID or Token found)"})
			return
		}

		// accessDetails, err := Asvc.ExtractTokenMetadata(c.Request)
		// if err != nil {
		// 	log.Println(err)
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication Required (No valid Guest ID or Token found)"})
		// 	return
		// }

		// c.Set("userID", accessDetails.UserID)
		c.Set("userID", userID)
		c.Next()
	}
}
