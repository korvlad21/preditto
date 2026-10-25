// Package apidocs describes anonymous HTTP envelopes for Swagger generation.
// Business payloads reuse the actual DTOs/models; these types do not serialize responses.
package apidocs

import "preditto/internal/model"

type Success struct {
	Success bool `json:"success" example:"true" enums:"true" validate:"required"`
	Data    any  `json:"data" validate:"required"`
}

type EmptySuccess struct {
	Success bool `json:"success" example:"true" enums:"true" validate:"required"`
}

type Error struct {
	Success bool   `json:"success" example:"false" enums:"false" validate:"required"`
	Error   string `json:"error" example:"INVALID_REQUEST" validate:"required"`
}

type TeamData struct {
	Team model.Team `json:"team" validate:"required"`
}

type FavoriteTeamData struct {
	// Null when the current profile has no favorite team.
	Team *model.Team `json:"team" extensions:"x-nullable" validate:"required"`
}

type TeamsData struct {
	// Empty results are [], never null.
	Teams []model.Team `json:"teams" validate:"required"`
}

type CountriesData struct {
	Countries []model.Country `json:"countries" validate:"required"`
}

type TournamentsData struct {
	Tournaments []model.Tournament `json:"tournaments" validate:"required"`
}

type Health struct {
	Status string `json:"status" example:"ok" validate:"required"`
}

type Hello struct {
	Message string `json:"message" example:"Hello from Go!" validate:"required"`
}
