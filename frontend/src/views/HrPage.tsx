import { useState } from 'react'

import { useProjects } from '../controllers/useErp'
import { useAttendanceCount, useEmployeeHeadcount, usePayrollSummary } from '../controllers/useHr'
import { EMPLOYMENT_TYPE_LABEL, type EmploymentCount } from '../models/hr'
import { monthLabel, number, rupiahShort } from '../shared/format'
import { currentPeriod, todayDate } from '../shared/localDate'
import { AppShell } from './components/AppShell'
import { Kpi, KpiRow } from './components/Data'
import { kpiText } from './components/kpiText'
import { AttendanceSection } from './hr/AttendanceSection'
import { EmployeesSection } from './hr/EmployeesSection'
import { PayrollSection } from './hr/PayrollSection'

function headcountNote(byType: EmploymentCount[]): string {
  if (byType.length === 0) {
    return 'Belum ada karyawan aktif'
  }
  return byType.map((count) => `${count.total} ${EMPLOYMENT_TYPE_LABEL[count.employmentType].toLowerCase()}`).join(', ')
}

export function HrPage() {
  const [today] = useState(todayDate)
  const [period, setPeriod] = useState(currentPeriod)
  const headcount = useEmployeeHeadcount()
  const presentToday = useAttendanceCount({ dateFrom: today, dateTo: today, status: 'present' })
  const recordedToday = useAttendanceCount({ dateFrom: today, dateTo: today })
  const payrollSummary = usePayrollSummary({ period })
  const projects = useProjects({ pageSize: 100 })
  const projectList = projects.data?.items ?? []

  const projectName = (id: string | null) => {
    if (id === null) {
      return 'Kantor pusat'
    }
    return projectList.find((project) => project.id === id)?.name ?? id.slice(0, 8)
  }

  return (
    <AppShell title="SDM" description="Karyawan, absensi harian, dan penggajian per periode">
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

      <div className="mt-6 grid gap-6">
        <EmployeesSection projects={projectList} projectName={projectName} today={today} />
        <AttendanceSection today={today} />
        <PayrollSection period={period} onPeriodChange={setPeriod} projects={projectList} />
      </div>
    </AppShell>
  )
}
