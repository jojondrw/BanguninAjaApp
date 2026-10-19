import { X } from 'lucide-react'
import { type ReactNode, useEffect, useRef } from 'react'
import { createPortal } from 'react-dom'

// Panel rincian yang menggeser masuk dari kanan di atas daftar. Halaman di
// belakangnya dibuat inert selama panel terbuka, jadi fokus keyboard tetap di
// panel; Escape, tombol tutup, atau klik di luar panel menutupnya.
export function SidePanel({ label, onClose, children }: { label: string; onClose: () => void; children: ReactNode }) {
  const panel = useRef<HTMLDivElement>(null)
  const closeRef = useRef(onClose)

  useEffect(() => {
    closeRef.current = onClose
  }, [onClose])

  useEffect(() => {
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null
    const root = document.getElementById('root')
    root?.setAttribute('inert', '')
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    panel.current?.focus()

    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !event.defaultPrevented) {
        closeRef.current()
      }
    }
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('keydown', onKey)
      root?.removeAttribute('inert')
      document.body.style.overflow = overflow
      previous?.focus()
    }
  }, [])

  return createPortal(
    <div className="fixed inset-0 z-40">
      <div aria-hidden="true" className="absolute inset-0 bg-slate-900/25 motion-safe:animate-fade" onClick={onClose} />
      <div
        ref={panel}
        role="dialog"
        aria-modal="true"
        aria-label={label}
        tabIndex={-1}
        className="absolute inset-y-0 right-0 flex w-full max-w-2xl flex-col bg-slate-100 shadow-[0_0_0_1px_rgb(0_0_0/0.06),-24px_0_48px_-24px_rgb(0_0_0/0.25)] outline-none motion-safe:animate-sheet sm:inset-y-2 sm:right-2 sm:rounded-2xl"
      >
        <div className="flex shrink-0 items-center justify-between gap-3 px-5 pt-4 pb-2 sm:px-6">
          <p className="truncate text-xs font-medium text-slate-500">{label}</p>
          <button
            type="button"
            onClick={onClose}
            aria-label="Tutup rincian"
            className="grid size-8 shrink-0 place-items-center rounded-full bg-white text-slate-600 shadow-hairline transition-colors hover:bg-slate-50 hover:text-slate-900"
          >
            <X aria-hidden="true" className="size-4" strokeWidth={2} />
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-3 pb-3 sm:px-4 sm:pb-4">{children}</div>
      </div>
    </div>,
    document.body,
  )
}
