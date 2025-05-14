package database

import (
	"context"
	"fmt"
	"users/models"
)

func (d *Database) InitDefaultDifficultyStats(ctx context.Context, userID string) error {
	skills := []string{"easy", "medium", "hard", "very_hard", "insane", "inhuman"}

	tx, err := d.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO user_difficulty_stats (user_id, difficulty)
		VALUES ($1, $2)
		ON CONFLICT (user_id, difficulty) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, skill := range skills {
		if _, err := stmt.ExecContext(ctx, userID, skill); err != nil {
			return fmt.Errorf("insert difficulty %s: %w", skill, err)
		}
	}

	return tx.Commit()
}

func (d *Database) GetUserStatistics(ctx context.Context, userID string) ([]models.DifficultyStatEntry, error) {
	const query = `
		SELECT difficulty, total_solved, total_time_seconds, best_time_seconds
		FROM user_difficulty_stats
		WHERE user_id = $1
		ORDER BY difficulty
	`

	var stats []models.DifficultyStatEntry
	if err := d.DB.SelectContext(ctx, &stats, query, userID); err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}
	return stats, nil
}

func (d *Database) UpdateDifficultyStats(ctx context.Context, userID, difficulty string, timeSeconds int) error {
	const query = `
		UPDATE user_difficulty_stats
		SET
			total_solved = total_solved + 1,
			total_time_seconds = total_time_seconds + $1,
			best_time_seconds = CASE
				WHEN best_time_seconds IS NULL THEN $1
				WHEN $1 < best_time_seconds THEN $1
				ELSE best_time_seconds
			END
		WHERE user_id = $2 AND difficulty = $3
	`

	_, err := d.DB.ExecContext(ctx, query, timeSeconds, userID, difficulty)
	if err != nil {
		return fmt.Errorf("update difficulty stats: %w", err)
	}
	return nil
}
