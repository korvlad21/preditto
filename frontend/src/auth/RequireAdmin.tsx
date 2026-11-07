import { Navigate, Outlet, useLocation } from 'react-router'
import { isAdmin, useCurrentUser } from './currentUser.ts'

// Layout route for admin-only pages. Roles come from get_user_info (user_roles table).
// Guests go to /login; signed-in non-admins go to the home page.
function RequireAdmin() {
  const { state } = useCurrentUser()
  const location = useLocation()

  if (state.status === 'guest') {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  if (state.status === 'loading') return null
  if (state.status === 'ready' && isAdmin(state.user)) return <Outlet />
  return <Navigate to="/" replace />
}

export default RequireAdmin
