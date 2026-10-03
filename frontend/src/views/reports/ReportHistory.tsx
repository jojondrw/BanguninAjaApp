import { useState } from 'react'

import { useDeleteReport, useDownloadReport, useReports } from '../../controllers/useReports'
import {
  createdAtLabel,
  REPORT_FORMAT_LABEL,
  REPORT_TYPES,
  reportPeriodLabel,
  reportTypeLabel,
  type Report,
} from '../../models/report'
import { errorMessage } from '../../shared/errorMessage'
import { Card, type Column, LoadFailed, Loading, Table } from '../components/Data'
import { Chip, FilterSelect, Pager } from '../components/RecordControls'
import { ActionGroup, ConfirmAction, RowNotice, TextAction } from '../hr/RowActions'

const PAGE_SIZE = 10

export function ReportHistory() {
  const [type, setType] = useState('')
  const [page, setPage] = useState(1)
  const [asking, setAsking] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const reports = useReports({ type: type || undefined, page, pageSize: PAGE_SIZE })
  const download = useDownloadReport()
  const remove = useDeleteReport()

  const clearNotes = () => {
    setNotice(null)
    download.reset()
    remove.reset()
  }

  const downloadAgain = (report: Report) => {
    clearNotes()
    download.mutate(report, { onSuccess: (fileName) => setNotice(`Berkas ${fileName} sudah diunduh.`) })
  }

  const confirmDelete = (report: Report) => {
    remove.mutate(report.id, {
      onSuccess: () => {
        setAsking(null)
        setNotice(`Laporan "${report.name}" dihapus dari riwayat.`)
      },
      onError: () => setAsking(null),
    })
  }

  const columns: Column<Report>[] = [
    {
      header: 'Nama',
      cell: (report) => (
        <span className="flex flex-col">
          <span className="font-medium text-slate-900">{report.name}</span>
          <span className="text-xs text-slate-500">{report.projectName ? `${report.projectCode} ${report.projectName}` : report.scope || 'Semua proyek'}</span>
        </span>
      ),
    },
    { header: 'Jenis', cell: (report) => reportTypeLabel(report.type) },
    { header: 'Periode', cell: (report) => <span className="whitespace-nowrap">{reportPeriodLabel(report)}</span> },
    { header: 'Format', cell: (report) => <Chip tone="bg-slate-100 text-slate-700">{REPORT_FORMAT_LABEL[report.format] ?? report.format}</Chip> },
    { header: 'Dibuat oleh', cell: (report) => report.createdByName || '-' },
    { header: 'Waktu', cell: (report) => <span className="whitespace-nowrap">{createdAtLabel(report.createdAt)}</span> },
    {
      header: 'Aksi',
      align: 'right',
      cell: (report) => (
        <ActionGroup>
          <TextAction
            label="Unduh lagi"
            srLabel={report.name}
            onClick={() => downloadAgain(report)}
            isPending={download.isPending && download.variables?.id === report.id}
            pendingLabel="Mengunduh"
          />
          <ConfirmAction
            label="Hapus"
            srLabel={report.name}
            question="Hapus dari riwayat?"
            confirmLabel="Ya, hapus"
            pendingLabel="Menghapus"
            isAsking={asking === report.id}
            isPending={remove.isPending && remove.variables === report.id}
            onAsk={() => {
              clearNotes()
              setAsking(report.id)
            }}
            onCancel={() => setAsking(null)}
            onConfirm={() => confirmDelete(report)}
          />
        </ActionGroup>
      ),
    },
  ]

  const failure = download.error ?? remove.error

  return (
    <Card
      title="Riwayat laporan"
      description="Berkas dibuat ulang dari data terbaru setiap kali diunduh."
      action={
        <FilterSelect
          id="report-history-type"
          label="Saring jenis laporan"
          value={type}
          onChange={(event) => {
            setType(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Semua jenis</option>
          {REPORT_TYPES.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </FilterSelect>
      }
    >
      <RowNotice success={notice} error={failure ? errorMessage(failure) : null} />
      {reports.isPending ? (
        <Loading />
      ) : reports.isError ? (
        <LoadFailed onRetry={() => void reports.refetch()} />
      ) : (
        <>
          <Table
            rows={reports.data.items}
            columns={columns}
            emptyMessage={type ? 'Belum ada laporan jenis ini.' : 'Belum ada laporan. Buat laporan pertama dari formulir di atas.'}
          />
          <Pager
            page={reports.data.page}
            totalPages={reports.data.totalPages}
            totalItems={reports.data.totalItems}
            unit="laporan"
            onChange={setPage}
          />
        </>
      )}
    </Card>
  )
}
