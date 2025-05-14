package main

import (
	"auth/config"
	"auth/handlers"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logger := logrus.New()

	// Инициализация обработчика
	authHandler := handlers.NewAuthHandler(cfg, logger)

	// Инициализация роутера
	router := gin.Default()

	// Маршруты
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// Запуск
	logger.Infof("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
