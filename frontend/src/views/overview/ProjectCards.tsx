import type { Project } from '../../models/project'
import { ProjectCard } from '../project/ProjectCard'

export function ProjectCards({ projects }: { projects: Project[] }) {
  return (
    <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {projects.map((project) => (
        <li key={project.id} className="min-w-0">
          <ProjectCard project={project} />
        </li>
      ))}
    </ul>
  )
}
