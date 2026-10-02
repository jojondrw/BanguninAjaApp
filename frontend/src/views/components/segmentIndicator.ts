import { type RefObject, useLayoutEffect, useState } from 'react'

export interface Indicator {
  left: number
  width: number
}

export const SEGMENT_LIST =
  'relative inline-flex max-w-full gap-0.5 overflow-x-auto rounded-[10px] bg-slate-100 p-[3px] [scrollbar-width:none]'

export function segmentButton(isActive: boolean, hasIndicator: boolean) {
  const resting = isActive ? 'text-slate-900' : 'text-slate-500 hover:text-slate-900'
  const fallback = isActive && !hasIndicator ? 'bg-white shadow-hairline' : ''
  return (
    'relative z-10 h-7 shrink-0 rounded-[7px] px-3 text-[13px] font-medium whitespace-nowrap ' +
    `transition-colors duration-150 ${resting} ${fallback}`
  )
}

// Latar putih tab aktif digambar sekali lalu digeser ke tombol yang dipilih,
// jadi perpindahan tab terasa menyambung, bukan berkedip.
export function useSegmentIndicator(listRef: RefObject<HTMLElement | null>, activeKey: string) {
  const [indicator, setIndicator] = useState<Indicator | null>(null)

  useLayoutEffect(() => {
    const list = listRef.current
    if (!list) {
      return
    }
    const measure = () => {
      const target = list.querySelector<HTMLElement>(`[data-segment="${CSS.escape(activeKey)}"]`)
      if (target) {
        setIndicator({ left: target.offsetLeft, width: target.offsetWidth })
      }
    }
    measure()
    const observer = new ResizeObserver(measure)
    observer.observe(list)
    return () => observer.disconnect()
  }, [listRef, activeKey])

  return indicator
}
