import type { OnchainStatus, OnchainTargetKind } from './onchain'

// Bentuk respons mengikuti backend/internal/billing/gateway_dto.go. Status
// memakai nilai yang sama dengan permintaan on-chain.
export type GatewayStatus = OnchainStatus

export interface GatewayConfig {
  enabled: boolean
  provider: string
  testMode: boolean
  channel: string
  maxAmount: number
  requestTtlMinutes: number
}

interface GatewayCommon {
  targetKind: OnchainTargetKind
  amountIdr: number
  provider: string
  channel: string
  qrString: string
  status: GatewayStatus
  failureReason: string
  expiresAt: string
  paidAt: string | null
  createdAt: string
}

export interface GatewayRequest extends GatewayCommon {
  id: string
  targetId: string
  providerId: string
  token: string
  createdBy: string | null
  createdByName: string | null
}

// Data halaman bayar QRIS publik. Tidak memuat nama pelanggan atau id internal.
export interface GatewayPublicPayment extends GatewayCommon {
  targetNumber: string
  projectName: string | null
  testMode: boolean
}

export const GATEWAY_PROVIDER_LABEL: Record<string, string> = {
  xendit: 'Xendit',
  fake: 'simulator lokal',
}

export function providerLabel(provider: string): string {
  return GATEWAY_PROVIDER_LABEL[provider] ?? provider
}

export function qrisPayLink(origin: string, token: string): string {
  return `${origin}/bayar/qris/${token}`
}

// Nominal awal formulir: sisa tagihan, tapi tidak lebih dari batas QRIS per
// transaksi supaya tagihan besar langsung terisi cicilan yang bisa dibayar.
export function defaultQrisAmount(outstanding: number, maxAmount: number): number {
  return Math.max(0, Math.min(outstanding, maxAmount))
}

export function qrisAmountError(value: number, outstanding: number, maxAmount: number): string | null {
  if (!Number.isInteger(value) || value <= 0) {
    return 'Isi nominal dalam rupiah tanpa koma.'
  }
  if (value > outstanding) {
    return 'Nominal melebihi sisa tagihan.'
  }
  if (value > maxAmount) {
    return 'QRIS dibatasi Rp 10 juta per transaksi. Tagih sebagian dulu.'
  }
  return null
}
