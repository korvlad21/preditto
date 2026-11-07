import type { ReactNode } from 'react'
import { Link, NavLink } from 'react-router'
import BrandLogo from '../BrandLogo/BrandLogo.tsx'
import { HomeIcon, TrophyIcon, UserIcon } from '../Icons/Icons.tsx'
import UserMenu from '../UserMenu/UserMenu.tsx'
import './AppHeader.css'

type NavItem = { to: string; label: string; icon: ReactNode; end?: boolean }

const NAV_ITEMS: NavItem[] = [
  { to: '/', label: 'Главная', icon: <HomeIcon />, end: true },
  { to: '/profile', label: 'Личный кабинет', icon: <UserIcon /> },
  { to: '/tournaments', label: 'Выбор турнира', icon: <TrophyIcon /> },
]

// Shared header of the signed-in area. At ≤ 640px the navigation becomes a bottom tab bar.
function AppHeader() {
  return (
    <header className="app-header ui-panel">
      <Link to="/" className="app-header__brand">
        <BrandLogo size="sm" />
      </Link>

      <nav className="app-nav" aria-label="Основная навигация">
        {NAV_ITEMS.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.end}
            className="app-nav__link"
            aria-label={item.label}
            title={item.label}
          >
            {item.icon}
            <span className="app-nav__label">{item.label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="app-header__user">
        <UserMenu />
      </div>
    </header>
  )
}

export default AppHeader
