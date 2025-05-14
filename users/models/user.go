package models

import (
	"time"

	"github.com/google/uuid"
)

// ------------------------------------------------------------

type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type User struct {
	ID        string    `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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

type SafeUser struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToSafeUser(u User) SafeUser {
	return SafeUser{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type UpdateUserInput struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=5"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UserInfo struct {
	ID       int    `json:"id" db:"id"`
	UserID   string `json:"user_id" db:"user_id"`
	FullName string `json:"full_name" db:"full_name"`
	Age      int    `json:"age" db:"age"`
	City     string `json:"city" db:"city"`
}

type UserInfoResponse struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
	City     string `json:"city"`
}

type UpdateUserInfoInput struct {
	FullName *string `json:"full_name"`
	Age      *int    `json:"age"`
	City     *string `json:"city"`
}

type DifficultyStatEntry struct {
	Difficulty       string `json:"difficulty" db:"difficulty"`
	TotalSolved      int    `json:"total_solved" db:"total_solved"`
	TotalTimeSeconds int    `json:"total_time_seconds" db:"total_time_seconds"`
	BestTimeSeconds  *int   `json:"best_time_seconds,omitempty" db:"best_time_seconds"`
}

type MeResponse struct {
	User       SafeUser              `json:"user"`
	Info       *UserInfo             `json:"info,omitempty"`
	Statistics []DifficultyStatEntry `json:"statistics,omitempty"`
}
