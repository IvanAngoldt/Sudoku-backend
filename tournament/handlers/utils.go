package handlers

import (
	"context"
	"fmt"
	"tournament/database"
	"tournament/models"

	"github.com/sirupsen/logrus"
)

type TournamentHandler struct {
	db     *database.Database
	logger *logrus.Logger
}

func NewTournamentHandler(db *database.Database, logger *logrus.Logger) *TournamentHandler {
	return &TournamentHandler{db: db, logger: logger}
}

func prepareTournamentResults(ctx context.Context, db *database.Database, tournamentID string) ([]models.TournamentResult, error) {
	const selectQuery = `
		SELECT user_id, username, score, solved_count
		FROM tournament_participants
		WHERE tournament_id = $1
		ORDER BY score DESC, solved_count DESC, joined_at ASC
	`

	type row struct {
		UserID      string `db:"user_id"`
		Username    string `db:"username"`
		Score       int    `db:"score"`
		SolvedCount int    `db:"solved_count"`
	}

	rows := []row{}
	if err := db.DB.SelectContext(ctx, &rows, selectQuery, tournamentID); err != nil {
		return nil, fmt.Errorf("fetch participants: %w", err)
	}

	results := make([]models.TournamentResult, len(rows))
	for i, r := range rows {
		results[i] = models.TournamentResult{
			TournamentID: tournamentID,
			UserID:       r.UserID,
			Username:     r.Username,
			Score:        r.Score,
			Rank:         i + 1,
			SolvedCount:  r.SolvedCount,
		}
	}

	return results, nil
}

func calculateScore(difficulty string, solveTimeSeconds int64) int {
	var base int
	var threshold int64

	switch difficulty {
	case "easy":
		base = 10
		threshold = 60
	case "medium":
		base = 20
		threshold = 120
	case "hard":
		base = 30
		threshold = 180
	case "very_hard":
		base = 40
		threshold = 300
	case "insane":
		base = 50
		threshold = 420
	case "inhuman":
		base = 60
		threshold = 600
	default:
		base = 0
		threshold = 1
	}

	// бонус за быстроту — максимум +100% к баллу
	bonusRatio := float64(threshold-solveTimeSeconds) / float64(threshold)
	if bonusRatio < 0 {
		bonusRatio = 0
	}
	final := float64(base) * (1.0 + bonusRatio)
	return int(final)
}
