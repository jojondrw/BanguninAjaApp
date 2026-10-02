import { useState } from 'react'

import { useBillingList } from '../../controllers/useBilling'
import {
  BILLING_STATUSES,
  BILLING_STATUS_LABEL,
  REFERENCE_SEARCH_MAX_LENGTH,
  type BillingStatus,
} from '../../models/billing'
import { rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { BalanceCell, DueDateCell, StatusChip } from './BillingCells'
import { BillingDetail } from './BillingDetail'
import { ReceivableForm } from './BillingForms'
import { customerName, type Directory } from './directory'

const PAGE_SIZE = 10

export function ReceivablesSection({ directory, today }: { directory: Directory; today: string }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<BillingStatus | ''>('')
  const [customerId, setCustomerId] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const receivables = useBillingList('receivables', {
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    customerId: customerId === '' ? undefined : customerId,
    page,
    pageSize: PAGE_SIZE,
  })
  const isFiltered = search.trim() !== '' || status !== '' || customerId !== ''

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  return (
    <div className="space-y-6">
      <Card
        title="Daftar piutang"
        description="Uang yang masih harus diterima dari pelanggan, diurutkan dari jatuh tempo paling awal"
      >
        <Toolbar>
          <ToolbarInput
            id="receivable-search"
            label="Cari referensi piutang"
            type="search"
            placeholder="Cari referensi"
            className="w-56"
            maxLength={REFERENCE_SEARCH_MAX_LENGTH}
            value={search}
            onChange={(event) => filterChanged(setSearch)(event.target.value)}
          />
          <FilterSelect
            id="receivable-filter-status"
            label="Saring menurut status piutang"
            value={status}
            onChange={(event) => filterChanged(setStatus)(event.target.value as BillingStatus | '')}
          >
            <option value="">Semua status</option>
            {BILLING_STATUSES.map((option) => (
              <option key={option} value={option}>
                {BILLING_STATUS_LABEL[option]}
              </option>
            ))}
          </FilterSelect>
          <FilterSelect
            id="receivable-filter-customer"
            label="Saring menurut pelanggan"
            value={customerId}
            onChange={(event) => filterChanged(setCustomerId)(event.target.value)}
          >
            <option value="">Semua pelanggan</option>
            {directory.customers.map((customer) => (
              <option key={customer.id} value={customer.id}>
                {customer.name}
              </option>
            ))}
          </FilterSelect>
          <FormToggle isOpen={isFormOpen} openLabel="Catat piutang" onToggle={() => setIsFormOpen((open) => !open)} />
        </Toolbar>

        {isFormOpen ? (
          <FormPanel>
            {directory.isLoading ? (
              <Loading label="Mengambil daftar pelanggan..." />
            ) : (
              <ReceivableForm directory={directory} />
            )}
          </FormPanel>
        ) : null}

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {receivables.isPending ? <Loading /> : null}
        {receivables.isError ? <LoadFailed onRetry={() => receivables.refetch()} /> : null}
        {receivables.data ? (
          <>
            <Table
              rows={receivables.data.items}
              emptyMessage={
                isFiltered
                  ? 'Tidak ada piutang yang cocok dengan penyaringan ini.'
                  : 'Belum ada piutang. Catat piutang pertama lewat tombol Catat piutang.'
              }
              columns={[
                {
                  header: 'Referensi',
                  cell: (row) => <span className="font-medium text-slate-900">{row.reference}</span>,
                },
                { header: 'Pelanggan', cell: (row) => customerName(directory, row.customerId) },
                { header: 'Jatuh tempo', cell: (row) => <DueDateCell record={row} today={today} /> },
                { header: 'Status', cell: (row) => <StatusChip status={row.status} /> },
                { header: 'Nilai', align: 'right', cell: (row) => rupiah(row.amount) },
                { header: 'Sisa', align: 'right', cell: (row) => <BalanceCell record={row} /> },
                {
                  header: 'Aksi',
                  align: 'right',
                  cell: (row) => (
                    <RowAction
                      label="Rincian"
                      isActive={row.id === selectedId}
                      onClick={() => {
                        setSelectedId((current) => (current === row.id ? null : row.id))
                        setNotice(null)
                      }}
                    />
                  ),
                },
              ]}
            />
            <Pager
              page={receivables.data.page}
              totalPages={receivables.data.totalPages}
              totalItems={receivables.data.totalItems}
              unit="piutang"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <BillingDetail
          key={selectedId}
          kind="receivables"
          id={selectedId}
          today={today}
          heading={(receivable) => `piutang ${receivable.reference}`}
          fields={(receivable) => [
            { label: 'Referensi', value: receivable.reference },
            { label: 'Pelanggan', value: customerName(directory, receivable.customerId) },
          ]}
          renderEdit={(receivable, controls) => (
            <ReceivableForm
              record={receivable}
              directory={directory}
              onSaved={controls.onSaved}
              onCancel={controls.onCancel}
            />
          )}
          onDeleted={(message) => {
            setSelectedId(null)
            setNotice(message)
          }}
        />
      ) : null}
    </div>
  )
}
