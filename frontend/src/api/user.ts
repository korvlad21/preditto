import type { User, UserInfo } from '../types/user.ts'
import { ApiError } from './client.ts'
import { authPost } from './session.ts'
import { getAllTeams } from './teams.ts'

// Concurrent callers (e.g. StrictMode's double effect run, several components) share one
// in-flight request. The result is not cached after it settles: the profile can change.
let userInfoRequest: Promise<UserInfo> | null = null

// POST /api/user/get_user_info (backend/docs/user.md): account, profile and roles.
export function getUserInfo(): Promise<UserInfo> {
  if (!userInfoRequest) {
    userInfoRequest = authPost<UserInfo>('/api/user/get_user_info').finally(() => {
      userInfoRequest = null
    })
  }
  return userInfoRequest
}

async function resolveFavoriteTeam(teamId: number | null): Promise<User['favoriteTeam']> {
  if (teamId === null) return null
  try {
    const team = (await getAllTeams()).find((item) => item.id === teamId)
    return team ? { id: team.id, name: team.name, logoUrl: team.logo_url } : null
  } catch {
    // The profile is still usable without the team details.
    return undefined
  }
}

export interface ProfileUpdate {
  firstName: string
  lastName: string
  email: string
  favoriteTeamId: number | null
  // undefined: keep the current avatar; null: remove it; File: upload a new one.
  avatar?: File | null
}

// Error code thrown by updateProfile() until the backend endpoint exists.
export const PROFILE_UPDATE_NOT_IMPLEMENTED = 'NOT_IMPLEMENTED'

// TODO: the backend has no profile-update endpoint or avatar upload yet. Replace the body
// with the real call once the contract exists, e.g. a draft (unconfirmed) shape:
//   authPost<UserInfo>('/api/user/update_profile', {
//     first_name, last_name, email, favorite_team_id,
//   })
// plus a separate avatar upload. Until then nothing is sent and nothing is saved.
export function updateProfile(update: ProfileUpdate): Promise<void> {
  void update
  return Promise.reject(new ApiError(501, PROFILE_UPDATE_NOT_IMPLEMENTED))
}

// The signed-in user for the UI. get_user_info returns only favorite_team_id, so the
// team name and logo come from the (page-load cached) teams directory.
export async function getCurrentUser(): Promise<User> {
  const info = await getUserInfo()
  return {
    id: info.id,
    login: info.username,
    firstName: info.first_name ?? undefined,
    lastName: info.last_name ?? undefined,
    email: info.email,
    avatarUrl: info.avatar_url,
    createdAt: info.created_at,
    roles: info.roles ?? [],
    favoriteTeam: await resolveFavoriteTeam(info.favorite_team_id),
  }
}
