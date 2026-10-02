export function BrandMark({ className = 'size-6' }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" className={className} aria-hidden="true">
      <rect width="24" height="24" rx="6.5" fill="#0b2b6b" />
      <path
        d="M6 12.25 12 7l6 5.25"
        fill="none"
        stroke="#fbbb16"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M8.75 13.25V17h6.5v-3.75"
        fill="none"
        stroke="#ffffff"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

export function BrandName({ className = '' }: { className?: string }) {
  return (
    <span className={`flex items-center gap-2 ${className}`}>
      <BrandMark />
      <span className="text-[15px] font-semibold tracking-[-0.02em] text-slate-900">BanguninAja</span>
    </span>
  )
}
