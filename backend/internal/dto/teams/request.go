package teams

type GetTeamRequest struct {
	// Required positive signed 64-bit team ID.
	ID int64 `json:"id" binding:"required,gt=0" minimum:"1" format:"int64" example:"1"`
}

type GetAllTeamsRequest struct {
	// Omitted, empty or ALL disables filtering; otherwise exact case-sensitive country code.
	Country string `json:"country" example:"ALL"`
}
