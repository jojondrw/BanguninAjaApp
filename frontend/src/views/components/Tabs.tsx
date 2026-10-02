import { useRef, type KeyboardEvent, type ReactNode } from 'react'

import { SegmentIndicator } from './Segmented'
import { SEGMENT_LIST, segmentButton, useSegmentIndicator } from './segmentIndicator'

export interface TabItem<T extends string> {
  id: T
  label: string
}

interface TabsProps<T extends string> {
  idPrefix: string
  label: string
  tabs: readonly TabItem<T>[]
  active: T
  onChange: (id: T) => void
  children: ReactNode
}

const NEXT_KEYS = ['ArrowRight', 'ArrowDown']
const PREVIOUS_KEYS = ['ArrowLeft', 'ArrowUp']

function tabId(prefix: string, id: string) {
  return `${prefix}-tab-${id}`
}

// Pola tab WAI-ARIA: hanya tab aktif yang bisa difokus lewat Tab, panah kiri
// dan kanan berpindah tab, Home dan End ke tab pertama dan terakhir.
export function Tabs<T extends string>({ idPrefix, label, tabs, active, onChange, children }: TabsProps<T>) {
  const listRef = useRef<HTMLDivElement>(null)
  const indicator = useSegmentIndicator(listRef, active)
  const panelId = `${idPrefix}-panel`

  function select(index: number) {
    const target = tabs[(index + tabs.length) % tabs.length]
    onChange(target.id)
    listRef.current?.querySelector<HTMLButtonElement>(`#${tabId(idPrefix, target.id)}`)?.focus()
  }

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    const current = tabs.findIndex((tab) => tab.id === active)
    if (NEXT_KEYS.includes(event.key)) {
      select(current + 1)
    } else if (PREVIOUS_KEYS.includes(event.key)) {
      select(current - 1)
    } else if (event.key === 'Home') {
      select(0)
    } else if (event.key === 'End') {
      select(tabs.length - 1)
    } else {
      return
    }
    event.preventDefault()
  }

  return (
    <div>
      <div ref={listRef} role="tablist" aria-label={label} onKeyDown={handleKeyDown} className={SEGMENT_LIST}>
        <SegmentIndicator indicator={indicator} />
        {tabs.map((tab) => {
          const isActive = tab.id === active
          return (
            <button
              key={tab.id}
              id={tabId(idPrefix, tab.id)}
              data-segment={tab.id}
              type="button"
              role="tab"
              aria-selected={isActive}
              aria-controls={panelId}
              tabIndex={isActive ? 0 : -1}
              onClick={() => onChange(tab.id)}
              className={segmentButton(isActive, indicator !== null)}
            >
              {tab.label}
            </button>
          )
        })}
      </div>

      <div
        key={active}
        id={panelId}
        role="tabpanel"
        aria-labelledby={tabId(idPrefix, active)}
        className="pt-5 motion-safe:animate-enter"
      >
        {children}
      </div>
    </div>
  )
}
