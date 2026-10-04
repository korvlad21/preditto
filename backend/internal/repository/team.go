package repository

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) GetAllTeams(ctx context.Context, country string) ([]model.Team, error) {
	query := `SELECT id, name, short_name, slug, country, logo_url FROM teams`
	var args []any
	if country != "" {
		query += ` WHERE country = $1`
		args = append(args, country)
	}
	query += ` ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query teams: %w", err)
	}
	defer rows.Close()

	teams := make([]model.Team, 0)
	for rows.Next() {
		var team model.Team
		if err := rows.Scan(&team.ID, &team.Name, &team.ShortName, &team.Slug, &team.Country, &team.LogoURL); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read teams: %w", err)
	}
	return teams, nil
}
