import { Navigate, Outlet } from 'react-router'
import { useCurrentUser } from './currentUser.ts'

// Layout route for guest-only pages (login). A signed-in user goes to the home page.
// While the user loads nothing is rendered to avoid flashing the form; if loading
// failed, the form stays available so the user can sign in again.
function GuestOnly() {
  const { state } = useCurrentUser()

  if (state.status === 'ready') return <Navigate to="/" replace />
  if (state.status === 'loading') return null
  return <Outlet />
}

export default GuestOnly
