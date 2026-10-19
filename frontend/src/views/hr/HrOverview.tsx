import { useHrOverview } from '../../controllers/useHr'
import { projectOptions } from '../../models/lookupApi'
import { EMPLOYMENT_TYPE_LABEL, type HrOverview as Overview } from '../../models/hr'
import { monthLabel, number, rupiahShort } from '../../shared/format'
import { Card, LoadFailed, Loading } from '../components/Data'
import { DonutChart } from '../components/DonutChart'
import { LinkButton } from '../components/LinkButton'
import { LookupName } from '../components/LookupName'
import { Money } from '../components/Money'

const TYPE_COLORS = ['#0b2b6b', '#1f5fd1', '#fbbb16', '#94a3b8']
const TOP_PROJECTS = 6
const ATTENDANCE = [
  { key: 'present', label: 'Hadir', color: 'bg-green-500' },
  { key: 'permitted', label: 'Izin', color: 'bg-sky-400' },
  { key: 'sick', label: 'Sakit', color: 'bg-amber-400' },
  { key: 'absent', label: 'Alpa', color: 'bg-red-500' },
] as const

export interface HrDrill {
  openEmployees: () => void
  openAttendance: () => void
  openPayroll: () => void
}

function attendanceRate(days: Overview['days']): number {
  const present = days.reduce((sum, day) => sum + day.present, 0)
  const recorded = days.reduce((sum, day) => sum + day.present + day.sick + day.permitted + day.absent, 0)
  return recorded > 0 ? Math.round((present / recorded) * 100) : 0
}

function AttendanceBars({ days }: { days: Overview['days'] }) {
  if (days.length === 0) {
    return <p className="text-[13px] text-slate-500">Belum ada absensi dalam 14 hari terakhir.</p>
  }
  const highest = Math.max(1, ...days.map((day) => day.present + day.sick + day.permitted + day.absent))
  return (
    <div>
      <div className="mb-3 flex flex-wrap gap-3 text-xs text-slate-500">
        {ATTENDANCE.map((item) => (
          <span key={item.key} className="inline-flex items-center gap-1.5">
            <span aria-hidden="true" className={`size-2 rounded-full ${item.color}`} />
            {item.label}
          </span>
        ))}
      </div>
      <ul className="flex h-36 items-end gap-1.5" aria-label="Absensi per hari">
        {days.map((day) => {
          const total = day.present + day.sick + day.permitted + day.absent
          return (
            <li key={day.date} className="flex h-full min-w-0 flex-1 flex-col items-center justify-end gap-1">
              <span
                className="flex w-full max-w-5 flex-col-reverse overflow-hidden rounded-t-[3px]"
                style={{ height: `${(total / highest) * 100}%` }}
                title={`${day.date}: hadir ${day.present}, izin ${day.permitted}, sakit ${day.sick}, alpa ${day.absent}`}
              >
                {ATTENDANCE.map((item) =>
                  day[item.key] > 0 ? (
                    <span key={item.key} className={item.color} style={{ height: `${(day[item.key] / total) * 100}%` }} />
                  ) : null,
                )}
              </span>
              <span className="text-[10px] text-slate-400 tabular-nums">{day.date.slice(8, 10)}</span>
            </li>
          )
        })}
      </ul>
    </div>
  )
}

function PayrollBars({ payrolls }: { payrolls: Overview['payrolls'] }) {
  const highest = Math.max(1, ...payrolls.map((payroll) => payroll.netPay))
  return (
    <ul className="space-y-2">
      {payrolls.map((payroll) => (
        <li key={payroll.period} className="grid grid-cols-[6.5rem_1fr_6rem] items-center gap-3 text-[13px]">
          <span className="text-slate-600">{monthLabel(payroll.period)}</span>
          <span aria-hidden="true" className="flex h-2.5 overflow-hidden rounded-full bg-slate-100">
            <span className="bar-fill h-full" style={{ width: `${((payroll.netPay - payroll.unpaid) / highest) * 100}%` }} />
            <span className="h-full bg-amber-400" style={{ width: `${(payroll.unpaid / highest) * 100}%` }} />
          </span>
          <span className="text-right text-slate-900 tabular-nums">
            <Money>{rupiahShort(payroll.netPay)}</Money>
          </span>
        </li>
      ))}
    </ul>
  )
}

// Ringkasan SDM: komposisi karyawan, sebaran per proyek, absensi 14 hari, dan
// gaji enam bulan. Rincian ada di tab Karyawan, Absensi, dan Penggajian.
export function HrOverview({ drill }: { drill: HrDrill }) {
  const overview = useHrOverview()

  if (overview.isPending) {
    return <Loading />
  }
  if (overview.isError) {
    return <LoadFailed onRetry={() => overview.refetch()} />
  }
  const data = overview.data
  const topProjects = data.byProject.slice(0, TOP_PROJECTS)
  const highest = Math.max(1, ...topProjects.map((project) => project.total))
  const unpaid = data.payrolls.reduce((sum, payroll) => sum + payroll.unpaid, 0)

  return (
    <div className="grid gap-6 xl:grid-cols-2">
      <Card title="Karyawan aktif" action={<LinkButton label="Semua karyawan" onClick={drill.openEmployees} />}>
        <div className="grid gap-6 sm:grid-cols-2">
          <DonutChart
            caption="Karyawan aktif menurut jenis kerja"
            colors={TYPE_COLORS}
            slices={data.byType.map((type) => ({ label: EMPLOYMENT_TYPE_LABEL[type.employmentType], value: type.total }))}
            format={(value) => `${number(value)} orang`}
          />
          <div>
            <p className="mb-2 text-xs font-medium text-slate-500">Proyek dengan karyawan terbanyak</p>
            <ul className="space-y-2">
              {topProjects.map((project) => (
                <li key={project.projectId ?? 'kantor'} className="text-[13px]">
                  <div className="flex items-baseline justify-between gap-3">
                    <span className="truncate text-slate-700">
                      {project.projectId ? <LookupName source={projectOptions} value={project.projectId} /> : 'Kantor pusat'}
                    </span>
                    <span className="shrink-0 text-slate-900 tabular-nums">{number(project.total)}</span>
                  </div>
                  <span aria-hidden="true" className="mt-1 block h-1.5 overflow-hidden rounded-full bg-slate-100">
                    <span className="bar-fill block h-full rounded-full" style={{ width: `${(project.total / highest) * 100}%` }} />
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </Card>

      <Card
        title="Absensi 14 hari"
        description={`Kehadiran ${attendanceRate(data.days)}% dari absensi yang tercatat`}
        action={<LinkButton label="Absensi" onClick={drill.openAttendance} />}
      >
        <AttendanceBars days={data.days} />
      </Card>

      <Card
        title="Gaji bersih 6 bulan"
        description={unpaid > 0 ? `${rupiahShort(unpaid)} belum dibayar (kuning)` : 'Semua slip sudah dibayar'}
        action={<LinkButton label="Penggajian" onClick={drill.openPayroll} />}
      >
        <PayrollBars payrolls={data.payrolls} />
      </Card>
    </div>
  )
}
