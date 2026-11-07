import { useEffect, useId, useRef, useState, type KeyboardEvent, type ReactNode } from 'react'
import { Link, useNavigate } from 'react-router'
import { displayName, isAdmin, useCurrentUser } from '../../auth/currentUser.ts'
import { ChevronDownIcon, LogoutIcon, UserIcon } from '../Icons/Icons.tsx'
import UserAvatar from '../UserAvatar/UserAvatar.tsx'
import './UserMenu.css'

type MenuAction = 'register' | 'logout'
type MenuItem = { action: MenuAction; label: string; icon: ReactNode }

const REGISTER_ITEM: MenuItem = { action: 'register', label: 'Регистрация пользователя', icon: <UserIcon /> }
const LOGOUT_ITEM: MenuItem = { action: 'logout', label: 'Выйти', icon: <LogoutIcon /> }

// Menu button pattern (WAI-ARIA APG): arrows move between items, Esc closes and returns
// focus to the button, Tab/outside click/selection close the menu.
function UserMenu() {
  const { state, signOut } = useCurrentUser()
  const navigate = useNavigate()
  const menuId = useId()
  const [open, setOpen] = useState(false)
  const [focusIndex, setFocusIndex] = useState(0)
  const [isSigningOut, setIsSigningOut] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const itemRefs = useRef<(HTMLButtonElement | null)[]>([])

  useEffect(() => {
    if (!open) return
    itemRefs.current[focusIndex]?.focus()
  }, [open, focusIndex])

  useEffect(() => {
    if (!open) return
    const handlePointerDown = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }
    document.addEventListener('pointerdown', handlePointerDown)
    return () => document.removeEventListener('pointerdown', handlePointerDown)
  }, [open])

  if (state.status === 'guest') {
    return (
      <Link to="/login" className="user-menu__button user-menu__button--login">
        Войти
      </Link>
    )
  }

  if (state.status === 'loading') {
    return <span className="user-menu__button user-menu__button--loading" aria-hidden="true" />
  }

  const user = state.status === 'ready' ? state.user : null
  const name = user ? displayName(user) : 'Аккаунт'
  // Registration is admin-only (roles from user_roles via get_user_info).
  const menuItems = user && isAdmin(user) ? [REGISTER_ITEM, LOGOUT_ITEM] : [LOGOUT_ITEM]

  const openMenu = (index: number) => {
    setFocusIndex(index)
    setOpen(true)
  }

  const closeMenu = (returnFocus: boolean) => {
    setOpen(false)
    if (returnFocus) buttonRef.current?.focus()
  }

  const handleLogout = async () => {
    closeMenu(false)
    setIsSigningOut(true)
    await signOut()
    navigate('/login', { replace: true })
  }

  const handleSelect = (action: MenuAction) => {
    if (action === 'logout') {
      void handleLogout()
    } else {
      closeMenu(false)
      navigate('/register')
    }
  }

  const lastIndex = menuItems.length - 1

  const handleButtonKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      openMenu(0)
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      openMenu(lastIndex)
    }
  }

  const handleMenuKeyDown = (event: KeyboardEvent<HTMLUListElement>) => {
    const moves: Record<string, number> = {
      ArrowDown: focusIndex === lastIndex ? 0 : focusIndex + 1,
      ArrowUp: focusIndex === 0 ? lastIndex : focusIndex - 1,
      Home: 0,
      End: lastIndex,
    }
    if (event.key in moves) {
      event.preventDefault()
      setFocusIndex(moves[event.key])
      itemRefs.current[moves[event.key]]?.focus()
    } else if (event.key === 'Escape') {
      event.preventDefault()
      closeMenu(true)
    } else if (event.key === 'Tab') {
      closeMenu(false)
    }
  }

  return (
    <div className="user-menu" ref={rootRef}>
      <button
        ref={buttonRef}
        type="button"
        className="user-menu__button"
        aria-haspopup="menu"
        aria-expanded={open}
        aria-controls={open ? menuId : undefined}
        aria-busy={isSigningOut || undefined}
        disabled={isSigningOut}
        onClick={() => (open ? closeMenu(false) : openMenu(0))}
        onKeyDown={handleButtonKeyDown}
      >
        <UserAvatar src={user?.avatarUrl} alt="" size={40} />
        <span className="user-menu__name">{name}</span>
        {/* Keeps the button named when the visible name is hidden on narrow screens. */}
        <span className="visually-hidden">, меню пользователя</span>
        <ChevronDownIcon className="user-menu__chevron" />
      </button>

      {open && (
        <ul
          id={menuId}
          className="user-menu__list ui-panel"
          role="menu"
          aria-label="Меню пользователя"
          onKeyDown={handleMenuKeyDown}
        >
          {menuItems.map((item, index) => (
            <li role="none" key={item.label}>
              <button
                ref={(element) => {
                  itemRefs.current[index] = element
                }}
                type="button"
                role="menuitem"
                className="user-menu__item"
                tabIndex={index === focusIndex ? 0 : -1}
                onClick={() => handleSelect(item.action)}
              >
                {item.icon}
                {item.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

export default UserMenu
