import { useEffect, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { ApiError } from '../../api/client.ts'
import {
  API_STATUS,
  CLEAR_SELECTION_NOT_IMPLEMENTED,
  clearSelectedTournament,
  getAllTournaments,
  selectTournament,
} from '../../api/tournaments.ts'
import { useCurrentUser } from '../../auth/currentUser.ts'
import { CalendarIcon, CheckIcon, TrophyIcon, UserIcon } from '../../components/Icons/Icons.tsx'
import type {
  GetAllTournamentsRequest,
  Tournament,
  TournamentStatus,
} from '../../types/tournament.ts'
import './TournamentsPage.css'

type StatusFilter = TournamentStatus | 'all'
// mine / not-mine go to the API (is_participate); current is filtered on the client.
type ParticipationFilter = 'all' | 'mine' | 'not-mine' | 'current'

const STATUS_FILTERS: { value: StatusFilter; label: string }[] = [
  { value: 'all', label: 'Все' },
  { value: 'active', label: 'Идут' },
  { value: 'upcoming', label: 'Скоро' },
  { value: 'finished', label: 'Завершённые' },
]

const PARTICIPATION_FILTERS: { value: ParticipationFilter; label: string }[] = [
  { value: 'all', label: 'Все турниры' },
  { value: 'mine', label: 'Я участвую' },
  { value: 'not-mine', label: 'Не участвую' },
  { value: 'current', label: 'Текущий' },
]

const STATUS_LABELS: Record<TournamentStatus, string> = {
  active: 'Идёт',
  upcoming: 'Скоро',
  finished: 'Завершён',
}

// select_tournament error codes (backend/docs/tournaments.md#select-a-tournament).
const SELECT_ERRORS: Record<string, string> = {
  TOURNAMENT_NOT_FOUND: 'Турнир не найден — возможно, он был удалён',
  USER_INFO_NOT_FOUND: 'Профиль пользователя не найден',
  USER_NOT_FOUND: 'Пользователь не найден',
  INVALID_TOURNAMENT_ID: 'Некорректный турнир',
  TOURNAMENT_ID_REQUIRED: 'Некорректный турнир',
  TOURNAMENT_UPDATE_FAILED: 'Не удалось сохранить выбор турнира',
}

function parseFilter<T extends string>(value: string | null, allowed: readonly { value: T }[]): T {
  return allowed.find((item) => item.value === value)?.value ?? allowed[0].value
}

const dateFormat = new Intl.DateTimeFormat('ru-RU', {
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
})

function formatDate(value?: string): string {
  if (!value) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '' : dateFormat.format(date)
}

function formatPeriod(start?: string, end?: string): string {
  const from = formatDate(start)
  const to = formatDate(end)
  if (from && to) return `${from} — ${to}`
  if (from) return `с ${from}`
  return to ? `до ${to}` : ''
}

function formatMoney(amount: number, currency?: string): string {
  if (amount === 0) return 'Бесплатно'
  if (!currency) return new Intl.NumberFormat('ru-RU').format(amount)
  try {
    return new Intl.NumberFormat('ru-RU', { style: 'currency', currency }).format(amount)
  } catch {
    return `${amount} ${currency}`
  }
}

function withCurrent(list: Tournament[], currentId: number | null): Tournament[] {
  return list.map((t) => ({ ...t, isCurrent: t.id === currentId }))
}

type CardNotice = { kind: 'error' | 'info'; text: string }

type CardProps = {
  tournament: Tournament
  notice?: CardNotice
  isBusy: boolean
  disabled: boolean
  onSelect: (id: number) => void
  onClear: (id: number) => void
}

function TournamentCard({ tournament: t, notice, isBusy, disabled, onSelect, onClear }: CardProps) {
  const titleId = `tournament-${t.id}-title`
  const period = formatPeriod(t.startDate, t.endDate)
  const participants =
    t.participantsLimit !== undefined
      ? `${t.participantsCount} / ${t.participantsLimit}`
      : String(t.participantsCount)

  return (
    <li
      className={t.isCurrent ? 'tournament-card tournament-card--current' : 'tournament-card'}
      aria-labelledby={titleId}
    >
      <div className="tournament-card__top">
        <span className={`tournament-status tournament-status--${t.status}`}>
          {STATUS_LABELS[t.status]}
        </span>
        {t.isCurrent && (
          // Receives focus after a successful selection.
          <span id={`tournament-${t.id}-current`} className="tournament-card__current" tabIndex={-1}>
            <CheckIcon />
            Текущий турнир
          </span>
        )}
      </div>

      <h2 id={titleId} className="tournament-card__name">
        {t.name}
      </h2>
      {t.type && <p className="tournament-card__type">{t.type}</p>}

      {notice && (
        <div
          className={
            notice.kind === 'info'
              ? 'ui-alert ui-alert--info tournament-card__alert'
              : 'ui-alert tournament-card__alert'
          }
          role={notice.kind === 'error' ? 'alert' : 'status'}
        >
          {notice.text}
        </div>
      )}

      <dl className="tournament-card__facts">
        {period && (
          <div>
            <dt>
              <CalendarIcon />
              <span className="visually-hidden">Даты</span>
            </dt>
            <dd>{period}</dd>
          </div>
        )}
        <div>
          <dt>
            <UserIcon />
            <span className="visually-hidden">Участники</span>
          </dt>
          <dd>Участников: {participants}</dd>
        </div>
        {t.entryFee !== undefined && (
          <div>
            <dt>
              <TrophyIcon />
              <span className="visually-hidden">Взнос</span>
            </dt>
            <dd>Взнос: {formatMoney(t.entryFee, t.currency)}</dd>
          </div>
        )}
        {t.prizePool !== undefined && (
          <div>
            <dt>
              <TrophyIcon />
              <span className="visually-hidden">Призовой фонд</span>
            </dt>
            <dd>Призовой фонд: {formatMoney(t.prizePool, t.currency)}</dd>
          </div>
        )}
      </dl>

      {t.prizes && (
        <div className="tournament-prizes">
          <span className="tournament-prizes__label">Распределение призов</span>
          <ol className="tournament-prizes__list">
            {t.prizes.map((prize) => (
              <li key={prize.place} className={`tournament-prize tournament-prize--${prize.place}`}>
                <span className="tournament-prize__place">{prize.place} место</span>
                <span className="tournament-prize__value">{prize.percent}%</span>
              </li>
            ))}
          </ol>
        </div>
      )}

      {/* The id lets focus return here after the selection is cleared. */}
      <button
        id={`tournament-${t.id}-action`}
        type="button"
        className={
          t.isCurrent
            ? 'tournament-card__action tournament-card__action--clear'
            : 'tournament-card__action ui-button-outline'
        }
        disabled={disabled}
        aria-busy={isBusy || undefined}
        aria-describedby={titleId}
        onClick={() => (t.isCurrent ? onClear(t.id) : onSelect(t.id))}
      >
        {isBusy && <span className="ui-spinner" aria-hidden="true" />}
        {t.isCurrent
          ? isBusy
            ? 'Снимаем…'
            : 'Снять выбор'
          : isBusy
            ? 'Выбираем…'
            : 'Выбрать турнир'}
      </button>
    </li>
  )
}

function TournamentsSkeleton() {
  return (
    <ul className="tournaments-grid" aria-hidden="true">
      {[0, 1, 2].map((key) => (
        <li key={key} className="tournament-card tournament-card--skeleton">
          <span className="tournament-skeleton__bar tournament-skeleton__bar--badge" />
          <span className="tournament-skeleton__bar tournament-skeleton__bar--title" />
          <span className="tournament-skeleton__bar" />
          <span className="tournament-skeleton__bar" />
          <span className="tournament-skeleton__bar tournament-skeleton__bar--button" />
        </li>
      ))}
    </ul>
  )
}

function SegmentGroup<T extends string>({
  label,
  items,
  value,
  onChange,
}: {
  label: string
  items: { value: T; label: string }[]
  value: T
  onChange: (value: T) => void
}) {
  return (
    <div className="tournaments-segments" role="group" aria-label={label}>
      {items.map((item) => (
        <button
          key={item.value}
          type="button"
          className="tournaments-segments__item"
          aria-pressed={value === item.value}
          onClick={() => onChange(item.value)}
        >
          {item.label}
        </button>
      ))}
    </div>
  )
}

type LoadResult =
  | { key: string; status: 'ok'; tournaments: Tournament[] }
  | { key: string; status: 'error' }

function TournamentsPage() {
  const { reload: reloadUser } = useCurrentUser()
  const [searchParams, setSearchParams] = useSearchParams()
  const status = parseFilter(searchParams.get('status'), STATUS_FILTERS)
  const participation = parseFilter(searchParams.get('participation'), PARTICIPATION_FILTERS)

  const [attempt, setAttempt] = useState(0)
  const [result, setResult] = useState<LoadResult | null>(null)
  // Card whose select/clear request is in flight; all action buttons are disabled meanwhile.
  const [busyId, setBusyId] = useState<number | null>(null)
  const [cardNotices, setCardNotices] = useState<Record<number, CardNotice>>({})
  const [toast, setToast] = useState('')
  // Element id to focus once it is rendered (badge after select, button after clear).
  const pendingFocus = useRef<string | null>(null)
  const latestKey = useRef('')

  // The API request for the URL filters; "current" is applied on the client.
  const request: GetAllTournamentsRequest = {
    ...(status !== 'all' && { status: API_STATUS[status] }),
    ...(participation === 'mine' && { is_participate: true }),
    ...(participation === 'not-mine' && { is_participate: false }),
  }
  const requestJson = JSON.stringify(request)
  // A result belongs to one request (and retry attempt); a newer key means loading.
  const requestKey = `${requestJson}|${attempt}`
  const isLoading = result?.key !== requestKey

  useEffect(() => {
    latestKey.current = requestKey
    let cancelled = false
    getAllTournaments(JSON.parse(requestJson) as GetAllTournamentsRequest)
      .then((tournaments) => {
        if (!cancelled) setResult({ key: requestKey, status: 'ok', tournaments })
      })
      .catch((error: unknown) => {
        if (cancelled) return
        if (error instanceof ApiError && error.status === 401) reloadUser()
        setResult({ key: requestKey, status: 'error' })
      })
    return () => {
      cancelled = true
    }
  }, [requestKey, requestJson, reloadUser])

  useEffect(() => {
    if (pendingFocus.current === null) return
    const target = document.getElementById(pendingFocus.current)
    if (target) {
      target.focus()
      pendingFocus.current = null
    }
  }, [result, toast])

  useEffect(() => {
    if (!toast) return
    const timer = window.setTimeout(() => setToast(''), 3000)
    return () => window.clearTimeout(timer)
  }, [toast])

  const setFilter = (name: 'status' | 'participation', value: string) => {
    setSearchParams(
      (previous) => {
        const next = new URLSearchParams(previous)
        if (value === 'all') next.delete(name)
        else next.set(name, value)
        return next
      },
      { replace: true },
    )
  }

  const updateList = (update: (list: Tournament[]) => Tournament[]) => {
    setResult((current) =>
      current?.status === 'ok' ? { ...current, tournaments: update(current.tournaments) } : current,
    )
  }

  const setNotice = (id: number, notice: CardNotice | null) => {
    setCardNotices((notices) => {
      const next = { ...notices }
      if (notice) next[id] = notice
      else delete next[id]
      return next
    })
  }

  // Re-read the list so counters and flags match the server; keep the local state if the
  // refresh fails or the filters changed meanwhile.
  const refreshList = (key: string) => {
    getAllTournaments(JSON.parse(requestJson) as GetAllTournamentsRequest)
      .then((tournaments) => {
        if (latestKey.current === key) setResult({ key, status: 'ok', tournaments })
      })
      .catch(() => {})
  }

  // Session ended: reloading the user lets RequireAuth send the guest to /login.
  const handleUnauthorized = (error: unknown) => {
    if (!(error instanceof ApiError && error.status === 401)) return false
    reloadUser()
    return true
  }

  const handleSelect = async (id: number) => {
    if (result?.status !== 'ok') return
    const previousId = result.tournaments.find((t) => t.isCurrent)?.id ?? null
    const key = requestKey

    setBusyId(id)
    setNotice(id, null)
    updateList((list) => withCurrent(list, id))

    try {
      await selectTournament(id)
    } catch (error) {
      updateList((list) => withCurrent(list, previousId))
      setBusyId(null)
      if (handleUnauthorized(error)) return
      const code = error instanceof ApiError ? error.code : ''
      setNotice(id, { kind: 'error', text: SELECT_ERRORS[code] ?? 'Не удалось выбрать турнир' })
      return
    }

    setBusyId(null)
    pendingFocus.current = `tournament-${id}-current`
    setToast('Турнир выбран')
    refreshList(key)
  }

  // Not optimistic: the API cannot clear the selection yet, so the state changes only
  // after a confirmed success.
  const handleClear = async (id: number) => {
    const key = requestKey
    setBusyId(id)
    setNotice(id, null)

    try {
      await clearSelectedTournament()
    } catch (error) {
      setBusyId(null)
      if (handleUnauthorized(error)) return
      const notImplemented =
        error instanceof ApiError && error.code === CLEAR_SELECTION_NOT_IMPLEMENTED
      setNotice(
        id,
        notImplemented
          ? { kind: 'info', text: 'Снять выбор пока нельзя: API ещё не готово. Выбор не изменился.' }
          : { kind: 'error', text: 'Не удалось снять выбор турнира' },
      )
      return
    }

    setBusyId(null)
    updateList((list) => withCurrent(list, null))
    pendingFocus.current = `tournament-${id}-action`
    setToast('Выбор снят')
    refreshList(key)
  }

  const loaded = !isLoading && result?.status === 'ok' ? result.tournaments : null
  const tournaments =
    loaded && participation === 'current' ? loaded.filter((t) => t.isCurrent) : loaded

  return (
    <>
    <section className="tournaments-page ui-panel" aria-labelledby="tournaments-title">
      <div className="ui-section-head">
        <div>
          <h1 id="tournaments-title">Выбор турнира</h1>
          <p className="tournaments-page__subtitle">
            Текущий турнир используется для прогнозов и таблиц на платформе
          </p>
        </div>
      </div>

      <div className="tournaments-filters">
        <SegmentGroup
          label="Статус турнира"
          items={STATUS_FILTERS}
          value={status}
          onChange={(value) => setFilter('status', value)}
        />
        <SegmentGroup
          label="Участие"
          items={PARTICIPATION_FILTERS}
          value={participation}
          onChange={(value) => setFilter('participation', value)}
        />
      </div>

      <p className="visually-hidden" role="status">
        {isLoading ? 'Загружаем турниры…' : tournaments ? `Найдено турниров: ${tournaments.length}` : ''}
      </p>

      {isLoading && <TournamentsSkeleton />}

      {!isLoading && result?.status === 'error' && (
        <div className="ui-alert" role="alert">
          Не удалось загрузить турниры
          <button type="button" className="ui-button-outline" onClick={() => setAttempt((n) => n + 1)}>
            Повторить
          </button>
        </div>
      )}

      {tournaments && tournaments.length === 0 && (
        <p className="tournaments-empty">
          {participation === 'current' && status === 'all'
            ? 'Текущий турнир не выбран'
            : 'Турниров с такими условиями нет'}
        </p>
      )}

      {tournaments && tournaments.length > 0 && (
        <ul className="tournaments-grid">
          {tournaments.map((tournament) => (
            <TournamentCard
              key={tournament.id}
              tournament={tournament}
              notice={cardNotices[tournament.id]}
              isBusy={busyId === tournament.id}
              disabled={busyId !== null}
              onSelect={handleSelect}
              onClear={handleClear}
            />
          ))}
        </ul>
      )}

    </section>
    {/* Outside the panel: its backdrop-filter would contain a fixed-position child. */}
    {toast && (
      <div className="ui-toast" role="status">
        <CheckIcon />
        {toast}
      </div>
    )}
    </>
  )
}

export default TournamentsPage
