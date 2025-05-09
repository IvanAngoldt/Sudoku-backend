package main

import (
	"log"
	"os"
	"users/config"
	"users/database"
	"users/handlers"
	"users/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	os.MkdirAll("uploads/avatars", os.ModePerm)
	os.MkdirAll("uploads/achievements", os.ModePerm)

	// Инициализация базы данных
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(db)

	// Настройка роутера
	router := gin.Default()

	router.Use(middleware.ExtractUserIDHeader())

	// Маршруты API
	router.GET("/", userHandler.GetUsers)
	router.GET("/me", userHandler.GetMe)
	router.GET(":id", userHandler.GetUser)
	router.POST("/", userHandler.CreateUser)
	router.PATCH("/:id", userHandler.PatchUser)
	router.PUT("/:id", userHandler.UpdateUser)
	router.DELETE("/:id", userHandler.DeleteUser)
	router.GET("/check-username", userHandler.CheckUsername)
	router.GET("/check-email", userHandler.CheckEmail)
	router.POST("/auth", userHandler.AuthUser)

	router.POST("/info", userHandler.CreateUserInfo)
	router.GET("/info", userHandler.GetMyUserInfo)
	router.PUT("/info", userHandler.UpdateUserInfo)
	router.DELETE("/info", userHandler.DeleteUserInfo)

	router.POST("/avatar", userHandler.UploadAvatar)
	router.GET("/avatar", userHandler.GetAvatar)

	router.GET("/statistics/:user_id", userHandler.GetUserStatistics)
	router.POST("/statistics/update", userHandler.UpdateUserStats)

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.ServerPort)

	router.Run(":" + cfg.ServerPort)
}
