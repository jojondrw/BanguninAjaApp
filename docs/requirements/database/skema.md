# Skema Database

Skema ini diturunkan dari livewire di `BanguninAja_livewire`, lalu dirapikan.
Livewire hanya gambaran tampilan, jadi bentuk datanya masih datar dan banyak
mengulang nama. Di sini semuanya dinormalkan.

## Aturan umum

| Hal | Keputusan |
|---|---|
| Kunci utama | UUID v4, dibuat di aplikasi lewat `entity.Base` |
| Waktu | `created_at` dan `updated_at` wajib, diisi otomatis |
| Uang | `bigint` dalam rupiah penuh. Tidak memakai float supaya tidak ada pembulatan aneh, dan rupiah tidak memakai sen |
| Jumlah barang | `numeric(14,2)` karena volume material bisa pecahan |
| Nama tabel | Bahasa Inggris bentuk tunggal, contoh `project`, `budget_item` |
| Status | Disimpan sebagai `varchar` dan dijaga `CHECK`, bukan enum Postgres, supaya menambah pilihan baru tidak perlu mengubah tipe |

## Slice dan entity

| Slice | Entity |
|---|---|
| `master` | Region, UnitOfMeasure, Account |
| `auth` | Role, User, RefreshToken, ProjectMember |
| `project` | Project, ProjectPhase, BudgetItem, Permit |
| `inventory` | Material, Warehouse, Stock, StockMovement |
| `procurement` | Vendor, PurchaseRequest, PurchaseRequestItem, PurchaseOrder, PurchaseOrderItem, GoodsReceipt, GoodsReceiptItem |
| `asset` | Asset, Equipment |
| `sales` | Customer, Unit, Lead, Contract, Installment |
| `finance` | Budget, CashTransaction, JournalEntry, JournalLine |
| `billing` | Invoice, Receivable, Payable |
| `hr` | Employee, Attendance, Payroll |
| `scoring` | Dimension, BuildingProfile, Weight |
| `location` | SavedLocation, DimensionScore, Comparison, ComparisonItem |
| `reporting` | Report |

Totalnya 47 tabel, 65 foreign key, 67 check, 13 unique, dan 216 index.

## Yang ditambahkan karena livewire belum mencerminkan data asli

| Tambahan | Alasan |
|---|---|
| Slice `master` | Livewire menulis kota, satuan, dan nama akun sebagai teks bebas. Kalau dibiarkan, "Jakarta Pusat" dan "Jakarta pusat" jadi dua hal berbeda |
| `PurchaseRequestItem`, `PurchaseOrderItem`, `GoodsReceiptItem` | Livewire hanya menyimpan jumlah item sebagai angka, misalnya "12 item". Barang apa saja yang dipesan tidak tercatat, padahal itu inti pengadaan |
| `Material` dan `Warehouse` | Halaman stok menulis nama material dan gudang sebagai teks. Stok harus melekat pada material dan gudang yang jelas |
| `Customer` | Livewire menulis nama pembeli di kontrak. Satu orang bisa punya beberapa kontrak, jadi datanya dipisah |
| `JournalLine` | Livewire menampilkan jurnal sebagai baris datar. Pembukuan berpasangan butuh satu jurnal berisi banyak baris debit dan kredit |
| `ProjectPhase` | Tab jadwal di livewire hanya memakai tanggal mulai dan selesai proyek. Proyek konstruksi punya tahapan, dan progres dihitung per tahap |
| `Budget`, `CashTransaction` | Livewire menghitung anggaran dari progres fisik dan menampilkan arus kas per bulan. Keduanya harus punya data sendiri |
| `BuildingProfile`, `Dimension`, `Weight` | Tabel bobot penilaian di halaman pengaturan tadinya angka mati di layar |
| `Role`, `ProjectMember` | Livewire menyimpan peran dan proyek sebagai teks pada pengguna. Satu orang bisa punya peran berbeda di proyek berbeda |

## Yang sengaja tidak dijadikan kolom

| Di livewire | Alasan |
|---|---|
| Umur piutang, misalnya "31 sampai 60 hari" | Dihitung dari `due_date` dan tanggal hari ini. Kalau disimpan, besoknya sudah salah |
| Nilai buku aset | Selisih nilai perolehan dan akumulasi penyusutan |
| Progres proyek dalam persen | Tetap disimpan untuk sementara, tapi nanti diturunkan dari tahapan |
| Arus kas per bulan | Hasil penjumlahan `cash_transaction`, bukan data mentah |
| Total RAB | Penjumlahan `budget_item` |

Aturannya: yang bisa dihitung ulang kapan saja tidak disimpan, supaya tidak ada
dua versi kebenaran.

## Kolom lokasi

`saved_location.point` bertipe `geometry(Point,4326)`. Nilainya ditulis lewat tipe
`geo.Point` di Go yang mengubah lintang bujur menjadi format yang dimengerti
PostGIS, dan membacanya kembali dari format biner PostGIS.

Pencarian berdasarkan area memakai index GiST `idx_saved_location_point`. Sudah
diperiksa dengan `EXPLAIN`, hasilnya Index Scan, bukan Seq Scan.

## Perilaku saat data induk dihapus

| Pola | Dipakai untuk |
|---|---|
| `CASCADE` | Anak yang tidak punya arti tanpa induknya, misalnya `budget_item` terhadap `project`, atau `installment` terhadap `contract` |
| `RESTRICT` | Data acuan yang dipakai catatan lain, misalnya `vendor`, `material`, `account`. Menghapusnya akan ditolak selama masih dipakai |
| `SET NULL` | Kaitan yang boleh kosong, misalnya `equipment.project_id` saat alatnya belum ditugaskan |

## Contoh aturan yang dijaga database

Ini beberapa yang sudah diuji dan benar-benar menolak data salah:

| Aturan | Constraint |
|---|---|
| Progres proyek 0 sampai 100 | `chk_project_progress` |
| Status proyek hanya lima pilihan | `chk_project_status` |
| Satu baris jurnal hanya boleh debit atau kredit, tidak dua duanya | `chk_journal_line_single_sided` |
| Penyusutan tidak boleh melebihi nilai perolehan | `chk_asset_depreciation_within_cost` |
| Satu karyawan satu absensi per hari | `uq_attendance_employee_date` |
| Stok tidak boleh minus | `chk_stock_quantity` |
| Mutasi harus punya gudang asal atau tujuan | `chk_stock_movement_direction` |
| Pembayaran tidak boleh melebihi tagihan | `chk_invoice_amount` |

## Menjalankan migrasi

```bash
docker compose exec backend go run ./cmd/migrate
```

Urutannya: pasang ekstensi, buat tabel per slice, lalu pasang index dan
constraint. Constraint dipasang setelah semua tabel ada, karena sebuah slice bisa
menunjuk tabel milik slice lain.

Semua perintahnya aman diulang. Index memakai `IF NOT EXISTS`, dan constraint
dibungkus pemeriksaan ke `pg_constraint` lebih dulu.
