<svg
  class="auth-illustration"
  viewBox="0 0 240 200"
  role="img"
  aria-label="Event tickets and confetti"
  xmlns="http://www.w3.org/2000/svg"
>
  <circle class="backdrop" cx="120" cy="110" r="84" />
  <circle cx="44" cy="40" r="4" fill="#8fb4ff" />
  <circle cx="206" cy="58" r="5" fill="#ffb74d" />
  <circle cx="196" cy="170" r="3" fill="#8fb4ff" />
  <path d="M30 150 l6 -6 l6 6 l-6 6 z" fill="#ffb74d" />
  <path d="M190 26 l5 -5 l5 5 l-5 5 z" fill="#5b8def" />
  <rect x="58" y="108" width="12" height="4" rx="2" fill="#f06292" transform="rotate(-30 64 110)" />
  <rect x="176" y="118" width="12" height="4" rx="2" fill="#4db6ac" transform="rotate(25 182 120)" />

  <g transform="rotate(-12 110 100)">
    <path
      class="ticket-back"
      d="M52 62 h116 a8 8 0 0 1 8 8 v14 a14 14 0 0 0 0 28 v14 a8 8 0 0 1 -8 8 h-116 a8 8 0 0 1 -8 -8 v-14 a14 14 0 0 0 0 -28 v-14 a8 8 0 0 1 8 -8 z"
    />
  </g>
  <path
    class="ticket"
    d="M60 76 h120 a8 8 0 0 1 8 8 v14 a14 14 0 0 0 0 28 v14 a8 8 0 0 1 -8 8 h-120 a8 8 0 0 1 -8 -8 v-14 a14 14 0 0 0 0 -28 v-14 a8 8 0 0 1 8 -8 z"
  />
  <line x1="146" y1="84" x2="146" y2="140" stroke="#ffffff" stroke-width="2" stroke-dasharray="4 4" opacity="0.6" />
  <rect x="72" y="94" width="56" height="8" rx="4" fill="#ffffff" />
  <rect x="72" y="110" width="40" height="6" rx="3" fill="#ffffff" opacity="0.7" />
  <rect x="72" y="122" width="48" height="6" rx="3" fill="#ffffff" opacity="0.7" />
  <circle cx="166" cy="112" r="10" fill="#ffb74d" />
  <path d="M161 112 l4 4 l7 -8" class="check" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round" />
</svg>

<style lang="sass">
  .auth-illustration
    width: 100%
    max-width: 260px
    height: auto

  .backdrop
    fill: var(--illustration-backdrop)

  .ticket-back
    fill: var(--illustration-ticket-back)

  .ticket
    fill: var(--illustration-ticket)

  .check
    stroke: var(--illustration-ticket)
</style>
