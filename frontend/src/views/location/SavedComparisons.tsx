import { type FormEvent, useState } from 'react'

import {
  useComparisons,
  useDeleteComparison,
  useOpenComparison,
  useRenameComparison,
} from '../../controllers/useLocations'
import {
  COMPARISON_NAME_MAX,
  MIN_COMPARED,
  type Comparison,
  type ComparisonDetail,
} from '../../models/location'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { Pager } from '../components/RecordControls'
import { ROW_BUTTON, ROW_BUTTON_DANGER, ROW_BUTTON_QUIET } from './buttons'

const PAGE_SIZE = 8
const NO_COMPARISONS: Comparison[] = []

interface SavedComparisonsProps {
  openedId: string | null
  onOpened: (detail: ComparisonDetail) => void
  onRenamed: (detail: ComparisonDetail) => void
  onDeleted: (id: string) => void
}

export function SavedComparisons({ openedId, onOpened, onRenamed, onDeleted }: SavedComparisonsProps) {
  const [page, setPage] = useState(1)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [confirmingId, setConfirmingId] = useState<string | null>(null)

  const comparisons = useComparisons({ page, pageSize: PAGE_SIZE })
  const open = useOpenComparison()
  const rename = useRenameComparison()
  const remove = useDeleteComparison()
  const items = comparisons.data?.items ?? NO_COMPARISONS

  // Satu catatan hasil saja di atas daftar, supaya catatan lama tidak
  // menumpuk setelah pengguna melakukan tindakan lain.
  function resetNotes() {
    open.reset()
    rename.reset()
    remove.reset()
  }

  function startRename(id: string) {
    resetNotes()
    setConfirmingId(null)
    setRenamingId(id)
  }

  function startDelete(id: string) {
    resetNotes()
    setRenamingId(null)
    setConfirmingId(id)
  }

  function handleOpen(id: string) {
    resetNotes()
    open.mutate(id, { onSuccess: onOpened })
  }

  function handleRename(id: string, name: string) {
    rename.mutate(
      { id, name },
      {
        onSuccess: (detail) => {
          setRenamingId(null)
          onRenamed(detail)
        },
      },
    )
  }

  function handleDelete(comparison: Comparison) {
    remove.mutate(
      { id: comparison.id, name: comparison.name },
      {
        onSuccess: () => {
          // Halaman yang isinya baru saja habis dihapus diganti ke halaman
          // sebelumnya, supaya tidak tampil daftar kosong.
          if (items.length === 1 && page > 1) {
            setPage(page - 1)
          }
          onDeleted(comparison.id)
        },
        onSettled: () => setConfirmingId(null),
      },
    )
  }

  return (
    <Card title="Perbandingan tersimpan" description="Buka lagi pilihan yang pernah disimpan, ganti namanya, atau hapus.">
      <div className="space-y-3">
        {open.isError ? <ErrorNote message={errorMessage(open.error)} /> : null}
        {remove.isError ? <ErrorNote message={errorMessage(remove.error)} /> : null}
        {rename.isSuccess ? <SuccessNote message={`Nama perbandingan diganti menjadi "${rename.data.name}".`} /> : null}
        {remove.isSuccess ? <SuccessNote message={`Perbandingan "${remove.variables.name}" dihapus.`} /> : null}

        {comparisons.isPending ? <Loading label="Mengambil perbandingan tersimpan..." /> : null}
        {comparisons.isError ? <LoadFailed onRetry={() => comparisons.refetch()} /> : null}
        {comparisons.data && items.length === 0 ? (
          <Empty message="Belum ada perbandingan tersimpan. Pilih minimal 2 lokasi, lalu simpan perbandingannya di bawah tabel." />
        ) : null}

        {items.length > 0 ? (
          <ul className="divide-y divide-slate-100">
            {items.map((comparison) => {
              const isOpening = open.isPending && open.variables === comparison.id
              const isDeleting = remove.isPending && remove.variables.id === comparison.id
              const isIncomplete = comparison.locationCount < MIN_COMPARED

              return (
                <li key={comparison.id} className="py-3 first:pt-1 last:pb-0">
                  <p className="flex flex-wrap items-baseline gap-x-2 text-[13px] font-medium text-slate-900">
                    <span className="min-w-0 break-words">{comparison.name}</span>
                    {comparison.id === openedId ? (
                      <span className="text-[11px] font-medium text-navy-700">sedang dibuka</span>
                    ) : null}
                  </p>
                  <p className="mt-0.5 text-xs text-slate-500 tabular-nums">
                    {comparison.locationCount} lokasi · diubah {shortDate(comparison.updatedAt)}
                  </p>

                  {isIncomplete ? (
                    <p className="mt-1.5 text-xs text-amber-800">
                      Tinggal {comparison.locationCount} lokasi karena lokasi lainnya sudah dihapus. Buka, pilih lokasi
                      lain, lalu perbarui.
                    </p>
                  ) : null}

                  {confirmingId === comparison.id || isDeleting ? (
                    // Hapus tidak bisa dibatalkan, jadi diminta konfirmasi sekali
                    // lagi di baris yang sama.
                    <div className="mt-2 flex flex-wrap items-center gap-2">
                      <span className="text-xs text-slate-600">Hapus perbandingan ini? Lokasi tersimpannya tetap ada.</span>
                      <button
                        type="button"
                        disabled={isDeleting}
                        aria-busy={isDeleting}
                        onClick={() => handleDelete(comparison)}
                        className={ROW_BUTTON_DANGER}
                      >
                        {isDeleting ? 'Menghapus...' : 'Ya, hapus'}
                      </button>
                      {isDeleting ? null : (
                        <button type="button" onClick={() => setConfirmingId(null)} className={ROW_BUTTON_QUIET}>
                          Batal
                        </button>
                      )}
                    </div>
                  ) : (
                    <div className="mt-2 flex flex-wrap gap-1.5">
                      <button
                        type="button"
                        disabled={isOpening}
                        aria-busy={isOpening}
                        onClick={() => handleOpen(comparison.id)}
                        className={ROW_BUTTON}
                      >
                        {isOpening ? 'Membuka...' : 'Buka'}
                        <span className="sr-only">: {comparison.name}</span>
                      </button>
                      {isIncomplete ? null : (
                        <button
                          type="button"
                          aria-expanded={renamingId === comparison.id}
                          onClick={() => (renamingId === comparison.id ? setRenamingId(null) : startRename(comparison.id))}
                          className={ROW_BUTTON}
                        >
                          Ganti nama<span className="sr-only">: {comparison.name}</span>
                        </button>
                      )}
                      <button type="button" onClick={() => startDelete(comparison.id)} className={ROW_BUTTON}>
                        Hapus<span className="sr-only">: {comparison.name}</span>
                      </button>
                    </div>
                  )}

                  {renamingId === comparison.id ? (
                    <RenameForm
                      comparison={comparison}
                      isPending={rename.isPending}
                      error={rename.isError ? errorMessage(rename.error) : null}
                      onSubmit={(name) => handleRename(comparison.id, name)}
                      onCancel={() => setRenamingId(null)}
                    />
                  ) : null}
                </li>
              )
            })}
          </ul>
        ) : null}
      </div>

      {comparisons.data ? (
        <Pager
          page={comparisons.data.page}
          totalPages={comparisons.data.totalPages}
          totalItems={comparisons.data.totalItems}
          unit="perbandingan"
          onChange={setPage}
        />
      ) : null}
    </Card>
  )
}

function RenameForm({ comparison, isPending, error, onSubmit, onCancel }: {
  comparison: Comparison
  isPending: boolean
  error: string | null
  onSubmit: (name: string) => void
  onCancel: () => void
}) {
  const [name, setName] = useState(comparison.name)
  const [isBlank, setIsBlank] = useState(false)
  const fieldId = `rename-comparison-${comparison.id}`

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const trimmed = name.trim()
    setIsBlank(trimmed === '')
    if (trimmed !== '') {
      onSubmit(trimmed)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="mt-3 space-y-3 rounded-xl bg-slate-50 p-3">
      <Field
        id={fieldId}
        label="Nama baru"
        value={name}
        onChange={(event) => setName(event.target.value)}
        required
        maxLength={COMPARISON_NAME_MAX}
        hint={`Maksimal ${COMPARISON_NAME_MAX} karakter.`}
      />
      <div className="flex flex-wrap gap-2">
        <Button type="submit" isPending={isPending} pendingLabel="Menyimpan nama...">
          Simpan nama
        </Button>
        <Button variant="subtle" onClick={onCancel}>
          Batal
        </Button>
      </div>
      {isBlank ? <ErrorNote message="Nama perbandingan tidak boleh kosong." /> : null}
      {error ? <ErrorNote message={error} /> : null}
    </form>
  )
}
