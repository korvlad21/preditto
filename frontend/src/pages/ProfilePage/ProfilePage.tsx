import { Link } from 'react-router'
import { displayName, useCurrentUser } from '../../auth/currentUser.ts'
import { CalendarIcon, PencilIcon } from '../../components/Icons/Icons.tsx'
import TeamLogo from '../../components/TeamLogo/TeamLogo.tsx'
import UserAvatar from '../../components/UserAvatar/UserAvatar.tsx'
import type { User } from '../../types/user.ts'
import './ProfilePage.css'

const dateFormat = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
})

function formatDate(value: string): string {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : dateFormat.format(date)
}

function FavoriteTeam({ team }: { team: User['favoriteTeam'] }) {
  // undefined: the teams directory failed to load; omit the row rather than mislead.
  if (team === undefined) return null

  if (team === null) {
    return (
      <p className="profile-team profile-team--empty">
        Любимая команда: не выбрана{' '}
        <Link to="/profile/edit" className="profile-link">
          Выбрать
        </Link>
      </p>
    )
  }

  return (
    <div className="profile-team">
      <TeamLogo src={team.logoUrl} name={team.name} size={52} />
      <div className="profile-team__text">
        <span className="profile-team__label">Любимая команда</span>
        <span className="profile-team__name">{team.name}</span>
      </div>
    </div>
  )
}

function ProfileDetails({ user }: { user: User }) {
  const name = displayName(user)
  const registered = formatDate(user.createdAt)

  return (
    <div className="profile-card__body">
      <UserAvatar src={user.avatarUrl} alt={name} size={228} className="profile-avatar" />
      <div className="profile-info">
        <p className="profile-info__name">{name}</p>
        <p className="profile-info__login">@{user.login}</p>
        {registered && (
          <p className="profile-info__date">
            <CalendarIcon />
            <span>
              Дата регистрации: <time dateTime={user.createdAt}>{registered}</time>
            </span>
          </p>
        )}
        <FavoriteTeam team={user.favoriteTeam} />
      </div>
    </div>
  )
}

function ProfileSkeleton() {
  return (
    <div className="profile-card__body profile-skeleton" aria-hidden="true">
      <span className="profile-skeleton__avatar" />
      <div className="profile-info">
        <span className="profile-skeleton__bar profile-skeleton__bar--wide" />
        <span className="profile-skeleton__bar" />
        <span className="profile-skeleton__bar profile-skeleton__bar--narrow" />
      </div>
    </div>
  )
}

function ProfilePage() {
  const { state, reload } = useCurrentUser()

  return (
    <section
      className="profile-card ui-panel"
      aria-labelledby="profile-title"
      aria-busy={state.status === 'loading'}
    >
      <div className="ui-section-head">
        <h1 id="profile-title">Профиль</h1>
        <Link to="/profile/edit" className="ui-button-outline profile-edit">
          <PencilIcon />
          Изменить профиль
        </Link>
      </div>

      {state.status === 'ready' && <ProfileDetails user={state.user} />}
      {state.status === 'loading' && (
        <>
          <span className="visually-hidden" role="status">
            Загружаем профиль…
          </span>
          <ProfileSkeleton />
        </>
      )}
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

export default ProfilePage
