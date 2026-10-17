import { type ChangeEvent, type FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'

import { useCan } from '../../controllers/useAccess'
import {
  useClaimOwnerTerm,
  useCreateOwnerTerm,
  useDeleteOwnerTerm,
  useOwnerSchedule,
  useReleaseOwnerRetention,
  useSaveOwnerContract,
} from '../../controllers/useProjectWorkspace'
import { customerOptions } from '../../models/lookupApi'
import {
  OWNER_STATE_LABEL,
  OWNER_TERM_DAYS_MAX,
  ownerContractFormValues,
  ownerTermNote,
  type OwnerContract,
  type OwnerContractFormValues,
  type OwnerSchedule,
  type OwnerTerm,
} from '../../models/ownerBilling'
import {
  CONTRACT_NUMBER_MAX_LENGTH,
  formatPercent,
  nextTermForm,
  remainingPercent,
  RETENTION_MAX_PERCENT,
  shareOfContract,
  TERM_STATE_TONE,
  type PaymentTermFormValues,
} from '../../models/paymentTerm'
import type { Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah, rupiahShort, shortDate } from '../../shared/format'
import { AccessScope } from '../access/AccessScope'
import { Card, type Column, Empty, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { Money, MoneyText } from '../components/Money'
import { SearchSelect } from '../components/SearchSelect'
import { Chip, FormSection, InlineConfirm } from './parts'
import { Fact, Stacked } from './PaymentTermsCard'

type Panel = 'contract' | 'term' | null

const CLOSED_STATUSES: readonly string[] = ['completed', 'cancelled']

// Sama seperti termin kontraktor, menagih termin butuh hak tulis Proyek:
// Manajer Proyek mengesahkan progres, Keuangan yang menerima pembayarannya.
export function OwnerBillingCard({ project }: { project: Project }) {
  return (
    <AccessScope module="projects">
      <OwnerBillingContent project={project} />
    </AccessScope>
  )
}

function OwnerBillingContent({ project }: { project: Project }) {
  const schedule = useOwnerSchedule(project.id)
  const canWrite = useCan('projects', 'write')
  const [panel, setPanel] = useState<Panel>(null)
  const data = schedule.data

  const toggle = (next: Exclude<Panel, null>) => setPanel((current) => (current === next ? null : next))
  const close = () => setPanel(null)

  return (
    <Card title="Termin pemberi kerja" description={scheduleDescription(data)}>
      {schedule.isPending ? <Loading /> : null}
      {schedule.isError ? <LoadFailed onRetry={() => schedule.refetch()} /> : null}
      {data && !data.contract ? (
        <Empty message="Belum ada kontrak dengan pemberi kerja. Atur kontraknya dulu, lalu susun termin penagihan berdasarkan progres." />
      ) : null}
      {data?.contract ? (
        <ScheduleBody project={project} schedule={data} contract={data.contract} canWrite={canWrite} />
      ) : null}

      {data && canWrite ? <ScheduleActions project={project} schedule={data} panel={panel} onToggle={toggle} /> : null}

      {panel === 'contract' && data ? (
        <ContractForm key={data.contract?.number ?? 'baru'} project={project} schedule={data} onDone={close} />
      ) : null}
      {panel === 'term' && data ? <TermForm projectId={project.id} schedule={data} onDone={close} /> : null}
    </Card>
  )
}

function scheduleDescription(schedule: OwnerSchedule | undefined): string | undefined {
  if (!schedule) {
    return undefined
  }
  if (!schedule.contract) {
    return 'Penagihan ke pemberi kerja mengikuti progres berbobot proyek'
  }
  const { customerName, number, retentionPercent, paymentTermDays } = schedule.contract
  return `${customerName}, kontrak ${number}, retensi ${formatPercent(retentionPercent)}, tempo ${paymentTermDays} hari`
}

function ScheduleBody({ project, schedule, contract, canWrite }: {
  project: Project
  schedule: OwnerSchedule
  contract: OwnerContract
  canWrite: boolean
}) {
  const claim = useClaimOwnerTerm(project.id)
  const remove = useDeleteOwnerTerm(project.id)
  const canOpenBilling = useCan('billing', 'read')
  const isOpen = !CLOSED_STATUSES.includes(project.status)
  const claimed = claim.data?.terms.find((term) => term.id === claim.variables)

  const columns: Column<OwnerTerm>[] = [
    { header: 'Termin', cell: (row) => <span className="font-medium text-slate-900">{row.label}</span> },
    { header: 'Syarat', cell: (row) => (row.kind === 'retention' ? 'Serah terima' : `Progres ${row.requiredProgress}%`) },
    {
      header: 'Nilai',
      align: 'right',
      cell: (row) => (
        <Stacked note={row.kind === 'progress' ? `${formatPercent(row.percent)} kontrak` : undefined}>
          <Money>{rupiahShort(row.grossAmount)}</Money>
        </Stacked>
      ),
    },
    {
      header: 'Ditagihkan',
      align: 'right',
      cell: (row) => (
        <Stacked note={row.retentionAmount > 0 ? `retensi ${rupiahShort(row.retentionAmount)}` : undefined}>
          <Money>{rupiahShort(row.netAmount)}</Money>
        </Stacked>
      ),
    },
    {
      header: 'Status',
      cell: (row) => (
        <Stacked note={ownerTermNote(row, schedule.progress, shortDate)} align="left">
          <Chip tone={TERM_STATE_TONE[row.state]}>{OWNER_STATE_LABEL[row.state]}</Chip>
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
        <Fact label="Nilai kontrak" note={`kontrak ${contract.number}`}>
          <Money>{rupiahShort(contract.value)}</Money>
        </Fact>
        <Fact label="Sudah ditagih" note={`${shareOfContract(schedule.claimedAmount, contract.value)} dari kontrak`}>
          <Money>{rupiahShort(schedule.claimedAmount)}</Money>
        </Fact>
        <Fact label="Sudah diterima" note="lewat modul Tagihan">
          <Money>{rupiahShort(schedule.paidAmount)}</Money>
        </Fact>
        <Fact label="Retensi tertahan" note={retentionNote(schedule)}>
          <Money>{rupiahShort(schedule.retentionHeld)}</Money>
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
          message={`${claimed.label} ditagih. Piutang ${rupiah(claimed.netAmount)} ke ${contract.customerName} tercatat di Tagihan, jatuh tempo ${contract.paymentTermDays} hari lagi.`}
        />
      ) : null}
    </div>
  )
}

function TermAction({ term, canWrite, canDelete, canOpenBilling, isClaiming, isDeleting, onClaim, onDelete }: {
  term: OwnerTerm
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
        label="Tagih"
        confirmLabel={`Ya, tagih ${rupiahShort(term.netAmount)}`}
        pendingLabel="Menagih"
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
  if (term.receivableId && canOpenBilling) {
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
  schedule: OwnerSchedule
  panel: Panel
  onToggle: (panel: Exclude<Panel, null>) => void
}) {
  const release = useReleaseOwnerRetention(project.id)
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
            {panel === 'contract' ? 'Tutup formulir kontrak' : schedule.contract ? 'Ubah kontrak' : 'Atur kontrak'}
          </Button>
        ) : null}
        {schedule.canReleaseRetention ? (
          <InlineConfirm
            tone="primary"
            label="Tagih retensi"
            confirmLabel={`Ya, tagih ${rupiahShort(schedule.retentionHeld)}`}
            pendingLabel="Menagih retensi"
            isPending={release.isPending}
            onConfirm={() => release.mutate()}
          />
        ) : null}
      </div>
      {hint ? (
        <p className="text-[13px] text-slate-500">
          <MoneyText>{hint}</MoneyText>
        </p>
      ) : null}
      {release.isError ? <ErrorNote message={errorMessage(release.error)} /> : null}
      {release.isSuccess ? (
        <SuccessNote message="Retensi ditagih. Piutangnya tercatat di Tagihan untuk diterima dari pemberi kerja." />
      ) : null}
    </div>
  )
}

function releaseHint(project: Project, schedule: OwnerSchedule): string {
  if (schedule.retentionHeld === 0 || schedule.retentionReleased || schedule.canReleaseRetention) {
    return ''
  }
  if (project.status !== 'completed') {
    return `Retensi ${rupiahShort(schedule.retentionHeld)} ditagih setelah proyek selesai dan diserahterimakan.`
  }
  return 'Tagih semua termin progres dulu sebelum menagih retensi.'
}

function retentionNote(schedule: OwnerSchedule): string {
  if (schedule.retentionReleased) {
    return 'sudah ditagih'
  }
  return schedule.retentionHeld > 0 ? 'ditagih setelah serah terima' : 'belum ada yang tertahan'
}

function ContractForm({ project, schedule, onDone }: {
  project: Project
  schedule: OwnerSchedule
  onDone: () => void
}) {
  const [values, setValues] = useState<OwnerContractFormValues>(() =>
    ownerContractFormValues(schedule.contract, project.contractValue),
  )
  const save = useSaveOwnerContract(project.id)

  const update = (key: keyof OwnerContractFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    save.mutate(values, { onSuccess: onDone })
  }

  return (
    <FormSection title={schedule.contract ? 'Ubah kontrak pemberi kerja' : 'Kontrak pemberi kerja'}>
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <SearchSelect
            {...customerOptions}
            id="owner-customer"
            label="Pemberi kerja"
            placeholder="Cari nama pelanggan"
            required
            hint="Dari daftar pelanggan di modul Penjualan. Piutang termin ditagihkan ke pelanggan ini"
            value={values.customerId}
            onChange={(customerId) => setValues((current) => ({ ...current, customerId }))}
          />
          <Field
            id="owner-number"
            label="Nomor kontrak"
            required
            autoComplete="off"
            maxLength={CONTRACT_NUMBER_MAX_LENGTH}
            placeholder="KTR/RSIA/2026"
            value={values.number}
            onChange={update('number')}
          />
          <Field
            id="owner-value"
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
            id="owner-retention"
            label="Retensi (%)"
            type="number"
            inputMode="decimal"
            min={0}
            max={RETENTION_MAX_PERCENT}
            step={0.5}
            required
            hint="Ditahan pemberi kerja dari setiap termin sampai serah terima"
            value={values.retentionPercent}
            onChange={update('retentionPercent')}
          />
          <Field
            id="owner-days"
            label="Tempo bayar (hari)"
            type="number"
            inputMode="numeric"
            min={0}
            max={OWNER_TERM_DAYS_MAX}
            step={1}
            required
            hint="Jatuh tempo piutang dihitung dari tanggal tagih"
            value={values.paymentTermDays}
            onChange={update('paymentTermDays')}
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
  schedule: OwnerSchedule
  onDone: () => void
}) {
  const [values, setValues] = useState<PaymentTermFormValues>(() => nextTermForm(schedule))
  const create = useCreateOwnerTerm(projectId)
  const remaining = remainingPercent(schedule)

  const update = (key: keyof PaymentTermFormValues) => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    create.mutate(values, { onSuccess: onDone })
  }

  return (
    <FormSection title="Termin tagihan baru">
      <form onSubmit={submit} className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <Field
            id="owner-term-progress"
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
            id="owner-term-percent"
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
