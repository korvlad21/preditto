package tournaments

type GetAllTournamentsRequest struct {
	Status        string `form:"status"`
	IsParticipate *bool  `form:"is_participate"`
}

type SelectTournamentRequest struct {
	ID *int64 `json:"id"`
}
