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
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
	Status      TournamentStatus `json:"status"`
	CreatedBy   string           `json:"created_by"`
	CreatedAt   time.Time        `json:"created_at"`
}

func NewTournament(name, description string, startTime, endTime time.Time,
	status TournamentStatus, createdBy string) *Tournament {
	return &Tournament{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      status,
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
}

type TournamentParticipant struct {
	ID                 string    `json:"id" db:"id"`
	TournamentID       string    `json:"tournament_id" db:"tournament_id"`
	UserID             string    `json:"user_id" db:"user_id"`
	Score              int       `json:"score" db:"score"`
	SolvedCount        int       `json:"solved_count" db:"solved_count"`
	JoinedAt           time.Time `json:"joined_at" db:"joined_at"`
	LastSolvedAt       time.Time `json:"last_solved_at" db:"last_solved_at"`
	LastSolvedSudokuID string    `json:"last_solved_sudoku_id" db:"last_solved_sudoku_id"`
}

type TournamentResult struct {
	ID           string    `json:"id" db:"id"`
	TournamentID string    `json:"tournament_id" db:"tournament_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Score        int       `json:"score" db:"score"`
	Rank         int       `json:"rank" db:"rank"`
	SolvedCount  int       `json:"solved_count" db:"solved_count"`
	FinishedAt   time.Time `json:"finished_at" db:"finished_at"`
}

type TournamentDashboard struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	Status       string                  `json:"status"`
	Participants []TournamentParticipant `json:"participants"`
}
