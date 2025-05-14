package models

import (
	"time"

	"github.com/google/uuid"
)

type TournamentStatus string

const (
	TournamentStatusPending   TournamentStatus = "upcoming"
	TournamentStatusActive    TournamentStatus = "active"
	TournamentStatusFinished  TournamentStatus = "finished"
	TournamentStatusCancelled TournamentStatus = "cancelled"
)

type Tournament struct {
	ID          string           `json:"id" db:"id"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	StartTime   time.Time        `json:"start_time" db:"start_time"`
	EndTime     time.Time        `json:"end_time" db:"end_time"`
	Status      TournamentStatus `json:"status" db:"status"`
	CreatedBy   string           `json:"created_by" db:"created_by"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
}

type CreateTournamentRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required"`
}

func NewTournament(name, description string, startTime, endTime time.Time,
	createdBy string) *Tournament {
	return &Tournament{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      TournamentStatusPending,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}
}

type UpdateTournamentInput struct {
	Name        *string           `json:"name"`
	Description *string           `json:"description"`
	StartTime   *time.Time        `json:"start_time"`
	EndTime     *time.Time        `json:"end_time"`
	Status      *TournamentStatus `json:"status"`
	CreatedBy   *string           `json:"created_by"`
}

type TournamentResult struct {
	TournamentID string `db:"tournament_id" json:"tournament_id"`
	UserID       string `db:"user_id" json:"user_id"`
	Username     string `db:"username" json:"username"`
	Score        int    `db:"score" json:"score"`
	Rank         int    `db:"rank" json:"rank"`
	SolvedCount  int    `db:"solved_count" json:"solved_count"`
}

type DashboardParticipant struct {
	UserID      string    `db:"user_id" json:"user_id"`
	Username    string    `db:"username" json:"username"`
	Score       int       `db:"score" json:"score"`
	SolvedCount int       `db:"solved_count" json:"solved_count"`
	JoinedAt    time.Time `db:"joined_at" json:"joined_at"`
	Rank        int       `json:"rank"`
}
