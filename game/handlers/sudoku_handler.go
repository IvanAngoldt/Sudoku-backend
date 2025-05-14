package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"game/config"
	"game/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *GameHandler) GetSudokuByDifficulty(c *gin.Context) {
	difficulty := c.Query("difficulty")
	if difficulty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty is required"})
		return
	}

	ctx := c.Request.Context()
	field, err := h.db.GetRandomByComplexity(ctx, difficulty)
	if err != nil {
		h.logger.Errorf("failed to get sudoku: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "нет доступных судоку"})
		return
	}

	var avgMs int64
	if field.SolvesSuccessful > 0 {
		avgMs = field.SolvesTotalTime / field.SolvesSuccessful
	}

	var successRate float64
	if field.SolveAttempts > 0 {
		successRate = float64(field.SolvesSuccessful) * 100.0 / float64(field.SolveAttempts)
	}

	resp := models.SudokuResponse{
		ID:               field.ID,
		InitialField:     field.InitialField,
		Solution:         field.Solution,
		Complexity:       field.Complexity,
		CreatedAt:        field.CreatedAt.Format(time.RFC3339),
		SolveAttempts:    field.SolveAttempts,
		SolvesSuccessful: field.SolvesSuccessful,
		AvgSolveTimeMs:   avgMs,
		SuccessRate:      successRate,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *GameHandler) GetAllSudokuByDifficulty(c *gin.Context) {
	difficulty := c.Query("difficulty")
	if difficulty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty is required"})
		return
	}

	fields, err := h.db.GetFieldsByComplexity(c.Request.Context(), difficulty)
	if err != nil {
		h.logger.Errorf("failed to get sudokus: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sudokus by difficulty"})
		return
	}

	responses := make([]models.SudokuResponse, 0, len(fields))
	for _, f := range fields {
		avg := int64(0)
		if f.SolvesSuccessful > 0 {
			avg = f.SolvesTotalTime / f.SolvesSuccessful
		}

		rate := 0.0
		if f.SolveAttempts > 0 {
			rate = float64(f.SolvesSuccessful) * 100 / float64(f.SolveAttempts)
		}

		responses = append(responses, models.SudokuResponse{
			ID:               f.ID,
			InitialField:     f.InitialField,
			Solution:         f.Solution,
			Complexity:       f.Complexity,
			CreatedAt:        f.CreatedAt.Format(time.RFC3339),
			SolveAttempts:    f.SolveAttempts,
			SolvesSuccessful: f.SolvesSuccessful,
			AvgSolveTimeMs:   avg,
			SuccessRate:      rate,
		})
	}

	c.JSON(http.StatusOK, responses)
}

func (h *GameHandler) GetSudokuByID(c *gin.Context) {
	id := c.Param("id")

	ctx := c.Request.Context()
	field, err := h.db.GetSudokuByID(ctx, id)
	if err != nil {
		h.logger.WithField("sudoku_id", id).Errorf("failed to get sudoku: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var avgSolveTimeMs int64
	if field.SolvesSuccessful > 0 {
		avgSolveTimeMs = field.SolvesTotalTime / field.SolvesSuccessful
	}

	var successRate float64
	if field.SolveAttempts > 0 {
		successRate = float64(field.SolvesSuccessful) * 100.0 / float64(field.SolveAttempts)
	}

	resp := models.SudokuResponse{
		ID:               field.ID,
		InitialField:     field.InitialField,
		Solution:         field.Solution,
		Complexity:       field.Complexity,
		CreatedAt:        field.CreatedAt.Format(time.RFC3339),
		SolveAttempts:    field.SolveAttempts,
		SolvesSuccessful: field.SolvesSuccessful,
		AvgSolveTimeMs:   avgSolveTimeMs,
		SuccessRate:      successRate,
	}

	h.logger.WithField("sudoku_id", field.ID).Info("sudoku fetched")
	c.JSON(http.StatusOK, resp)
}

func (h *GameHandler) ReportSolved(c *gin.Context) {
	id := c.Param("id")
	cfg := c.MustGet("config").(*config.Config)

	if id == "" {
		h.logger.Warn("missing sudoku id in path")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req models.SudokuSolvedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("invalid sudoku solve report: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.SolveTimeMs <= 0 {
		h.logger.Warn("solve_time_ms must be > 0")
		c.JSON(http.StatusBadRequest, gin.H{"error": "solve_time_ms must be > 0"})
		return
	}

	ctx := c.Request.Context()

	difficulty, err := h.db.GetSudokuDifficultyByID(ctx, id)
	if err != nil {
		h.logger.Errorf("failed to get difficulty for sudoku %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get difficulty"})
		return
	}

	// 1. Отмечаем, что судоку решено
	if err := h.db.MarkSudokuSolved(ctx, id, req.SolveTimeMs); err != nil {
		h.logger.Errorf("failed to mark sudoku as solved: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark sudoku as solved"})
		return
	}

	// 2. Отправляем обновление статистики
	updateReq := models.UpdateStatsRequest{
		UserID:      req.UserID,
		Difficulty:  difficulty,
		TimeSeconds: req.SolveTimeMs / 1000,
	}
	body, err := json.Marshal(updateReq)
	if err != nil {
		h.logger.Errorf("failed to marshal update stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal update request"})
		return
	}

	patchURL := fmt.Sprintf("%s/%s/statistics", cfg.UsersURL, req.UserID)
	patchReq, err := http.NewRequest(http.MethodPatch, patchURL, bytes.NewBuffer(body))
	if err != nil {
		h.logger.Errorf("failed to create PATCH request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create update request"})
		return
	}
	patchReq.Header.Set("Content-Type", "application/json")

	patchResp, err := http.DefaultClient.Do(patchReq)
	if err != nil {
		h.logger.Errorf("failed to send update stats request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to contact users service"})
		return
	}
	defer patchResp.Body.Close()

	if patchResp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		_ = json.NewDecoder(patchResp.Body).Decode(&errResp)
		h.logger.Warnf("users service update failed: %v", errResp.Error)
		c.JSON(patchResp.StatusCode, gin.H{"error": "users service update failed"})
		return
	}

	// 3. Получаем обновлённую статистику
	statsURL := fmt.Sprintf("%s/%s/statistics", cfg.UsersURL, req.UserID)
	statsResp, err := http.Get(statsURL)
	if err != nil {
		h.logger.Errorf("failed to fetch updated stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated statistics"})
		return
	}
	defer statsResp.Body.Close()

	if statsResp.StatusCode != http.StatusOK {
		h.logger.Warnf("users service GET stats failed: status %d", statsResp.StatusCode)
		c.JSON(statsResp.StatusCode, gin.H{"error": "failed to fetch statistics"})
		return
	}

	var stats models.UserStatisticsResponse

	// 4. Получаем все достижения из базы
	achievements, err := h.db.GetAllAchievements(ctx)
	if err != nil {
		h.logger.Errorf("failed to load achievements: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load achievements"})
		return
	}

	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		h.logger.Errorf("failed to decode user statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid statistics response"})
		return
	}

	// 5. Проверяем какие достижения выполнены
	qualified := EvaluateAchievements(stats, achievements)

	// 6. Получаем уже выданные пользователю достижения
	existingCodes, err := h.db.GetUserAchievements(ctx, req.UserID)
	if err != nil {
		h.logger.Errorf("failed to fetch user achievements: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user achievements"})
		return
	}
	existingSet := make(map[string]struct{}, len(existingCodes))
	for _, ach := range existingCodes {
		existingSet[ach.Code] = struct{}{}
	}

	// 7. Отфильтровываем уже выданные
	var newAchievements []models.Achievement
	for _, ach := range qualified {
		if _, exists := existingSet[ach.Code]; !exists {
			newAchievements = append(newAchievements, ach)
		}
	}

	// 8. Сохраняем новые достижения
	var codes []string
	for _, a := range newAchievements {
		codes = append(codes, a.Code)
	}

	if err := h.db.AssignAchievements(ctx, req.UserID, codes); err != nil {
		h.logger.Errorf("failed to assign achievements: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to assign achievements"})
		return
	}

	h.logger.Infof("sudoku %s solved by user %s, stats updated and fetched", id, req.UserID)

	c.JSON(http.StatusOK, gin.H{
		"status":            "solved and stats updated",
		"stats":             stats,
		"qualified_rewards": newAchievements,
	})
}
