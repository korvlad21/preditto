package development

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type tournamentParticipant struct {
	tournamentSlug string
	username       string
	status         string
}

var tournamentParticipants = []tournamentParticipant{
	{championsLeague20262027Slug, "korvlad21", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "ines_costa", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "alex_morgan", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "felix_schmidt", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "sofia_rossi", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "leo_moretti", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "emma_wilson", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "oscar_lind", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "mateo_garcia", model.TournamentParticipantStatusActive},
	{championsLeague20262027Slug, "noah_clarke", model.TournamentParticipantStatusNotPaid},
	{championsLeague20262027Slug, "zara_ahmed", model.TournamentParticipantStatusBlocked},

	{championsLeague20262027GELSlug, "ines_costa", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "alex_morgan", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "felix_schmidt", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "sofia_rossi", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "leo_moretti", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "emma_wilson", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "oscar_lind", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "mateo_garcia", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "daniel_weber", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "lucas_silva", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "olivia_brown", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "mia_jensen", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "elena_popov", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "hanna_stoma", model.TournamentParticipantStatusActive},
	{championsLeague20262027GELSlug, "noah_clarke", model.TournamentParticipantStatusNotPaid},

	{laLiga20262027Slug, "lucas_silva", model.TournamentParticipantStatusActive},
	{laLiga20262027Slug, "mia_jensen", model.TournamentParticipantStatusActive},
	{laLiga20262027Slug, "ines_costa", model.TournamentParticipantStatusActive},
	{laLiga20262027Slug, "daniel_weber", model.TournamentParticipantStatusActive},
	{laLiga20262027Slug, "olivia_brown", model.TournamentParticipantStatusActive},
	{laLiga20262027Slug, "leo_moretti", model.TournamentParticipantStatusActive},
	{laLiga20262027Slug, "elena_popov", model.TournamentParticipantStatusBlocked},
}

func seedTournamentParticipants(ctx context.Context, tx *sql.Tx) error {
	tournamentIDs := make(map[string]int64)
	for _, item := range tournamentParticipants {
		tournamentID, ok := tournamentIDs[item.tournamentSlug]
		if !ok {
			if err := tx.QueryRowContext(ctx, `
				SELECT id FROM tournaments WHERE slug = $1
			`, item.tournamentSlug).Scan(&tournamentID); err != nil {
				return fmt.Errorf("find tournament for participants %q: %w", item.tournamentSlug, err)
			}
			tournamentIDs[item.tournamentSlug] = tournamentID
		}
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
		`, tournamentID, userID, item.status)
		if err != nil {
			return fmt.Errorf("upsert tournament %d participant %q: %w", tournamentID, item.username, err)
		}
	}
	return nil
}
