package database

import (
	"context"
	"fmt"
	"tournament/models"
)

func (d *Database) GetParticipants(ctx context.Context, tournamentID string) ([]models.TournamentParticipant, error) {
	var participants []models.TournamentParticipant
	const query = `
		SELECT tournament_id, user_id, username, score, solved_count, joined_at, 
		last_solved_at
		FROM tournament_participants
		WHERE tournament_id = $1
		ORDER BY joined_at DESC
	`

	if err := d.DB.SelectContext(ctx, &participants, query, tournamentID); err != nil {
		return nil, fmt.Errorf("get participants: %w", err)
	}

	return participants, nil
}

func (d *Database) RegisterParticipant(ctx context.Context, participant *models.TournamentParticipant) error {
	const query = `
		INSERT INTO tournament_participants (
			tournament_id, user_id, username, score, 
			solved_count, joined_at, last_solved_at
		)
		VALUES (
			:tournament_id, :user_id, :username, :score, 
			:solved_count, :joined_at, :last_solved_at
		)
		ON CONFLICT (tournament_id, user_id) DO NOTHING
	`

	res, err := d.DB.NamedExecContext(ctx, query, participant)
	if err != nil {
		return fmt.Errorf("register participant: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("participant already exists")
	}

	return nil
}

// DeleteParticipant удаляет участника из турнира
func (d *Database) DeleteParticipant(ctx context.Context, tournamentID, userID string) error {
	const query = `DELETE FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2`

	res, err := d.DB.ExecContext(ctx, query, tournamentID, userID)
	if err != nil {
		return fmt.Errorf("delete participant: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected error: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("participant or tournament not found")
	}

	return nil
}
