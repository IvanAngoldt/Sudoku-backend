package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"users/models"
)

func (d *Database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := d.DB.GetContext(ctx, &user, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}
