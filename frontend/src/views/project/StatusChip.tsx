import { PROJECT_STATUS_LABEL, PROJECT_STATUS_TONE, type Project } from '../../models/project'
import { CHIP_CLASS } from '../components/ListTools'

export function StatusChip({ project }: { project: Project }) {
  return (
    <span className={`${CHIP_CLASS} ${PROJECT_STATUS_TONE[project.status]}`}>
      {PROJECT_STATUS_LABEL[project.status]}
    </span>
  )
}
