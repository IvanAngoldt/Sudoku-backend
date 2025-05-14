package handlers

import (
	"net/http"
	"strings"
	"time"

	"auth/config"
	"auth/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	cfg    *config.Config
	logger *logrus.Logger
}

func NewAuthHandler(cfg *config.Config, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{cfg: cfg, logger: logger}
}

func isValidEmail(email string) bool {
	// Проверяем наличие @ и точки
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return false
	}

	// Проверяем, что email не начинается и не заканчивается на @ или точку
	if strings.HasPrefix(email, "@") || strings.HasPrefix(email, ".") ||
		strings.HasSuffix(email, "@") || strings.HasSuffix(email, ".") {
		return false
	}

	return true
}

func isValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	// Проверяем наличие хотя бы одной цифры
	hasNumber := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasNumber = true
			break
		}
	}

	return hasNumber
}

func (h *AuthHandler) validateRegisterRequest(req *models.RegisterRequest) (string, bool) {
	if len(req.Username) < 5 {
		return "Username must be at least 5 characters long", false
	}
	if !isValidEmail(req.Email) {
		return "Invalid email format", false
	}
	if !isValidPassword(req.Password) {
		return "Password must be at least 6 characters long and contain at least one number", false
	}

	// Проверка уникальности никнейма
	resp, err := http.Get(h.cfg.UsersURL + "/check-username?username=" + req.Username)
	if err != nil {
		return "Failed to check username uniqueness", false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return "Username already exists", false
	}

	// Проверка уникальности email
	resp, err = http.Get(h.cfg.UsersURL + "/check-email?email=" + req.Email)
	if err != nil {
		return "Failed to check email uniqueness", false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return "Email already exists", false
	}

	return "", true
}

func GenerateJWT(userID string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}
