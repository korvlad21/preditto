import { Navigate, Route, Routes } from 'react-router'
import ForgotPasswordPage from './pages/ForgotPasswordPage/ForgotPasswordPage.tsx'
import HomePage from './pages/HomePage/HomePage.tsx'
import LoginPage from './pages/LoginPage/LoginPage.tsx'
import RegisterPage from './pages/RegisterPage/RegisterPage.tsx'
import './App.css'

function App() {
  return (
    <main className="app-main">
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </main>
  )
}

export default App
