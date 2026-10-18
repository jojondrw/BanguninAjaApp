import { useContractSummary, useUnitSummary } from '../../controllers/useSales'
import { percentOf, unitFunnel, type FunnelStage } from '../../models/overview'
import { number } from '../../shared/format'
import { Empty, LoadFailed, Loading } from '../components/Data'

const BAND_HEIGHT = 54
const MIN_WIDTH = 16
const LAST_TAPER = 0.78
const BAND_COLORS = ['var(--color-navy-500)', 'var(--color-navy-600)', 'var(--color-navy-700)', 'var(--color-navy-800)']

// Corong unit dari dipasarkan sampai kontrak lunas. Lebar tiap pita sebanding
// dengan jumlahnya, tepi bawah menyambung ke pita berikutnya sehingga bentuknya
// benar-benar mengerucut seperti corong.
export function SalesFunnel() {
  const units = useUnitSummary()
  const contracts = useContractSummary()

  if (units.isPending || contracts.isPending) {
    return <Loading />
  }
  if (units.isError || contracts.isError) {
    return (
      <LoadFailed
        onRetry={() => {
          void units.refetch()
          void contracts.refetch()
        }}
      />
    )
  }

  const stages = unitFunnel(units.data, contracts.data)
  if (stages[0].value === 0) {
    return <Empty message="Belum ada unit yang dipasarkan." />
  }

  const widths = bandWidths(stages)
  const overall = percentOf(stages[stages.length - 1].value, stages[0].value)

  return (
    <div>
      <ol className="space-y-[3px]">
        {stages.map((stage, index) => (
          <li key={stage.label} className="grid grid-cols-[minmax(0,1fr)_minmax(0,9rem)] items-center gap-4">
            <Band
              top={widths[index]}
              bottom={index + 1 < widths.length ? widths[index + 1] : widths[index] * LAST_TAPER}
              color={BAND_COLORS[index % BAND_COLORS.length]}
              value={stage.value}
            />
            <div className="min-w-0">
              <p className="truncate text-[13px] font-medium text-slate-900">{stage.label}</p>
              <p className="text-xs text-slate-500 tabular-nums">
                {index === 0 ? 'titik awal' : `konversi ${percentOf(stage.value, stages[index - 1].value)}%`}
              </p>
            </div>
          </li>
        ))}
      </ol>
      <p className="mt-4 text-xs text-slate-500">
        <span className="font-semibold text-slate-900 tabular-nums">{overall}%</span> unit yang dipasarkan sudah menjadi
        kontrak lunas. Unit yang ditahan tidak dihitung.
      </p>
    </div>
  )
}

function bandWidths(stages: FunnelStage[]): number[] {
  const top = Math.max(1, stages[0].value)
  const widths: number[] = []
  stages.forEach((stage, index) => {
    const share = Math.max(MIN_WIDTH, (stage.value / top) * 100)
    widths.push(index === 0 ? share : Math.min(widths[index - 1], share))
  })
  return widths
}

function Band({ top, bottom, color, value }: { top: number; bottom: number; color: string; value: number }) {
  const left = (100 - top) / 2
  const right = left + top
  const lowerLeft = (100 - bottom) / 2
  const lowerRight = lowerLeft + bottom

  return (
    <div className="relative" style={{ height: BAND_HEIGHT }}>
      <svg
        viewBox={`0 0 100 ${BAND_HEIGHT}`}
        preserveAspectRatio="none"
        aria-hidden="true"
        className="absolute inset-0 size-full motion-safe:animate-fade"
      >
        <polygon points={`${left},0 ${right},0 ${lowerRight},${BAND_HEIGHT} ${lowerLeft},${BAND_HEIGHT}`} fill={color} />
      </svg>
      <span className="absolute inset-0 grid place-items-center text-[17px] font-semibold tracking-[-0.02em] text-white tabular-nums">
        {number(value)}
      </span>
    </div>
  )
}
