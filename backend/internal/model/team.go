package model

type Team struct {
	ID        int64   `json:"id" format:"int64"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name" extensions:"x-nullable"`
	Slug      string  `json:"slug"`
	Country   string  `json:"country"`
	LogoURL   *string `json:"logo_url" extensions:"x-nullable"`
}
