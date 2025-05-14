package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func AuthRequired(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 📌 Request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 🔐 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c, "Authorization header is required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			unauthorized(c, "Invalid authorization header format. Expected: Bearer <token>")
			return
		}

		tokenStr := parts[1]
		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			unauthorized(c, "Invalid or expired token")
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			unauthorized(c, "Token has expired")
			return
		}

		// Прокидываем user_id и роль
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Writer.Header().Set("X-User-Role", claims.Role)

		// Продолжаем выполнение
		c.Next()

		// 📜 Логгирование после обработки
		latency := time.Since(start)
		status := c.Writer.Status()
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.FullPath(),
			"status":     status,
			"user_id":    claims.UserID,
			"user_role":  claims.Role,
			"request_id": requestID,
			"latency":    latency,
		}).Info("Handled request")
	}
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(401, gin.H{"error": message})
}
