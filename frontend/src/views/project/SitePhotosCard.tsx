import { ImagePlus } from 'lucide-react'
import { type ChangeEvent, type FormEvent, useEffect, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useDeleteSitePhoto, useSitePhotos, useSitePhotoUrl, useUploadSitePhoto } from '../../controllers/useSitePhotos'
import { isProjectClosed, type Project } from '../../models/project'
import { SITE_PHOTO_CAPTION_MAX_LENGTH, SITE_PHOTO_TYPES, sitePhotoProblem, type SitePhoto } from '../../models/sitePhoto'
import { errorMessage } from '../../shared/errorMessage'
import { shortDate } from '../../shared/format'
import { todayDate } from '../../shared/localDate'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, Field } from '../components/Form'
import { FormPanel, FormToggle } from '../components/RecordControls'
import { InlineConfirm } from './parts'

// Foto kondisi lapangan. Foto dengan tanggal terbaru menjadi sampul kartu
// proyek, jadi kantor bisa melihat keadaan proyek tanpa membuka laporan.
export function SitePhotosCard({ project }: { project: Project }) {
  const [isUploading, setIsUploading] = useState(false)
  const photos = useSitePhotos(project.id)
  const deletePhoto = useDeleteSitePhoto(project.id)
  const canWrite = useScopedCan('write') && !isProjectClosed(project)

  return (
    <Card
      title="Foto lapangan"
      description="Foto terbaru menjadi sampul proyek di daftar proyek dan Ringkasan"
      action={
        canWrite ? (
          <FormToggle isOpen={isUploading} openLabel="Unggah foto" onToggle={() => setIsUploading((open) => !open)} />
        ) : undefined
      }
    >
      {isUploading && canWrite ? (
        <FormPanel>
          <UploadForm projectId={project.id} onDone={() => setIsUploading(false)} />
        </FormPanel>
      ) : null}

      {deletePhoto.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(deletePhoto.error)} />
        </div>
      ) : null}

      {photos.isPending ? <Loading /> : null}
      {photos.isError ? <LoadFailed onRetry={() => photos.refetch()} /> : null}
      {photos.data && photos.data.length === 0 ? (
        <Empty message="Belum ada foto lapangan. Unggah foto kondisi terkini supaya tim kantor bisa melihat perkembangan proyek." />
      ) : null}
      {photos.data && photos.data.length > 0 ? (
        <ul className="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-4">
          {photos.data.map((photo, index) => (
            <PhotoTile
              key={photo.id}
              photo={photo}
              isCover={index === 0}
              canDelete={canWrite}
              isDeleting={deletePhoto.isPending && deletePhoto.variables === photo.id}
              onDelete={() => deletePhoto.mutate(photo.id)}
            />
          ))}
        </ul>
      ) : null}
    </Card>
  )
}

function PhotoTile({ photo, isCover, canDelete, isDeleting, onDelete }: {
  photo: SitePhoto
  isCover: boolean
  canDelete: boolean
  isDeleting: boolean
  onDelete: () => void
}) {
  const url = useSitePhotoUrl(photo.projectId, photo.id)
  const description = photo.caption || 'Tanpa keterangan'

  return (
    <li className="min-w-0">
      <figure>
        <div className="relative aspect-[4/3] overflow-hidden rounded-xl bg-slate-100">
          {url ? (
            <a href={url} target="_blank" rel="noreferrer" aria-label={`Buka foto ${description} ukuran penuh`}>
              <img src={url} alt={description} className="size-full object-cover motion-safe:animate-fade" />
            </a>
          ) : (
            <div aria-hidden="true" className="size-full animate-pulse" />
          )}
          {isCover ? (
            <span className="absolute top-2 left-2 rounded-full bg-white/90 px-2 py-0.5 text-[11px] font-medium text-navy-700 shadow-sm backdrop-blur">
              Sampul
            </span>
          ) : null}
        </div>
        <figcaption className="mt-2 flex items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="line-clamp-2 text-[13px] text-slate-900">{description}</p>
            <p className="text-xs text-slate-500">{shortDate(photo.takenOn)}</p>
          </div>
          {canDelete ? (
            <InlineConfirm label="Hapus" confirmLabel="Ya, hapus" pendingLabel="Menghapus" isPending={isDeleting} onConfirm={onDelete} />
          ) : null}
        </figcaption>
      </figure>
    </li>
  )
}

function UploadForm({ projectId, onDone }: { projectId: string; onDone: () => void }) {
  const upload = useUploadSitePhoto(projectId)
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState<string | null>(null)
  const [caption, setCaption] = useState('')
  const [takenOn, setTakenOn] = useState(todayDate())
  const [problem, setProblem] = useState<string | null>(null)

  useEffect(() => () => {
    if (preview) {
      URL.revokeObjectURL(preview)
    }
  }, [preview])

  function choose(event: ChangeEvent<HTMLInputElement>) {
    const next = event.target.files?.[0] ?? null
    setFile(next)
    setProblem(next ? sitePhotoProblem(next) : null)
    setPreview(next && !sitePhotoProblem(next) ? URL.createObjectURL(next) : null)
  }

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const found = sitePhotoProblem(file)
    if (found || !file) {
      setProblem(found)
      return
    }
    upload.mutate({ file, caption, takenOn }, { onSuccess: onDone })
  }

  return (
    <form onSubmit={submit} className="grid gap-4 md:grid-cols-[minmax(0,14rem)_minmax(0,1fr)]">
      <label
        htmlFor="foto-lapangan"
        className="relative grid aspect-[4/3] cursor-pointer place-items-center overflow-hidden rounded-xl bg-white text-center
                   shadow-[inset_0_0_0_1px_rgb(0_0_0/0.08)] transition-colors hover:bg-navy-50/50"
      >
        {preview ? (
          <img src={preview} alt="Pratinjau foto yang dipilih" className="absolute inset-0 size-full object-cover" />
        ) : (
          <span className="flex flex-col items-center gap-1.5 px-4 text-xs text-slate-500">
            <ImagePlus aria-hidden="true" className="size-6 text-navy-600" strokeWidth={1.6} />
            Pilih foto JPG, PNG, atau WebP, paling besar 3 MB
          </span>
        )}
        <input
          id="foto-lapangan"
          type="file"
          accept={SITE_PHOTO_TYPES.join(',')}
          onChange={choose}
          className="sr-only"
        />
      </label>
      <div className="flex flex-col gap-3">
        <Field
          id="keterangan-foto"
          label="Keterangan"
          value={caption}
          maxLength={SITE_PHOTO_CAPTION_MAX_LENGTH}
          placeholder="Contoh: pengecoran pelat lantai 2 selesai"
          onChange={(event) => setCaption(event.target.value)}
        />
        <Field
          id="tanggal-foto"
          label="Tanggal foto"
          type="date"
          value={takenOn}
          max={todayDate()}
          onChange={(event) => setTakenOn(event.target.value)}
        />
        {problem ? <ErrorNote message={problem} /> : null}
        {upload.isError ? <ErrorNote message={errorMessage(upload.error)} /> : null}
        <div>
          <Button type="submit" isPending={upload.isPending} pendingLabel="Mengunggah">
            Unggah foto
          </Button>
        </div>
      </div>
    </form>
  )
}
