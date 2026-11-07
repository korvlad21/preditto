import { Navigate, Outlet, Route, Routes } from 'react-router'
import CurrentUserProvider from './auth/CurrentUserProvider.tsx'
import RequireAuth from './auth/RequireAuth.tsx'
import AppLayout from './components/AppLayout/AppLayout.tsx'
import ForgotPasswordPage from './pages/ForgotPasswordPage/ForgotPasswordPage.tsx'
import HomePage from './pages/HomePage/HomePage.tsx'
import LoginPage from './pages/LoginPage/LoginPage.tsx'
import ProfileEditPage from './pages/ProfileEditPage/ProfileEditPage.tsx'
import ProfilePage from './pages/ProfilePage/ProfilePage.tsx'
import RegisterPage from './pages/RegisterPage/RegisterPage.tsx'
import TournamentsPage from './pages/TournamentsPage/TournamentsPage.tsx'
import './App.css'

// Auth screens: full-screen background, no app header.
function AuthLayout() {
  return (
    <main className="app-main">
      <Outlet />
    </main>
  )
}

function App() {
  return (
    <CurrentUserProvider>
      <Routes>
        <Route element={<AuthLayout />}>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        </Route>

        <Route element={<AppLayout />}>
          <Route path="/" element={<HomePage />} />
          <Route element={<RequireAuth />}>
            <Route path="/profile" element={<ProfilePage />} />
            <Route path="/profile/edit" element={<ProfileEditPage />} />
            <Route path="/tournaments" element={<TournamentsPage />} />
          </Route>
        </Route>

        {/* Former home path. */}
        <Route path="/home" element={<Navigate to="/" replace />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </CurrentUserProvider>
  )
}

export default App
