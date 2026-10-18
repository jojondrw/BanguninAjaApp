// Logo BanguninAja: pin lokasi dengan Bani berhelm proyek. Pin untuk analisis
// lokasi, helm untuk konstruksi, berang-berang untuk pembangunnya.
export function BrandMark({ className = 'h-7 w-auto' }: { className?: string }) {
  return <img src="/brand/logo-128.webp" alt="" aria-hidden="true" width={102} height={128} draggable={false} className={className} />
}

export function BrandName({ className = '' }: { className?: string }) {
  return (
    <span className={`flex items-center gap-2 ${className}`}>
      <BrandMark />
      <span className="text-[15px] font-semibold tracking-[-0.02em] text-slate-900">BanguninAja</span>
    </span>
  )
}
