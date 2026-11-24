package predictions

type GetStageRequest struct {
	// Required positive signed 64-bit tournament stage ID.
	ID int64 `json:"id" binding:"required,gt=0" minimum:"1" format:"int64" example:"1"`
}

type SavePredictionRequest struct {
	MatchID   int64  `json:"match_id" binding:"required,gt=0" minimum:"1" format:"int64" example:"1"`
	HomeScore *int16 `json:"home_score" binding:"required,gte=0" minimum:"0" maximum:"32767" example:"2"`
	AwayScore *int16 `json:"away_score" binding:"required,gte=0" minimum:"0" maximum:"32767" example:"0"`
}
