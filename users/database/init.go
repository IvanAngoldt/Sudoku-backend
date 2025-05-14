package database

import (
	"context"
	"fmt"
	"users/config"

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

	// Проверка соединения
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
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_info (
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			full_name VARCHAR(100) NOT NULL DEFAULT '',
			age INT NOT NULL DEFAULT 0,
			city VARCHAR(50) NOT NULL DEFAULT '',
			PRIMARY KEY (user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_difficulty_stats (
			user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			difficulty TEXT NOT NULL,
			total_solved INTEGER NOT NULL DEFAULT 0,
			total_time_seconds INTEGER NOT NULL DEFAULT 0,
			best_time_seconds INTEGER,
			PRIMARY KEY (user_id, difficulty)
		)`,
		`CREATE TABLE IF NOT EXISTS user_avatar (
			user_id VARCHAR(36) PRIMARY KEY,
			avatar_url TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		if _, err := d.DB.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("failed to exec query: %v\nquery: %s", err, q)
		}
	}

	return nil
}
