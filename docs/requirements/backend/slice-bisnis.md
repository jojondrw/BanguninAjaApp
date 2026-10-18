# Slice Bisnis

Dokumen ini mencatat pola yang dipakai slice bisnis (`master`, `project`,
`inventory`, dan slice berikutnya), daftar endpoint, serta aturan bisnis yang
dijaga di service. Aturan dasar arsitektur tetap mengacu ke `arsitektur.md`.

## Pola yang dipakai ulang

| Kebutuhan | Tempat | Cara pakai |
|---|---|---|
| Paginasi | `shared/pagination` | DTO query menanam `pagination.Query` (`page`, `pageSize`, bawaan 20, maksimal 100). Service membalas `pagination.Page[T]` |
| Daftar berhalaman | `database.FindPage` | Repository mengisi `database.Listing` berisi filter, urutan, offset, limit. Hitung total dan ambil data memakai dua query terpisah dari filter yang sama |
| Pencarian teks | `database.ContainsPattern` | Membungkus kata dengan `%` dan meloloskan `%`, `_`, `\` supaya masukan pengguna tidak dibaca sebagai wildcard |
| Error database | `database.Translate` | Repository selalu membungkus error GORM dengan ini, sehingga service hanya melihat `ErrNotFound`, `ErrDuplicate`, `ErrReferenceViolated`, `ErrCheckViolated` |
| Error untuk pengguna | `database.ErrorMap` | Service memetakan error repository ke `apperror` milik slice. Error yang tidak dipetakan menjadi 500 |
| Binding request | `shared/httprequest` | `BindJSON`, `BindQuery`, `PathID`. Membalas `invalid_payload`, `invalid_query`, atau `invalid_id` dengan status 400. `UserID` mengambil id pengguna dari access token untuk data yang dicatat atas nama pengguna |
| Menulis response | `httpresponse.Respond`, `RespondCreated`, `RespondNoContent` | Controller cukup meneruskan hasil dan error dari service |
| Status jatuh tempo | `shared/duedate` | Dipakai `sales` dan `billing`. `Status` menghitung `not_due`, `due`, `overdue`, `paid` dari tanggal jatuh tempo. `RangeOf` menerjemahkan filter status menjadi rentang tanggal untuk repository |

### Kenapa RESTRICT perlu ditangani sendiri

Foreign key yang dibuat dengan `ON DELETE RESTRICT` membalas SQLSTATE `23001`,
bukan `23503`. Penerjemah error bawaan GORM hanya mengenal `23503`, jadi tanpa
penanganan tambahan, menghapus data yang masih dipakai berakhir sebagai 500.
`database.Translate` memetakan `23001` ke `ErrReferenceViolated` supaya service
bisa membalas 409.

### Transaksi

Service tidak boleh tahu GORM, jadi transaksi dibuka lewat method repository:

```go
Transaction(ctx context.Context, work func(Repository) error) error
```

Repository di dalam `work` sudah terikat ke transaksi yang sama. Aturan bisnis
tetap ditulis di service, repository hanya menyediakan batas transaksinya.

### Status HTTP

| Status | Dipakai saat |
|---|---|
| 400 | Format body, query, atau id salah |
| 404 | Data yang dituju lewat path tidak ada |
| 409 | Kode sudah dipakai, data masih dipakai data lain, atau data sudah terkunci karena statusnya |
| 422 | Format benar tapi melanggar aturan bisnis, termasuk data acuan di body yang tidak ditemukan |

Semua endpoint slice bisnis wajib memakai access token.

## Master

Semua di bawah `/api/master`.

| Method | Path |
|---|---|
| GET, POST | `/regions`, `/units-of-measure`, `/accounts` |
| GET, PUT, DELETE | `/regions/:id`, `/units-of-measure/:id`, `/accounts/:id` |

Filter daftar: `search` untuk semua, `type` dan `parentId` untuk wilayah dan akun.

| Aturan | Kode error |
|---|---|
| Provinsi tidak punya induk. Kota dan kabupaten di bawah provinsi. Kecamatan di bawah kota atau kabupaten | `region_parent_required`, `region_parent_invalid` |
| Jenis wilayah atau akun tidak bisa diubah selama masih punya turunan | `region_type_locked`, `account_type_locked` |
| Akun induk harus sejenis dengan akun anaknya | `account_parent_type_mismatch` |
| Akun induk tidak boleh akun itu sendiri atau salah satu turunannya. Diperiksa dengan satu recursive CTE, bukan query berulang | `account_parent_cycle` |

## Project

Semua di bawah `/api/projects`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/` | Filter `search` (nama atau kode), `status`, `regionId`, `mine=true` (hanya proyek yang timnya memuat pemanggil) |
| GET, PUT, DELETE | `/:id` | PUT tidak mengubah status dan progres |
| PATCH | `/:id/status` | Body `{ "status": "ongoing" }` |
| GET, POST | `/:id/phases` | Daftar tahap urut `sortOrder` |
| PUT, DELETE | `/:id/phases/:childId` | |
| GET, POST | `/:id/budget-items` | Berhalaman, ditambah `grandTotal` untuk seluruh RAB proyek |
| PUT, DELETE | `/:id/budget-items/:childId` | |
| GET | `/:id/progress-curve` | Kurva S rencana dan realisasi per minggu |
| GET, POST | `/:id/daily-reports` | Laporan harian lapangan, berhalaman, terbaru di atas |
| PUT, DELETE | `/:id/daily-reports/:childId` | |
| GET | `/:id/payment-schedule` | Kontrak kontraktor, semua termin dengan statusnya, dan ringkasan pembayaran |
| PUT | `/:id/construction-contract` | Buat atau ubah kontrak kontraktor. Body `{ "vendorId", "number", "value", "retentionPercent" }` |
| POST | `/:id/payment-terms` | Body `{ "requiredProgress", "percent" }` |
| DELETE | `/:id/payment-terms/:childId` | Hanya termin yang belum diajukan |
| POST | `/:id/payment-terms/:childId/claim` | Ajukan termin, membuat hutang di Tagihan |
| POST | `/:id/retention-release` | Lepas retensi yang ditahan, membuat hutang di Tagihan |
| GET | `/:id/owner-billing` | Kontrak dengan pemberi kerja, semua termin tagihannya, dan ringkasan penerimaan |
| PUT | `/:id/owner-billing/contract` | Buat atau ubah kontrak pemberi kerja. Body `{ "customerId", "number", "value", "retentionPercent", "paymentTermDays" }` |
| POST | `/:id/owner-billing/terms` | Body `{ "requiredProgress", "percent" }` |
| DELETE | `/:id/owner-billing/terms/:childId` | Hanya termin yang belum ditagih |
| POST | `/:id/owner-billing/terms/:childId/claim` | Tagih termin, membuat piutang di Tagihan |
| POST | `/:id/owner-billing/retention-release` | Tagih retensi yang ditahan pemberi kerja, membuat piutang di Tagihan |
| POST | `/:id/contract-changes` | Usulkan adendum. Body `{ "description", "valueChange", "daysChange" }`, `valueChange` negatif untuk pekerjaan kurang |
| PATCH | `/:id/contract-changes/:childId` | Putuskan adendum. Body `{ "status": "approved" }` atau `rejected`, hanya Admin |
| DELETE | `/:id/contract-changes/:childId` | Hapus usulan yang belum diputuskan |
| GET, POST | `/:id/permits` | |
| PUT, DELETE | `/:id/permits/:childId` | |
| GET, POST | `/:id/members` | Tim proyek, urut nama. Body tambah `{ "userId", "position"?, "role"? }` |
| PATCH, DELETE | `/:id/members/:childId` | Body ubah `{ "position"?, "role"? }`, field yang tidak dikirim tidak berubah |
| GET | `/:id/member-candidates` | Pengguna aktif yang belum masuk tim, hanya `id`, `name`, `email`. Query `search`, `page`, `pageSize` |

### Status proyek

| Dari | Boleh ke |
|---|---|
| `planning` | `ongoing`, `cancelled` |
| `ongoing` | `on_hold`, `completed`, `cancelled` |
| `on_hold` | `ongoing`, `cancelled` |
| `completed`, `cancelled` | tidak ada, keduanya status akhir |

Proyek baru selalu mulai dari `planning`. Mengirim status yang sama dengan
status sekarang tidak dianggap salah dan tidak mengubah apa pun.

### Progres diturunkan dari tahapan

Setiap tahap atau item RAB ditambah, diubah, atau dihapus, progres proyek
dihitung ulang sebagai rata-rata progres tahap yang **dibobot nilai RAB** tiap
tahap: bobot tahap = total item RAB yang ditautkan ke tahap itu dibagi total
RAB yang tertaut. Kalau belum ada item RAB yang ditautkan, semua tahap berbobot
sama. Hasilnya dibulatkan dan disimpan juga sebagai riwayat harian
(`project_progress`) untuk kurva S. Proses ini berjalan dalam satu transaksi dan
baris proyek dikunci dengan `FOR UPDATE`, supaya dua perubahan yang bersamaan
tidak saling menimpa hasil hitungan.

Status tahap juga diturunkan dari progresnya: 0 berarti `not_started`, 100
berarti `completed`, selain itu `in_progress`. Jadi status tahap tidak dikirim
dari frontend.

| Aturan | Kode error |
|---|---|
| Proyek hanya bisa `completed` kalau semua tahapnya sudah 100%. Saat selesai, progres dijadikan 100 | `project_unfinished` |
| Tahap dan RAB tidak bisa diubah kalau proyek sudah `completed` atau `cancelled` | `project_closed` |
| Tanggal selesai tidak boleh sebelum tanggal mulai, berlaku untuk proyek, tahap, dan izin | `invalid_date_range` |
| `total` item RAB selalu dihitung dari volume dikali harga satuan, dibulatkan ke rupiah penuh. Volume dibulatkan dua desimal mengikuti `numeric(14,2)` | |
| Total RAB proyek tidak disimpan, dihitung dengan `SUM` saat daftar diminta | |
| Izin berstatus `issued` wajib punya tanggal terbit | `permit_issued_date_required` |

### Termin kontraktor dan retensi

Developer membayar kontraktor pelaksana bertahap sesuai progres dan menahan
retensi sampai serah terima. Satu proyek punya satu kontrak kontraktor, dan
setiap termin berisi syarat progres berbobot serta persen nilai kontrak.

Mengajukan termin berarti mengesahkan progres. Dalam satu transaksi, sistem
membuat hutang di modul Tagihan sebesar nilai termin dikurangi retensi, dengan
jatuh tempo hari ini ditambah termin bayar vendor (`payment_term_days`). Hutang
dibuat lewat `billing.PayableLedger`, dipanggil melalui interface `PayableLedger`
yang dideklarasikan di slice `project` dan dirakit di `module.go`, sama seperti
pengadaan memakai ledger stok inventaris.

Pemisahan tugas: endpoint termin ikut modul akses `projects`, jadi yang
mengajukan adalah Manajer Proyek atau Admin. Membayar hutangnya tetap di modul
Tagihan yang butuh hak tulis `billing`, yaitu Keuangan atau Admin.

| Aturan | Kode error |
|---|---|
| Termin dan kontrak hanya disusun selama proyek belum `completed` atau `cancelled` | `project_closed` (409) |
| Termin butuh kontrak kontraktor | `construction_contract_not_found` (404) |
| Kontraktor harus vendor yang ada dan aktif | `vendor_not_found`, `vendor_inactive` (422) |
| Setelah ada termin diajukan, kontraktor, nilai, dan retensi tidak bisa diubah. Nomor SPK tetap boleh | `construction_contract_locked` (409) |
| Jumlah persen semua termin paling banyak 100 | `payment_term_percent_exceeded` (422) |
| Syarat progres termin baru tidak boleh di bawah termin yang sudah diajukan | `payment_term_before_claimed` (422) |
| Termin diajukan kalau progres berbobot sudah mencapai syarat | `payment_term_progress_not_reached` (422) |
| Termin diajukan berurutan: syarat progres terkecil dulu | `payment_term_out_of_order` (422) |
| Termin yang sudah diajukan tidak bisa diajukan ulang atau dihapus | `payment_term_claimed` (409) |
| Proyek `cancelled` tidak bisa mengajukan termin. Proyek `on_hold` dan `completed` masih bisa, karena progresnya sudah tercapai | `payment_term_project_cancelled` (409) |
| Retensi dilepas setelah proyek `completed` | `retention_not_due` (422) |
| Semua termin progres harus sudah diajukan sebelum retensi dilepas | `retention_terms_open` (422) |
| Retensi hanya dilepas sekali, dijaga juga oleh indeks unik parsial | `retention_released` (409) |

Nilai termin dihitung **kumulatif**: persen kumulatif sampai termin itu dikali
nilai kontrak berjalan, dikurangi nilai termin sebelumnya. Retensi = nilai
termin x persen retensi. Semua dihitung dengan bilangan bulat dan dibulatkan ke
rupiah penuh. Termin yang sudah diajukan memakai angka saat diajukan, jadi
setelah adendum disetujui, termin sisanya menyesuaikan sendiri dan totalnya tetap
sama dengan nilai kontrak berjalan. Kalau hutang termin yang belum dibayar dihapus di Tagihan, `payable_id`
menjadi NULL dan termin itu kembali bisa diajukan.

### Adendum (pekerjaan tambah-kurang)

Setelah ada termin diajukan, nilai kontrak kontraktor terkunci. Perubahan
nilai dan waktu sesudahnya lewat adendum: uraian dan alasan, selisih nilai
(negatif untuk pekerjaan kurang), dan perpanjangan hari. Nilai kontrak berjalan =
nilai awal + adendum yang disetujui.

Manajer Proyek mengusulkan. Menyetujui atau menolak hanya boleh Admin: aturan
ini diperiksa di service dengan peran yang sudah dibaca middleware otorisasi,
karena matriks akses hanya mengenal hak per modul.

| Aturan | Kode error |
|---|---|
| Adendum hanya diusulkan selama proyek belum `completed` atau `cancelled` | `project_closed` (409) |
| Adendum butuh kontrak kontraktor | `construction_contract_not_found` (404) |
| Adendum harus mengubah nilai atau waktu | `contract_change_empty` (422) |
| Hanya Admin yang menyetujui atau menolak | `contract_change_approver` (403) |
| Keputusan bersifat final; adendum yang sudah diputuskan tidak bisa diputuskan ulang atau dihapus | `contract_change_decided` (409) |
| Nilai kontrak berjalan harus tetap lebih dari nol | `contract_change_non_positive` (422) |
| Nilai kontrak berjalan tidak boleh di bawah nilai termin yang sudah diajukan | `contract_change_below_claimed` (422) |
| Proyek `cancelled` tidak bisa memutuskan adendum | `contract_change_project_cancelled` (409) |
| Kalau pengurangan membuat sisa nilai habis, termin berikutnya tidak bisa diajukan | `payment_term_nothing_to_claim` (422) |

Nomor adendum (`sequence`) unik per proyek dan diambil dari nomor terbesar
ditambah satu, di dalam transaksi yang mengunci baris proyek.

### Termin pemberi kerja

Sisi lain termin kontraktor: developer menagih pemberi kerja (misalnya yayasan
pemilik rumah sakit) sesuai progres berbobot proyek, dan setiap tagihan menjadi
piutang di Tagihan. Begitu piutang itu dibayar, kasnya masuk otomatis lewat
aturan "Pembayaran tercatat sebagai kas".

| Aturan | Isi |
|---|---|
| Kontrak | Satu per proyek (`owner_contract`, unik per `project_id`): pemberi kerja dari tabel `customer`, nomor, nilai, retensi 0 sampai 10%, tempo bayar 0 sampai 180 hari |
| Termin | Tabel `owner_term`, bentuknya sama dengan `payment_term` tetapi menaut `receivable_id`. Hitungan nilai kumulatif, urutan pengajuan, syarat progres, dan retensi memakai fungsi yang sama dengan termin kontraktor |
| Tagih | Membuat piutang ke pemberi kerja senilai termin dikurangi retensi, referensi "Tagihan termin N KODE", jatuh tempo tanggal tagih ditambah tempo bayar |
| Retensi | Ditagih sekali setelah proyek selesai dan semua termin progres sudah ditagih, referensi "Tagihan retensi KODE" |
| Kunci kontrak | Setelah ada termin yang ditagih, pemberi kerja, nilai, retensi, dan tempo tidak bisa diubah (409 `owner_contract_locked`); nomor kontrak tetap boleh diubah |
| Hak akses | Sama dengan termin kontraktor: butuh hak tulis Proyek dan keanggotaan tim proyek |

Modul Proyek tidak mengimpor Tagihan secara langsung untuk ini: antarmuka
`BillingLedger` (dulu `PayableLedger`) punya `RecordPayable` dan
`RecordReceivable`, dan adapternya ke `billing.PayableLedger` serta
`billing.ReceivableLedger` ada di `project/module.go`.

### Tim proyek

Tabel `project_member` milik slice `project` (sebelumnya entity-nya ada di
`auth`, nama tabel, index, dan constraint tidak berubah). Satu baris berarti
satu pengguna masuk tim satu proyek.

| Kolom | Arti |
|---|---|
| `role_id` | Peran pengguna di proyek ini, salah satu dari 7 peran aplikasi di tabel `role`. Kalau tidak dipilih saat ditambahkan, diisi peran akun pengguna saat itu |
| `position` | Posisi bebas di proyek, misalnya "Site manager" atau "Estimator". Boleh kosong |
| `created_at` | Tanggal masuk tim, dikirim sebagai `joinedAt` |

Semua endpoint tim ikut modul akses `projects`: GET butuh hak baca, selain itu
butuh hak tulis. `GET /api/users` tetap khusus admin, jadi pencarian pengguna
untuk formulir tambah anggota memakai `member-candidates`.

| Aturan | Kode error |
|---|---|
| Proyek harus ada | `project_not_found` (404) |
| Anggota yang diubah atau dihapus harus ada di proyek itu | `project_member_not_found` (404) |
| Satu pengguna hanya sekali di satu tim, dijaga unique `(user_id, project_id)` | `project_member_exists` (409) |
| Pengguna harus ada dan aktif | `member_user_not_found`, `member_user_inactive` (422) |
| Kode peran harus ada di tabel `role`. Pengguna tanpa peran akun wajib dipilihkan peran | `member_role_not_found`, `member_role_required` (422) |

### Hak akses per proyek

Admin, Pengamat, dan Keuangan melihat semua proyek. Peran lain (Manajer Proyek,
Pengadaan, Penjualan, SDM) hanya melihat proyek yang timnya memuat mereka:

| Bagian | Aturan |
|---|---|
| `GET /` dan `GET /summary` | Daftar dan ringkasan otomatis disaring ke proyek tim pemanggil. `mine=true` tetap tersedia sebagai penyaring untuk tiga peran yang melihat semua |
| Semua rute `/:id/...` | Diperiksa middleware `RequireProjectAccess` sebelum handler. Bukan anggota mendapat `project_member_required` (403) |
| Peran tidak diketahui | Ditolak dengan `project_viewer_unknown` (403), bukan diloloskan |
| `POST /` | Pembuat proyek otomatis masuk tim dengan perannya, supaya Manajer Proyek bisa membuka proyek yang baru ia buat |

Aturan ini hanya berlaku di modul Proyek. Modul lain yang menyaring per proyek
(keuangan, tagihan, pengadaan, laporan) masih mengikuti hak akses per modul.

## Inventory

Semua di bawah `/api/inventory`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/materials` | Filter `search`, `category` |
| GET | `/materials/low-stock` | Material dengan total stok semua gudang di bawah `minimumStock` |
| GET, PUT, DELETE | `/materials/:id` | |
| GET, POST | `/warehouses` | Filter `search`, `projectId` |
| GET, PUT, DELETE | `/warehouses/:id` | |
| GET | `/stocks` | Filter `materialId`, `warehouseId`. Sudah berisi kode dan nama material, kode satuan, serta gudang |
| GET, POST | `/stock-movements` | Filter `materialId`, `warehouseId`, `type`, `dateFrom`, `dateTo` dengan format `YYYY-MM-DD`. Sudah berisi kode dan nama material, kode satuan, serta kode dan nama gudang asal dan tujuan |
| GET | `/stock-movements/:id` | |

### Stok hanya berubah lewat mutasi

Tidak ada endpoint untuk mengubah angka stok secara langsung. Setiap perubahan
harus tercatat sebagai mutasi, dan mutasi tidak bisa diubah maupun dihapus,
sama seperti buku besar. Salah catat diperbaiki dengan mutasi `adjustment`.

Mencatat mutasi, mengurangi stok gudang asal, dan menambah stok gudang tujuan
berjalan dalam satu transaksi. Pengurangan memakai
`UPDATE ... WHERE quantity >= ?` sehingga pengecekan dan pengurangan terjadi
dalam satu langkah, tanpa celah bagi permintaan lain untuk menyela. Penambahan
memakai `INSERT ... ON CONFLICT` pada `uq_stock_material_warehouse`, jadi baris
stok dibuat otomatis saat material pertama kali masuk ke sebuah gudang.

| Jenis | Gudang asal | Gudang tujuan |
|---|---|---|
| `in` | kosong | wajib |
| `out` | wajib | kosong |
| `transfer` | wajib | wajib, dan tidak sama dengan asal |
| `adjustment` | salah satu saja: asal untuk mengurangi, tujuan untuk menambah | |

| Aturan | Kode error |
|---|---|
| Arah gudang tidak sesuai jenis mutasi | `stock_movement_direction_invalid` |
| Transfer ke gudang yang sama | `stock_movement_same_warehouse` |
| Stok gudang asal kurang | `insufficient_stock` |
| Material atau gudang tidak ada | `stock_movement_reference_not_found` |

## Procurement

Semua di bawah `/api/procurement`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/vendors` | Filter `search`, `category`, `rating`, `active` |
| GET, PUT, DELETE | `/vendors/:id` | |
| GET, POST | `/purchase-requests` | Filter `search` (nomor), `projectId`, `status`, `dateFrom`, `dateTo`. Pemohon diambil dari access token |
| GET, PUT, DELETE | `/purchase-requests/:id` | GET berisi daftar item. PUT mengganti seluruh item |
| PATCH | `/purchase-requests/:id/status` | `submitted`, `approved`, `rejected`, `completed` |
| GET, POST | `/purchase-orders` | Filter `search`, `vendorId`, `projectId`, `status`, `dateFrom`, `dateTo` |
| GET, PUT, DELETE | `/purchase-orders/:id` | GET berisi item, jumlah yang sudah diterima, dan sisanya |
| PATCH | `/purchase-orders/:id/status` | Hanya `sent` atau `cancelled` |
| GET, POST | `/goods-receipts` | Filter `search`, `purchaseOrderId`, `warehouseId`, `dateFrom`, `dateTo`. POST membalas rincian yang sama dengan GET `/:id` |
| GET | `/goods-receipts/:id` | Berisi baris beserta material, satuan pesanan, satuan stok, dan status stok per baris, ditambah `unpostedLines` |

| Dokumen | Dari | Boleh ke |
|---|---|---|
| Permintaan | `draft` | `submitted` |
| | `submitted` | `approved`, `rejected` |
| | `approved` | `completed` |
| Pesanan | `draft` | `sent`, `cancelled` |
| | `sent` | `cancelled` |

`partially_received` dan `completed` pada pesanan tidak bisa dikirim dari
frontend. Keduanya diatur oleh penerimaan barang: setelah penerimaan dicatat,
pesanan menjadi `completed` kalau semua baris sudah diterima penuh, selain itu
`partially_received`. Pesanan dikunci dengan `FOR UPDATE` selama penerimaan
dicatat supaya dua penerimaan yang bersamaan tidak melewati jumlah pesanan.

Penerimaan barang tidak bisa diubah maupun dihapus, sama seperti mutasi stok.

### Penerimaan langsung menambah stok

Dalam transaksi yang sama dengan penyimpanan penerimaan, setiap baris dengan
jumlah diterima lebih dari nol dicatat sebagai satu mutasi `in` ke gudang
penerimaan, bertanggal tanggal penerimaan, dengan `reference` berisi nomor
penerimaan. Jumlah yang ditolak tidak masuk stok. Mutasi dan penambahan stok
memakai logika yang sama dengan `POST /api/inventory/stock-movements`
(`INSERT ... ON CONFLICT` pada stok), jadi hasilnya tidak bisa dibedakan dari
mutasi masuk biasa.

`procurement` tidak memanggil repository `inventory` secara langsung. Service
`procurement` mendeklarasikan interface kecil `StockLedger` dengan satu method
`RecordIncoming`. Repository `procurement` membuat ledger itu dari koneksi yang
sedang dipakai, jadi di dalam `Transaction` ledger ikut terikat ke transaksi yang
sama. Implementasinya, `inventory.StockLedger`, dirakit di `module.go` milik
`procurement`, satu-satunya berkas `procurement` yang mengimpor `inventory`.
`inventory` tidak tahu apa pun soal `procurement`, jadi tidak ada impor
melingkar.

Stok dihitung dalam satuan material (`material.unit_of_measure_id`), sedangkan
baris pesanan punya satuannya sendiri. Tabel `unit_of_measure` tidak menyimpan
konversi, jadi baris yang satuannya berbeda tidak ditebak: baris itu dilewati,
`stock_skip_reason` diisi `unit_mismatch`, dan dilaporkan di `unpostedLines`
supaya pengguna mencatatnya manual setelah jumlahnya dikonversi. Kalau mutasi
untuk baris yang satuannya cocok gagal dicatat, seluruh penerimaan dibatalkan
(rollback), termasuk perubahan status pesanan.

Setiap baris penerimaan menyimpan `stock_movement_id` (unik, `RESTRICT` ke
`stock_movement`) untuk mutasi yang dibuat otomatis. Status stok per baris di
respons diturunkan dari kolom itu:

| `stockStatus` | Arti |
|---|---|
| `posted` | Mutasi masuk dibuat otomatis, `stockMovementId` terisi, `stockPosted` bernilai `true` |
| `skipped` | Tidak dicatat otomatis, alasan di `unpostedReason` dan baris ini muncul di `unpostedLines` (`lineId`, `materialId`, `materialName`, `reason`) |
| `legacy` | Penerimaan dicatat sebelum fitur ini ada. Stok masuknya mungkin sudah dicatat manual, jadi tidak diisi ulang otomatis supaya tidak tercatat dua kali |
| `none` | Tidak ada jumlah diterima, tidak ada yang perlu masuk stok |

Untuk baris `legacy` dan `skipped`, frontend masih mencari mutasi manual dengan
`reference` nomor penerimaan dan menyediakan tombol "Catat stok masuk".

| Aturan | Kode error |
|---|---|
| Permintaan dan pesanan hanya bisa diubah atau dihapus selama `draft` | `purchase_request_locked`, `purchase_order_locked` |
| Satu material hanya sekali dalam satu dokumen | `material_duplicate` |
| Nilai pesanan dihitung dari jumlah kali harga satuan per baris, dibulatkan ke rupiah penuh. Jumlah dibulatkan dua desimal | |
| Vendor pesanan harus aktif | `vendor_inactive` |
| Pesanan yang mengacu permintaan butuh permintaan `approved` dengan proyek yang sama | `purchase_request_not_approved`, `purchase_request_project_mismatch` |
| Barang hanya diterima untuk pesanan `sent` atau `partially_received` | `purchase_order_not_receivable` |
| Baris penerimaan harus milik pesanan itu, tidak ganda, dan punya jumlah diterima atau ditolak | `goods_receipt_item_invalid`, `goods_receipt_item_duplicate`, `goods_receipt_quantity_invalid` |
| Total diterima per baris tidak boleh melebihi jumlah pesanan | `goods_receipt_exceeds_order` |
| Stok gagal ditambah untuk baris yang satuannya cocok. Seluruh penerimaan dibatalkan, pesan menyebut material dan sebabnya. Status mengikuti sebabnya (422 untuk aturan stok, 500 untuk kegagalan database) | `goods_receipt_stock_failed` |

## Asset

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/api/assets` | Filter `search`, `category`, `status`. Daftar ditambah total nilai perolehan, akumulasi penyusutan, dan nilai buku |
| GET, PUT, DELETE | `/api/assets/:id` | |
| GET, POST | `/api/equipment` | Filter `search`, `status`, `projectId`, `assetId`, `serviceDueBefore` (`YYYY-MM-DD`) |
| GET, PUT, DELETE | `/api/equipment/:id` | |

Nilai buku dan penyusutan per bulan (garis lurus, nilai perolehan dibagi umur
manfaat) dihitung saat dibaca, tidak disimpan.

| Aturan | Kode error |
|---|---|
| Akumulasi penyusutan tidak boleh melebihi nilai perolehan | `asset_depreciation_exceeds_cost` |
| Alat berstatus `operating` wajib ditugaskan ke proyek | `equipment_project_required` |
| Alat tidak boleh dikaitkan ke aset yang sudah `disposed` | `equipment_asset_disposed` |

## Sales

Semua di bawah `/api/sales`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/customers` | Filter `search` (nama, email, nomor identitas) |
| GET, PUT, DELETE | `/customers/:id` | |
| GET, POST | `/units` | Filter `search`, `projectId`, `status` |
| GET | `/units/summary` | Jumlah unit per status, bisa disaring `projectId` |
| GET, PUT, DELETE | `/units/:id` | |
| GET, POST | `/leads` | Filter `search`, `projectId`, `stage` |
| GET, PUT, DELETE | `/leads/:id` | |
| POST | `/leads/:id/convert` | Body sama dengan pelanggan baru. Membuat pelanggan, menautkannya ke prospek (`customerId`), dan menjadikan tahap `won` dalam satu transaksi |
| GET, POST | `/contracts` | Filter `search`, `customerId`, `unitId`, `status`, `type`. Sudah berisi nama pelanggan dan kode unit |
| GET, PUT, DELETE | `/contracts/:id` | |
| PATCH | `/contracts/:id/status` | `active`, `paid`, `cancelled` |
| POST | `/contracts/:id/installments` | |
| PUT, DELETE | `/contracts/:id/installments/:childId` | |
| POST | `/contracts/:id/installments/:childId/payment` | Body `{ "paidDate": "..." }` |
| GET | `/installments` | Filter `contractId`, `customerId`, `status`. Urut jatuh tempo terdekat |

Nomor identitas pelanggan wajib diisi karena kolomnya unik. Kalau boleh kosong,
dua pelanggan tanpa nomor identitas akan bentrok di index unik.

### Prospek menjadi pelanggan

Satu prospek hanya bisa menjadi satu pelanggan (`lead.customer_id` unik, FK
`ON DELETE SET NULL`). Konversi mengunci baris prospek dengan `FOR UPDATE`,
menolak prospek yang sudah tertaut (409 `lead_already_converted`) atau yang
`cancelled` (422 `lead_cancelled`), dan gagal seluruhnya kalau nomor identitas
pelanggan bentrok.

### Unit mengikuti kontrak

Status `reserved` dan `sold` tidak bisa dikirim dari frontend. Membuat kontrak
mengunci baris unit dengan `FOR UPDATE`, memastikan unit masih `available`, lalu
menjadikannya `reserved`. Kontrak `paid` menjadikan unit `sold`. Kontrak
`cancelled` atau draf yang dihapus mengembalikan unit ke `available`.

| Kontrak dari | Boleh ke |
|---|---|
| `draft` | `active`, `cancelled` |
| `active` | `paid`, `cancelled` |
| `paid`, `cancelled` | tidak ada |

| Aturan | Kode error |
|---|---|
| Unit yang sudah dipesan atau terjual tidak bisa dikontrak lagi | `unit_not_available` |
| Unit yang terikat kontrak tidak bisa diubah manual | `unit_locked` |
| Kontrak hanya bisa diubah atau dihapus selama `draft`, dan unitnya tidak bisa diganti | `contract_locked`, `contract_unit_change_invalid` |
| Kontrak baru bisa `paid` kalau tidak ada cicilan yang belum lunas | `contract_unpaid` |
| Cicilan tidak bisa diubah kalau kontrak sudah `paid` atau `cancelled` | `contract_closed` |
| Total cicilan tidak boleh melebihi nilai kontrak | `installment_exceeds_contract` |
| Cicilan hanya bisa dibayar pada kontrak `active`, dan cicilan lunas terkunci | `contract_not_active`, `installment_paid` |

### NIK tersamar

Nomor identitas (`identity_number`) pelanggan dan karyawan adalah data pribadi
menurut UU No. 27 Tahun 2022.

| Pemanggil | `identityNumber` di respons | Pencarian `search` |
|---|---|---|
| Peran dengan hak tulis modulnya (Penjualan untuk pelanggan, SDM untuk karyawan, Admin) | Utuh, karena dipakai untuk mengubah data | Nama, email, dan NIK |
| Peran yang hanya membaca (misalnya Pengamat, Keuangan) | Hanya empat karakter terakhir, sisanya `•` | Nama dan email saja, supaya NIK tidak bisa ditebak lewat pencarian |

Laporan gaji (PDF, Excel, CSV) selalu memakai NIK tersamar, karena berkas
laporan bisa diunduh peran lain. Penyamaran ada di `shared/privacy`; NIK di
database belum dienkripsi.

## Finance

Semua di bawah `/api/finance`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/budgets` | Filter `projectId`, `year`. Berisi `realized`, `remaining`, dan `absorption` dalam persen |
| GET, PUT, DELETE | `/budgets/:id` | |
| GET, POST | `/cash-transactions` | Filter `type`, `accountId`, `projectId`, `dateFrom`, `dateTo`. Berisi `journalEntryId` |
| GET, PUT, DELETE | `/cash-transactions/:id` | Ikut mengubah atau menghapus jurnal otomatisnya |
| GET | `/cash-flow` | Kas masuk dan keluar per bulan, bawaan enam bulan terakhir. Filter `projectId`, `dateFrom`, `dateTo` |
| GET, POST | `/journal-entries` | Filter `search`, `accountId`, `dateFrom`, `dateTo`. Berisi `cashTransactionId`, `null` untuk jurnal manual |
| GET | `/journal-entries/:id` | Berisi baris debit dan kredit |
| GET | `/ledger` | Buku besar satu akun. Wajib `accountId`, filter `dateFrom`, `dateTo`. Tiap baris berisi `cashTransactionId` |

Realisasi anggaran adalah jumlah kas keluar proyek itu pada tahun anggaran,
dihitung saat dibaca. Arus kas juga dijumlahkan dari `cash_transaction`, bulan
tanpa transaksi tetap muncul dengan nilai nol, dan `balance` adalah saldo kas
sampai akhir periode.

Jurnal yang dicatat lewat `POST /journal-entries` tidak bisa diubah maupun
dihapus. Salah catat diperbaiki dengan jurnal balik. Buku besar memakai saldo
debit dikurangi kredit: saldo awal dihitung dari semua baris sebelum
`dateFrom`, lalu saldo berjalan dihitung dengan window function sebelum
paginasi supaya tetap benar di halaman mana pun.

### Jurnal otomatis dari transaksi kas

Setiap transaksi kas punya tepat satu jurnal sistem, jadi buku besar akun Kas
ikut mencatat setiap kas masuk dan keluar.

| Hal | Aturan |
|---|---|
| Akun Kas | Dicari dari bagan akun lewat kode `1110`, bukan UUID. Kalau tidak ada, transaksi kas ditolak dengan `cash_account_not_found` |
| Baris jurnal | `in`: debit Kas, kredit akun lawan. `out`: debit akun lawan, kredit Kas. Nilainya sama dengan `amount` |
| Kepala jurnal | Tanggal sama dengan transaksi, `source = "kas"`, keterangan dari `note` transaksi atau "Kas masuk"/"Kas keluar" bila kosong, `cash_transaction_id` menunjuk transaksinya |
| Nomor | `KAS-<tahun tanggal transaksi>-<urutan 6 digit>`, contoh `KAS-2026-000123`. Urutan diambil dari sequence `journal_entry_cash_number_seq`, jadi unik tanpa kunci dan aman saat dua transaksi dicatat bersamaan. Sequence tidak ikut rollback, jadi nomor bisa melompat. Nomor tidak berubah saat transaksi diubah, termasuk saat tahunnya pindah |
| Buat | Transaksi kas, kepala jurnal, dan barisnya ditulis dalam satu transaksi database |
| Ubah | Dalam satu transaksi: tanggal dan keterangan jurnal ditulis ulang, baris lama dihapus, baris baru ditulis. Id dan nomor jurnal tetap. Transaksi lama yang belum punya jurnal langsung dibuatkan |
| Hapus | Dalam satu transaksi: baris jurnal, jurnal, lalu transaksi kas dihapus. Foreign key `journal_entry.cash_transaction_id` juga `ON DELETE CASCADE` sebagai pengaman |
| Jurnal manual | `source` "kas" (tanpa beda huruf besar kecil) dan nomor berawalan `KAS-` ditolak supaya tidak bisa menyamar atau bentrok dengan nomor sistem. Tidak ada endpoint untuk mengubah atau menghapus jurnal, termasuk jurnal sistem; jurnal sistem hanya berubah lewat transaksi kasnya |

`journal_entry.cash_transaction_id` nullable dengan unique index
`uq_journal_cash_transaction` (banyak `NULL` boleh, satu transaksi kas hanya
satu jurnal) dan check `chk_journal_entry_cash_source` yang mewajibkan jurnal
bertaut kas bersumber `kas`. Kolom tautan dipilih ketimbang mencocokkan nomor
atau keterangan, karena tautan tetap utuh walau tanggal, nominal, atau catatan
transaksi diubah.

`cmd/migrate` menjalankan isi ulang (backfill) setelah semua tabel, index, dan
constraint siap: setiap transaksi kas yang belum punya jurnal dibuatkan dalam
satu transaksi database, urut tanggal lalu waktu catat. Transaksi yang sudah
bertaut dilewati lewat `NOT EXISTS`, dan unique index menjaga tidak ada jurnal
ganda, jadi migrate aman dijalankan berulang. Kalau tidak ada transaksi yang
perlu diisi, akun Kas tidak diperiksa, sehingga migrate di database kosong
(sebelum seed) tetap jalan.

| Aturan | Kode error |
|---|---|
| Satu proyek satu anggaran per tahun | `budget_year_used` |
| Jurnal minimal dua baris, tiap baris hanya debit atau kredit | `journal_line_single_sided` |
| Total debit harus sama dengan total kredit | `journal_unbalanced` |
| Akun Kas kode 1110 wajib ada untuk mencatat transaksi kas | `cash_account_not_found` |
| Akun lawan transaksi kas bukan akun Kas itu sendiri | `cash_counter_account_invalid` |
| Jurnal otomatis bentrok dengan unique index (seharusnya tidak terjadi) | `cash_journal_conflict` |
| Jurnal manual tidak boleh bersumber `kas` | `journal_source_reserved` |
| Nomor jurnal manual tidak boleh berawalan `KAS-` | `journal_number_reserved` |

## Billing

Semua di bawah `/api/billing`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/invoices` | Filter `search`, `status`, `partyType`, `partyId`, `projectId`, `recorded` (`false` = belum dicatat sebagai piutang). Sudah berisi `partyName` (pelanggan atau vendor sesuai `partyType`), `projectName`, `receivableId`, `receivableReference`, `lastPaidAt` |
| GET, PUT, DELETE | `/invoices/:id` | |
| POST | `/invoices/:id/payments` | Body `{ "amount": 1000000, "paidAt"?, "method"?, "reference"?, "note"? }`, lihat [Riwayat pembayaran](#riwayat-pembayaran) |
| GET | `/invoices/:id/payments` | Riwayat pembayaran, terbaru dulu, `page` dan `pageSize` |
| GET, POST | `/receivables` | Filter `search`, `status`, `customerId`, `projectId`, `contractId`. Body opsional `projectId`, `contractId`, `invoiceId`. Sudah berisi `customerName`, `projectName`, `contractNumber`, `invoiceNumber`, `lastPaidAt` |
| GET, PUT, DELETE | `/receivables/:id` | |
| POST, GET | `/receivables/:id/payments` | Sama seperti faktur |
| GET, POST | `/payables` | Filter `search`, `status`, `vendorId`, `projectId`, `purchaseOrderId`. Body opsional `projectId`, `purchaseOrderId`. Sudah berisi `vendorName`, `projectName`, `purchaseOrderNumber`, `lastPaidAt` |
| GET, PUT, DELETE | `/payables/:id` | |
| POST, GET | `/payables/:id/payments` | Sama seperti faktur |
| POST, GET | `/invoices/:id/onchain-requests`, `/receivables/:id/onchain-requests` | Permintaan bayar on-chain. POST body opsional `{ "amount": 50000000 }` (kosong berarti sisa tagihan). GET berisi permintaan tagihan itu dan pasangan tertautnya, terbaru dulu, `page` dan `pageSize`. Lihat [Pembayaran on-chain](#pembayaran-on-chain-sepolia) |
| GET | `/onchain/config` | `{ enabled, chainId, receiver, idrPerEth, explorerUrl, minConfirmations, requestTtlMinutes }` |
| POST | `/onchain/requests/:id/cancel` | Membatalkan permintaan yang masih `pending` |

Dua endpoint publik untuk pembeli ada di luar `/api/billing`, tanpa login:
`GET /api/pay/:token` dan `POST /api/pay/:token/confirm`.

### Tautan piutang dan utang ke sumbernya

Piutang bisa ditautkan ke proyek, kontrak penjualan, dan faktur. Utang bisa
ditautkan ke proyek dan pesanan pembelian. Semua tautan opsional dan nama atau
nomornya ikut di response lewat join, jadi frontend tidak perlu mencari
sendiri. PUT mengganti seluruh isi, jadi tautan yang tidak dikirim ikut lepas.

Faktur ke pelanggan dan piutang bisa mencatat utang yang sama dua kali. Karena
itu `receivable.invoice_id` unik: satu faktur paling banyak satu piutang, dan
faktur menampilkan `receivableId` serta `receivableReference` kalau sudah
dicatat.

| Kolom | Foreign key | Saat data induk dihapus |
|---|---|---|
| `receivable.project_id`, `payable.project_id` | `project` | `SET NULL` |
| `receivable.contract_id` | `contract` | `SET NULL` |
| `payable.purchase_order_id` | `purchase_order` | `SET NULL` |
| `receivable.invoice_id` | `invoice`, unik (`uq_receivable_invoice`) | `RESTRICT`, supaya faktur tidak hilang di bawah piutang yang mencatatnya. Hapus piutangnya dulu |

### Riwayat pembayaran

Setiap pembayaran menjadi satu baris `billing_payment`, dan `paid_amount` di
induknya tetap disimpan sebagai total berjalan supaya daftar dan KPI tidak
perlu menjumlah ulang. Baris riwayat dan `paid_amount` ditulis dalam satu
transaksi dengan baris induk dikunci (`SELECT ... FOR UPDATE`).

| Kolom | Isi |
|---|---|
| `invoice_id`, `receivable_id`, `payable_id` | Minimal satu terisi (`chk_billing_payment_parent`). Faktur dan piutang boleh terisi bersamaan, utang selalu sendiri. FK `ON DELETE CASCADE`, index per kolom |
| `amount` | Lebih dari nol (`chk_billing_payment_amount`) |
| `paid_at` | Tanggal bayar. Kalau tidak dikirim, hari ini menurut Asia/Jakarta. Tanggal setelah hari ini ditolak |
| `method` | `transfer` (bawaan), `tunai`, `cek`, `lainnya`, `onchain`. Divalidasi `oneof` di DTO, bukan check constraint, supaya metode baru cukup menambah nilai di DTO |
| `reference`, `note` | Opsional, maksimal 80 dan 200 karakter. 80 supaya hash transaksi on-chain (66 karakter) muat; kolomnya dilebarkan dari `varchar(60)` oleh AutoMigrate |
| `created_by` | User yang mencatat, dari token. FK ke `users`, `SET NULL` |

Respons riwayat berisi `invoiceNumber`, `receivableReference`, dan
`createdByName` lewat join, supaya rincian faktur bisa menulis "Juga tercatat
di piutang X" dan sebaliknya.

### Pembayaran tercatat sebagai kas

Setiap pembayaran di Tagihan, termasuk pembayaran on-chain, langsung menjadi
satu transaksi kas di Keuangan dalam transaksi database yang sama. Kalau kasnya
gagal dicatat, pembayarannya ikut dibatalkan, jadi sisa tagihan dan kas tidak
pernah berbeda.

| Pembayaran | Transaksi kas | Akun lawan | Jurnal |
|---|---|---|---|
| Faktur atau piutang (termasuk pasangan yang tertaut, tetap satu kas) | `in` | `1200` Piutang Usaha | Debit Kas, kredit Piutang Usaha |
| Utang | `out` | `2100` Utang Usaha | Debit Utang Usaha, kredit Kas |

- Tanggal kas sama dengan `paid_at`, nominal sama dengan pembayaran, dan
  proyeknya diambil dari tagihan (untuk faktur tanpa proyek, dari piutang
  tautannya). Keterangannya "Pembayaran faktur INV-...", "Pembayaran piutang
  ...", atau "Pembayaran hutang ...".
- `cash_transaction.billing_payment_id` menunjuk baris `billing_payment`, unik,
  FK `ON DELETE CASCADE`. Kas seperti ini tidak bisa diubah atau dihapus dari
  Keuangan (422 `cash_from_billing`); kalau pembayarannya terhapus, kas dan
  jurnalnya ikut terhapus.
- Akun lawan dicari lewat kode. Kalau `1200` atau `2100` tidak ada,
  pembayaran ditolak dengan `billing_cash_account_not_found`.
- Slice `billing` mendeklarasikan antarmuka `CashBook`; adapternya ke
  `finance.BillingCashLedger` ada di `billing/module.go`. Repositori yang
  dibuat dengan `NewRepository(db, nil)` (migrasi, ledger utang dari modul
  Proyek) tidak mencatat kas.
- Pembayaran yang dicatat sebelum fitur ini ada tidak diisi ulang ke kas,
  karena kasnya mungkin sudah dicatat manual di Keuangan.

### Faktur dan piutang yang tertaut dibayar bersama

Piutang yang mencatat faktur (`receivable.invoice_id`) dan fakturnya adalah
utang yang sama, jadi saldonya harus selalu sama.

1. Pembayaran ke salah satunya menulis **satu** baris riwayat berisi
   `invoice_id` dan `receivable_id`, lalu menaikkan `paid_amount` dan status
   keduanya di transaksi yang sama.
2. Urutan kunci selalu faktur dulu, lalu piutang, dari arah mana pun
   pembayarannya datang, supaya dua pembayaran bersamaan tidak saling
   mengunci. Bayar lewat piutang membaca `invoice_id` dulu tanpa kunci, lalu
   mengunci faktur, lalu piutang. Kalau tautannya berubah di antara itu,
   pembayaran ditolak dengan `billing_link_changed` (409) dan cukup diulang.
3. Pembayaran yang melebihi sisa salah satu sisi ditolak (422
   `payment_exceeds_outstanding`), dengan pesan yang menyebut sisi mana dan
   sisanya, misalnya "Pembayaran melebihi sisa piutang PTG-1 yang tertaut,
   sisanya tinggal Rp 500.000".

Aturan saat piutang ditautkan ke faktur (buat piutang dengan `invoiceId`, atau
ubah piutang ke faktur lain):

| Keadaan | Hasil |
|---|---|
| Keduanya belum dibayar, atau total bayarnya sudah sama | Ditautkan apa adanya |
| Hanya faktur yang sudah dibayar | Total bayar faktur disalin ke piutang, baris riwayat faktur ikut diberi `receivable_id`. Nilai piutang harus minimal sebesar total itu (`receivable_below_invoice_paid`) |
| Hanya piutang yang sudah dibayar | Total bayar piutang disalin ke faktur, baris riwayat piutang ikut diberi `invoice_id` |
| Keduanya sudah dibayar dengan total berbeda | Ditolak (422 `link_paid_mismatch`), karena tidak bisa diketahui pembayaran mana yang sama |
| Piutang yang sudah dibayar dilepas dari fakturnya atau dipindah ke faktur lain | Ditolak (422 `receivable_link_has_payment`), karena riwayatnya sudah dipakai bersama |

Karena total bayar faktur ikut disalin, "Catat sebagai piutang" di frontend
mengisi nilai piutang dengan nilai penuh faktur, bukan sisanya.

### Backfill riwayat

`cmd/migrate` mengisi riwayat untuk data lama yang `paid_amount > 0` tapi
belum punya baris, dengan satu baris senilai total itu, `paid_at =
updated_at::date`, metode `lainnya`, catatan "Saldo terbayar sebelum riwayat
pembayaran dicatat". Pasangan faktur dan piutang tertaut yang total bayarnya
sama mendapat satu baris bersama. Pasangan yang totalnya sudah berbeda sebelum
riwayat ada tidak disamakan otomatis: tiap sisi mendapat barisnya sendiri dan
selisihnya perlu dirapikan manual. Semua `INSERT` dijaga `NOT EXISTS`, jadi
menjalankan migrate kedua kalinya tidak menambah baris.

### Pembayaran on-chain (Sepolia)

Simulasi pembayaran masuk lewat jaringan uji Ethereum Sepolia untuk faktur ke
pelanggan dan piutang (bukan utang dan bukan faktur vendor). Perusahaan hanya
**menerima** ke satu alamat yang diatur lewat env; server tidak pernah memegang
kunci privat dan tidak pernah mengirim transaksi.

Alur:

1. Staf membuat permintaan bayar di rincian faktur atau piutang. Nominal bawaan
   adalah sisa tagihan (untuk pasangan faktur dan piutang yang tertaut, sisa
   terkecil dari keduanya), harus lebih dari nol dan tidak melebihi sisa itu.
   Selama masih ada permintaan `pending` yang belum kedaluwarsa untuk tagihan
   itu atau pasangannya, permintaan baru ditolak (`onchain_request_active`).
2. Server menyimpan nominal rupiah, nilai wei (dibulatkan ke atas memakai kurs
   simulasi), alamat penerima, chain id, kurs, token publik acak (24 byte,
   base64url), dan batas waktu. Frontend menampilkan QR dari URI EIP-681
   `ethereum:<penerima>@<chainId>?value=<wei>` dan tautan `/bayar/<token>`.
3. Pembeli membayar dari dompetnya (MetaMask di peramban atau memindai QR),
   lalu halaman bayar mengirim hash transaksinya ke
   `POST /api/pay/:token/confirm` `{ "txHash": "0x..." }`.
4. Server memeriksa transaksi sendiri lewat JSON-RPC (`net/http`, tanpa
   dependensi baru): `eth_chainId`, `eth_getTransactionByHash`,
   `eth_getTransactionReceipt`, `eth_blockNumber`, dan `eth_getBlockByNumber`
   kalau node tidak menyertakan `blockTimestamp`.
5. Kalau lolos, dalam satu transaksi database: baris permintaan dikunci, lalu
   pembayaran dicatat lewat jalur yang sama dengan `POST .../payments` (faktur
   dikunci dulu, lalu piutang tertautnya, satu baris `billing_payment` dengan
   metode `onchain`, referensi = hash transaksi, catatan berisi alamat pengirim
   dan tautan explorer, `created_by` = staf yang membuat permintaan), lalu
   permintaan ditandai `paid` beserta alamat pengirim dan waktu bayar.

Konfigurasi (`internal/shared/config/onchain.go`):

| Env | Bawaan | Keterangan |
|---|---|---|
| `ONCHAIN_RECEIVER_ADDRESS` | kosong | Alamat penerima `0x...`. Kosong berarti fitur mati: `/onchain/config` membalas `enabled: false` dan frontend menyembunyikan "Bayar on-chain". Huruf campuran harus lolos checksum EIP-55; alamat nol ditolak. Server gagal start kalau alamatnya rusak |
| `ONCHAIN_RPC_URL` | `https://ethereum-sepolia-rpc.publicnode.com` | Node JSON-RPC publik |
| `ONCHAIN_CHAIN_ID` | `11155111` | Sepolia |
| `ONCHAIN_IDR_PER_ETH` | `1e12` | **Kurs simulasi**, bukan harga pasar. Rp 1 triliun = 1 ETH, jadi Rp 50 juta = 0,00005 SepoliaETH, cukup dari faucet. Boleh ditulis `1e12` atau `1000000000000` |
| `ONCHAIN_MIN_CONFIRMATIONS` | `1` | Jumlah blok (termasuk blok transaksi) sebelum dicatat |
| `ONCHAIN_REQUEST_TTL` | `30m` | Masa berlaku permintaan, minimal `1m` |
| `ONCHAIN_EXPLORER_URL` | `https://sepolia.etherscan.io` | Dasar tautan `/tx/<hash>` |

Tabel `onchain_payment_request`:

| Kolom | Isi |
|---|---|
| `target_kind`, `target_id` | `invoice` atau `receivable` (`chk_onchain_payment_request_target_kind`) dan id-nya, index gabungan. Tanpa FK karena polimorfik; tagihan yang sudah dihapus membuat halaman bayar membalas 404 `onchain_target_not_found` |
| `amount_idr`, `amount_wei`, `idr_per_eth` | Wei disimpan sebagai teks angka (`varchar(80)`, dicek regex) karena melebihi int64. Kurs ikut disimpan supaya tampilan tidak berubah kalau env diganti |
| `receiver`, `chain_id` | Disalin dari env saat dibuat, dan verifikasi memakai nilai yang tersimpan ini |
| `token` | Unik (`uq_onchain_payment_request_token`) |
| `status` | `pending`, `paid`, `expired`, `failed`, `cancelled`. `paid` wajib punya `tx_hash` dan `paid_at` (`chk_onchain_payment_request_paid`) |
| `tx_hash` | Unik bila terisi (`uq_onchain_payment_request_tx_hash`), huruf kecil |
| `payer_address`, `failure_reason`, `created_by` (FK `users`, `SET NULL`), `expires_at`, `paid_at` | |

Status `pending` yang sudah lewat `expires_at` dibalas sebagai `expired` di
semua respons tanpa mengubah baris; baris baru benar-benar diubah ke `expired`
kalau ada transaksi yang terbukti masuk blok setelah batas waktu. Kolom unik
memakai tag `uniqueIndex`, jadi `cmd/migrate` tetap bisa dijalankan berulang.

Endpoint publik (segmen `pay` terbuka di `access/matrix.go`):

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/pay/:token` | Nomor faktur atau referensi piutang, nama proyek, nominal rupiah, wei dan ETH, kurs, penerima, chain id, URI EIP-681, status, batas waktu, hash dan tautan explorer kalau sudah dibayar. Tidak ada nama pelanggan atau id internal |
| POST | `/api/pay/:token/confirm` | 200 kalau sudah dicatat (`verification: "paid"`), 202 kalau masih menunggu (`verification: "waiting"` beserta pesan seperti "menunggu konfirmasi jaringan (1 dari 2)"), atau error dengan alasan berbahasa Indonesia |

Token yang bentuknya salah langsung 404 tanpa query, `txHash` divalidasi
binding (`len=66,startswith=0x,hexadecimal`), dan setiap token dibatasi 60
GET dan 20 confirm per menit di memori (429 `too_many_requests`).

Aturan verifikasi, berurutan:

| Aturan | Hasil |
|---|---|
| Node RPC harus di chain yang sama dengan permintaan | 502 `chain_mismatch` |
| Node tidak bisa dihubungi atau jawabannya rusak | 502 `chain_unavailable`, halaman bayar mencoba lagi |
| Transaksi belum terlihat atau belum masuk blok | 202 menunggu; setelah batas waktu lewat jadi 422 `onchain_request_expired` |
| `chainId` transaksi (kalau ada) sama dengan permintaan | 422 `onchain_wrong_chain` |
| `to` sama dengan penerima, tanpa membedakan huruf besar kecil | 422 `onchain_wrong_receiver` |
| `value` minimal sebesar wei permintaan (lebih boleh, yang dicatat tetap nominal rupiahnya) | 422 `onchain_value_too_low` dengan kedua nilai |
| Status receipt `0x1` | 422 `onchain_tx_reverted` |
| Waktu blok tidak sebelum permintaan dibuat (toleransi jam 2 menit) | 422 `onchain_tx_before_request` |
| Waktu blok tidak setelah `expires_at` (toleransi 2 menit). Transaksi yang masuk tepat waktu tetap dicatat walaupun konfirmasinya dikirim setelah kedaluwarsa | 422 `onchain_tx_after_expiry`, permintaan ditandai `expired` beserta hash di alasan |
| Konfirmasi minimal `ONCHAIN_MIN_CONFIRMATIONS` | 202 menunggu |
| Hash belum dipakai permintaan lain atau pembayaran `onchain` lain di `billing_payment` (dicek sebelum RPC) | 409 `onchain_tx_used` |

Idempoten: konfirmasi ulang permintaan yang sudah `paid` dengan hash yang sama
membalas status lunas tanpa pembayaran kedua (juga kalau dua konfirmasi datang
bersamaan, karena baris permintaan dikunci). Hash lain untuk permintaan yang
sudah lunas dibalas 409 `onchain_request_paid`. Kalau transaksi sah tetapi
pencatatan ditolak (misalnya tagihan sudah dilunasi lewat transfer sementara
itu, atau tagihannya dihapus), permintaan ditandai `failed` beserta hash,
pengirim, dan alasannya, lalu dibalas 422 `onchain_settlement_failed`; staf
perlu mengurus dana itu secara manual. Permintaan yang dibatalkan (409
`onchain_request_cancelled`) tidak bisa dikonfirmasi lagi.

Klien RPC (`internal/shared/ethereum`) adalah implementasi interface
`billing.ChainReader`, jadi tes memakai node palsu. Respons asli Sepolia (chain
id, nomor blok, transaksi, receipt, blok) dipakai sebagai fixture tes pengurai.

Yang belum: server tidak memindai blok untuk menemukan pembayaran sendiri.
Pembayaran dari HP lewat QR tetap perlu hash transaksinya ditempel di halaman
bayar (tombol "Saya sudah bayar, tempel hash transaksi").

### Status jatuh tempo tidak dipercaya dari kolom

`status` tetap disimpan supaya index parsial `WHERE status <> 'paid'` terpakai,
tapi selain `paid` nilainya bisa basi keesokan harinya. Karena itu status di
response selalu dihitung ulang dari `dueDate`, begitu juga `daysOverdue` dan
`outstanding`. Filter `status` diterjemahkan menjadi rentang tanggal:

| Status | Arti |
|---|---|
| `overdue` | Belum lunas dan jatuh tempo sebelum hari ini |
| `due` | Belum lunas dan jatuh tempo hari ini sampai tujuh hari ke depan |
| `not_due` | Belum lunas dan jatuh tempo lebih dari tujuh hari lagi |
| `paid` | Sudah dibayar penuh |

Aturan yang sama dipakai untuk cicilan di slice `sales`.

| Aturan | Kode error |
|---|---|
| Pembayaran dicatat dalam transaksi dengan baris dikunci, dan tidak boleh melebihi sisa tagihan (di kedua sisi untuk faktur dan piutang yang tertaut) | `payment_exceeds_outstanding` |
| Tanggal bayar tidak boleh setelah hari ini (WIB) | `payment_date_in_future` |
| Tautan faktur dan piutang berubah saat pembayaran menunggu kunci (409) | `billing_link_changed` |
| Aturan penautan piutang ke faktur yang sudah dibayar (lihat di atas) | `link_paid_mismatch`, `receivable_below_invoice_paid`, `receivable_link_has_payment` |
| Jumlah tagihan tidak boleh diubah di bawah yang sudah dibayar | `amount_below_paid` |
| Data yang sudah menerima pembayaran tidak bisa dihapus | `invoice_has_payment`, `receivable_has_payment`, `payable_has_payment` |
| `invoice.partyId` tidak punya foreign key karena bisa menunjuk pelanggan atau vendor, jadi keberadaannya diperiksa repository | `invoice_party_not_found` |
| Pelanggan piutang, vendor utang, dan proyek yang ditautkan harus ada (422) | `customer_not_found`, `vendor_not_found`, `project_not_found` |
| Kontrak yang ditautkan harus ada dan milik pelanggan piutang (422) | `contract_not_found`, `contract_customer_mismatch` |
| Pesanan pembelian yang ditautkan harus ada dan milik vendor utang (422) | `purchase_order_not_found`, `purchase_order_vendor_mismatch` |
| Faktur yang ditautkan harus ada, ditagihkan ke pelanggan piutang, dan nilai piutang tidak melebihi nilai faktur (422) | `linked_invoice_not_found`, `invoice_customer_mismatch`, `receivable_exceeds_invoice` |
| Satu faktur hanya satu piutang (409, "Faktur ini sudah dicatat sebagai piutang <referensi>") | `invoice_already_recorded` |
| Faktur yang sudah dicatat sebagai piutang tidak bisa dihapus (409), dan saat diubah pelanggannya harus tetap sama serta nilainya tidak di bawah nilai piutang (422) | `invoice_has_receivable`, `invoice_receivable_mismatch` |
| Permintaan bayar on-chain hanya untuk faktur pelanggan dan piutang yang masih bersisa, nominal tidak melebihi sisa, satu yang aktif per tagihan, dan fitur harus aktif (503) | `onchain_vendor_invoice`, `onchain_nothing_outstanding`, `onchain_amount_exceeds_outstanding`, `onchain_request_active`, `onchain_disabled` |
| Verifikasi transaksi on-chain (lihat tabel di atas) | `onchain_*`, `chain_unavailable`, `chain_mismatch` |

## Evidence (jejak bukti blockchain)

Slice `internal/evidence`. Masalahnya: tagihan termin dibayar berdasarkan progres, dan buktinya (laporan harian, foto lapangan, progres, pembayaran) disimpan di database pengembang sendiri. Pemberi kerja, bank, dan pembeli tidak bisa memastikan bukti itu tidak diubah setelah tagihan diajukan.

### Sidik jari dan pencatatan

- Sumber bukti dibaca lewat adapter, mengikuti pola lintas slice: `project.EvidenceFeed` (laporan harian, foto lapangan dengan SHA-256 isi berkasnya, riwayat progres, termin pemberi kerja yang sudah ditagih) dan `billing.PaymentFeed` (semua pembayaran tagihan, **apa pun metodenya**: transfer, tunai, cek, on-chain).
- Tiap bukti diberi sidik jari SHA-256 dari JSON isinya. Versi baru (isi berubah) menjadi `evidence_record` baru, jadi riwayatnya tidak hilang.
- `POST /evidence/batches` mengumpulkan semua sidik jari yang belum tercatat menjadi satu akar Merkle (keccak256, daun diurutkan supaya deterministik). Kumpulan yang belum tercatat diganti kalau dibuat ulang.
- Pengguna mengirim **transaksi 0 ETH ke alamat sendiri dari dompetnya** dengan data `0x424e474e` ("BNGN") + akar. Server tidak memegang kunci privat.
- `POST /evidence/batches/:id/anchor` memeriksa transaksi di Sepolia: data harus sama persis, receipt sukses, jaringan benar, jumlah konfirmasi cukup. Belum terlihat atau belum terkonfirmasi dijawab `waiting`.

### Status bukti

| Status | Arti |
|---|---|
| Tercatat, utuh | Sidik jari saat ini sama dengan yang tercatat di blockchain |
| Belum dicatat | Bukti baru, belum masuk transaksi apa pun |
| Berubah setelah dicatat | Isi sekarang berbeda dari yang tercatat |
| Dihapus setelah dicatat | Pernah tercatat, sekarang tidak ada |

### Persetujuan pemberi kerja (EIP-712)

- Alamat dompet pemberi kerja didaftarkan per proyek (`PUT /evidence/projects/:id/approver`, divalidasi EIP-55).
- Tiap termin yang sudah ditagih bisa diberi tautan verifikasi publik `/verifikasi/:token` (segmen API `verify` terbuka seperti `pay`).
- Pemberi kerja melihat tagihan dan status semua bukti, lalu menandatangani `OwnerClaimApproval(project, contract, term, progress, amount, claimedOn, evidenceRoot)` dengan `eth_signTypedData_v4`. Gratis, tanpa transaksi.
- Server memulihkan alamat penanda tangan (secp256k1) dan hanya menerima alamat yang terdaftar. Tanda tangan mengikat nilai tagihan **dan** akar sidik jari bukti proyek saat itu; kalau bukti berubah sesudahnya, persetujuan ditandai "bukti berubah sejak disetujui".
- Pengodean EIP-712 diuji terhadap vektor resmi di spesifikasi EIP-712 (contoh "Mail").

## HR

Semua di bawah `/api/hr`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/employees` | Filter `search`, `employmentType`, `projectId`, `active` |
| GET | `/employees/summary` | Jumlah karyawan aktif per jenis kepegawaian |
| GET, PUT, DELETE | `/employees/:id` | |
| GET, POST | `/attendances` | Filter `employeeId`, `projectId`, `status`, `dateFrom`, `dateTo`. Sudah berisi nama karyawan |
| GET, PUT, DELETE | `/attendances/:id` | Jam ditulis `HH:MM` |
| GET, POST | `/payrolls` | Filter `employeeId`, `projectId` (proyek karyawan), `period` (`YYYY-MM`), `paid` |
| GET | `/payrolls/summary` | Filter `period`, `projectId`. `count`, `grossPay` (gaji pokok ditambah tunjangan), `netPay`, `paidNetPay`, `unpaidNetPay`, `unpaidCount` dijumlah dengan SQL dari semua slip yang cocok |
| GET, PUT, DELETE | `/payrolls/:id` | |
| PATCH | `/payrolls/:id/pay` | Menandai sudah dibayar |

Gaji pokok tidak dikirim dari frontend. Pekerja harian dibayar gaji dasar dikali
jumlah hari `present` pada periode itu, jenis lain memakai gaji dasar. Gaji
bersih adalah gaji pokok ditambah tunjangan dikurangi potongan.

Kolom `check_in_time` dan `check_out_time` bertipe `time`. Driver pgx membaca
tipe itu sebagai teks, jadi field Go-nya `*string`, bukan `*time.Time`.

| Aturan | Kode error |
|---|---|
| Karyawan yang sudah punya absensi atau penggajian tidak bisa dihapus. Isi tanggal keluar | `employee_in_use` |
| Absensi hanya di dalam masa kerja, satu per karyawan per hari | `attendance_outside_employment`, `attendance_already_recorded` |
| Status `present` wajib jam masuk, status lain tanpa jam, jam pulang tidak sebelum jam masuk | `attendance_check_in_required`, `attendance_time_not_allowed`, `attendance_time_invalid` |
| Penggajian hanya di dalam masa kerja, satu per karyawan per periode | `payroll_outside_employment`, `payroll_already_exists` |
| Potongan tidak boleh membuat gaji bersih minus | `payroll_net_negative` |
| Penggajian yang sudah dibayar tidak bisa diubah atau dihapus | `payroll_paid` |

## Scoring

Semua di bawah `/api/scoring`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/dimensions` | Tanpa paginasi, urut `sortOrder` |
| GET, PUT, DELETE | `/dimensions/:id` | |
| GET, POST | `/building-profiles` | |
| GET, PUT, DELETE | `/building-profiles/:id` | GET berisi bobot |
| GET, PUT | `/building-profiles/:id/weights` | PUT mengganti seluruh bobot profil |

| Aturan | Kode error |
|---|---|
| Total bobot satu profil harus 100 | `weight_total_invalid` |
| Satu dimensi sekali per profil | `weight_dimension_duplicate` |
| Dimensi yang masih dipakai bobot atau skor lokasi tidak bisa dihapus | `dimension_in_use` |

## Location

Semua di bawah `/api/locations` dan hanya melihat data milik pengguna yang
sedang masuk. Lokasi atau perbandingan milik orang lain dibalas 404.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/saved` | Filter `search`, `regionId`, `buildingProfileId`, `projectId`, `minScore`, serta area peta `minLongitude`, `minLatitude`, `maxLongitude`, `maxLatitude`. Urut skor tertinggi |
| GET, PUT, DELETE | `/saved/:id` | Berisi skor per dimensi. PUT mengganti seluruh skor dimensi |
| GET, POST | `/comparisons` | Body berisi dua sampai lima `savedLocationIds` |
| GET, PUT, DELETE | `/comparisons/:id` | GET berisi skor per dimensi setiap lokasi |
| GET | `/saved/:id/report` | PDF Laporan Analisis Lokasi. `Content-Disposition: attachment; filename=analisis-lokasi-<nama>-<tanggal>.pdf` |
| GET | `/comparisons/:id/report` | PDF Laporan Perbandingan Lokasi, satu kolom per lokasi. `filename=perbandingan-lokasi-<nama>-<tanggal>.pdf` |

Kedua laporan PDF disusun hanya dari data tersimpan dan tidak memanggil layanan
skoring: skor keseluruhan, skor dimensi, penanda risiko, indeks banjir dan
gempa, serta harga tanah dari `saved_location` dan `dimension_score`; nama,
deskripsi, dan bobot dimensi dari tabel `dimension`, `weight`, dan
`building_profile`; konteks regulasi dari tabel `regulation` di titik lokasi
(nilai simulasi kalau RDTR belum ada). Penjelasan per dimensi dari layanan
skoring tidak disimpan, jadi laporan memakai deskripsi dimensi sebagai kolom
"Yang diukur". `risk_flags` kosong (`null`) ditulis "tidak tercatat", bukan
"tidak ada risiko". Waktu di laporan dan tanggal di nama berkas memakai WIB.

Area peta disaring dengan `point && ST_MakeEnvelope(...)` supaya index GiST
`idx_saved_location_point` terpakai. Skor lokasi dikirim dari hasil analisis
karena mesin skoring belum diputuskan. `savedAt` diperbarui hanya kalau skor
berubah, sesuai catatan prototipe bahwa skor yang tampil adalah skor saat
lokasi disimpan.

Lokasi boleh dikaitkan ke satu proyek lewat `projectId`. Evaluasi lewat
`POST /api/site/evaluate` menyimpan `project_id` yang dikirim ke lokasi hasil
evaluasi. Kaitan ini memakai `ON DELETE SET NULL`, jadi menghapus proyek tidak
ikut menghapus lokasinya.

| Aturan | Kode error |
|---|---|
| Batas area peta harus lengkap empat nilai dan minimum lebih kecil dari maksimum | `location_bounds_invalid` |
| Satu dimensi sekali per lokasi | `dimension_score_duplicate` |
| Lokasi yang dibandingkan harus milik pengguna dan tidak ganda | `comparison_location_not_found`, `comparison_location_duplicate` |

## Reporting

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/api/reports` | Filter `search`, `type`, `format`, `projectId`. Pembuat diambil dari access token. Daftar membawa `projectCode`, `projectName`, dan `createdByName` |
| GET | `/api/reports/preview` | Query `type` (wajib), `projectId`, `periodStart`, `periodEnd` (`YYYY-MM-DD`). Isi laporan sebagai JSON, paling banyak 200 baris per bagian plus `totalRows` |
| GET, DELETE | `/api/reports/:id` | |
| GET | `/api/reports/:id/download` | Berkas dibuat ulang dari data terbaru saat diunduh. `Content-Disposition: attachment; filename=laporan-arus-kas-2026-10.pdf`, header itu dibuka lewat `Access-Control-Expose-Headers` |

Baris `report` hanya menyimpan kriteria (jenis, proyek, periode, format), bukan
berkasnya. Isinya selalu mengikuti data saat diunduh.

Jenis laporan (`type`):

| Jenis | Modul data | Isi | Arti periode |
|---|---|---|---|
| `arus_kas` | `finance` | Kas masuk, keluar, arus bersih, dan saldo per bulan mulai dari saldo awal, lalu semua transaksi kas | Tanggal transaksi. Kosong berarti bulan berjalan |
| `anggaran` | `finance` | Anggaran per proyek per tahun, realisasi (kas keluar proyek di tahun itu), sisa, serapan | Tahun anggaran dari tahun tanggal awal sampai tahun tanggal akhir |
| `proyek` | `projects` | Status, progres, jadwal, nilai kontrak, total RAB | Proyek yang jadwalnya beririsan dengan periode |
| `piutang_utang` | `billing` | Piutang dan utang yang belum lunas, kelompok umur 0–30, 31–60, 61–90, > 90 hari terlambat (belum jatuh tempo masuk 0–30) | Tanggal jatuh tempo. Umur dihitung sampai hari ini |
| `stok` | `inventory` | Stok per gudang dan material di bawah stok minimum (stok minimum dibanding jumlah di semua gudang yang tersaring) | Tidak memakai periode, posisi hari ini |
| `penggajian` | `hr` | Ringkasan per periode gaji dan rincian slip | Periode gaji `YYYY-MM`. Kosong berarti bulan berjalan |
| `penjualan` | `sales` | Kontrak dengan nilai terbayar dan jumlah cicilan lunas atau terlambat, lalu status cicilan | Tanggal kontrak untuk kontrak, tanggal jatuh tempo untuk cicilan |

### Hak akses per jenis laporan

Middleware hanya memeriksa modul `reports` (baca untuk GET, ubah untuk POST dan
DELETE). Karena isi laporan berasal dari modul lain, service juga mewajibkan hak
**baca** pada modul data di kolom "Modul data" di atas, memakai peran yang sudah
disimpan middleware di context (`middleware.AuthenticatedRole`), tanpa query
tambahan. Pemetaannya ada di `reportKinds` pada `internal/reporting/catalog.go`.

| Endpoint | Pemeriksaan tambahan |
|---|---|
| `GET /api/reports/preview`, `POST /api/reports` | Jenis yang diminta harus terbaca. Ditolak sebelum data apa pun dibaca |
| `GET /api/reports/:id`, `GET /api/reports/:id/download`, `DELETE /api/reports/:id` | Jenis laporan yang tersimpan harus terbaca. Jenis lama yang tidak dikenal hanya untuk admin |
| `GET /api/reports` | Hanya laporan dengan jenis yang terbaca (`report.type IN (...)`), jadi total halaman ikut benar |

Penolakan dibalas 403 `forbidden` dengan pesan yang sama dengan middleware,
misalnya "Peran Anda tidak punya akses melihat modul SDM". Frontend memakai
pemetaan yang sama di `models/report.ts` dan hanya menawarkan jenis yang terbaca.

Filter proyek berlaku untuk semua jenis: kas, anggaran, piutang, dan utang lewat
`project_id`, stok lewat gudang proyek, penggajian lewat proyek karyawan,
penjualan lewat proyek unit. Kontrak batal tetap tampil tetapi tidak masuk total.
Alias lama (`cash_flow`, `budget`, `projects`, `stock`, `payroll`, `sales`, dan
sejenisnya) masih diterima dan disimpan sebagai nama baku di atas. Untuk laporan
bulanan, periode yang dipakai saat dibuat disimpan di baris laporan supaya
"Unduh lagi" memberi periode yang sama.

Format berkas (`format`), ditulis tanpa dependensi tambahan di paket
`reporting/export`:

| Format | Bentuk |
|---|---|
| `pdf` | A4, Helvetica WinAnsi, judul, baris filter, tabel per bagian dengan kepala tabel yang diulang di halaman baru, nomor halaman "Halaman x dari y". Otomatis landscape kalau tabel terlalu lebar |
| `xlsx` | Satu sheet per bagian, kepala tabel tebal dan dibekukan, angka tersimpan sebagai angka (`#,##0`), tanggal sebagai tanggal Excel, lebar kolom menyesuaikan isi |
| `csv` | UTF-8 dengan BOM, pemisah titik koma (`;`) supaya langsung terbaca Excel berpengaturan Indonesia. Angka tanpa pemisah ribuan, desimal memakai koma, tanggal `YYYY-MM-DD`. Setiap bagian diawali baris judul |

| Aturan | Kode error |
|---|---|
| Jenis laporan harus salah satu dari tabel di atas | `report_type_invalid` |
| Peran harus bisa membaca modul data jenis laporan itu | `forbidden` (403) |
| Laporan stok tidak menerima periode | `report_period_unsupported` |
| Tanggal akhir tidak boleh sebelum tanggal awal | `invalid_date_range` |
| Proyek pada filter harus ada | `project_not_found` |

Setiap bagian dibatasi 10.000 baris.

## Yang belum dikerjakan

1. Hak akses masih per modul, belum per proyek. Keanggotaan tim
   (`project_member`) belum membatasi proyek mana yang bisa dilihat atau diubah.
2. Penagihan termin ke pemberi kerja (arah piutang) belum ada. Termin yang ada
   sekarang adalah pembayaran ke kontraktor pelaksana.
3. Pembayaran hutang di Tagihan belum otomatis tercatat sebagai kas keluar di
   modul Keuangan.
