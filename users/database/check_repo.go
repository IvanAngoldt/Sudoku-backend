package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (d *Database) UsernameExists(ctx context.Context, username string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`

	var exists int
	err := d.DB.GetContext(ctx, &exists, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check username exists: %w", err)
	}
	return true, nil
}

func (d *Database) EmailExists(ctx context.Context, email string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE email = $1 LIMIT 1`

	var exists int
	err := d.DB.GetContext(ctx, &exists, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check email exists: %w", err)
	}
	return true, nil
}
