package repository

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type CountryRepository struct {
	db *sql.DB
}

func NewCountryRepository(db *sql.DB) *CountryRepository {
	return &CountryRepository{db: db}
}

func (r *CountryRepository) GetAllCountries(ctx context.Context) ([]model.Country, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, short_name, created_at FROM countries ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query countries: %w", err)
	}
	defer rows.Close()

	countries := make([]model.Country, 0)
	for rows.Next() {
		var country model.Country
		if err := rows.Scan(&country.ID, &country.Name, &country.ShortName, &country.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan country: %w", err)
		}
		countries = append(countries, country)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read countries: %w", err)
	}
	return countries, nil
}
