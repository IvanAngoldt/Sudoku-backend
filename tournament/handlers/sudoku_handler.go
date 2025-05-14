package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"tournament/config"
	"tournament/models"

	"github.com/gin-gonic/gin"
)

func (h *TournamentHandler) GetSudoku(c *gin.Context) {
	difficulty := c.Query("difficulty")

	if difficulty == "" {
		h.logger.Errorf("difficulty is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty is required"})
		return
	}

	var req models.GetSudokuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid get sudoku request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == "" || req.TournamentID == "" {
		h.logger.Warn("missing user_id or tournament_id in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and tournament_id are required"})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists || userID != req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to get sudoku for this tournament"})
		return
	}

	cfg := c.MustGet("config").(*config.Config)
	sudokuURL := fmt.Sprintf("%s/sudoku?difficulty=%s", cfg.GameServiceURL, difficulty)

	sudokuResp, err := http.Get(sudokuURL)
	if err != nil {
		h.logger.Errorf("failed to fetch sudoku: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sudoku"})
		return
	}
	defer sudokuResp.Body.Close()

	if sudokuResp.StatusCode != http.StatusOK {
		h.logger.Warnf("game service GET sudoku failed: status %d", sudokuResp.StatusCode)
		c.JSON(sudokuResp.StatusCode, gin.H{"error": "failed to fetch sudoku"})
		return
	}

	var sudoku models.SudokuResponse
	if err := json.NewDecoder(sudokuResp.Body).Decode(&sudoku); err != nil {
		h.logger.Errorf("failed to decode sudoku response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid sudoku response"})
		return
	}

	c.JSON(http.StatusOK, sudoku)
}

func (h *TournamentHandler) GetSudokuByID(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		h.logger.Errorf("id is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req models.GetSudokuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid get sudoku request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == "" || req.TournamentID == "" {
		h.logger.Warn("missing user_id or tournament_id in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and tournament_id are required"})
		return
	}

	userID, exists := c.Get("user_id")

	if !exists || userID != req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to get sudoku for this tournament"})
		return
	}

	cfg := c.MustGet("config").(*config.Config)
	sudokuURL := fmt.Sprintf("%s/sudoku/%s", cfg.GameServiceURL, id)

	sudokuResp, err := http.Get(sudokuURL)
	if err != nil {
		h.logger.Errorf("failed to fetch sudoku: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch sudoku"})
		return
	}
	defer sudokuResp.Body.Close()

	if sudokuResp.StatusCode != http.StatusOK {
		h.logger.Warnf("game service GET sudoku failed: status %d", sudokuResp.StatusCode)
		c.JSON(sudokuResp.StatusCode, gin.H{"error": "failed to fetch sudoku"})
		return
	}

	var sudoku models.SudokuResponse
	if err := json.NewDecoder(sudokuResp.Body).Decode(&sudoku); err != nil {
		h.logger.Errorf("failed to decode sudoku response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid sudoku response"})
		return
	}

	c.JSON(http.StatusOK, sudoku)
}

func (h *TournamentHandler) ReportSolved(c *gin.Context) {

	var req models.SudokuSolvedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid solved request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	score := calculateScore(req.Difficulty, req.SolveTimeMs/1000)

	ctx := c.Request.Context()
	now := time.Now()

	err := h.db.UpdateTournamentParticipant(ctx, req.TournamentID, req.UserID, now, score)
	if err != nil {
		h.logger.Errorf("failed to update participant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tournament stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "solved recorded"})
}
