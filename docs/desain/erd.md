# ERD (Entity Relationship Diagram)

Diturunkan dari skema database yang berjalan (PostgreSQL 18 + PostGIS), bukan
dari rancangan awal. Daftar foreign key diambil langsung dari `pg_constraint`,
jadi setiap garis di bawah ada di database sungguhan.

| | Jumlah |
|---|---|
| Tabel aplikasi | 55 (ditambah `spatial_ref_sys` milik PostGIS) |
| Foreign key | 91 |
| Slice backend yang memiliki tabel | 14 |

Satu diagram berisi 55 tabel tidak terbaca, jadi ERD dipecah per domain. Tabel
dari domain lain hanya ditampilkan namanya supaya garis relasinya tetap kelihatan.

## Cara membaca

| Notasi | Arti |
|---|---|
| `PK` | Kunci utama. Semua tabel memakai UUID v4 yang dibuat aplikasi |
| `FK` | Foreign key |
| `UK` | Unik |
| `\|\|--o{` | Satu ke banyak, sisi kiri wajib ada |
| `\|o--o{` | Satu ke banyak, kolom FK boleh kosong |
| `\|o--o\|` | Satu ke satu, FK boleh kosong dan unik |

Kolom `created_at` dan `updated_at` ada di semua tabel dan tidak ditulis ulang.
Uang disimpan sebagai `bigint` rupiah penuh, kuantitas barang sebagai `numeric`.

## 1. Akses dan data master

Peran utama pengguna ada di `users.role_id`. Peran per proyek ada di
`project_member`, sehingga satu orang bisa menjadi manajer di satu proyek dan
pengamat di proyek lain.

```mermaid
erDiagram
    role |o--o{ users : "peran utama"
    users ||--o{ refresh_tokens : "sesi login"
    users ||--o{ project_member : "anggota"
    role ||--o{ project_member : "peran di proyek"
    project ||--o{ project_member : "punya anggota"
    region |o--o{ region : "induk wilayah"
    account |o--o{ account : "induk akun"

    role {
        uuid id PK
        varchar code UK "admin, manajer_proyek, keuangan, ..."
        varchar name
        text description
    }
    users {
        uuid id PK
        varchar name
        varchar email UK
        varchar password_hash "bcrypt"
        uuid role_id FK
        boolean active
        timestamptz last_login_at
    }
    refresh_tokens {
        uuid id PK
        uuid user_id FK
        varchar token_hash UK "token disimpan sebagai hash"
        timestamptz expires_at
        timestamptz revoked_at
    }
    project_member {
        uuid id PK
        uuid user_id FK
        uuid project_id FK
        uuid role_id FK
        varchar position
    }
    region {
        uuid id PK
        varchar code UK
        varchar name
        varchar type "province, regency, city, district"
        uuid parent_id FK
    }
    unit_of_measure {
        uuid id PK
        varchar code UK
        varchar name
    }
    account {
        uuid id PK
        varchar code UK "bagan akun"
        varchar name
        varchar type "asset, liability, equity, revenue, expense"
        uuid parent_id FK
    }
```

## 2. Proyek

```mermaid
erDiagram
    region |o--o{ project : "lokasi"
    project ||--o{ project_phase : "tahap"
    project ||--o{ budget_item : "RAB"
    project_phase |o--o{ budget_item : "bobot tahap"
    unit_of_measure ||--o{ budget_item : "satuan"
    project ||--o{ project_progress : "riwayat progres"
    project ||--o{ daily_report : "laporan harian"
    users |o--o{ daily_report : "dibuat oleh"
    project ||--o| construction_contract : "kontrak kontraktor"
    vendor ||--o{ construction_contract : "kontraktor"
    project ||--o{ payment_term : "termin"
    payable |o--o| payment_term : "hutang termin"
    users |o--o{ payment_term : "diajukan oleh"
    project ||--o{ contract_change : "adendum"
    users |o--o{ contract_change : "diusulkan atau diputuskan oleh"
    project ||--o{ permit : "izin"
    project ||--o{ project_member : "anggota"
    project |o--o{ report : "laporan"
    users ||--o{ report : "dibuat oleh"

    project {
        uuid id PK
        varchar code UK
        varchar name
        uuid region_id FK
        varchar type "teks bebas, mis. Perumahan, Ruko"
        varchar status "planning, ongoing, on_hold, completed, cancelled"
        int progress "rata-rata berbobot nilai RAB per tahap"
        date start_date
        date target_end_date
        bigint contract_value
    }
    project_phase {
        uuid id PK
        uuid project_id FK
        varchar name
        int sort_order
        date start_date
        date target_end_date
        int progress
        varchar status
    }
    budget_item {
        uuid id PK
        uuid project_id FK
        uuid phase_id FK "boleh kosong"
        varchar code
        varchar description
        numeric volume
        uuid unit_of_measure_id FK
        bigint unit_price
        bigint total "volume x harga satuan, dihitung server"
    }
    project_progress {
        uuid id PK
        uuid project_id FK
        date recorded_on "unik per proyek"
        int progress
    }
    daily_report {
        uuid id PK
        uuid project_id FK
        date report_date "unik per proyek"
        varchar weather_morning
        varchar weather_afternoon
        int workers
        varchar equipment
        text work_done
        text issues
        uuid created_by FK
    }
    construction_contract {
        uuid id PK
        uuid project_id FK "UK, satu per proyek"
        uuid vendor_id FK
        varchar number "nomor SPK"
        bigint value
        numeric retention_percent "0 sampai 10"
    }
    payment_term {
        uuid id PK
        uuid project_id FK
        varchar kind "progress, retention"
        int required_progress
        numeric percent "persen nilai kontrak"
        bigint gross_amount
        bigint retention_amount
        uuid payable_id FK "UK"
        date claimed_on
        int claimed_progress
        uuid claimed_by FK
    }
    contract_change {
        uuid id PK
        uuid project_id FK
        int sequence "unik per proyek"
        varchar description
        bigint value_change "negatif untuk pekerjaan kurang"
        int days_change
        varchar status "proposed, approved, rejected"
        uuid proposed_by FK
        uuid decided_by FK
        date decided_on
    }
    permit {
        uuid id PK
        uuid project_id FK
        varchar type "teks bebas, mis. PKKPR, PBG, SLF"
        varchar number
        date issued_date
        date valid_until
        varchar status "submitted, issued, expired, rejected"
    }
    report {
        uuid id PK
        varchar name
        varchar type
        varchar scope
        uuid project_id FK
        date period_start
        date period_end
        varchar format "pdf, xlsx, csv"
        uuid created_by FK
    }
```

## 3. Keuangan dan tagihan

Buku besar (`journal_entry`, `journal_line`) tidak terikat proyek, sedangkan
anggaran dan transaksi kas terikat proyek. `invoice.party_id` bisa menunjuk
pelanggan atau vendor (dibedakan `party_type`), dan
`onchain_payment_request.target_id` bisa menunjuk invoice atau piutang
(dibedakan `target_kind`). Karena polimorfik, keduanya tidak punya foreign key
di database dan dijaga oleh service.

```mermaid
erDiagram
    project ||--o{ budget : "anggaran tahunan"
    account ||--o{ cash_transaction : "akun"
    project |o--o{ cash_transaction : "proyek"
    cash_transaction |o--o| journal_entry : "jurnal otomatis"
    journal_entry ||--|{ journal_line : "baris debit/kredit"
    account ||--o{ journal_line : "akun"
    project |o--o{ invoice : "proyek"
    customer ||--o{ receivable : "pelanggan"
    contract |o--o{ receivable : "kontrak"
    invoice |o--o| receivable : "invoice piutang"
    project |o--o{ receivable : "proyek"
    vendor ||--o{ payable : "vendor"
    purchase_order |o--o{ payable : "pesanan"
    project |o--o{ payable : "proyek"
    invoice |o--o{ billing_payment : "pembayaran"
    receivable |o--o{ billing_payment : "pembayaran"
    payable |o--o{ billing_payment : "pembayaran"
    users |o--o{ billing_payment : "dicatat oleh"
    users |o--o{ onchain_payment_request : "dibuat oleh"

    budget {
        uuid id PK
        uuid project_id FK
        int year
        bigint value
        text note
    }
    cash_transaction {
        uuid id PK
        date date
        varchar type "in, out"
        uuid account_id FK
        uuid project_id FK
        bigint amount
        text note
    }
    journal_entry {
        uuid id PK
        varchar number UK
        date date
        text note
        varchar source "kas untuk jurnal otomatis, selain itu bebas"
        uuid cash_transaction_id FK, UK
    }
    journal_line {
        uuid id PK
        uuid journal_entry_id FK
        uuid account_id FK
        bigint debit
        bigint kredit
    }
    invoice {
        uuid id PK
        varchar number UK
        varchar party_type "customer, vendor"
        uuid party_id "polimorfik, tanpa FK"
        uuid project_id FK
        date due_date
        bigint amount
        bigint paid_amount
        varchar status "not_due, due, paid, overdue"
    }
    receivable {
        uuid id PK
        uuid customer_id FK
        uuid project_id FK
        uuid contract_id FK
        uuid invoice_id FK, UK
        varchar reference
        date due_date
        bigint amount
        bigint paid_amount
        varchar status "not_due, due, paid, overdue"
    }
    payable {
        uuid id PK
        uuid vendor_id FK
        uuid project_id FK
        uuid purchase_order_id FK
        varchar reference
        date due_date
        bigint amount
        bigint paid_amount
        varchar status "not_due, due, paid, overdue"
    }
    billing_payment {
        uuid id PK
        uuid invoice_id FK "tepat satu dari tiga FK terisi"
        uuid receivable_id FK
        uuid payable_id FK
        bigint amount
        timestamptz paid_at
        varchar method "transfer, tunai, cek, lainnya, onchain"
        varchar reference "nomor bukti atau hash transaksi"
        text note
        uuid created_by FK
    }
    onchain_payment_request {
        uuid id PK
        varchar target_kind "invoice, receivable"
        uuid target_id "polimorfik, tanpa FK"
        bigint amount_idr
        varchar amount_wei
        bigint idr_per_eth "kurs saat permintaan dibuat"
        varchar receiver "alamat dompet perusahaan"
        bigint chain_id "11155111 = Sepolia"
        varchar token UK "dipakai di tautan publik /bayar/:token"
        varchar status "pending, paid, failed, expired, cancelled"
        varchar tx_hash UK
        varchar payer_address
        text failure_reason
        uuid created_by FK
        timestamptz expires_at
        timestamptz paid_at
    }
```

## 4. Penjualan

Membuat kontrak mengubah unit dari `available` menjadi `reserved`. Kontrak
`paid` membuat unit `sold`, kontrak `cancelled` mengembalikan unit ke
`available`. Satu prospek (`lead`) paling banyak menjadi satu pelanggan.

```mermaid
erDiagram
    project ||--o{ unit : "unit dijual"
    project |o--o{ lead : "minat proyek"
    customer |o--o| lead : "prospek jadi pelanggan"
    customer ||--o{ contract : "pembeli"
    unit ||--o{ contract : "unit"
    contract ||--o{ installment : "cicilan"
    contract |o--o{ receivable : "piutang"

    customer {
        uuid id PK
        varchar name
        varchar contact
        varchar email
        varchar identity_number UK "NIK"
        text address
    }
    unit {
        uuid id PK
        varchar code UK
        uuid project_id FK
        varchar unit_type
        numeric area_sqm
        bigint price
        varchar status "available, reserved, sold, on_hold"
    }
    lead {
        uuid id PK
        varchar name
        varchar contact
        uuid project_id FK
        varchar source
        varchar stage "new, interested, negotiating, won, cancelled"
        timestamptz last_contacted_at
        uuid customer_id FK, UK
    }
    contract {
        uuid id PK
        varchar number UK
        uuid customer_id FK
        uuid unit_id FK
        varchar type "cash, mortgage, installment"
        bigint value
        date date
        varchar status "draft, active, paid, cancelled"
    }
    installment {
        uuid id PK
        uuid contract_id FK
        int installment_number
        date due_date
        bigint amount
        date paid_date
        varchar status "not_due, due, paid, overdue"
    }
```

## 5. Pengadaan, inventaris, dan aset

Setiap baris penerimaan barang yang diterima mencatat `stock_movement` masuk ke
gudang tujuan. Kalau satuan barang di pesanan berbeda dengan satuan material,
pencatatan stok dilewati dan alasannya disimpan di `stock_skip_reason`.

```mermaid
erDiagram
    project ||--o{ purchase_request : "untuk proyek"
    users ||--o{ purchase_request : "pemohon"
    purchase_request ||--|{ purchase_request_item : "barang diminta"
    purchase_request |o--o{ purchase_order : "dasar pesanan"
    vendor ||--o{ purchase_order : "vendor"
    project ||--o{ purchase_order : "untuk proyek"
    purchase_order ||--|{ purchase_order_item : "barang dipesan"
    purchase_order ||--o{ goods_receipt : "penerimaan"
    warehouse ||--o{ goods_receipt : "gudang tujuan"
    goods_receipt ||--|{ goods_receipt_item : "barang diterima"
    purchase_order_item ||--o{ goods_receipt_item : "baris pesanan"
    stock_movement |o--o| goods_receipt_item : "stok masuk"
    material ||--o{ purchase_request_item : ""
    material ||--o{ purchase_order_item : ""
    unit_of_measure ||--o{ material : "satuan"
    material ||--o{ stock : "saldo"
    warehouse ||--o{ stock : "saldo"
    material ||--o{ stock_movement : "mutasi"
    warehouse |o--o{ stock_movement : "asal atau tujuan"
    project |o--o{ warehouse : "gudang proyek"
    asset |o--o{ equipment : "alat dari aset"
    project |o--o{ equipment : "dipakai di proyek"

    vendor {
        uuid id PK
        varchar code UK
        varchar name
        varchar category
        varchar tax_number "NPWP"
        int payment_term_days
        varchar rating "new, good, fair, poor"
        boolean active
    }
    purchase_request {
        uuid id PK
        varchar number UK
        uuid project_id FK
        uuid requester_id FK
        date date
        varchar status "draft, submitted, approved, rejected, completed"
    }
    purchase_request_item {
        uuid id PK
        uuid purchase_request_id FK
        uuid material_id FK
        numeric quantity
        uuid unit_of_measure_id FK
    }
    purchase_order {
        uuid id PK
        varchar number UK
        uuid vendor_id FK
        uuid project_id FK
        uuid purchase_request_id FK
        date date
        date due_date
        bigint value
        varchar status "draft, sent, partially_received, completed, cancelled"
    }
    purchase_order_item {
        uuid id PK
        uuid purchase_order_id FK
        uuid material_id FK
        numeric quantity
        uuid unit_of_measure_id FK
        bigint unit_price
    }
    goods_receipt {
        uuid id PK
        varchar number UK
        uuid purchase_order_id FK
        uuid warehouse_id FK
        date date
        varchar condition "good, partially_damaged, broken"
    }
    goods_receipt_item {
        uuid id PK
        uuid goods_receipt_id FK
        uuid purchase_order_item_id FK
        numeric accepted_quantity
        numeric rejected_quantity
        uuid stock_movement_id FK, UK
        varchar stock_skip_reason "kosong atau unit_mismatch"
    }
    material {
        uuid id PK
        varchar code UK
        varchar name
        varchar category
        uuid unit_of_measure_id FK
        numeric minimum_stock
        bigint last_price
    }
    warehouse {
        uuid id PK
        varchar code UK
        varchar name
        uuid project_id FK
    }
    stock {
        uuid id PK
        uuid material_id FK
        uuid warehouse_id FK
        numeric quantity
    }
    stock_movement {
        uuid id PK
        date date
        varchar type "in, out, transfer, adjustment"
        uuid material_id FK
        numeric quantity
        uuid source_warehouse_id FK
        uuid target_warehouse_id FK
        varchar reference
    }
    asset {
        uuid id PK
        varchar code UK
        varchar name
        varchar category
        date acquisition_date
        bigint acquisition_value
        bigint accumulated_depreciation
        int useful_life_months
        varchar status "in_use, maintenance, for_sale, disposed"
    }
    equipment {
        uuid id PK
        varchar code UK
        varchar name
        uuid asset_id FK
        uuid project_id FK
        int operating_hours
        date next_service_date
        varchar status "operating, idle, maintenance, broken"
    }
```

## 6. SDM

```mermaid
erDiagram
    project |o--o{ employee : "ditempatkan di"
    users |o--o{ employee : "akun login"
    employee ||--o{ attendance : "absensi"
    employee ||--o{ payroll : "gaji"

    employee {
        uuid id PK
        varchar identity_number UK
        varchar name
        varchar position
        uuid project_id FK
        uuid user_id FK
        varchar employment_type "permanent, contract, daily, subcontractor"
        date joined_date
        date left_date
        bigint base_salary
    }
    attendance {
        uuid id PK
        uuid employee_id FK
        date date
        varchar check_in_time
        varchar check_out_time
        varchar status "present, permitted, sick, absent, holiday"
    }
    payroll {
        uuid id PK
        uuid employee_id FK
        varchar period "YYYY-MM"
        bigint basic_pay
        bigint allowance
        bigint deduction
        bigint net_pay
        boolean paid
    }
```

## 7. Analisis lokasi

`regulation` tidak punya foreign key. Tabel ini dicocokkan dengan titik lokasi
secara spasial (`ST_DWithin` pada kolom `point`), bukan lewat kunci. Baris
dengan `is_simulated = true` adalah data simulasi untuk wilayah yang belum punya
data RDTR asli.

```mermaid
erDiagram
    users ||--o{ saved_location : "disimpan oleh"
    region |o--o{ saved_location : "wilayah"
    building_profile |o--o{ saved_location : "profil bangunan"
    project |o--o{ saved_location : "kandidat lokasi proyek"
    saved_location ||--o{ dimension_score : "skor per dimensi"
    dimension ||--o{ dimension_score : "dimensi"
    users ||--o{ comparison : "dibuat oleh"
    comparison ||--|{ comparison_item : "isi perbandingan"
    saved_location ||--o{ comparison_item : "lokasi"
    building_profile ||--o{ weight : "bobot"
    dimension ||--o{ weight : "bobot"

    saved_location {
        uuid id PK
        uuid user_id FK
        varchar name
        uuid region_id FK
        uuid building_profile_id FK
        uuid project_id FK
        geometry point "Point, SRID 4326"
        numeric area_sqm
        bigint land_price_per_sqm "dari ZNT ATR/BPN"
        int score "0-100"
        numeric flood_index
        numeric earthquake_index
        jsonb risk_flags
        timestamptz saved_at
    }
    dimension_score {
        uuid id PK
        uuid saved_location_id FK
        uuid dimension_id FK
        int value "0-100"
    }
    comparison {
        uuid id PK
        uuid user_id FK
        varchar name
    }
    comparison_item {
        uuid id PK
        uuid comparison_id FK
        uuid saved_location_id FK
        int sort_order
    }
    dimension {
        uuid id PK
        varchar code UK "fisik_lingkungan, infrastruktur, demografi_sosial, pasar_kompetisi, finansial_proyek"
        varchar name
        int sort_order
    }
    building_profile {
        uuid id PK
        varchar code UK "housing, hospital, mall, entertainment"
        varchar name
    }
    weight {
        uuid id PK
        uuid building_profile_id FK
        uuid dimension_id FK
        int percent "sama dengan scoring/model/model.json, total 100 per profil"
    }
    regulation {
        uuid id PK
        varchar rtr_id
        varchar region_code
        varchar zone_name
        varchar sub_zone_name
        varchar district
        varchar village
        geometry point
        numeric kdb
        numeric klb
        numeric kdh
        varchar gsb
        varchar max_height
        varchar source
        boolean is_simulated
    }
```
