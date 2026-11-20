package model

import "time"

// TournamentStage includes prediction progress for the requesting participant.
type TournamentStage struct {
	ID                int64      `json:"id" format:"int64"`
	TournamentID      int64      `json:"tournament_id" format:"int64"`
	Code              string     `json:"code"`
	Name              string     `json:"name"`
	StageKind         string     `json:"stage_kind" enums:"group,round,play-off"`
	SequenceNo        int16      `json:"sequence_no"`
	Status            string     `json:"status" enums:"active,planned,finished,cancelled"`
	PredictionsOpenAt *time.Time `json:"predictions_open_at" format:"date-time" extensions:"x-nullable"`
	OutcomePoints     int16      `json:"outcome_points"`
	ExactScorePoints  int16      `json:"exact_score_points"`
	// Total matches in this stage, across all match statuses.
	CountMatches int64 `json:"count_matches" format:"int64"`
	// Matches in this stage with a prediction from the current participant.
	CountPredictions int64 `json:"count_predictions" format:"int64"`
}
