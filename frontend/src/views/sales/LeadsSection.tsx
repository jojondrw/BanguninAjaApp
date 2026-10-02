import { ChevronRight } from 'lucide-react'
import { type ChangeEvent, type FormEvent, Fragment, useState } from 'react'

import {
  useCreateLead,
  useDeleteLead,
  useLeadStageCounts,
  useLeads,
  useMoveLeadStage,
  useUpdateLead,
} from '../../controllers/useSales'
import { projectOptions } from '../../models/lookupApi'
import {
  EMPTY_LEAD_FORM,
  LEAD_CONTACT_MAX_LENGTH,
  LEAD_NAME_MAX_LENGTH,
  LEAD_SOURCE_MAX_LENGTH,
  LEAD_STAGES,
  LEAD_STAGE_LABEL,
  LEAD_STAGE_TONE,
  leadToCustomerForm,
  leadToForm,
  nextLeadStage,
  todayInput,
  type Lead,
  type LeadFormValues,
  type LeadStage,
  type LeadStageCounts,
} from '../../models/sales'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import {
  Chip,
  FormPanel,
  FormToggle,
  Pager,
  SelectField,
  Toolbar,
  ToolbarInput,
} from '../components/RecordControls'
import { CustomerForm } from './CustomerForm'
import { ActionGroup, ConfirmAction, TextAction } from './RowActions'
import { LookupName } from '../components/LookupName'
import { SearchSelect } from '../components/SearchSelect'
import { PAGE_SIZE, pageAfterRemoval, refusalText } from './salesShared'

const SOURCE_SUGGESTIONS = ['Pameran', 'Iklan online', 'Media sosial', 'Rujukan', 'Datang langsung']

type StageFilter = LeadStage | ''

type Panel =
  | { kind: 'create' }
  | { kind: 'edit'; lead: Lead }
  | { kind: 'convert'; lead: Lead }

function stageChipClass(isActive: boolean): string {
  return `inline-flex h-8 items-center gap-1.5 rounded-full px-3 text-[13px] font-medium whitespace-nowrap
          transition-[background-color,color,transform] motion-safe:active:scale-[0.97] ${
            isActive
              ? 'bg-slate-900 text-white'
              : 'bg-white text-slate-600 shadow-hairline hover:bg-slate-50 hover:text-slate-900'
          }`
}

function StageChip({ label, count, isActive, onClick }: {
  label: string
  count: number | undefined
  isActive: boolean
  onClick: () => void
}) {
  return (
    <button type="button" aria-pressed={isActive} onClick={onClick} className={stageChipClass(isActive)}>
      {label}{' '}
      {count === undefined ? null : (
        <span className={`tabular-nums ${isActive ? 'text-white/75' : 'text-slate-500'}`}>{number(count)}</span>
      )}
    </button>
  )
}

// Saluran prospek: tahap maju dihubungkan panah, Batal dipisah di ujung karena
// bisa terjadi dari tahap mana pun. Menekan tahap sekaligus menyaring tabel.
function StagePipeline({ counts, active, onChange }: {
  counts: LeadStageCounts | undefined
  active: StageFilter
  onChange: (stage: StageFilter) => void
}) {
  const total = counts ? LEAD_STAGES.reduce((sum, stage) => sum + counts[stage], 0) : undefined
  const forward = LEAD_STAGES.filter((stage) => stage !== 'cancelled')

  return (
    <div className="mb-4 flex flex-wrap items-center gap-1.5" role="group" aria-label="Saring menurut tahap prospek">
      <StageChip label="Semua" count={total} isActive={active === ''} onClick={() => onChange('')} />
      <span aria-hidden="true" className="mx-1 h-5 w-px bg-slate-200" />
      {forward.map((stage, index) => (
        <Fragment key={stage}>
          {index > 0 ? (
            <ChevronRight aria-hidden="true" className="size-4 text-slate-400" strokeWidth={1.8} />
          ) : null}
          <StageChip
            label={LEAD_STAGE_LABEL[stage]}
            count={counts?.[stage]}
            isActive={active === stage}
            onClick={() => onChange(stage)}
          />
        </Fragment>
      ))}
      <span aria-hidden="true" className="mx-1 h-5 w-px bg-slate-200" />
      <StageChip
        label={LEAD_STAGE_LABEL.cancelled}
        count={counts?.cancelled}
        isActive={active === 'cancelled'}
        onClick={() => onChange('cancelled')}
      />
    </div>
  )
}

function LeadForm({ lead, defaultStage, onSaved, onCancel }: {
  lead: Lead | null
  defaultStage: StageFilter
  onSaved: (message: string) => void
  onCancel: () => void
}) {
  // Prospek baru mengikuti tahap yang sedang disaring supaya langsung terlihat
  // di tabel, kecuali saringan Batal.
  const [values, setValues] = useState<LeadFormValues>(() =>
    lead
      ? leadToForm(lead)
      : { ...EMPTY_LEAD_FORM, stage: defaultStage === '' || defaultStage === 'cancelled' ? 'new' : defaultStage },
  )
  const createLead = useCreateLead()
  const updateLead = useUpdateLead()
  const saving = lead ? updateLead : createLead

  const update = (key: keyof LeadFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (lead) {
      updateLead.mutate(
        { id: lead.id, values },
        { onSuccess: (saved) => onSaved(`Prospek ${saved.name} berhasil diperbarui.`) },
      )
      return
    }
    createLead.mutate(values, {
      onSuccess: () => setValues({ ...EMPTY_LEAD_FORM, stage: values.stage }),
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-sm font-semibold text-slate-900">{lead ? `Ubah prospek ${lead.name}` : 'Prospek baru'}</h3>
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="lead-name"
            label="Nama calon pembeli"
            autoComplete="off"
            autoFocus={lead !== null}
            maxLength={LEAD_NAME_MAX_LENGTH}
            required
            value={values.name}
            onChange={update('name')}
          />
          <Field
            id="lead-contact"
            label="Kontak (opsional)"
            placeholder="0812xxxx atau email"
            autoComplete="off"
            maxLength={LEAD_CONTACT_MAX_LENGTH}
            hint="Telepon atau email, dipakai saat dijadikan pelanggan"
            value={values.contact}
            onChange={update('contact')}
          />
          <SearchSelect
            {...projectOptions}
            id="lead-project"
            label="Minat proyek (opsional)"
            placeholder="Cari proyek"
            allowEmpty
            emptyLabel="Belum menentukan proyek"
            value={values.projectId}
            onChange={(projectId) => setValues((current) => ({ ...current, projectId }))}
          />
          <Field
            id="lead-source"
            label="Sumber (opsional)"
            placeholder="Pameran"
            autoComplete="off"
            list="lead-source-options"
            maxLength={LEAD_SOURCE_MAX_LENGTH}
            hint="Dari mana prospek ini datang"
            value={values.source}
            onChange={update('source')}
          />
          <datalist id="lead-source-options">
            {SOURCE_SUGGESTIONS.map((source) => (
              <option key={source} value={source} />
            ))}
          </datalist>
          <SelectField id="lead-stage" label="Tahap" required value={values.stage} onChange={update('stage')}>
            {LEAD_STAGES.map((stage) => (
              <option key={stage} value={stage}>
                {LEAD_STAGE_LABEL[stage]}
              </option>
            ))}
          </SelectField>
          <Field
            id="lead-contacted"
            label="Terakhir dihubungi (opsional)"
            type="date"
            max={todayInput()}
            value={values.lastContactedAt}
            onChange={update('lastContactedAt')}
          />
        </div>

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={saving.isPending} pendingLabel="Menyimpan prospek">
            {lead ? 'Simpan perubahan' : 'Simpan prospek'}
          </Button>
          {lead ? (
            <Button variant="subtle" onClick={onCancel}>
              Batal ubah
            </Button>
          ) : null}
        </div>

        {saving.isError ? <ErrorNote message={errorMessage(saving.error)} /> : null}
        {createLead.isSuccess && !lead ? (
          <SuccessNote
            message={`Prospek ${createLead.data.name} dicatat di tahap ${LEAD_STAGE_LABEL[createLead.data.stage]}.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}

export function LeadsSection() {
  const [stage, setStage] = useState<StageFilter>('')
  const [projectId, setProjectId] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [panel, setPanel] = useState<Panel | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const counts = useLeadStageCounts(projectId === '' ? undefined : projectId)
  const leads = useLeads({
    stage: stage === '' ? undefined : stage,
    projectId: projectId === '' ? undefined : projectId,
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })
  const moveStage = useMoveLeadStage()
  const deleteLead = useDeleteLead()
  const panelLeadId = panel && panel.kind !== 'create' ? panel.lead.id : null

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  const openPanel = (next: Panel | null) => {
    setNotice(null)
    setPanel(next)
  }

  const togglePanel = (kind: 'edit' | 'convert', lead: Lead) => {
    const isSame = panel?.kind === kind && panel.lead.id === lead.id
    openPanel(isSame ? null : { kind, lead })
  }

  const advance = (lead: Lead, next: LeadStage) => {
    setNotice(null)
    deleteLead.reset()
    moveStage.mutate(
      { lead, stage: next },
      {
        onSuccess: () => {
          setNotice(`${lead.name} dipindah ke tahap ${LEAD_STAGE_LABEL[next]}.`)
          if (stage !== '') {
            setPage((current) => pageAfterRemoval(current, leads.data?.items.length ?? 0))
          }
        },
      },
    )
  }

  const askDelete = (lead: Lead) => {
    setNotice(null)
    moveStage.reset()
    deleteLead.reset()
    setAskingId(lead.id)
  }

  const confirmDelete = (lead: Lead) => {
    deleteLead.mutate(lead, {
      onSuccess: () => {
        setNotice(`Prospek ${lead.name} berhasil dihapus.`)
        setPage((current) => pageAfterRemoval(current, leads.data?.items.length ?? 0))
        if (panelLeadId === lead.id) {
          setPanel(null)
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  const emptyMessage =
    stage === '' && search.trim() === '' && projectId === ''
      ? 'Belum ada prospek. Catat calon pembeli pertama lewat tombol Tambah prospek.'
      : 'Tidak ada prospek yang cocok dengan penyaringan ini.'

  return (
    <Card
      title="Prospek"
      description="Calon pembeli dari pameran, iklan, atau rujukan. Majukan tahapnya sampai berhasil, lalu jadikan pelanggan."
    >
      <StagePipeline counts={counts.data} active={stage} onChange={filterChanged(setStage)} />

      <Toolbar>
        <ToolbarInput
          id="lead-search"
          label="Cari prospek"
          type="search"
          placeholder="Cari nama atau kontak"
          maxLength={LEAD_NAME_MAX_LENGTH}
          className="w-56"
          value={search}
          onChange={(event) => filterChanged(setSearch)(event.target.value)}
        />
        <SearchSelect
          {...projectOptions}
          id="lead-project-filter"
          label="Saring menurut minat proyek"
          compact
          allowEmpty
          emptyLabel="Semua proyek"
          className="w-52"
          value={projectId}
          onChange={(value) => filterChanged(setProjectId)(value)}
        />
        <FormToggle
          isOpen={panel !== null}
          openLabel="Tambah prospek"
          onToggle={() => openPanel(panel === null ? { kind: 'create' } : null)}
        />
      </Toolbar>

      {panel?.kind === 'create' || panel?.kind === 'edit' ? (
        <LeadForm
          key={panel.kind === 'edit' ? panel.lead.id : 'new'}
          lead={panel.kind === 'edit' ? panel.lead : null}
          defaultStage={stage}
          onCancel={() => setPanel(null)}
          onSaved={(message) => {
            setPanel(null)
            setNotice(message)
          }}
        />
      ) : null}
      {panel?.kind === 'convert' ? (
        <CustomerForm
          key={`convert-${panel.lead.id}`}
          customer={null}
          initial={leadToCustomerForm(panel.lead)}
          title={`Jadikan pelanggan: ${panel.lead.name}`}
          onCancel={() => setPanel(null)}
          onSaved={(saved) => {
            setPanel(null)
            setNotice(`${saved.name} sudah terdaftar sebagai pelanggan. Buat kontraknya di bagian Kontrak.`)
          }}
        />
      ) : null}

      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}
      {moveStage.isError ? (
        <div className="mb-4">
          <ErrorNote message={refusalText(`Tahap ${moveStage.variables.lead.name} gagal dipindah.`, moveStage.error)} />
        </div>
      ) : null}
      {deleteLead.isError ? (
        <div className="mb-4">
          <ErrorNote message={refusalText(`Prospek ${deleteLead.variables.name} tidak bisa dihapus.`, deleteLead.error)} />
        </div>
      ) : null}

      {leads.isPending ? <Loading /> : null}
      {leads.isError ? <LoadFailed onRetry={() => leads.refetch()} /> : null}
      {leads.data ? (
        <>
          <Table
            rows={leads.data.items}
            emptyMessage={emptyMessage}
            columns={[
              {
                header: 'Prospek',
                cell: (row) => (
                  <span className="flex flex-col">
                    <span className="font-medium text-slate-900">{row.name}</span>
                    <span className="text-xs text-slate-500">{row.contact || 'Kontak belum diisi'}</span>
                  </span>
                ),
              },
              {
                header: 'Minat proyek',
                cell: (row) => (row.projectId ? <LookupName source={projectOptions} value={row.projectId} /> : '-'),
              },
              { header: 'Sumber', cell: (row) => row.source || '-' },
              {
                header: 'Tahap',
                cell: (row) => <Chip tone={LEAD_STAGE_TONE[row.stage]}>{LEAD_STAGE_LABEL[row.stage]}</Chip>,
              },
              {
                header: 'Terakhir dihubungi',
                cell: (row) => (row.lastContactedAt ? shortDate(row.lastContactedAt) : 'Belum dicatat'),
              },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => {
                  const next = nextLeadStage(row.stage)
                  const isAsking = askingId === row.id
                  return (
                    <ActionGroup>
                      {next && !isAsking ? (
                        <TextAction
                          label={next === 'won' ? 'Tandai berhasil' : `Pindah ke ${LEAD_STAGE_LABEL[next]}`}
                          pendingLabel="Memindahkan"
                          isPending={moveStage.isPending && moveStage.variables.lead.id === row.id}
                          onClick={() => advance(row, next)}
                        />
                      ) : null}
                      {row.stage === 'won' && !isAsking ? (
                        <RowAction
                          label="Jadikan pelanggan"
                          isActive={panel?.kind === 'convert' && panel.lead.id === row.id}
                          onClick={() => togglePanel('convert', row)}
                        />
                      ) : null}
                      {isAsking ? null : (
                        <RowAction
                          label="Ubah"
                          isActive={panel?.kind === 'edit' && panel.lead.id === row.id}
                          onClick={() => togglePanel('edit', row)}
                        />
                      )}
                      <ConfirmAction
                        label="Hapus"
                        question={`Hapus prospek ${row.name}?`}
                        confirmLabel="Ya, hapus"
                        pendingLabel="Menghapus"
                        isAsking={isAsking}
                        isPending={deleteLead.isPending && deleteLead.variables.id === row.id}
                        onAsk={() => askDelete(row)}
                        onCancel={() => setAskingId(null)}
                        onConfirm={() => confirmDelete(row)}
                      />
                    </ActionGroup>
                  )
                },
              },
            ]}
          />
          <Pager
            page={leads.data.page}
            totalPages={leads.data.totalPages}
            totalItems={leads.data.totalItems}
            unit="prospek"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
