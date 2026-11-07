import { useState, type CSSProperties } from 'react'
import './UserAvatar.css'

type UserAvatarProps = {
  // User.avatarUrl: nullable, and may point to something that is not an image.
  src?: string | null
  // Accessible name; pass '' when the user's name is already shown next to the avatar.
  alt: string
  size: number
  className?: string
}

// Round avatar; falls back to a head-and-shoulders silhouette when there is no image
// or it fails to load.
function UserAvatar({ src, alt, size, className }: UserAvatarProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null)
  const style = { '--user-avatar-size': `${size}px` } as CSSProperties
  const classes = (extra: string) => ['user-avatar', extra, className].filter(Boolean).join(' ')

  if (!src || failedSrc === src) {
    return (
      <span
        className={classes('user-avatar--fallback')}
        style={style}
        {...(alt ? { role: 'img', 'aria-label': alt } : { 'aria-hidden': true })}
      >
        <svg viewBox="0 0 100 100" aria-hidden="true" focusable="false">
          <circle cx="50" cy="38" r="18" />
          <path d="M12 104c2-23 18-38 38-38s36 15 38 38Z" />
        </svg>
      </span>
    )
  }

  return (
    <img
      className={classes('')}
      style={style}
      src={src}
      alt={alt}
      width={size}
      height={size}
      referrerPolicy="no-referrer"
      onError={() => setFailedSrc(src)}
    />
  )
}

export default UserAvatar
