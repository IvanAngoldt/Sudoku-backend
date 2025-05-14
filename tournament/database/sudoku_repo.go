package database

import (
	"context"
	"fmt"
	"time"
)

func (d *Database) UpdateTournamentParticipant(ctx context.Context, tournamentID, userID string, solvedAt time.Time, scoreDelta int) error {
	const query = `
		UPDATE tournament_participants
		SET 
			score = score + $1,
			solved_count = solved_count + 1,
			last_solved_at = $2
		WHERE tournament_id = $3 AND user_id = $4
	`
	_, err := d.DB.ExecContext(ctx, query, scoreDelta, solvedAt, tournamentID, userID)
	if err != nil {
		return fmt.Errorf("update participant stats: %w", err)
	}
	return nil
}
