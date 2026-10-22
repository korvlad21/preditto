package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"preditto/internal/model"
)

var ErrTournamentNotFound = errors.New("tournament not found")

type TournamentRepository struct {
	db *sql.DB
}

func NewTournamentRepository(db *sql.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM tournaments WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check tournament existence: %w", err)
	}
	return exists, nil
}

func (r *TournamentRepository) GetAllTournaments(ctx context.Context, userID int64, status string, isParticipate *bool) ([]model.Tournament, error) {
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
	var conditions []string
	if status != "" {
		args = append(args, status)
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", len(args)))
	}
	if isParticipate != nil {
		args = append(args, userID)
		condition := fmt.Sprintf(`EXISTS (
			SELECT 1 FROM tournament_participants tp
			WHERE tp.tournament_id = t.id AND tp.user_id = $%d AND tp.status = $1
		)`, len(args))
		if !*isParticipate {
			condition = "NOT " + condition
		}
		conditions = append(conditions, condition)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
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
