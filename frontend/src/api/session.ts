// Authorized requests with access-token refresh (backend/docs/auth.md#refresh-rotation).
//
// Refresh tokens are single-use: a rotated token is rejected on reuse. All refreshes are
// therefore serialized — one in-flight request per tab, and a Web Lock across tabs that
// share localStorage. Inside the lock the stored pair is re-read: if another tab already
// rotated it, the new pair is used without a second refresh.

import { ApiError, apiPost } from './client.ts'
import { clearTokens, getTokens, replaceTokens, type TokenPair } from './tokens.ts'

const REFRESH_LOCK = 'preditto.auth.refresh'
// Refresh proactively when the access token expires within this margin.
const EXPIRY_MARGIN_MS = 30_000

let refreshInFlight: Promise<string | null> | null = null

function withLock<T>(task: () => Promise<T>): Promise<T> {
  if (typeof navigator !== 'undefined' && navigator.locks) {
    return navigator.locks.request(REFRESH_LOCK, task)
  }
  return task()
}

// Rotates the pair unless it already differs from `staleRefreshToken`.
// Resolves to the current access token, or null when the session has ended.
async function rotate(staleRefreshToken: string): Promise<string | null> {
  return withLock(async () => {
    const current = getTokens()
    if (!current) return null
    if (current.refresh_token !== staleRefreshToken) {
      return current.access_token
    }
    try {
      const pair = await apiPost<TokenPair>('/api/auth/refresh', {
        refresh_token: current.refresh_token,
      })
      replaceTokens(pair)
      return pair.access_token
    } catch (error) {
      // Rejected refresh credential (expired, revoked, reused) or blocked account ends the
      // session. Network and server errors keep the pair so a later attempt can succeed.
      if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
        clearTokens()
        return null
      }
      throw error
    }
  })
}

function refreshAccessToken(staleRefreshToken: string): Promise<string | null> {
  if (!refreshInFlight) {
    refreshInFlight = rotate(staleRefreshToken).finally(() => {
      refreshInFlight = null
    })
  }
  return refreshInFlight
}

async function validAccessToken(): Promise<string | null> {
  const tokens = getTokens()
  if (!tokens) return null
  if (tokens.access_expires_at - Date.now() > EXPIRY_MARGIN_MS) {
    return tokens.access_token
  }
  return refreshAccessToken(tokens.refresh_token)
}

function isUnauthorized(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401
}

// POST with `Authorization: Bearer <access token>`. On 401 the pair is refreshed once and
// the request retried. Throws ApiError(401, 'UNAUTHORIZED') when there is no valid session.
export async function authPost<T>(path: string, body?: unknown): Promise<T> {
  const accessToken = await validAccessToken()
  if (!accessToken) throw new ApiError(401, 'UNAUTHORIZED')

  const send = (token: string) => apiPost<T>(path, body, { Authorization: `Bearer ${token}` })
  try {
    return await send(accessToken)
  } catch (error) {
    if (!isUnauthorized(error)) throw error
  }

  // The access token was rejected (e.g. expired early or clock skew): refresh and retry once.
  const tokens = getTokens()
  if (!tokens) throw new ApiError(401, 'UNAUTHORIZED')
  const refreshed =
    tokens.access_token !== accessToken
      ? tokens.access_token
      : await refreshAccessToken(tokens.refresh_token)
  if (!refreshed) throw new ApiError(401, 'UNAUTHORIZED')
  return send(refreshed)
}

// Revokes the current refresh session on the backend and always clears local tokens.
// The access token stays valid server-side until it expires (no blacklist).
// Runs under the refresh lock so it revokes the latest pair, not one another tab is rotating.
export async function logout(): Promise<void> {
  try {
    await withLock(async () => {
      const tokens = getTokens()
      clearTokens()
      if (tokens) {
        await apiPost<void>('/api/auth/logout', { refresh_token: tokens.refresh_token })
      }
    })
  } catch {
    // Local sign-out already happened; the server session expires on its own.
    clearTokens()
  }
}
