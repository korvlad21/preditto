package development

import (
	"context"
	"database/sql"
	"fmt"
)

func seedTournamentPrizeRules(ctx context.Context, tx *sql.Tx) error {
	// Only 95% is allocated; the remaining 5% has no prize rule.
	for _, item := range []struct {
		slug        string
		percentages [3]string
	}{
		{championsLeague20262027Slug, [3]string{"50.00", "30.00", "15.00"}},
		{championsLeague20262027GELSlug, [3]string{"50.00", "30.00", "15.00"}},
		{laLiga20262027Slug, [3]string{"45.00", "30.00", "20.00"}},
	} {
		var tournamentID int64
		if err := tx.QueryRowContext(ctx, `
			SELECT id FROM tournaments WHERE slug = $1
		`, item.slug).Scan(&tournamentID); err != nil {
			return fmt.Errorf("find tournament for prize rules %q: %w", item.slug, err)
		}
		for i, percentage := range item.percentages {
			place := i + 1
			_, err := tx.ExecContext(ctx, `
				INSERT INTO tournament_prize_rules (tournament_id, place, percentage)
				VALUES ($1, $2, $3)
				ON CONFLICT (tournament_id, place) DO UPDATE SET percentage = EXCLUDED.percentage
			`, tournamentID, place, percentage)
			if err != nil {
				return fmt.Errorf("upsert tournament %q prize place %d: %w", item.slug, place, err)
			}
		}
	}
	return nil
}
