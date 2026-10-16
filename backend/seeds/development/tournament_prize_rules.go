package development

import (
	"context"
	"database/sql"
	"fmt"
)

func seedTournamentPrizeRules(ctx context.Context, tx *sql.Tx) error {
	var tournamentID int64
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM tournaments WHERE slug = $1
	`, championsLeague20262027Slug).Scan(&tournamentID); err != nil {
		return fmt.Errorf("find tournament for prize rules %q: %w", championsLeague20262027Slug, err)
	}
	// Only 95% is allocated; the remaining 5% has no prize rule.
	for _, item := range []struct {
		place      int
		percentage string
	}{
		{1, "50.00"},
		{2, "30.00"},
		{3, "15.00"},
	} {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO tournament_prize_rules (tournament_id, place, percentage)
			VALUES ($1, $2, $3)
			ON CONFLICT (tournament_id, place) DO UPDATE SET percentage = EXCLUDED.percentage
		`, tournamentID, item.place, item.percentage)
		if err != nil {
			return fmt.Errorf("upsert tournament %q prize place %d: %w", championsLeague20262027Slug, item.place, err)
		}
	}
	return nil
}
