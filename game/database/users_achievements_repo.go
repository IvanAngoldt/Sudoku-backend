package database

import (
	"context"
	"database/sql"
	"fmt"
	"game/models"

	"github.com/lib/pq"
)

func (d *Database) GetUserAchievements(ctx context.Context, userID string) ([]models.SimpleAchievement, error) {
	rows, err := d.DB.QueryContext(ctx, `
		SELECT a.code, a.title, a.description
		FROM user_achievements ua
		JOIN achievements a ON a.id = ua.achievement_id
		WHERE ua.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("get user achievements: %w", err)
	}
	defer rows.Close()

	var achievements []models.SimpleAchievement
	for rows.Next() {
		var a models.SimpleAchievement
		if err := rows.Scan(&a.Code, &a.Title, &a.Description); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		achievements = append(achievements, a)
	}
	return achievements, nil
}

func (d *Database) AssignAchievements(ctx context.Context, userID string, codes []string) error {
	if len(codes) == 0 {
		return nil
	}

	// Получаем соответствующие ID достижений
	rows, err := d.DB.QueryContext(ctx, `
		SELECT id, code FROM achievements WHERE code = ANY($1)
	`, pq.Array(codes))
	if err != nil {
		return fmt.Errorf("fetch achievement IDs: %w", err)
	}
	defer rows.Close()

	achMap := make(map[string]int)
	for rows.Next() {
		var id int
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return fmt.Errorf("scan achievement ID: %w", err)
		}
		achMap[code] = id
	}

	// Готовим батч вставки
	tx, err := d.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO user_achievements (user_id, achievement_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, code := range codes {
		id, ok := achMap[code]
		if !ok {
			continue
		}
		if _, err := stmt.ExecContext(ctx, userID, id); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert user achievement: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit insert achievements: %w", err)
	}

	return nil
}

func (d *Database) DeleteUserAchievement(ctx context.Context, userID string, code string) error {
	const getQuery = `SELECT id FROM achievements WHERE code = $1`

	var achievementID int
	err := d.DB.QueryRowContext(ctx, getQuery, code).Scan(&achievementID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("achievement not found")
	}
	if err != nil {
		return fmt.Errorf("query achievement: %w", err)
	}

	const deleteQuery = `
		DELETE FROM user_achievements
		WHERE user_id = $1 AND achievement_id = $2
	`

	res, err := d.DB.ExecContext(ctx, deleteQuery, userID, achievementID)
	if err != nil {
		return fmt.Errorf("delete user_achievement: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("user did not have this achievement")
	}

	return nil
}
