package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"tournament/config"
	"tournament/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
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
		log.Println("Error opening database:", err)
		return nil, err
	}

	if err = db.Ping(); err != nil {
		log.Println("Error pinging database:", err)
		return nil, err
	}

	// tournaments
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tournaments (
			id VARCHAR(36) PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Println("Error creating tournaments table:", err)
		return nil, err
	}

	// sudoku_tags
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tournament_participants (
			id VARCHAR(36) PRIMARY KEY,
			tournament_id VARCHAR(36) NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			solved_count INTEGER NOT NULL DEFAULT 0,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_solved_at TIMESTAMP
		)
	`)
	if err != nil {
		log.Println("Error creating tournament_participants table:", err)
		return nil, err
	}

	// solved_sudokus
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS solved_sudokus (
			id VARCHAR(36) PRIMARY KEY,
			tournament_id VARCHAR(36) NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			sudoku_id VARCHAR(36) NOT NULL,
			solved_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tournament_id, user_id, sudoku_id)
		)
	`)
	if err != nil {
		log.Println("Error creating solved_sudokus table:", err)
		return nil, err
	}

	// sudoku_field_tags
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tournament_results (
			id VARCHAR(36) PRIMARY KEY,
			tournament_id VARCHAR(36) NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
			user_id VARCHAR(36) NOT NULL,
			score INTEGER NOT NULL DEFAULT 0,
			rank INTEGER NOT NULL,
			solved_count INTEGER NOT NULL DEFAULT 0,
			finished_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		log.Println("Error creating tournament_results table:", err)
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// CreateTournament создает новый турнир
func (d *Database) CreateTournament(tournament *models.Tournament) error {
	query := `INSERT INTO tournaments (id, name, description, start_time, end_time, status, created_by, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := d.db.Exec(
		query,
		tournament.ID, tournament.Name, tournament.Description, tournament.StartTime,
		tournament.EndTime, tournament.Status, tournament.CreatedBy,
		tournament.CreatedAt,
	)
	if err != nil {
		log.Println("Error creating tournament:", err)
		return err
	}

	return nil
}

// GetTournaments возвращает список всех турниров
func (d *Database) GetTournaments() ([]models.Tournament, error) {
	query := `SELECT * FROM tournaments ORDER BY created_at DESC`
	rows, err := d.db.Query(query)
	if err != nil {
		log.Println("Error getting tournaments:", err)
		return nil, err
	}
	defer rows.Close()

	var tournaments []models.Tournament
	for rows.Next() {
		var t models.Tournament
		err := rows.Scan(
			&t.ID, &t.Name, &t.Description, &t.StartTime, &t.EndTime,
			&t.Status, &t.CreatedBy, &t.CreatedAt,
		)
		if err != nil {
			log.Println("Error scanning tournament:", err)
			return nil, err
		}
		tournaments = append(tournaments, t)
	}
	return tournaments, nil
}

// GetTournament возвращает турнир по ID
func (d *Database) GetTournament(id string) (*models.Tournament, error) {
	query := `SELECT * FROM tournaments WHERE id = $1`
	var tournament models.Tournament
	err := d.db.QueryRow(query, id).Scan(
		&tournament.ID, &tournament.Name, &tournament.Description,
		&tournament.StartTime, &tournament.EndTime, &tournament.Status,
		&tournament.CreatedBy, &tournament.CreatedAt,
	)
	if err != nil {
		log.Println("Error getting tournament:", err)
		return nil, err
	}
	return &tournament, nil
}

// UpdateTournament обновляет информацию о турнире
func (d *Database) UpdateTournament(id string, input *models.UpdateTournamentInput) error {
	setClauses := []string{}
	args := []interface{}{}
	argIndex := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *input.Description)
		argIndex++
	}
	if input.StartTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("start_time = $%d", argIndex))
		args = append(args, *input.StartTime)
		argIndex++
	}
	if input.EndTime != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_time = $%d", argIndex))
		args = append(args, *input.EndTime)
		argIndex++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *input.Status)
		argIndex++
	}

	if len(setClauses) == 0 {
		return nil // ничего не обновляется
	}

	query := fmt.Sprintf(`
		UPDATE tournaments
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	args = append(args, id)

	_, err := d.db.Exec(query, args...)
	if err != nil {
		log.Println("Error updating tournament:", err)
	}
	return err
}

// DeleteTournament удаляет турнир
func (d *Database) DeleteTournament(id string) error {
	query := `DELETE FROM tournaments WHERE id = $1`
	_, err := d.db.Exec(query, id)
	if err != nil {
		log.Println("Error deleting tournament:", err)
		return err
	}
	return nil
}

// AddParticipant добавляет участника в турнир
func (d *Database) AddParticipant(participant *models.TournamentParticipant) error {
	query := `INSERT INTO tournament_participants (id, tournament_id, user_id, score, solved_count, joined_at, last_solved_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := d.db.Exec(query,
		participant.ID, participant.TournamentID, participant.UserID,
		participant.Score, participant.SolvedCount, participant.JoinedAt,
		participant.LastSolvedAt,
	)
	if err != nil {
		log.Println("Error adding participant:", err)
		return err
	}
	return nil
}

// GetParticipant получает информацию об участнике турнира
func (d *Database) GetParticipant(tournamentID, userID string) (*models.TournamentParticipant, error) {
	query := `SELECT * FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2`
	var participant models.TournamentParticipant
	err := d.db.QueryRow(query, tournamentID, userID).Scan(
		&participant.ID, &participant.TournamentID, &participant.UserID,
		&participant.Score, &participant.SolvedCount, &participant.JoinedAt,
		&participant.LastSolvedAt,
	)
	if err != nil {
		log.Println("Error getting participant:", err)
		return nil, err
	}
	return &participant, nil
}

// UpdateParticipant обновляет информацию об участнике
func (d *Database) UpdateParticipant(participant *models.TournamentParticipant) error {
	query := `UPDATE tournament_participants 
				SET score = $1, solved_count = $2, last_solved_at = $3
				WHERE id = $4`
	_, err := d.db.Exec(query,
		participant.Score, participant.SolvedCount, participant.LastSolvedAt,
		participant.ID,
	)
	if err != nil {
		log.Println("Error updating participant:", err)
		return err
	}
	return nil
}

// GetParticipants получает список всех участников турнира
func (d *Database) GetParticipants(tournamentID string) ([]*models.TournamentParticipant, error) {
	query := `SELECT * FROM tournament_participants WHERE tournament_id = $1 ORDER BY score DESC`
	rows, err := d.db.Query(query, tournamentID)
	if err != nil {
		log.Println("Error getting participants:", err)
		return nil, err
	}
	defer rows.Close()

	var participants []*models.TournamentParticipant
	for rows.Next() {
		var p models.TournamentParticipant
		err := rows.Scan(
			&p.ID, &p.TournamentID, &p.UserID,
			&p.Score, &p.SolvedCount, &p.JoinedAt,
			&p.LastSolvedAt,
		)
		if err != nil {
			log.Println("Error scanning participant:", err)
			return nil, err
		}
		participants = append(participants, &p)
	}
	return participants, nil
}

// DeleteParticipant удаляет участника из турнира
func (d *Database) DeleteParticipant(tournamentID, userID string) error {
	query := `DELETE FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2`
	_, err := d.db.Exec(query, tournamentID, userID)
	if err != nil {
		log.Println("Error deleting participant:", err)
		return err
	}
	return nil
}

// AddResult добавляет результат турнира
func (d *Database) AddResult(result *models.TournamentResult) error {
	query := `INSERT INTO tournament_results (id, tournament_id, user_id, score, rank, solved_count, finished_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := d.db.Exec(query,
		result.ID, result.TournamentID, result.UserID,
		result.Score, result.Rank, result.SolvedCount,
		result.FinishedAt,
	)
	if err != nil {
		log.Println("Error adding result:", err)
		return err
	}
	return nil
}

func (d *Database) GetTopParticipants(client *redis.Client, tournamentID string, count int64) ([]redis.Z, error) {
	ctx := context.Background()
	return client.ZRevRangeWithScores(ctx, "tournament:"+tournamentID+":scores", 0, count-1).Result()
}

// AddSolvedSudoku добавляет запись о решенной судоку
func (d *Database) AddSolvedSudoku(tournamentID, userID, sudokuID string) error {
	query := `INSERT INTO solved_sudokus (id, tournament_id, user_id, sudoku_id)
				VALUES ($1, $2, $3, $4)`
	_, err := d.db.Exec(query,
		uuid.New().String(),
		tournamentID,
		userID,
		sudokuID,
	)
	if err != nil {
		log.Println("Error adding solved sudoku:", err)
		return err
	}
	return nil
}

// IsSudokuSolved проверяет, решал ли пользователь эту судоку в рамках турнира
func (d *Database) IsSudokuSolved(tournamentID, userID, sudokuID string) (bool, error) {
	query := `SELECT EXISTS(
		SELECT 1 FROM solved_sudokus 
		WHERE tournament_id = $1 AND user_id = $2 AND sudoku_id = $3
	)`
	var exists bool
	err := d.db.QueryRow(query, tournamentID, userID, sudokuID).Scan(&exists)
	if err != nil {
		log.Println("Error checking if sudoku is solved:", err)
		return false, err
	}
	return exists, nil
}
