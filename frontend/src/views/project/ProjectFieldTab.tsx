import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import {
  useCreateDailyReport,
  useDailyReports,
  useDeleteDailyReport,
  useUpdateDailyReport,
} from '../../controllers/useProjectWorkspace'
import {
  DAILY_REPORT_EQUIPMENT_MAX_LENGTH,
  DAILY_REPORT_TEXT_MAX_LENGTH,
  DAILY_REPORT_WORKERS_MAX,
  dailyReportFormValues,
  emptyDailyReportForm,
  isRainy,
  recentSummary,
  WEATHER_LABEL,
  WEATHER_OPTIONS,
  weatherSummary,
  type DailyReport,
  type DailyReportFormValues,
  type Weather,
} from '../../models/dailyReport'
import { isProjectClosed, type Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { todayDate } from '../../shared/localDate'
import { Card, LoadFailed, Loading, Table, type Column } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FormPanel, FormToggle, Pager } from '../components/RecordControls'
import { TextAreaField } from '../procurement/parts'
import { FormSection, InlineConfirm } from './parts'
import { EvidenceCard } from './EvidenceCard'
import { SitePhotosCard } from './SitePhotosCard'

const PAGE_SIZE = 14

type UpdateReport = ReturnType<typeof useUpdateDailyReport>

// Laporan harian lapangan: satu laporan per hari. Dipakai kantor untuk tahu
// apa yang terjadi di lokasi, dan sebagai bukti saat ada klaim keterlambatan
// karena cuaca.
export function ProjectFieldTab({ project }: { project: Project }) {
  const [page, setPage] = useState(1)
  const [isWriting, setIsWriting] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const reports = useDailyReports(project.id, page, PAGE_SIZE)
  const updateReport = useUpdateDailyReport(project.id)
  const deleteReport = useDeleteDailyReport(project.id)
  const canWrite = useScopedCan('write') && !isProjectClosed(project)
  const editing = editingId ? reports.data?.items.find((report) => report.id === editingId) : undefined

  const toggleEdit = (reportId: string) => {
    updateReport.reset()
    setIsWriting(false)
    setEditingId((current) => (current === reportId ? null : reportId))
  }

  const columns: Column<DailyReport>[] = [
    { header: 'Tanggal', cell: (row) => <span className="whitespace-nowrap">{shortDate(row.reportDate)}</span> },
    {
      header: 'Cuaca',
      cell: (row) => (
        <span className={isRainy(row.weatherMorning) || isRainy(row.weatherAfternoon) ? 'text-sky-800' : undefined}>
          {weatherSummary(row)}
        </span>
      ),
    },
    { header: 'Pekerja', align: 'right', cell: (row) => <span className="tabular-nums">{number(row.workers)}</span> },
    {
      header: 'Pekerjaan',
      cell: (row) => (
        <div className="max-w-md min-w-56">
          <p className="whitespace-pre-line">{row.workDone}</p>
          {row.issues ? <p className="mt-1 text-xs whitespace-pre-line text-amber-800">Kendala: {row.issues}</p> : null}
        </div>
      ),
    },
  ]
  if (canWrite) {
    columns.push({
      header: 'Aksi',
      align: 'right',
      cell: (row) => (
        <div className="flex items-center justify-end gap-1">
          <RowAction label={editingId === row.id ? 'Tutup' : 'Ubah'} isActive={editingId === row.id} onClick={() => toggleEdit(row.id)} />
          <InlineConfirm
            label="Hapus"
            confirmLabel="Ya, hapus laporan"
            pendingLabel="Menghapus"
            isPending={deleteReport.isPending && deleteReport.variables === row.id}
            onConfirm={() => deleteReport.mutate(row.id)}
          />
        </div>
      ),
    })
  }

  return (
    <div className="space-y-8">
      <SitePhotosCard project={project} />
      <EvidenceCard project={project} />
      <Card
        title="Laporan harian"
        description={reports.data ? recentSummary(reports.data) : 'Cuaca, tenaga kerja, pekerjaan, dan kendala di lokasi setiap hari'}
        action={
          canWrite ? (
            <FormToggle
              isOpen={isWriting}
              openLabel="Tulis laporan"
              onToggle={() => {
                setEditingId(null)
                setIsWriting((open) => !open)
              }}
            />
          ) : undefined
        }
      >
        {isWriting && canWrite ? (
          <FormPanel>
            <NewReportForm projectId={project.id} onDone={() => setIsWriting(false)} />
          </FormPanel>
        ) : null}

        {deleteReport.isError ? (
          <div className="mb-4">
            <ErrorNote message={errorMessage(deleteReport.error)} />
          </div>
        ) : null}

        {reports.isPending ? <Loading /> : null}
        {reports.isError ? <LoadFailed onRetry={() => reports.refetch()} /> : null}
        {reports.data ? (
          <>
            <Table
              rows={reports.data.items}
              columns={columns}
              emptyMessage="Belum ada laporan harian. Laporan pertama biasanya ditulis mandor atau pengawas di akhir hari kerja."
            />
            <Pager
              page={reports.data.page}
              totalPages={reports.data.totalPages}
              totalItems={reports.data.totalItems}
              unit="laporan"
              onChange={setPage}
            />
          </>
        ) : null}

        {editing && canWrite ? (
          <EditReportForm key={editing.id} report={editing} update={updateReport} onDone={() => setEditingId(null)} />
        ) : null}
      </Card>
    </div>
  )
}

function NewReportForm({ projectId, onDone }: { projectId: string; onDone: () => void }) {
  const [values, setValues] = useState<DailyReportFormValues>(emptyDailyReportForm)
  const createReport = useCreateDailyReport(projectId)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createReport.mutate(values, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <ReportFields idPrefix="report-new" values={values} onChange={setValues} />
      <Button type="submit" isPending={createReport.isPending} pendingLabel="Menyimpan laporan">
        Simpan laporan
      </Button>
      {createReport.isError ? <ErrorNote message={errorMessage(createReport.error)} /> : null}
    </form>
  )
}

function EditReportForm({ report, update, onDone }: { report: DailyReport; update: UpdateReport; onDone: () => void }) {
  const [values, setValues] = useState<DailyReportFormValues>(() => dailyReportFormValues(report))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.mutate({ reportId: report.id, values }, { onSuccess: onDone })
  }

  return (
    <FormSection title={`Ubah laporan ${shortDate(report.reportDate)}`}>
      <form onSubmit={submit} className="space-y-4">
        <ReportFields idPrefix="report-edit" values={values} onChange={setValues} />
        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan laporan">
            Simpan perubahan
          </Button>
          <Button variant="subtle" onClick={onDone}>
            Batal
          </Button>
        </div>
        {update.isError ? <ErrorNote message={errorMessage(update.error)} /> : null}
        {update.isSuccess ? <SuccessNote message="Laporan disimpan." /> : null}
      </form>
    </FormSection>
  )
}

function ReportFields({ idPrefix, values, onChange }: {
  idPrefix: string
  values: DailyReportFormValues
  onChange: (update: (current: DailyReportFormValues) => DailyReportFormValues) => void
}) {
  const set =
    (key: keyof DailyReportFormValues) =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) =>
      onChange((current) => ({ ...current, [key]: event.target.value }))

  return (
    <div className="grid gap-4 md:grid-cols-4">
      <Field
        id={`${idPrefix}-date`}
        label="Tanggal"
        type="date"
        required
        max={todayDate()}
        value={values.reportDate}
        onChange={set('reportDate')}
      />
      <WeatherSelect id={`${idPrefix}-morning`} label="Cuaca pagi" value={values.weatherMorning} onChange={set('weatherMorning')} />
      <WeatherSelect
        id={`${idPrefix}-afternoon`}
        label="Cuaca siang"
        value={values.weatherAfternoon}
        onChange={set('weatherAfternoon')}
      />
      <Field
        id={`${idPrefix}-workers`}
        label="Tenaga kerja"
        type="number"
        inputMode="numeric"
        min={0}
        max={DAILY_REPORT_WORKERS_MAX}
        step={1}
        required
        placeholder="0"
        hint="Orang di lokasi hari ini"
        value={values.workers}
        onChange={set('workers')}
      />
      <div className="md:col-span-4">
        <Field
          id={`${idPrefix}-equipment`}
          label="Alat yang bekerja"
          placeholder="1 excavator, 2 dump truck"
          maxLength={DAILY_REPORT_EQUIPMENT_MAX_LENGTH}
          value={values.equipment}
          onChange={set('equipment')}
        />
      </div>
      <div className="md:col-span-2">
        <TextAreaField
          id={`${idPrefix}-work`}
          label="Pekerjaan hari ini"
          placeholder="Pengecoran kolom lantai 2 as A-C"
          required
          rows={3}
          maxLength={DAILY_REPORT_TEXT_MAX_LENGTH}
          value={values.workDone}
          onChange={set('workDone')}
        />
      </div>
      <div className="md:col-span-2">
        <TextAreaField
          id={`${idPrefix}-issues`}
          label="Kendala"
          placeholder="Hujan deras pukul 14.00, pekerjaan berhenti 2 jam"
          hint="Kosongkan kalau tidak ada"
          rows={3}
          maxLength={DAILY_REPORT_TEXT_MAX_LENGTH}
          value={values.issues}
          onChange={set('issues')}
        />
      </div>
    </div>
  )
}

function WeatherSelect({ id, label, value, onChange }: {
  id: string
  label: string
  value: Weather
  onChange: (event: ChangeEvent<HTMLSelectElement>) => void
}) {
  return (
    <SelectField id={id} label={label} value={value} onChange={onChange}>
      {WEATHER_OPTIONS.map((weather) => (
        <option key={weather} value={weather}>
          {WEATHER_LABEL[weather]}
        </option>
      ))}
    </SelectField>
  )
}
