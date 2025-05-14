package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (d *Database) UpsertUserAvatar(ctx context.Context, userID string, avatarURL string) error {
	const query = `
		INSERT INTO user_avatar (user_id, avatar_url)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET avatar_url = EXCLUDED.avatar_url
	`

	_, err := d.DB.ExecContext(ctx, query, userID, avatarURL)
	if err != nil {
		return fmt.Errorf("upsert user_avatar: %w", err)
	}
	return nil
}

func (d *Database) GetUserAvatarFilename(ctx context.Context, userID string) (string, error) {
	const query = `SELECT avatar_url FROM user_avatar WHERE user_id = $1`

	var filename string
	err := d.DB.GetContext(ctx, &filename, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("get user_avatar: %w", err)
	}
	return filename, nil
}
