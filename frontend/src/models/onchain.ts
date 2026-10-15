import { rupiah } from '../shared/format'

// Bentuk respons mengikuti backend/internal/billing/onchain_dto.go. Nilai wei
// dikirim sebagai teks karena bisa melebihi batas angka JavaScript.
export type OnchainStatus = 'pending' | 'paid' | 'expired' | 'failed' | 'cancelled'
export type OnchainTargetKind = 'invoice' | 'receivable'
export type OnchainVerification = 'paid' | 'waiting' | 'failed'
// Hanya faktur pelanggan dan piutang yang bisa dibayar on-chain, bukan utang.
export type OnchainBillingKind = 'invoices' | 'receivables'

export interface OnchainConfig {
  enabled: boolean
  chainId: number
  receiver: string
  idrPerEth: number
  explorerUrl: string
  minConfirmations: number
  requestTtlMinutes: number
}

interface OnchainAmount {
  amountIdr: number
  amountWei: string
  amountEth: string
  idrPerEth: number
  receiver: string
  chainId: number
  paymentUri: string
  status: OnchainStatus
  txHash: string | null
  txUrl: string | null
  payerAddress: string | null
  failureReason: string
  expiresAt: string
  paidAt: string | null
  createdAt: string
}

export interface OnchainRequest extends OnchainAmount {
  id: string
  targetKind: OnchainTargetKind
  targetId: string
  token: string
  createdBy: string | null
  createdByName: string | null
}

// Data halaman bayar publik. Tidak memuat nama pelanggan atau id internal.
export interface PublicPayment extends OnchainAmount {
  targetKind: OnchainTargetKind
  targetNumber: string
  projectName: string | null
  explorerUrl: string
}

export interface ConfirmOutcome extends PublicPayment {
  verification: OnchainVerification
  message: string
  confirmations: number
}

export interface OnchainRequestCreate {
  amount?: number
}

export const SEPOLIA_CHAIN_ID = 11155111
export const ONCHAIN_LIST_PAGE_SIZE = 5

export const ONCHAIN_STATUS_LABEL: Record<OnchainStatus, string> = {
  pending: 'Menunggu pembayaran',
  paid: 'Lunas',
  expired: 'Kedaluwarsa',
  failed: 'Gagal dicatat',
  cancelled: 'Dibatalkan',
}

export const ONCHAIN_STATUS_TONE: Record<OnchainStatus, string> = {
  pending: 'bg-amber-100 text-amber-800',
  paid: 'bg-green-100 text-green-800',
  expired: 'bg-slate-100 text-slate-700',
  failed: 'bg-red-100 text-red-800',
  cancelled: 'bg-slate-100 text-slate-700',
}

export const ONCHAIN_TARGET_LABEL: Record<OnchainTargetKind, string> = {
  invoice: 'faktur',
  receivable: 'piutang',
}

export const TARGET_KIND_OF: Record<OnchainBillingKind, OnchainTargetKind> = {
  invoices: 'invoice',
  receivables: 'receivable',
}

const TX_HASH_PATTERN = /^0x[0-9a-fA-F]{64}$/
const WEI_PER_ETHER = 10n ** 18n

export function networkName(chainId: number): string {
  return chainId === SEPOLIA_CHAIN_ID ? 'Sepolia' : `chain ${chainId}`
}

export function currencyName(chainId: number): string {
  return chainId === SEPOLIA_CHAIN_ID ? 'SepoliaETH' : 'ETH'
}

// Hash yang ditempel pembeli boleh berspasi atau berhuruf besar. Hasilnya null
// kalau bukan 0x diikuti 64 karakter heksadesimal.
export function normalizeTxHash(raw: string): string | null {
  const trimmed = raw.trim()
  return TX_HASH_PATTERN.test(trimmed) ? trimmed.toLowerCase() : null
}

export function payLink(origin: string, token: string): string {
  return `${origin}/bayar/${token}`
}

// Sama dengan ethereum.WeiForAmount di backend: dibulatkan ke atas supaya
// nilai yang dibayar tidak pernah kurang dari tagihannya.
export function weiForRupiah(amount: number, idrPerEth: number): bigint {
  const rate = BigInt(idrPerEth)
  return (BigInt(Math.trunc(amount)) * WEI_PER_ETHER + rate - 1n) / rate
}

export function formatEther(wei: bigint): string {
  const whole = wei / WEI_PER_ETHER
  const fraction = (wei % WEI_PER_ETHER).toString().padStart(18, '0').replace(/0+$/, '')
  return fraction === '' ? whole.toString() : `${whole}.${fraction}`
}

// Angka ETH ditulis dengan koma desimal seperti angka lain di aplikasi.
export function etherLabel(amountEth: string, chainId: number): string {
  return `${amountEth.replace('.', ',')} ${currencyName(chainId)}`
}

export function rateLabel(idrPerEth: number, chainId: number): string {
  return `Kurs simulasi: 1 ${currencyName(chainId)} = ${rupiah(idrPerEth)}`
}

export function toHexQuantity(wei: string): string {
  return `0x${BigInt(wei).toString(16)}`
}

export function shortHex(value: string): string {
  return value.length > 14 ? `${value.slice(0, 8)}…${value.slice(-6)}` : value
}

export function isOpenRequest(request: Pick<OnchainRequest, 'status'>): boolean {
  return request.status === 'pending'
}

export function timeLabel(iso: string): string {
  return new Date(iso).toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}
