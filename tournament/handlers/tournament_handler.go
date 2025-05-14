package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"tournament/models"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func (h *TournamentHandler) GetTournaments(c *gin.Context) {
	ctx := c.Request.Context()
	tournaments, err := h.db.GetTournaments(ctx)
	if err != nil {
		h.logger.Errorf("failed to get tournaments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tournaments"})
		return
	}

	c.JSON(http.StatusOK, tournaments)
}

func (h *TournamentHandler) GetTournament(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	tournament, err := h.db.GetTournament(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		} else {
			h.logger.Errorf("failed to get tournament by id: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, tournament)
}

func (h *TournamentHandler) CreateTournament(c *gin.Context) {
	var req models.CreateTournamentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid tournament create request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx := c.Request.Context()

	newTournament := models.NewTournament(
		req.Name,
		req.Description,
		req.StartTime,
		req.EndTime,
		userID.(string),
	)

	if err := h.db.CreateTournament(ctx, newTournament); err != nil {
		h.logger.Errorf("failed to create tournament: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament"})
		return
	}

	c.JSON(http.StatusCreated, newTournament)
}

func (h *TournamentHandler) PatchTournament(c *gin.Context) {
	id := c.Param("id")

	var input models.UpdateTournamentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name == nil && input.Description == nil && input.StartTime == nil &&
		input.EndTime == nil && input.Status == nil && input.CreatedBy == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty update payload"})
		return
	}

	ctx := c.Request.Context()
	tournament, err := h.db.GetTournament(ctx, id)
	if err != nil || tournament == nil {
		h.logger.Errorf("failed to get tournament by id: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if input.Name != nil {
		tournament.Name = *input.Name
	}
	if input.Description != nil {
		tournament.Description = *input.Description
	}
	if input.StartTime != nil {
		tournament.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		tournament.EndTime = *input.EndTime
	}
	if input.Status != nil {
		tournament.Status = *input.Status
	}
	if input.CreatedBy != nil {
		tournament.CreatedBy = *input.CreatedBy
	}

	if err := h.db.UpdateTournament(ctx, tournament); err != nil {
		h.logger.Errorf("failed to update tournament: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tournament"})
		return
	}

	c.JSON(http.StatusOK, tournament)
}

func (h *TournamentHandler) DeleteTournament(c *gin.Context) {
	tournamentID := c.Param("id")

	ctx := c.Request.Context()

	ok, err := h.db.DeleteTournament(ctx, tournamentID)
	if err != nil {
		h.logger.Errorf("failed to delete tournament: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if !ok {
		h.logger.Errorf("failed to delete tournament: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament deleted successfully"})
}

func (h *TournamentHandler) GetCurrentTournament(c *gin.Context) {
	ctx := c.Request.Context()

	tournament, err := h.db.GetLatestTournament(ctx)
	if err != nil {
		h.logger.Errorf("failed to fetch latest tournament: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch latest tournament"})
		return
	}

	if tournament == nil {
		h.logger.Infof("no latest tournament found")
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, tournament)
}

func (h *TournamentHandler) StartTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	ctx := c.Request.Context()

	ok, err := h.db.UpdateTournamentStatus(ctx, tournamentID, models.TournamentStatusActive)
	if err != nil {
		if strings.Contains(err.Error(), "cannot update finished tournament") {
			h.logger.Warnf("attempt to start finished tournament: %s", tournamentID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя запустить завершённый турнир"})
			return
		}

		h.logger.Errorf("failed to start tournament: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Турнир успешно запущен"})
}

func (h *TournamentHandler) FinishTournament(c *gin.Context) {
	tournamentID := c.Param("id")
	ctx := c.Request.Context()

	// Обновим статус турнира (можно вне транзакции)
	ok, err := h.db.UpdateTournamentStatus(ctx, tournamentID, models.TournamentStatusFinished)
	if err != nil {
		h.logger.Errorf("failed to finish tournament: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	// Подготовим результаты отдельно
	results, err := prepareTournamentResults(ctx, h.db, tournamentID)
	if err != nil {
		h.logger.Errorf("failed to prepare results: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	// Гарантированно сохраняем и чистим в одной транзакции
	if err := h.db.WithTx(ctx, func(tx *sqlx.Tx) error {
		if err := h.db.SaveTournamentResultsTx(ctx, tx, results); err != nil {
			return err
		}
		if err := h.db.DeleteTournamentParticipantsTx(ctx, tx, tournamentID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		h.logger.Errorf("transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tournament finished successfully"})
}

func (h *TournamentHandler) GetDashboard(c *gin.Context) {
	tournamentID := c.Param("id")

	ctx := c.Request.Context()

	participants, err := h.db.GetTournamentDashboard(ctx, tournamentID)
	if err != nil {
		h.logger.Errorf("failed to get tournament dashboard: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, participants)
}

func (h *TournamentHandler) GetResults(c *gin.Context) {
	tournamentID := c.Param("id")

	ctx := c.Request.Context()

	results, err := h.db.GetTournamentResults(ctx, tournamentID)
	if err != nil {
		h.logger.Errorf("failed to get tournament results: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, results)
}
