package tournaments

type GetAllTournamentsRequest struct {
	Status        string `json:"status"`
	IsParticipate *bool  `json:"is_participate"`
}

type SelectTournamentRequest struct {
	ID *int64 `json:"id"`
}
