import { useState } from 'react'

import { useCustomers, useDeleteCustomer } from '../../controllers/useSales'
import { CUSTOMER_NAME_MAX_LENGTH, type Customer } from '../../models/sales'
import { shortDate } from '../../shared/format'
import { Card, LoadFailed, Loading, Table } from '../components/Data'
import { ErrorNote, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { CustomerForm } from './CustomerForm'
import { ActionGroup, ConfirmAction } from './RowActions'
import { PAGE_SIZE, pageAfterRemoval, refusalText } from './salesShared'

export function CustomersSection() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Customer | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const customers = useCustomers({
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })
  const deleteCustomer = useDeleteCustomer()
  const isFormOpen = isCreating || editing !== null

  const closeForm = () => {
    setIsCreating(false)
    setEditing(null)
  }

  const toggleForm = () => {
    setNotice(null)
    if (isFormOpen) {
      closeForm()
      return
    }
    setIsCreating(true)
  }

  const startEdit = (customer: Customer) => {
    setNotice(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === customer.id ? null : customer))
  }

  const askDelete = (customer: Customer) => {
    setNotice(null)
    deleteCustomer.reset()
    setAskingId(customer.id)
  }

  const confirmDelete = (customer: Customer) => {
    deleteCustomer.mutate(customer, {
      onSuccess: () => {
        setNotice(`Pelanggan ${customer.name} berhasil dihapus.`)
        setPage((current) => pageAfterRemoval(current, customers.data?.items.length ?? 0))
        if (editing?.id === customer.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  return (
    <Card title="Daftar pelanggan" description="Pelanggan yang sudah punya kontrak tidak bisa dihapus.">
      <Toolbar>
        <ToolbarInput
          id="customer-search"
          label="Cari pelanggan"
          type="search"
          placeholder="Cari nama, email, atau NIK"
          maxLength={CUSTOMER_NAME_MAX_LENGTH}
          className="w-60"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        <FormToggle isOpen={isFormOpen} openLabel="Tambah pelanggan" onToggle={toggleForm} />
      </Toolbar>

      {isFormOpen ? (
        <CustomerForm
          key={editing?.id ?? 'new'}
          customer={editing}
          title={editing ? `Ubah pelanggan ${editing.name}` : 'Pelanggan baru'}
          onCancel={editing ? closeForm : undefined}
          onSaved={
            editing
              ? (saved) => {
                  closeForm()
                  setNotice(`Data pelanggan ${saved.name} berhasil diperbarui.`)
                }
              : undefined
          }
        />
      ) : null}

      {notice ? (
        <div className="mb-4">
          <SuccessNote message={notice} />
        </div>
      ) : null}
      {deleteCustomer.isError ? (
        <div className="mb-4">
          <ErrorNote
            message={refusalText(`Pelanggan ${deleteCustomer.variables.name} tidak bisa dihapus.`, deleteCustomer.error)}
          />
        </div>
      ) : null}

      {customers.isPending ? <Loading /> : null}
      {customers.isError ? <LoadFailed onRetry={() => customers.refetch()} /> : null}
      {customers.data ? (
        <>
          <Table
            rows={customers.data.items}
            emptyMessage={
              search.trim() === ''
                ? 'Belum ada pelanggan. Tambahkan lewat tombol Tambah pelanggan.'
                : 'Tidak ada pelanggan yang cocok dengan pencarian itu.'
            }
            columns={[
              { header: 'Nama', cell: (row) => <span className="font-medium text-slate-900">{row.name}</span> },
              { header: 'No. identitas', cell: (row) => row.identityNumber },
              { header: 'Telepon', cell: (row) => row.contact || '-' },
              { header: 'Email', cell: (row) => row.email || '-' },
              { header: 'Terdaftar', cell: (row) => shortDate(row.createdAt) },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <ActionGroup>
                    {askingId !== row.id ? (
                      <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                    ) : null}
                    <ConfirmAction
                      label="Hapus"
                      question={`Hapus ${row.name}?`}
                      confirmLabel="Ya, hapus"
                      pendingLabel="Menghapus"
                      isAsking={askingId === row.id}
                      isPending={deleteCustomer.isPending && deleteCustomer.variables.id === row.id}
                      onAsk={() => askDelete(row)}
                      onCancel={() => setAskingId(null)}
                      onConfirm={() => confirmDelete(row)}
                    />
                  </ActionGroup>
                ),
              },
            ]}
          />
          <Pager
            page={customers.data.page}
            totalPages={customers.data.totalPages}
            totalItems={customers.data.totalItems}
            unit="pelanggan"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
