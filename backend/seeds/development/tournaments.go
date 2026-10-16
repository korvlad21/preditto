package development

import (
	"context"
	"database/sql"
	"fmt"
)

const championsLeague2026Slug = "uefa-champions-league-2026-2027"

func seedTournaments(ctx context.Context, tx *sql.Tx) error {
	// UEFA's full season includes qualifying; timestamps represent calendar dates.
	// https://www.uefa.com/uefachampionsleague/news/02a6-20d57cfcd03e-407c22a7f465-1000--2026-27-champions-league-teams-dates-draws-format-final/
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tournaments (name, slug, status, starts_at, ends_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			starts_at = EXCLUDED.starts_at,
			ends_at = EXCLUDED.ends_at,
			updated_at = CURRENT_TIMESTAMP
	`, "Лига чемпионов УЕФА 2026/2027", championsLeague2026Slug, "active", "2026-09-08", "2027-06-05")
	if err != nil {
		return fmt.Errorf("upsert tournament %q: %w", championsLeague2026Slug, err)
	}
	return nil
}
