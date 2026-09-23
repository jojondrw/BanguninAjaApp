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
| GET, POST | `/` | Filter `search` (nama atau kode), `status`, `regionId` |
| GET, PUT, DELETE | `/:id` | PUT tidak mengubah status dan progres |
| PATCH | `/:id/status` | Body `{ "status": "ongoing" }` |
| GET, POST | `/:id/phases` | Daftar tahap urut `sortOrder` |
| PUT, DELETE | `/:id/phases/:childId` | |
| GET, POST | `/:id/budget-items` | Berhalaman, ditambah `grandTotal` untuk seluruh RAB proyek |
| PUT, DELETE | `/:id/budget-items/:childId` | |
| GET, POST | `/:id/permits` | |
| PUT, DELETE | `/:id/permits/:childId` | |

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

`skema.md` mencatat progres proyek nantinya diturunkan dari tahapan. Sekarang
sudah begitu: setiap tahap ditambah, diubah, atau dihapus, progres proyek
dihitung ulang sebagai rata-rata progres semua tahap lalu dibulatkan. Proses ini
berjalan dalam satu transaksi dan baris proyek dikunci dengan `FOR UPDATE`,
supaya dua perubahan tahap yang bersamaan tidak saling menimpa hasil hitungan.

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

## Inventory

Semua di bawah `/api/inventory`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/materials` | Filter `search`, `category` |
| GET | `/materials/low-stock` | Material dengan total stok semua gudang di bawah `minimumStock` |
| GET, PUT, DELETE | `/materials/:id` | |
| GET, POST | `/warehouses` | Filter `search`, `projectId` |
| GET, PUT, DELETE | `/warehouses/:id` | |
| GET | `/stocks` | Filter `materialId`, `warehouseId`. Sudah berisi kode dan nama material serta gudang |
| GET, POST | `/stock-movements` | Filter `materialId`, `warehouseId`, `type`, `dateFrom`, `dateTo` dengan format `YYYY-MM-DD` |
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
| GET, POST | `/goods-receipts` | Filter `search`, `purchaseOrderId`, `warehouseId`, `dateFrom`, `dateTo` |
| GET | `/goods-receipts/:id` | |

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
Penerimaan belum menambah stok gudang secara otomatis, karena stok milik slice
`inventory` dan slice tidak boleh saling impor. Untuk sementara stok masuk
dicatat lewat `POST /api/inventory/stock-movements` jenis `in` dengan
`reference` berisi nomor penerimaan.

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
| GET, POST | `/contracts` | Filter `search`, `customerId`, `unitId`, `status`, `type`. Sudah berisi nama pelanggan dan kode unit |
| GET, PUT, DELETE | `/contracts/:id` | |
| PATCH | `/contracts/:id/status` | `active`, `paid`, `cancelled` |
| POST | `/contracts/:id/installments` | |
| PUT, DELETE | `/contracts/:id/installments/:childId` | |
| POST | `/contracts/:id/installments/:childId/payment` | Body `{ "paidDate": "..." }` |
| GET | `/installments` | Filter `contractId`, `customerId`, `status`. Urut jatuh tempo terdekat |

Nomor identitas pelanggan wajib diisi karena kolomnya unik. Kalau boleh kosong,
dua pelanggan tanpa nomor identitas akan bentrok di index unik.

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

## Finance

Semua di bawah `/api/finance`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/budgets` | Filter `projectId`, `year`. Berisi `realized`, `remaining`, dan `absorption` dalam persen |
| GET, PUT, DELETE | `/budgets/:id` | |
| GET, POST | `/cash-transactions` | Filter `type`, `accountId`, `projectId`, `dateFrom`, `dateTo` |
| GET, PUT, DELETE | `/cash-transactions/:id` | |
| GET | `/cash-flow` | Kas masuk dan keluar per bulan, bawaan enam bulan terakhir. Filter `projectId`, `dateFrom`, `dateTo` |
| GET, POST | `/journal-entries` | Filter `search`, `accountId`, `dateFrom`, `dateTo` |
| GET | `/journal-entries/:id` | Berisi baris debit dan kredit |
| GET | `/ledger` | Buku besar satu akun. Wajib `accountId`, filter `dateFrom`, `dateTo` |

Realisasi anggaran adalah jumlah kas keluar proyek itu pada tahun anggaran,
dihitung saat dibaca. Arus kas juga dijumlahkan dari `cash_transaction`, bulan
tanpa transaksi tetap muncul dengan nilai nol, dan `balance` adalah saldo kas
sampai akhir periode.

Jurnal tidak bisa diubah maupun dihapus. Salah catat diperbaiki dengan jurnal
balik. Buku besar memakai saldo debit dikurangi kredit: saldo awal dihitung
dari semua baris sebelum `dateFrom`, lalu saldo berjalan dihitung dengan window
function sebelum paginasi supaya tetap benar di halaman mana pun.

| Aturan | Kode error |
|---|---|
| Satu proyek satu anggaran per tahun | `budget_year_used` |
| Jurnal minimal dua baris, tiap baris hanya debit atau kredit | `journal_line_single_sided` |
| Total debit harus sama dengan total kredit | `journal_unbalanced` |

## Billing

Semua di bawah `/api/billing`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/invoices` | Filter `search`, `status`, `partyType`, `partyId`, `projectId` |
| GET, PUT, DELETE | `/invoices/:id` | |
| POST | `/invoices/:id/payments` | Body `{ "amount": 1000000 }` |
| GET, POST | `/receivables` | Filter `search`, `status`, `customerId` |
| GET, PUT, DELETE | `/receivables/:id` | |
| POST | `/receivables/:id/payments` | |
| GET, POST | `/payables` | Filter `search`, `status`, `vendorId` |
| GET, PUT, DELETE | `/payables/:id` | |
| POST | `/payables/:id/payments` | |

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
| Pembayaran dicatat dalam transaksi dengan baris dikunci, dan tidak boleh melebihi sisa tagihan | `payment_exceeds_outstanding` |
| Jumlah tagihan tidak boleh diubah di bawah yang sudah dibayar | `amount_below_paid` |
| Data yang sudah menerima pembayaran tidak bisa dihapus | `invoice_has_payment`, `receivable_has_payment`, `payable_has_payment` |
| `invoice.partyId` tidak punya foreign key karena bisa menunjuk pelanggan atau vendor, jadi keberadaannya diperiksa repository | `invoice_party_not_found` |

## HR

Semua di bawah `/api/hr`.

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/employees` | Filter `search`, `employmentType`, `projectId`, `active` |
| GET | `/employees/summary` | Jumlah karyawan aktif per jenis kepegawaian |
| GET, PUT, DELETE | `/employees/:id` | |
| GET, POST | `/attendances` | Filter `employeeId`, `projectId`, `status`, `dateFrom`, `dateTo`. Sudah berisi nama karyawan |
| GET, PUT, DELETE | `/attendances/:id` | Jam ditulis `HH:MM` |
| GET, POST | `/payrolls` | Filter `employeeId`, `period` (`YYYY-MM`), `paid` |
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
| GET, POST | `/saved` | Filter `search`, `regionId`, `buildingProfileId`, `minScore`, serta area peta `minLongitude`, `minLatitude`, `maxLongitude`, `maxLatitude`. Urut skor tertinggi |
| GET, PUT, DELETE | `/saved/:id` | Berisi skor per dimensi. PUT mengganti seluruh skor dimensi |
| GET, POST | `/comparisons` | Body berisi dua sampai lima `savedLocationIds` |
| GET, PUT, DELETE | `/comparisons/:id` | GET berisi skor per dimensi setiap lokasi |

Area peta disaring dengan `point && ST_MakeEnvelope(...)` supaya index GiST
`idx_saved_location_point` terpakai. Skor lokasi dikirim dari hasil analisis
karena mesin skoring belum diputuskan. `savedAt` diperbarui hanya kalau skor
berubah, sesuai catatan prototipe bahwa skor yang tampil adalah skor saat
lokasi disimpan.

| Aturan | Kode error |
|---|---|
| Batas area peta harus lengkap empat nilai dan minimum lebih kecil dari maksimum | `location_bounds_invalid` |
| Satu dimensi sekali per lokasi | `dimension_score_duplicate` |
| Lokasi yang dibandingkan harus milik pengguna dan tidak ganda | `comparison_location_not_found`, `comparison_location_duplicate` |

## Reporting

| Method | Path | Keterangan |
|---|---|---|
| GET, POST | `/api/reports` | Filter `search`, `type`, `format`, `projectId`. Pembuat diambil dari access token |
| GET, DELETE | `/api/reports/:id` | |

Slice ini baru mencatat daftar laporan. Pembuatan berkas PDF, XLSX, atau CSV
belum dikerjakan.

## Yang belum dikerjakan

1. Belum ada pembatasan akses berdasarkan peran. Tabel `role` dan
   `project_member` sudah ada, tapi semua user yang punya access token saat ini
   bisa memakai semua endpoint.
2. Penerimaan barang belum menambah stok secara otomatis. Perlu diputuskan cara
   slice `procurement` meminta `inventory` mencatat mutasi tanpa saling impor.
3. Pembuatan berkas laporan di slice `reporting`.
