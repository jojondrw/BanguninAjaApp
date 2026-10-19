import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { SidePanel } from '../views/components/SidePanel'

describe('SidePanel', () => {
  it('membuka rincian sebagai dialog dan menutup lewat tombol, Escape, atau latar', async () => {
    const onClose = vi.fn()
    const { container } = render(
      <SidePanel label="Rincian piutang" onClose={onClose}>
        <p>Isi rincian</p>
      </SidePanel>,
    )

    const dialog = screen.getByRole('dialog', { name: 'Rincian piutang' })
    expect(dialog).toHaveAttribute('aria-modal', 'true')
    expect(screen.getByText('Isi rincian')).toBeInTheDocument()
    expect(container).toBeEmptyDOMElement()
    expect(document.body.style.overflow).toBe('hidden')

    await userEvent.click(screen.getByRole('button', { name: 'Tutup rincian' }))
    await userEvent.keyboard('{Escape}')
    await userEvent.click(dialog.previousElementSibling as HTMLElement)
    expect(onClose).toHaveBeenCalledTimes(3)
  })

  it('mengembalikan halaman seperti semula saat ditutup', () => {
    const root = document.createElement('div')
    root.id = 'root'
    document.body.appendChild(root)
    const { unmount } = render(
      <SidePanel label="Rincian kontrak" onClose={() => {}}>
        <p>Isi</p>
      </SidePanel>,
    )
    expect(root).toHaveAttribute('inert')

    unmount()
    expect(root).not.toHaveAttribute('inert')
    expect(document.body.style.overflow).toBe('')
    root.remove()
  })
})
