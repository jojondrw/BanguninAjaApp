import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useScopedCan } from '../../controllers/useAccess'
import { useAccounts, useCreateAccount, useDeleteAccount, useUpdateAccount } from '../../controllers/useMaster'
import type { OptionSource } from '../../models/lookup'
import { accountOptions } from '../../models/lookupApi'
import {
  ACCOUNT_CODE_MAX_LENGTH,
  ACCOUNT_NAME_MAX_LENGTH,
  ACCOUNT_SEARCH_MAX_LENGTH,
  ACCOUNT_TYPE_LABEL,
  ACCOUNT_TYPES,
  EMPTY_ACCOUNT_FORM,
  accountFormValues,
  accountParentText,
  accountTypeLocked,
  isCashAccount,
  withAccountType,
  type Account,
  type AccountFormValues,
  type AccountType,
  type PostableFilter,
} from '../../models/master'
import { errorMessage } from '../../shared/errorMessage'
import { type Column, Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { Chip, FilterSelect, FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { pageAfterRemoval, refusalText } from '../hr/hrShared'
import { ActionGroup, ConfirmAction, RowNotice } from '../hr/RowActions'
import { PAGE_SIZE } from './masterShared'

// Akun induk harus sejenis, jadi pilihan induk dibatasi ke jenis akun di formulir.
const PARENT_SOURCES = Object.fromEntries(ACCOUNT_TYPES.map((type) => [type, accountOptions(type)])) as Record<
  AccountType,
  OptionSource
>

type ChangeAccount = (key: 'code' | 'name') => (event: ChangeEvent<HTMLInputElement>) => void

function useAccountValues(initial: AccountFormValues) {
  const [values, setValues] = useState<AccountFormValues>(initial)

  const change: ChangeAccount = (key) => (event) => setValues((current) => ({ ...current, [key]: event.target.value }))
  const pickType = (type: AccountType) => setValues((current) => withAccountType(current, type))
  const pickParent = (parentId: string) => setValues((current) => ({ ...current, parentId }))

  return { values, setValues, change, pickType, pickParent }
}

function AccountFields({ idPrefix, account, values, onChange, onType, onParent }: {
  idPrefix: string
  account: Account | null
  values: AccountFormValues
  onChange: ChangeAccount
  onType: (type: AccountType) => void
  onParent: (parentId: string) => void
}) {
  const isCash = account !== null && isCashAccount(account)
  const isTypeLocked = accountTypeLocked(account)

  return (
    <div className="grid gap-4 md:grid-cols-2">
      <Field
        id={`${idPrefix}-code`}
        label="Kode akun"
        placeholder="1130"
        autoComplete="off"
        maxLength={ACCOUNT_CODE_MAX_LENGTH}
        hint={isCash ? 'Dipakai jurnal kas otomatis, jadi kodenya tetap' : `Unik, maksimal ${ACCOUNT_CODE_MAX_LENGTH} karakter`}
        required
        disabled={isCash}
        value={values.code}
        onChange={onChange('code')}
      />
      <Field
        id={`${idPrefix}-name`}
        label="Nama akun"
        placeholder="Kas Kecil Proyek"
        autoComplete="off"
        maxLength={ACCOUNT_NAME_MAX_LENGTH}
        required
        value={values.name}
        onChange={onChange('name')}
      />
      <SelectField
        id={`${idPrefix}-type`}
        label="Jenis akun"
        hint={isTypeLocked ? 'Tidak bisa diubah karena akun ini punya akun turunan' : 'Akun induk dan turunannya harus sejenis'}
        disabled={isTypeLocked}
        value={values.type}
        onChange={(event) => onType(event.target.value as AccountType)}
      >
        {ACCOUNT_TYPES.map((type) => (
          <option key={type} value={type}>
            {ACCOUNT_TYPE_LABEL[type]}
          </option>
        ))}
      </SelectField>
      <SearchSelect
        key={values.type}
        {...PARENT_SOURCES[values.type]}
        id={`${idPrefix}-parent`}
        label="Akun induk (opsional)"
        placeholder="Cari kode atau nama akun"
        hint="Akun yang menjadi induk tidak bisa lagi dipakai di transaksi kas dan jurnal"
        allowEmpty
        emptyLabel="Tanpa induk, akun utama"
        exclude={account ? [account.id] : undefined}
        value={values.parentId}
        onChange={onParent}
      />
    </div>
  )
}

function NewAccountForm() {
  const { values, setValues, change, pickType, pickParent } = useAccountValues(EMPTY_ACCOUNT_FORM)
  const createAccount = useCreateAccount()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createAccount.mutate(values, { onSuccess: () => setValues({ ...EMPTY_ACCOUNT_FORM, type: values.type }) })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <AccountFields
          idPrefix="account-new"
          account={null}
          values={values}
          onChange={change}
          onType={pickType}
          onParent={pickParent}
        />

        <Button type="submit" isPending={createAccount.isPending} pendingLabel="Menyimpan akun">
          Simpan akun
        </Button>

        {createAccount.isError ? <ErrorNote message={errorMessage(createAccount.error)} /> : null}
        {createAccount.isSuccess ? (
          <SuccessNote message={`Akun ${createAccount.data.code} ${createAccount.data.name} berhasil ditambahkan.`} />
        ) : null}
      </form>
    </FormPanel>
  )
}

function EditAccountForm({ account, onSaved, onCancel }: {
  account: Account
  onSaved: (saved: Account) => void
  onCancel: () => void
}) {
  const { values, change, pickType, pickParent } = useAccountValues(accountFormValues(account))
  const updateAccount = useUpdateAccount()

  const submit = (event: FormEvent) => {
    event.preventDefault()
    updateAccount.mutate({ account, values }, { onSuccess: onSaved })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        <h3 className="text-[15px] font-semibold text-slate-900">
          Ubah akun {account.code} {account.name}
        </h3>
        <AccountFields
          idPrefix="account-edit"
          account={account}
          values={values}
          onChange={change}
          onType={pickType}
          onParent={pickParent}
        />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={updateAccount.isPending} pendingLabel="Menyimpan perubahan">
            Simpan perubahan akun
          </Button>
          <Button variant="subtle" onClick={onCancel}>
            Batal
          </Button>
        </div>

        {updateAccount.isError ? <ErrorNote message={errorMessage(updateAccount.error)} /> : null}
      </form>
    </FormPanel>
  )
}

function PostableChip({ account }: { account: Account }) {
  return account.postable ? (
    <Chip tone="bg-green-100 text-green-800">Ya</Chip>
  ) : (
    <Chip tone="bg-slate-100 text-slate-700">Tidak, akun induk</Chip>
  )
}

export function AccountsSection() {
  const canWrite = useScopedCan('write')
  const [search, setSearch] = useState('')
  const [type, setType] = useState<AccountType | ''>('')
  const [postable, setPostable] = useState<PostableFilter>('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Account | null>(null)
  const [askingId, setAskingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  const deleteAccount = useDeleteAccount()
  const isFormOpen = isCreating || editing !== null

  const accounts = useAccounts({
    search: search.trim() === '' ? undefined : search.trim(),
    type: type === '' ? undefined : type,
    postable: postable === '' ? undefined : postable,
    page,
    pageSize: PAGE_SIZE,
  })

  const filterChanged = <T,>(apply: (value: T) => void) => (value: T) => {
    apply(value)
    setPage(1)
  }

  const clearNotes = () => {
    setNotice(null)
    deleteAccount.reset()
  }

  const closeForm = () => {
    setIsCreating(false)
    setEditing(null)
  }

  const toggleForm = () => {
    clearNotes()
    if (isFormOpen) {
      closeForm()
      return
    }
    setIsCreating(true)
  }

  const startEdit = (account: Account) => {
    clearNotes()
    setAskingId(null)
    setIsCreating(false)
    setEditing((current) => (current?.id === account.id ? null : account))
  }

  const remove = (account: Account) => {
    deleteAccount.mutate(account, {
      onSuccess: () => {
        setNotice(`Akun ${account.code} ${account.name} dihapus.`)
        setPage((current) => pageAfterRemoval(current, accounts.data?.items.length ?? 0))
        if (editing?.id === account.id) {
          closeForm()
        }
      },
      onSettled: () => setAskingId(null),
    })
  }

  const columns: Column<Account>[] = [
    { header: 'Kode', cell: (row) => <span className="font-medium text-slate-900 tabular-nums">{row.code}</span> },
    { header: 'Nama', cell: (row) => row.name },
    { header: 'Jenis', cell: (row) => ACCOUNT_TYPE_LABEL[row.type] },
    {
      header: 'Induk',
      cell: (row) => <span className={row.parentId === null ? 'text-slate-500' : ''}>{accountParentText(row)}</span>,
    },
    { header: 'Bisa diposting', cell: (row) => <PostableChip account={row} /> },
  ]

  if (canWrite) {
    columns.push({
      header: 'Aksi',
      align: 'right',
      cell: (row) => (
        <ActionGroup>
          {askingId === row.id ? null : (
            <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
          )}
          {isCashAccount(row) ? null : (
            <ConfirmAction
              label="Hapus"
              srLabel={`${row.code} ${row.name}`}
              question={`Hapus akun ${row.code}?`}
              confirmLabel="Ya, hapus"
              pendingLabel="Menghapus"
              isAsking={askingId === row.id}
              isPending={deleteAccount.isPending && deleteAccount.variables.id === row.id}
              onAsk={() => {
                clearNotes()
                setAskingId(row.id)
              }}
              onCancel={() => setAskingId(null)}
              onConfirm={() => remove(row)}
            />
          )}
        </ActionGroup>
      ),
    })
  }

  return (
    <Card
      title="Bagan akun"
      description="Akun tanpa akun turunan bisa dipakai di transaksi kas dan jurnal. Akun yang sudah dipakai atau masih punya turunan tidak bisa dihapus."
    >
      <Toolbar>
        <ToolbarInput
          id="account-search"
          label="Cari kode atau nama akun"
          type="search"
          placeholder="Cari kode atau nama akun"
          className="w-56"
          maxLength={ACCOUNT_SEARCH_MAX_LENGTH}
          value={search}
          onChange={(event) => filterChanged(setSearch)(event.target.value)}
        />
        <FilterSelect
          id="account-filter-type"
          label="Saring menurut jenis akun"
          value={type}
          onChange={(event) => filterChanged(setType)(event.target.value as AccountType | '')}
        >
          <option value="">Semua jenis</option>
          {ACCOUNT_TYPES.map((accountType) => (
            <option key={accountType} value={accountType}>
              {ACCOUNT_TYPE_LABEL[accountType]}
            </option>
          ))}
        </FilterSelect>
        <FilterSelect
          id="account-filter-postable"
          label="Saring menurut bisa diposting"
          value={postable}
          onChange={(event) => filterChanged(setPostable)(event.target.value as PostableFilter)}
        >
          <option value="">Semua akun</option>
          <option value="true">Bisa diposting</option>
          <option value="false">Akun induk</option>
        </FilterSelect>
        <FormToggle isOpen={isFormOpen} openLabel="Tambah akun" onToggle={toggleForm} />
      </Toolbar>

      {isCreating ? <NewAccountForm /> : null}
      {editing ? (
        <EditAccountForm
          key={editing.id}
          account={editing}
          onCancel={closeForm}
          onSaved={(saved) => {
            closeForm()
            setNotice(`Akun ${saved.code} ${saved.name} berhasil diperbarui.`)
          }}
        />
      ) : null}

      <RowNotice
        success={notice}
        error={
          deleteAccount.isError
            ? refusalText(`Akun ${deleteAccount.variables.code} tidak bisa dihapus.`, deleteAccount.error)
            : null
        }
      />

      {accounts.isPending ? <Loading /> : null}
      {accounts.isError ? <LoadFailed onRetry={() => accounts.refetch()} /> : null}
      {accounts.data ? (
        <>
          <Table rows={accounts.data.items} emptyMessage="Tidak ada akun yang cocok dengan penyaringan ini." columns={columns} />
          <Pager
            page={accounts.data.page}
            totalPages={accounts.data.totalPages}
            totalItems={accounts.data.totalItems}
            unit="akun"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
