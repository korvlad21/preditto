import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
  type MouseEvent,
} from 'react'
import { createPortal } from 'react-dom'
import { getAllCountries, getAllTeams } from '../../api/teams.ts'
import { COUNTRY_LABELS_RU, COUNTRY_PRIORITY } from '../../data/countries.ts'
import type { Country, Team } from '../../types/team.ts'
import { CheckIcon, CloseIcon, SearchIcon, ShieldStarIcon } from '../Icons/Icons.tsx'
import TeamLogo from '../TeamLogo/TeamLogo.tsx'
import '../../styles/auth.css'
import './TeamPickerModal.css'

type TeamPickerModalProps = {
  open: boolean
  selectedTeamId: number | null
  onSelect: (team: Team) => void
  onClose: () => void
  // On wide screens, dock beside a card shifted left instead of centering.
  beside?: boolean
}

type LoadState =
  | { status: 'loading' }
  | { status: 'error' }
  | { status: 'ready'; teams: Team[]; countries: Country[] }

const ALL = 'ALL'
const FOCUSABLE =
  'button:not([disabled]), input:not([disabled]), [tabindex]:not([tabindex="-1"])'

// Case- and diacritic-insensitive: "Atlético" matches "atletico".
function normalize(value: string) {
  return value.normalize('NFD').replace(/\p{Diacritic}/gu, '').toLowerCase().trim()
}

function countryLabel(code: string, countries: Country[]) {
  return (
    COUNTRY_LABELS_RU[code] ??
    countries.find((country) => country.short_name === code)?.name ??
    code
  )
}

function priority(code: string) {
  const index = COUNTRY_PRIORITY.indexOf(code)
  return index === -1 ? COUNTRY_PRIORITY.length : index
}

function TeamPickerModal({ open, ...props }: TeamPickerModalProps) {
  // Mounting on open gives every opening a fresh draft, search and filter.
  return open ? createPortal(<TeamPickerDialog {...props} />, document.body) : null
}

function TeamPickerDialog({
  selectedTeamId,
  onSelect,
  onClose,
  beside = false,
}: Omit<TeamPickerModalProps, 'open'>) {
  const [load, setLoad] = useState<LoadState>({ status: 'loading' })
  const [attempt, setAttempt] = useState(0)
  const [query, setQuery] = useState('')
  const [country, setCountry] = useState(ALL)
  const [draftId, setDraftId] = useState(selectedTeamId)
  const dialogRef = useRef<HTMLDivElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const pressedOnBackdrop = useRef(false)
  const titleId = useId()
  const subtitleId = useId()

  useEffect(() => {
    let cancelled = false
    Promise.all([getAllTeams(), getAllCountries().catch(() => [])]).then(
      ([teams, countries]) => {
        if (!cancelled) setLoad({ status: 'ready', teams, countries })
      },
      () => {
        if (!cancelled) setLoad({ status: 'error' })
      },
    )
    return () => {
      cancelled = true
    }
  }, [attempt])

  // Lock page scroll, focus search, and restore focus to the opener on close.
  useEffect(() => {
    const opener = document.activeElement as HTMLElement | null
    const { overflow } = document.body.style
    document.body.style.overflow = 'hidden'
    searchRef.current?.focus()
    return () => {
      document.body.style.overflow = overflow
      opener?.focus()
    }
  }, [])

  const teams = useMemo(() => (load.status === 'ready' ? load.teams : []), [load])
  const countries = load.status === 'ready' ? load.countries : []

  const countryCodes = useMemo(
    () =>
      [...new Set(teams.map((team) => team.country))].sort(
        (a, b) =>
          priority(a) - priority(b) ||
          (COUNTRY_LABELS_RU[a] ?? a).localeCompare(COUNTRY_LABELS_RU[b] ?? b, 'ru'),
      ),
    [teams],
  )

  const visibleTeams = useMemo(() => {
    const needle = normalize(query)
    return teams.filter(
      (team) =>
        (country === ALL || team.country === country) &&
        normalize(team.name).includes(needle),
    )
  }, [teams, query, country])

  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Escape') {
      event.stopPropagation()
      onClose()
      return
    }
    if (event.key !== 'Tab' || !dialogRef.current) return

    const items = dialogRef.current.querySelectorAll<HTMLElement>(FOCUSABLE)
    if (items.length === 0) return
    const first = items[0]
    const last = items[items.length - 1]
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }

  // Close only when both press and release happen on the backdrop itself.
  const handleBackdropClick = (event: MouseEvent<HTMLDivElement>) => {
    if (pressedOnBackdrop.current && event.target === event.currentTarget) {
      onClose()
    }
  }

  const handleConfirm = () => {
    const team = teams.find((item) => item.id === draftId)
    if (!team) return
    onSelect(team)
    onClose()
  }

  const retry = () => {
    setLoad({ status: 'loading' })
    setAttempt((value) => value + 1)
  }

  return (
    <div
      className="team-picker-overlay auth-theme"
      onMouseDown={(event) => {
        pressedOnBackdrop.current = event.target === event.currentTarget
      }}
      onClick={handleBackdropClick}
      onKeyDown={handleKeyDown}
    >
      <div
        ref={dialogRef}
        className={beside ? 'team-picker team-picker--beside' : 'team-picker'}
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={subtitleId}
      >
        <button
          type="button"
          className="team-picker__close"
          aria-label="Закрыть"
          onClick={onClose}
        >
          <CloseIcon />
        </button>

        <header className="team-picker__header">
          <span className="team-picker__emblem" aria-hidden="true">
            <ShieldStarIcon />
          </span>
          <h2 id={titleId}>Выберите любимую команду</h2>
          <p id={subtitleId}>
            Поддерживайте любимую команду и получайте персональные рекомендации
          </p>
        </header>

        <div className="team-picker__search">
          <SearchIcon className="team-picker__search-icon" />
          <input
            ref={searchRef}
            className="auth-input"
            type="search"
            placeholder="Поиск команды…"
            aria-label="Поиск команды"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
          {query && (
            <button
              type="button"
              className="team-picker__clear"
              aria-label="Очистить поиск"
              onClick={() => {
                setQuery('')
                searchRef.current?.focus()
              }}
            >
              <CloseIcon />
            </button>
          )}
        </div>

        {countryCodes.length > 0 && (
          <div className="team-picker__chips" role="group" aria-label="Страна">
            {[ALL, ...countryCodes].map((code) => (
              <button
                key={code}
                type="button"
                className="team-picker__chip"
                aria-pressed={country === code}
                onClick={() => setCountry(code)}
              >
                {code === ALL ? 'Все' : countryLabel(code, countries)}
              </button>
            ))}
          </div>
        )}

        <div className="team-picker__body">
          {load.status === 'loading' && (
            <p className="team-picker__status">Загрузка команд…</p>
          )}
          {load.status === 'error' && (
            <div className="team-picker__status" role="alert">
              <p>Не удалось загрузить команды</p>
              <button type="button" className="team-picker__retry" onClick={retry}>
                Повторить
              </button>
            </div>
          )}
          {load.status === 'ready' && visibleTeams.length === 0 && (
            <p className="team-picker__status">Ничего не найдено</p>
          )}
          {visibleTeams.length > 0 && (
            <div className="team-picker__grid">
              {visibleTeams.map((team) => (
                <button
                  key={team.id}
                  type="button"
                  className="team-picker__team"
                  aria-pressed={draftId === team.id}
                  onClick={() => setDraftId(team.id)}
                >
                  <TeamLogo src={team.logo_url} name={team.name} size={44} />
                  <span className="team-picker__team-name">{team.name}</span>
                  {draftId === team.id && (
                    <span className="team-picker__check" aria-hidden="true">
                      <CheckIcon />
                    </span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>

        <button
          type="button"
          className="auth-submit team-picker__confirm"
          disabled={!teams.some((team) => team.id === draftId)}
          onClick={handleConfirm}
        >
          Выбрать
        </button>
      </div>
    </div>
  )
}

export default TeamPickerModal
