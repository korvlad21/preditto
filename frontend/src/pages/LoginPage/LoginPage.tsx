import { useRef, useState, type FormEvent } from 'react'
import { Link } from 'react-router'
import { login as loginRequest } from '../../api/auth.ts'
import BrandLogo from '../../components/BrandLogo/BrandLogo.tsx'
import {
  ArrowRightIcon,
  CheckIcon,
  EyeIcon,
  EyeOffIcon,
  LockIcon,
  UserIcon,
} from '../../components/Icons/Icons.tsx'
import '../../styles/auth.css'
import './LoginPage.css'

// Temporary dev-only prefill with an active account from backend/seeds/development.
const devCredentials = import.meta.env.DEV
  ? { login: 'hanna_stoma', password: '12345678' }
  : { login: '', password: '' }

type FieldErrors = { login?: string; password?: string }

function validate(login: string, password: string): FieldErrors {
  const errors: FieldErrors = {}
  if (login.trim() === '') {
    errors.login = 'Введите логин или email'
  }
  if (password === '') {
    errors.password = 'Введите пароль'
  } else if (password.length < 8) {
    errors.password = 'Минимум 8 символов'
  }
  return errors
}

function LoginPage() {
  const [login, setLogin] = useState(devCredentials.login)
  const [password, setPassword] = useState(devCredentials.password)
  const [remember, setRemember] = useState(true)
  const [showPassword, setShowPassword] = useState(false)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [serverError, setServerError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const loginRef = useRef<HTMLInputElement>(null)
  const passwordRef = useRef<HTMLInputElement>(null)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isSubmitting) return

    const nextErrors = validate(login, password)
    setErrors(nextErrors)
    if (nextErrors.login) {
      loginRef.current?.focus()
      return
    }
    if (nextErrors.password) {
      passwordRef.current?.focus()
      return
    }

    setIsSubmitting(true)
    try {
      // TODO: POST /api/auth/login { login, password, remember }; store tokens and redirect.
      await loginRequest({ login, password, remember })
    } catch (error) {
      setServerError(
        error instanceof Error ? error.message : 'Не удалось войти. Попробуйте ещё раз',
      )
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <section className="page auth-page">
      <div className="auth-card">
        <BrandLogo />
        <h1>Авторизация</h1>
        <p className="auth-subtitle">
          Войдите в свой аккаунт, чтобы продолжить пользоваться платформой
        </p>

        <form className="auth-form" onSubmit={handleSubmit} noValidate>
          <div className="auth-field">
            <label htmlFor="login" className="visually-hidden">
              Логин или email
            </label>
            <div className="auth-field__control">
              <UserIcon className="auth-field__icon" />
              <input
                ref={loginRef}
                id="login"
                name="login"
                autoComplete="username"
                required
                maxLength={255}
                placeholder="Логин или email"
                value={login}
                aria-invalid={errors.login ? true : undefined}
                aria-describedby={errors.login ? 'login-error' : undefined}
                onChange={(event) => {
                  setLogin(event.target.value)
                  setErrors((prev) => ({ ...prev, login: undefined }))
                  setServerError('')
                }}
              />
            </div>
            {errors.login && (
              <p className="form-error" id="login-error">
                {errors.login}
              </p>
            )}
          </div>

          <div className="auth-field">
            <label htmlFor="password" className="visually-hidden">
              Пароль
            </label>
            <div className="auth-field__control">
              <LockIcon className="auth-field__icon" />
              <input
                ref={passwordRef}
                id="password"
                name="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                required
                placeholder="Пароль"
                value={password}
                aria-invalid={errors.password ? true : undefined}
                aria-describedby={errors.password ? 'password-error' : undefined}
                onChange={(event) => {
                  setPassword(event.target.value)
                  setErrors((prev) => ({ ...prev, password: undefined }))
                  setServerError('')
                }}
              />
              <button
                type="button"
                className="auth-field__toggle"
                aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
                aria-pressed={showPassword}
                aria-controls="password"
                onClick={() => setShowPassword((prev) => !prev)}
              >
                {showPassword ? <EyeIcon /> : <EyeOffIcon />}
              </button>
            </div>
            {errors.password && (
              <p className="form-error" id="password-error">
                {errors.password}
              </p>
            )}
          </div>

          <div className="auth-options">
            <label className="auth-remember">
              <input
                type="checkbox"
                className="visually-hidden"
                name="remember"
                checked={remember}
                onChange={(event) => setRemember(event.target.checked)}
              />
              <span className="auth-checkbox" aria-hidden="true">
                <CheckIcon />
              </span>
              Запомнить меня
            </label>
            <Link to="/forgot-password" className="auth-forgot">
              Забыли пароль?
            </Link>
          </div>

          {serverError && (
            <div className="auth-alert" role="alert">
              {serverError}
            </div>
          )}

          <button
            type="submit"
            className="auth-submit"
            disabled={isSubmitting}
            aria-busy={isSubmitting}
          >
            {isSubmitting ? (
              <>
                <span className="auth-spinner" aria-hidden="true" />
                <span className="auth-loading-text">Входим…</span>
              </>
            ) : (
              <>
                Войти
                <ArrowRightIcon className="auth-submit__arrow" />
              </>
            )}
          </button>
        </form>

        <p className="auth-switch">
          Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
        </p>
      </div>
    </section>
  )
}

export default LoginPage
