package middleware

import (
	"os"
	"strings"

	"github.com/HappyNaCl/new-debyut/internal/common/tokens"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func OptionalJwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &tokens.AccessTokenClaims{}

		if _, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		}); err != nil {
			c.Next()
			return
		}

		userId, err := claims.GetSubject()
		if err != nil || userId == "" || claims.Role == "" {
			c.Next()
			return
		}

		c.Set("userId", userId)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}
