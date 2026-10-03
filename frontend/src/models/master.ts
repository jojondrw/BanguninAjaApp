export type AccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense'

export interface Account {
  id: string
  code: string
  name: string
  type: AccountType
  parentId: string | null
  createdAt: string
  updatedAt: string
}

export interface UnitOfMeasure {
  id: string
  code: string
  name: string
  createdAt: string
  updatedAt: string
}

export const ACCOUNT_TYPE_LABEL: Record<AccountType, string> = {
  asset: 'Aset',
  liability: 'Liabilitas',
  equity: 'Ekuitas',
  revenue: 'Pendapatan',
  expense: 'Beban',
}
