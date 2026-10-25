import { useRef, useState, type ChangeEvent, type FormEvent, type ReactNode } from 'react'
import { Link, useNavigate } from 'react-router'
import { register } from '../../api/auth.ts'
import { ApiError } from '../../api/client.ts'
import AvatarUpload from '../../components/AvatarUpload/AvatarUpload.tsx'
import BrandLogo from '../../components/BrandLogo/BrandLogo.tsx'
import FavoriteTeamButton from '../../components/FavoriteTeamButton/FavoriteTeamButton.tsx'
import {
  ArrowRightIcon,
  EyeIcon,
  EyeOffIcon,
  LockIcon,
  MailIcon,
  UserIcon,
} from '../../components/Icons/Icons.tsx'
import TeamPickerModal from '../../components/TeamPickerModal/TeamPickerModal.tsx'
import type { Team } from '../../types/team.ts'
import '../../styles/auth.css'
import './RegisterPage.css'

type Field = 'login' | 'lastName' | 'firstName' | 'email' | 'password' | 'confirmPassword'
type FormValues = Record<Field, string>
type FieldErrors = Partial<Record<Field, string>>

// Order defines both rendering and focus after a failed submit.
const FIELD_ORDER: Field[] = ['login', 'lastName', 'firstName', 'email', 'password', 'confirmPassword']

// Temporary dev-only prefill; a fresh username/email so the real API would accept it.
const initialValues: FormValues = import.meta.env.DEV
  ? {
      login: 'new_user',
      lastName: 'Иванов',
      firstName: 'Иван',
      email: 'new.user@example.com',
      password: '12345678',
      confirmPassword: '12345678',
    }
  : { login: '', lastName: '', firstName: '', email: '', password: '', confirmPassword: '' }

// Limits mirror backend RegisterRequest (dto/auth/request.go).
const USERNAME_PATTERN = /^[a-zA-Z0-9_.-]{3,50}$/
const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function validate(values: FormValues): FieldErrors {
  const errors: FieldErrors = {}

  if (values.login === '') {
    errors.login = 'Введите логин'
  } else if (!USERNAME_PATTERN.test(values.login)) {
    errors.login = 'Логин: 3–50 символов, латиница, цифры, . _ -'
  }

  if (values.lastName.trim() === '') {
    errors.lastName = 'Введите фамилию'
  } else if (values.lastName.length > 100) {
    errors.lastName = 'Не более 100 символов'
  }

  if (values.firstName.trim() === '') {
    errors.firstName = 'Введите имя'
  } else if (values.firstName.length > 100) {
    errors.firstName = 'Не более 100 символов'
  }

  const email = values.email.trim()
  if (email === '') {
    errors.email = 'Введите email'
  } else if (!EMAIL_PATTERN.test(email) || email.length > 255) {
    errors.email = 'Введите корректный email'
  }

  if (values.password === '') {
    errors.password = 'Введите пароль'
  } else if (values.password.length < 8) {
    errors.password = 'Минимум 8 символов'
  } else if (values.password.length > 72) {
    errors.password = 'Не более 72 символов'
  }

  if (values.confirmPassword === '') {
    errors.confirmPassword = 'Повторите пароль'
  } else if (values.confirmPassword !== values.password) {
    errors.confirmPassword = 'Пароли не совпадают'
  }

  return errors
}

type FieldConfig = {
  label: string
  icon: ReactNode
  type?: string
  autoComplete: string
  maxLength?: number
}

const FIELDS: Record<Field, FieldConfig> = {
  login: { label: 'Логин', icon: <UserIcon />, autoComplete: 'username', maxLength: 50 },
  lastName: { label: 'Фамилия', icon: <UserIcon />, autoComplete: 'family-name', maxLength: 100 },
  firstName: { label: 'Имя', icon: <UserIcon />, autoComplete: 'given-name', maxLength: 100 },
  email: { label: 'Email почта', icon: <MailIcon />, type: 'email', autoComplete: 'email', maxLength: 255 },
  password: { label: 'Пароль', icon: <LockIcon />, autoComplete: 'new-password', maxLength: 72 },
  confirmPassword: { label: 'Повторите пароль', icon: <LockIcon />, autoComplete: 'new-password', maxLength: 72 },
}

function RegisterPage() {
  const [values, setValues] = useState(initialValues)
  const [errors, setErrors] = useState<FieldErrors>({})
  const [serverError, setServerError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [visiblePasswords, setVisiblePasswords] = useState({ password: false, confirmPassword: false })
  const [avatar, setAvatar] = useState<File | null>(null)
  const [team, setTeam] = useState<Team | null>(null)
  const [isPickerOpen, setIsPickerOpen] = useState(false)
  const inputRefs = useRef<Partial<Record<Field, HTMLInputElement | null>>>({})
  const navigate = useNavigate()

  const focusFirstError = (fieldErrors: FieldErrors) => {
    const field = FIELD_ORDER.find((name) => fieldErrors[name])
    if (field) inputRefs.current[field]?.focus()
  }

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const name = event.target.name as Field
    setValues((prev) => ({ ...prev, [name]: event.target.value }))
    setErrors((prev) => ({ ...prev, [name]: undefined }))
    setServerError('')
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isSubmitting) return

    const nextErrors = validate(values)
    setErrors(nextErrors)
    if (Object.keys(nextErrors).length > 0) {
      focusFirstError(nextErrors)
      return
    }

    setIsSubmitting(true)
    try {
      // The user signs in explicitly afterwards, so returned tokens are not stored.
      await register({ ...values, favoriteTeamId: team?.id ?? null, avatar })
    } catch (error) {
      setIsSubmitting(false)
      const code = error instanceof ApiError ? error.code : ''
      if (code === 'USERNAME_ALREADY_EXISTS' || code === 'EMAIL_ALREADY_EXISTS') {
        const conflict: FieldErrors =
          code === 'USERNAME_ALREADY_EXISTS'
            ? { login: 'Этот логин уже занят' }
            : { email: 'Этот email уже зарегистрирован' }
        setErrors(conflict)
        focusFirstError(conflict)
      } else if (code === 'INVALID_REQUEST') {
        setServerError('Проверьте введённые данные')
      } else {
        setServerError('Не удалось зарегистрироваться. Попробуйте ещё раз')
      }
      return
    }
    navigate('/login', { replace: true })
  }

  return (
    <section
      className={isPickerOpen ? 'page auth-page auth-page--picker-open' : 'page auth-page'}
    >
      <div className="auth-card auth-card--register">
        <BrandLogo size="sm" />
        <h1>Регистрация</h1>
        <p className="auth-subtitle">
          Создайте аккаунт и станьте частью футбольного сообщества
        </p>

        <div className="register-extras">
          <AvatarUpload onChange={setAvatar} />
          <FavoriteTeamButton team={team} onClick={() => setIsPickerOpen(true)} />
        </div>

        <form className="auth-form auth-form--compact" onSubmit={handleSubmit} noValidate>
          {FIELD_ORDER.map((name) => {
            const config = FIELDS[name]
            const isPassword = name === 'password' || name === 'confirmPassword'
            const isVisible = isPassword && visiblePasswords[name]
            const errorId = `${name}-error`
            const type = isPassword ? (isVisible ? 'text' : 'password') : (config.type ?? 'text')

            return (
              <div className="auth-field" key={name}>
                <label htmlFor={name} className="visually-hidden">
                  {config.label}
                </label>
                <div className="auth-field__control">
                  <span className="auth-field__icon" aria-hidden="true">
                    {config.icon}
                  </span>
                  <input
                    ref={(element) => {
                      inputRefs.current[name] = element
                    }}
                    id={name}
                    name={name}
                    type={type}
                    autoComplete={config.autoComplete}
                    required
                    maxLength={config.maxLength}
                    placeholder={config.label}
                    value={values[name]}
                    aria-invalid={errors[name] ? true : undefined}
                    aria-describedby={errors[name] ? errorId : undefined}
                    onChange={handleChange}
                  />
                  {isPassword && (
                    <button
                      type="button"
                      className="auth-field__toggle"
                      aria-label={isVisible ? 'Скрыть пароль' : 'Показать пароль'}
                      aria-pressed={isVisible}
                      aria-controls={name}
                      onClick={() =>
                        setVisiblePasswords((prev) => ({ ...prev, [name]: !prev[name] }))
                      }
                    >
                      {isVisible ? <EyeIcon /> : <EyeOffIcon />}
                    </button>
                  )}
                </div>
                {errors[name] && (
                  <p className="form-error" id={errorId}>
                    {errors[name]}
                  </p>
                )}
              </div>
            )
          })}

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
                <span className="auth-loading-text">Регистрируем…</span>
              </>
            ) : (
              <>
                Зарегистрироваться
                <ArrowRightIcon className="auth-submit__arrow" />
              </>
            )}
          </button>
        </form>

        <p className="auth-switch auth-switch--plain">
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </p>
      </div>

      <TeamPickerModal
        open={isPickerOpen}
        selectedTeamId={team?.id ?? null}
        onSelect={setTeam}
        onClose={() => setIsPickerOpen(false)}
        beside
      />
    </section>
  )
}

export default RegisterPage
