// Mirrors backend internal/model/team.go and country.go.

// model.Team: items of `data.teams` from POST /api/teams/get_all_teams.
export interface Team {
  id: number
  name: string
  short_name: string | null
  slug: string
  // Three-letter country code matching Country.short_name, e.g. "ENG".
  country: string
  // Public path such as "/logos/teams/arsenal.svg"; see docs/assets/team-logos.md.
  logo_url: string | null
}

// model.Country: items of `data.countries` from POST /api/teams/get_all_countries.
export interface Country {
  id: number
  name: string
  short_name: string
  created_at: string
}
