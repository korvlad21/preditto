package model

import "time"

type Country struct {
	ID        int64     `json:"id" format:"int64"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
}
