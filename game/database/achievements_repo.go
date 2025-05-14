package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"game/models"

	_ "github.com/lib/pq"
)

func (d *Database) GetAllAchievements(ctx context.Context) ([]models.Achievement, error) {
	rows, err := d.DB.QueryContext(ctx, `
		SELECT id, code, title, description, icon_url, condition, created_at
		FROM achievements
	`)
	if err != nil {
		return nil, fmt.Errorf("get achievements: %w", err)
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var rawCondition []byte
		var a models.Achievement
		if err := rows.Scan(
			&a.ID,
			&a.Code,
			&a.Title,
			&a.Description,
			&a.IconURL,
			&rawCondition,
			&a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		if err := json.Unmarshal(rawCondition, &a.Condition); err != nil {
			return nil, fmt.Errorf("unmarshal condition: %w", err)
		}
		achievements = append(achievements, a)
	}
	return achievements, nil
}

func (d *Database) GetAchievementByCode(code string) (models.Achievement, error) {
	var (
		a            models.Achievement
		rawCondition []byte
	)
	err := d.DB.QueryRowContext(context.Background(), `
		SELECT id, code, title, description, icon_url, condition, created_at
		FROM achievements
		WHERE code = $1
	`, code).Scan(
		&a.ID,
		&a.Code,
		&a.Title,
		&a.Description,
		&a.IconURL,
		&rawCondition,
		&a.CreatedAt,
	)
	if err != nil {
		return models.Achievement{}, fmt.Errorf("get achievement: %w", err)
	}

	if err := json.Unmarshal(rawCondition, &a.Condition); err != nil {
		return models.Achievement{}, fmt.Errorf("unmarshal condition: %w", err)
	}

	return a, nil
}

func (d *Database) InsertAchievement(ctx context.Context, a models.Achievement) error {
	conditionJSON, err := json.Marshal(a.Condition)
	if err != nil {
		return fmt.Errorf("marshal condition: %w", err)
	}

	_, err = d.DB.ExecContext(ctx, `
		INSERT INTO achievements (code, title, description, icon_url, condition)
		VALUES ($1, $2, $3, $4, $5)
	`, a.Code, a.Title, a.Description, a.IconURL, conditionJSON)
	if err != nil {
		return fmt.Errorf("insert achievement: %w", err)
	}

	return nil
}

func (d *Database) UpdateAchievement(ctx context.Context, code, title, description, icon, condition string) error {
	query := `
		UPDATE achievements
		SET title = COALESCE(NULLIF($2, ''), title),
		    description = COALESCE(NULLIF($3, ''), description),
		    icon_url = CASE WHEN $4 != '' THEN $4 ELSE icon_url END,
		    condition = CASE WHEN $5 != '' THEN $5::jsonb ELSE condition END
		WHERE code = $1
	`
	_, err := d.DB.ExecContext(ctx, query, code, title, description, icon, condition)
	if err != nil {
		return fmt.Errorf("update achievement: %w", err)
	}
	return nil
}

func (d *Database) GetAchievementIconFilename(code string) (string, error) {
	const query = `SELECT icon_url FROM achievements WHERE code = $1`

	var filename string
	err := d.DB.QueryRow(query, code).Scan(&filename)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get icon filename: %w", err)
	}
	return filename, nil
}

func (d *Database) DeleteAchievement(code string) error {
	const query = `DELETE FROM achievements WHERE code = $1`
	_, err := d.DB.Exec(query, code)
	if err != nil {
		return fmt.Errorf("delete achievement: %w", err)
	}
	return nil
}
