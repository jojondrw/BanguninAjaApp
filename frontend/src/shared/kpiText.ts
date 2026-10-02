interface QueryState<T> {
  isPending: boolean
  isError: boolean
  data: T | undefined
}

// Isi kartu KPI: titik tiga selama memuat, strip kalau gagal.
export function kpiText<T>(query: QueryState<T>, format: (data: T) => string): string {
  if (query.isPending) {
    return '...'
  }
  if (query.isError || query.data === undefined) {
    return '-'
  }
  return format(query.data)
}
