package main

import (
	"log"
	"tournament/config"
	"tournament/database"
	"tournament/handlers"
	"tournament/middleware"
	"tournament/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Инициализация базы данных
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Инициализация Redis
	redisClient := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	defer redisClient.Close()

	// Инициализация сервисов
	tournamentService := services.NewTournamentService(db)

	// Инициализация сервисов
	gameService := services.NewGameService(cfg)

	// Инициализация обработчиков
	tournamentHandler := handlers.NewTournamentHandler(db, redisClient, cfg, tournamentService, gameService)

	// Настройка роутера
	router := gin.Default()

	router.Use(middleware.ExtractUserIDHeader())

	// Базовые CRUD операции
	router.POST("/", tournamentHandler.CreateTournament)
	router.GET("/", tournamentHandler.GetTournaments)
	router.GET("/:id", tournamentHandler.GetTournament)
	router.PUT("/:id", tournamentHandler.UpdateTournament)
	router.DELETE("/:id", tournamentHandler.DeleteTournament)

	// Операции с участниками турнира
	router.POST("/:id/register", tournamentHandler.RegisterParticipant)
	router.DELETE("/:id/delete", tournamentHandler.DeleteParticipant)

	router.POST("/:id/start", tournamentHandler.StartTournament)
	router.POST("/:id/finish", tournamentHandler.FinishTournament)

	router.GET("/sudoku/:difficulty", tournamentHandler.GetUnsolvedSudoku)
	router.POST("/sudoku/solved", tournamentHandler.UpdateProgress)

	router.GET("/dashboard", tournamentHandler.GetDashboard)

	// Запуск сервера
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
