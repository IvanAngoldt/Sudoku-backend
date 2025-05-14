package handlers

import (
	"encoding/json"
	"game/models"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *GameHandler) GetAllAchievements(c *gin.Context) {
	achievements, err := h.db.GetAllAchievements(c.Request.Context())
	if err != nil {
		h.logger.Errorf("failed to fetch achievements: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch achievements"})
		return
	}
	c.JSON(http.StatusOK, achievements)
}

func (h *GameHandler) GetAchievementByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	achievement, err := h.db.GetAchievementByCode(code)
	if err != nil {
		h.logger.Errorf("failed to fetch achievement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch achievement"})
		return
	}
	c.JSON(http.StatusOK, achievement)
}

func (h *GameHandler) CreateAchievement(c *gin.Context) {
	var form models.CreateAchievementForm
	if err := c.ShouldBind(&form); err != nil {
		h.logger.Warnf("invalid achievement form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	// Парсим JSON-условие
	var cond models.AchievementCondition
	if err := json.Unmarshal([]byte(form.Condition), &cond); err != nil {
		h.logger.Warnf("invalid condition json: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid condition JSON"})
		return
	}

	// Загружаем иконку
	file, header, err := c.Request.FormFile("icon")
	if err != nil {
		h.logger.Warn("icon file is missing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Icon file is required"})
		return
	}
	defer file.Close()

	_ = os.MkdirAll("uploads/achievements", os.ModePerm)
	ext := path.Ext(header.Filename)
	filename := uuid.New().String() + ext
	filePath := path.Join("uploads/achievements", filename)

	out, err := os.Create(filePath)
	if err != nil {
		h.logger.Errorf("failed to create file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		h.logger.Errorf("failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Собираем структуру
	achievement := models.Achievement{
		Code:        form.Code,
		Title:       form.Title,
		Description: form.Description,
		IconURL:     filename,
		Condition:   cond,
	}

	if err := h.db.InsertAchievement(c.Request.Context(), achievement); err != nil {
		h.logger.Errorf("failed to insert achievement: %v", err)

		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Achievement with this code already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert achievement"})
		return
	}

	h.logger.Infof("achievement created: %s", form.Code)
	c.JSON(http.StatusCreated, gin.H{
		"code":        achievement.Code,
		"title":       achievement.Title,
		"description": achievement.Description,
		"icon_url":    achievement.IconURL,
	})
}

func (h *GameHandler) UpdateAchievement(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		h.logger.Warn("missing achievement code in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}

	var form models.UpdateAchievementForm
	if err := c.ShouldBind(&form); err != nil {
		h.logger.Warnf("invalid update form: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form fields"})
		return
	}

	// Получаем старую иконку
	oldIcon, err := h.db.GetAchievementIconFilename(code)
	if err != nil {
		h.logger.Errorf("failed to get old icon filename: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current icon"})
		return
	}

	var filename string
	file, header, err := c.Request.FormFile("icon")
	if err == nil {
		defer file.Close()
		_ = os.MkdirAll("uploads/achievements", os.ModePerm)
		ext := path.Ext(header.Filename)
		filename = uuid.New().String() + ext
		filePath := path.Join("uploads/achievements", filename)

		out, err := os.Create(filePath)
		if err != nil {
			h.logger.Errorf("failed to create file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			h.logger.Errorf("failed to save file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
	}

	// Обновляем в БД, включая условие
	err = h.db.UpdateAchievement(c.Request.Context(), code, form.Title, form.Description, filename, form.Condition)
	if err != nil {
		h.logger.Errorf("failed to update achievement: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update achievement"})
		return
	}

	// Удаляем старую иконку, если загружена новая
	if filename != "" && oldIcon != "" && oldIcon != filename {
		oldPath := path.Join("uploads/achievements", oldIcon)
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			h.logger.Warnf("failed to remove old icon: %v", err)
		}
	}

	h.logger.Infof("achievement updated: %s", code)
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *GameHandler) DeleteAchievement(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	filename, err := h.db.GetAchievementIconFilename(code)
	if err != nil {
		h.logger.Errorf("failed to get icon filename for %s: %v", code, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get icon filename"})
		return
	}

	if err := h.db.DeleteAchievement(code); err != nil {
		h.logger.Errorf("failed to delete achievement %s: %v", code, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete achievement"})
		return
	}

	if filename != "" {
		filePath := path.Join("uploads/achievements", filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			h.logger.Warnf("failed to delete icon file %s: %v", filePath, err)
		}
	}

	h.logger.Infof("achievement %s deleted", code)
	c.Status(http.StatusNoContent)
}

func (h *GameHandler) GetAchievementIcon(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing achievement code"})
		return
	}

	filename, err := h.db.GetAchievementIconFilename(code)
	if err != nil {
		h.logger.Errorf("failed to get icon filename: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get achievement"})
		return
	}
	if filename == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "achievement not found"})
		return
	}

	filePath := path.Join("uploads/achievements", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Warnf("icon file not found on disk: %s", filePath)
		c.JSON(http.StatusNotFound, gin.H{"error": "achievement icon not found"})
		return
	}

	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.File(filePath)
}
