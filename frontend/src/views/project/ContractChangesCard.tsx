import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useCan } from '../../controllers/useAccess'
import {
  useDecideContractChange,
  useDeleteContractChange,
  usePaymentSchedule,
  useProposeContractChange,
} from '../../controllers/useProjectWorkspace'
import { useSessionStore } from '../../controllers/sessionStore'
import {
  CHANGE_DAYS_MAX,
  CHANGE_DESCRIPTION_MAX_LENGTH,
  CHANGE_STATUS_LABEL,
  CHANGE_STATUS_TONE,
  changeTrail,
  EMPTY_CHANGE_FORM,
  signedRupiahShort,
  type ConstructionContract,
  type ContractChange,
  type ContractChangeFormValues,
} from '../../models/paymentTerm'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, rupiahShort, shortDate } from '../../shared/format'
import { AccessScope } from '../access/AccessScope'
import { Card, type Column, Table } from '../components/Data'
import { Button, ErrorNote, Field, SelectField } from '../components/Form'
import { Chip, FormSection, InlineConfirm } from './parts'
import { Money } from '../components/Money'

const ADMIN_ROLE = 'admin'
const CLOSED_STATUSES: readonly string[] = ['completed', 'cancelled']

// Adendum mengubah kontrak kontraktor, jadi kartunya hanya muncul setelah
// kontrak diatur. Data diambil dari jadwal termin yang sama, tanpa permintaan
// tambahan.
export function ContractChangesCard({ project }: { project: Project }) {
  const schedule = usePaymentSchedule(project.id)
  const contract = schedule.data?.contract

  if (!schedule.data || !contract) {
    return null
  }

  return (
    <AccessScope module="projects">
      <ChangesContent project={project} contract={contract} changes={schedule.data.changes} />
    </AccessScope>
  )
}

function ChangesContent({ project, contract, changes }: {
  project: Project
  contract: ConstructionContract
  changes: ContractChange[]
}) {
  const canWrite = useCan('projects', 'write')
  const isAdmin = useSessionStore((state) => state.user?.role?.code === ADMIN_ROLE)
  const decide = useDecideContractChange(project.id)
  const remove = useDeleteContractChange(project.id)
  const [isProposing, setIsProposing] = useState(false)
  const isOpen = !CLOSED_STATUSES.includes(project.status)
  const isDeciding = (changeId: string, status: 'approved' | 'rejected') =>
    decide.isPending && decide.variables?.changeId === changeId && decide.variables.status === status

  const columns: Column<ContractChange>[] = [
    { header: 'Adendum', cell: (row) => <span className="font-medium text-slate-900">{row.label}</span> },
    { header: 'Uraian', cell: (row) => <span className="block max-w-sm text-slate-700">{row.description}</span> },
    {
      header: 'Nilai',
      align: 'right',
      cell: (row) => (row.valueChange === 0 ? '-' : <span className="tabular-nums"><Money>{signedRupiahShort(row.valueChange, rupiahShort)}</Money></span>),
    },
    { header: 'Waktu', align: 'right', cell: (row) => (row.daysChange > 0 ? `+${row.daysChange} hari` : '-') },
    {
      header: 'Status',
      cell: (row) => (
        <span className="inline-flex flex-col items-start gap-1">
          <Chip tone={CHANGE_STATUS_TONE[row.status]}>{CHANGE_STATUS_LABEL[row.status]}</Chip>
          <span className="text-xs text-slate-500">{changeTrail(row, shortDate)}</span>
        </span>
      ),
    },
  ]
  if (canWrite) {
    columns.push({
      header: 'Aksi',
      align: 'right',
      cell: (row) =>
        row.status === 'proposed' ? (
          <span className="inline-flex items-center justify-end gap-1">
            {isAdmin ? (
              <>
                <InlineConfirm
                  tone="primary"
                  label="Setujui"
                  confirmLabel={`Ya, setujui ${row.label.toLowerCase()}`}
                  pendingLabel="Menyimpan"
                  isPending={isDeciding(row.id, 'approved')}
                  onConfirm={() => decide.mutate({ changeId: row.id, status: 'approved' })}
                />
                <InlineConfirm
                  label="Tolak"
                  confirmLabel="Ya, tolak"
                  pendingLabel="Menyimpan"
                  isPending={isDeciding(row.id, 'rejected')}
                  onConfirm={() => decide.mutate({ changeId: row.id, status: 'rejected' })}
                />
              </>
            ) : null}
            {isOpen ? (
              <InlineConfirm
                label="Hapus"
                confirmLabel="Ya, hapus usulan"
                pendingLabel="Menghapus"
                isPending={remove.isPending && remove.variables === row.id}
                onConfirm={() => remove.mutate(row.id)}
              />
            ) : null}
          </span>
        ) : null,
    })
  }

  return (
    <Card title="Pekerjaan tambah-kurang" description={changesDescription(contract)}>
      <Table
        rows={changes}
        emptyMessage="Belum ada adendum. Nilai kontrak masih sama dengan nilai awal."
        columns={columns}
      />

      {decide.isError ? (
        <div className="mt-4">
          <ErrorNote message={errorMessage(decide.error)} />
        </div>
      ) : null}
      {remove.isError ? (
        <div className="mt-4">
          <ErrorNote message={errorMessage(remove.error)} />
        </div>
      ) : null}
      {changes.some((change) => change.status === 'proposed') && !isAdmin ? (
        <p className="mt-4 text-[13px] text-slate-500">Usulan adendum menunggu persetujuan Admin.</p>
      ) : null}

      {canWrite && isOpen ? (
        <div className="mt-6">
          <Button variant="subtle" onClick={() => setIsProposing((open) => !open)}>
            {isProposing ? 'Tutup formulir adendum' : 'Usulkan adendum'}
          </Button>
        </div>
      ) : null}
      {isProposing && canWrite && isOpen ? (
        <ChangeForm projectId={project.id} contract={contract} onDone={() => setIsProposing(false)} />
      ) : null}
    </Card>
  )
}

function changesDescription(contract: ConstructionContract): string {
  if (contract.approvedChange === 0 && contract.extensionDays === 0) {
    return 'Perubahan nilai dan waktu kontrak kontraktor tercatat lewat adendum'
  }
  const extension = contract.extensionDays > 0 ? `, waktu +${contract.extensionDays} hari` : ''
  return `Nilai berjalan ${rupiahShort(contract.currentValue)} dari awal ${rupiahShort(contract.value)}${extension}`
}

function ChangeForm({ projectId, contract, onDone }: {
  projectId: string
  contract: ConstructionContract
  onDone: () => void
}) {
  const [values, setValues] = useState<ContractChangeFormValues>(EMPTY_CHANGE_FORM)
  const propose = useProposeContractChange(projectId)

  const update =
    (key: keyof ContractChangeFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    propose.mutate(values, { onSuccess: onDone })
  }

  const amount = Number(values.amount)
  const amountHint =
    values.amount === '' || !Number.isFinite(amount) || amount <= 0
      ? 'Kosongkan kalau hanya waktu yang berubah'
      : `${rupiah(amount)}, nilai berjalan menjadi ${rupiahShort(contract.currentValue + (values.kind === 'deduct' ? -amount : amount))}`

  return (
    <FormSection title="Usulan adendum">
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-3">
          <SelectField id="change-kind" label="Jenis" value={values.kind} onChange={update('kind')}>
            <option value="add">Pekerjaan tambah</option>
            <option value="deduct">Pekerjaan kurang</option>
          </SelectField>
          <Field
            id="change-amount"
            label="Nilai"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            hint={amountHint}
            value={values.amount}
            onChange={update('amount')}
          />
          <Field
            id="change-days"
            label="Perpanjangan waktu (hari)"
            type="number"
            inputMode="numeric"
            min={0}
            max={CHANGE_DAYS_MAX}
            step={1}
            placeholder="0"
            value={values.days}
            onChange={update('days')}
          />
          <div className="md:col-span-3">
            <Field
              id="change-description"
              label="Uraian dan alasan"
              required
              autoComplete="off"
              maxLength={CHANGE_DESCRIPTION_MAX_LENGTH}
              placeholder="Penambahan ruang NICU di lantai 3 atas permintaan pemberi kerja"
              value={values.description}
              onChange={update('description')}
            />
          </div>
        </div>

        <Button type="submit" isPending={propose.isPending} pendingLabel="Mengirim usulan">
          Kirim usulan
        </Button>
        {propose.isError ? <ErrorNote message={errorMessage(propose.error)} /> : null}
      </form>
    </FormSection>
  )
}
