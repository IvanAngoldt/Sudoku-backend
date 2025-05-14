package database

import (
	"context"
	"fmt"

	"tournament/config"

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
		`CREATE TABLE IF NOT EXISTS tournaments (
			id VARCHAR(36) PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tournament_participants (
			tournament_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			username VARCHAR(36) NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			solved_count INTEGER NOT NULL DEFAULT 0,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_solved_at TIMESTAMP,
			PRIMARY KEY (tournament_id, user_id),
			FOREIGN KEY (tournament_id) REFERENCES tournaments(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS tournament_results (
			tournament_id VARCHAR(36) NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			username VARCHAR(36) NOT NULL,
			score INTEGER NOT NULL,
			rank INTEGER NOT NULL,
			solved_count INTEGER NOT NULL,
			finished_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tournament_id, user_id),
			UNIQUE (tournament_id, rank)
		)`,
		`CREATE TABLE IF NOT EXISTS solved_sudokus (
			id VARCHAR(36) PRIMARY KEY,
			tournament_id VARCHAR(36) NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			sudoku_id VARCHAR(36) NOT NULL,
			solved_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tournament_id, user_id, sudoku_id)
		)`,
	}

	for _, q := range queries {
		if _, err := d.DB.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("failed to exec query: %v\nquery: %s", err, q)
		}
	}

	return nil
}
