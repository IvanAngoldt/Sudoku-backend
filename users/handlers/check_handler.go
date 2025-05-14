package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if strings.TrimSpace(username) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	ctx := c.Request.Context()
	exists, err := h.db.UsernameExists(ctx, username)
	if err != nil {
		h.logger.Errorf("failed to check username: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if exists {
		c.Status(http.StatusConflict)
	} else {
		c.Status(http.StatusNoContent)
	}
}

func (h *UserHandler) CheckEmail(c *gin.Context) {
	email := c.Query("email")
	if strings.TrimSpace(email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	ctx := c.Request.Context()
	exists, err := h.db.EmailExists(ctx, email)
	if err != nil {
		h.logger.Errorf("failed to check email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if exists {
		c.Status(http.StatusConflict) // 409: занято
	} else {
		c.Status(http.StatusNoContent) // 204: свободно
	}
}
