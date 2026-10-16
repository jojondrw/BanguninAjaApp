# Sequence Diagram

Lima alur yang paling menentukan perilaku sistem. Nama endpoint, fungsi, dan
kode galat di diagram sama dengan yang ada di kode, sehingga diagram ini bisa
dipakai untuk menelusuri kode.

| No | Alur | Kode utama |
|---|---|---|
| 1 | Masuk dan memperbarui sesi | `auth/service.go`, `shared/cookie`, `frontend/src/shared/apiClient.ts` |
| 2 | Pemeriksaan hak akses setiap request | `shared/middleware/authorization.go`, `shared/access/matrix.go` |
| 3 | Evaluasi lokasi | `site/service.go`, `scoring/banguninaja_scoring` |
| 4 | Pembayaran on-chain | `billing/onchain.go`, `billing/onchain_verify.go`, `views/pay` |
| 5 | Penerimaan barang | `procurement/service.go` (`receiveGoods`) |

## 1. Masuk dan memperbarui sesi

Access token (JWT, 15 menit) hanya disimpan di memori frontend (store Zustand
tanpa `persist`), tidak di `localStorage`, sehingga tidak tertinggal di peramban.
Saat halaman dimuat ulang, sesi dipulihkan lewat `/auth/refresh`. Refresh token (7 hari)
dikirim sebagai cookie `HttpOnly` dan hanya hash SHA-256-nya yang disimpan di
database.

```mermaid
sequenceDiagram
    autonumber
    actor U as Pengguna
    participant FE as Frontend (React)
    participant API as Backend Go
    participant DB as PostgreSQL

    U->>FE: Isi email dan kata sandi
    FE->>API: POST /api/auth/login
    API->>DB: Cari user berdasarkan email
    API->>API: Bandingkan bcrypt hash
    alt Salah atau akun nonaktif
        API-->>FE: 401 invalid_credential / account_inactive
    else Benar
        API->>DB: Simpan hash refresh token baru
        API-->>FE: 200 accessToken + user, Set-Cookie refresh (HttpOnly)
        FE->>FE: Simpan access token di memori
    end

    Note over FE,API: 15 menit kemudian access token kedaluwarsa
    FE->>API: GET /api/projects (Bearer token lama)
    API-->>FE: 401
    FE->>API: POST /api/auth/refresh (cookie dikirim otomatis)
    API->>DB: Cari refresh token berdasarkan hash
    alt Token sudah pernah dipakai (dicabut)
        API->>DB: Cabut semua refresh token milik user
        API-->>FE: 401, pengguna harus masuk lagi
    else Token masih berlaku
        API->>DB: Cabut token lama, simpan token baru (rotasi)
        API-->>FE: 200 accessToken baru + cookie baru
        FE->>API: Ulangi GET /api/projects
        API-->>FE: 200
    end
```

Kalau beberapa request gagal 401 bersamaan, frontend hanya mengirim satu
`/auth/refresh` (`pendingRefresh` di `apiClient.ts`) dan semua request menunggu
hasil yang sama. Tanpa ini, rotasi token akan menganggap request kedua sebagai
pemakaian ulang dan mengeluarkan pengguna.

## 2. Pemeriksaan hak akses setiap request

Hak akses tidak dibaca dari isi JWT. Peran dan status aktif diambil dari
database di setiap request, jadi perubahan peran atau penonaktifan oleh Admin
langsung berlaku tanpa menunggu token kedaluwarsa.

```mermaid
sequenceDiagram
    autonumber
    participant FE as Frontend
    participant MW as Middleware Authorization
    participant AC as access.Matrix
    participant DB as PostgreSQL
    participant H as Controller slice

    FE->>MW: PATCH /api/procurement/purchase-orders/:id/status
    MW->>AC: ModuleForPath("/procurement/...")
    AC-->>MW: modul procurement
    MW->>MW: Verifikasi tanda tangan dan masa berlaku JWT
    alt Token tidak sah
        MW-->>FE: 401
    end
    MW->>DB: Ambil users.active dan role.code
    alt Akun nonaktif
        MW-->>FE: 401 account_inactive
    end
    MW->>AC: LevelForMethod(PATCH) = tulis
    MW->>AC: Allows(peran, procurement, tulis)
    alt Peran tidak punya hak tulis
        MW-->>FE: 403 forbidden
    else Diizinkan
        MW->>H: Lanjutkan ke handler
        H-->>FE: 200
    end
```

## 3. Evaluasi lokasi

Skor prediktif dan konteks deskriptif sengaja dipisah. Kalau regulasi atau
berita gagal diambil, evaluasi tetap berhasil dan bagian itu kosong.

Kalau layanan penilaian tidak bisa dihubungi (timeout, koneksi ditolak, galat
5xx), backend tidak gagal tetapi memakai **skor cadangan** dari
`stubScoreClient`: nilai semu 55 sampai 95 yang diturunkan dari hash koordinat,
dengan penjelasan berlabel "data simulasi". Galat `scoring_unavailable` (502)
hanya muncul kalau profil bangunan tidak dikenal.

```mermaid
sequenceDiagram
    autonumber
    actor PM as Manajer Proyek
    participant FE as Frontend (peta)
    participant API as Backend Go (site)
    participant DB as PostgreSQL + PostGIS
    participant SC as Layanan penilaian (Python)
    participant GN as Google News RSS

    PM->>FE: Klik titik di peta, pilih profil bangunan
    FE->>API: POST /api/site/evaluate {lat, lon, building_profile_id, project_id?}
    API->>DB: Cari building_profile, pastikan project ada
    API->>SC: POST /score {latitude, longitude, building_profile_code}
    alt Layanan menjawab
        SC->>SC: Baca bundle GIS, hitung 5 dimensi, kalikan bobot profil
        SC-->>API: overall_score, dimension_scores, risk_flags, region
    else Profil tidak dikenal
        SC-->>API: 4xx unknown_building_profile
        API-->>FE: 502 scoring_unavailable
    else Timeout, koneksi ditolak, atau 5xx
        API->>API: Skor cadangan (stubScoreClient), berlabel data simulasi
    end
    API->>DB: Transaksi: simpan saved_location + 5 dimension_score
    API->>DB: Cari regulasi terdekat (ST_DWithin pada titik)
    opt Ada kata kunci wilayah
        API->>GN: Cari berita kecamatan atau kabupaten
        GN-->>API: Daftar artikel
    end
    API-->>FE: 200 {saved_location_id, predictive, descriptive}
    FE-->>PM: Skor, rincian dimensi, risiko, regulasi, berita
```

## 4. Pembayaran on-chain

Pembeli membayar ke alamat dompet perusahaan di jaringan uji Ethereum Sepolia,
lalu backend memverifikasi transaksinya sendiri lewat node RPC. Backend tidak
pernah memegang kunci privat; ia hanya membaca blockchain.

```mermaid
sequenceDiagram
    autonumber
    actor K as Staf Keuangan
    actor P as Pembeli
    participant FE as Frontend
    participant API as Backend Go (billing)
    participant DB as PostgreSQL
    participant W as Dompet Pembeli
    participant RPC as Node RPC Sepolia

    K->>FE: Buat permintaan bayar untuk invoice
    FE->>API: POST /api/billing/invoices/:id/onchain-requests
    API->>API: Konversi rupiah ke wei dengan kurs tetap
    API->>DB: Simpan onchain_payment_request (pending, token acak, kedaluwarsa)
    API-->>FE: Tautan /bayar/:token
    K-->>P: Kirim tautan

    P->>FE: Buka /bayar/:token (tanpa login)
    FE->>API: GET /api/pay/:token
    API-->>FE: Jumlah ETH, alamat tujuan, chain, batas waktu
    FE-->>P: QR EIP-681 dan tombol bayar
    P->>W: Setujui transfer
    W->>RPC: Kirim transaksi
    W-->>FE: Hash transaksi
    FE->>API: POST /api/pay/:token/confirm {tx_hash}
    API->>API: Pembatas laju per token
    API->>DB: Pastikan hash belum dipakai permintaan lain
    API->>RPC: Ambil transaksi, chain id, blok, konfirmasi
    alt Penerima, jumlah, atau chain tidak cocok
        API-->>FE: 422 onchain_wrong_receiver, onchain_tx_reverted, dll.
    else Belum cukup konfirmasi
        API-->>FE: 200 waiting
        FE->>API: Kirim ulang konfirmasi berkala selama waiting
    else Terverifikasi
        API->>DB: Transaksi: kunci permintaan dan invoice
        API->>DB: Simpan billing_payment (method onchain, reference = hash)
        API->>DB: Tambah paid_amount, perbarui status invoice
        API->>DB: Tandai permintaan paid
        API-->>FE: 200 paid
        FE-->>P: Lunas, tautan ke Etherscan
    end
```

Pemeriksaan yang mencegah pembayaran palsu atau ganda:

| Pemeriksaan | Mencegah |
|---|---|
| `tx_hash` unik di seluruh permintaan | Satu transaksi dipakai melunasi dua tagihan |
| Penerima dan nilai harus sama dengan permintaan | Transfer ke alamat lain atau kurang bayar dianggap lunas |
| Chain id node harus 11155111 | Transaksi dari jaringan lain |
| Blok transaksi harus sebelum `expires_at` | Membayar dengan kurs yang sudah kedaluwarsa |
| Baris permintaan dikunci (`SELECT ... FOR UPDATE`) | Dua konfirmasi bersamaan mencatat pembayaran dua kali |

## 5. Penerimaan barang

Penerimaan barang mengubah tiga hal sekaligus dalam satu transaksi database:
dokumen penerimaan, stok gudang, dan status pesanan.

```mermaid
sequenceDiagram
    autonumber
    actor PG as Staf Pengadaan
    participant FE as Frontend
    participant API as Backend Go (procurement)
    participant INV as Inventaris (StockLedger)
    participant DB as PostgreSQL

    PG->>FE: Isi jumlah diterima dan ditolak per barang
    FE->>API: POST /api/procurement/goods-receipts
    API->>DB: BEGIN, kunci purchase_order
    alt Status pesanan bukan sent atau partially_received
        API-->>FE: 409 purchase_order_not_receivable
    end
    API->>DB: Hitung jumlah yang sudah diterima sebelumnya
    alt Penerimaan melebihi jumlah dipesan
        API-->>FE: 422 goods_receipt_exceeds_order
    end
    API->>DB: Simpan goods_receipt dan item-itemnya
    loop Setiap barang yang diterima
        alt Satuan pesanan sama dengan satuan material
            API->>INV: RecordIncoming(material, gudang, jumlah)
            INV->>DB: Simpan stock_movement (in), tambah stock
            API->>DB: Hubungkan item dengan stock_movement
        else Satuan berbeda
            API->>DB: Simpan stock_skip_reason = unit_mismatch
        end
    end
    API->>DB: Status pesanan = partially_received atau completed
    API->>DB: COMMIT
    API-->>FE: 201 dokumen penerimaan
```
