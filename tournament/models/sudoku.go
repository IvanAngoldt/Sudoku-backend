package models

type GetSudokuRequest struct {
	UserID       string `json:"user_id"`
	TournamentID string `json:"tournament_id"`
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
	UserID       string `json:"user_id"`
	TournamentID string `json:"tournament_id"`
	Difficulty   string `json:"difficulty"`
	SolveTimeMs  int64  `json:"solve_time_ms"`
}

type SudokuSolvedResponse struct {
	Message string `json:"message"`
}
