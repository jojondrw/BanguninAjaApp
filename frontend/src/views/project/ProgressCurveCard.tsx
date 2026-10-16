import { useProgressCurve } from '../../controllers/useProjectWorkspace'
import { scheduleStatus, type Project } from '../../models/project'
import { Card, Empty, LoadFailed, Loading } from '../components/Data'
import { Chip } from './parts'
import { ProgressCurveChart } from './ProgressCurveChart'

const STATUS_TONE = {
  good: 'bg-green-100 text-green-800',
  warn: 'bg-amber-100 text-amber-900',
  neutral: 'bg-slate-100 text-slate-700',
} as const

export function ProgressCurveCard({ project }: { project: Project }) {
  const curve = useProgressCurve(project.id)
  const status = curve.data ? scheduleStatus(curve.data.today.deviation) : null

  return (
    <Card
      title="Kurva S"
      description="Progres rencana dari jadwal tahap dibanding progres nyata"
      action={status && curve.data && curve.data.points.length > 0 ? <Chip tone={STATUS_TONE[status.tone]}>{status.label}</Chip> : undefined}
    >
      {curve.isPending ? <Loading /> : null}
      {curve.isError ? <LoadFailed onRetry={() => curve.refetch()} /> : null}
      {curve.data ? (
        curve.data.points.length === 0 ? (
          <Empty message="Isi tanggal mulai dan target selesai tiap tahap untuk melihat kurva S." />
        ) : (
          <>
            <ProgressCurveChart points={curve.data.points} today={curve.data.today.date} />
            <CurveNotes weighting={curve.data.weighting} undatedPhases={curve.data.undatedPhases} />
          </>
        )
      ) : null}
    </Card>
  )
}

function CurveNotes({ weighting, undatedPhases }: { weighting: 'budget' | 'equal'; undatedPhases: number }) {
  const notes: string[] = []
  if (weighting === 'equal') {
    notes.push('Bobot tahap masih sama rata. Kaitkan item RAB ke tahap supaya progres mengikuti nilai pekerjaan.')
  }
  if (undatedPhases > 0) {
    notes.push(`${undatedPhases} tahap belum punya tanggal, jadi belum masuk garis rencana.`)
  }
  if (notes.length === 0) {
    return null
  }
  return (
    <ul className="mt-4 space-y-1 text-xs text-slate-500">
      {notes.map((note) => (
        <li key={note}>{note}</li>
      ))}
    </ul>
  )
}
