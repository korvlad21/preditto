package tournaments

type GetAllTournamentsRequest struct {
	// Omitted, empty or all disables filtering. Known values: planned, active, finished, cancelled. Unknown strings return no results.
	Status string `json:"status" example:"all"`
	// True: active participation; false: no active participation; omitted/null: no filter.
	IsParticipate *bool `json:"is_participate" extensions:"x-nullable"`
}

type SelectTournamentRequest struct {
	// Required non-null positive signed 64-bit tournament ID.
	ID *int64 `json:"id" validate:"required" minimum:"1" format:"int64" example:"1"`
}
