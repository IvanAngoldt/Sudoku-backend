package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"auth/config"
	"auth/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	cfg *config.Config
}

func NewAuthHandler(cfg *config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if msg, ok := h.validateRegisterRequest(&req); !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: msg})
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
		return
	}

	// Создаем запрос для сервиса пользователей
	userRequest := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	userData, err := json.Marshal(userRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user request"})
		return
	}

	resp, err := http.Post(h.cfg.UsersURL+"/", "application/json", bytes.NewBuffer(userData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to connect to users service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create user"})
			return
		}
		c.JSON(resp.StatusCode, errorResp)
		return
	}

	// Декодируем ответ и получаем user ID
	var createdUser models.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse user creation response"})
		return
	}

	// Создаем JWT токен с user_id = UUID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": createdUser.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{Token: tokenString})
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

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Отправляем запрос на аутентификацию в сервис users
	userData, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to authenticate"})
		return
	}

	resp, err := http.Post(h.cfg.UsersURL+"/auth", "application/json", bytes.NewBuffer(userData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to connect to users service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
			return
		}
		c.JSON(resp.StatusCode, errorResp)
		return
	}

	// Декодируем ответ, чтобы получить user ID
	var user models.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse user data"})
		return
	}

	// Создаем JWT токен с user_id = UUID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{Token: tokenString})
}
