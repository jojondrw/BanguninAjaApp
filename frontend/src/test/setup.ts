import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// jsdom tidak punya tata letak, jadi API yang hanya berarti di peramban asli
// diganti versi kosong supaya komponen bisa dirender.
class NoopResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

if (!('ResizeObserver' in globalThis)) {
  globalThis.ResizeObserver = NoopResizeObserver as unknown as typeof ResizeObserver
}

if (!Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = () => {}
}

if (typeof CSS === 'undefined' || typeof CSS.escape !== 'function') {
  globalThis.CSS = { ...(globalThis.CSS ?? {}), escape: (value: string) => value.replace(/[^\w-]/g, '\\$&') } as typeof CSS
}

afterEach(() => {
  cleanup()
})
