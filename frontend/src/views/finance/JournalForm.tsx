import { type ChangeEvent, type FormEvent, useState } from 'react'

import { useCreateJournalEntry } from '../../controllers/useFinance'
import {
  JOURNAL_MIN_LINES,
  JOURNAL_NOTE_MAX_LENGTH,
  JOURNAL_NUMBER_MAX_LENGTH,
  JOURNAL_SIDE_LABEL,
  JOURNAL_SOURCE_MAX_LENGTH,
  emptyJournalForm,
  emptyJournalLine,
  journalBalance,
  postableAccountGroups,
  type JournalBalance,
  type JournalEntryDetail,
  type JournalFormValues,
  type JournalLineValues,
  type JournalSide,
} from '../../models/accounting'
import type { AccountGroup, Account } from '../../models/master'
import { errorMessage } from '../../shared/errorMessage'
import { rupiah } from '../../shared/format'
import { todayDate } from '../../shared/localDate'
import { Button, ErrorNote, Field, SelectField, SuccessNote } from '../components/Form'
import { FormPanel } from '../components/RecordControls'
import { amountHint } from './financeTabs'
import { AccountOptions, SubmitButton } from './parts'

const BALANCE_STATUS_ID = 'journal-balance-status'
const SIDES: JournalSide[] = ['debit', 'credit']

function JournalLineFields({ line, index, groups, canRemove, onChange, onRemove }: {
  line: JournalLineValues
  index: number
  groups: AccountGroup[]
  canRemove: boolean
  onChange: (line: JournalLineValues) => void
  onRemove: () => void
}) {
  const id = line.key

  const update = (key: 'accountId' | 'side' | 'amount') =>
    (event: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => onChange({ ...line, [key]: event.target.value })

  return (
    <fieldset className="rounded-xl bg-white p-4 shadow-hairline">
      <legend className="float-left mb-3 w-full text-xs font-medium text-slate-500">Baris {index + 1}</legend>
      <div className="clear-both grid gap-4 md:grid-cols-[2fr_1fr_1.2fr]">
        <SelectField id={`${id}-account`} label="Akun" required value={line.accountId} onChange={update('accountId')}>
          <AccountOptions groups={groups} placeholder="Pilih akun" />
        </SelectField>
        <SelectField id={`${id}-side`} label="Posisi" value={line.side} onChange={update('side')}>
          {SIDES.map((side) => (
            <option key={side} value={side}>
              {JOURNAL_SIDE_LABEL[side]}
            </option>
          ))}
        </SelectField>
        <Field
          id={`${id}-amount`}
          label={`Jumlah ${JOURNAL_SIDE_LABEL[line.side].toLowerCase()}`}
          type="number"
          inputMode="numeric"
          min={1}
          step={1}
          required
          placeholder="0"
          hint={amountHint(line.amount, 'Dalam Rupiah, lebih dari nol')}
          value={line.amount}
          onChange={update('amount')}
        />
      </div>
      {canRemove ? (
        <div className="mt-3 flex justify-end">
          <button
            type="button"
            onClick={onRemove}
            className="rounded-md px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
          >
            Hapus baris {index + 1}
          </button>
        </div>
      ) : null}
    </fieldset>
  )
}

function balanceMessage(balance: JournalBalance): string {
  if (balance.debit === 0 && balance.credit === 0) {
    return 'Isi jumlah setiap baris. Jurnal baru bisa disimpan setelah total debit sama dengan total kredit.'
  }
  if (balance.isBalanced) {
    return `Seimbang: total debit dan total kredit sama-sama ${rupiah(balance.debit)}.`
  }
  if (balance.difference > 0) {
    return `Belum seimbang: kredit kurang ${rupiah(balance.difference)} dari debit. Tambah atau ubah baris kredit.`
  }
  return `Belum seimbang: debit kurang ${rupiah(-balance.difference)} dari kredit. Tambah atau ubah baris debit.`
}

function BalancePanel({ balance }: { balance: JournalBalance }) {
  const cells = [
    { label: 'Total debit', value: rupiah(balance.debit), tone: 'text-slate-900' },
    { label: 'Total kredit', value: rupiah(balance.credit), tone: 'text-slate-900' },
    {
      label: 'Selisih',
      value: rupiah(Math.abs(balance.difference)),
      tone: balance.difference === 0 ? 'text-slate-900' : 'text-red-700',
    },
  ]

  return (
    <div className="rounded-xl bg-white p-4 shadow-hairline">
      <dl className="grid grid-cols-3 gap-4">
        {cells.map((cell) => (
          <div key={cell.label} className="min-w-0">
            <dt className="text-xs font-medium text-slate-500">{cell.label}</dt>
            <dd className={`mt-1 text-sm font-semibold tabular-nums ${cell.tone}`}>{cell.value}</dd>
          </div>
        ))}
      </dl>
      <p
        id={BALANCE_STATUS_ID}
        role="status"
        className={`mt-3 text-[13px] ${balance.isBalanced ? 'text-green-700' : 'text-slate-600'}`}
      >
        {balanceMessage(balance)}
      </p>
    </div>
  )
}

export function JournalForm({ accounts, isLoadingAccounts, onRecorded }: {
  accounts: Account[]
  isLoadingAccounts: boolean
  onRecorded: (entry: JournalEntryDetail) => void
}) {
  const [values, setValues] = useState<JournalFormValues>(() => emptyJournalForm(todayDate()))
  const createEntry = useCreateJournalEntry()
  const groups = postableAccountGroups(accounts)
  const balance = journalBalance(values.lines)

  const update = (key: 'number' | 'date' | 'note' | 'source') => (event: ChangeEvent<HTMLInputElement>) =>
    setValues((current) => ({ ...current, [key]: event.target.value }))

  const changeLine = (key: string) => (line: JournalLineValues) =>
    setValues((current) => ({ ...current, lines: current.lines.map((item) => (item.key === key ? line : item)) }))

  const removeLine = (key: string) => () =>
    setValues((current) => ({ ...current, lines: current.lines.filter((item) => item.key !== key) }))

  // Baris baru langsung berada di sisi yang masih kurang.
  const addLine = () =>
    setValues((current) => ({
      ...current,
      lines: [...current.lines, emptyJournalLine(balance.difference > 0 ? 'credit' : 'debit')],
    }))

  const submit = (event: FormEvent) => {
    event.preventDefault()
    if (!balance.isBalanced) {
      return
    }
    createEntry.mutate(values, {
      onSuccess: (entry) => {
        setValues(emptyJournalForm(values.date))
        onRecorded(entry)
      },
    })
  }

  return (
    <FormPanel>
      <form onSubmit={submit} className="space-y-5">
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <Field
            id="journal-number"
            label="Nomor jurnal"
            placeholder="JU-2026-001"
            autoComplete="off"
            maxLength={JOURNAL_NUMBER_MAX_LENGTH}
            hint={`Unik, maksimal ${JOURNAL_NUMBER_MAX_LENGTH} karakter`}
            required
            value={values.number}
            onChange={update('number')}
          />
          <Field id="journal-date" label="Tanggal" type="date" required value={values.date} onChange={update('date')} />
          <Field
            id="journal-source"
            label="Sumber (opsional)"
            placeholder="Manual"
            autoComplete="off"
            maxLength={JOURNAL_SOURCE_MAX_LENGTH}
            hint="Misalnya Manual, Penyesuaian, atau nomor bukti"
            value={values.source}
            onChange={update('source')}
          />
          <Field
            id="journal-note"
            label="Keterangan (opsional)"
            placeholder="Setoran modal awal"
            autoComplete="off"
            maxLength={JOURNAL_NOTE_MAX_LENGTH}
            value={values.note}
            onChange={update('note')}
          />
        </div>

        <div className="space-y-3">
          {isLoadingAccounts ? <p className="text-sm text-slate-600">Memuat daftar akun...</p> : null}
          {values.lines.map((line, index) => (
            <JournalLineFields
              key={line.key}
              line={line}
              index={index}
              groups={groups}
              canRemove={values.lines.length > JOURNAL_MIN_LINES}
              onChange={changeLine(line.key)}
              onRemove={removeLine(line.key)}
            />
          ))}
          <Button variant="subtle" onClick={addLine}>
            Tambah baris
          </Button>
        </div>

        <BalancePanel balance={balance} />

        <SubmitButton
          isPending={createEntry.isPending}
          pendingLabel="Menyimpan jurnal"
          isBlocked={!balance.isBalanced}
          describedBy={BALANCE_STATUS_ID}
        >
          Simpan jurnal
        </SubmitButton>

        {createEntry.isError ? <ErrorNote message={errorMessage(createEntry.error)} /> : null}
        {createEntry.isSuccess ? (
          <SuccessNote
            message={`Jurnal ${createEntry.data.number} tercatat dengan total ${rupiah(createEntry.data.total)}. Rinciannya tampil di bawah daftar.`}
          />
        ) : null}
      </form>
    </FormPanel>
  )
}
