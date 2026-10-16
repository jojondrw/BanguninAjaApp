# State Diagram dan Activity Diagram

State diagram menggambarkan siklus hidup satu dokumen. Setiap panah adalah
transisi yang diizinkan service backend; transisi lain ditolak dengan galat
409 atau 422, sehingga data tidak bisa melompati tahap. Activity diagram
menggambarkan pekerjaan nyata pengguna yang melintasi beberapa modul.

Sumber aturan transisi:

| Dokumen | Kode |
|---|---|
| Proyek | `projectTransitions` di `project/service.go` |
| Tahap proyek | `phaseStatus` di `project/service.go` |
| Termin kontraktor | `newPaymentScheduleResponse`, `claimBlocker`, `releasableRetention` di `project/payment_term.go` |
| Adendum kontrak | `DecideContractChange` di `project/contract_change.go` |
| Permintaan dan pesanan pembelian | `requestTransitions`, `orderTransitions`, `statusAfterReceipt` di `procurement/service.go` |
| Kontrak dan unit | `contractTransitions`, `unitStatusAfterContract` di `sales/service.go` |
| Jatuh tempo tagihan dan cicilan | `duedate.Status` di `shared/duedate` |
| Permintaan bayar on-chain | `billing/onchain.go` |

## State diagram

### Proyek

Proyek `completed` atau `cancelled` tidak bisa diubah lagi (`project_closed`).

```mermaid
stateDiagram-v2
    [*] --> planning
    planning --> ongoing
    planning --> cancelled
    ongoing --> on_hold
    ongoing --> completed
    ongoing --> cancelled
    on_hold --> ongoing
    on_hold --> cancelled
    completed --> [*]
    cancelled --> [*]
```

### Tahap proyek

Status tahap tidak dipilih pengguna. Status diturunkan dari progres yang diisi.
Progres proyek adalah rata-rata progres tahap yang dibobot nilai RAB tiap tahap,
atau bobot sama kalau belum ada item RAB yang ditautkan ke tahap.

```mermaid
stateDiagram-v2
    [*] --> not_started
    not_started --> in_progress : progres 1 sampai 99
    not_started --> completed : progres 100
    in_progress --> completed : progres 100
    in_progress --> not_started : progres 0
    completed --> in_progress : progres diturunkan
```

### Termin kontraktor

Status termin tidak disimpan sebagai kolom. Status dihitung dari progres
berbobot proyek dan dari hutang yang tertaut di modul Tagihan, jadi selalu
sesuai keadaan terakhir.

```mermaid
stateDiagram-v2
    [*] --> planned : termin disusun
    planned --> ready : progres mencapai syarat dan termin sebelumnya sudah diajukan
    ready --> planned : progres turun di bawah syarat
    ready --> claimed : Manajer Proyek mengajukan, hutang dibuat
    claimed --> paid : Keuangan melunasi hutang
    claimed --> ready : hutang dihapus di Tagihan sebelum dibayar
    planned --> [*] : dihapus
    ready --> [*] : dihapus
    paid --> [*]
```

### Adendum kontrak kontraktor

Manajer Proyek mengusulkan, hanya Admin yang memutuskan, dan keputusan bersifat
final. Hanya adendum yang disetujui yang mengubah nilai kontrak berjalan.

```mermaid
stateDiagram-v2
    [*] --> proposed : Manajer Proyek mengusulkan
    proposed --> approved : Admin menyetujui, nilai berjalan berubah
    proposed --> rejected : Admin menolak
    proposed --> [*] : usulan dihapus
    approved --> [*]
    rejected --> [*]
```

### Permintaan pembelian dan pesanan pembelian

```mermaid
stateDiagram-v2
    state "Permintaan pembelian" as PR {
        [*] --> draft
        draft --> submitted : diajukan
        submitted --> approved : disetujui
        submitted --> rejected : ditolak
        approved --> completed : dipenuhi pesanan
        rejected --> [*]
        completed --> [*]
    }
    state "Pesanan pembelian" as PO {
        [*] --> po_draft
        po_draft --> sent : dikirim ke vendor
        po_draft --> po_cancelled
        sent --> po_cancelled
        sent --> partially_received : penerimaan sebagian
        sent --> po_completed : semua barang diterima
        partially_received --> partially_received : penerimaan berikutnya
        partially_received --> po_completed : sisa barang diterima
        po_cancelled --> [*]
        po_completed --> [*]
    }
```

Status pesanan `po_draft`, `po_cancelled`, dan `po_completed` di diagram adalah
`draft`, `cancelled`, dan `completed` di database; awalan hanya untuk membedakan
dari status permintaan. Perpindahan ke `partially_received` dan `completed`
hanya terjadi lewat pencatatan penerimaan barang, tidak bisa dipilih manual.

### Kontrak penjualan dan unit

Kontrak dan unit berubah bersama. Unit hanya bisa dipakai satu kontrak yang
masih berjalan, dan kontrak hanya bisa `paid` kalau semua cicilannya lunas
(`contract_unpaid`).

```mermaid
stateDiagram-v2
    state "Kontrak" as C {
        [*] --> draft
        draft --> active
        draft --> cancelled
        active --> paid : semua cicilan lunas
        active --> cancelled
        paid --> [*]
        cancelled --> [*]
    }
    state "Unit" as U {
        [*] --> available
        available --> on_hold : ditahan manual
        on_hold --> available : dilepas manual
        available --> reserved : kontrak dibuat
        reserved --> sold : kontrak paid
        reserved --> available : kontrak cancelled atau dihapus
        sold --> [*]
    }
```

### Jatuh tempo invoice, piutang, hutang, dan cicilan

Status ini dihitung dari tanggal jatuh tempo dan tanggal hari ini setiap kali
data dibaca, jadi tidak pernah basi.

```mermaid
stateDiagram-v2
    [*] --> not_due
    not_due --> due : 7 hari sebelum jatuh tempo
    due --> overdue : lewat jatuh tempo
    not_due --> paid : lunas
    due --> paid : lunas
    overdue --> paid : lunas
    paid --> [*]
```

### Permintaan bayar on-chain

```mermaid
stateDiagram-v2
    [*] --> pending : staf keuangan membuat tautan
    pending --> paid : transaksi terverifikasi di Sepolia
    pending --> failed : transaksi tidak cocok atau gagal dicatat
    pending --> expired : lewat batas waktu
    pending --> cancelled : dibatalkan staf
    paid --> [*]
    failed --> [*]
    expired --> [*]
    cancelled --> [*]
```

Status `expired` juga ditampilkan tanpa menunggu baris diperbarui: permintaan
`pending` yang sudah lewat `expires_at` langsung dianggap `expired` saat dibaca
(`effectiveStatus`).

## Activity diagram

### Memilih lokasi proyek

Proses bisnis inti produk: dari beberapa titik kandidat sampai satu lokasi
dipilih untuk proyek.

```mermaid
flowchart TD
    start((Mulai)) --> profil[Pilih profil bangunan: perumahan, rumah sakit, mal, hiburan]
    profil --> titik[Klik titik kandidat di peta]
    titik --> eval[Sistem menghitung skor 5 dimensi dan risiko]
    eval --> konteks[Sistem menampilkan regulasi zonasi dan berita sekitar]
    konteks --> layak{Regulasi mengizinkan dan skor memadai?}
    layak -- Tidak --> lagi{Masih ada kandidat lain?}
    layak -- Ya --> simpan[Simpan sebagai lokasi kandidat]
    simpan --> lagi
    lagi -- Ya --> titik
    lagi -- Tidak --> ada{Ada kandidat tersimpan?}
    ada -- Tidak --> selesai((Selesai))
    ada -- Ya --> banding[Bandingkan kandidat berdampingan]
    banding --> pilih[Pilih lokasi terbaik]
    pilih --> kait[Kaitkan lokasi ke proyek]
    kait --> laporan[Unduh laporan lokasi PDF untuk rapat investasi]
    laporan --> selesai
```

### Pengadaan material sampai pembayaran vendor

Melintasi dua peran dan empat modul. Manajer Proyek hanya punya hak baca di
modul Pengadaan, jadi permintaan pembelian dibuat dan disetujui oleh peran
Pengadaan (atau Admin).

```mermaid
flowchart TD
    subgraph pg["Pengadaan"]
        a1[Buat permintaan pembelian untuk proyek]
        a1b[Ajukan permintaan]
        a2{Setujui permintaan?}
        a3[Tolak permintaan]
        a4[Buat pesanan ke vendor dari permintaan]
        a5[Kirim pesanan ke vendor]
        a6[Catat penerimaan barang di gudang]
        a7{Semua barang sudah diterima?}
    end
    subgraph sys["Sistem"]
        s1[Tambah stok material di gudang]
        s2[Status pesanan partially_received]
        s3[Status pesanan completed]
    end
    subgraph keu["Keuangan"]
        k1[Catat hutang ke vendor dari pesanan]
        k2[Catat pembayaran hutang]
        k3[Hutang lunas]
    end

    a1 --> a1b --> a2
    a2 -- Tidak --> a3
    a2 -- Ya --> a4
    a4 --> a5 --> a6 --> s1 --> a7
    a7 -- Belum --> s2 --> a6
    a7 -- Sudah --> s3 --> k1 --> k2 --> k3
```

### Pembayaran kontraktor per termin sampai retensi

Manajer Proyek mengesahkan progres dengan mengajukan termin, Keuangan yang
membayar. Keduanya tidak bisa saling menggantikan karena hak aksesnya berbeda:
pengajuan butuh hak tulis Proyek, pembayaran butuh hak tulis Tagihan.

```mermaid
flowchart TD
    subgraph pm["Manajer Proyek"]
        m1[Atur kontrak kontraktor: vendor, nilai, retensi]
        m2[Susun termin: syarat progres dan persen nilai]
        m3[Perbarui progres tahap dan laporan harian]
        m4[Ajukan termin yang siap]
        m6[Usulkan adendum pekerjaan tambah atau kurang]
        m5[Lepas retensi]
    end
    subgraph sys["Sistem"]
        s1[Hitung progres berbobot RAB]
        s2{Progres mencapai syarat dan termin sebelumnya sudah diajukan?}
        s3[Buat hutang sebesar nilai termin dikurangi retensi]
        s4[Kunci kontrak dari perubahan]
        s5{Proyek selesai dan semua termin diajukan?}
        s6[Buat hutang sebesar retensi yang ditahan]
    end
    subgraph adm["Admin"]
        a1{Setujui adendum?}
    end
    subgraph keu["Keuangan"]
        k1[Bayar hutang termin di Tagihan]
        k2[Bayar hutang retensi]
    end

    m1 --> m2 --> m3 --> s1 --> s2
    s2 -- Belum --> m3
    s2 -- Sudah --> m4 --> s3 --> s4 --> k1
    k1 --> s5
    s5 -- Belum --> m3
    s5 -- Sudah --> m5 --> s6 --> k2
    m3 -.-> m6 --> a1
    a1 -- Ya --> s1
```

### Penjualan unit sampai lunas

```mermaid
flowchart TD
    l1[Catat prospek dari pameran, iklan, atau rujukan] --> l2[Hubungi dan negosiasi]
    l2 --> l3{Prospek membeli?}
    l3 -- Tidak --> l4[Tandai prospek cancelled]
    l3 -- Ya --> l5[Ubah prospek menjadi pelanggan]
    l5 --> c1[Buat kontrak untuk unit tersedia]
    c1 --> u1[Sistem: unit menjadi reserved]
    u1 --> c2[Aktifkan kontrak]
    c2 --> c3[Susun jadwal cicilan]
    c3 --> p2[Catat pembayaran cicilan yang diterima]
    p2 --> p5{Masih ada cicilan belum lunas?}
    p5 -- Ya --> p2
    p5 -- Tidak --> c4[Ubah status kontrak menjadi paid]
    c4 --> u2[Sistem: unit menjadi sold]
```

Pembayaran on-chain tersedia untuk invoice dan piutang di modul Tagihan, tetapi
belum terhubung ke cicilan di modul Penjualan. Kalau pembeli membayar piutang
kontrak secara on-chain, cicilan yang bersangkutan masih perlu ditandai lunas di
modul Penjualan.
