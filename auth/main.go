package main

import (
	"auth/config"
	"auth/handlers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Println("AUTH JWT_SECRET:", os.Getenv("JWT_SECRET"))

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(cfg)

	// Настройка роутера
	router := gin.Default()

	// Маршруты API
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// Запуск сервера
	log.Printf("Auth server starting on port %s", cfg.ServerPort)
	router.Run(":" + cfg.ServerPort)
}
