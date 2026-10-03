import { type ChangeEvent, type FormEvent, type ReactNode, useState } from 'react'

import { useBudgets } from '../../controllers/useErp'
import {
  useBudgetItemPage,
  useCreateBudgetItem,
  useDeleteBudgetItem,
  useUpdateBudgetItem,
} from '../../controllers/useProjectWorkspace'
import { summarizeBudgets, type Budget } from '../../models/finance'
import { unitOfMeasureOptions } from '../../models/lookupApi'
import {
  BUDGET_ITEM_CODE_MAX_LENGTH,
  BUDGET_ITEM_DESCRIPTION_MAX_LENGTH,
  budgetItemFormValues,
  budgetItemTotal,
  budgetShare,
  EMPTY_BUDGET_ITEM_FORM,
  isProjectClosed,
  type BudgetItem,
  type BudgetItemFormValues,
  type Project,
} from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { kpiValue, number, rupiah } from '../../shared/format'
import { Bar, Card, LoadFailed, Loading, Table, type Column } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { LookupName } from '../components/LookupName'
import { FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { FormSection, InlineConfirm } from './parts'

const PAGE_SIZE = 20
const NO_BUDGETS: Budget[] = []

type UpdateBudgetItem = ReturnType<typeof useUpdateBudgetItem>
type BudgetPanel = { kind: 'new' } | { kind: 'edit'; itemId: string } | null

export function BudgetItemsCard({ project }: { project: Project }) {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [panel, setPanel] = useState<BudgetPanel>(null)
  const items = useBudgetItemPage(project.id, {
    search: search.trim() === '' ? undefined : search.trim(),
    page,
    pageSize: PAGE_SIZE,
  })
  const budgets = useBudgets(project.id)
  const updateItem = useUpdateBudgetItem(project.id)
  const deleteItem = useDeleteBudgetItem(project.id)
  const isClosed = isProjectClosed(project)

  const data = items.data
  const grandTotal = data?.grandTotal ?? 0
  const selected = panel?.kind === 'edit' ? data?.items.find((item) => item.id === panel.itemId) : undefined

  const toggleEdit = (itemId: string) => {
    updateItem.reset()
    setPanel((current) => (current?.kind === 'edit' && current.itemId === itemId ? null : { kind: 'edit', itemId }))
  }

  // Menghapus satu-satunya baris di halaman terakhir mundur satu halaman,
  // supaya tabel tidak kosong padahal masih ada item.
  const remove = (itemId: string) =>
    deleteItem.mutate(itemId, {
      onSuccess: () => {
        if (data && data.items.length === 1 && page > 1) {
          setPage(page - 1)
        }
      },
    })

  const columns: Column<BudgetItem>[] = [
    { header: 'Kode', cell: (row) => <span className="font-medium text-slate-900">{row.code}</span> },
    { header: 'Uraian', cell: (row) => row.description },
    {
      header: 'Volume',
      align: 'right',
      cell: (row) => (
        <>
          {number(row.volume)} <LookupName source={unitOfMeasureOptions} value={row.unitOfMeasureId} fallback="" />
        </>
      ),
    },
    { header: 'Harga satuan', align: 'right', cell: (row) => rupiah(row.unitPrice) },
    { header: 'Jumlah', align: 'right', cell: (row) => rupiah(row.total) },
    { header: 'Porsi RAB', align: 'right', cell: (row) => <Bar percent={budgetShare(row.total, grandTotal)} /> },
  ]

  const actionColumn: Column<BudgetItem> = {
    header: 'Aksi',
    align: 'right',
    cell: (row) => (
      <span className="inline-flex items-center justify-end gap-1">
        <RowAction
          label="Ubah"
          isActive={panel?.kind === 'edit' && panel.itemId === row.id}
          onClick={() => toggleEdit(row.id)}
        />
        <InlineConfirm
          label="Hapus"
          confirmLabel="Ya, hapus item"
          pendingLabel="Menghapus"
          isPending={deleteItem.isPending && deleteItem.variables === row.id}
          onConfirm={() => remove(row.id)}
        />
      </span>
    ),
  }

  return (
    <Card
      title="Rencana anggaran biaya (RAB)"
      description="Jumlah tiap item dihitung server dari volume kali harga satuan. Realisasi dicatat per proyek lewat transaksi kas, belum per item."
    >
      <BudgetTotals
        project={project}
        grandTotal={grandTotal}
        itemsState={{ isPending: items.isPending, isError: items.isError }}
        budgets={budgets.data?.items ?? NO_BUDGETS}
        budgetsState={{ isPending: budgets.isPending, isError: budgets.isError }}
      />

      <Toolbar>
        <ToolbarInput
          id="rab-search"
          label="Cari item RAB"
          type="search"
          placeholder="Cari kode atau uraian"
          value={search}
          onChange={(event) => {
            setSearch(event.target.value)
            setPage(1)
          }}
        />
        {isClosed ? null : (
          <FormToggle
            isOpen={panel?.kind === 'new'}
            openLabel="Tambah item RAB"
            onToggle={() => setPanel((current) => (current?.kind === 'new' ? null : { kind: 'new' }))}
          />
        )}
      </Toolbar>

      {isClosed ? (
        <p className="mb-4 text-xs text-slate-500">
          Proyek sudah selesai atau dibatalkan, jadi item RAB tidak bisa ditambah, diubah, atau dihapus.
        </p>
      ) : null}

      {panel?.kind === 'new' && !isClosed ? (
        <FormPanel>
          <NewBudgetItemForm projectId={project.id} />
        </FormPanel>
      ) : null}

      {deleteItem.isError ? (
        <div className="mb-4">
          <ErrorNote message={errorMessage(deleteItem.error)} />
        </div>
      ) : null}
      {updateItem.isSuccess && panel === null ? (
        <div className="mb-4">
          <SuccessNote
            message={`Item ${updateItem.data.code} disimpan, jumlahnya sekarang ${rupiah(updateItem.data.total)}.`}
          />
        </div>
      ) : null}

      {items.isPending ? <Loading /> : null}
      {items.isError ? <LoadFailed onRetry={() => items.refetch()} /> : null}
      {data ? (
        <>
          <Table
            rows={data.items}
            emptyMessage={
              search.trim() === '' ? 'Belum ada item RAB untuk proyek ini.' : 'Tidak ada item RAB yang cocok dengan pencarian.'
            }
            columns={isClosed ? columns : [...columns, actionColumn]}
          />
          <Pager
            page={data.page}
            totalPages={data.totalPages}
            totalItems={data.totalItems}
            unit="item RAB"
            onChange={setPage}
          />
        </>
      ) : null}

      {selected && !isClosed ? (
        <EditBudgetItemForm
          key={selected.id}
          item={selected}
          update={updateItem}
          onDone={() => setPanel(null)}
        />
      ) : null}
    </Card>
  )
}

interface LoadState {
  isPending: boolean
  isError: boolean
}

function realizedNote(years: number, realizedShare: number, grandTotal: number): string {
  if (years === 0) {
    return 'belum ada anggaran tahunan di Keuangan'
  }
  return grandTotal > 0 ? `${realizedShare}% dari total RAB` : `dari ${years} tahun anggaran`
}

// Rencana (RAB) dibandingkan dengan realisasi di tingkat proyek. Realisasi
// diambil dari anggaran tahunan di Keuangan, yaitu kas keluar proyek.
function BudgetTotals({ project, grandTotal, itemsState, budgets, budgetsState }: {
  project: Project
  grandTotal: number
  itemsState: LoadState
  budgets: Budget[]
  budgetsState: LoadState
}) {
  const summary = summarizeBudgets(budgets)
  const remaining = grandTotal - summary.realized

  return (
    <dl className="mb-5 grid grid-cols-2 gap-x-6 gap-y-4 rounded-xl bg-slate-50 px-4 py-3.5 lg:grid-cols-4">
      <Total label="Total RAB" note="semua item, bukan hanya halaman ini">
        {kpiValue(itemsState.isPending, itemsState.isError, rupiah(grandTotal))}
      </Total>
      <Total
        label="Nilai kontrak"
        note={
          project.contractValue > 0
            ? `RAB ${budgetShare(grandTotal, project.contractValue)}% dari kontrak`
            : 'kontrak belum diisi'
        }
      >
        {rupiah(project.contractValue)}
      </Total>
      <Total
        label="Realisasi kas keluar"
        note={realizedNote(summary.years, budgetShare(summary.realized, grandTotal), grandTotal)}
      >
        {kpiValue(budgetsState.isPending, budgetsState.isError, rupiah(summary.realized))}
      </Total>
      <Total label="Sisa RAB" note={remaining < 0 ? 'realisasi melebihi RAB' : 'RAB dikurangi realisasi'}>
        {kpiValue(
          budgetsState.isPending || itemsState.isPending,
          budgetsState.isError || itemsState.isError,
          rupiah(remaining),
        )}
      </Total>
    </dl>
  )
}

function Total({ label, note, children }: { label: string; note: string; children: ReactNode }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs font-medium text-slate-500">{label}</dt>
      <dd className="mt-1 text-[15px] font-semibold tracking-[-0.01em] text-slate-900 tabular-nums">{children}</dd>
      <dd className="mt-0.5 text-xs text-slate-500">{note}</dd>
    </div>
  )
}

function totalHint(values: BudgetItemFormValues): string {
  const total = budgetItemTotal(values)
  return total === null ? 'Jumlah = volume x harga satuan' : `Jumlah ${rupiah(total)}`
}

function BudgetItemFields({ idPrefix, values, onChange, onUnitChange, autoFocus = false }: {
  idPrefix: string
  values: BudgetItemFormValues
  onChange: (key: keyof BudgetItemFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onUnitChange: (unitOfMeasureId: string) => void
  autoFocus?: boolean
}) {
  return (
    <div className="grid gap-4 md:grid-cols-3">
      <Field
        id={`${idPrefix}-code`}
        label="Kode"
        placeholder="PEK-02"
        autoComplete="off"
        autoFocus={autoFocus}
        maxLength={BUDGET_ITEM_CODE_MAX_LENGTH}
        hint="Unik di dalam proyek ini"
        required
        value={values.code}
        onChange={onChange('code')}
      />
      <div className="md:col-span-2">
        <Field
          id={`${idPrefix}-description`}
          label="Uraian pekerjaan"
          placeholder="Pengecoran pelat lantai 2"
          autoComplete="off"
          maxLength={BUDGET_ITEM_DESCRIPTION_MAX_LENGTH}
          required
          value={values.description}
          onChange={onChange('description')}
        />
      </div>
      <Field
        id={`${idPrefix}-volume`}
        label="Volume"
        type="number"
        inputMode="decimal"
        min={0}
        step={0.01}
        required
        placeholder="0"
        hint="Dibulatkan dua angka di belakang koma"
        value={values.volume}
        onChange={onChange('volume')}
      />
      <SearchSelect
        {...unitOfMeasureOptions}
        id={`${idPrefix}-unit`}
        label="Satuan"
        placeholder="Cari kode atau nama satuan"
        required
        value={values.unitOfMeasureId}
        onChange={onUnitChange}
      />
      <Field
        id={`${idPrefix}-unit-price`}
        label="Harga satuan"
        type="number"
        inputMode="numeric"
        min={0}
        step={1}
        required
        placeholder="0"
        hint={totalHint(values)}
        value={values.unitPrice}
        onChange={onChange('unitPrice')}
      />
    </div>
  )
}

function useBudgetItemValues(initial: BudgetItemFormValues) {
  const [values, setValues] = useState<BudgetItemFormValues>(initial)

  const change =
    (key: keyof BudgetItemFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const changeUnit = (unitOfMeasureId: string) => setValues((current) => ({ ...current, unitOfMeasureId }))

  return { values, setValues, change, changeUnit }
}

// Formulir tetap terbuka setelah menyimpan supaya beberapa item bisa dicatat
// berturut-turut. Satuan terakhir dipertahankan karena biasanya sama.
function NewBudgetItemForm({ projectId }: { projectId: string }) {
  const { values, setValues, change, changeUnit } = useBudgetItemValues(EMPTY_BUDGET_ITEM_FORM)
  const createItem = useCreateBudgetItem(projectId)

  const submit = (event: FormEvent) => {
    event.preventDefault()
    createItem.mutate(values, {
      onSuccess: () => setValues((current) => ({ ...EMPTY_BUDGET_ITEM_FORM, unitOfMeasureId: current.unitOfMeasureId })),
    })
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <BudgetItemFields idPrefix="rab-new" values={values} onChange={change} onUnitChange={changeUnit} />

      <Button type="submit" isPending={createItem.isPending} pendingLabel="Menyimpan item">
        Simpan item RAB
      </Button>

      {createItem.isError ? <ErrorNote message={errorMessage(createItem.error)} /> : null}
      {createItem.isSuccess ? (
        <SuccessNote
          message={`Item ${createItem.data.code} ditambahkan dengan jumlah ${rupiah(createItem.data.total)}.`}
        />
      ) : null}
    </form>
  )
}

function EditBudgetItemForm({ item, update, onDone }: {
  item: BudgetItem
  update: UpdateBudgetItem
  onDone: () => void
}) {
  const { values, change, changeUnit } = useBudgetItemValues(budgetItemFormValues(item))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    update.mutate({ itemId: item.id, values }, { onSuccess: onDone })
  }

  return (
    <FormSection title={`Ubah item ${item.code}`}>
      <form onSubmit={submit} className="space-y-4">
        <BudgetItemFields idPrefix="rab-edit" values={values} onChange={change} onUnitChange={changeUnit} autoFocus />

        <div className="flex flex-wrap gap-2">
          <Button type="submit" isPending={update.isPending} pendingLabel="Menyimpan item">
            Simpan perubahan item
          </Button>
          <Button variant="subtle" onClick={onDone}>
            Batal
          </Button>
        </div>

        {update.isError ? <ErrorNote message={errorMessage(update.error)} /> : null}
      </form>
    </FormSection>
  )
}
