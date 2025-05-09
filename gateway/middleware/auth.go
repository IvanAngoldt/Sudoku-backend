package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization header format. Expected: Bearer <token>"})
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token is required"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
				c.AbortWithStatusJSON(401, gin.H{"error": "Token has expired"})
				return
			}
			c.Set("user_id", claims.UserID)
			c.Next()
		} else {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token claims"})
			return
		}
	}
}
