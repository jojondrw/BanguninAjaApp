import { useState } from 'react'

import type { ReceiptStockLine } from '../../models/procurement'
import { number } from '../../shared/format'
import { ErrorNote, Field } from '../components/Form'
import { ConfirmAction } from './parts'

// Satu baris penerimaan yang tidak dicatat otomatis karena satuan pesanannya
// beda dengan satuan stok material. Jumlahnya diisi sendiri dalam satuan stok,
// karena tidak ada data konversi satuan yang bisa dipakai untuk menebak.
export function ManualStockLine({
  receiptId,
  line,
  orderUnit,
  stockUnit,
  warehouseName,
  isPending,
  error,
  onRecord,
}: {
  receiptId: string
  line: ReceiptStockLine
  orderUnit: string
  stockUnit: string
  warehouseName: string
  isPending: boolean
  error: string | null
  onRecord: (quantity: number, onSettled: () => void) => void
}) {
  const [quantity, setQuantity] = useState('')
  const [isConfirming, setIsConfirming] = useState(false)
  const [hasTried, setHasTried] = useState(false)

  const amount = Number(quantity)
  const isValid = quantity !== '' && Number.isFinite(amount) && amount > 0
  const reason =
    line.reason ||
    `Satuan pesanan (${orderUnit}) berbeda dengan satuan stok material (${stockUnit}). Catat stok masuk manual setelah jumlahnya dikonversi`

  return (
    <li className="rounded-xl bg-white p-4 shadow-hairline">
      <p className="text-[13px] font-medium text-slate-900">{line.materialName}</p>
      <p className="mt-0.5 text-[13px] text-slate-600">
        {reason}. Diterima baik {number(line.acceptedQuantity)} {orderUnit}.
      </p>
      <div className="mt-3 flex flex-wrap items-end gap-3">
        <div className="w-56">
          <Field
            id={`receipt-${receiptId}-manual-${line.receiptItemId}`}
            label={`Jumlah masuk stok (${stockUnit})`}
            type="number"
            inputMode="decimal"
            min={0.01}
            step={0.01}
            placeholder="0"
            disabled={isPending}
            value={quantity}
            onChange={(event) => {
              setQuantity(event.target.value)
              setIsConfirming(false)
            }}
          />
        </div>
        <ConfirmAction
          size="md"
          label="Catat stok masuk"
          prompt={`Tambah ${number(amount)} ${stockUnit} ${line.materialName} ke ${warehouseName}?`}
          confirmLabel="Ya, catat stok masuk"
          pendingLabel="Mencatat stok masuk"
          isAsking={isConfirming}
          isPending={isPending}
          onAsk={() => {
            setHasTried(true)
            setIsConfirming(isValid)
          }}
          onCancel={() => setIsConfirming(false)}
          onConfirm={() => onRecord(amount, () => setIsConfirming(false))}
        />
      </div>
      {hasTried && !isValid ? (
        <div className="mt-3">
          <ErrorNote message={`Isi jumlah masuk stok dalam ${stockUnit}, lebih dari nol.`} />
        </div>
      ) : null}
      {error ? (
        <div className="mt-3">
          <ErrorNote message={error} />
        </div>
      ) : null}
    </li>
  )
}
