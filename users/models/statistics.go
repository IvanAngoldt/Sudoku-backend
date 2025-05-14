package models

type UserStatisticsResponse struct {
	UserID     string                `json:"user_id"`
	Statistics []DifficultyStatEntry `json:"statistics"`
}

type UpdateStatsRequest struct {
	UserID      string `json:"user_id"`
	Difficulty  string `json:"difficulty"`
	TimeSeconds int    `json:"time_seconds"`
}
