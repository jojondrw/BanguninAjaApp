import { CornerDownLeft, FolderOpen, type LucideIcon, Search } from 'lucide-react'
import { type KeyboardEvent, useDeferredValue, useEffect, useId, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { useProjects } from '../../controllers/useErp'
import { MENU_ITEMS } from './navigation'

interface Command {
  id: string
  group: string
  label: string
  hint?: string
  icon: LucideIcon
  to: string
}

const PAGE_COMMANDS: Command[] = MENU_ITEMS.map((item) => ({
  id: `page:${item.to}`,
  group: 'Halaman',
  label: item.label,
  hint: item.keywords,
  icon: item.icon,
  to: item.to,
}))

const PROJECT_LIMIT = 5

function matches(command: Command, query: string): boolean {
  if (query === '') {
    return true
  }
  const haystack = `${command.label} ${command.hint ?? ''}`.toLowerCase()
  return query
    .toLowerCase()
    .split(/\s+/)
    .every((word) => haystack.includes(word))
}

export function CommandPalette({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) {
  if (!isOpen) {
    return null
  }
  return <PaletteDialog onClose={onClose} />
}

function PaletteDialog({ onClose }: { onClose: () => void }) {
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLUListElement>(null)
  const listId = useId()
  const [query, setQuery] = useState('')
  const [active, setActive] = useState(0)
  const search = useDeferredValue(query.trim())
  const projects = useProjects({ search: search || undefined, pageSize: PROJECT_LIMIT })

  const projectCommands: Command[] = (projects.data?.items ?? []).map((project) => ({
    id: `project:${project.id}`,
    group: search ? 'Proyek' : 'Proyek terbaru',
    label: project.name,
    hint: project.code,
    icon: FolderOpen,
    to: `/proyek/${project.id}`,
  }))
  const commands = [...PAGE_COMMANDS.filter((command) => matches(command, search)), ...projectCommands]
  const current = Math.min(active, Math.max(0, commands.length - 1))

  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null
    inputRef.current?.focus()
    return () => previous?.focus()
  }, [])

  useEffect(() => {
    listRef.current?.querySelector('[aria-selected="true"]')?.scrollIntoView({ block: 'nearest' })
  }, [current])

  function run(command: Command | undefined) {
    if (!command) {
      return
    }
    onClose()
    navigate(command.to)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      onClose()
    } else if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActive((current + 1) % Math.max(1, commands.length))
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActive((current - 1 + commands.length) % Math.max(1, commands.length))
    } else if (event.key === 'Enter') {
      event.preventDefault()
      run(commands[current])
    } else if (event.key === 'Tab') {
      event.preventDefault()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center px-4 pt-[14vh]" onKeyDown={handleKeyDown}>
      <div
        aria-hidden="true"
        className="absolute inset-0 bg-slate-900/25 motion-safe:animate-fade"
        onClick={onClose}
      />
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Cari halaman atau proyek"
        className="relative w-full max-w-[560px] overflow-hidden rounded-2xl bg-white shadow-panel motion-safe:animate-enter"
      >
        <div className="flex items-center gap-3 border-b border-slate-100 px-4">
          <Search aria-hidden="true" className="size-4 shrink-0 text-slate-400" strokeWidth={2} />
          <input
            ref={inputRef}
            role="combobox"
            aria-expanded="true"
            aria-controls={listId}
            aria-activedescendant={commands[current] ? `${listId}-${current}` : undefined}
            aria-autocomplete="list"
            aria-label="Cari halaman atau proyek"
            placeholder="Cari halaman atau proyek…"
            autoComplete="off"
            spellCheck={false}
            value={query}
            onChange={(event) => {
              setQuery(event.target.value)
              setActive(0)
            }}
            className="h-13 flex-1 bg-transparent text-[15px] text-slate-900 placeholder:text-slate-400 focus:outline-none"
          />
          <kbd className="rounded-md bg-slate-100 px-1.5 py-0.5 font-sans text-[11px] font-medium text-slate-500">Esc</kbd>
        </div>

        <ul ref={listRef} id={listId} role="listbox" aria-label="Hasil" className="max-h-[50vh] overflow-y-auto p-2">
          {commands.length === 0 ? (
            <li className="px-3 py-8 text-center text-[13px] text-slate-500">
              {projects.isPending ? 'Mencari…' : `Tidak ada yang cocok dengan "${query}".`}
            </li>
          ) : null}
          {commands.map((command, index) => {
            const Icon = command.icon
            const isActive = index === current
            const startsGroup = index === 0 || commands[index - 1].group !== command.group
            return (
              <li key={command.id} role="presentation">
                {startsGroup ? (
                  <p role="presentation" className="px-3 pt-2 pb-1 text-[11px] font-semibold text-slate-500">
                    {command.group}
                  </p>
                ) : null}
                <div
                  id={`${listId}-${index}`}
                  role="option"
                  aria-selected={isActive}
                  onPointerMove={() => setActive(index)}
                  onClick={() => run(command)}
                  className={`flex h-10 cursor-pointer items-center gap-3 rounded-lg px-3 text-[13px] ${
                    isActive ? 'bg-navy-50 text-slate-900' : 'text-slate-700'
                  }`}
                >
                  <Icon
                    aria-hidden="true"
                    className={`size-4 shrink-0 ${isActive ? 'text-navy-600' : 'text-slate-400'}`}
                    strokeWidth={1.8}
                  />
                  <span className="min-w-0 flex-1 truncate font-medium">{command.label}</span>
                  {command.group !== 'Halaman' && command.hint ? (
                    <span className="text-xs text-slate-500 tabular-nums">{command.hint}</span>
                  ) : null}
                  {isActive ? (
                    <CornerDownLeft aria-hidden="true" className="size-3.5 text-slate-400" strokeWidth={2} />
                  ) : null}
                </div>
              </li>
            )
          })}
        </ul>

        <div className="flex items-center gap-4 border-t border-slate-100 px-4 py-2 text-[11px] text-slate-500">
          <span>
            <Kbd>↑</Kbd> <Kbd>↓</Kbd> pilih
          </span>
          <span>
            <Kbd>Enter</Kbd> buka
          </span>
        </div>
      </div>
    </div>
  )
}

function Kbd({ children }: { children: string }) {
  return <kbd className="rounded bg-slate-100 px-1 py-px font-sans text-[10px] font-medium text-slate-600">{children}</kbd>
}
