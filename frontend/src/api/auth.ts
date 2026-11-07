import type { AuthUser } from '../types/user.ts'
import { apiPost } from './client.ts'
import { authPost } from './session.ts'

// Shapes mirror backend dto/auth (LoginRequest, AuthResponse).
export interface LoginPayload {
  login: string
  password: string
  // Sent as remember_me: true selects the long refresh session, false the short one.
  remember: boolean
}

export interface AuthResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: AuthUser
}

export interface RegisterPayload {
  login: string
  firstName: string
  lastName: string
  email: string
  password: string
  confirmPassword: string
  favoriteTeamId: number | null
  // Backend registration has no avatar upload yet; kept for the future endpoint.
  avatar: File | null
}

// Body of backend RegisterRequest (JSON, snake_case).
interface RegisterRequest {
  username: string
  first_name: string
  last_name: string
  email: string
  password: string
  password_confirmation: string
  favorite_team_id?: number
}

function toRegisterRequest(payload: RegisterPayload): RegisterRequest {
  return {
    username: payload.login,
    first_name: payload.firstName.trim(),
    last_name: payload.lastName.trim(),
    email: payload.email.trim(),
    password: payload.password,
    password_confirmation: payload.confirmPassword,
    ...(payload.favoriteTeamId !== null && { favorite_team_id: payload.favoriteTeamId }),
  }
}

// POST /api/auth/register (201). Only admins register users, so the request carries the
// admin's access token; the created account's returned tokens must not be stored.
// The avatar is not sent: the backend has no upload endpoint yet.
// Errors arrive as ApiError: 409 USERNAME_ALREADY_EXISTS / EMAIL_ALREADY_EXISTS,
// 400 INVALID_REQUEST / FAVORITE_TEAM_NOT_FOUND, 401 UNAUTHORIZED, 403 FORBIDDEN.
export function register(payload: RegisterPayload): Promise<AuthResponse> {
  return authPost<AuthResponse>('/api/auth/register', toRegisterRequest(payload))
}

// POST /api/auth/login (200). Errors arrive as ApiError: 401 INVALID_CREDENTIALS,
// 403 USER_BLOCKED, 400 INVALID_REQUEST.
export function login(payload: LoginPayload): Promise<AuthResponse> {
  return apiPost<AuthResponse>('/api/auth/login', {
    login: payload.login,
    password: payload.password,
    remember_me: payload.remember,
  })
}