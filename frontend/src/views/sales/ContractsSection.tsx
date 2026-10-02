import { useState } from 'react'

import { useContracts, useDeleteContract } from '../../controllers/useSales'
import {
  CONTRACT_NUMBER_MAX_LENGTH,
  CONTRACT_STATUS_LABEL,
  CONTRACT_STATUS_TONE,
  CONTRACT_TYPE_LABEL,
  type Contract,
  type ContractStatus,
} from '../../models/sales'
import { rupiahShort, shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FilterSelect, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { ContractDetailCard } from './ContractDetail'
import { ContractForm } from './ContractForm'
import { PAGE_SIZE } from './salesShared'

const CONTRACT_STATUSES: ContractStatus[] = ['draft', 'active', 'paid', 'cancelled']

export function ContractsSection() {
  const [status, setStatus] = useState<ContractStatus | ''>('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const contracts = useContracts({
    status: status === '' ? undefined : status,
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })
  const deleteContract = useDeleteContract()

  const deleteDraft = (contract: Contract, onSettled: () => void) => {
    deleteContract.mutate(contract, {
      onSuccess: () => {
        setSelectedId(null)
        setNotice(`Draf kontrak ${contract.number} berhasil dihapus dan unit ${contract.unitCode} kembali tersedia.`)
      },
      onSettled,
    })
  }

  return (
    <div className="space-y-6">
      <Card title="Daftar kontrak" description="Buka Rincian untuk mengubah status, jadwal cicilan, dan pembayarannya.">
        <Toolbar>
          <ToolbarInput
            id="contract-search"
            label="Cari nomor kontrak"
            type="search"
            placeholder="Cari nomor kontrak"
            maxLength={CONTRACT_NUMBER_MAX_LENGTH}
            className="w-52"
            value={search}
            onChange={(event) => {
              setSearch(event.target.value)
              setPage(1)
            }}
          />
          <FilterSelect
            id="contract-status-filter"
            label="Saring menurut status kontrak"
            value={status}
            onChange={(event) => {
              setStatus(event.target.value as ContractStatus | '')
              setPage(1)
            }}
          >
            <option value="">Semua status</option>
            {CONTRACT_STATUSES.map((item) => (
              <option key={item} value={item}>
                {CONTRACT_STATUS_LABEL[item]}
              </option>
            ))}
          </FilterSelect>
          <FormToggle
            isOpen={isFormOpen}
            openLabel="Tambah kontrak"
            onToggle={() => {
              setNotice(null)
              setIsFormOpen((open) => !open)
            }}
          />
        </Toolbar>

        {isFormOpen ? <ContractForm contract={null} /> : null}

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {contracts.isPending ? <Loading /> : null}
        {contracts.isError ? <LoadFailed onRetry={() => contracts.refetch()} /> : null}
        {contracts.data ? (
          <>
            <Table
              rows={contracts.data.items}
              emptyMessage={
                status === '' && search.trim() === ''
                  ? 'Belum ada kontrak. Buat kontrak dari pelanggan dan unit yang tersedia.'
                  : 'Tidak ada kontrak yang cocok dengan penyaringan ini.'
              }
              columns={[
                { header: 'Nomor', cell: (row) => <span className="font-medium text-slate-900">{row.number}</span> },
                { header: 'Pelanggan', cell: (row) => row.customerName },
                { header: 'Unit', cell: (row) => row.unitCode },
                { header: 'Cara bayar', cell: (row) => CONTRACT_TYPE_LABEL[row.type] },
                { header: 'Tanggal', cell: (row) => shortDate(row.date) },
                { header: 'Nilai', align: 'right', cell: (row) => rupiahShort(row.value) },
                {
                  header: 'Status',
                  cell: (row) => (
                    <Chip tone={CONTRACT_STATUS_TONE[row.status]}>{CONTRACT_STATUS_LABEL[row.status]}</Chip>
                  ),
                },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Rincian"
                      isActive={row.id === selectedId}
                      onClick={() => {
                        setNotice(null)
                        setSelectedId((current) => (current === row.id ? null : row.id))
                      }}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={contracts.data.page}
              totalPages={contracts.data.totalPages}
              totalItems={contracts.data.totalItems}
              unit="kontrak"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <ContractDetailCard
          key={selectedId}
          contractId={selectedId}
          deletion={deleteContract}
          onDelete={deleteDraft}
        />
      ) : null}
    </div>
  )
}
