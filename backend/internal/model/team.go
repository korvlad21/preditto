package model

type Team struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name"`
	Slug      string  `json:"slug"`
	Country   string  `json:"country"`
	LogoURL   *string `json:"logo_url"`
}
