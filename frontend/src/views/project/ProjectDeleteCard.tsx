import { useNavigate } from 'react-router-dom'

import { useDeleteProject } from '../../controllers/useProjectWorkspace'
import type { DeletedProjectState, Project } from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { Card } from '../components/Data'
import { Button, ErrorNote } from '../components/Form'

const DELETE_CONFIRM_CLASS =
  'inline-flex h-9 items-center justify-center gap-2 rounded-lg bg-red-600 px-3.5 text-[13px] font-medium ' +
  'whitespace-nowrap text-white shadow-button transition-[background-color,transform] duration-150 hover:bg-red-700 ' +
  'motion-safe:active:scale-[0.97] disabled:cursor-not-allowed disabled:bg-red-600/60 disabled:active:scale-100'

// Langkah kedua dari "Hapus proyek" di kepala halaman. Isi akibatnya mengikuti
// foreign key di backend: anak proyek ikut terhapus, data lain dilepas, dan
// unit penjualan atau dokumen pengadaan membuat server menolak.
export function ProjectDeleteCard({ project, onClose }: { project: Project; onClose: () => void }) {
  const deleteProject = useDeleteProject(project.id)
  const navigate = useNavigate()

  // replace dipakai supaya tombol kembali tidak membuka proyek yang sudah hilang.
  const confirm = () =>
    deleteProject.mutate(undefined, {
      onSuccess: () =>
        navigate('/proyek', {
          replace: true,
          state: { deletedProject: `${project.name} (${project.code})` } satisfies DeletedProjectState,
        }),
    })

  return (
    <Card title="Hapus proyek" description="Penghapusan permanen dan tidak bisa dibatalkan.">
      <div className="space-y-3 text-[13px] text-slate-700">
        <p>
          Proyek <span className="font-medium text-slate-900">{project.name}</span> ({project.code}) akan dihapus beserta
          tahap, izin, item RAB, anggaran tahunan, dan laporan proyeknya.
        </p>
        <p>
          Transaksi kas, faktur, lokasi tersimpan, gudang, karyawan, dan peralatan tetap ada, tetapi tidak lagi terhubung
          ke proyek ini.
        </p>
        <p className="text-slate-500">
          Server menolak penghapusan kalau proyek masih punya unit penjualan, permintaan pembelian, atau pesanan
          pembelian. Kalau proyek hanya berhenti, pertimbangkan status Batal supaya riwayatnya tetap tersimpan.
        </p>
      </div>

      <div className="mt-5 flex flex-wrap gap-2">
        <button
          type="button"
          onClick={confirm}
          disabled={deleteProject.isPending}
          aria-busy={deleteProject.isPending}
          className={DELETE_CONFIRM_CLASS}
        >
          {deleteProject.isPending ? (
            <span
              aria-hidden="true"
              className="size-3.5 rounded-full border-2 border-current border-r-transparent motion-safe:animate-spin"
            />
          ) : null}
          {deleteProject.isPending ? 'Menghapus proyek' : 'Ya, hapus proyek'}
        </button>
        {deleteProject.isPending ? null : (
          <Button variant="subtle" onClick={onClose}>
            Batal
          </Button>
        )}
      </div>

      {deleteProject.isError ? (
        <div className="mt-4">
          <ErrorNote message={errorMessage(deleteProject.error)} />
        </div>
      ) : null}
    </Card>
  )
}
