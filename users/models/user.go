package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(username, email, hashedPassword string) *User {
	return &User{
		ID:        uuid.New().String(),
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == password
}

type UpdateStatsRequest struct {
	UserID      string `json:"user_id"`
	Difficulty  string `json:"difficulty"`
	TimeSeconds int    `json:"time_seconds"`
}

type UserDifficultyStat struct {
	Difficulty       string `json:"difficulty"`
	TotalSolved      int    `json:"total_solved"`
	TotalTimeSeconds int    `json:"total_time_seconds"`
	BestTimeSeconds  *int   `json:"best_time_seconds,omitempty"`
}

type UserStatisticsResponse struct {
	UserID     string               `json:"user_id"`
	Statistics []UserDifficultyStat `json:"statistics"`
}
