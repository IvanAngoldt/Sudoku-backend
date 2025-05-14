package main

import (
	"os"
	"users/config"
	"users/database"
	"users/handlers"
	"users/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	// Создание папки под аватары
	_ = os.MkdirAll("uploads/avatars", os.ModePerm)

	db, err := database.NewDatabase(cfg)
	if err != nil {
		logrus.Fatalf("failed to init database: %v", err)
	}

	// Обработчики
	userHandler := handlers.NewUserHandler(db, logger)

	// Роутер
	router := gin.Default()

	router.Use(middleware.ExtractUserIDHeader())

	// Публичные маршруты
	router.GET("/check-username", userHandler.CheckUsername)
	router.GET("/check-email", userHandler.CheckEmail)
	router.POST("/", userHandler.CreateUser)
	router.POST("/auth", userHandler.AuthUser)

	// Защищённые маршруты
	router.GET("/", userHandler.GetUsers)
	router.GET("/:id", userHandler.GetUser)
	router.PATCH("/:id", userHandler.PatchUser)
	router.DELETE("/:id", userHandler.DeleteUser)

	router.GET("/me", userHandler.GetMe)
	router.GET("/me/info", userHandler.GetMyUserInfo)
	router.PATCH("/me/info", userHandler.UpdateUserInfo)

	router.POST("/me/avatar", userHandler.UploadAvatar)
	router.GET("/:id/avatar", userHandler.GetAvatar)

	router.GET("/:id/statistics", userHandler.GetUserStatistics)
	router.PATCH("/:id/statistics", userHandler.UpdateUserStats)

	// Запуск
	logger.Infof("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
