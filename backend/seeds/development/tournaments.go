package development

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

const (
	championsLeague20262027Slug    = "uefa-champions-league-2026-2027"
	championsLeague20262027GELSlug = "uefa-champions-league-2026-2027-gel"
	laLiga20262027Slug             = "la-liga-2026-2027"
)

type tournament struct {
	name             string
	slug             string
	status           string
	startsAt         string
	endsAt           string
	costAmount       string
	amountCurrency   string
	typeOfTournament string
}

var tournaments = []tournament{
	// The tournament starts with the league phase; timestamps represent calendar dates.
	// https://www.uefa.com/uefachampionsleague/news/02a6-20d57cfcd03e-407c22a7f465-1000--2026-27-champions-league-teams-dates-draws-format-final/
	{
		name:             "Лига чемпионов УЕФА 2026/2027",
		slug:             championsLeague20262027Slug,
		status:           model.TournamentStatusActive,
		startsAt:         "2026-09-08",
		endsAt:           "2027-06-05",
		costAmount:       "10.00",
		amountCurrency:   model.CurrencyUSD,
		typeOfTournament: model.TournamentTypeChampionsLeague,
	},
	{
		name:             "Лига чемпионов УЕФА 2026/2027 (GEL)",
		slug:             championsLeague20262027GELSlug,
		status:           model.TournamentStatusActive,
		startsAt:         "2026-09-08",
		endsAt:           "2027-06-05",
		costAmount:       "40.00",
		amountCurrency:   model.CurrencyGEL,
		typeOfTournament: model.TournamentTypeChampionsLeague,
	},
	// First match and final round dates, stored as calendar dates rather than kickoff times.
	// https://www.laliga.com/noticias/horarios-de-la-primera-jornada-de-laliga-ea-sports-2026-27
	// https://rfef.es/es/noticias/la-temporada-202627-ya-tiene-establecidas-sus-fechas-clave
	{
		name:             "Ла Лига 2026-2027",
		slug:             laLiga20262027Slug,
		status:           model.TournamentStatusActive,
		startsAt:         "2026-08-15",
		endsAt:           "2027-05-30",
		costAmount:       "20.00",
		amountCurrency:   model.CurrencyUSD,
		typeOfTournament: model.TournamentTypeLaLiga,
	},
}

func seedTournaments(ctx context.Context, tx *sql.Tx) error {
	for _, item := range tournaments {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO tournaments (name, slug, status, starts_at, ends_at, cost_amount, amount_currency, type_of_tournament)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			starts_at = EXCLUDED.starts_at,
			ends_at = EXCLUDED.ends_at,
			cost_amount = EXCLUDED.cost_amount,
			amount_currency = EXCLUDED.amount_currency,
			type_of_tournament = EXCLUDED.type_of_tournament,
			updated_at = CURRENT_TIMESTAMP
		`, item.name, item.slug, item.status, item.startsAt, item.endsAt, item.costAmount, item.amountCurrency, item.typeOfTournament)
		if err != nil {
			return fmt.Errorf("upsert tournament %q: %w", item.slug, err)
		}
	}
	return nil
}
