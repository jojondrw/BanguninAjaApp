import { TriangleAlert } from 'lucide-react'
import { Component, type ErrorInfo, type ReactNode } from 'react'

import { BUTTON_BASE, BUTTON_PRIMARY, BUTTON_SUBTLE } from './Form'

interface ErrorBoundaryProps {
  resetKey: string
  children: ReactNode
}

interface ErrorBoundaryState {
  error: Error | null
  resetKey: string
}

// Galat saat merender satu halaman tidak boleh mengosongkan seluruh aplikasi.
// Batas ini dipasang ulang setiap kali alamat berubah, jadi pindah halaman
// lewat menu langsung memulihkan tampilan.
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { error: null, resetKey: this.props.resetKey }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { error }
  }

  static getDerivedStateFromProps(props: ErrorBoundaryProps, state: ErrorBoundaryState): Partial<ErrorBoundaryState> | null {
    if (props.resetKey !== state.resetKey) {
      return { error: null, resetKey: props.resetKey }
    }
    return null
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error(error, info.componentStack)
  }

  render() {
    if (!this.state.error) {
      return this.props.children
    }

    const isStaleChunk = /dynamically imported module|Failed to fetch|Loading chunk/i.test(this.state.error.message)

    return (
      <div role="alert" className="flex min-h-screen items-center justify-center bg-slate-100 px-6">
        <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-panel">
          <span className="grid size-10 place-items-center rounded-xl bg-red-50 text-red-600">
            <TriangleAlert aria-hidden="true" className="size-5" strokeWidth={1.8} />
          </span>
          <h1 className="mt-4 text-lg font-semibold text-slate-900">
            {isStaleChunk ? 'Versi aplikasi baru tersedia' : 'Halaman ini gagal ditampilkan'}
          </h1>
          <p className="mt-1.5 text-[13px] leading-relaxed text-slate-600">
            {isStaleChunk
              ? 'Sebagian berkas aplikasi sudah diperbarui di server. Muat ulang untuk memakai versi terbaru.'
              : 'Terjadi kesalahan di tampilan, bukan di data kamu. Muat ulang halaman atau kembali ke Ringkasan.'}
          </p>
          <div className="mt-5 flex gap-2">
            <button type="button" onClick={() => window.location.reload()} className={`h-9 ${BUTTON_BASE} ${BUTTON_PRIMARY}`}>
              Muat ulang
            </button>
            <a href="/" className={`h-9 ${BUTTON_BASE} ${BUTTON_SUBTLE}`}>
              Ke Ringkasan
            </a>
          </div>
        </div>
      </div>
    )
  }
}
