package handlers

import (
	"game/database"
	"game/models"

	"github.com/sirupsen/logrus"
)

type GameHandler struct {
	db     *database.Database
	logger *logrus.Logger
}

func NewGameHandler(db *database.Database, logger *logrus.Logger) *GameHandler {
	return &GameHandler{db: db, logger: logger}
}

type ConditionFunc func(stats models.UserStatisticsResponse, cond models.AchievementCondition) bool

var conditionFuncs = map[string]ConditionFunc{
	"solved_count": func(stats models.UserStatisticsResponse, cond models.AchievementCondition) bool {
		for _, s := range stats.Statistics {
			if s.Difficulty == cond.Difficulty && s.TotalSolved >= cond.Count {
				return true
			}
		}
		return false
	},
	"best_time": func(stats models.UserStatisticsResponse, cond models.AchievementCondition) bool {
		for _, s := range stats.Statistics {
			if s.Difficulty == cond.Difficulty && s.BestTimeSeconds != nil && *s.BestTimeSeconds <= cond.MaxSeconds {
				return true
			}
		}
		return false
	},
	"multi_difficulty_solved": func(stats models.UserStatisticsResponse, cond models.AchievementCondition) bool {
		count := 0
		for _, level := range cond.Levels {
			for _, s := range stats.Statistics {
				if s.Difficulty == level && s.TotalSolved > 0 {
					count++
					break
				}
			}
		}
		return count == len(cond.Levels)
	},
}

func EvaluateAchievements(stats models.UserStatisticsResponse, achievements []models.Achievement) []models.Achievement {
	var result []models.Achievement
	for _, ach := range achievements {
		checkFunc, ok := conditionFuncs[ach.Condition.Type]
		if !ok {
			continue // неизвестный тип
		}
		if checkFunc(stats, ach.Condition) {
			result = append(result, ach)
		}
	}
	return result
}
