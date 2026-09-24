export interface Budget {
  id: string
  projectId: string
  year: number
  value: number
  note: string
  realized: number
  remaining: number
  absorption: number
  createdAt: string
  updatedAt: string
}

export interface CashTransaction {
  id: string
  date: string
  type: 'in' | 'out'
  accountId: string
  projectId: string | null
  amount: number
  note: string
  createdAt: string
  updatedAt: string
}

export interface CashFlowPeriod {
  period: string
  cashIn: number
  cashOut: number
  net: number
}

export interface CashFlow {
  dateFrom: string
  dateTo: string
  periods: CashFlowPeriod[]
  totalIn: number
  totalOut: number
  net: number
  balance: number
}

export interface CashFlowFilter {
  projectId?: string
  dateFrom?: string
  dateTo?: string
}
