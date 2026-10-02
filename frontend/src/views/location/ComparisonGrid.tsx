import { X } from 'lucide-react'
import type { ReactNode } from 'react'

import type { ComparisonTable, SavedLocationDetail } from '../../models/location'
import { number, rupiah, shortDate } from '../../shared/format'
import { RiskChips, ScoreCell, ValueWithMark } from './parts'

const COORDINATE_DIGITS = 4
const LABEL_COLUMN_REM = 11
const LOCATION_COLUMN_REM = 13

const LABEL_CELL =
  'sticky left-0 z-10 w-40 border-b border-slate-100 bg-white py-3 pr-3 pl-5 text-left align-top font-medium ' +
  'text-slate-700 sm:w-44'
const VALUE_CELL = 'border-b border-slate-100 px-3 py-3 align-top text-slate-700 last:pr-5'

export interface LocationNames {
  project: (id: string | null) => string
  profile: (id: string | null) => string
  region: (id: string | null) => string
}

interface ComparisonGridProps {
  locations: SavedLocationDetail[]
  table: ComparisonTable
  names: LocationNames
  onRemove: (id: string) => void
}

function Row<T>({ label, hint, items, render }: {
  label: string
  hint?: string
  items: T[]
  render: (item: T, index: number) => ReactNode
}) {
  return (
    <tr>
      <th scope="row" className={LABEL_CELL}>
        {label}
        {hint ? <span className="mt-0.5 block text-xs font-normal text-slate-500">{hint}</span> : null}
      </th>
      {items.map((item, index) => (
        <td key={index} className={VALUE_CELL}>
          {render(item, index)}
        </td>
      ))}
    </tr>
  )
}

function GroupRow({ label, span }: { label: string; span: number }) {
  return (
    <tr>
      <th
        scope="colgroup"
        colSpan={span}
        className="border-b border-slate-100 bg-slate-50/70 py-2 pr-5 pl-5 text-left text-xs font-medium text-slate-500"
      >
        {label}
      </th>
    </tr>
  )
}

function emptyText(text: string) {
  return <span className="text-slate-400">{text}</span>
}

// Tabel dibalik: baris adalah aspek, kolom adalah lokasi, supaya lokasi bisa
// dibaca berdampingan. Kolom aspek menempel di kiri saat tabel digeser.
export function ComparisonGrid({ locations, table, names, onRemove }: ComparisonGridProps) {
  const span = locations.length + 1
  const minWidth = `${LABEL_COLUMN_REM + locations.length * LOCATION_COLUMN_REM}rem`

  return (
    <div className="-mx-5 overflow-x-auto">
      <table className="w-full border-separate border-spacing-0 text-[13px]" style={{ minWidth }}>
        <caption className="sr-only">
          Perbandingan {locations.length} lokasi. Skor 0 sampai 100, makin tinggi makin baik.
        </caption>
        <thead>
          <tr>
            <th scope="col" className={`${LABEL_CELL} border-t bg-slate-50 text-xs font-medium text-slate-500`}>
              Aspek
            </th>
            {locations.map((location, index) => (
              <th
                key={location.id}
                scope="col"
                className="border-y border-slate-100 bg-slate-50 px-3 py-2.5 text-left align-top last:pr-5"
              >
                <span className="flex items-start justify-between gap-2">
                  <span className="text-[13px] font-semibold text-slate-900">{table.labels[index]}</span>
                  <button
                    type="button"
                    onClick={() => onRemove(location.id)}
                    aria-label={`Keluarkan ${table.labels[index]} dari perbandingan`}
                    className="-m-1 rounded-md p-1 text-slate-400 transition-colors hover:bg-slate-200/70 hover:text-slate-700"
                  >
                    <X aria-hidden="true" className="size-4" strokeWidth={1.8} />
                  </button>
                </span>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          <GroupRow label="Ringkasan" span={span} />
          <Row
            label="Skor keseluruhan"
            hint="0 sampai 100, makin tinggi makin baik"
            items={locations}
            render={(location, index) => <ScoreCell value={location.score} isBest={table.scoreBest.has(index)} />}
          />
          <Row
            label="Penanda risiko"
            hint="Tinggi atau sedang"
            items={table.risks}
            render={(flags) => <RiskChips flags={flags} />}
          />

          <GroupRow label="Skor per dimensi (0 sampai 100, makin tinggi makin baik)" span={span} />
          {table.dimensionRows.map((row) => (
            <Row
              key={row.dimension.id}
              label={row.dimension.name}
              hint={row.dimension.description}
              items={row.values}
              render={(value, index) => <ScoreCell value={value} isBest={row.best.has(index)} />}
            />
          ))}

          <GroupRow label="Data lokasi" span={span} />
          <Row label="Proyek" items={locations} render={(location) => names.project(location.projectId)} />
          <Row label="Profil bangunan" items={locations} render={(location) => names.profile(location.buildingProfileId)} />
          {table.hasRegion ? (
            <Row label="Wilayah" items={locations} render={(location) => names.region(location.regionId)} />
          ) : null}
          <Row
            label="Koordinat"
            items={locations}
            render={(location) => (
              <span className="tabular-nums">
                {location.latitude.toFixed(COORDINATE_DIGITS)}, {location.longitude.toFixed(COORDINATE_DIGITS)}
              </span>
            )}
          />
          {table.hasArea ? (
            <Row
              label="Luas lahan"
              items={locations}
              render={(location) =>
                location.areaSqm > 0 ? (
                  <span className="tabular-nums">{number(location.areaSqm)} m²</span>
                ) : (
                  emptyText('Belum diisi')
                )
              }
            />
          ) : null}
          {table.hasLandPrice ? (
            <Row
              label="Harga tanah per m²"
              hint="Makin murah makin baik"
              items={locations}
              render={(location, index) =>
                location.landPricePerSqm > 0 ? (
                  <ValueWithMark isBest={table.landPriceBest.has(index)}>{rupiah(location.landPricePerSqm)}</ValueWithMark>
                ) : (
                  emptyText('Belum diisi')
                )
              }
            />
          ) : null}
          {table.hasHazardIndex ? (
            <>
              <Row
                label="Indeks banjir"
                hint="0 sampai 1, makin rendah makin aman"
                items={locations}
                render={(location, index) => (
                  <ValueWithMark isBest={table.floodBest.has(index)}>{number(location.floodIndex)}</ValueWithMark>
                )}
              />
              <Row
                label="Indeks gempa bumi"
                hint="0 sampai 1, makin rendah makin aman"
                items={locations}
                render={(location, index) => (
                  <ValueWithMark isBest={table.earthquakeBest.has(index)}>
                    {number(location.earthquakeIndex)}
                  </ValueWithMark>
                )}
              />
            </>
          ) : null}
          {table.hasNote ? (
            <Row
              label="Catatan"
              items={locations}
              render={(location) => (location.note.trim() ? location.note : emptyText('Tidak ada'))}
            />
          ) : null}
          <Row label="Dievaluasi" items={locations} render={(location) => shortDate(location.savedAt)} />
        </tbody>
      </table>
    </div>
  )
}
