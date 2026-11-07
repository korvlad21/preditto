import type {
  GetAllTournamentsRequest,
  SelectTournamentRequest,
  Tournament,
  TournamentDto,
  TournamentStatus,
} from '../types/tournament.ts'
import { ApiError } from './client.ts'
import { authPost } from './session.ts'

function parseDate(value: string | null): Date | null {
  if (!value) return null
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

// Without a usable status the phase is derived from the dates.
function statusFromDates(dto: TournamentDto, now: Date): TournamentStatus {
  const start = parseDate(dto.starts_at)
  const end = parseDate(dto.ends_at)
  if (end && end < now) return 'finished'
  if (start && start <= now) return 'active'
  return 'upcoming'
}

function toStatus(dto: TournamentDto, now: Date): TournamentStatus {
  switch (dto.status?.toLowerCase()) {
    case 'active':
      return 'active'
    case 'planned':
      return 'upcoming'
    // A cancelled tournament can no longer be played; the UI has no separate phase for it.
    case 'finished':
    case 'cancelled':
      return 'finished'
    default:
      return statusFromDates(dto, now)
  }
}

export function toTournament(dto: TournamentDto, now: Date = new Date()): Tournament {
  const prizes = [dto.prize_first_place, dto.prize_second_place, dto.prize_third_place]
    .map((percent, index) => ({ place: index + 1, percent }))
    .filter((prize): prize is { place: number; percent: number } => prize.percent !== null)

  return {
    id: dto.id,
    name: dto.name,
    type: dto.type_of_tournament ?? undefined,
    status: toStatus(dto, now),
    startDate: dto.starts_at ?? undefined,
    endDate: dto.ends_at ?? undefined,
    participantsCount: dto.count_participant,
    entryFee: dto.cost_amount ?? undefined,
    currency: dto.amount_currency ?? undefined,
    prizes: prizes.length > 0 ? prizes : undefined,
    isCurrent: dto.is_selected_tournament,
  }
}

// UI status → API status filter value.
export const API_STATUS: Record<TournamentStatus, string> = {
  active: 'active',
  upcoming: 'planned',
  finished: 'finished',
}

// POST /api/tournaments/get_all_tournaments (backend/docs/tournaments.md), newest first.
export async function getAllTournaments(
  request: GetAllTournamentsRequest = {},
): Promise<Tournament[]> {
  const data = await authPost<{ tournaments: TournamentDto[] }>(
    '/api/tournaments/get_all_tournaments',
    request,
  )
  const now = new Date()
  return data.tournaments.map((dto) => toTournament(dto, now))
}

// POST /api/tournaments/select_tournament: makes the tournament current for the user.
// Selection is not membership or payment. Errors arrive as ApiError with the API code:
// 404 TOURNAMENT_NOT_FOUND / USER_INFO_NOT_FOUND, 400 INVALID_TOURNAMENT_ID /
// TOURNAMENT_ID_REQUIRED, 500 TOURNAMENT_UPDATE_FAILED.
export function selectTournament(id: number): Promise<void> {
  const body: SelectTournamentRequest = { id }
  return authPost<void>('/api/tournaments/select_tournament', body)
}

// Error code thrown by clearSelectedTournament() until the backend exposes a reset.
export const CLEAR_SELECTION_NOT_IMPLEMENTED = 'NOT_IMPLEMENTED'

// TODO: the HTTP API cannot clear the selection yet: select_tournament requires a positive
// id (the repository already supports writing NULL). Replace the body with the real call
// once an endpoint or contract exists. Until then nothing is sent and nothing changes.
export function clearSelectedTournament(): Promise<void> {
  return Promise.reject(new ApiError(501, CLEAR_SELECTION_NOT_IMPLEMENTED))
}
