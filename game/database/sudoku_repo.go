package database

import (
	"context"
	"database/sql"
	"fmt"
	"game/models"

	_ "github.com/lib/pq"
)

func (d *Database) GetRandomByComplexity(ctx context.Context, complexity string) (*models.SudokuField, error) {
	const query = `
		SELECT id, initial_field, solution, complexity, created_at,
		       solve_attempts, solves_successful, solves_total_time
		FROM sudoku_fields
		WHERE complexity = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := d.DB.QueryRowContext(ctx, query, complexity)

	var field models.SudokuField
	err := row.Scan(
		&field.ID,
		&field.InitialField,
		&field.Solution,
		&field.Complexity,
		&field.CreatedAt,
		&field.SolveAttempts,
		&field.SolvesSuccessful,
		&field.SolvesTotalTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan sudoku: %w", err)
	}

	go d.incrementAttempts(context.Background(), field.ID)

	return &field, nil
}

func (d *Database) incrementAttempts(ctx context.Context, id string) {
	const update = `UPDATE sudoku_fields SET solve_attempts = solve_attempts + 1 WHERE id = $1`
	_, _ = d.DB.ExecContext(ctx, update, id)
}

func (d *Database) GetFieldsByComplexity(ctx context.Context, difficulty string) ([]models.SudokuField, error) {
	const query = `
		SELECT id, initial_field, solution, complexity, created_at,
		       solve_attempts, solves_successful, solves_total_time
		FROM sudoku_fields
		WHERE complexity = $1
		ORDER BY created_at DESC
	`

	rows, err := d.DB.QueryContext(ctx, query, difficulty)
	if err != nil {
		return nil, fmt.Errorf("query fields: %w", err)
	}
	defer rows.Close()

	var fields []models.SudokuField
	for rows.Next() {
		var f models.SudokuField
		if err := rows.Scan(
			&f.ID,
			&f.InitialField,
			&f.Solution,
			&f.Complexity,
			&f.CreatedAt,
			&f.SolveAttempts,
			&f.SolvesSuccessful,
			&f.SolvesTotalTime,
		); err != nil {
			return nil, fmt.Errorf("scan field: %w", err)
		}
		fields = append(fields, f)
	}
	return fields, nil
}

func (d *Database) GetSudokuByID(ctx context.Context, id string) (*models.SudokuField, error) {
	const query = `
		SELECT id, initial_field, solution, complexity, created_at,
		       solve_attempts, solves_successful, solves_total_time
		FROM sudoku_fields
		WHERE id = $1
	`

	var field models.SudokuField
	err := d.DB.QueryRowContext(ctx, query, id).Scan(
		&field.ID,
		&field.InitialField,
		&field.Solution,
		&field.Complexity,
		&field.CreatedAt,
		&field.SolveAttempts,
		&field.SolvesSuccessful,
		&field.SolvesTotalTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get sudoku by id: %w", err)
	}

	go d.incrementAttempts(context.Background(), id)
	return &field, nil
}

func (d *Database) MarkSudokuSolved(ctx context.Context, id string, solveTimeMs int64) error {
	const query = `
		UPDATE sudoku_fields
		SET solves_successful = solves_successful + 1,
		    solves_total_time = solves_total_time + $2
		WHERE id = $1
	`

	_, err := d.DB.ExecContext(ctx, query, id, solveTimeMs)
	if err != nil {
		return fmt.Errorf("update solved stats: %w", err)
	}
	return nil
}

func (d *Database) GetSudokuDifficultyByID(ctx context.Context, sudokuID string) (string, error) {
	var difficulty string

	err := d.DB.QueryRowContext(ctx, `
		SELECT complexity 
		FROM sudoku_fields 
		WHERE id = $1
	`, sudokuID).Scan(&difficulty)

	if err != nil {
		return "", fmt.Errorf("get difficulty by sudoku id: %w", err)
	}

	return difficulty, nil
}
