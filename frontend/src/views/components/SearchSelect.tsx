import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { ChevronDown, X } from 'lucide-react'
import { useEffect, useRef, useState, type KeyboardEvent } from 'react'

import type { SearchOption } from '../../models/lookup'
import { SEARCH_LIMIT } from '../../models/lookupApi'
import { CONTROL_CLASS, LABEL_CLASS } from './Form'
import { optionDomId, resolveLabel, stepIndex, useDebounced } from './searchSelectLogic'

interface SearchSelectProps<O extends SearchOption> {
  id: string
  label: string
  value: string
  onChange: (value: string, option: O | null) => void
  fetchOptions: (search: string) => Promise<O[]>
  fetchSelected?: (value: string) => Promise<O>
  cacheKey: string
  initial?: SearchOption | null
  exclude?: string[]
  placeholder?: string
  allowEmpty?: boolean
  emptyLabel?: string
  hint?: string
  required?: boolean
  disabled?: boolean
  compact?: boolean
  className?: string
}

// Kotak cari-dan-pilih (WAI-ARIA combobox + listbox). Pilihan dicari ke server
// sambil mengetik, jadi data di luar halaman pertama tetap bisa dipilih.
export function SearchSelect<O extends SearchOption = SearchOption>({
  id,
  label,
  value,
  onChange,
  fetchOptions,
  fetchSelected,
  cacheKey,
  initial,
  exclude,
  placeholder = 'Ketik untuk mencari',
  allowEmpty = false,
  emptyLabel = 'Semua',
  hint,
  required = false,
  disabled = false,
  compact = false,
  className = '',
}: SearchSelectProps<O>) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [isOpen, setOpen] = useState(false)
  const [text, setText] = useState('')
  const [isTyping, setTyping] = useState(false)
  const [active, setActive] = useState(-1)
  const [picked, setPicked] = useState<O | null>(null)

  const listId = `${id}-listbox`
  const hintId = hint ? `${id}-hint` : undefined
  const search = useDebounced(isTyping ? text.trim() : '')

  const results = useQuery({
    queryKey: ['search-select', cacheKey, search],
    queryFn: () => fetchOptions(search),
    enabled: isOpen && !disabled,
    staleTime: 30_000,
    placeholderData: keepPreviousData,
  })
  const found = (results.data ?? []).filter((option) => option.value === value || !exclude?.includes(option.value))

  const knownLabel = resolveLabel(value, [picked, initial], found)
  const selected = useQuery({
    queryKey: ['search-select', cacheKey, 'selected', value],
    queryFn: () => fetchSelected?.(value) ?? Promise.reject(new Error('no lookup')),
    enabled: Boolean(fetchSelected) && value !== '' && knownLabel === '',
    staleTime: 5 * 60_000,
    retry: false,
  })
  const selectedLabel = knownLabel || (selected.data?.value === value ? selected.data.label : '')
  const shownLabel = value === '' ? (allowEmpty ? emptyLabel : '') : selectedLabel || (selected.isError ? 'Data tidak ditemukan' : 'Memuat…')

  const options: SearchOption[] =
    allowEmpty && search === '' ? [{ value: '', label: emptyLabel }, ...found] : found
  const activeIndex = active < options.length ? active : options.length - 1
  const activeId = isOpen && activeIndex >= 0 ? optionDomId(listId, activeIndex) : undefined

  useEffect(() => {
    inputRef.current?.setCustomValidity(required && value === '' ? `Pilih ${label.toLowerCase()} dari daftar.` : '')
  }, [required, value, label])

  useEffect(() => {
    if (activeId) document.getElementById(activeId)?.scrollIntoView({ block: 'nearest' })
  }, [activeId])

  function open() {
    if (disabled) return
    setOpen(true)
    setActive(Math.max(0, options.findIndex((option) => option.value === value)))
  }

  function close() {
    setOpen(false)
    setTyping(false)
    setText('')
    setActive(-1)
  }

  function choose(option: SearchOption) {
    const chosen = option.value === '' ? null : (option as O)
    setPicked(chosen)
    onChange(option.value, chosen)
    close()
  }

  function clear() {
    setPicked(null)
    onChange('', null)
    close()
    inputRef.current?.focus()
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
      event.preventDefault()
      if (!isOpen) {
        open()
        return
      }
      setActive(stepIndex(activeIndex, options.length, event.key === 'ArrowDown' ? 1 : -1))
    } else if (event.key === 'Enter' && isOpen) {
      event.preventDefault()
      const option = options[activeIndex]
      if (option) choose(option)
    } else if (event.key === 'Escape' && isOpen) {
      event.preventDefault()
      close()
    }
  }

  const height = compact ? 'h-8 text-[13px]' : 'h-9'
  const canClear = value !== '' && !disabled
  const status = results.isError
    ? 'Pilihan gagal dimuat. Coba ketik lagi.'
    : results.isFetching && found.length === 0
      ? 'Mencari…'
      : options.length === 0
        ? 'Tidak ada yang cocok.'
        : found.length >= SEARCH_LIMIT
          ? `Menampilkan ${SEARCH_LIMIT} teratas. Ketik untuk mempersempit.`
          : ''

  return (
    <div className={`flex flex-col gap-1.5 ${className}`}>
      <label htmlFor={id} className={compact ? 'sr-only' : LABEL_CLASS}>
        {label}
      </label>
      <div className="relative">
        <input
          ref={inputRef}
          id={id}
          type="text"
          role="combobox"
          aria-expanded={isOpen}
          aria-controls={listId}
          aria-autocomplete="list"
          aria-activedescendant={activeId}
          aria-describedby={hintId}
          aria-required={required || undefined}
          autoComplete="off"
          spellCheck={false}
          disabled={disabled}
          placeholder={placeholder}
          value={isTyping ? text : shownLabel}
          onChange={(event) => {
            setText(event.target.value)
            setTyping(true)
            setOpen(true)
            setActive(0)
          }}
          onFocus={(event) => event.target.select()}
          onClick={() => (isOpen ? undefined : open())}
          onBlur={close}
          onKeyDown={handleKeyDown}
          className={`${CONTROL_CLASS} ${height} w-full min-w-0 pr-14 pl-2.5`}
        />
        <div className="absolute inset-y-0 right-1.5 flex items-center gap-0.5">
          {canClear ? (
            <button
              type="button"
              aria-label={`Kosongkan ${label.toLowerCase()}`}
              onMouseDown={(event) => event.preventDefault()}
              onClick={clear}
              className="grid size-6 place-items-center rounded-md text-slate-400 hover:bg-slate-100 hover:text-slate-600"
            >
              <X aria-hidden="true" className="size-3.5" strokeWidth={2} />
            </button>
          ) : null}
          <ChevronDown aria-hidden="true" className="pointer-events-none size-4 text-slate-400" strokeWidth={1.8} />
        </div>
        {isOpen ? (
          <div className="absolute z-40 mt-1 w-full min-w-56 overflow-hidden rounded-lg border border-slate-200 bg-white shadow-lg">
            <ul id={listId} role="listbox" aria-label={label} className="max-h-64 overflow-y-auto py-1">
              {options.map((option, index) => (
                <li
                  key={option.value || 'empty'}
                  id={optionDomId(listId, index)}
                  role="option"
                  aria-selected={option.value === value}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => choose(option)}
                  onMouseEnter={() => setActive(index)}
                  className={`flex cursor-pointer items-baseline justify-between gap-3 px-3 py-1.5 text-[13px] ${
                    index === activeIndex ? 'bg-slate-100' : ''
                  } ${option.value === value ? 'font-medium text-navy-700' : 'text-slate-800'}`}
                >
                  <span className="truncate">{option.label}</span>
                  {option.hint ? <span className="shrink-0 text-xs text-slate-500">{option.hint}</span> : null}
                </li>
              ))}
            </ul>
            {status ? (
              <p role="status" className="border-t border-slate-100 px-3 py-1.5 text-xs text-slate-500">
                {status}
              </p>
            ) : null}
          </div>
        ) : null}
      </div>
      {hint ? (
        <p id={hintId} className="text-xs text-slate-500">
          {hint}
        </p>
      ) : null}
    </div>
  )
}
