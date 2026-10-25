import { useState, type CSSProperties } from 'react'
import './TeamLogo.css'

type TeamLogoProps = {
  src: string | null
  name: string
  size: number
}

function initials(name: string) {
  return name
    .split(/[\s/-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((word) => word[0])
    .join('')
    .toUpperCase()
}

// Decorative: callers always show the team name next to the logo.
function TeamLogo({ src, name, size }: TeamLogoProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null)
  const style = { '--team-logo-size': `${size}px` } as CSSProperties

  if (!src || failedSrc === src) {
    return (
      <span className="team-logo team-logo--fallback" style={style} aria-hidden="true">
        {initials(name)}
      </span>
    )
  }

  return (
    <img
      className="team-logo"
      style={style}
      src={src}
      alt=""
      width={size}
      height={size}
      loading="lazy"
      onError={() => setFailedSrc(src)}
    />
  )
}

export default TeamLogo
