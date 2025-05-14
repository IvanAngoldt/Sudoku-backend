package database

import (
	"context"
	"fmt"
	"game/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	DB *sqlx.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	database := &Database{DB: db}

	if err := database.createTables(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return database, nil
}

func (d *Database) createTables(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sudoku_fields (
			id VARCHAR(36) PRIMARY KEY,
			initial_field VARCHAR(81) NOT NULL,
			solution VARCHAR(81) NOT NULL,
			complexity TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

			solve_attempts INTEGER NOT NULL DEFAULT 0,
			solves_successful INTEGER NOT NULL DEFAULT 0,
			solves_total_time BIGINT NOT NULL DEFAULT 0,

			UNIQUE (initial_field, solution)
		)`,
		`CREATE TABLE IF NOT EXISTS sudoku_tags (
			id VARCHAR(36) PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sudoku_field_tags (
			field_id VARCHAR(36) NOT NULL REFERENCES sudoku_fields(id) ON DELETE CASCADE,
			tag_id VARCHAR(36) NOT NULL REFERENCES sudoku_tags(id) ON DELETE CASCADE,
			PRIMARY KEY (field_id, tag_id)
		)`,
		`CREATE TABLE IF NOT EXISTS achievements (
			id SERIAL PRIMARY KEY,
			code TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			icon_url TEXT,
			condition JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_achievements (
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			achievement_id INTEGER NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
			earned_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (user_id, achievement_id)
		)`,
	}

	for _, q := range queries {
		if _, err := d.DB.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("failed to exec query: %v\nquery: %s", err, q)
		}
	}

	return nil
}
