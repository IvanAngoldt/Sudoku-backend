package handlers

import (
	"game/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *GameHandler) GetUserAchievements(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}

	achievements, err := h.db.GetUserAchievements(c.Request.Context(), userID)
	if err != nil {
		h.logger.Errorf("failed to fetch achievements for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch achievements"})
		return
	}

	c.JSON(http.StatusOK, achievements)
}

func (h *GameHandler) AssignAchievement(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
		return
	}

	var req models.AssignAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("invalid assign request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.db.AssignAchievements(c.Request.Context(), userID, []string{req.Code})
	if err != nil {
		if err.Error() == "achievement not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "achievement not found"})
			return
		}
		h.logger.Errorf("failed to assign achievement to user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign achievement"})
		return
	}

	c.Status(http.StatusNoContent)
}

// handlers/achievement_handler.go

func (h *GameHandler) DeleteUserAchievement(c *gin.Context) {
	userID := c.Param("id")
	code := c.Param("code")

	if userID == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id or achievement code"})
		return
	}

	err := h.db.DeleteUserAchievement(c.Request.Context(), userID, code)
	if err != nil {
		switch err.Error() {
		case "achievement not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "achievement not found"})
		case "user did not have this achievement":
			c.JSON(http.StatusNotFound, gin.H{"error": "user does not have this achievement"})
		default:
			h.logger.Errorf("failed to delete achievement: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	h.logger.Infof("achievement %s removed from user %s", code, userID)
	c.Status(http.StatusNoContent)
}
