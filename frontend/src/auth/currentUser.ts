import { createContext, useContext } from 'react'
import type { User } from '../types/user.ts'

export type CurrentUserState =
  | { status: 'loading' }
  | { status: 'guest' }
  | { status: 'ready'; user: User }
  | { status: 'error' }

export interface CurrentUserContextValue {
  state: CurrentUserState
  // Reloads the user from the API (after login, or to retry after an error).
  reload: () => void
  // Revokes the session (POST /api/auth/logout), clears tokens and switches to guest.
  signOut: () => Promise<void>
}

export const CurrentUserContext = createContext<CurrentUserContextValue | null>(null)

export function useCurrentUser(): CurrentUserContextValue {
  const value = useContext(CurrentUserContext)
  if (!value) throw new Error('useCurrentUser must be used inside CurrentUserProvider')
  return value
}

// First and last name when present, otherwise the login.
export function displayName(user: User): string {
  return [user.firstName, user.lastName].filter(Boolean).join(' ') || user.login
}
