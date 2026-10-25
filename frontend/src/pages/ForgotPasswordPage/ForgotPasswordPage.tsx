import { Link } from 'react-router'
import BrandLogo from '../../components/BrandLogo/BrandLogo.tsx'
import '../../styles/auth.css'

// Placeholder until password recovery exists in the API.
function ForgotPasswordPage() {
  return (
    <section className="page auth-page">
      <div className="auth-card">
        <BrandLogo />
        <h1>Восстановление пароля</h1>
        <p className="auth-subtitle">
          Восстановление пароля пока недоступно
        </p>
        <p className="auth-switch">
          <Link to="/login">Вернуться ко входу</Link>
        </p>
      </div>
    </section>
  )
}

export default ForgotPasswordPage
