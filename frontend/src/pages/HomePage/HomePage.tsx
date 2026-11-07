import { Link } from 'react-router'
import { displayName, useCurrentUser } from '../../auth/currentUser.ts'
import UserAvatar from '../../components/UserAvatar/UserAvatar.tsx'
import './HomePage.css'

// The user comes from CurrentUserProvider; logout lives in the header's user menu.
function HomePage() {
  const { state, reload } = useCurrentUser()

  return (
    <section className="page home-page">
      <div> Главная страница</div>
      {state.status === 'guest' && (
        <Link to="/login" className="home-login-button">
          Войти
        </Link>
      )}
      {state.status === 'loading' && <p className="home-status">Загружаем профиль…</p>}
      {state.status === 'ready' && (
        <>
          <UserAvatar src={state.user.avatarUrl} alt="" size={96} />
          <p className="home-status">
            {`Вы вошли как ${displayName(state.user)} (@${state.user.login}), ${state.user.email}`}
          </p>
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

export default HomePage
