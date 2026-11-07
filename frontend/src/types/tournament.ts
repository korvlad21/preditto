// API shapes follow the Swagger schemas 1:1 (snake_case):
// model.Tournament, tournaments.GetAllTournamentsRequest, tournaments.SelectTournamentRequest.

// model.Tournament: items of `data.tournaments` from POST /api/tournaments/get_all_tournaments.
export interface TournamentDto {
  id: number
  name: string
  type_of_tournament: string | null
  // Free string in the schema; backend constants: planned, active, finished, cancelled.
  status: string | null
  // RFC 3339 date-time.
  starts_at: string | null
  ends_at: string | null
  cost_amount: number | null
  amount_currency: string | null
  // Prize percentages for places 1–3.
  prize_first_place: number | null
  prize_second_place: number | null
  prize_third_place: number | null
  count_participant: number
  is_selected_tournament: boolean
}

// tournaments.GetAllTournamentsRequest; omitted fields apply no filter.
export interface GetAllTournamentsRequest {
  // Omitted, empty or "all": no filter. Known values: planned, active, finished, cancelled.
  status?: string
  // true: active participation; false: no active participation; omitted/null: no filter.
  is_participate?: boolean | null
}

// tournaments.SelectTournamentRequest.
export interface SelectTournamentRequest {
  id: number
}

export type TournamentStatus = 'active' | 'upcoming' | 'finished'

// UI model built by toTournament(). Optional fields are absent when the API has no data
// for them; the card then omits the corresponding cell.
export interface Tournament {
  id: number
  name: string
  type?: string
  status: TournamentStatus
  startDate?: string
  endDate?: string
  participantsCount: number
  // Not provided by the API.
  participantsLimit?: number
  entryFee?: number
  currency?: string
  // Not provided by the API (only per-place percentages are).
  prizePool?: number
  // Percent of the pool per place.
  prizes?: { place: number; percent: number }[]
  isCurrent: boolean
}
