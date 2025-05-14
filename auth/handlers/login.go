package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"auth/models"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("invalid login request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid request"})
		return
	}

	body, err := json.Marshal(req)
	if err != nil {
		h.logger.Errorf("failed to marshal login request: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to marshal request"})
		return
	}

	resp, err := http.Post(h.cfg.UsersURL+"/auth", "application/json", bytes.NewBuffer(body))
	if err != nil {
		h.logger.Errorf("failed to contact users service: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to contact users service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			h.logger.Warn("invalid credentials: unable to decode error response")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
			return
		}
		h.logger.Warnf("login failed: %s", errResp.Error)
		c.JSON(resp.StatusCode, errResp)
		return
	}

	var user models.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		h.logger.Errorf("failed to decode user response: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Invalid user response"})
		return
	}
	if user.ID == "" {
		h.logger.Error("empty user ID in response from users service")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "User ID is missing in response"})
		return
	}

	token, err := GenerateJWT(user.ID, h.cfg.JWTSecret)
	if err != nil {
		h.logger.Errorf("failed to generate JWT: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	h.logger.Infof("user %s logged in successfully", user.ID)
	c.JSON(http.StatusOK, models.AuthResponse{Token: token})
}
