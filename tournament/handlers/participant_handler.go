package handlers

import (
	"net/http"
	"strings"
	"time"
	"tournament/models"

	"github.com/gin-gonic/gin"
)

func (h *TournamentHandler) GetParticipants(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	participants, err := h.db.GetParticipants(ctx, id)
	if err != nil {
		h.logger.Errorf("failed to get participants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get participants"})
		return
	}

	c.JSON(http.StatusOK, participants)
}

func (h *TournamentHandler) RegisterParticipant(c *gin.Context) {
	tournamentID := c.Param("id")

	var req models.RegisterParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid register participant request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == "" || req.Username == "" {
		h.logger.Warn("missing user_id or username in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and username are required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID != req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to register for this tournament"})
		return
	}

	ctx := c.Request.Context()

	participant := &models.TournamentParticipant{
		TournamentID: tournamentID,
		UserID:       req.UserID,
		Username:     req.Username,
		Score:        0,
		SolvedCount:  0,
		JoinedAt:     time.Now(),
		LastSolvedAt: nil,
	}

	if err := h.db.RegisterParticipant(ctx, participant); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			h.logger.Warnf("participant already registered: %v", err)
			c.JSON(http.StatusConflict, gin.H{"error": "Participant already registered"})
			return
		}
		h.logger.Errorf("failed to register participant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant registered successfully"})
}

func (h *TournamentHandler) DeleteParticipant(c *gin.Context) {
	tournamentID := c.Param("id")

	var req models.DeleteParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid delete participant request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == "" {
		h.logger.Warn("missing user_id in request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists || userID != req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this participant"})
		return
	}

	ctx := c.Request.Context()

	if err := h.db.DeleteParticipant(ctx, tournamentID, req.UserID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Participant or tournament not found"})
			return
		}
		h.logger.Errorf("failed to delete participant: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted participant"})
}
