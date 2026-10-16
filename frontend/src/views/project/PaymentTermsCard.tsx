import { type ChangeEvent, type FormEvent, type ReactNode, useState } from 'react'
import { Link } from 'react-router-dom'

import { useCan } from '../../controllers/useAccess'
import {
  useClaimPaymentTerm,
  useCreatePaymentTerm,
  useDeletePaymentTerm,
  usePaymentSchedule,
  useReleaseRetention,
  useSaveContract,
} from '../../controllers/useProjectWorkspace'
import { activeVendorOptions } from '../../models/lookupApi'
import {
  CONTRACT_NUMBER_MAX_LENGTH,
  contractFormValues,
  formatPercent,
  nextTermForm,
  remainingPercent,
  RETENTION_MAX_PERCENT,
  signedRupiahShort,
  TERM_STATE_LABEL,
  TERM_STATE_TONE,
  termNote,
  type ConstructionContract,
  type ContractFormValues,
  type PaymentSchedule,
  type PaymentTerm,
  type PaymentTermFormValues,
} from '../../models/paymentTerm'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, rupiahShort, shortDate } from '../../shared/format'
import { AccessScope } from '../access/AccessScope'
import { Card, type Column, Empty, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { SearchSelect } from '../components/SearchSelect'
import { Chip, FormSection, InlineConfirm } from './parts'

type Panel = 'contract' | 'term' | null

const CLOSED_STATUSES: readonly string[] = ['completed', 'cancelled']

// Aksi termin butuh hak tulis Proyek, bukan Keuangan, walaupun kartunya ada di
// tab Keuangan: Manajer Proyek mengesahkan progres, Keuangan yang membayar.
export function PaymentTermsCard({ project }: { project: Project }) {
  return (
    <AccessScope module="projects">
      <PaymentTermsContent project={project} />
    </AccessScope>
  )
}

function PaymentTermsContent({ project }: { project: Project }) {
  const schedule = usePaymentSchedule(project.id)
  const canWrite = useCan('projects', 'write')
  const [panel, setPanel] = useState<Panel>(null)
  const data = schedule.data

  const toggle = (next: Exclude<Panel, null>) => setPanel((current) => (current === next ? null : next))
  const close = () => setPanel(null)

  return (
    <Card title="Termin kontraktor" description={scheduleDescription(data)}>
      {schedule.isPending ? <Loading /> : null}
      {schedule.isError ? <LoadFailed onRetry={() => schedule.refetch()} /> : null}
      {data && !data.contract ? (
        <Empty message="Belum ada kontrak kontraktor. Atur kontraknya dulu, lalu susun termin pembayaran berdasarkan progres." />
      ) : null}
      {data?.contract ? (
        <ScheduleBody project={project} schedule={data} contract={data.contract} canWrite={canWrite} />
      ) : null}

      {data && canWrite ? (
        <ScheduleActions project={project} schedule={data} panel={panel} onToggle={toggle} />
      ) : null}

      {panel === 'contract' && data ? (
        <ContractForm key={data.contract?.number ?? 'baru'} projectId={project.id} schedule={data} onDone={close} />
      ) : null}
      {panel === 'term' && data ? <TermForm projectId={project.id} schedule={data} onDone={close} /> : null}
    </Card>
  )
}

function scheduleDescription(schedule: PaymentSchedule | undefined): string | undefined {
  if (!schedule) {
    return undefined
  }
  if (!schedule.contract) {
    return 'Pembayaran ke kontraktor mengikuti progres berbobot proyek'
  }
  const { vendorName, number, retentionPercent } = schedule.contract
  return `${vendorName}, SPK ${number}, retensi ${formatPercent(retentionPercent)}`
}

function ScheduleBody({ project, schedule, contract, canWrite }: {
  project: Project
  schedule: PaymentSchedule
  contract: ConstructionContract
  canWrite: boolean
}) {
  const claim = useClaimPaymentTerm(project.id)
  const remove = useDeletePaymentTerm(project.id)
  const canOpenBilling = useCan('billing', 'read')
  const isOpen = !CLOSED_STATUSES.includes(project.status)
  const claimed = claim.data?.terms.find((term) => term.id === claim.variables)

  const columns: Column<PaymentTerm>[] = [
    { header: 'Termin', cell: (row) => <span className="font-medium text-slate-900">{row.label}</span> },
    { header: 'Syarat', cell: (row) => (row.kind === 'retention' ? 'Serah terima' : `Progres ${row.requiredProgress}%`) },
    {
      header: 'Nilai',
      align: 'right',
      cell: (row) => (
        <Stacked note={row.kind === 'progress' ? `${formatPercent(row.percent)} kontrak` : undefined}>
          {rupiahShort(row.grossAmount)}
        </Stacked>
      ),
    },
    {
      header: 'Dibayarkan',
      align: 'right',
      cell: (row) => (
        <Stacked note={row.retentionAmount > 0 ? `retensi ${rupiahShort(row.retentionAmount)}` : undefined}>
          {rupiahShort(row.netAmount)}
        </Stacked>
      ),
    },
    {
      header: 'Status',
      cell: (row) => (
        <Stacked note={termNote(row, schedule.progress, shortDate)} align="left">
          <Chip tone={TERM_STATE_TONE[row.state]}>{TERM_STATE_LABEL[row.state]}</Chip>
        </Stacked>
      ),
    },
  ]
  if (canWrite || canOpenBilling) {
    columns.push({
      header: 'Aksi',
      align: 'right',
      cell: (row) => (
        <TermAction
          term={row}
          canWrite={canWrite}
          canDelete={isOpen}
          canOpenBilling={canOpenBilling}
          isClaiming={claim.isPending && claim.variables === row.id}
          isDeleting={remove.isPending && remove.variables === row.id}
          onClaim={() => claim.mutate(row.id)}
          onDelete={() => remove.mutate(row.id)}
        />
      ),
    })
  }

  return (
    <div className="space-y-6">
      <dl className="grid gap-x-8 gap-y-5 sm:grid-cols-2 lg:grid-cols-4">
        <Fact label="Nilai kontrak" note={contractValueNote(contract)}>
          {rupiahShort(contract.currentValue)}
        </Fact>
        <Fact label="Sudah diajukan" note={`${shareOf(schedule.claimedAmount, contract.currentValue)} dari kontrak`}>
          {rupiahShort(schedule.claimedAmount)}
        </Fact>
        <Fact label="Sudah dibayar" note="lewat modul Tagihan">
          {rupiahShort(schedule.paidAmount)}
        </Fact>
        <Fact label="Retensi ditahan" note={retentionNote(schedule)}>
          {rupiahShort(schedule.retentionHeld)}
        </Fact>
      </dl>

      <Table
        rows={schedule.terms}
        emptyMessage="Belum ada termin. Tambahkan termin dengan syarat progres dan persen nilai kontrak."
        columns={columns}
      />

      {claim.isError ? <ErrorNote message={errorMessage(claim.error)} /> : null}
      {remove.isError ? <ErrorNote message={errorMessage(remove.error)} /> : null}
      {claim.isSuccess && claimed ? (
        <SuccessNote
          message={`${claimed.label} diajukan. Hutang ${rupiah(claimed.netAmount)} ke ${contract.vendorName} tercatat di Tagihan, jatuh tempo ${contract.paymentTermDays} hari lagi.`}
        />
      ) : null}
    </div>
  )
}

function TermAction({ term, canWrite, canDelete, canOpenBilling, isClaiming, isDeleting, onClaim, onDelete }: {
  term: PaymentTerm
  canWrite: boolean
  canDelete: boolean
  canOpenBilling: boolean
  isClaiming: boolean
  isDeleting: boolean
  onClaim: () => void
  onDelete: () => void
}) {
  if (term.state === 'ready' && canWrite) {
    return (
      <InlineConfirm
        tone="primary"
        label="Ajukan"
        confirmLabel={`Ya, ajukan ${rupiahShort(term.netAmount)}`}
        pendingLabel="Mengajukan"
        isPending={isClaiming}
        onConfirm={onClaim}
      />
    )
  }
  if (term.state === 'planned' && canWrite && canDelete) {
    return (
      <InlineConfirm
        label="Hapus"
        confirmLabel="Ya, hapus termin"
        pendingLabel="Menghapus"
        isPending={isDeleting}
        onConfirm={onDelete}
      />
    )
  }
  if (term.payableId && canOpenBilling) {
    return (
      <Link to="/tagihan" className="rounded-md px-2 py-1 text-xs font-medium text-slate-600 hover:bg-slate-100">
        Lihat di Tagihan
      </Link>
    )
  }
  return null
}

function ScheduleActions({ project, schedule, panel, onToggle }: {
  project: Project
  schedule: PaymentSchedule
  panel: Panel
  onToggle: (panel: Exclude<Panel, null>) => void
}) {
  const release = useReleaseRetention(project.id)
  const isOpen = !CLOSED_STATUSES.includes(project.status)
  const canAddTerm = isOpen && schedule.contract !== null && remainingPercent(schedule) > 0
  const canEditContract = isOpen && !schedule.contractLocked
  const hint = releaseHint(project, schedule)

  if (!canAddTerm && !canEditContract && !schedule.canReleaseRetention && !hint) {
    return null
  }

  return (
    <div className="mt-6 space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        {canAddTerm ? (
          <Button variant="subtle" onClick={() => onToggle('term')}>
            {panel === 'term' ? 'Tutup formulir termin' : 'Tambah termin'}
          </Button>
        ) : null}
        {canEditContract ? (
          <Button variant="subtle" onClick={() => onToggle('contract')}>
            {contractToggleLabel(schedule, panel)}
          </Button>
        ) : null}
        {schedule.canReleaseRetention ? (
          <InlineConfirm
            tone="primary"
            label="Lepas retensi"
            confirmLabel={`Ya, lepas ${rupiahShort(schedule.retentionHeld)}`}
            pendingLabel="Melepas retensi"
            isPending={release.isPending}
            onConfirm={() => release.mutate()}
          />
        ) : null}
      </div>
      {hint ? <p className="text-[13px] text-slate-500">{hint}</p> : null}
      {release.isError ? <ErrorNote message={errorMessage(release.error)} /> : null}
      {release.isSuccess ? (
        <SuccessNote message="Retensi dilepas. Hutangnya tercatat di Tagihan untuk dibayar ke kontraktor." />
      ) : null}
    </div>
  )
}

function contractToggleLabel(schedule: PaymentSchedule, panel: Panel): string {
  if (panel === 'contract') {
    return 'Tutup formulir kontrak'
  }
  return schedule.contract ? 'Ubah kontrak' : 'Atur kontrak'
}

function releaseHint(project: Project, schedule: PaymentSchedule): string {
  if (schedule.retentionHeld === 0 || schedule.retentionReleased || schedule.canReleaseRetention) {
    return ''
  }
  if (project.status !== 'completed') {
    return `Retensi ${rupiahShort(schedule.retentionHeld)} dilepas setelah proyek selesai dan diserahterimakan.`
  }
  return 'Ajukan semua termin progres dulu sebelum melepas retensi.'
}

function contractValueNote(contract: ConstructionContract): string {
  if (contract.approvedChange === 0) {
    return 'belum ada adendum'
  }
  return `awal ${rupiahShort(contract.value)}, adendum ${signedRupiahShort(contract.approvedChange, rupiahShort)}`
}

function retentionNote(schedule: PaymentSchedule): string {
  if (schedule.retentionReleased) {
    return 'sudah dilepas'
  }
  return schedule.retentionHeld > 0 ? 'dilepas setelah serah terima' : 'belum ada yang ditahan'
}

function shareOf(part: number, whole: number): string {
  return whole > 0 ? formatPercent(Math.round((part / whole) * 1000) / 10) : '0%'
}

function Fact({ label, note, children }: { label: string; note: string; children: ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-medium text-slate-500">{label}</dt>
      <dd className="mt-1.5 text-[15px] font-semibold tabular-nums text-slate-900">{children}</dd>
      <dd className="mt-0.5 text-xs text-slate-500">{note}</dd>
    </div>
  )
}

function Stacked({ children, note, align = 'right' }: { children: ReactNode; note?: string; align?: 'left' | 'right' }) {
  return (
    <span className={`inline-flex flex-col gap-1 ${align === 'right' ? 'items-end' : 'items-start'}`}>
      <span className="tabular-nums">{children}</span>
      {note ? <span className="text-xs text-slate-500">{note}</span> : null}
    </span>
  )
}

function ContractForm({ projectId, schedule, onDone }: {
  projectId: string
  schedule: PaymentSchedule
  onDone: () => void
}) {
  const [values, setValues] = useState<ContractFormValues>(() => contractFormValues(schedule.contract))
  const save = useSaveContract(projectId)

  const update = (key: keyof ContractFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    save.mutate(values, { onSuccess: onDone })
  }

  return (
    <FormSection title={schedule.contract ? 'Ubah kontrak kontraktor' : 'Kontrak kontraktor'}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <SearchSelect
            {...activeVendorOptions}
            id="contract-vendor"
            label="Kontraktor"
            placeholder="Cari kode atau nama vendor"
            required
            hint="Vendor aktif dari modul Pengadaan. Jatuh tempo hutang mengikuti termin bayarnya"
            value={values.vendorId}
            onChange={(vendorId) => setValues((current) => ({ ...current, vendorId }))}
          />
          <Field
            id="contract-number"
            label="Nomor SPK atau kontrak"
            required
            autoComplete="off"
            maxLength={CONTRACT_NUMBER_MAX_LENGTH}
            placeholder="SPK/012/2026"
            value={values.number}
            onChange={update('number')}
          />
          <Field
            id="contract-value"
            label="Nilai kontrak"
            type="number"
            inputMode="numeric"
            min={1}
            step={1}
            required
            placeholder="0"
            hint={values.value === '' || Number(values.value) <= 0 ? 'Dalam Rupiah' : rupiah(Number(values.value))}
            value={values.value}
            onChange={update('value')}
          />
          <Field
            id="contract-retention"
            label="Retensi (%)"
            type="number"
            inputMode="decimal"
            min={0}
            max={RETENTION_MAX_PERCENT}
            step={0.5}
            required
            hint="Umumnya 5%, ditahan dari setiap termin sampai serah terima"
            value={values.retentionPercent}
            onChange={update('retentionPercent')}
          />
        </div>

        <Button type="submit" isPending={save.isPending} pendingLabel="Menyimpan kontrak">
          Simpan kontrak
        </Button>
        {save.isError ? <ErrorNote message={errorMessage(save.error)} /> : null}
      </form>
    </FormSection>
  )
}

function TermForm({ projectId, schedule, onDone }: {
  projectId: string
  schedule: PaymentSchedule
  onDone: () => void
}) {
  const [values, setValues] = useState<PaymentTermFormValues>(() => nextTermForm(schedule))
  const create = useCreatePaymentTerm(projectId)
  const remaining = remainingPercent(schedule)

  const update = (key: keyof PaymentTermFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    create.mutate(values, { onSuccess: onDone })
  }

  return (
    <FormSection title="Termin baru">
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <Field
            id="term-progress"
            label="Syarat progres (%)"
            type="number"
            inputMode="numeric"
            min={0}
            max={100}
            step={1}
            required
            hint={`Progres berbobot sekarang ${schedule.progress}%`}
            value={values.requiredProgress}
            onChange={update('requiredProgress')}
          />
          <Field
            id="term-percent"
            label="Nilai (% kontrak)"
            type="number"
            inputMode="decimal"
            min={0.01}
            max={remaining}
            step={0.01}
            required
            hint={`Sisa ${formatPercent(remaining)} dari nilai kontrak`}
            value={values.percent}
            onChange={update('percent')}
          />
        </div>

        <Button type="submit" isPending={create.isPending} pendingLabel="Menyimpan termin">
          Simpan termin
        </Button>
        {create.isError ? <ErrorNote message={errorMessage(create.error)} /> : null}
      </form>
    </FormSection>
  )
}
