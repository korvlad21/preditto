package model

import "time"

const (
	TournamentParticipantStatusActive  = "active"
	TournamentParticipantStatusNotPaid = "not_paid"
	TournamentParticipantStatusBlocked = "blocked"
)

type ParticipantInfo struct {
	ID       int64     `json:"id" format:"int64"`
	JoinedAt time.Time `json:"joined_at" format:"date-time"`
}
