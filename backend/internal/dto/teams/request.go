package teams

type GetTeamRequest struct {
	ID int64 `json:"id" binding:"required,gt=0"`
}

type GetAllTeamsRequest struct {
	Country string `json:"country" form:"country"`
}
