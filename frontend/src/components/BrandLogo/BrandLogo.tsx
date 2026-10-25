import { useId } from 'react'
import './BrandLogo.css'

function BrandLogo() {
  const gradientId = useId()

  return (
    <div className="brand-logo" role="img" aria-label="Preditto">
      <svg
        className="brand-logo__mark"
        viewBox="0 0 56 62"
        fill="none"
        aria-hidden="true"
      >
        <defs>
          <linearGradient
            id={gradientId}
            x1="0"
            y1="62"
            x2="56"
            y2="0"
            gradientUnits="userSpaceOnUse"
          >
            <stop offset="0" stopColor="#7B5CFF" />
            <stop offset="0.5" stopColor="#2F6BFF" />
            <stop offset="1" stopColor="#3FD0FF" />
          </linearGradient>
        </defs>
        <path
          d="M14 54V10h16a13 13 0 0 1 0 26H14"
          stroke={`url(#${gradientId})`}
          strokeWidth="10"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
      <span className="brand-logo__text" aria-hidden="true">
        Preditto
      </span>
    </div>
  )
}

export default BrandLogo
