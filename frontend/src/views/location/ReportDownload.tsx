import { FileDown } from 'lucide-react'

import { useDownloadSiteReport } from '../../controllers/useLocationReports'
import { errorMessage } from '../../shared/errorMessage'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { ROW_BUTTON } from './buttons'

export const REPORT_BUTTON_LABEL = 'Unduh laporan (PDF)'
export const REPORT_PENDING_LABEL = 'Menyiapkan PDF...'

export function ReportIcon() {
  return <FileDown aria-hidden="true" className="size-4" strokeWidth={1.8} />
}

// Satu catatan hasil unduhan: gagal dengan pesan dari backend, atau berhasil
// dengan nama berkasnya, sama dengan riwayat di halaman Laporan.
export function DownloadNotes({ fileName, error }: { fileName: string | null | undefined; error: Error | null }) {
  if (error) {
    return <ErrorNote message={`Laporan PDF gagal diunduh: ${errorMessage(error)}`} />
  }
  return fileName ? <SuccessNote message={`Berkas ${fileName} sudah diunduh.`} /> : null
}

// Tombol kecil di baris daftar. Nama baris ikut dibacakan pembaca layar supaya
// tombol yang sama di banyak baris tetap bisa dibedakan.
export function ReportRowButton({ srLabel, isPending, onClick }: {
  srLabel: string
  isPending: boolean
  onClick: () => void
}) {
  return (
    <button type="button" disabled={isPending} aria-busy={isPending} onClick={onClick} className={ROW_BUTTON}>
      {isPending ? 'Mengunduh...' : 'Unduh PDF'}
      <span className="sr-only">: laporan {srLabel}</span>
    </button>
  )
}

// Ditampilkan di atas hasil evaluasi. Evaluasi selalu menyimpan lokasinya, jadi
// laporannya bisa langsung diunduh. Dipasang dengan key id lokasi supaya catatan
// unduhan evaluasi sebelumnya tidak terbawa.
export function SiteReportPanel({ savedLocationId }: { savedLocationId: string }) {
  const download = useDownloadSiteReport()

  return (
    <section className="space-y-3 rounded-2xl bg-white px-5 py-4 shadow-hairline">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-[15px] font-semibold text-slate-900">Laporan analisis lokasi</h2>
          <p className="mt-0.5 text-[13px] text-slate-500">
            Hasil evaluasi ini sudah tersimpan. Unduh skor, penanda risiko, dan konteks regulasinya sebagai PDF.
          </p>
        </div>
        <Button
          variant="subtle"
          isPending={download.isPending}
          pendingLabel={REPORT_PENDING_LABEL}
          onClick={() => download.mutate(savedLocationId)}
        >
          <ReportIcon />
          {REPORT_BUTTON_LABEL}
        </Button>
      </div>
      <DownloadNotes fileName={download.data} error={download.error} />
    </section>
  )
}
