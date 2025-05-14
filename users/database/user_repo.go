package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"users/models"
)

func (d *Database) GetUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	const query = `SELECT id, username, email, created_at, updated_at FROM users ORDER BY created_at DESC`

	if err := d.DB.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	return users, nil
}

func (d *Database) GetUser(ctx context.Context, id string) (*models.User, error) {
	const query = `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := d.DB.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (d *Database) CreateUser(ctx context.Context, user *models.User) error {
	const query = `
		INSERT INTO users (id, username, email, password, created_at, updated_at)
		VALUES (:id, :username, :email, :password, :created_at, :updated_at)
	`

	_, err := d.DB.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (d *Database) UpdateUser(ctx context.Context, user *models.User) error {
	const query = `
		UPDATE users
		SET username = :username,
		    email = :email,
		    password = :password,
		    updated_at = :updated_at
		WHERE id = :id
	`

	_, err := d.DB.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (d *Database) DeleteUser(ctx context.Context, id string) (bool, error) {
	const query = `DELETE FROM users WHERE id = $1`

	result, err := d.DB.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}

	return rows > 0, nil
}
