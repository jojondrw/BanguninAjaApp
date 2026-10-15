import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import type { SearchOption } from '../models/lookup'
import { SearchSelect } from '../views/components/SearchSelect'
import { renderWithProviders } from './render'

const CUSTOMERS: SearchOption[] = [
  { value: 'c1', label: 'Budi Santoso', hint: 'KTP-1' },
  { value: 'c2', label: 'Citra Lestari', hint: 'KTP-2' },
  { value: 'c3', label: 'Dewi Anggraini', hint: 'KTP-3' },
]

function setup(props: { initial?: SearchOption | null; value?: string } = {}) {
  const fetchOptions = vi.fn(async (search: string) =>
    CUSTOMERS.filter((option) => option.label.toLowerCase().includes(search.toLowerCase())),
  )
  const onChange = vi.fn()

  function Harness() {
    const [value, setValue] = useState(props.value ?? '')
    return (
      <SearchSelect
        id="customer"
        label="Pelanggan"
        cacheKey="customers-test"
        placeholder="Cari pelanggan"
        fetchOptions={fetchOptions}
        initial={props.initial}
        value={value}
        onChange={(next, option) => {
          setValue(next)
          onChange(next, option)
        }}
      />
    )
  }

  renderWithProviders(<Harness />)
  return { fetchOptions, onChange, user: userEvent.setup() }
}

describe('SearchSelect', () => {
  it('memakai label yang terhubung ke kotak isian dan tertutup di awal', () => {
    setup()
    const input = screen.getByRole('combobox', { name: 'Pelanggan' })
    expect(input).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
  })

  it('membuka daftar saat diklik dan mengambil pilihan dari server', async () => {
    const { fetchOptions, user } = setup()
    await user.click(screen.getByRole('combobox', { name: 'Pelanggan' }))

    expect(await screen.findAllByRole('option')).toHaveLength(3)
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-expanded', 'true')
    expect(fetchOptions).toHaveBeenCalledWith('')
  })

  it('panah bawah dan Enter memilih pilihan yang disorot', async () => {
    const { onChange, user } = setup()
    const input = screen.getByRole('combobox', { name: 'Pelanggan' })
    await user.click(input)
    await screen.findAllByRole('option')

    await user.keyboard('{ArrowDown}{ArrowDown}{ArrowUp}{ArrowDown}{Enter}')

    expect(onChange).toHaveBeenLastCalledWith('c3', CUSTOMERS[2])
    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(input).toHaveValue('Dewi Anggraini')
  })

  it('Escape menutup daftar tanpa mengubah pilihan', async () => {
    const { onChange, user } = setup()
    await user.click(screen.getByRole('combobox', { name: 'Pelanggan' }))
    await screen.findAllByRole('option')

    await user.keyboard('{Escape}')

    expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    expect(onChange).not.toHaveBeenCalled()
  })

  it('mencari ke server sesuai ketikan', async () => {
    const { fetchOptions, user } = setup()
    await user.click(screen.getByRole('combobox', { name: 'Pelanggan' }))
    await user.keyboard('cit')

    await waitFor(() => expect(fetchOptions).toHaveBeenCalledWith('cit'))
    await waitFor(() => expect(screen.getAllByRole('option')).toHaveLength(1))
    expect(screen.getByRole('option')).toHaveTextContent('Citra Lestari')
  })

  it('menampilkan label isian awal walau pilihannya tidak ada di hasil', () => {
    setup({ value: 'lama', initial: { value: 'lama', label: 'Pelanggan Lama' } })
    expect(screen.getByRole('combobox', { name: 'Pelanggan' })).toHaveValue('Pelanggan Lama')
  })

  it('tombol kosongkan melepas pilihan', async () => {
    const { onChange, user } = setup({ value: 'lama', initial: { value: 'lama', label: 'Pelanggan Lama' } })

    await user.click(screen.getByRole('button', { name: 'Kosongkan pelanggan' }))

    expect(onChange).toHaveBeenCalledWith('', null)
    expect(screen.getByRole('combobox', { name: 'Pelanggan' })).toHaveValue('')
    expect(screen.queryByRole('button', { name: 'Kosongkan pelanggan' })).not.toBeInTheDocument()
  })
})
