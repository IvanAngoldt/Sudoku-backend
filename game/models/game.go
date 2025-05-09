package models

import "time"

type SudokuField struct {
	ID               string
	InitialField     string
	Solution         string
	Complexity       string
	CreatedAt        time.Time
	SolveAttempts    int
	SolvesSuccessful int
	SolvesTotalTime  int64
}

type SudokuSolvedRequest struct {
	ID          string `json:"id" binding:"required"`
	SolveTimeMs int64  `json:"solve_time_ms" binding:"required"`
}
