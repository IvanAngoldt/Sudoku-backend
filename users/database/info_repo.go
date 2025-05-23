package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"users/models"
)

func (d *Database) CreateUserInfo(ctx context.Context, userID string) error {
	const query = `
		INSERT INTO user_info (user_id, firstname, secondname, age, city)
		VALUES ($1, '', '', 0, '')
	`

	_, err := d.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("create user_info: %w", err)
	}
	return nil
}

func (d *Database) GetUserInfo(ctx context.Context, userID string) (*models.UserInfo, error) {
	const query = `
		SELECT firstname, secondname, age, city
		FROM user_info
		WHERE user_id = $1
	`

	var info models.UserInfo
	err := d.DB.GetContext(ctx, &info, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user info: %w", err)
	}
	return &info, nil
}

func (d *Database) GetUserInfoByUserID(ctx context.Context, userID string) (*models.UserInfo, error) {
	const query = `
		SELECT user_id, firstname, secondname, age, city
		FROM user_info
		WHERE user_id = $1
	`

	var info models.UserInfo
	err := d.DB.GetContext(ctx, &info, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user_info by user_id: %w", err)
	}

	return &info, nil
}

func (d *Database) UpdateUserInfo(ctx context.Context, userID string, input models.UpdateUserInfoInput) error {
	const query = `
		UPDATE user_info
		SET
			firstname = COALESCE(:firstname, firstname),
			secondname = COALESCE(:secondname, secondname),
			age = COALESCE(:age, age),
			city = COALESCE(:city, city)
		WHERE user_id = :user_id
	`

	type updatePayload struct {
		UserID     string  `db:"user_id"`
		FirstName  *string `db:"firstname"`
		SecondName *string `db:"secondname"`
		Age        *int    `db:"age"`
		City       *string `db:"city"`
	}

	payload := updatePayload{
		UserID:     userID,
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		Age:        input.Age,
		City:       input.City,
	}

	_, err := d.DB.NamedExecContext(ctx, query, payload)
	if err != nil {
		return fmt.Errorf("update user_info: %w", err)
	}
	return nil
}
