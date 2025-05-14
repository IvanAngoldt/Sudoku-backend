package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"
	"users/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (h *UserHandler) GetUsers(c *gin.Context) {
	ctx := c.Request.Context()

	users, err := h.db.GetUsers(ctx)
	if err != nil {
		h.logger.Errorf("failed to get users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	safeUsers := make([]models.SafeUser, 0, len(users))
	for _, user := range users {
		safeUsers = append(safeUsers, models.ToSafeUser(user))
	}

	c.JSON(http.StatusOK, safeUsers)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	user, err := h.db.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			h.logger.Errorf("failed to get user by id: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, models.ToSafeUser(*user))
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("invalid user create request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx := c.Request.Context()

	existingUser, err := h.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Errorf("failed to check email: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	newUser := models.NewUser(req.Username, req.Email, req.Password)

	if err := h.db.CreateUser(ctx, newUser); err != nil {
		h.logger.Errorf("failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	if err := h.db.CreateUserInfo(ctx, newUser.ID); err != nil {
		h.logger.Errorf("failed to create user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init user info"})
		return
	}

	if err := h.db.InitDefaultDifficultyStats(ctx, newUser.ID); err != nil {
		h.logger.Errorf("failed to create stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init user stats"})
		return
	}

	c.JSON(http.StatusCreated, models.ToSafeUser(*newUser))
}

func (h *UserHandler) PatchUser(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists || userID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to edit this user"})
		return
	}

	var input models.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warnf("invalid patch input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Отказаться от пустого PATCH
	if input.Username == nil && input.Email == nil && input.Password == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty update payload"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.db.GetUser(ctx, id)
	if err != nil || user == nil {
		h.logger.Errorf("failed to get user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			h.logger.Errorf("failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = string(hashed)
	}
	user.UpdatedAt = time.Now()

	if err := h.db.UpdateUser(ctx, user); err != nil {
		h.logger.Errorf("failed to update user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, models.ToSafeUser(*user))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists || userID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this user"})
		return
	}

	ctx := c.Request.Context()

	ok, err := h.db.DeleteUser(ctx, id)
	if err != nil {
		h.logger.Errorf("failed to delete user %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
