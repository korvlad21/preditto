import { Navigate, Outlet, useLocation } from 'react-router'
import { useCurrentUser } from './currentUser.ts'

// Layout route for signed-in pages. Guests go to /login, which returns them here afterwards.
// While the user loads (or failed to load) the page renders and shows its own state.
function RequireAuth() {
  const { state } = useCurrentUser()
  const location = useLocation()

  if (state.status === 'guest') {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  return <Outlet />
}

export default RequireAuth
