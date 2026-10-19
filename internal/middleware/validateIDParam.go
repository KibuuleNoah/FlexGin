package middleware

import (
	"flexgin/internal/apperror"
	"regexp"

	"github.com/gin-gonic/gin"
)

// Must be exactly 6 alphanumeric characters
var idParamRe = regexp.MustCompile(`^[a-zA-Z0-9]{6}$`)

func ValidateIDParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// If an 'id' parameter exists on the route but fails validation
		if id != "" && !idParamRe.MatchString(id) {
			_ = c.Error(&apperror.Error{
				Kind:    apperror.KindValidation,
				Code:    "invalid_id_param",
				Message: "id must be exactly 6 alphanumeric characters",
			},
			)
			c.Abort()
			return
		}

		c.Next()
	}
}
