import type { Country, Team } from '../types/team.ts'
import { apiPost } from './client.ts'

// Reference data rarely changes: share one request per page load, retry after failure.
let teamsRequest: Promise<Team[]> | null = null
let countriesRequest: Promise<Country[]> | null = null

export function getAllTeams(): Promise<Team[]> {
  if (!teamsRequest) {
    teamsRequest = apiPost<{ teams: Team[] }>('/api/teams/get_all_teams', {}).then(
      (data) => data.teams,
    )
    teamsRequest.catch(() => {
      teamsRequest = null
    })
  }
  return teamsRequest
}

export function getAllCountries(): Promise<Country[]> {
  if (!countriesRequest) {
    countriesRequest = apiPost<{ countries: Country[] }>(
      '/api/teams/get_all_countries',
    ).then((data) => data.countries)
    countriesRequest.catch(() => {
      countriesRequest = null
    })
  }
  return countriesRequest
}
