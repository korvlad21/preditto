package model

import "time"

type MatchWithPrediction struct {
	Match      Match       `json:"match"`
	HomeTeam   MatchTeam   `json:"home_team"`
	AwayTeam   MatchTeam   `json:"away_team"`
	Prediction *Prediction `json:"prediction" extensions:"x-nullable"`
}

type Match struct {
	ID               int64      `json:"id" format:"int64"`
	HomeScore        *int16     `json:"home_score" extensions:"x-nullable"`
	AwayScore        *int16     `json:"away_score" extensions:"x-nullable"`
	ScheduledStartAt time.Time  `json:"scheduled_start_at" format:"date-time"`
	PredictionLockAt time.Time  `json:"prediction_lock_at" format:"date-time"`
	ActualStartAt    *time.Time `json:"actual_start_at" format:"date-time" extensions:"x-nullable"`
	FinishedAt       *time.Time `json:"finished_at" format:"date-time" extensions:"x-nullable"`
	Status           string     `json:"status" enums:"scheduled,live,finished,postponed,cancelled"`
}

type MatchTeam struct {
	ID        int64   `json:"id" format:"int64"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name" extensions:"x-nullable"`
	LogoURL   *string `json:"logo_url" extensions:"x-nullable"`
}

type Prediction struct {
	ID            int64   `json:"id" format:"int64"`
	HomeScore     int16   `json:"home_score"`
	AwayScore     int16   `json:"away_score"`
	ResultType    *string `json:"result_type" enums:"exact_score,outcome,missed" extensions:"x-nullable"`
	AwardedPoints *int16  `json:"awarded_points" extensions:"x-nullable"`
}
