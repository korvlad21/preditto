package predictions

type GetStageRequest struct {
	// Required positive signed 64-bit tournament stage ID.
	ID int64 `json:"id" binding:"required,gt=0" minimum:"1" format:"int64" example:"1"`
}
