import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { Tabs, type TabItem } from '../views/components/Tabs'

type Section = 'ringkasan' | 'keuangan' | 'lokasi'

const TABS: TabItem<Section>[] = [
  { id: 'ringkasan', label: 'Ringkasan' },
  { id: 'keuangan', label: 'Keuangan' },
  { id: 'lokasi', label: 'Lokasi' },
]

function Harness() {
  const [active, setActive] = useState<Section>('ringkasan')
  return (
    <Tabs idPrefix="proyek" label="Bagian proyek" tabs={TABS} active={active} onChange={setActive}>
      <p>Isi {active}</p>
    </Tabs>
  )
}

describe('Tabs', () => {
  it('hanya tab aktif yang terpilih dan bisa difokus lewat Tab', () => {
    render(<Harness />)
    const [first, second] = screen.getAllByRole('tab')

    expect(screen.getByRole('tablist', { name: 'Bagian proyek' })).toBeInTheDocument()
    expect(first).toHaveAttribute('aria-selected', 'true')
    expect(first).toHaveAttribute('tabindex', '0')
    expect(second).toHaveAttribute('aria-selected', 'false')
    expect(second).toHaveAttribute('tabindex', '-1')
    expect(screen.getByRole('tabpanel')).toHaveAccessibleName('Ringkasan')
  })

  it('panah kanan dan kiri berpindah tab, berputar di ujung', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    await user.click(screen.getByRole('tab', { name: 'Ringkasan' }))

    await user.keyboard('{ArrowRight}')
    expect(screen.getByRole('tab', { name: 'Keuangan' })).toHaveAttribute('aria-selected', 'true')
    expect(screen.getByRole('tab', { name: 'Keuangan' })).toHaveFocus()
    expect(screen.getByRole('tabpanel')).toHaveTextContent('Isi keuangan')

    await user.keyboard('{ArrowLeft}{ArrowLeft}')
    expect(screen.getByRole('tab', { name: 'Lokasi' })).toHaveAttribute('aria-selected', 'true')
  })

  it('Home dan End melompat ke tab pertama dan terakhir', async () => {
    const user = userEvent.setup()
    render(<Harness />)
    await user.click(screen.getByRole('tab', { name: 'Ringkasan' }))

    await user.keyboard('{End}')
    expect(screen.getByRole('tab', { name: 'Lokasi' })).toHaveAttribute('aria-selected', 'true')
    await user.keyboard('{Home}')
    expect(screen.getByRole('tab', { name: 'Ringkasan' })).toHaveAttribute('aria-selected', 'true')
  })
})
