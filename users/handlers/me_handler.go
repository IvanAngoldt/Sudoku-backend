package handlers

import (
	"net/http"
	"users/models"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetMe(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		h.logger.Errorf("user_id in context is not a string: %v", userIDRaw)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	ctx := c.Request.Context()

	user, err := h.db.GetUser(ctx, userID)
	if err != nil {
		h.logger.Errorf("failed to get user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	info, err := h.db.GetUserInfo(ctx, userID)
	if err != nil {
		h.logger.Errorf("failed to get user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user info"})
		return
	}

	stats, err := h.db.GetUserStatistics(ctx, userID)
	if err != nil {
		h.logger.Errorf("failed to get user stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
		return
	}

	resp := models.MeResponse{
		User:       models.ToSafeUser(*user),
		Info:       info,
		Statistics: stats,
	}
	c.JSON(http.StatusOK, resp)
}
