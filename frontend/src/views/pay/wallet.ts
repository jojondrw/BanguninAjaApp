import { SEPOLIA_CHAIN_ID, toHexQuantity } from '../../models/onchain'

// Dompet di peramban (MetaMask dan sejenisnya) menyuntikkan window.ethereum
// sesuai EIP-1193. Aplikasi hanya meminta dompet mengirim transaksi; kunci
// privat tidak pernah keluar dari dompet pembeli.
export interface Eip1193Provider {
  request: (args: { method: string; params?: unknown[] }) => Promise<unknown>
}

interface WalletTransfer {
  receiver: string
  chainId: number
  amountWei: string
}

const USER_REJECTED = 4001
const REQUEST_PENDING = -32002
const UNRECOGNIZED_CHAIN = 4902

const SEPOLIA_NETWORK = {
  chainId: `0x${SEPOLIA_CHAIN_ID.toString(16)}`,
  chainName: 'Sepolia',
  nativeCurrency: { name: 'Sepolia Ether', symbol: 'SepoliaETH', decimals: 18 },
  rpcUrls: ['https://ethereum-sepolia-rpc.publicnode.com'],
  blockExplorerUrls: ['https://sepolia.etherscan.io'],
}

export function injectedWallet(): Eip1193Provider | null {
  if (typeof window === 'undefined') {
    return null
  }
  const candidate = (window as unknown as { ethereum?: Partial<Eip1193Provider> }).ethereum
  return candidate && typeof candidate.request === 'function' ? (candidate as Eip1193Provider) : null
}

// Tautan yang membuka halaman ini di peramban bawaan aplikasi MetaMask di HP.
export function metamaskDeepLink(pageUrl: string): string {
  return `https://metamask.app.link/dapp/${pageUrl.replace(/^https?:\/\//, '')}`
}

function errorCode(error: unknown): number | null {
  if (typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'number') {
    return error.code
  }
  return null
}

function errorText(error: unknown): string {
  if (typeof error === 'object' && error !== null && 'message' in error && typeof error.message === 'string') {
    return error.message
  }
  return ''
}

async function switchChain(wallet: Eip1193Provider, chainId: number) {
  const hexChainId = `0x${chainId.toString(16)}`
  try {
    await wallet.request({ method: 'wallet_switchEthereumChain', params: [{ chainId: hexChainId }] })
  } catch (error) {
    if (errorCode(error) !== UNRECOGNIZED_CHAIN || chainId !== SEPOLIA_CHAIN_ID) {
      throw error
    }
    await wallet.request({ method: 'wallet_addEthereumChain', params: [SEPOLIA_NETWORK] })
  }
}

// Meminta akun, memindah dompet ke jaringan yang benar, lalu mengirim nilai
// persis sebesar permintaan ke alamat penerima. Hasilnya hash transaksi.
export async function payWithWallet(wallet: Eip1193Provider, transfer: WalletTransfer): Promise<string> {
  const accounts = (await wallet.request({ method: 'eth_requestAccounts' })) as string[]
  const from = accounts[0]
  if (!from) {
    throw new Error('Dompet tidak memberikan akun. Buka kunci MetaMask lalu coba lagi.')
  }
  await switchChain(wallet, transfer.chainId)
  const hash = await wallet.request({
    method: 'eth_sendTransaction',
    params: [{ from, to: transfer.receiver, value: toHexQuantity(transfer.amountWei) }],
  })
  if (typeof hash !== 'string') {
    throw new Error('Dompet tidak mengembalikan hash transaksi.')
  }
  return hash
}

async function connect(wallet: Eip1193Provider, chainId: number): Promise<string> {
  const accounts = (await wallet.request({ method: 'eth_requestAccounts' })) as string[]
  const from = accounts[0]
  if (!from) {
    throw new Error('Dompet tidak memberikan akun. Buka kunci MetaMask lalu coba lagi.')
  }
  await switchChain(wallet, chainId)
  return from
}

// Mencatat sidik jari bukti: transaksi 0 ETH ke alamat sendiri yang membawa
// data (akar Merkle). Tidak ada dana yang berpindah, hanya biaya gas jaringan uji.
export async function sendCalldata(wallet: Eip1193Provider, chainId: number, data: string, to?: string): Promise<string> {
  const from = await connect(wallet, chainId)
  const hash = await wallet.request({ method: 'eth_sendTransaction', params: [{ from, to: to || from, value: '0x0', data }] })
  if (typeof hash !== 'string') {
    throw new Error('Dompet tidak mengembalikan hash transaksi.')
  }
  return hash
}

// Tanda tangan EIP-712: gratis, tanpa transaksi. Dompet menampilkan isi
// persetujuan apa adanya sebelum ditandatangani.
export async function signTypedData(wallet: Eip1193Provider, typedData: { domain: { chainId: number } }): Promise<string> {
  const from = await connect(wallet, typedData.domain.chainId)
  const signature = await wallet.request({ method: 'eth_signTypedData_v4', params: [from, JSON.stringify(typedData)] })
  if (typeof signature !== 'string') {
    throw new Error('Dompet tidak mengembalikan tanda tangan.')
  }
  return signature
}

export function walletErrorMessage(error: unknown): string {
  const code = errorCode(error)
  const text = errorText(error)
  if (code === USER_REJECTED) {
    return 'Permintaan dibatalkan di dompet. Tidak ada dana yang dikirim.'
  }
  if (code === REQUEST_PENDING) {
    return 'MetaMask masih membuka permintaan lain. Buka MetaMask untuk menyelesaikannya dulu.'
  }
  if (code === UNRECOGNIZED_CHAIN) {
    return 'Jaringan uji ini belum ada di dompet. Tambahkan jaringannya di MetaMask lalu coba lagi.'
  }
  if (/insufficient funds/i.test(text)) {
    return 'Saldo SepoliaETH di dompet tidak cukup. Ambil SepoliaETH gratis dari faucet dulu.'
  }
  return text ? `Dompet menolak transaksi: ${text}` : 'Dompet menolak transaksi. Coba lagi.'
}
