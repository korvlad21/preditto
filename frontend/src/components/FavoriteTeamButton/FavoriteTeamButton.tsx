import type { Ref } from 'react'
import type { Team } from '../../types/team.ts'
import { ChevronRightIcon, ShieldStarIcon } from '../Icons/Icons.tsx'
import TeamLogo from '../TeamLogo/TeamLogo.tsx'
import './FavoriteTeamButton.css'

type FavoriteTeamButtonProps = {
  // Only the name and logo are shown, so a profile's team summary also fits.
  team: Pick<Team, 'name' | 'logo_url'> | null
  onClick: () => void
  ref?: Ref<HTMLButtonElement>
}

function FavoriteTeamButton({ team, onClick, ref }: FavoriteTeamButtonProps) {
  return (
    <button
      ref={ref}
      type="button"
      className="favorite-team-button"
      aria-haspopup="dialog"
      onClick={onClick}
    >
      {team ? (
        <TeamLogo src={team.logo_url} name={team.name} size={28} />
      ) : (
        <ShieldStarIcon className="favorite-team-button__icon" />
      )}
      <span className="favorite-team-button__text">
        <span className="favorite-team-button__title">
          {team ? team.name : 'Выбрать любимую команду'}
        </span>
        <span className="favorite-team-button__hint">
          {team ? 'Изменить' : 'Необязательно'}
        </span>
      </span>
      <ChevronRightIcon className="favorite-team-button__chevron" />
    </button>
  )
}

export default FavoriteTeamButton
