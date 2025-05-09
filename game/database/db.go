package database

import (
	"context"
	"database/sql"
	"fmt"
	"game/config"
	"game/models"
	"log"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func (d *Database) Config() {
	panic("unimplemented")
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// sudoku_fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sudoku_fields (
			id VARCHAR(36) PRIMARY KEY,
			initial_field VARCHAR(81) NOT NULL,
			solution VARCHAR(81) NOT NULL,
			complexity TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

			solve_attempts INTEGER NOT NULL DEFAULT 0,
			solves_successful INTEGER NOT NULL DEFAULT 0,
			solves_total_time BIGINT NOT NULL DEFAULT 0,

			UNIQUE (initial_field, solution)
		)
	`)
	if err != nil {
		return nil, err
	}

	// sudoku_tags
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sudoku_tags (
			id VARCHAR(36) PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		return nil, err
	}

	// sudoku_field_tags
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sudoku_field_tags (
			field_id VARCHAR(36) NOT NULL REFERENCES sudoku_fields(id) ON DELETE CASCADE,
			tag_id VARCHAR(36) NOT NULL REFERENCES sudoku_tags(id) ON DELETE CASCADE,
			PRIMARY KEY (field_id, tag_id)
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetFieldByComplexity(level string) (*models.SudokuField, error) {
	query := `
		SELECT id, initial_field, solution, complexity, created_at,
		       solve_attempts, solves_successful, solves_total_time
		FROM sudoku_fields
		WHERE complexity = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	field := &models.SudokuField{}
	err := d.db.QueryRow(query, level).Scan(
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
		return nil, err
	}

	_, _ = d.db.ExecContext(context.Background(), `
		UPDATE sudoku_fields
		SET solve_attempts = solve_attempts + 1
		WHERE id = $1
	`, field.ID)

	return field, nil
}

func (d *Database) GetFieldsByComplexity(difficulty string) ([]models.SudokuField, error) {
	query := `
		SELECT id, initial_field, solution, complexity, created_at,
		       solve_attempts, solves_successful, solves_total_time
		FROM sudoku_fields 
		WHERE complexity = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	rows, err := d.db.Query(query, difficulty)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []models.SudokuField
	for rows.Next() {
		var field models.SudokuField
		if err := rows.Scan(
			&field.ID,
			&field.InitialField,
			&field.Solution,
			&field.Complexity,
			&field.CreatedAt,
			&field.SolveAttempts,
			&field.SolvesSuccessful,
			&field.SolvesTotalTime,
		); err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

func (d *Database) MarkSudokuSolved(ctx context.Context, id string, solveTimeMs int64) error {
	_, err := d.db.ExecContext(ctx, `
		UPDATE sudoku_fields
		SET solves_successful = solves_successful + 1,
		    solves_total_time = solves_total_time + $2
		WHERE id = $1
	`, id, solveTimeMs)
	return err
}

func (d *Database) GetSudokuByID(ctx context.Context, id string) (*models.SudokuField, error) {
	query := `
		SELECT id, initial_field, solution, complexity, created_at,
		       solve_attempts, solves_successful, solves_total_time
		FROM sudoku_fields
		WHERE id = $1
	`
	field := &models.SudokuField{}
	err := d.db.QueryRowContext(ctx, query, id).Scan(
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
		return nil, err
	}

	_, _ = d.db.ExecContext(context.Background(), `
		UPDATE sudoku_fields
		SET solve_attempts = solve_attempts + 1
		WHERE id = $1
	`, field.ID)

	return field, nil
}

func (d *Database) InsertAchievement(a models.Achievement) error {
	_, err := d.db.Exec(`
		INSERT INTO achievements (code, title, description, icon_url)
		VALUES ($1, $2, $3, $4)
	`, a.Code, a.Title, a.Description, a.IconURL)
	if err != nil {
		// Можно позже добавить проверку на конфликт кода
		return err
	}
	return nil
}

func (d *Database) GetAllAchievements() ([]models.Achievement, error) {
	rows, err := d.db.Query(`SELECT code, title, description, icon_url FROM achievements`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var a models.Achievement
		if err := rows.Scan(&a.Code, &a.Title, &a.Description, &a.IconURL); err != nil {
			return nil, err
		}
		achievements = append(achievements, a)
	}

	return achievements, nil
}

func (d *Database) GetAchievementIconFilename(achievementCode string) (string, error) {
	query := `SELECT icon_url FROM achievements WHERE code = $1`
	var filename string
	err := d.db.QueryRow(query, achievementCode).Scan(&filename)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return filename, nil
}

func (d *Database) DeleteAchievement(code string) error {
	_, err := d.db.Exec(`DELETE FROM achievements WHERE code = $1`, code)
	return err
}

func (d *Database) GetUserAchievements(userID string) ([]models.UserAchievementResponse, error) {
	rows, err := d.db.Query(`
		SELECT a.code, a.title, a.description
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		WHERE ua.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.UserAchievementResponse
	for rows.Next() {
		var a models.UserAchievementResponse
		if err := rows.Scan(&a.Code, &a.Title, &a.Description); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func (d *Database) AssignAchievementByCode(userID, code string) (models.AchievementResponse, error) {

	log.Printf("code: %v", code)     // добавь лог
	log.Printf("userID: %v", userID) // добавь лог

	var id int
	var title, description string

	err := d.db.QueryRow(`SELECT id, title, description FROM achievements WHERE code = $1`, code).
		Scan(&id, &title, &description)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.AchievementResponse{}, fmt.Errorf("achievement not found")
		}
		return models.AchievementResponse{}, err
	}

	log.Printf("checked") // добавь лог

	_, err = d.db.Exec(`
		INSERT INTO user_achievements (user_id, achievement_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, id)
	if err != nil {
		return models.AchievementResponse{}, err
	}

	log.Printf("assigned") // добавь лог

	return models.AchievementResponse{
		Code:        code,
		Title:       title,
		Description: description,
	}, nil
}

func (d *Database) GetUserAchievementsCodes(userID string) ([]string, error) {
	rows, err := d.db.Query(`
		SELECT a.code
		FROM user_achievements ua
		JOIN achievements a ON ua.achievement_id = a.id
		WHERE ua.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (d *Database) CheckAndAssignAchievements(userID string, stats []models.UserDifficultyStat) ([]models.AchievementResponse, error) {
	existing, err := d.GetUserAchievementsCodes(userID)
	if err != nil {
		return nil, err
	}
	existingSet := make(map[string]struct{})
	for _, code := range existing {
		existingSet[code] = struct{}{}
	}

	var newAchievements []models.AchievementResponse

	// 1. Первое решённое судоку
	totalSolved := 0
	for _, s := range stats {
		totalSolved += s.TotalSolved
	}

	if totalSolved >= 1 && !has(existingSet, "first_sudoku") {
		if a, err := d.AssignAchievementByCode(userID, "first_sudoku"); err == nil {
			newAchievements = append(newAchievements, a)
		}
	}

	// 2. 10 лёгких
	for _, s := range stats {
		if s.Difficulty == "easy" && s.TotalSolved >= 10 && !has(existingSet, "ten_easy") {
			if a, err := d.AssignAchievementByCode(userID, "ten_easy"); err == nil {
				newAchievements = append(newAchievements, a)
			}
		}
	}

	// 3. Первое medium
	for _, s := range stats {
		if s.Difficulty == "medium" && s.TotalSolved >= 1 && !has(existingSet, "first_medium") {
			if a, err := d.AssignAchievementByCode(userID, "first_medium"); err == nil {
				newAchievements = append(newAchievements, a)
			}
		}
	}

	return newAchievements, nil
}

func has(set map[string]struct{}, key string) bool {
	_, ok := set[key]
	return ok
}

func (d *Database) DeleteUserAchievement(userID, code string) error {
	_, err := d.db.Exec(`
		DELETE FROM user_achievements
		WHERE user_id = $1 AND achievement_id = (SELECT id FROM achievements WHERE code = $2)
	`, userID, code)
	return err
}
