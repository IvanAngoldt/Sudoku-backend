package handlers

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok {
		h.logger.Error("user_id is not a string")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id"})
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(path.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type"})
		return
	}

	if err := os.MkdirAll("uploads/avatars", os.ModePerm); err != nil {
		h.logger.Errorf("failed to create avatars dir: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload dir"})
		return
	}

	filename := fmt.Sprintf("%s%s", userID, ext)
	filePath := path.Join("uploads/avatars", filename)

	out, err := os.Create(filePath)
	if err != nil {
		h.logger.Errorf("failed to create file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		h.logger.Errorf("failed to copy file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	ctx := c.Request.Context()
	if err := h.db.UpsertUserAvatar(ctx, userID, filename); err != nil {
		h.logger.Errorf("failed to update avatar in db: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.File(filePath)
}

func (h *UserHandler) GetAvatar(c *gin.Context) {
	userID := c.Param("id") // из URL, как и ожидается

	ctx := c.Request.Context()
	filename, err := h.db.GetUserAvatarFilename(ctx, userID)
	if err != nil {
		h.logger.Errorf("failed to get avatar from db: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get avatar"})
		return
	}
	if filename == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar not found"})
		return
	}

	filePath := path.Join("uploads/avatars", filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Warnf("avatar file not found for user %s: %s", userID, filePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar file not found"})
		return
	}

	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.File(filePath)
}
