package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"auth/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("invalid register request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	if msg, ok := h.validateRegisterRequest(&req); !ok {
		h.logger.Warnf("validation failed: %s", msg)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: msg})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Errorf("failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to hash password"})
		return
	}

	userReq := models.UserCreateRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	body, err := json.Marshal(userReq)
	if err != nil {
		h.logger.Errorf("failed to marshal user request: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to marshal user request"})
		return
	}

	resp, err := http.Post(h.cfg.UsersURL+"/", "application/json", bytes.NewBuffer(body))
	if err != nil {
		h.logger.Errorf("failed to contact users service: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to contact users service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errorResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			h.logger.Errorf("user service error (unreadable response): %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "User service returned unexpected error"})
			return
		}
		h.logger.Warnf("user service responded with error: %s", errorResp.Error)
		c.JSON(resp.StatusCode, errorResp)
		return
	}

	var createdUser models.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
		h.logger.Errorf("failed to parse user creation response: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to parse user creation response"})
		return
	}

	if createdUser.ID == "" {
		h.logger.Error("user service returned empty ID")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Invalid user creation response"})
		return
	}

	token, err := GenerateJWT(createdUser.ID, h.cfg.JWTSecret)
	if err != nil {
		h.logger.Errorf("failed to generate token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	h.logger.Infof("user registered successfully: id=%s, email=%s", createdUser.ID, createdUser.Email)
	c.JSON(http.StatusCreated, models.AuthResponse{Token: token})
}
