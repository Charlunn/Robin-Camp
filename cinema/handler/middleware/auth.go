package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type authError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RequireBearerToken(expected string) gin.HandlerFunc {
	token := strings.TrimSpace(expected)
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			unauthorised(c)
			return
		}

		provided := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if provided == "" || provided != token {
			unauthorised(c)
			return
		}

		c.Next()
	}
}

func unauthorised(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, authError{
		Code:    "UNAUTHORIZED",
		Message: "Missing or invalid authentication information",
	})
}
