import { useState } from 'react'

import { usePhases, useUpdateProjectStatus } from '../../controllers/useProjectWorkspace'
import {
  isReversibleTransition,
  PROGRESS_MAX,
  PROJECT_STATUS_DESCRIPTION,
  PROJECT_STATUS_LABEL,
  PROJECT_STATUS_TONE,
  PROJECT_TRANSITIONS,
  transitionLabel,
  type Project,
  type ProjectStatus,
} from '../../models/project'
import { errorMessage } from '../../shared/errorMessage'
import { Card } from '../components/Data'
import { Button, ErrorNote, SuccessNote } from '../components/Form'
import { Chip, GuardedButton } from './parts'

function irreversibleWarning(current: ProjectStatus, next: ProjectStatus): string {
  if (PROJECT_TRANSITIONS[next].length === 0) {
    return `${PROJECT_STATUS_LABEL[next]} adalah status akhir. Setelah ini status, tahap, dan RAB tidak bisa diubah lagi.`
  }
  return `Status tidak bisa dikembalikan ke ${PROJECT_STATUS_LABEL[current]} setelah ini.`
}

// Hanya perpindahan yang diterima backend yang ditawarkan, masing-masing dengan
// arti status tujuannya. Selesai dan Batal adalah status akhir.
export function ProjectStatusCard({ project, onClose }: { project: Project; onClose: () => void }) {
  const updateStatus = useUpdateProjectStatus(project.id)
  const phases = usePhases(project.id)
  const [asking, setAsking] = useState<ProjectStatus | null>(null)
  const nextStatuses = PROJECT_TRANSITIONS[project.status]

  const unfinished = phases.data?.filter((phase) => phase.progress < PROGRESS_MAX).length ?? 0
  const hasNoPhases = phases.data?.length === 0

  const change = (next: ProjectStatus) =>
    updateStatus.mutate(next, { onSettled: () => setAsking(null) })

  const completionNote = (): string | null => {
    if (unfinished > 0) {
      return `Masih ada ${unfinished} tahap di bawah 100%. Selesaikan dulu progres tahap di tab Ringkasan.`
    }
    if (hasNoPhases) {
      return 'Belum ada tahap, jadi progres proyek akan dicatat 100% saat ditandai selesai.'
    }
    return null
  }

  return (
    <Card
      title="Ubah status proyek"
      description={`Status sekarang ${PROJECT_STATUS_LABEL[project.status]}. ${PROJECT_STATUS_DESCRIPTION[project.status]}`}
    >
      {nextStatuses.length === 0 ? (
        <p className="rounded-xl bg-slate-50 px-4 py-3 text-[13px] text-slate-600">
          {PROJECT_STATUS_LABEL[project.status]} adalah status akhir, jadi statusnya tidak bisa diubah lagi. Data proyek,
          izin, dan penghapusan proyek tetap bisa dilakukan.
        </p>
      ) : (
        <ul className="grid gap-3 md:grid-cols-3" aria-label="Status yang bisa dipilih">
          {nextStatuses.map((next) => {
            const label = transitionLabel(project.status, next)
            const isReversible = isReversibleTransition(project.status, next)
            const isBlocked = next === 'completed' && unfinished > 0
            const note = next === 'completed' ? completionNote() : null
            const noteId = `project-status-note-${next}`
            const isPending = updateStatus.isPending && updateStatus.variables === next

            return (
              <li key={`${project.status}-${next}`} className="flex flex-col rounded-xl bg-slate-50 p-4 shadow-[inset_0_0_0_1px_rgb(0_0_0/0.04)]">
                <Chip tone={PROJECT_STATUS_TONE[next]}>{PROJECT_STATUS_LABEL[next]}</Chip>
                <p className="mt-2 text-[13px] text-slate-600">{PROJECT_STATUS_DESCRIPTION[next]}</p>
                {note ? (
                  <p id={noteId} className={`mt-2 text-xs ${isBlocked ? 'text-amber-800' : 'text-slate-500'}`}>
                    {note}
                  </p>
                ) : null}

                <div className="mt-auto pt-4">
                  {asking === next ? (
                    <div className="space-y-2">
                      <p className="text-xs text-slate-600">{irreversibleWarning(project.status, next)}</p>
                      <div className="flex flex-wrap gap-2">
                        <GuardedButton isPending={isPending} pendingLabel="Mengubah status" onClick={() => change(next)}>
                          Ya, {label.toLowerCase()}
                        </GuardedButton>
                        {isPending ? null : (
                          <Button variant="subtle" onClick={() => setAsking(null)}>
                            Batal
                          </Button>
                        )}
                      </div>
                    </div>
                  ) : (
                    <GuardedButton
                      isPending={isPending}
                      pendingLabel="Mengubah status"
                      disabled={isBlocked || updateStatus.isPending}
                      describedBy={note ? noteId : undefined}
                      onClick={() => (isReversible ? change(next) : setAsking(next))}
                    >
                      {label}
                    </GuardedButton>
                  )}
                </div>
              </li>
            )
          })}
        </ul>
      )}

      <div className="mt-5 flex flex-wrap items-start gap-3">
        <Button variant="subtle" onClick={onClose}>
          Tutup panel status
        </Button>
      </div>

      {updateStatus.isError ? (
        <div className="mt-4">
          <ErrorNote message={errorMessage(updateStatus.error)} />
        </div>
      ) : null}
      {updateStatus.isSuccess ? (
        <div className="mt-4">
          <SuccessNote message={`Status proyek sekarang ${PROJECT_STATUS_LABEL[updateStatus.data.status]}.`} />
        </div>
      ) : null}
    </Card>
  )
}
