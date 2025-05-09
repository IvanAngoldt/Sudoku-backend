package models

type Achievement struct {
	Code        string
	Title       string
	Description string
	IconURL     string
}

type CreateAchievementForm struct {
	Code        string `form:"code" binding:"required"`
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
}

type UserAchievementResponse struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AssignAchievementRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type AchievementResponse struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UserDifficultyStat struct {
	Difficulty       string `json:"difficulty"`
	TotalSolved      int    `json:"total_solved"`
	TotalTimeSeconds int    `json:"total_time_seconds"`
	BestTimeSeconds  int    `json:"best_time_seconds"`
}

type DeleteUserAchievementRequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}
