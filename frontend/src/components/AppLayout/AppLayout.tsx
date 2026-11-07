import { Outlet } from 'react-router'
import AppHeader from '../AppHeader/AppHeader.tsx'
import './AppLayout.css'

// Layout route of the main application (everything except the auth screens).
function AppLayout() {
  return (
    <div className="app-shell">
      <AppHeader />
      <main className="app-shell__main">
        <Outlet />
      </main>
    </div>
  )
}

export default AppLayout
