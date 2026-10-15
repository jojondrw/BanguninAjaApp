import { useState } from 'react'

import { type ManagedUser, ROLE_OPTIONS, type RoleCode } from '../../models/user'
import { BUTTON_BASE, BUTTON_PRIMARY, BUTTON_SUBTLE } from '../components/Form'
import { FilterSelect } from '../components/RecordControls'

const NO_ROLE = ''

// Peran dikirim setelah Simpan ditekan, supaya salah pilih di daftar tidak
// langsung mengubah hak akses orang lain. Setelah tersimpan, baris dipasang
// ulang lewat key sehingga pilihan kembali bersih.
export function RoleCell({ user, isPending, onSave }: {
  user: ManagedUser
  isPending: boolean
  onSave: (user: ManagedUser, role: RoleCode) => void
}) {
  const current = user.role?.code ?? NO_ROLE
  const [draft, setDraft] = useState(current)
  const isChanged = draft !== current && draft !== NO_ROLE
  const hint = ROLE_OPTIONS.find((option) => option.code === draft)?.hint

  return (
    <div className="flex flex-wrap items-center gap-2">
      <FilterSelect
        id={`user-role-${user.id}`}
        label={`Peran ${user.name}`}
        value={draft}
        disabled={isPending}
        title={hint}
        onChange={(event) => setDraft(event.target.value)}
      >
        {current === NO_ROLE ? <option value={NO_ROLE}>Tanpa peran</option> : null}
        {ROLE_OPTIONS.map((option) => (
          <option key={option.code} value={option.code}>
            {option.label}
          </option>
        ))}
      </FilterSelect>
      {isChanged ? (
        <>
          <button
            type="button"
            disabled={isPending}
            aria-busy={isPending}
            onClick={() => onSave(user, draft as RoleCode)}
            className={`h-8 ${BUTTON_BASE} ${BUTTON_PRIMARY}`}
          >
            {isPending ? 'Menyimpan' : 'Simpan peran'}
          </button>
          {isPending ? null : (
            <button type="button" onClick={() => setDraft(current)} className={`h-8 ${BUTTON_BASE} ${BUTTON_SUBTLE}`}>
              Batal
            </button>
          )}
        </>
      ) : null}
    </div>
  )
}
