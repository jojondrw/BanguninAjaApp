import { useState } from 'react'

import { useAttendanceCount, useEmployeeHeadcount, usePayrollSummary } from '../controllers/useHr'
import { EMPLOYMENT_TYPE_LABEL, type EmploymentCount } from '../models/hr'
import { monthLabel, number, rupiahShort } from '../shared/format'
import { currentPeriod, todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { Kpi, KpiRow } from './components/Data'
import { kpiText } from './components/kpiText'
import { SectionTabs } from './components/ListTools'
import { AttendanceSection } from './hr/AttendanceSection'
import { EmployeesSection } from './hr/EmployeesSection'
import { HrOverview } from './hr/HrOverview'
import { PayrollSection } from './hr/PayrollSection'

type HrTab = 'overview' | 'employees' | 'attendance' | 'payroll'

const TABS: { value: HrTab; label: string }[] = [
  { value: 'overview', label: 'Ringkasan' },
  { value: 'employees', label: 'Karyawan' },
  { value: 'attendance', label: 'Absensi' },
  { value: 'payroll', label: 'Penggajian' },
]

function headcountNote(byType: EmploymentCount[]): string {
  if (byType.length === 0) {
    return 'Belum ada karyawan aktif'
  }
  return byType.map((count) => `${count.total} ${EMPLOYMENT_TYPE_LABEL[count.employmentType].toLowerCase()}`).join(', ')
}

export function HrPage() {
  const [today] = useState(todayDate)
  const [period, setPeriod] = useState(currentPeriod)
  const [tab, setTab] = useState<HrTab>('overview')
  const headcount = useEmployeeHeadcount()
  const presentToday = useAttendanceCount({ dateFrom: today, dateTo: today, status: 'present' })
  const recordedToday = useAttendanceCount({ dateFrom: today, dateTo: today })
  const payrollSummary = usePayrollSummary({ period })
  return (
    <AppShell
      title="SDM"
      description="Karyawan, absensi harian, dan penggajian per periode"
      tabs={<SectionTabs tabs={TABS} active={tab} onChange={setTab} label="Bagian SDM" />}
    >
      <KpiRow>
        <Kpi
          label="Karyawan aktif"
          value={kpiText(headcount, (data) => number(data.total))}
          note={headcount.data ? headcountNote(headcount.data.byType) : undefined}
        />
        <Kpi
          label="Hadir hari ini"
          value={kpiText(presentToday, number)}
          note={recordedToday.data === undefined ? undefined : `dari ${number(recordedToday.data)} absensi tercatat`}
        />
        <Kpi
          label={`Gaji bersih ${monthLabel(period)}`}
          value={kpiText(payrollSummary, (data) => rupiahShort(data.netPay))}
          note={payrollSummary.data ? `${number(payrollSummary.data.count)} slip gaji` : undefined}
        />
        <Kpi
          label="Belum dibayar"
          value={kpiText(payrollSummary, (data) => rupiahShort(data.unpaidNetPay))}
          note={payrollSummary.data ? `${number(payrollSummary.data.unpaidCount)} slip menunggu pembayaran` : undefined}
        />
      </KpiRow>

      <div className="mt-6">
        {tab === 'overview' ? (
          <HrOverview
            drill={{
              openEmployees: () => setTab('employees'),
              openAttendance: () => setTab('attendance'),
              openPayroll: () => setTab('payroll'),
            }}
          />
        ) : null}
        {tab === 'employees' ? <EmployeesSection today={today} /> : null}
        {tab === 'attendance' ? <AttendanceSection today={today} /> : null}
        {tab === 'payroll' ? <PayrollSection period={period} onPeriodChange={setPeriod} /> : null}
      </div>
    </AppShell>
  )
}
