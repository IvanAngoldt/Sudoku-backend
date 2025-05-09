package middleware

import "github.com/gin-gonic/gin"

func ExtractUserIDHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", userID)
		}
		c.Next()
	}
}
