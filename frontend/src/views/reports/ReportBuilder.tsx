import { type FormEvent, useState } from 'react'

import { useCreateReport, useReportPreview } from '../../controllers/useReports'
import { projectOptions } from '../../models/lookupApi'
import {
  defaultReportName,
  periodIsOrdered,
  REPORT_FORMATS,
  REPORT_NAME_MAX_LENGTH,
  REPORT_TYPES,
  reportTypeInfo,
  type ReportCriteria,
  type ReportFormat,
  type ReportFormValues,
  type ReportType,
} from '../../models/report'
import { errorMessage } from '../../shared/errorMessage'
import { Card } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { SearchSelect } from '../components/SearchSelect'
import { ReportPreviewPanel } from './ReportPreviewPanel'

const INITIAL_VALUES: ReportFormValues = {
  type: 'arus_kas',
  projectId: '',
  periodStart: '',
  periodEnd: '',
  name: '',
  format: 'pdf',
}

const PERIOD_ORDER_ERROR = 'Tanggal akhir tidak boleh lebih awal dari tanggal awal.'

export function ReportBuilder() {
  const [values, setValues] = useState<ReportFormValues>(INITIAL_VALUES)
  const [projectLabel, setProjectLabel] = useState('')
  const [formError, setFormError] = useState<string | null>(null)
  const [previewCriteria, setPreviewCriteria] = useState<ReportCriteria | null>(null)
  const create = useCreateReport()
  const preview = useReportPreview(previewCriteria)
  const info = reportTypeInfo(values.type) ?? REPORT_TYPES[0]

  const update = (changes: Partial<ReportFormValues>) => {
    setValues((current) => ({ ...current, ...changes }))
    setFormError(null)
    create.reset()
  }

  const changeType = (type: ReportType) => {
    const usesPeriod = reportTypeInfo(type)?.usesPeriod ?? true
    update(usesPeriod ? { type } : { type, periodStart: '', periodEnd: '' })
  }

  const criteria = (): ReportCriteria | null => {
    if (!periodIsOrdered(values.periodStart, values.periodEnd)) {
      setFormError(PERIOD_ORDER_ERROR)
      return null
    }
    return { type: values.type, projectId: values.projectId, periodStart: values.periodStart, periodEnd: values.periodEnd }
  }

  const showPreview = () => {
    const next = criteria()
    if (!next) {
      return
    }
    if (JSON.stringify(next) === JSON.stringify(previewCriteria)) {
      void preview.refetch()
      return
    }
    setPreviewCriteria(next)
  }

  const download = (event: FormEvent) => {
    event.preventDefault()
    if (criteria()) {
      create.mutate({ values, scope: values.projectId ? projectLabel : 'Semua proyek' })
    }
  }

  return (
    <div className="grid gap-6">
      <Card title="Buat laporan" description={info.description}>
        <form onSubmit={download} className="grid gap-4" noValidate>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <SelectField
              id="report-type"
              label="Jenis laporan"
              value={values.type}
              onChange={(event) => changeType(event.target.value as ReportType)}
            >
              {REPORT_TYPES.map((type) => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </SelectField>
            <SearchSelect
              id="report-project"
              label="Proyek"
              value={values.projectId}
              onChange={(projectId, option) => {
                setProjectLabel(option ? `${option.hint ?? ''} ${option.label}`.trim() : '')
                update({ projectId })
              }}
              allowEmpty
              emptyLabel="Semua proyek"
              placeholder="Semua proyek"
              {...projectOptions}
            />
            <SelectField
              id="report-format"
              label="Format berkas"
              value={values.format}
              onChange={(event) => update({ format: event.target.value as ReportFormat })}
              hint={values.format === 'csv' ? 'Pemisah titik koma, terbaca langsung di Excel berbahasa Indonesia.' : undefined}
            >
              {REPORT_FORMATS.map((format) => (
                <option key={format.value} value={format.value}>
                  {format.label}
                </option>
              ))}
            </SelectField>
            <Field
              id="report-name"
              label="Nama laporan"
              value={values.name}
              maxLength={REPORT_NAME_MAX_LENGTH}
              placeholder={defaultReportName(values.type, values.periodStart, values.periodEnd)}
              onChange={(event) => update({ name: event.target.value })}
              hint="Boleh dikosongkan, nama dibuat dari jenis dan periode."
            />
          </div>

          {info.usesPeriod ? (
            <fieldset className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
              <legend className="sr-only">Periode laporan</legend>
              <Field
                id="report-period-start"
                label="Dari tanggal"
                type="date"
                value={values.periodStart}
                max={values.periodEnd || undefined}
                onChange={(event) => update({ periodStart: event.target.value })}
              />
              <Field
                id="report-period-end"
                label="Sampai tanggal"
                type="date"
                value={values.periodEnd}
                min={values.periodStart || undefined}
                onChange={(event) => update({ periodEnd: event.target.value })}
              />
              <p className="self-end pb-2 text-xs text-slate-500 sm:col-span-2">{info.periodHint}</p>
            </fieldset>
          ) : (
            <p className="text-xs text-slate-500">{info.periodHint}</p>
          )}

          <div className="flex flex-wrap gap-2">
            <Button type="submit" isPending={create.isPending} pendingLabel="Menyiapkan berkas">
              Unduh laporan
            </Button>
            <Button variant="subtle" onClick={showPreview} isPending={preview.isFetching} pendingLabel="Memuat pratinjau">
              Lihat pratinjau
            </Button>
          </div>

          {formError ? <ErrorNote message={formError} /> : null}
          {create.isError ? <ErrorNote message={errorMessage(create.error)} /> : null}
          {create.isSuccess ? (
            <SuccessNote message={`Berkas ${create.data} sudah diunduh dan tercatat di riwayat laporan.`} />
          ) : null}
        </form>
      </Card>

      {previewCriteria ? <ReportPreviewPanel preview={preview} /> : null}
    </div>
  )
}
