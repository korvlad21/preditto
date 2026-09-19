package development

import (
	"context"
	"database/sql"
	"fmt"
)

type team struct {
	name      string
	shortName string
	slug      string
	country   string
}

var teams = []team{
	{"AEK Athens", "AEK", "aek-athens", "GRE"},
	{"Arsenal", "ARS", "arsenal", "ENG"},
	{"Aston Villa", "AVL", "aston-villa", "ENG"},
	{"Atletico Madrid", "ATM", "atletico-madrid", "ESP"},
	{"Barcelona", "BAR", "barcelona", "ESP"},
	{"Bayern Munich", "BAY", "bayern-munich", "GER"},
	{"Bodo/Glimt", "BOD", "bodo-glimt", "NOR"},
	{"Borussia Dortmund", "BVB", "borussia-dortmund", "GER"},
	{"Club Brugge", "BRU", "club-brugge", "BEL"},
	{"Como", "COM", "como", "ITA"},
	{"Fenerbahce", "FEN", "fenerbahce", "TUR"},
	{"Feyenoord", "FEY", "feyenoord", "NED"},
	{"Galatasaray", "GAL", "galatasaray", "TUR"},
	{"Inter Milan", "INT", "inter-milan", "ITA"},
	{"LASK", "LAS", "lask", "AUT"},
	{"RB Leipzig", "RBL", "rb-leipzig", "GER"},
	{"Lens", "LEN", "lens", "FRA"},
	{"Lille", "LIL", "lille", "FRA"},
	{"Liverpool", "LIV", "liverpool", "ENG"},
	{"Manchester City", "MCI", "manchester-city", "ENG"},
	{"Manchester United", "MUN", "manchester-united", "ENG"},
	{"Napoli", "NAP", "napoli", "ITA"},
	{"Paris Saint-Germain", "PSG", "paris-saint-germain", "FRA"},
	{"Porto", "POR", "porto", "POR"},
	{"PSV Eindhoven", "PSV", "psv-eindhoven", "NED"},
	{"Real Betis", "BET", "real-betis", "ESP"},
	{"Real Madrid", "RMA", "real-madrid", "ESP"},
	{"Roma", "ROM", "roma", "ITA"},
	{"Sabah", "SAB", "sabah", "AZE"},
	{"Shakhtar Donetsk", "SHK", "shakhtar-donetsk", "UKR"},
	{"Slavia Prague", "SLA", "slavia-prague", "CZE"},
	{"Slovan Bratislava", "SLO", "slovan-bratislava", "SVK"},
	{"Sporting CP", "SCP", "sporting-cp", "POR"},
	{"Stuttgart", "STU", "stuttgart", "GER"},
	{"Viking", "VIK", "viking", "NOR"},
	{"Villarreal", "VIL", "villarreal", "ESP"},
}

func validateTeams(items []team) error {
	seen := make(map[string]bool, len(items))
	knownCountries := make(map[string]bool, len(countries))
	for _, country := range countries {
		knownCountries[country.shortName] = true
	}
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
		if !knownCountries[item.country] {
			return fmt.Errorf("team %q has unknown country %q", item.slug, item.country)
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
			INSERT INTO teams (name, short_name, slug, country, logo_url)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (slug) DO UPDATE
			SET name = EXCLUDED.name, short_name = EXCLUDED.short_name, country = EXCLUDED.country,
			    logo_url = EXCLUDED.logo_url
		`, item.name, item.shortName, item.slug, item.country, "/logos/teams/"+item.slug+".svg")
		if err != nil {
			return fmt.Errorf("upsert team %q: %w", item.slug, err)
		}
	}
	return nil
}
