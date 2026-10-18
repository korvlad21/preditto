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
	ID                   int64      `json:"id"`
	Name                 string     `json:"name"`
	TypeOfTournament     *string    `json:"type_of_tournament"`
	Status               *string    `json:"status"`
	StartsAt             *time.Time `json:"starts_at"`
	EndsAt               *time.Time `json:"ends_at"`
	CostAmount           *float64   `json:"cost_amount"`
	AmountCurrency       *string    `json:"amount_currency"`
	PrizeFirstPlace      *float64   `json:"prize_first_place"`
	PrizeSecondPlace     *float64   `json:"prize_second_place"`
	PrizeThirdPlace      *float64   `json:"prize_third_place"`
	CountParticipant     int64      `json:"count_participant"`
	IsSelectedTournament bool       `json:"is_selected_tournament"`
}
