import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreateInstallment,
  useDeleteInstallment,
  useInstallments,
  usePayInstallment,
  useUpdateInstallment,
} from '../../controllers/useSales'
import {
  EMPTY_INSTALLMENT_FORM,
  INSTALLMENT_STATUS_LABEL,
  INSTALLMENT_STATUS_TONE,
  installmentToForm,
  todayInput,
  type Contract,
  type Installment,
  type InstallmentFormValues,
} from '../../models/sales'
import { errorMessage } from '../../shared/errorMessage'
import { number, rupiah, shortDate } from '../../shared/format'
import { LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, ToolbarInput } from '../components/RecordControls'
import { ActionGroup, ConfirmAction } from './RowActions'
import { amountHint, refusalText } from './salesShared'

type RowIntent = { id: string; action: 'pay' | 'delete' }

function StatusCell({ installment }: { installment: Installment }) {
  return (
    <span className="flex flex-wrap items-center gap-2">
      <Chip tone={INSTALLMENT_STATUS_TONE[installment.status]}>{INSTALLMENT_STATUS_LABEL[installment.status]}</Chip>
      {installment.daysOverdue > 0 ? (
        <span className="text-xs font-medium text-red-700">lewat {number(installment.daysOverdue)} hari</span>
      ) : null}
    </span>
  )
}

function InstallmentForm({ contract, installment, items, onSaved, onCancel }: {
  contract: Contract
  installment: Installment | null
  items: Installment[]
  onSaved: (message: string) => void
  onCancel: () => void
}) {
  const [values, setValues] = useState<InstallmentFormValues>(() =>
    installment ? installmentToForm(installment) : EMPTY_INSTALLMENT_FORM,
  )
  const createInstallment = useCreateInstallment()
  const updateInstallment = useUpdateInstallment()
  const saving = installment ? updateInstallment : createInstallment

  // Batas sisa mengikuti ensureWithinContract di backend: angsuran yang sedang
  // diubah tidak ikut dihitung.
  const scheduledElsewhere = items
    .filter((item) => item.id !== installment?.id)
    .reduce((total, item) => total + item.amount, 0)
  const room = Math.max(0, contract.value - scheduledElsewhere)
  const nextNumber = items.reduce((highest, item) => Math.max(highest, item.installmentNumber), 0) + 1

  const update = (key: keyof InstallmentFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (installment) {
      updateInstallment.mutate(
        { installment, values },
        { onSuccess: (saved) => onSaved(`Angsuran ke-${saved.installmentNumber} berhasil diperbarui.`) },
      )
      return
    }
    const filled = {
      ...values,
      installmentNumber: values.installmentNumber === '' ? String(nextNumber) : values.installmentNumber,
    }
    createInstallment.mutate(
      { contractId: contract.id, values: filled },
      { onSuccess: () => setValues(EMPTY_INSTALLMENT_FORM) },
    )
  }

  return (
    <form onSubmit={submit} className="mt-6 space-y-4 border-t border-slate-200 pt-5">
      <h3 className="text-sm font-semibold text-slate-900">
        {installment ? `Ubah angsuran ke-${installment.installmentNumber}` : 'Tambah jadwal cicilan'}
      </h3>
      <div className="grid gap-4 md:grid-cols-3">
        <Field
          id="installment-number"
          label="Angsuran ke"
          type="number"
          inputMode="numeric"
          min={1}
          step={1}
          autoFocus={installment !== null}
          required={installment !== null}
          placeholder={String(nextNumber)}
          hint={installment ? 'Unik di kontrak ini' : `Kosongkan untuk memakai nomor ${nextNumber}`}
          value={values.installmentNumber}
          onChange={update('installmentNumber')}
        />
        <Field
          id="installment-due-date"
          label="Jatuh tempo"
          type="date"
          required
          value={values.dueDate}
          onChange={update('dueDate')}
        />
        <Field
          id="installment-amount"
          label="Nominal"
          type="number"
          inputMode="numeric"
          min={1}
          max={room > 0 ? room : undefined}
          step={1}
          required
          hint={amountHint(values.amount, `Paling banyak ${rupiah(room)} agar tidak melebihi nilai kontrak`)}
          value={values.amount}
          onChange={update('amount')}
        />
      </div>
      <div className="flex flex-wrap gap-2">
        <Button type="submit" isPending={saving.isPending} pendingLabel="Menyimpan cicilan">
          {installment ? 'Simpan perubahan' : 'Simpan cicilan'}
        </Button>
        {installment ? (
          <Button variant="subtle" onClick={onCancel}>
            Batal ubah
          </Button>
        ) : null}
      </div>
      {saving.isError ? <ErrorNote message={errorMessage(saving.error)} /> : null}
      {createInstallment.isSuccess && !installment ? (
        <SuccessNote message={`Angsuran ke-${createInstallment.data.installmentNumber} berhasil dijadwalkan.`} />
      ) : null}
    </form>
  )
}

// Jadwal cicilan satu kontrak. Aturannya mengikuti service backend:
// - jadwal bisa ditambah, diubah, dan dihapus selama kontrak Draf atau Aktif,
//   dan hanya untuk angsuran yang belum lunas;
// - pembayaran hanya bisa dicatat saat kontrak Aktif.
export function InstallmentSchedule({ contract, installments }: {
  contract: Contract
  installments: ReturnType<typeof useInstallments>
}) {
  const [intent, setIntent] = useState<RowIntent | null>(null)
  const [paidDate, setPaidDate] = useState(todayInput)
  const [editing, setEditing] = useState<Installment | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const pay = usePayInstallment()
  const deleteInstallment = useDeleteInstallment()

  const items = installments.data?.items ?? []
  const isOpen = contract.status === 'draft' || contract.status === 'active'
  const canPay = contract.status === 'active'

  const ask = (installment: Installment, action: RowIntent['action']) => {
    setNotice(null)
    pay.reset()
    deleteInstallment.reset()
    setPaidDate(todayInput())
    setIntent({ id: installment.id, action })
  }

  const isAsking = (installment: Installment, action: RowIntent['action']) =>
    intent?.id === installment.id && intent.action === action

  const confirmPay = (installment: Installment) => {
    pay.mutate(
      { installment, paidDate },
      {
        onSuccess: (saved) => {
          setNotice(`Angsuran ke-${saved.installmentNumber} dicatat lunas pada ${shortDate(saved.paidDate)}.`)
          if (editing?.id === installment.id) {
            setEditing(null)
          }
        },
        onSettled: () => setIntent(null),
      },
    )
  }

  const confirmDelete = (installment: Installment) => {
    deleteInstallment.mutate(installment, {
      onSuccess: () => {
        setNotice(`Angsuran ke-${installment.installmentNumber} berhasil dihapus.`)
        if (editing?.id === installment.id) {
          setEditing(null)
        }
      },
      onSettled: () => setIntent(null),
    })
  }

  const startEdit = (installment: Installment) => {
    setNotice(null)
    setEditing((current) => (current?.id === installment.id ? null : installment))
  }

  const actionCell = (row: Installment) => {
    if (row.status === 'paid' || !isOpen) {
      return null
    }
    const isBusy = intent?.id === row.id
    return (
      <ActionGroup>
        {canPay && !isAsking(row, 'delete') ? (
          <ConfirmAction
            label="Catat pembayaran"
            confirmLabel="Ya, catat lunas"
            pendingLabel="Menyimpan"
            tone="primary"
            isAsking={isAsking(row, 'pay')}
            isPending={pay.isPending && pay.variables.installment.id === row.id}
            onAsk={() => ask(row, 'pay')}
            onCancel={() => setIntent(null)}
            onConfirm={() => confirmPay(row)}
          >
            <ToolbarInput
              id={`installment-paid-date-${row.id}`}
              label="Dibayar tanggal"
              showLabel
              type="date"
              required
              autoFocus
              max={todayInput()}
              value={paidDate}
              onChange={(event) => setPaidDate(event.target.value)}
            />
          </ConfirmAction>
        ) : null}
        {isBusy ? null : (
          <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
        )}
        {isAsking(row, 'pay') ? null : (
          <ConfirmAction
            label="Hapus"
            question={`Hapus angsuran ke-${row.installmentNumber}?`}
            confirmLabel="Ya, hapus"
            pendingLabel="Menghapus"
            isAsking={isAsking(row, 'delete')}
            isPending={deleteInstallment.isPending && deleteInstallment.variables.id === row.id}
            onAsk={() => ask(row, 'delete')}
            onCancel={() => setIntent(null)}
            onConfirm={() => confirmDelete(row)}
          />
        )}
      </ActionGroup>
    )
  }

  return (
    <div>
      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}
      {pay.isError ? (
        <div className="mb-4">
          <ErrorNote
            message={refusalText(`Pembayaran angsuran ke-${pay.variables.installment.installmentNumber} gagal dicatat.`, pay.error)}
          />
        </div>
      ) : null}
      {deleteInstallment.isError ? (
        <div className="mb-4">
          <ErrorNote
            message={refusalText(
              `Angsuran ke-${deleteInstallment.variables.installmentNumber} tidak bisa dihapus.`,
              deleteInstallment.error,
            )}
          />
        </div>
      ) : null}

      {installments.isPending ? <Loading /> : null}
      {installments.isError ? <LoadFailed onRetry={() => installments.refetch()} /> : null}
      {installments.data ? (
        <Table
          rows={items}
          emptyMessage="Kontrak ini belum punya jadwal cicilan."
          columns={[
            { header: 'Ke', cell: (row) => row.installmentNumber },
            { header: 'Jatuh tempo', cell: (row) => shortDate(row.dueDate) },
            { header: 'Nominal', align: 'right', cell: (row) => rupiah(row.amount) },
            { header: 'Status', cell: (row) => <StatusCell installment={row} /> },
            { header: 'Dibayar', cell: (row) => (row.paidDate ? shortDate(row.paidDate) : 'Belum') },
            { header: 'Aksi', align: 'right', cell: actionCell },
          ]}
        />
      ) : null}

      {contract.status === 'draft' && items.length > 0 ? (
        <p className="mt-3 text-xs text-slate-500">Pembayaran bisa dicatat setelah kontrak diaktifkan.</p>
      ) : null}

      {isOpen ? (
        <InstallmentForm
          key={editing?.id ?? 'new'}
          contract={contract}
          installment={editing}
          items={items}
          onCancel={() => setEditing(null)}
          onSaved={(message) => {
            setEditing(null)
            setNotice(message)
          }}
        />
      ) : (
        <p className="mt-4 text-xs text-slate-500">
          Kontrak yang sudah lunas atau batal tidak bisa diubah jadwal cicilannya.
        </p>
      )}
    </div>
  )
}
