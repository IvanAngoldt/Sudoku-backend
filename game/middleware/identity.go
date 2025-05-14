package middleware

import (
	"game/config"

	"github.com/gin-gonic/gin"
)

func ExtractUserIDHeader(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}

		c.Set("config", cfg)

		c.Next()
	}
}
