import type { ReactNode } from 'react'

import { splitMoney } from '../../shared/money'

export function Money({ children }: { children: string }) {
  return <span className="font-uang">{children}</span>
}

// Teks biasa dilewatkan apa adanya; hanya potongan Rupiah yang diberi font uang.
export function MoneyText({ children }: { children: ReactNode }) {
  if (typeof children !== 'string') {
    return children
  }
  const parts = splitMoney(children)
  if (!parts.some((part) => part.money)) {
    return children
  }
  return parts.map((part, index) =>
    part.money ? <Money key={`${index}-${part.text}`}>{part.text}</Money> : part.text,
  )
}
