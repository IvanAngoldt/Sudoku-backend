package models

type UserAchievement struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	EarnedAt    string `json:"earned_at"`
}

type AssignAchievementRequest struct {
	Code string `json:"code" binding:"required"` // код достижения
}

type AchievementResponse struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
}
