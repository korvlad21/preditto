package development

import (
	"context"
	"database/sql"
	"fmt"
)

type country struct {
	name      string
	shortName string
}

var countries = []country{
	{"England", "ENG"},
	{"Spain", "ESP"},
	{"Germany", "GER"},
	{"Italy", "ITA"},
	{"France", "FRA"},
	{"Portugal", "POR"},
	{"Netherlands", "NED"},
	{"Turkey", "TUR"},
	{"Norway", "NOR"},
	{"Belgium", "BEL"},
	{"Greece", "GRE"},
	{"Austria", "AUT"},
	{"Ukraine", "UKR"},
	{"Czech Republic", "CZE"},
	{"Slovakia", "SVK"},
	{"Azerbaijan", "AZE"},
}

func validateCountries(items []country) error {
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if item.name == "" {
			return fmt.Errorf("country name must be set")
		}
		if len(item.shortName) != 3 {
			return fmt.Errorf("country %q short_name must contain exactly three uppercase ASCII letters", item.name)
		}
		for _, letter := range item.shortName {
			if letter < 'A' || letter > 'Z' {
				return fmt.Errorf("country %q short_name must contain exactly three uppercase ASCII letters", item.name)
			}
		}
		if seen[item.shortName] {
			return fmt.Errorf("duplicate country short_name %q", item.shortName)
		}
		seen[item.shortName] = true
	}
	return nil
}

func seedCountries(ctx context.Context, tx *sql.Tx) error {
	if err := validateCountries(countries); err != nil {
		return err
	}
	for _, item := range countries {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO countries (name, short_name) VALUES ($1, $2)
			ON CONFLICT (short_name) DO UPDATE SET name = EXCLUDED.name
		`, item.name, item.shortName)
		if err != nil {
			return fmt.Errorf("upsert country %q: %w", item.shortName, err)
		}
	}
	return nil
}
