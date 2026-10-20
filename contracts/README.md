# EvidenceRegistry

Kontrak kecil di Sepolia yang menyimpan akar Merkle tiap kumpulan bukti
BanguninAja. Yang disimpan hanya akar (32 byte), jumlah catatan, waktu, dan
dompet pengirimnya. Isi laporan, foto, dan pembayaran tetap di server.

| Fungsi | Siapa | Kegunaan |
|---|---|---|
| `anchor(bytes32 root, uint32 count)` | dompet yang diizinkan | Mencatat satu kumpulan bukti. Satu root hanya bisa dicatat sekali |
| `anchoredAt(bytes32 root)` | siapa saja | Waktu root dicatat (0 bila belum pernah) |
| `recordCount(root)`, `anchoredBy(root)`, `isAnchored(root)` | siapa saja | Rincian pencatatan |
| `setAnchorer(address, bool)` | pemilik | Menambah atau mencabut dompet yang boleh mencatat |
| `transferOwnership(address)` | pemilik | Memindahkan kepemilikan |

Setiap pencatatan memancarkan event `Anchored(root, recordCount, anchorer, anchoredAt)`
sehingga riwayatnya terbaca di tab Events Etherscan.

## Menyiapkan dompet server

1. Di MetaMask, tambah akun baru khusus server (jangan pakai akun pribadi).
2. Isi akun itu dengan SepoliaETH dari faucet, misalnya
   https://cloud.google.com/application/web3/faucet/ethereum/sepolia. Satu kali
   pencatatan memakai sekitar 50.000 gas.
3. Ekspor kunci privat akun server (Account details, Show private key) dan tulis
   sendiri ke `.env` backend sebagai `EVIDENCE_SIGNER_KEY`. Jangan kirim kunci ini
   lewat chat, email, atau commit.

## Deploy lewat Remix

1. Buka https://remix.ethereum.org, buat file `EvidenceRegistry.sol`, dan tempel
   isi kontrak ini.
2. Tab Solidity compiler: versi 0.8.24 atau lebih baru, lalu Compile.
3. Tab Deploy & run: Environment `Injected Provider - MetaMask`, pastikan MetaMask
   di jaringan Sepolia dan memakai akun pribadimu (akun ini menjadi pemilik).
4. Isi argumen `initialAnchorers` dengan alamat dompet server dalam bentuk array,
   misalnya `["0xAlamatDompetServer"]`, lalu Deploy dan setujui di MetaMask.
5. Salin alamat kontrak yang muncul di Remix.
6. Verifikasi kode di Etherscan (plugin Contract Verification di Remix, atau
   https://sepolia.etherscan.io/verifyContract dengan single file, compiler dan
   optimisasi yang sama dengan langkah 2), supaya kodenya bisa dibaca semua orang.

## Menghubungkan ke backend

Tambahkan ke `.env`:

```
EVIDENCE_REGISTRY_ADDRESS=0xAlamatKontrak
EVIDENCE_SIGNER_KEY=<kunci privat dompet server, diisi sendiri>
EVIDENCE_AUTO_ANCHOR_INTERVAL=24h
```

Lalu buat ulang container backend (bersama override situs publik) supaya env
baru terbaca. Tanpa `EVIDENCE_AUTO_ANCHOR_INTERVAL`, pencatatan lewat server
hanya terjadi saat tombol "Catat bukti ke blockchain" ditekan.
