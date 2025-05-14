package main

import (
	"game/config"
	"game/database"
	"game/handlers"
	"game/middleware"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	_ = os.MkdirAll("uploads/achievements", os.ModePerm)

	db, err := database.NewDatabase(cfg)
	if err != nil {
		logrus.Fatalf("failed to init database: %v", err)
	}

	// Инициализация обработчиков
	gameHandler := handlers.NewGameHandler(db, logger)

	// Настройка роутера
	router := gin.Default()

	router.Use(middleware.ExtractUserIDHeader(cfg))

	// Sudoku
	router.GET("/sudoku", gameHandler.GetSudokuByDifficulty)
	router.GET("/sudoku/all", gameHandler.GetAllSudokuByDifficulty)
	router.GET("/sudoku/:id", gameHandler.GetSudokuByID)
	router.POST("/sudoku/:id/solved", gameHandler.ReportSolved)

	// Achievements
	router.GET("/achievements", gameHandler.GetAllAchievements)
	router.GET("/achievements/:code", gameHandler.GetAchievementByCode)
	router.POST("/achievements", gameHandler.CreateAchievement)
	router.PATCH("/achievements/:code", gameHandler.UpdateAchievement)
	router.DELETE("/achievements/:code", gameHandler.DeleteAchievement)
	router.GET("/achievements/:code/icon", gameHandler.GetAchievementIcon)

	// User Achievements
	router.GET("/:id/achievements", gameHandler.GetUserAchievements)
	router.POST("/:id/achievements", gameHandler.AssignAchievement)
	router.DELETE("/:id/achievements/:code", gameHandler.DeleteUserAchievement)

	// Auto Achievements
	// router.POST("/:id/achievements/check", gameHandler.CheckAndAssignAchievements)

	// Запуск
	logger.Infof("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}
