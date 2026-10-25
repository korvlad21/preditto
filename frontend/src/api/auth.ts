import type { AuthUser } from '../types/user.ts'

// Shapes mirror backend dto/auth (LoginRequest, AuthResponse).
export interface LoginPayload {
  login: string
  password: string
  // Not part of backend LoginRequest yet; the API ignores unknown fields.
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

// TODO: replace the mock with `apiPost<AuthResponse>('/api/auth/register', request)`
// from ./client.ts and upload payload.avatar once the backend accepts it.
// Errors arrive as ApiError: 409 USERNAME_ALREADY_EXISTS / EMAIL_ALREADY_EXISTS, 400 INVALID_REQUEST.
export function register(payload: RegisterPayload): Promise<AuthResponse> {
  const request = toRegisterRequest(payload)
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        access_token: 'mock-access-token',
        refresh_token: 'mock-refresh-token',
        token_type: 'Bearer',
        expires_in: 900,
        user: {
          id: 0,
          username: request.username,
          email: request.email,
          status: 'ACTIVE',
          first_name: request.first_name,
          last_name: request.last_name,
          favorite_team_id: request.favorite_team_id ?? null,
        },
      })
    }, 800)
  })
}

// TODO: send POST /api/auth/login once the API client (proxy or CORS) is wired up.
export function login(payload: LoginPayload): Promise<AuthResponse> {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve({
        access_token: 'mock-access-token',
        refresh_token: 'mock-refresh-token',
        token_type: 'Bearer',
        expires_in: 900,
        user: {
          id: 0,
          username: payload.login,
          email: '',
          status: 'ACTIVE',
          first_name: '',
          last_name: '',
          favorite_team_id: null,
        },
      })
    }, 800)
  })
}
