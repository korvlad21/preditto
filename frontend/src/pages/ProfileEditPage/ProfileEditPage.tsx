import { useRef, useState, type ChangeEvent, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router'
import { ApiError } from '../../api/client.ts'
import { PROFILE_UPDATE_NOT_IMPLEMENTED, updateProfile } from '../../api/user.ts'
import { useCurrentUser } from '../../auth/currentUser.ts'
import AvatarUpload from '../../components/AvatarUpload/AvatarUpload.tsx'
import FavoriteTeamButton from '../../components/FavoriteTeamButton/FavoriteTeamButton.tsx'
import TeamPickerModal from '../../components/TeamPickerModal/TeamPickerModal.tsx'
import type { Team } from '../../types/team.ts'
import type { User } from '../../types/user.ts'
import '../../styles/auth.css'
import './ProfileEditPage.css'

type Field = 'firstName' | 'lastName' | 'email'
type FormValues = Record<Field, string>
type FieldErrors = Partial<Record<Field, string>>
type TeamSummary = Pick<Team, 'id' | 'name' | 'logo_url'>

const FIELD_ORDER: Field[] = ['firstName', 'lastName', 'email']

const FIELDS: Record<Field, { label: string; type: string; autoComplete: string; maxLength: number }> = {
  firstName: { label: 'Имя', type: 'text', autoComplete: 'given-name', maxLength: 100 },
  lastName: { label: 'Фамилия', type: 'text', autoComplete: 'family-name', maxLength: 100 },
  email: { label: 'Email', type: 'email', autoComplete: 'email', maxLength: 255 },
}

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// Same limits as registration (backend RegisterRequest), pending a real update contract.
function validate(values: FormValues): FieldErrors {
  const errors: FieldErrors = {}
  if (values.firstName.trim() === '') errors.firstName = 'Введите имя'
  else if (values.firstName.length > 100) errors.firstName = 'Не более 100 символов'

  if (values.lastName.trim() === '') errors.lastName = 'Введите фамилию'
  else if (values.lastName.length > 100) errors.lastName = 'Не более 100 символов'

  const email = values.email.trim()
  if (email === '') errors.email = 'Введите email'
  else if (!EMAIL_PATTERN.test(email) || email.length > 255) errors.email = 'Введите корректный email'
  return errors
}

function ProfileEditForm({ user }: { user: User }) {
  const { reload: reloadUser } = useCurrentUser()
  const navigate = useNavigate()
  const initialValues: FormValues = {
    firstName: user.firstName ?? '',
    lastName: user.lastName ?? '',
    email: user.email,
  }
  const [values, setValues] = useState(initialValues)
  const [errors, setErrors] = useState<FieldErrors>({})
  // undefined: unchanged; null: removed; File: replaced.
  const [avatar, setAvatar] = useState<File | null | undefined>(undefined)
  const [team, setTeam] = useState<TeamSummary | null>(
    user.favoriteTeam
      ? { id: user.favoriteTeam.id, name: user.favoriteTeam.name, logo_url: user.favoriteTeam.logoUrl }
      : null,
  )
  const [teamTouched, setTeamTouched] = useState(false)
  const [isPickerOpen, setIsPickerOpen] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [notice, setNotice] = useState<{ kind: 'info' | 'error'; text: string } | null>(null)
  const inputRefs = useRef<Partial<Record<Field, HTMLInputElement | null>>>({})

  const isDirty =
    FIELD_ORDER.some((name) => values[name] !== initialValues[name]) ||
    teamTouched ||
    avatar !== undefined

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    const name = event.target.name as Field
    setValues((previous) => ({ ...previous, [name]: event.target.value }))
    setErrors((previous) => ({ ...previous, [name]: undefined }))
    setNotice(null)
  }

  const chooseTeam = (next: TeamSummary | null) => {
    setTeam(next)
    setTeamTouched(true)
    setNotice(null)
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (isSubmitting) return

    const nextErrors = validate(values)
    setErrors(nextErrors)
    const firstInvalid = FIELD_ORDER.find((name) => nextErrors[name])
    if (firstInvalid) {
      inputRefs.current[firstInvalid]?.focus()
      return
    }

    setIsSubmitting(true)
    setNotice(null)
    try {
      await updateProfile({
        firstName: values.firstName.trim(),
        lastName: values.lastName.trim(),
        email: values.email.trim(),
        favoriteTeamId: team?.id ?? null,
        avatar,
      })
    } catch (error) {
      setIsSubmitting(false)
      if (error instanceof ApiError && error.status === 401) {
        reloadUser()
        return
      }
      const notImplemented = error instanceof ApiError && error.code === PROFILE_UPDATE_NOT_IMPLEMENTED
      setNotice(
        notImplemented
          ? {
              kind: 'info',
              text: 'Сохранение профиля пока недоступно: API ещё не готово. Изменения не сохранены.',
            }
          : { kind: 'error', text: 'Не удалось сохранить профиль. Попробуйте ещё раз' },
      )
      return
    }
    reloadUser()
    navigate('/profile', { replace: true })
  }

  return (
    <>
      <form className="profile-edit__form" onSubmit={handleSubmit} noValidate>
        <div className="profile-edit__avatar">
          {/* The upload tile names itself ("Загрузить/Заменить аватарку"). */}
          <span className="profile-edit__label" aria-hidden="true">
            Аватар
          </span>
          <AvatarUpload initialUrl={user.avatarUrl} onChange={setAvatar} />
        </div>

        <div className="profile-edit__fields">
          {FIELD_ORDER.map((name) => {
            const config = FIELDS[name]
            const errorId = `${name}-error`
            return (
              <div className="profile-edit__field" key={name}>
                <label htmlFor={name} className="profile-edit__label">
                  {config.label}
                </label>
                <input
                  ref={(element) => {
                    inputRefs.current[name] = element
                  }}
                  id={name}
                  name={name}
                  type={config.type}
                  className="auth-input"
                  autoComplete={config.autoComplete}
                  maxLength={config.maxLength}
                  required
                  value={values[name]}
                  aria-invalid={errors[name] ? true : undefined}
                  aria-describedby={errors[name] ? errorId : undefined}
                  onChange={handleChange}
                />
                {errors[name] && (
                  <p className="form-error" id={errorId}>
                    {errors[name]}
                  </p>
                )}
              </div>
            )
          })}

          <div className="profile-edit__field">
            <label htmlFor="login" className="profile-edit__label">
              Логин
            </label>
            <input
              id="login"
              className="auth-input"
              value={user.login}
              readOnly
              aria-describedby="login-hint"
            />
            <p className="profile-edit__hint" id="login-hint">
              Логин изменить нельзя
            </p>
          </div>

          <div className="profile-edit__field profile-edit__field--wide">
            <span className="profile-edit__label">Любимая команда</span>
            <div className="profile-edit__team">
              <FavoriteTeamButton team={team} onClick={() => setIsPickerOpen(true)} />
              {team && (
                <button type="button" className="profile-edit__clear" onClick={() => chooseTeam(null)}>
                  Убрать
                </button>
              )}
            </div>
          </div>
        </div>

        {notice && (
          <div
            className={notice.kind === 'info' ? 'ui-alert ui-alert--info profile-edit__notice' : 'ui-alert profile-edit__notice'}
            role={notice.kind === 'error' ? 'alert' : 'status'}
          >
            {notice.text}
          </div>
        )}

        <div className="profile-edit__actions">
          <Link to="/profile" className="ui-button-outline profile-edit__cancel">
            Отмена
          </Link>
          <button
            type="submit"
            className="profile-edit__save"
            disabled={!isDirty || isSubmitting}
            aria-busy={isSubmitting || undefined}
          >
            {isSubmitting && <span className="ui-spinner" aria-hidden="true" />}
            {isSubmitting ? 'Сохраняем…' : 'Сохранить'}
          </button>
        </div>
      </form>

      <TeamPickerModal
        open={isPickerOpen}
        selectedTeamId={team?.id ?? null}
        onSelect={(selected) => chooseTeam(selected)}
        onClose={() => setIsPickerOpen(false)}
      />
    </>
  )
}

function ProfileEditPage() {
  const { state, reload } = useCurrentUser()

  return (
    <section
      className="profile-edit ui-panel auth-theme"
      aria-labelledby="profile-edit-title"
      aria-busy={state.status === 'loading'}
    >
      <div className="ui-section-head">
        <h1 id="profile-edit-title">Редактирование профиля</h1>
      </div>

      {/* Keyed by user so the form re-initializes if another user signs in. */}
      {state.status === 'ready' && <ProfileEditForm key={state.user.id} user={state.user} />}
      {state.status === 'loading' && <p className="profile-edit__status">Загружаем профиль…</p>}
      {state.status === 'error' && (
        <div className="ui-alert" role="alert">
          Не удалось загрузить профиль
          <button type="button" className="ui-button-outline" onClick={reload}>
            Повторить
          </button>
        </div>
      )}
    </section>
  )
}

export default ProfileEditPage
