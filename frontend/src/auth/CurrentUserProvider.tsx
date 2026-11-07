import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { ApiError } from '../api/client.ts'
import { logout } from '../api/session.ts'
import { hasTokens } from '../api/tokens.ts'
import { getCurrentUser } from '../api/user.ts'
import { CurrentUserContext, type CurrentUserState } from './currentUser.ts'

function initialState(): CurrentUserState {
  return hasTokens() ? { status: 'loading' } : { status: 'guest' }
}

// Single source of the signed-in user for the header, guards and pages.
function CurrentUserProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<CurrentUserState>(initialState)
  const [requestId, setRequestId] = useState(0)

  useEffect(() => {
    if (!hasTokens()) return
    let cancelled = false
    getCurrentUser()
      .then((user) => {
        if (!cancelled) setState({ status: 'ready', user })
      })
      .catch((error: unknown) => {
        if (cancelled) return
        // 401: the session has ended (refresh rejected) and tokens are already cleared.
        const ended = error instanceof ApiError && error.status === 401
        setState(ended ? { status: 'guest' } : { status: 'error' })
      })
    return () => {
      cancelled = true
    }
  }, [requestId])

  const reload = useCallback(() => {
    setState(initialState())
    setRequestId((id) => id + 1)
  }, [])

  const signOut = useCallback(async () => {
    await logout()
    setState({ status: 'guest' })
  }, [])

  const value = useMemo(() => ({ state, reload, signOut }), [state, reload, signOut])
  return <CurrentUserContext.Provider value={value}>{children}</CurrentUserContext.Provider>
}

export default CurrentUserProvider
