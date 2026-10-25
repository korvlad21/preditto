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
