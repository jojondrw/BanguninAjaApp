import { useState } from 'react'

import { useBillingList } from '../../controllers/useBilling'
import {
  BILLING_STATUSES,
  BILLING_STATUS_LABEL,
  REFERENCE_SEARCH_MAX_LENGTH,
  type BillingStatus,
} from '../../models/billing'
import { projectOptions } from '../../models/lookupApi'
import { rupiah } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { BalanceCell, DueDateCell, ReferenceCell, StatusChip } from './BillingCells'
import { BillingDetail } from './BillingDetail'
import { PayableForm } from './BillingForms'
import { allVendorOptions } from './billingLookups'

const PAGE_SIZE = 10

export function PayablesSection({ today }: { today: string }) {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<BillingStatus | ''>('')
  const [vendorId, setVendorId] = useState('')
  const [projectId, setProjectId] = useState('')
  const [page, setPage] = useState(1)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)

  const payables = useBillingList('payables', {
    search: search.trim() === '' ? undefined : search.trim(),
    status: status === '' ? undefined : status,
    vendorId: vendorId === '' ? undefined : vendorId,
    projectId: projectId === '' ? undefined : projectId,
    page,
    pageSize: PAGE_SIZE,
  })
  const isFiltered = search.trim() !== '' || status !== '' || vendorId !== '' || projectId !== ''

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  return (
    <div className="space-y-6">
      <Card
        title="Daftar utang"
        description="Kewajiban bayar ke vendor, diurutkan dari jatuh tempo paling awal"
      >
        <Toolbar>
          <ToolbarInput
            id="payable-search"
            label="Cari referensi utang"
            type="search"
            placeholder="Cari referensi"
            className="w-56"
            maxLength={REFERENCE_SEARCH_MAX_LENGTH}
            value={search}
            onChange={(event) => filterChanged(setSearch)(event.target.value)}
          />
          <FilterSelect
            id="payable-filter-status"
            label="Saring menurut status utang"
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
          <SearchSelect
            {...allVendorOptions}
            id="payable-filter-vendor"
            label="Saring menurut vendor"
            compact
            allowEmpty
            emptyLabel="Semua vendor"
            className="w-52"
            value={vendorId}
            onChange={(value) => filterChanged(setVendorId)(value)}
          />
          <SearchSelect
            {...projectOptions}
            id="payable-filter-project"
            label="Saring menurut proyek"
            compact
            allowEmpty
            emptyLabel="Semua proyek"
            className="w-52"
            value={projectId}
            onChange={(value) => filterChanged(setProjectId)(value)}
          />
          <FormToggle isOpen={isFormOpen} openLabel="Catat utang" onToggle={() => setIsFormOpen((open) => !open)} />
        </Toolbar>

        {isFormOpen ? (
          <FormPanel>
            <PayableForm />
          </FormPanel>
        ) : null}

        {notice ? (
          <div className="mb-4">
            <SuccessNote message={notice} />
          </div>
        ) : null}

        {payables.isPending ? <Loading /> : null}
        {payables.isError ? <LoadFailed onRetry={() => payables.refetch()} /> : null}
        {payables.data ? (
          <>
            <Table
              rows={payables.data.items}
              emptyMessage={
                isFiltered
                  ? 'Tidak ada utang yang cocok dengan penyaringan ini.'
                  : 'Belum ada utang. Catat utang pertama lewat tombol Catat utang.'
              }
              columns={[
                {
                  header: 'Referensi',
                  cell: (row) => <ReferenceCell reference={row.reference} sources={[['PO', row.purchaseOrderNumber]]} />,
                },
                { header: 'Vendor', cell: (row) => row.vendorName },
                { header: 'Proyek', cell: (row) => row.projectName ?? '-' },
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
              page={payables.data.page}
              totalPages={payables.data.totalPages}
              totalItems={payables.data.totalItems}
              unit="utang"
              onChange={setPage}
            />
          </>
        ) : null}
      </Card>

      {selectedId ? (
        <BillingDetail
          key={selectedId}
          kind="payables"
          id={selectedId}
          today={today}
          heading={(payable) => `utang ${payable.reference}`}
          fields={(payable) => [
            { label: 'Referensi', value: payable.reference },
            { label: 'Vendor', value: payable.vendorName },
            { label: 'Proyek', value: payable.projectName ?? 'Tanpa proyek' },
            { label: 'Pesanan pembelian', value: payable.purchaseOrderNumber ?? 'Tidak ditautkan' },
          ]}
          renderEdit={(payable, controls) => (
            <PayableForm
              record={payable}
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
