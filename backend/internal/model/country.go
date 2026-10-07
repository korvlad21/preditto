package model

import "time"

type Country struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	CreatedAt time.Time `json:"created_at"`
}
