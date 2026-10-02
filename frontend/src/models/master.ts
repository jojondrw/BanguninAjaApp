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

export interface AccountGroup {
  type: AccountType
  label: string
  accounts: Account[]
}

// Kas keluar biasanya dibebankan ke akun beban, kas masuk ke akun pendapatan,
// jadi kelompok itu ditaruh paling atas. Akun induk (yang punya anak) tidak
// ditawarkan karena transaksi dicatat di akun rinciannya.
const GROUP_ORDER: Record<'in' | 'out', AccountType[]> = {
  out: ['expense', 'asset', 'liability', 'equity', 'revenue'],
  in: ['revenue', 'liability', 'equity', 'asset', 'expense'],
}

export function accountGroups(accounts: Account[], cashType: 'in' | 'out'): AccountGroup[] {
  const parents = new Set(accounts.map((account) => account.parentId).filter((id) => id !== null))
  const postable = accounts
    .filter((account) => !parents.has(account.id))
    .sort((a, b) => a.code.localeCompare(b.code))

  return GROUP_ORDER[cashType]
    .map((type) => ({
      type,
      label: ACCOUNT_TYPE_LABEL[type],
      accounts: postable.filter((account) => account.type === type),
    }))
    .filter((group) => group.accounts.length > 0)
}
