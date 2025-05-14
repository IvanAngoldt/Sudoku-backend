package models

import "time"

type SudokuField struct {
	ID               string    `db:"id"`
	InitialField     string    `db:"initial_field"`
	Solution         string    `db:"solution"`
	Complexity       string    `db:"complexity"`
	CreatedAt        time.Time `db:"created_at"`
	SolveAttempts    int64     `db:"solve_attempts"`
	SolvesSuccessful int64     `db:"solves_successful"`
	SolvesTotalTime  int64     `db:"solves_total_time"`
}

type SudokuResponse struct {
	ID               string  `json:"id"`
	InitialField     string  `json:"initial_field"`
	Solution         string  `json:"solution"`
	Complexity       string  `json:"complexity"`
	CreatedAt        string  `json:"created_at"`
	SolveAttempts    int64   `json:"solve_attempts"`
	SolvesSuccessful int64   `json:"solves_successful"`
	AvgSolveTimeMs   int64   `json:"avg_solve_time_ms"`
	SuccessRate      float64 `json:"success_rate"`
}

type SudokuSolvedRequest struct {
	UserID      string `json:"user_id"`
	SolveTimeMs int64  `json:"solve_time_ms"`
}

type UpdateStatsRequest struct {
	UserID      string `json:"user_id"`
	Difficulty  string `json:"difficulty"`
	TimeSeconds int64  `json:"time_seconds"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UserStatisticsResponse struct {
	UserID     string                `json:"user_id"`
	Statistics []DifficultyStatEntry `json:"statistics"`
}

type DifficultyStatEntry struct {
	Difficulty       string `json:"difficulty"`
	TotalSolved      int    `json:"total_solved"`
	TotalTimeSeconds int    `json:"total_time_seconds"`
	BestTimeSeconds  *int   `json:"best_time_seconds,omitempty"`
}
