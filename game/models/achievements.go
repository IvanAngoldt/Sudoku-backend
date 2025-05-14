package models

import "time"

type Achievement struct {
	ID          int                  `json:"id"`
	Code        string               `json:"code"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	IconURL     string               `json:"icon_url"`
	Condition   AchievementCondition `json:"condition"`
	CreatedAt   time.Time            `json:"created_at"`
}

type AchievementCondition struct {
	Type       string   `json:"type"` // e.g. "solved_count", "best_time", ...
	Difficulty string   `json:"difficulty,omitempty"`
	Count      int      `json:"count,omitempty"`
	MaxSeconds int      `json:"max_seconds,omitempty"`
	Levels     []string `json:"levels,omitempty"`
}

type CreateAchievementForm struct {
	Code        string `form:"code" binding:"required"`
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	Condition   string `form:"condition" binding:"required"` // raw JSON-строка
}

type UpdateAchievementForm struct {
	Title       string `form:"title"`
	Description string `form:"description"`
	Condition   string `form:"condition"` // JSON-строка, опционально
}

type SimpleAchievement struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
