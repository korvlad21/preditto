package repository

import (
	"context"
	"database/sql"
	"fmt"

	"preditto/internal/model"
)

type TournamentRepository struct {
	db *sql.DB
}

func NewTournamentRepository(db *sql.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) GetAllTournaments(ctx context.Context, status string) ([]model.Tournament, error) {
	query := `
		SELECT t.id, t.name, t.type_of_tournament, t.status, t.starts_at, t.ends_at,
		       t.cost_amount, t.amount_currency, first.percentage, second.percentage,
		       third.percentage, COALESCE(participants.count_participant, 0)
		FROM tournaments t
		LEFT JOIN tournament_prize_rules first ON first.tournament_id = t.id AND first.place = 1
		LEFT JOIN tournament_prize_rules second ON second.tournament_id = t.id AND second.place = 2
		LEFT JOIN tournament_prize_rules third ON third.tournament_id = t.id AND third.place = 3
		LEFT JOIN (
			SELECT tournament_id, COUNT(DISTINCT user_id) AS count_participant
			FROM tournament_participants
			WHERE status = $1
			GROUP BY tournament_id
		) participants ON participants.tournament_id = t.id`
	args := []any{model.TournamentParticipantStatusActive}
	if status != "" {
		query += ` WHERE t.status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY t.created_at DESC, t.id DESC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tournaments: %w", err)
	}
	defer rows.Close()

	tournaments := make([]model.Tournament, 0)
	for rows.Next() {
		var tournament model.Tournament
		if err := rows.Scan(
			&tournament.ID, &tournament.Name, &tournament.TypeOfTournament, &tournament.Status,
			&tournament.StartsAt, &tournament.EndsAt, &tournament.CostAmount, &tournament.AmountCurrency,
			&tournament.PrizeFirstPlace, &tournament.PrizeSecondPlace, &tournament.PrizeThirdPlace,
			&tournament.CountParticipant,
		); err != nil {
			return nil, fmt.Errorf("scan tournament: %w", err)
		}
		tournaments = append(tournaments, tournament)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read tournaments: %w", err)
	}
	return tournaments, nil
}
