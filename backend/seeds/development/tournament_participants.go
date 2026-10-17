package development

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type tournamentParticipant struct {
	tournamentID int64
	username     string
	status       string
}

var tournamentParticipants = []tournamentParticipant{
	{1, "ines_costa", model.TournamentParticipantStatusActive},
	{1, "alex_morgan", model.TournamentParticipantStatusActive},
	{1, "felix_schmidt", model.TournamentParticipantStatusActive},
	{1, "sofia_rossi", model.TournamentParticipantStatusActive},
	{1, "leo_moretti", model.TournamentParticipantStatusActive},
	{1, "emma_wilson", model.TournamentParticipantStatusActive},
	{1, "oscar_lind", model.TournamentParticipantStatusActive},
	{1, "mateo_garcia", model.TournamentParticipantStatusActive},
	{1, "noah_clarke", model.TournamentParticipantStatusNotPaid},
	{1, "zara_ahmed", model.TournamentParticipantStatusBlocked},
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
