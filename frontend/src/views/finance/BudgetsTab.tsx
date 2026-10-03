import { type ChangeEvent, type FormEvent, useState } from 'react'
import { Link } from 'react-router-dom'

import { useProjects } from '../../controllers/useErp'
import { useBudgetList, useCreateBudget, useDeleteBudget, useUpdateBudget } from '../../controllers/useFinance'
import {
  BUDGET_NOTE_MAX_LENGTH,
  BUDGET_YEAR_MAX,
  BUDGET_YEAR_MIN,
  budgetFormFrom,
  emptyBudgetForm,
  isBudgetYear,
  type BudgetFormValues,
} from '../../models/accounting'
import type { Budget } from '../../models/finance'
import { projectOptions } from '../../models/lookupApi'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { Bar, Card, LoadFailed, Loading, Table } from '../components/Data'
import { Button, ErrorNote, Field, SuccessNote } from '../components/Form'
import { RowAction } from '../components/ListTools'
import { FormPanel, FormToggle, Pager, Toolbar, ToolbarInput } from '../components/RecordControls'
import { SearchSelect } from '../components/SearchSelect'
import { amountHint } from './financeTabs'
import { ConfirmDelete, Notice, RowActions } from './parts'

const PAGE_SIZE = 10
const FULL_PERCENT = 100

function budgetName(budget: Budget): string {
  return `${budget.projectName || 'proyek'} tahun ${budget.year}`
}

function BudgetForm({ budget, thisYear, onUpdated }: {
  budget: Budget | null
  thisYear: number
  onUpdated?: (message: string) => void
}) {
  const [values, setValues] = useState<BudgetFormValues>(() =>
    budget ? budgetFormFrom(budget) : emptyBudgetForm(thisYear),
  )
  const createBudget = useCreateBudget()
  const updateBudget = useUpdateBudget()
  const mutation = budget ? updateBudget : createBudget

  const update =
    (key: keyof BudgetFormValues) => (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setValues((current) => ({ ...current, [key]: event.target.value }))

  const savedMessage = (saved: Budget, verb: string) =>
    `Anggaran ${budgetName(saved)} ${verb}, nilai ${rupiah(saved.value)}.`

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (budget) {
      updateBudget.mutate(
        { id: budget.id, values },
        { onSuccess: (saved) => onUpdated?.(savedMessage(saved, 'diperbarui')) },
      )
      return
    }
    createBudget.mutate(values, {
      onSuccess: () => setValues((current) => ({ ...emptyBudgetForm(thisYear), year: current.year })),
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-4">
        {budget ? (
          <p className="text-sm text-slate-600">
            Mengubah anggaran <span className="font-medium text-slate-900">{budgetName(budget)}</span>.
          </p>
        ) : null}
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <SearchSelect
            {...projectOptions}
            id="budget-project"
            label="Proyek"
            placeholder="Cari nama atau kode proyek"
            required
            initial={budget ? { value: budget.projectId, label: budget.projectName } : null}
            value={values.projectId}
            onChange={(projectId) => setValues((current) => ({ ...current, projectId }))}
          />
          <Field
            id="budget-year"
            label="Tahun anggaran"
            type="number"
            inputMode="numeric"
            min={BUDGET_YEAR_MIN}
            max={BUDGET_YEAR_MAX}
            step={1}
            required
            hint="Satu anggaran per proyek per tahun"
            value={values.year}
            onChange={update('year')}
          />
          <Field
            id="budget-value"
            label="Nilai anggaran"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            required
            placeholder="0"
            hint={amountHint(values.value, 'Dalam Rupiah')}
            value={values.value}
            onChange={update('value')}
          />
          <Field
            id="budget-note"
            label="Catatan (opsional)"
            placeholder="Konstruksi tahap 1"
            autoComplete="off"
            maxLength={BUDGET_NOTE_MAX_LENGTH}
            value={values.note}
            onChange={update('note')}
          />
        </div>

        <Button
          type="submit"
          isPending={mutation.isPending}
          pendingLabel={budget ? 'Menyimpan perubahan' : 'Menyimpan anggaran'}
        >
          {budget ? 'Simpan perubahan' : 'Simpan anggaran'}
        </Button>

        {mutation.isError ? <ErrorNote message={errorMessage(mutation.error)} /> : null}
        {createBudget.isSuccess ? <SuccessNote message={savedMessage(createBudget.data, 'tersimpan')} /> : null}
      </form>
    </FormPanel>
  )
}

function Absorption({ budget }: { budget: Budget }) {
  return (
    <span className="flex flex-col items-end gap-0.5">
      <Bar percent={budget.absorption} />
      {budget.absorption > FULL_PERCENT ? (
        <span className="text-xs font-medium text-red-700">Lewat anggaran, terserap {budget.absorption}%</span>
      ) : null}
    </span>
  )
}

export function BudgetsTab() {
  const [thisYear] = useState(() => new Date().getFullYear())
  const [projectId, setProjectId] = useState('')
  const [yearDraft, setYearDraft] = useState('')
  const [page, setPage] = useState(1)
  const [isCreating, setIsCreating] = useState(false)
  const [editing, setEditing] = useState<Budget | null>(null)
  const [confirmingId, setConfirmingId] = useState<string | null>(null)
  const [notice, setNotice] = useState<string | null>(null)
  // Satu baris cukup untuk tahu apakah sudah ada proyek sama sekali.
  const projects = useProjects({ pageSize: 1 })
  const deleteBudget = useDeleteBudget()

  const budgets = useBudgetList({
    projectId: projectId === '' ? undefined : projectId,
    year: isBudgetYear(yearDraft) ? Number(yearDraft) : undefined,
    page,
    pageSize: PAGE_SIZE,
  })

  const emptyMessage =
    projects.isSuccess && projects.data.totalItems === 0
      ? 'Belum ada proyek. Buat proyek dulu di menu Proyek, lalu susun anggarannya di sini.'
      : projectId === '' && yearDraft === ''
        ? 'Belum ada anggaran. Tambahkan lewat tombol Tambah anggaran.'
        : 'Tidak ada anggaran untuk proyek atau tahun ini.'

  const startEdit = (budget: Budget) => {
    setIsCreating(false)
    setNotice(null)
    setEditing((current) => (current?.id === budget.id ? null : budget))
  }

  const remove = (budget: Budget) => {
    setNotice(null)
    deleteBudget.mutate(budget.id, {
      onSuccess: () => {
        setNotice(`Anggaran ${budgetName(budget)} dihapus.`)
        if (editing?.id === budget.id) {
          setEditing(null)
        }
        if (budgets.data?.items.length === 1 && page > 1) {
          setPage(page - 1)
        }
      },
      onSettled: () => setConfirmingId(null),
    })
  }

  return (
    <Card
      title="Anggaran proyek"
      description="Satu anggaran per proyek per tahun. Realisasi adalah kas keluar proyek di tahun yang sama, dihitung server."
    >
      <Toolbar>
        <SearchSelect
          {...projectOptions}
          id="budget-filter-project"
          label="Saring menurut proyek"
          compact
          allowEmpty
          emptyLabel="Semua proyek"
          className="w-52"
          value={projectId}
          onChange={(value) => {
            setProjectId(value)
            setPage(1)
          }}
        />
        <ToolbarInput
          id="budget-filter-year"
          label="Saring menurut tahun"
          type="number"
          inputMode="numeric"
          min={BUDGET_YEAR_MIN}
          max={BUDGET_YEAR_MAX}
          step={1}
          placeholder="Semua tahun"
          className="w-32"
          value={yearDraft}
          onChange={(event) => {
            setYearDraft(event.target.value)
            setPage(1)
          }}
        />
        <FormToggle
          isOpen={isCreating || editing !== null}
          openLabel="Tambah anggaran"
          onToggle={() => {
            setNotice(null)
            if (editing !== null) {
              setEditing(null)
              return
            }
            setIsCreating((open) => !open)
          }}
        />
      </Toolbar>

      {isCreating ? <BudgetForm budget={null} thisYear={thisYear} /> : null}
      {editing ? (
        <BudgetForm
          key={editing.id}
          budget={editing}
          thisYear={thisYear}
          onUpdated={(message) => {
            setEditing(null)
            setNotice(message)
          }}
        />
      ) : null}

      {deleteBudget.isError ? (
        <Notice>
          <ErrorNote message={errorMessage(deleteBudget.error)} />
        </Notice>
      ) : null}
      {notice ? (
        <Notice>
          <SuccessNote message={notice} />
        </Notice>
      ) : null}

      {budgets.isPending ? <Loading /> : null}
      {budgets.isError ? <LoadFailed onRetry={() => budgets.refetch()} /> : null}
      {budgets.data ? (
        <>
          <Table
            rows={budgets.data.items}
            emptyMessage={emptyMessage}
            columns={[
              {
                header: 'Proyek',
                cell: (row) => (
                  <Link
                    to={`/proyek/${row.projectId}?tab=keuangan`}
                    className="font-medium text-slate-900 underline-offset-4 hover:text-navy-600 hover:underline"
                  >
                    {row.projectName || '-'}
                  </Link>
                ),
              },
              { header: 'Tahun', cell: (row) => String(row.year) },
              { header: 'Anggaran', align: 'right', cell: (row) => rupiah(row.value) },
              { header: 'Realisasi', align: 'right', cell: (row) => rupiah(row.realized) },
              {
                header: 'Sisa',
                align: 'right',
                cell: (row) => (
                  <span className={row.remaining < 0 ? 'font-medium text-red-700' : undefined}>
                    {rupiah(row.remaining)}
                  </span>
                ),
              },
              { header: 'Serapan', align: 'right', cell: (row) => <Absorption budget={row} /> },
              { header: 'Catatan', cell: (row) => row.note || '-' },
              {
                header: 'Aksi',
                align: 'right',
                cell: (row) => (
                  <RowActions>
                    <RowAction label="Ubah" isActive={editing?.id === row.id} onClick={() => startEdit(row)} />
                    <ConfirmDelete
                      subject={`anggaran ${budgetName(row)}`}
                      isConfirming={confirmingId === row.id}
                      isPending={deleteBudget.isPending && deleteBudget.variables === row.id}
                      onAsk={() => setConfirmingId(row.id)}
                      onCancel={() => setConfirmingId(null)}
                      onConfirm={() => remove(row)}
                    />
                  </RowActions>
                ),
              },
            ]}
          />
          <Pager
            page={budgets.data.page}
            totalPages={budgets.data.totalPages}
            totalItems={budgets.data.totalItems}
            unit="anggaran"
            onChange={setPage}
          />
        </>
      ) : null}
    </Card>
  )
}
