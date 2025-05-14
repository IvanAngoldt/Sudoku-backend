package database

import (
	_ "github.com/lib/pq"
)

// // AddParticipant добавляет участника в турнир
// func (d *Database) AddParticipant(participant *models.TournamentParticipant) error {
// 	query := `INSERT INTO tournament_participants (id, tournament_id, user_id, score, solved_count, joined_at, last_solved_at)
// 				VALUES ($1, $2, $3, $4, $5, $6, $7)`
// 	_, err := d.db.Exec(query,
// 		participant.ID, participant.TournamentID, participant.UserID,
// 		participant.Score, participant.SolvedCount, participant.JoinedAt,
// 		participant.LastSolvedAt,
// 	)
// 	if err != nil {
// 		log.Println("Error adding participant:", err)
// 		return err
// 	}
// 	return nil
// }

// // GetParticipant получает информацию об участнике турнира
// func (d *Database) GetParticipant(tournamentID, userID string) (*models.TournamentParticipant, error) {
// 	query := `SELECT * FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2`
// 	var participant models.TournamentParticipant
// 	err := d.db.QueryRow(query, tournamentID, userID).Scan(
// 		&participant.ID, &participant.TournamentID, &participant.UserID,
// 		&participant.Score, &participant.SolvedCount, &participant.JoinedAt,
// 		&participant.LastSolvedAt,
// 	)
// 	if err != nil {
// 		log.Println("Error getting participant:", err)
// 		return nil, err
// 	}
// 	return &participant, nil
// }

// // UpdateParticipant обновляет информацию об участнике
// func (d *Database) UpdateParticipant(participant *models.TournamentParticipant) error {
// 	query := `UPDATE tournament_participants
// 				SET score = $1, solved_count = $2, last_solved_at = $3
// 				WHERE id = $4`
// 	_, err := d.db.Exec(query,
// 		participant.Score, participant.SolvedCount, participant.LastSolvedAt,
// 		participant.ID,
// 	)
// 	if err != nil {
// 		log.Println("Error updating participant:", err)
// 		return err
// 	}
// 	return nil
// }

// // AddResult добавляет результат турнира
// func (d *Database) AddResult(result *models.TournamentResult) error {
// 	query := `INSERT INTO tournament_results (id, tournament_id, user_id, score, rank, solved_count, finished_at)
// 				VALUES ($1, $2, $3, $4, $5, $6, $7)`
// 	_, err := d.db.Exec(query,
// 		result.ID, result.TournamentID, result.UserID,
// 		result.Score, result.Rank, result.SolvedCount,
// 		result.FinishedAt,
// 	)
// 	if err != nil {
// 		log.Println("Error adding result:", err)
// 		return err
// 	}
// 	return nil
// }

// func (d *Database) GetTopParticipants(client *redis.Client, tournamentID string, count int64) ([]redis.Z, error) {
// 	ctx := context.Background()
// 	return client.ZRevRangeWithScores(ctx, "tournament:"+tournamentID+":scores", 0, count-1).Result()
// }

// // AddSolvedSudoku добавляет запись о решенной судоку
// func (d *Database) AddSolvedSudoku(tournamentID, userID, sudokuID string) error {
// 	query := `INSERT INTO solved_sudokus (id, tournament_id, user_id, sudoku_id)
// 				VALUES ($1, $2, $3, $4)`
// 	_, err := d.db.Exec(query,
// 		uuid.New().String(),
// 		tournamentID,
// 		userID,
// 		sudokuID,
// 	)
// 	if err != nil {
// 		log.Println("Error adding solved sudoku:", err)
// 		return err
// 	}
// 	return nil
// }

// // IsSudokuSolved проверяет, решал ли пользователь эту судоку в рамках турнира
// func (d *Database) IsSudokuSolved(tournamentID, userID, sudokuID string) (bool, error) {
// 	query := `SELECT EXISTS(
// 		SELECT 1 FROM solved_sudokus
// 		WHERE tournament_id = $1 AND user_id = $2 AND sudoku_id = $3
// 	)`
// 	var exists bool
// 	err := d.db.QueryRow(query, tournamentID, userID, sudokuID).Scan(&exists)
// 	if err != nil {
// 		log.Println("Error checking if sudoku is solved:", err)
// 		return false, err
// 	}
// 	return exists, nil
// }
