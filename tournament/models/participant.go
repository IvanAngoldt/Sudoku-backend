package models

import "time"

type TournamentParticipant struct {
	TournamentID string     `json:"tournament_id" db:"tournament_id"`
	UserID       string     `json:"user_id" db:"user_id"`
	Username     string     `json:"username" db:"username"`
	Score        int        `json:"score" db:"score"`
	SolvedCount  int        `json:"solved_count" db:"solved_count"`
	JoinedAt     time.Time  `json:"joined_at" db:"joined_at"`
	LastSolvedAt *time.Time `json:"last_solved_at" db:"last_solved_at"`
}

type RegisterParticipantRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type DeleteParticipantRequest struct {
	UserID string `json:"user_id"`
}
