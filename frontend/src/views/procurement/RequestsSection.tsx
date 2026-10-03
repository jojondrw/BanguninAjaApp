import { type ChangeEvent, type FormEvent, useState } from 'react'

import {
  useCreatePurchaseRequest,
  useDeletePurchaseRequest,
  usePurchaseRequest,
  usePurchaseRequests,
  useUpdatePurchaseRequest,
  useUpdatePurchaseRequestStatus,
} from '../../controllers/useProcurement'
import {
  REQUEST_NOTE_MAX_LENGTH,
  REQUEST_NUMBER_MAX_LENGTH,
  REQUEST_STATUS_LABEL,
  REQUEST_STATUS_TONE,
  REQUEST_TRANSITIONS,
  emptyPurchaseRequestForm,
  purchaseOrderFormFromRequest,
  purchaseRequestFormFrom,
  type PurchaseOrderFormValues,
  type PurchaseRequestDetail,
  type PurchaseRequestFormValues,
  type PurchaseRequestStatus,
  type PurchaseRequestStatusChange,
} from '../../models/procurement'
import { materialOptions, projectOptions, unitOfMeasureOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { number, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { Chip, FilterChips, Pager, RowAction } from '../components/ListTools'
import { LookupName } from '../components/LookupName'
import { ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { useLookupLabel } from '../components/searchSelectLogic'
import { MaterialLinesEditor } from './MaterialLines'
import { PAGE_SIZE, useMaterialLookups } from './lookup'
import { ConfirmAction, Facts, TextAreaField } from './parts'

const REQUEST_FILTERS: { value: PurchaseRequestStatus | ''; label: string }[] = [
  { value: '', label: 'Semua' },
  { value: 'draft', label: REQUEST_STATUS_LABEL.draft },
  { value: 'submitted', label: REQUEST_STATUS_LABEL.submitted },
  { value: 'approved', label: REQUEST_STATUS_LABEL.approved },
  { value: 'rejected', label: REQUEST_STATUS_LABEL.rejected },
  { value: 'completed', label: REQUEST_STATUS_LABEL.completed },
]

const STATUS_HINT: Record<PurchaseRequestStatus, string> = {
  draft: 'Draf masih bisa diubah atau dihapus. Ajukan supaya bisa disetujui.',
  submitted: 'Menunggu persetujuan. Setelah disetujui, pesanan pembelian bisa dibuat dari permintaan ini.',
  approved: 'Sudah disetujui. Buat pesanan pembelian dari permintaan ini, lalu tandai selesai kalau semua kebutuhan sudah dipesan.',
  rejected: 'Permintaan ditolak. Buat permintaan baru kalau kebutuhannya masih ada.',
  completed: 'Permintaan selesai.',
}

// Langkah status yang perlu konfirmasi karena tidak bisa dikembalikan.
const STATUS_CONFIRM: Record<Exclude<PurchaseRequestStatusChange, 'submitted'>, {
  label: string
  prompt: string
  confirmLabel: string
  pendingLabel: string
  tone: 'neutral' | 'danger'
}> = {
  approved: {
    label: 'Setujui',
    prompt: 'Setujui permintaan ini?',
    confirmLabel: 'Ya, setujui',
    pendingLabel: 'Menyetujui',
    tone: 'neutral',
  },
  rejected: {
    label: 'Tolak',
    prompt: 'Tolak permintaan ini? Permintaan yang ditolak tidak bisa diajukan lagi.',
    confirmLabel: 'Ya, tolak',
    pendingLabel: 'Menolak',
    tone: 'danger',
  },
  completed: {
    label: 'Tandai selesai',
    prompt: 'Tandai selesai? Statusnya tidak bisa dikembalikan.',
    confirmLabel: 'Ya, tandai selesai',
    pendingLabel: 'Menyimpan',
    tone: 'neutral',
  },
}

const CONFIRMED_STEPS = ['approved', 'rejected', 'completed'] as const

function PurchaseRequestFormCard({ editing, onSaved, onCancel }: {
  editing: PurchaseRequestDetail | null
  onSaved: (request: PurchaseRequestDetail) => void
  onCancel?: () => void
}) {
  const [values, setValues] = useState<PurchaseRequestFormValues>(() =>
    editing ? purchaseRequestFormFrom(editing) : emptyPurchaseRequestForm(),
  )
  const createRequest = useCreatePurchaseRequest()
  const updateRequest = useUpdatePurchaseRequest()
  const isEditing = editing !== null
  const title = editing ? `Ubah permintaan ${editing.number}` : 'Permintaan pembelian baru'
  const isSaving = createRequest.isPending || updateRequest.isPending
  const saveError = isEditing ? updateRequest.error : createRequest.error

  const update = (key: 'number' | 'date' | 'note') =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (editing) {
      updateRequest.mutate({ id: editing.id, values }, { onSuccess: onSaved })
      return
    }
    createRequest.mutate(values, {
      onSuccess: (created) => {
        setValues(emptyPurchaseRequestForm())
        onSaved(created)
      },
    })
  }

  const cancelAction = onCancel ? (
    <Button variant="subtle" onClick={onCancel}>
      Batal ubah
    </Button>
  ) : undefined

  return (
    <Card
      title={title}
      description={
        isEditing
          ? 'Hanya permintaan berstatus Draf yang bisa diubah. Semua baris material diganti dengan isi formulir ini.'
          : 'Permintaan baru berstatus Draf dan tercatat atas nama akun yang sedang masuk.'
      }
      action={cancelAction}
    >
      <form onSubmit={submit} className="space-y-5">
        <div className="grid gap-4 md:grid-cols-3">
          <Field
            id="pr-number"
            label="Nomor permintaan"
            placeholder="PR-2026-001"
            autoComplete="off"
            maxLength={REQUEST_NUMBER_MAX_LENGTH}
            required
            value={values.number}
            onChange={update('number')}
          />
          <SearchSelect
            {...projectOptions}
            id="pr-project"
            label="Proyek"
            placeholder="Cari nama atau kode proyek"
            required
            value={values.projectId}
            onChange={(projectId) => setValues((current) => ({ ...current, projectId }))}
          />
          <Field
            id="pr-date"
            label="Tanggal permintaan"
            type="date"
            required
            value={values.date}
            onChange={update('date')}
          />
        </div>

        <TextAreaField
          id="pr-note"
          label="Catatan (opsional)"
          placeholder="Misalnya untuk pekerjaan apa dan kapan barang harus sampai di lokasi"
          hint="Permintaan tidak punya kolom tanggal dibutuhkan, jadi tulis tenggatnya di sini."
          maxLength={REQUEST_NOTE_MAX_LENGTH}
          value={values.note}
          onChange={update('note')}
        />

        <MaterialLinesEditor
          idPrefix="pr"
          lines={values.items}
          showPrice={false}
          onChange={(items) => setValues((current) => ({ ...current, items }))}
        />

        <Button type="submit" isPending={isSaving} pendingLabel={isEditing ? 'Menyimpan perubahan' : 'Menyimpan permintaan'}>
          {isEditing ? 'Simpan perubahan' : 'Simpan permintaan'}
        </Button>

        {saveError ? <ErrorNote message={errorMessage(saveError)} /> : null}
        {!isEditing && createRequest.isSuccess ? (
          <SuccessNote
            message={`Permintaan ${createRequest.data.number} tersimpan sebagai draf. Ajukan lewat rincian permintaan supaya bisa disetujui.`}
          />
        ) : null}
      </form>
    </Card>
  )
}

type RequestConfirm = (typeof CONFIRMED_STEPS)[number] | 'delete'

function PurchaseRequestDetailCard({ id, onDeleted, onCreateOrder }: {
  id: string
  onDeleted: (number: string) => void
  onCreateOrder: (values: PurchaseOrderFormValues, requestNumber: string) => void
}) {
  const [isEditing, setIsEditing] = useState(false)
  const [confirming, setConfirming] = useState<RequestConfirm | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const request = usePurchaseRequest(id)
  const materials = useMaterialLookups(request.data?.items.map((item) => item.materialId) ?? [])
  const updateStatus = useUpdatePurchaseRequestStatus()
  const deleteRequest = useDeletePurchaseRequest()
  const projectName = useLookupLabel(projectOptions, request.data?.projectId) ?? '-'

  if (request.isPending) {
    return (
      <Card title="Rincian permintaan">
        <Loading />
      </Card>
    )
  }

  if (request.isError) {
    return (
      <Card title="Rincian permintaan">
        <LoadFailed onRetry={() => request.refetch()} />
      </Card>
    )
  }

  const detail = request.data
  const allowed = REQUEST_TRANSITIONS[detail.status]

  if (isEditing) {
    return (
      <PurchaseRequestFormCard
        editing={detail}
        onSaved={(saved) => {
          setIsEditing(false)
          setNotice(`Perubahan permintaan ${saved.number} tersimpan.`)
        }}
        onCancel={() => setIsEditing(false)}
      />
    )
  }

  const startAction = () => {
    setNotice(null)
    updateStatus.reset()
    deleteRequest.reset()
  }

  const changeStatus = (status: PurchaseRequestStatusChange) =>
    updateStatus.mutate(
      { id: detail.id, status },
      {
        onSuccess: (saved) =>
          setNotice(`Permintaan ${saved.number} sekarang berstatus ${REQUEST_STATUS_LABEL[saved.status]}.`),
        onSettled: () => setConfirming(null),
      },
    )

  const remove = () =>
    deleteRequest.mutate(detail.id, {
      onSuccess: () => onDeleted(detail.number),
      onSettled: () => setConfirming(null),
    })

  const pendingStatus = updateStatus.isPending ? updateStatus.variables?.status : undefined
  const activeConfirm: RequestConfirm | null =
    confirming ?? (deleteRequest.isPending ? 'delete' : pendingStatus && pendingStatus !== 'submitted' ? pendingStatus : null)

  const deleteConfirm = (
    <ConfirmAction
      size="md"
      tone="danger"
      label="Hapus permintaan"
      prompt="Hapus permintaan draf ini beserta semua barisnya?"
      confirmLabel="Ya, hapus"
      pendingLabel="Menghapus"
      isAsking={confirming === 'delete'}
      isPending={deleteRequest.isPending}
      onAsk={() => {
        startAction()
        setConfirming('delete')
      }}
      onCancel={() => setConfirming(null)}
      onConfirm={remove}
    />
  )

  const stepConfirm = (step: (typeof CONFIRMED_STEPS)[number]) => (
    <ConfirmAction
      key={step}
      size="md"
      tone={STATUS_CONFIRM[step].tone}
      label={STATUS_CONFIRM[step].label}
      prompt={STATUS_CONFIRM[step].prompt}
      confirmLabel={STATUS_CONFIRM[step].confirmLabel}
      pendingLabel={STATUS_CONFIRM[step].pendingLabel}
      isAsking={confirming === step}
      isPending={pendingStatus === step}
      onAsk={() => {
        startAction()
        setConfirming(step)
      }}
      onCancel={() => setConfirming(null)}
      onConfirm={() => changeStatus(step)}
    />
  )

  let actions = null
  if (activeConfirm === 'delete') {
    actions = deleteConfirm
  } else if (activeConfirm) {
    actions = stepConfirm(activeConfirm)
  } else {
    actions = (
      <>
        {detail.status === 'draft' ? (
          <Button
            variant="subtle"
            onClick={() => {
              startAction()
              setIsEditing(true)
            }}
          >
            Ubah permintaan
          </Button>
        ) : null}
        {allowed.includes('submitted') ? (
          <Button
            isPending={pendingStatus === 'submitted'}
            pendingLabel="Mengajukan"
            onClick={() => {
              startAction()
              changeStatus('submitted')
            }}
          >
            Ajukan permintaan
          </Button>
        ) : null}
        {detail.status === 'approved' ? (
          <Button
            isPending={materials.isPending}
            pendingLabel="Mengambil harga material"
            onClick={() => onCreateOrder(purchaseOrderFormFromRequest(detail, materials.byId), detail.number)}
          >
            Buat PO dari permintaan
          </Button>
        ) : null}
        {CONFIRMED_STEPS.filter((step) => allowed.includes(step)).map(stepConfirm)}
        {detail.status === 'draft' ? deleteConfirm : null}
      </>
    )
  }

  return (
    <Card
      title={`Rincian permintaan ${detail.number}`}
      description={`${projectName}, ${shortDate(detail.date)}`}
    >
      <Facts
        items={[
          {
            label: 'Status',
            value: <Chip label={REQUEST_STATUS_LABEL[detail.status]} tone={REQUEST_STATUS_TONE[detail.status]} />,
          },
          { label: 'Tanggal permintaan', value: shortDate(detail.date) },
          { label: 'Jumlah barang', value: number(detail.items.length) },
          { label: 'Catatan', value: detail.note || '-' },
        ]}
      />

      <div className="mb-4 space-y-3">
        <p className="text-[13px] text-slate-600">{STATUS_HINT[detail.status]}</p>
        {detail.status !== 'rejected' && detail.status !== 'completed' ? (
          <div className="flex flex-wrap items-center gap-2">{actions}</div>
        ) : null}
        {updateStatus.isError ? <ErrorNote message={errorMessage(updateStatus.error)} /> : null}
        {deleteRequest.isError ? <ErrorNote message={errorMessage(deleteRequest.error)} /> : null}
        {notice ? <SuccessNote message={notice} /> : null}
      </div>

      <Table
        rows={detail.items}
        emptyMessage="Permintaan ini tidak punya baris material."
        columns={[
          { header: 'Material', cell: (row) => <LookupName source={materialOptions} value={row.materialId} /> },
          { header: 'Jumlah', align: 'right', cell: (row) => number(row.quantity) },
          { header: 'Satuan', cell: (row) => <LookupName source={unitOfMeasureOptions} value={row.unitOfMeasureId} /> },
        ]}
      />
    </Card>
  )
}

export function RequestsSection({ onCreateOrder }: {
  onCreateOrder: (values: PurchaseOrderFormValues, requestNumber: string) => void
}) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<PurchaseRequestStatus | ''>('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const requests = usePurchaseRequests({
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    page,
    pageSize: PAGE_SIZE,
  })

  return (
    <div className="space-y-6">
      {isFormOpen ? (
        <PurchaseRequestFormCard
          editing={null}
          onSaved={(created) => {
            setNotice(null)
            setSelectedId(created.id)
          }}
        />
      ) : null}

      <Card
        title="Permintaan pembelian"
        description="Kebutuhan material dari proyek. Pesanan bisa dibuat dari permintaan yang sudah disetujui."
      >
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <ToolbarInput
              id="request-search"
              label="Cari nomor permintaan"
              type="search"
              placeholder="Cari nomor permintaan"
              maxLength={REQUEST_NUMBER_MAX_LENGTH}
              className="w-52"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value)
                setPage(1)
              }}
            />
            <FilterChips
              filters={REQUEST_FILTERS}
              active={status}
              label="Saring status permintaan"
              onChange={(value) => {
                setStatus(value)
                setPage(1)
              }}
            />
          </div>
          <Button variant={isFormOpen ? 'subtle' : 'primary'} onClick={() => setIsFormOpen((open) => !open)}>
            {isFormOpen ? 'Tutup formulir' : 'Buat permintaan'}
          </Button>
        </div>

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {requests.isPending ? <Loading /> : null}
        {requests.isError ? <LoadFailed onRetry={() => requests.refetch()} /> : null}
        {requests.data ? (
          <>
            <Table
              rows={requests.data.items}
              emptyMessage={
                status === '' && search.trim() === ''
                  ? 'Belum ada permintaan pembelian. Buat lewat tombol Buat permintaan.'
                  : 'Tidak ada permintaan yang cocok dengan penyaringan ini.'
              }
              columns={[
                { header: 'Nomor', cell: (row) => row.number },
                { header: 'Proyek', cell: (row) => row.projectName ?? '-' },
                { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                { header: 'Jumlah barang', align: 'right', cell: (row) => number(row.itemCount) },
                { header: 'Catatan', cell: (row) => row.note || '-' },
                {
                  header: 'Status',
                  cell: (row) => (
                    <Chip label={REQUEST_STATUS_LABEL[row.status]} tone={REQUEST_STATUS_TONE[row.status]} />
                  ),
                },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Rincian"
                      isActive={row.id === selectedId}
                      onClick={() => setSelectedId((current) => (current === row.id ? null : row.id))}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={requests.data.page}
              totalPages={requests.data.totalPages}
              totalItems={requests.data.totalItems}
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <PurchaseRequestDetailCard
          key={selectedId}
          id={selectedId}
          onDeleted={(deletedNumber) => {
            setSelectedId(null)
            setNotice(`Permintaan ${deletedNumber} dihapus.`)
          }}
          onCreateOrder={onCreateOrder}
        />
      ) : null}
    </div>
  )
}
