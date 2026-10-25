import { useState, type ChangeEvent, type FormEvent } from 'react'
import { Link } from 'react-router'
import '../../styles/auth.css'
import './RegisterPage.css'

// Field names and limits mirror backend RegisterRequest.
const initialForm = {
  username: '',
  first_name: '',
  last_name: '',
  email: '',
  password: '',
  password_confirmation: '',
}

function RegisterPage() {
  const [form, setForm] = useState(initialForm)
  const passwordsMismatch =
    form.password_confirmation !== '' &&
    form.password !== form.password_confirmation

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { name, value } = event.target
    setForm((prev) => ({ ...prev, [name]: value }))
  }

  // TODO: send to POST /api/auth/register once the API client is wired up.
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
  }

  return (
    <section className="page auth-page">
      <div className="auth-card">
        <h1>Регистрация</h1>
        <form className="auth-form" onSubmit={handleSubmit}>
          <label>
            Имя пользователя
            <input
              name="username"
              autoComplete="username"
              required
              pattern="[a-zA-Z0-9_.\-]{3,50}"
              title="3–50 символов: латиница, цифры, _ . -"
              value={form.username}
              onChange={handleChange}
            />
          </label>
          <label>
            Имя
            <input
              name="first_name"
              autoComplete="given-name"
              required
              maxLength={100}
              value={form.first_name}
              onChange={handleChange}
            />
          </label>
          <label>
            Фамилия
            <input
              name="last_name"
              autoComplete="family-name"
              required
              maxLength={100}
              value={form.last_name}
              onChange={handleChange}
            />
          </label>
          <label>
            Email
            <input
              name="email"
              type="email"
              autoComplete="email"
              required
              maxLength={255}
              value={form.email}
              onChange={handleChange}
            />
          </label>
          <label>
            Пароль
            <input
              name="password"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
              maxLength={72}
              value={form.password}
              onChange={handleChange}
            />
          </label>
          <label>
            Повторите пароль
            <input
              name="password_confirmation"
              type="password"
              autoComplete="new-password"
              required
              aria-invalid={passwordsMismatch}
              value={form.password_confirmation}
              onChange={handleChange}
            />
          </label>
          {passwordsMismatch && (
            <p className="form-error" role="alert">
              Пароли не совпадают
            </p>
          )}
          <button type="submit" disabled={passwordsMismatch}>
            Зарегистрироваться
          </button>
        </form>
        <p className="auth-switch">
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </p>
      </div>
    </section>
  )
}

export default RegisterPage
