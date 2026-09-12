package seed

import (
	"context"
	"database/sql"
	"fmt"
)

type team struct {
	name      string
	shortName string
	slug      string
}

var teams = []team{
	{"AEK Athens", "AEK", "aek-athens"},
	{"Arsenal", "ARS", "arsenal"},
	{"Aston Villa", "AVL", "aston-villa"},
	{"Atletico Madrid", "ATM", "atletico-madrid"},
	{"Barcelona", "BAR", "barcelona"},
	{"Bayern Munich", "BAY", "bayern-munich"},
	{"Bodo/Glimt", "BOD", "bodo-glimt"},
	{"Borussia Dortmund", "BVB", "borussia-dortmund"},
	{"Club Brugge", "BRU", "club-brugge"},
	{"Como", "COM", "como"},
	{"Fenerbahce", "FEN", "fenerbahce"},
	{"Feyenoord", "FEY", "feyenoord"},
	{"Galatasaray", "GAL", "galatasaray"},
	{"Inter Milan", "INT", "inter-milan"},
	{"LASK", "LAS", "lask"},
	{"RB Leipzig", "RBL", "rb-leipzig"},
	{"Lens", "LEN", "lens"},
	{"Lille", "LIL", "lille"},
	{"Liverpool", "LIV", "liverpool"},
	{"Manchester City", "MCI", "manchester-city"},
	{"Manchester United", "MUN", "manchester-united"},
	{"Napoli", "NAP", "napoli"},
	{"Paris Saint-Germain", "PSG", "paris-saint-germain"},
	{"Porto", "POR", "porto"},
	{"PSV Eindhoven", "PSV", "psv-eindhoven"},
	{"Real Betis", "BET", "real-betis"},
	{"Real Madrid", "RMA", "real-madrid"},
	{"Roma", "ROM", "roma"},
	{"Sabah", "SAB", "sabah"},
	{"Shakhtar Donetsk", "SHK", "shakhtar-donetsk"},
	{"Slavia Prague", "SLA", "slavia-prague"},
	{"Slovan Bratislava", "SLO", "slovan-bratislava"},
	{"Sporting CP", "SCP", "sporting-cp"},
	{"Stuttgart", "STU", "stuttgart"},
	{"Viking", "VIK", "viking"},
	{"Villarreal", "VIL", "villarreal"},
}

func validateTeams(items []team) error {
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if item.name == "" || item.slug == "" {
			return fmt.Errorf("team name and slug must be set")
		}
		if len(item.shortName) != 3 {
			return fmt.Errorf("team %q short_name must contain exactly three uppercase ASCII letters", item.slug)
		}
		for _, letter := range item.shortName {
			if letter < 'A' || letter > 'Z' {
				return fmt.Errorf("team %q short_name must contain exactly three uppercase ASCII letters", item.slug)
			}
		}
		if seen[item.slug] {
			return fmt.Errorf("duplicate team slug %q", item.slug)
		}
		seen[item.slug] = true
	}
	return nil
}

func seedTeams(ctx context.Context, tx *sql.Tx) error {
	if err := validateTeams(teams); err != nil {
		return err
	}
	for _, item := range teams {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO teams (name, short_name, slug, logo_url)
			VALUES ($1, $2, $3, NULL)
			ON CONFLICT (slug) DO UPDATE
			SET name = EXCLUDED.name, short_name = EXCLUDED.short_name
		`, item.name, item.shortName, item.slug)
		if err != nil {
			return fmt.Errorf("upsert team %q: %w", item.slug, err)
		}
	}
	return nil
}
