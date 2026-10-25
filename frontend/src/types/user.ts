// Mirrors backend internal/model/user.go. Field names keep the API's snake_case.

// model.UserStatusActive / model.UserStatusBlocked.
export type UserStatus = 'ACTIVE' | 'BLOCKED'

// model.User: `user` in POST /api/auth/login and /api/auth/register responses.
export interface AuthUser {
  id: number
  username: string
  email: string
  status: UserStatus
  first_name: string
  last_name: string
  favorite_team_id: number | null
}

// model.UserInfo: `data` of POST /api/user/get_user_info (account + profile).
export interface UserInfo {
  id: number
  username: string
  email: string
  status: UserStatus
  // Profile fields are nullable: the profile row may be missing.
  first_name: string | null
  last_name: string | null
  avatar_url: string | null
  favorite_team_id: number | null
  selected_tournament_id: number | null
  // RFC 3339 timestamp, e.g. "2026-10-25T12:00:00Z".
  created_at: string
}
