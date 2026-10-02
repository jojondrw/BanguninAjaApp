import { useState } from 'react'

import {
  useContract,
  useDeleteContract,
  useInstallments,
  useUpdateContractStatus,
} from '../../controllers/useSales'
import {
  CONTRACT_STATUS_LABEL,
  CONTRACT_STATUS_TONE,
  CONTRACT_TRANSITIONS,
  CONTRACT_TYPE_LABEL,
  type Contract,
  type ContractTarget,
} from '../../models/sales'
import { number, rupiah, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { Chip } from '../components/RecordControls'
import { ContractForm } from './ContractForm'
import { InstallmentSchedule } from './InstallmentSchedule'
import { ConfirmPanel } from './RowActions'
import { OPTION_LIMIT, refusalText } from './salesShared'

type ContractAction = ContractTarget | 'delete'

interface ActionCopy {
  trigger: string
  question: string
  confirm: string
  pending: string
  tone: 'primary' | 'danger'
}

const ACTION_COPY: Record<ContractAction, ActionCopy> = {
  active: {
    trigger: 'Aktifkan kontrak',
    question: 'Aktifkan kontrak ini? Setelah aktif, isi kontrak terkunci dan tidak bisa dihapus, tetapi pembayaran cicilan mulai bisa dicatat.',
    confirm: 'Ya, aktifkan',
    pending: 'Mengaktifkan',
    tone: 'primary',
  },
  paid: {
    trigger: 'Tandai lunas',
    question: 'Tandai kontrak ini lunas? Unitnya menjadi Terjual dan kontrak tidak bisa diubah lagi.',
    confirm: 'Ya, tandai lunas',
    pending: 'Menyimpan',
    tone: 'primary',
  },
  cancelled: {
    trigger: 'Batalkan kontrak',
    question: 'Batalkan kontrak ini? Unitnya kembali Tersedia, dan kontrak yang batal tidak bisa diaktifkan atau dihapus lagi.',
    confirm: 'Ya, batalkan kontrak',
    pending: 'Membatalkan',
    tone: 'danger',
  },
  delete: {
    trigger: 'Hapus draf',
    question: 'Hapus draf kontrak ini beserta jadwal cicilannya? Unitnya kembali Tersedia.',
    confirm: 'Ya, hapus draf',
    pending: 'Menghapus',
    tone: 'danger',
  },
}

const FINAL_NOTE: Partial<Record<Contract['status'], string>> = {
  paid: 'Kontrak sudah lunas. Status ini final, isi kontrak dan cicilannya tidak bisa diubah lagi.',
  cancelled: 'Kontrak sudah dibatalkan. Status ini final, isi kontrak dan cicilannya tidak bisa diubah lagi.',
}

function Figure({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-slate-500">{label}</dt>
      <dd className="mt-0.5 text-sm font-medium text-slate-900 tabular-nums">{value}</dd>
    </div>
  )
}

function ContractDetail({ contract, installments, deletion, onDelete }: {
  contract: Contract
  installments: ReturnType<typeof useInstallments>
  deletion: ReturnType<typeof useDeleteContract>
  onDelete: (contract: Contract, onSettled: () => void) => void
}) {
  const [asking, setAsking] = useState<ContractAction | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [notice, setNotice] = useState<string | null>(null)
  const updateStatus = useUpdateContractStatus()

  const items = installments.data?.items ?? []
  const scheduled = items.reduce((total, item) => total + item.amount, 0)
  const paid = items.filter((item) => item.status === 'paid').reduce((total, item) => total + item.amount, 0)
  const unpaidCount = items.filter((item) => item.status !== 'paid').length
  const isDraft = contract.status === 'draft'

  // Lunas hanya ditawarkan setelah jadwal cicilan termuat dan semuanya
  // terbayar, sama dengan pemeriksaan contract_unpaid di backend.
  const canMarkPaid = installments.data !== undefined && unpaidCount === 0
  const transitions = CONTRACT_TRANSITIONS[contract.status].filter((target) => target !== 'paid' || canMarkPaid)
  const forward = transitions.filter((target) => target !== 'cancelled')
  const hasCancel = transitions.includes('cancelled')
  const isDeleting = deletion.isPending && deletion.variables.id === contract.id

  const ask = (action: ContractAction) => {
    setNotice(null)
    updateStatus.reset()
    deletion.reset()
    setIsEditing(false)
    setAsking(action)
  }

  const confirm = (action: ContractAction) => {
    if (action === 'delete') {
      onDelete(contract, () => setAsking(null))
      return
    }
    updateStatus.mutate(
      { id: contract.id, status: action },
      {
        onSuccess: (saved) =>
          setNotice(`Kontrak ${saved.number} sekarang berstatus ${CONTRACT_STATUS_LABEL[saved.status]}.`),
        onSettled: () => setAsking(null),
      },
    )
  }

  const copy = asking ? ACTION_COPY[asking] : null

  return (
    <Card
      title={`Kontrak ${contract.number}`}
      description={`${contract.customerName}, unit ${contract.unitCode}, ${CONTRACT_TYPE_LABEL[contract.type]}, ${shortDate(contract.date)}`}
      action={<Chip tone={CONTRACT_STATUS_TONE[contract.status]}>{CONTRACT_STATUS_LABEL[contract.status]}</Chip>}
    >
      <dl className="mb-5 grid grid-cols-2 gap-4 rounded-xl bg-slate-50 px-4 py-3 md:grid-cols-4">
        <Figure label="Nilai kontrak" value={rupiah(contract.value)} />
        <Figure label="Dijadwalkan" value={installments.data ? rupiah(scheduled) : '...'} />
        <Figure label="Terbayar" value={installments.data ? rupiah(paid) : '...'} />
        <Figure
          label="Belum dijadwalkan"
          value={installments.data ? rupiah(Math.max(0, contract.value - scheduled)) : '...'}
        />
      </dl>

      {copy && asking ? (
        <div className="mb-4">
          <ConfirmPanel
            question={copy.question}
            confirmLabel={copy.confirm}
            pendingLabel={copy.pending}
            tone={copy.tone}
            isPending={asking === 'delete' ? isDeleting : updateStatus.isPending}
            onConfirm={() => confirm(asking)}
            onCancel={() => setAsking(null)}
          />
        </div>
      ) : null}

      {!copy && FINAL_NOTE[contract.status] ? (
        <p className="mb-4 text-[13px] text-slate-600">{FINAL_NOTE[contract.status]}</p>
      ) : null}

      {!copy && !FINAL_NOTE[contract.status] ? (
        <div className="mb-4 flex flex-wrap items-center gap-2">
          {forward.map((target) => (
            <Button key={target} onClick={() => ask(target)}>
              {ACTION_COPY[target].trigger}
            </Button>
          ))}
          {isDraft ? (
            <Button
              variant="subtle"
              onClick={() => {
                setNotice(null)
                setIsEditing((open) => !open)
              }}
            >
              {isEditing ? 'Tutup formulir ubah' : 'Ubah kontrak'}
            </Button>
          ) : null}
          {isDraft ? (
            <Button variant="subtle" onClick={() => ask('delete')}>
              {ACTION_COPY.delete.trigger}
            </Button>
          ) : null}
          {hasCancel ? (
            <Button variant="subtle" onClick={() => ask('cancelled')}>
              {ACTION_COPY.cancelled.trigger}
            </Button>
          ) : null}
          {contract.status === 'active' && installments.data && unpaidCount > 0 ? (
            <p className="text-xs text-slate-500">
              Tandai lunas tersedia setelah {number(unpaidCount)} angsuran yang tersisa dibayar.
            </p>
          ) : null}
        </div>
      ) : null}

      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}
      {updateStatus.isError ? (
        <div className="mb-4">
          <ErrorNote message={refusalText('Status kontrak gagal diubah.', updateStatus.error)} />
        </div>
      ) : null}
      {deletion.isError && deletion.variables.id === contract.id ? (
        <div className="mb-4">
          <ErrorNote message={refusalText('Draf kontrak tidak bisa dihapus.', deletion.error)} />
        </div>
      ) : null}

      {isEditing && isDraft ? (
        <ContractForm
          contract={contract}
          onCancel={() => setIsEditing(false)}
          onSaved={(saved) => {
            setIsEditing(false)
            setNotice(`Draf kontrak ${saved.number} berhasil diperbarui.`)
          }}
        />
      ) : null}

      <h3 className="mb-3 text-sm font-semibold text-slate-900">Jadwal cicilan</h3>
      <InstallmentSchedule contract={contract} installments={installments} />
    </Card>
  )
}

export function ContractDetailCard({ contractId, deletion, onDelete }: {
  contractId: string
  deletion: ReturnType<typeof useDeleteContract>
  onDelete: (contract: Contract, onSettled: () => void) => void
}) {
  const isDeleting = deletion.isPending && deletion.variables.id === contractId
  const contract = useContract(contractId, !isDeleting)
  const installments = useInstallments({ contractId, pageSize: OPTION_LIMIT }, !isDeleting)

  if (contract.isPending) {
    return (
      <Card title="Rincian kontrak">
        <Loading />
      </Card>
    )
  }

  if (contract.isError) {
    return (
      <Card title="Rincian kontrak">
        <LoadFailed onRetry={() => contract.refetch()} />
      </Card>
    )
  }

  return (
    <ContractDetail contract={contract.data} installments={installments} deletion={deletion} onDelete={onDelete} />
  )
}
