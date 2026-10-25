package model

import "time"

const (
	CurrencyUSD                   = "USD"
	CurrencyGEL                   = "GEL"
	TournamentTypeChampionsLeague = "Champions league"
	TournamentTypeLaLiga          = "La Liga"
)

const (
	TournamentStatusPlanned   = "planned"
	TournamentStatusActive    = "active"
	TournamentStatusFinished  = "finished"
	TournamentStatusCancelled = "cancelled"
)

type Tournament struct {
	ID                   int64      `json:"id" format:"int64"`
	Name                 string     `json:"name"`
	TypeOfTournament     *string    `json:"type_of_tournament" extensions:"x-nullable"`
	Status               *string    `json:"status" extensions:"x-nullable"`
	StartsAt             *time.Time `json:"starts_at" extensions:"x-nullable" format:"date-time"`
	EndsAt               *time.Time `json:"ends_at" extensions:"x-nullable" format:"date-time"`
	CostAmount           *float64   `json:"cost_amount" extensions:"x-nullable"`
	AmountCurrency       *string    `json:"amount_currency" extensions:"x-nullable"`
	PrizeFirstPlace      *float64   `json:"prize_first_place" extensions:"x-nullable"`
	PrizeSecondPlace     *float64   `json:"prize_second_place" extensions:"x-nullable"`
	PrizeThirdPlace      *float64   `json:"prize_third_place" extensions:"x-nullable"`
	CountParticipant     int64      `json:"count_participant" format:"int64"`
	IsSelectedTournament bool       `json:"is_selected_tournament"`
}
