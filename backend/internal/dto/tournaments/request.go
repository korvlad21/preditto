package tournaments

type GetAllTournamentsRequest struct {
	Status string `form:"status"`
}

type SelectTournamentRequest struct {
	ID *int64 `json:"id"`
}
