package development

import (
	"context"
	"database/sql"
	"fmt"
)

type tournamentParticipant struct {
	tournamentID int64
	username     string
	status       string
}

var tournamentParticipants = []tournamentParticipant{
	{1, "ines_costa", "active"},
	{1, "alex_morgan", "active"},
	{1, "felix_schmidt", "active"},
	{1, "sofia_rossi", "active"},
	{1, "leo_moretti", "active"},
	{1, "emma_wilson", "active"},
	{1, "oscar_lind", "active"},
	{1, "mateo_garcia", "active"},
	{1, "noah_clarke", "not_paid"},
	{1, "zara_ahmed", "blocked"},
}

func seedTournamentParticipants(ctx context.Context, tx *sql.Tx) error {
	for _, item := range tournamentParticipants {
		var userID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM users WHERE username = $1
		`, item.username).Scan(&userID); err != nil {
			return fmt.Errorf("find user for tournament_participants %q: %w", item.username, err)
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tournament_participants (tournament_id, user_id, status)
			VALUES ($1, $2, $3)
			ON CONFLICT (tournament_id, user_id) DO UPDATE SET status = EXCLUDED.status
		`, item.tournamentID, userID, item.status)
		if err != nil {
			return fmt.Errorf("upsert tournament %d participant %q: %w", item.tournamentID, item.username, err)
		}
	}
	return nil
}
