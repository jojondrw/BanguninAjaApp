import type { ChangeEvent } from 'react'

import { materialOptions, unitOfMeasureOptions, type MaterialOption } from '../../models/lookupApi'
import { emptyOrderLine, orderLineTotal, type OrderLineValues } from '../../models/procurement'
import { rupiah } from '../../shared/format'
import { Button, Field } from '../components/Form'
import { SearchSelect } from '../components/SearchSelect'

function MaterialLineFields({ idPrefix, line, index, usedMaterials, showPrice, canRemove, onChange, onRemove }: {
  idPrefix: string
  line: OrderLineValues
  index: number
  usedMaterials: string[]
  showPrice: boolean
  canRemove: boolean
  onChange: (line: OrderLineValues) => void
  onRemove: () => void
}) {
  const id = `${idPrefix}-line-${line.key}`

  // Memilih material mengisi satuan bawaan dan harga terakhirnya. Keduanya
  // tetap boleh diubah sesudahnya.
  const chooseMaterial = (materialId: string, material: MaterialOption | null) => {
    onChange({
      ...line,
      materialId,
      unitOfMeasureId: material?.unitOfMeasureId ?? line.unitOfMeasureId,
      unitPrice: line.unitPrice === '' && material ? String(material.lastPrice) : line.unitPrice,
    })
  }

  const update = (key: 'quantity' | 'unitPrice') =>
    (event: ChangeEvent<HTMLInputElement>) => onChange({ ...line, [key]: event.target.value })

  return (
    <fieldset className="rounded-xl bg-white p-4 shadow-hairline">
      <legend className="float-left mb-3 w-full text-xs font-medium text-slate-500">Baris {index + 1}</legend>
      <div className={`clear-both grid gap-4 ${showPrice ? 'md:grid-cols-[2fr_1fr_1fr_1fr]' : 'md:grid-cols-[2fr_1fr_1fr]'}`}>
        <SearchSelect
          {...materialOptions}
          id={`${id}-material`}
          label="Material"
          placeholder="Cari nama atau kode material"
          required
          exclude={usedMaterials}
          value={line.materialId}
          onChange={chooseMaterial}
        />
        <Field
          id={`${id}-quantity`}
          label="Jumlah"
          type="number"
          inputMode="decimal"
          min={0.01}
          step={0.01}
          required
          value={line.quantity}
          onChange={update('quantity')}
        />
        <SearchSelect
          {...unitOfMeasureOptions}
          id={`${id}-unit`}
          label="Satuan"
          placeholder="Cari satuan"
          required
          value={line.unitOfMeasureId}
          onChange={(unitOfMeasureId) => onChange({ ...line, unitOfMeasureId })}
        />
        {showPrice ? (
          <Field
            id={`${id}-price`}
            label="Harga satuan"
            type="number"
            inputMode="numeric"
            min={0}
            step={1}
            placeholder="0"
            value={line.unitPrice}
            onChange={update('unitPrice')}
          />
        ) : null}
      </div>
      {showPrice || canRemove ? (
        <div className="mt-3 flex items-center justify-between gap-3">
          {showPrice ? (
            <p className="text-sm text-slate-600 tabular-nums">Subtotal {rupiah(orderLineTotal(line))}</p>
          ) : (
            <span />
          )}
          {canRemove ? (
            <button
              type="button"
              onClick={onRemove}
              className="rounded-md px-2 py-1 text-xs font-medium text-red-700 hover:bg-red-50"
            >
              Hapus baris {index + 1}
            </button>
          ) : null}
        </div>
      ) : null}
    </fieldset>
  )
}

// Daftar baris material untuk permintaan (tanpa harga) dan pesanan (dengan
// harga). Satu material hanya boleh sekali per dokumen, sama dengan aturan
// material_duplicate di backend, jadi material yang sudah dipakai dikunci.
export function MaterialLinesEditor({ idPrefix, lines, showPrice, onChange }: {
  idPrefix: string
  lines: OrderLineValues[]
  showPrice: boolean
  onChange: (lines: OrderLineValues[]) => void
}) {
  const usedMaterials = lines.map((line) => line.materialId).filter((id) => id !== '')
  const total = lines.reduce((sum, line) => sum + orderLineTotal(line), 0)

  return (
    <div className="space-y-3">
      {lines.map((line, index) => (
        <MaterialLineFields
          key={line.key}
          idPrefix={idPrefix}
          line={line}
          index={index}
          usedMaterials={usedMaterials}
          showPrice={showPrice}
          canRemove={lines.length > 1}
          onChange={(next) => onChange(lines.map((item) => (item.key === line.key ? next : item)))}
          onRemove={() => onChange(lines.filter((item) => item.key !== line.key))}
        />
      ))}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Button variant="subtle" onClick={() => onChange([...lines, emptyOrderLine()])}>
          Tambah baris material
        </Button>
        {showPrice ? <p className="text-sm font-medium text-slate-900 tabular-nums">Total {rupiah(total)}</p> : null}
      </div>
    </div>
  )
}
