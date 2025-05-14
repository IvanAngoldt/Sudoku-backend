package handlers

import (
	"net/http"
	"strings"
	"users/models"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetMyUserInfo(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		h.logger.Errorf("user_id in context is not string: %v", userIDRaw)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id"})
		return
	}

	ctx := c.Request.Context()
	info, err := h.db.GetUserInfoByUserID(ctx, userID)
	if err != nil {
		h.logger.Errorf("failed to get user_info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User info not found"})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	var input models.UpdateUserInfoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if input.FullName == nil && input.Age == nil && input.City == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty update payload"})
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID := userIDRaw.(string)
	ctx := c.Request.Context()

	// валидация
	if input.FullName != nil && strings.TrimSpace(*input.FullName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Full name is required"})
		return
	}
	if input.City != nil && strings.TrimSpace(*input.City) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City is required"})
		return
	}
	if input.Age != nil && (*input.Age <= 0 || *input.Age > 120) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Age must be between 1 and 120"})
		return
	}

	if err := h.db.UpdateUserInfo(ctx, userID, input); err != nil {
		h.logger.Errorf("failed to update user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user info"})
		return
	}

	info, err := h.db.GetUserInfoByUserID(ctx, userID)
	if err != nil || info == nil {
		h.logger.Errorf("failed to fetch updated user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated info"})
		return
	}

	resp := models.UserInfoResponse{
		UserID:   info.UserID,
		FullName: info.FullName,
		Age:      info.Age,
		City:     info.City,
	}
	c.JSON(http.StatusOK, resp)
}
