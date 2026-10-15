import type { UseQueryResult } from '@tanstack/react-query'

import {
  formatReportValue,
  isNumericColumn,
  reportRows,
  type ReportPreview,
  type ReportPreviewSection,
  type ReportRow,
} from '../../models/report'
import { errorMessage } from '../../shared/errorMessage'
import { number } from '../../shared/format'
import { Card, type Column, Loading, Table } from '../components/Data'
import { ErrorNote } from '../components/Form'

function sectionColumns(section: ReportPreviewSection): Column<ReportRow>[] {
  return section.columns.map((column, index) => ({
    header: column.header,
    align: isNumericColumn(column.kind) ? 'right' : 'left',
    cell: (row) => {
      const text = formatReportValue(column.kind, row.values[index])
      return row.isTotal ? <strong className="font-semibold text-slate-900">{text}</strong> : text
    },
  }))
}

function PreviewSection({ section }: { section: ReportPreviewSection }) {
  const shown = section.rows.length

  return (
    <section aria-label={section.title} className="grid gap-2">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h3 className="text-[13px] font-semibold text-slate-800">{section.title}</h3>
        <p className="text-xs text-slate-500 tabular-nums">
          {shown < section.totalRows
            ? `Menampilkan ${number(shown)} dari ${number(section.totalRows)} baris. Unduh berkas untuk data lengkap.`
            : `${number(section.totalRows)} baris`}
        </p>
      </div>
      <Table rows={reportRows(section)} columns={sectionColumns(section)} emptyMessage="Tidak ada data untuk filter ini." />
    </section>
  )
}

export function ReportPreviewPanel({ preview }: { preview: UseQueryResult<ReportPreview> }) {
  if (preview.isPending) {
    return (
      <Card title="Pratinjau laporan">
        <Loading label="Menyusun pratinjau..." />
      </Card>
    )
  }

  if (preview.isError) {
    return (
      <Card title="Pratinjau laporan">
        <ErrorNote message={errorMessage(preview.error)} />
      </Card>
    )
  }

  const data = preview.data

  return (
    <Card title={`Pratinjau: ${data.title}`} description={data.filters}>
      <div className="grid gap-6">
        {data.sections.map((section) => (
          <PreviewSection key={section.title} section={section} />
        ))}
      </div>
    </Card>
  )
}
