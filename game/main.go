package main

import (
	"game/config"
	"game/database"
	"game/handlers"
	"game/middleware"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	os.MkdirAll("uploads/achievements", os.ModePerm)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Инициализация базы данных
	db, err := database.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Инициализация обработчиков
	gameHandler := handlers.NewGameHandler(db, cfg)

	// Настройка роутера
	router := gin.Default()

	router.Use(middleware.ExtractUserIDHeader())
	router.Use(gin.Logger())
	router.Use(gin.RecoveryWithWriter(os.Stderr))

	// Маршруты API
	router.GET("/get-random-sudoku/:difficulty", gameHandler.GetRandomSudokuFromDB)
	router.GET("/get-sudokus-by-difficulty/:difficulty", gameHandler.GetSudokusByDifficulty)

	router.POST("/sudoku/solved", gameHandler.ReportSolved)
	router.GET("/sudoku/:id", gameHandler.GetSudokuByID)

	router.GET("/achievements", gameHandler.GetAchievements)
	router.POST("/achievements", gameHandler.CreateAchievement)
	router.DELETE("/achievements/:code", gameHandler.DeleteAchievement)
	router.GET("/achievements/:code/icon", gameHandler.GetAchievementIcon)
	router.GET("/my-achievements", gameHandler.GetMyAchievements)
	router.POST("/assign-achievement", gameHandler.AssignAchievement)
	router.POST("/check-achievements", gameHandler.CheckAchievements)

	router.DELETE("/achievements", gameHandler.DeleteUserAchievement)

	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.ServerPort)

	router.Run(":" + cfg.ServerPort)
}
