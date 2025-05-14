package main

import (
	"log"
	"tournament/config"
	"tournament/database"
	"tournament/handlers"
	"tournament/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	db, err := database.NewDatabase(cfg)
	if err != nil {
		logrus.Fatalf("failed to init database: %v", err)
	}

	// Инициализация обработчиков
	tournamentHandler := handlers.NewTournamentHandler(db, logger)

	// Настройка роутера
	router := gin.Default()

	router.Use(middleware.ExtractUserIDHeader(cfg))

	// Базовые CRUD операции
	router.GET("/", tournamentHandler.GetTournaments)
	router.GET("/:id", tournamentHandler.GetTournament)
	router.POST("/", tournamentHandler.CreateTournament)
	router.PATCH("/:id", tournamentHandler.PatchTournament)
	router.DELETE("/:id", tournamentHandler.DeleteTournament)

	router.GET("/current", tournamentHandler.GetCurrentTournament)

	// Операции с участниками турнира
	router.GET("/:id/participants", tournamentHandler.GetParticipants)
	router.POST("/:id/register", tournamentHandler.RegisterParticipant)
	router.DELETE("/:id/delete", tournamentHandler.DeleteParticipant)

	router.GET("/:id/dashboard", tournamentHandler.GetDashboard)
	router.GET("/:id/results", tournamentHandler.GetResults)

	router.POST("/:id/start", tournamentHandler.StartTournament)
	router.POST("/:id/finish", tournamentHandler.FinishTournament)

	router.POST("/sudoku", tournamentHandler.GetSudoku)
	router.POST("/sudoku/:id", tournamentHandler.GetSudokuByID)
	router.POST("/sudoku/:id/solved", tournamentHandler.ReportSolved)

	// Запуск сервера
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Error starting server:", err)
	}
}
