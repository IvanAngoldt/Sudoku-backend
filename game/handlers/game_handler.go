package handlers

import (
	"encoding/json"
	"fmt"
	"game/config"
	"game/database"
	"game/models"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SudokuHandler struct {
	db  *database.Database
	cfg *config.Config
}

func NewGameHandler(db *database.Database, cfg *config.Config) *SudokuHandler {
	return &SudokuHandler{
		db:  db,
		cfg: cfg,
	}
}

func (h *SudokuHandler) GetRandomSudokuFromDB(c *gin.Context) {
	difficulty := c.Param("difficulty")

	if difficulty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty is required"})
		return
	}

	field, err := h.db.GetFieldByComplexity(difficulty)
	if err != nil {
		log.Printf("Ошибка при получении судоку: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "нет доступных судоку"})
		return
	}

	// вычисляем статистику
	var avgSolveTimeMs int64
	if field.SolvesSuccessful > 0 {
		avgSolveTimeMs = field.SolvesTotalTime / int64(field.SolvesSuccessful)
	}

	var successRate float64
	if field.SolveAttempts > 0 {
		successRate = float64(field.SolvesSuccessful) * 100.0 / float64(field.SolveAttempts)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                field.ID,
		"initial_field":     field.InitialField,
		"solution":          field.Solution,
		"complexity":        field.Complexity,
		"created_at":        field.CreatedAt,
		"solve_attempts":    field.SolveAttempts,
		"solves_successful": field.SolvesSuccessful,
		"avg_solve_time_ms": avgSolveTimeMs,
		"success_rate":      fmt.Sprintf("%.2f", successRate),
	})
}

func (h *SudokuHandler) GetSudokusByDifficulty(c *gin.Context) {
	difficulty := c.Param("difficulty")

	fields, err := h.db.GetFieldsByComplexity(difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sudokus by difficulty"})
		return
	}

	c.JSON(http.StatusOK, fields)
}

func (h *SudokuHandler) ReportSolved(c *gin.Context) {
	var req models.SudokuSolvedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.SolveTimeMs <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "solve_time_ms must be > 0"})
		return
	}

	err := h.db.MarkSudokuSolved(c.Request.Context(), req.ID, req.SolveTimeMs)
	if err != nil {
		log.Printf("Ошибка при обновлении статистики: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (h *SudokuHandler) GetSudokuByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	field, err := h.db.GetSudokuByID(c.Request.Context(), id)
	if err != nil {
		log.Printf("Ошибка при получении судоку по id: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// рассчитываем статистику
	var avgSolveTimeMs int64
	if field.SolvesSuccessful > 0 {
		avgSolveTimeMs = field.SolvesTotalTime / int64(field.SolvesSuccessful)
	}
	var successRate float64
	if field.SolveAttempts > 0 {
		successRate = float64(field.SolvesSuccessful) * 100.0 / float64(field.SolveAttempts)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                field.ID,
		"initial_field":     field.InitialField,
		"solution":          field.Solution,
		"complexity":        field.Complexity,
		"created_at":        field.CreatedAt,
		"solve_attempts":    field.SolveAttempts,
		"solves_successful": field.SolvesSuccessful,
		"avg_solve_time_ms": avgSolveTimeMs,
		"success_rate":      fmt.Sprintf("%.2f", successRate),
	})
}

func (h *SudokuHandler) CreateAchievement(c *gin.Context) {
	var form models.CreateAchievementForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	file, header, err := c.Request.FormFile("icon")
	if err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	iconURL := filename

	// Сохраняем в БД
	err = h.db.InsertAchievement(models.Achievement{
		Code:        form.Code,
		Title:       form.Title,
		Description: form.Description,
		IconURL:     iconURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert achievement"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":        form.Code,
		"title":       form.Title,
		"description": form.Description,
		"icon_url":    iconURL,
	})
}

func (h *SudokuHandler) GetAchievements(c *gin.Context) {
	achievements, err := h.db.GetAllAchievements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch achievements"})
		return
	}

	c.JSON(http.StatusOK, achievements)
}

func (h *SudokuHandler) GetAchievementIcon(c *gin.Context) {
	achievementCode := c.Param("code")
	if achievementCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing achievement_code"})
		return
	}

	achievement, err := h.db.GetAchievementIconFilename(achievementCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get achievement"})
		return
	}
	if achievement == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Achievement not found"})
		return
	}

	filePath := path.Join("uploads/achievements", achievement)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Achievement icon file not found"})
		return
	}

	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.File(filePath)
}

func (h *SudokuHandler) DeleteAchievement(c *gin.Context) {
	code := c.Param("code")

	// Получаем имя файла для удаления
	filename, err := h.db.GetAchievementIconFilename(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get icon filename"})
		return
	}

	// Удаляем запись из БД
	if err := h.db.DeleteAchievement(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete achievement"})
		return
	}

	// Удаляем файл с диска, если был
	if filename != "" {
		filePath := path.Join("uploads/achievements", filename)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			// Не критично, но логируем
			fmt.Println("Failed to delete icon file:", err)
		}
	}

	c.Status(http.StatusNoContent)
}

func (h *SudokuHandler) GetMyAchievements(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	achievements, err := h.db.GetUserAchievements(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch achievements"})
		return
	}

	c.JSON(http.StatusOK, achievements)
}

func (h *SudokuHandler) fetchUserStats(userID string) ([]models.UserDifficultyStat, error) {
	url := fmt.Sprintf("%s/statistics/%s", h.cfg.UsersURL, userID)
	log.Printf("fetchUserStats: making request to %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact users service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("users service returned %d: %s", resp.StatusCode, string(body))
	}

	// Обёртка для декодинга
	var wrapper struct {
		Statistics []models.UserDifficultyStat `json:"statistics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("failed to decode users stats: %w", err)
	}

	return wrapper.Statistics, nil
}

func (h *SudokuHandler) AssignAchievement(c *gin.Context) {
	var req models.AssignAchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	achievement, err := h.db.AssignAchievementByCode(req.UserID, req.Code)
	if err != nil {
		if err.Error() == "achievement not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Achievement not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign achievement"})
		return
	}

	c.JSON(http.StatusOK, achievement)
}

func (h *SudokuHandler) CheckAchievements(c *gin.Context) {
	// Получаем user_id из контекста (установлен middleware'ом авторизации)
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	// Получаем статистику пользователя из users-сервиса
	stats, err := h.fetchUserStats(userID)
	if err != nil {
		log.Printf("fetchUserStats error: %v", err) // добавь лог
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user statistics"})
		return
	}

	// Вызываем проверку и возможную выдачу новых достижений
	newAchievements, err := h.db.CheckAndAssignAchievements(userID, stats)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check achievements"})
		return
	}

	// Возвращаем список новых достижений (может быть пустым)
	c.JSON(http.StatusOK, newAchievements)
}

func (h *SudokuHandler) DeleteUserAchievement(c *gin.Context) {
	var req models.DeleteUserAchievementRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	if req.UserID == "" || req.Code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id or code"})
		return
	}

	err := h.db.DeleteUserAchievement(req.UserID, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete achievement"})
		return
	}

	c.Status(http.StatusNoContent)
}
