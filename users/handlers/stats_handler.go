package handlers

import (
	"net/http"
	"strings"
	"users/models"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetUserStatistics(c *gin.Context) {
	userID := c.Param("id")
	if strings.TrimSpace(userID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}

	ctx := c.Request.Context()
	stats, err := h.db.GetUserStatistics(ctx, userID)
	if err != nil {
		h.logger.Errorf("failed to get stats for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, models.UserStatisticsResponse{
		UserID:     userID,
		Statistics: stats,
	})
}

func (h *UserHandler) UpdateUserStats(c *gin.Context) {
	var req models.UpdateStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.UserID == "" || req.Difficulty == "" || req.TimeSeconds <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	ctx := c.Request.Context()

	if err := h.db.UpdateDifficultyStats(ctx, req.UserID, req.Difficulty, req.TimeSeconds); err != nil {
		h.logger.Errorf("failed to update stats for %s (%s): %v", req.UserID, req.Difficulty, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update difficulty stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Statistics updated"})
}
