package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"tournament/models"

	"github.com/jmoiron/sqlx"
)

func (d *Database) GetTournaments(ctx context.Context) ([]models.Tournament, error) {
	var tournaments []models.Tournament
	const query = `
		SELECT id, name, description, start_time, end_time, status, 
		created_by, created_at FROM tournaments ORDER BY created_at DESC
	`

	if err := d.DB.SelectContext(ctx, &tournaments, query); err != nil {
		return nil, fmt.Errorf("get tournaments: %w", err)
	}
	return tournaments, nil
}

func (d *Database) GetTournament(ctx context.Context, id string) (*models.Tournament, error) {
	const query = `
		SELECT id, name, description, start_time, end_time, status, 
		created_by, created_at FROM tournaments WHERE id = $1
	`

	var tournament models.Tournament
	err := d.DB.GetContext(ctx, &tournament, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("get tournament by id: %w", err)
	}
	return &tournament, nil
}

func (d *Database) CreateTournament(ctx context.Context, tournament *models.Tournament) error {
	const query = `
		INSERT INTO tournaments (
			id, name, description, start_time, end_time, status, created_by, created_at
		) VALUES (
			:id, :name, :description, :start_time, :end_time, :status, :created_by, :created_at
		)
	`

	_, err := d.DB.NamedExecContext(ctx, query, tournament)
	if err != nil {
		return fmt.Errorf("create tournament: %w", err)
	}
	return nil
}

func (d *Database) UpdateTournament(ctx context.Context, tournament *models.Tournament) error {
	const query = `
		UPDATE tournaments
		SET name = :name,
		    description = :description,
		    start_time = :start_time,
		    end_time = :end_time,
		    status = :status
			created_by = :created_by
		WHERE id = :id
	`

	_, err := d.DB.NamedExecContext(ctx, query, tournament)
	if err != nil {
		return fmt.Errorf("update tournament: %w", err)
	}
	return nil
}

func (d *Database) UpdateTournamentStatus(ctx context.Context, id string, newStatus models.TournamentStatus) (bool, error) {
	// 1. Проверка текущего статуса
	var currentStatus models.TournamentStatus
	err := d.DB.GetContext(ctx, &currentStatus, "SELECT status FROM tournaments WHERE id = $1", id)
	if err != nil {
		return false, fmt.Errorf("get current status: %w", err)
	}

	if currentStatus == models.TournamentStatusFinished {
		return false, fmt.Errorf("cannot update finished tournament")
	}

	// 2. Обновление
	const query = `
		UPDATE tournaments
		SET status = $1
		WHERE id = $2
	`

	result, err := d.DB.ExecContext(ctx, query, newStatus, id)
	if err != nil {
		return false, fmt.Errorf("update tournament status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (d *Database) DeleteTournament(ctx context.Context, id string) (bool, error) {
	const query = `
	    DELETE FROM tournaments WHERE id = $1
	`

	result, err := d.DB.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("delete tournament: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("rows affected: %w", err)
	}

	return rows > 0, nil
}

func (d *Database) GetLatestTournament(ctx context.Context) (*models.Tournament, error) {
	const query = `
		SELECT id, name, description, start_time, end_time, status, created_by, created_at
		FROM tournaments
		ORDER BY created_at DESC
		LIMIT 1
	`

	var tournament models.Tournament
	err := d.DB.GetContext(ctx, &tournament, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest tournament: %w", err)
	}

	return &tournament, nil
}

func (d *Database) GetTournamentDashboard(ctx context.Context, tournamentID string) ([]models.DashboardParticipant, error) {
	const query = `
		SELECT user_id, username, score, solved_count, joined_at
		FROM tournament_participants
		WHERE tournament_id = $1
		ORDER BY score DESC, solved_count DESC, joined_at ASC
	`

	type row struct {
		UserID      string    `db:"user_id"`
		Username    string    `db:"username"`
		Score       int       `db:"score"`
		SolvedCount int       `db:"solved_count"`
		JoinedAt    time.Time `db:"joined_at"`
	}

	var rows []row
	if err := d.DB.SelectContext(ctx, &rows, query, tournamentID); err != nil {
		return nil, fmt.Errorf("select dashboard participants: %w", err)
	}

	result := make([]models.DashboardParticipant, 0, len(rows))
	for i, r := range rows {
		result = append(result, models.DashboardParticipant{
			UserID:      r.UserID,
			Username:    r.Username,
			Score:       r.Score,
			SolvedCount: r.SolvedCount,
			JoinedAt:    r.JoinedAt,
			Rank:        i + 1,
		})
	}

	return result, nil
}

func (d *Database) GetTournamentResults(ctx context.Context, tournamentID string) ([]models.TournamentResult, error) {
	const query = `
		SELECT tournament_id, user_id, username, score, rank, solved_count
		FROM tournament_results
		WHERE tournament_id = $1
		ORDER BY rank ASC
	`

	var results []models.TournamentResult
	if err := d.DB.SelectContext(ctx, &results, query, tournamentID); err != nil {
		return nil, fmt.Errorf("select tournament results: %w", err)
	}

	return results, nil
}

func (d *Database) WithTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := d.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *Database) SaveTournamentResultsTx(ctx context.Context, tx *sqlx.Tx, results []models.TournamentResult) error {
	const insertQuery = `
		INSERT INTO tournament_results (
			tournament_id, user_id, username, score, rank, solved_count
		) VALUES (
			:tournament_id, :user_id, :username, :score, :rank, :solved_count
		)
	`

	for _, res := range results {
		if _, err := tx.NamedExecContext(ctx, insertQuery, res); err != nil {
			return fmt.Errorf("insert result: %w", err)
		}
	}

	return nil
}

func (d *Database) DeleteTournamentParticipantsTx(ctx context.Context, tx *sqlx.Tx, tournamentID string) error {
	const query = `
		DELETE FROM tournament_participants
		WHERE tournament_id = $1
	`
	if _, err := tx.ExecContext(ctx, query, tournamentID); err != nil {
		return fmt.Errorf("delete participants: %w", err)
	}
	return nil
}
