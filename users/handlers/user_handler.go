package handlers

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"users/database"
	"users/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	db *database.Database
}

func NewUserHandler(db *database.Database) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) validateUpdateRequest(username, email, password *string, isPatch bool) (string, bool) {
	if username != nil {
		if len(*username) < 5 {
			return "Username must be at least 5 characters long", false
		}
		exists, err := h.db.UsernameExists(*username)
		if err != nil {
			return "Failed to check username uniqueness", false
		}
		if exists {
			return "Username already exists", false
		}
	} else if !isPatch {
		return "Username is required", false
	}

	if email != nil {
		if !isValidEmail(*email) {
			return "Invalid email format", false
		}
		exists, err := h.db.EmailExists(*email)
		if err != nil {
			return "Failed to check email uniqueness", false
		}
		if exists {
			return "Email already exists", false
		}
	} else if !isPatch {
		return "Email is required", false
	}

	if password != nil {
		if !isValidPassword(*password) {
			return "Password must be at least 6 characters long and contain at least one number", false
		}
	} else if !isPatch {
		return "Password is required", false
	}

	return "", true
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

func ToSafeUser(u *models.User) models.SafeUser {
	return models.SafeUser{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.db.GetUsers()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	safeUsers := make([]models.SafeUser, 0, len(users))
	for i := range users {
		safeUsers = append(safeUsers, ToSafeUser(users[i]))
	}
	c.JSON(200, safeUsers)
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDRaw.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.db.GetUser(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	safeUser := ToSafeUser(user)
	c.JSON(http.StatusOK, safeUser)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.db.GetUser(id)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	safeUser := ToSafeUser(user)
	c.JSON(200, safeUser)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли пользователь с таким email
	existingUser, err := h.db.GetUserByEmail(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email existence"})
		return
	}
	if existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
		return
	}

	// Создаем нового пользователя с уже хешированным паролем
	newUser := models.NewUser(user.Username, user.Email, user.Password)

	if err := h.db.CreateUser(newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Инициализируем статистику пользователя
	if _, err := h.db.EnsureUserStatistics(newUser.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user statistics"})
		return
	}

	newUser.Password = ""
	c.JSON(http.StatusCreated, newUser)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	userID, exists := c.Get("user_id")
	if !exists || userID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to edit this user"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// Проверка наличия всех обязательных полей
	if user.Username == "" || user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields (username, email, password) are required"})
		return
	}

	// Валидация содержимого
	msg, ok := h.validateUpdateRequest(&user.Username, &user.Email, &user.Password, false)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Хеширование пароля
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user.ID = id
	user.Password = string(hashed)
	user.UpdatedAt = user.CreatedAt

	if err := h.db.UpdateUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
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
		c.JSON(400, gin.H{"error": "invalid input"})
		return
	}

	user, err := h.db.GetUser(id)
	if err != nil || user == nil {
		c.JSON(404, gin.H{"error": "user not found"})
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
			c.JSON(500, gin.H{"error": "failed to hash password"})
			return
		}
		user.Password = string(hashed)
	}
	user.UpdatedAt = user.CreatedAt

	if err := h.db.UpdateUser(user); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	user.Password = ""
	c.JSON(200, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists || userID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this user"})
		return
	}

	if err := h.db.DeleteUser(id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(204)
}

func (h *UserHandler) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(400, gin.H{"error": "username is required"})
		return
	}
	exists, err := h.db.UsernameExists(username)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}
	if exists {
		c.Status(409)
	} else {
		c.Status(204)
	}
}

func (h *UserHandler) CheckEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(400, gin.H{"error": "email is required"})
		return
	}
	exists, err := h.db.EmailExists(email)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}
	if exists {
		c.Status(409)
	} else {
		c.Status(204)
	}
}

func (h *UserHandler) AuthUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.db.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(401, gin.H{"error": "Invalid email"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}

	c.JSON(200, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (h *UserHandler) CreateUserInfo(c *gin.Context) {
	var info models.UserInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Получаем user_id из токена и устанавливаем его в info
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)
	info.UserID = userID

	// Минимальная валидация
	if strings.TrimSpace(info.FullName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Full name is required"})
		return
	}
	if info.Age <= 0 || info.Age > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Age must be between 1 and 120"})
		return
	}
	if strings.TrimSpace(info.City) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City is required"})
		return
	}

	if err := h.db.CreateUserInfo(&info); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user info"})
		return
	}

	resp := models.UserInfoResponse{
		UserID:   info.UserID,
		FullName: info.FullName,
		Age:      info.Age,
		City:     info.City,
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) GetMyUserInfo(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	info, err := h.db.GetUserInfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User info not found"})
		return
	}

	resp := models.UserInfoResponse{
		UserID:   info.UserID,
		FullName: info.FullName,
		Age:      info.Age,
		City:     info.City,
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	var input models.UpdateUserInfoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Получаем user_id из контекста
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	// Загружаем текущую информацию
	currentInfo, err := h.db.GetUserInfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load current info"})
		return
	}
	if currentInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User info not found"})
		return
	}

	// Применяем только изменённые поля
	if input.FullName != nil {
		if strings.TrimSpace(*input.FullName) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Full name is required"})
			return
		}
		currentInfo.FullName = *input.FullName
	}
	if input.Age != nil {
		if *input.Age <= 0 || *input.Age > 120 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Age must be between 1 and 120"})
			return
		}
		currentInfo.Age = *input.Age
	}
	if input.City != nil {
		if strings.TrimSpace(*input.City) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "City is required"})
			return
		}
		currentInfo.City = *input.City
	}

	// Обновляем в базе
	if err := h.db.UpdateUserInfo(currentInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user info"})
		return
	}

	resp := models.UserInfoResponse{
		UserID:   currentInfo.UserID,
		FullName: currentInfo.FullName,
		Age:      currentInfo.Age,
		City:     currentInfo.City,
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *UserHandler) DeleteUserInfo(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	if err := h.db.DeleteUserInfo(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user info"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
		return
	}
	defer file.Close()

	_ = os.MkdirAll("uploads/avatars", os.ModePerm)

	ext := path.Ext(header.Filename)
	filename := userID + ext
	filePath := path.Join("uploads/avatars", filename)

	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	if err := h.db.UpsertUserAvatar(userID, filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar filename"})
		return
	}

	// Возвращаем саму картинку, как в GET /users/avatar
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.File(filePath)
}

func (h *UserHandler) GetAvatar(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDRaw.(string)

	filename, err := h.db.GetUserAvatarFilename(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get avatar"})
		return
	}
	if filename == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar not found"})
		return
	}

	filePath := path.Join("uploads/avatars", filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar file not found"})
		return
	}

	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.File(filePath)
}

func (h *UserHandler) UpdateUserStats(c *gin.Context) {
	var req models.UpdateStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.UserID == "" || req.Difficulty == "" || req.TimeSeconds <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid fields"})
		return
	}

	if _, err := h.db.EnsureUserStatistics(req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user statistics"})
		return
	}

	if err := h.db.UpdateDifficultyStats(req.UserID, req.Difficulty, req.TimeSeconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update difficulty stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Statistics updated"})
}

func (h *UserHandler) GetUserStatistics(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}

	stats, err := h.db.GetUserStatistics(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
		return
	}

	c.JSON(http.StatusOK, models.UserStatisticsResponse{
		UserID:     userID,
		Statistics: stats,
	})
}
