package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"tournament/config"
	"tournament/database"
	"tournament/models"
	"tournament/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type TournamentHandler struct {
	db                *database.Database
	redisClient       *redis.Client
	cfg               *config.Config
	tournamentService *services.TournamentService
	gameService       *services.GameService
}

func NewTournamentHandler(db *database.Database, redisClient *redis.Client, cfg *config.Config, tournamentService *services.TournamentService, gameService *services.GameService) *TournamentHandler {
	return &TournamentHandler{
		db:                db,
		redisClient:       redisClient,
		cfg:               cfg,
		tournamentService: tournamentService,
		gameService:       gameService,
	}
}

func (h *TournamentHandler) CreateTournament(c *gin.Context) {
	var tournament models.Tournament
	if err := c.ShouldBindJSON(&tournament); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newTournament := models.NewTournament(tournament.Name, tournament.Description,
		tournament.StartTime, tournament.EndTime, tournament.Status, tournament.CreatedBy)

	if err := h.db.CreateTournament(newTournament); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament"})
		return
	}

	c.JSON(http.StatusCreated, newTournament)
}

func (h *TournamentHandler) GetTournaments(c *gin.Context) {
	tournaments, err := h.db.GetTournaments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournaments"})
		return
	}

	c.JSON(http.StatusOK, tournaments)
}

func (h *TournamentHandler) GetTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	tournament, err := h.db.GetTournament(tournamentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	c.JSON(http.StatusOK, tournament)
}

func (h *TournamentHandler) UpdateTournament(c *gin.Context) {
	id := c.Param("id")

	var input models.UpdateTournamentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.db.UpdateTournament(id, &input)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament updated successfully"})
}

func (h *TournamentHandler) DeleteTournament(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.DeleteTournament(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament deleted successfully"})
}

type RegisterParticipantRequest struct {
	UserID string `json:"user_id"`
}

func (h *TournamentHandler) RegisterParticipant(c *gin.Context) {
	tournamentID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	if err := h.tournamentService.RegisterParticipant(tournamentID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered for tournament"})
}

func (h *TournamentHandler) DeleteParticipant(c *gin.Context) {
	tournamentID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	if err := h.tournamentService.DeleteParticipant(tournamentID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted participant"})
}

func (h *TournamentHandler) StartTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	if err := h.tournamentService.StartTournament(tournamentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament started successfully"})
}

type UpdateProgressRequest struct {
	TournamentID string `json:"tournament_id"` // ID турнира
	FieldID      string `json:"field_id"`      // ID судоку из сервиса game
	SolveTimeMs  int64  `json:"solve_time_ms"` // Время решения в миллисекундах
}

func (h *TournamentHandler) UpdateProgress(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")

	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Error binding request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sudokuInfo, err := h.gameService.GetSudokuInfo(req.FieldID)

	if err != nil {
		log.Println("Error getting sudoku info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sudoku information"})
		return
	}

	progressUpdate := services.ProgressUpdate{
		SudokuID:    sudokuInfo.ID,
		Difficulty:  sudokuInfo.Complexity,
		SolveTimeMs: req.SolveTimeMs,
	}

	if err := h.tournamentService.UpdateParticipantProgress(req.TournamentID, userID, progressUpdate); err != nil {
		log.Println("Error updating progress:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress updated successfully"})
}

type GetUnsolvedSudokuRequest struct {
	TournamentID string `json:"tournament_id"`
}

func (h *TournamentHandler) GetUnsolvedSudoku(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	difficulty := c.Param("difficulty")

	var req GetUnsolvedSudokuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Error binding request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что турнир активен
	tournament, err := h.db.GetTournament(req.TournamentID)
	if err != nil {
		log.Println("Error getting tournament:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tournament information"})
		return
	}

	if tournament.Status != models.TournamentStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tournament is not active"})
		return
	}

	// Проверяем, что пользователь зарегистрирован в турнире
	// participant, err := h.db.GetParticipant(req.TournamentID, userID)
	// if err != nil {
	// 	log.Println("Error getting participant:", err)
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "user is not registered in this tournament"})
	// 	return
	// }

	// Получаем нерешенную судоку
	sudoku, err := h.gameService.GetUnsolvedSudokuByDifficulty(req.TournamentID, userID, difficulty, h.db)
	if err != nil {
		log.Println("Error getting unsolved sudoku:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sudoku)
}

func (h *TournamentHandler) GetSudoku(c *gin.Context) {
	difficulty := c.Param("difficulty")

	// userID := c.GetHeader("X-User-ID")

	if difficulty == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty is required"})
		return
	}

	sudoku, err := h.gameService.GetSudokuByDifficulty(difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sudoku"})
		return
	}

	c.JSON(http.StatusOK, sudoku)
}

func (h *TournamentHandler) FinishTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	if err := h.tournamentService.FinishTournament(tournamentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Подведём итоги турнира
	if err := h.tournamentService.UpdateTournamentStats(tournamentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament finished successfully"})
}

func (h *TournamentHandler) GetDashboard(c *gin.Context) {
	tournamentID := c.Param("id")
	userID := c.GetHeader("X-User-ID")

	dashboard, err := h.tournamentService.GetDashboard(tournamentID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if dashboard == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "dashboard not found"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
