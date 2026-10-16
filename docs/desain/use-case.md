# Use Case

Aktor dan use case di bawah diturunkan dari kode: peran dan haknya dari
`backend/internal/shared/access/matrix.go`, use case dari rute API di setiap
`module.go` (224 endpoint) dan halaman di `frontend/src/views`.

## Aktor

| Aktor | Jenis | Keterangan |
|---|---|---|
| Admin | Pengguna internal | Semua modul, termasuk kelola pengguna. Pengguna pertama yang mendaftar otomatis menjadi Admin |
| Manajer Proyek | Pengguna internal | Proyek, analisis lokasi, inventaris; melihat pengadaan dan keuangan |
| Keuangan | Pengguna internal | Keuangan dan tagihan; melihat proyek, penjualan, pengadaan |
| Pengadaan | Pengguna internal | Pengadaan dan inventaris |
| Penjualan | Pengguna internal | Penjualan; melihat proyek dan tagihan |
| SDM | Pengguna internal | Karyawan, absensi, penggajian |
| Pengamat | Pengguna internal | Melihat semua modul tanpa mengubah data, kecuali kelola pengguna |
| Pendaftar baru | Calon pengguna | Sudah punya akun tapi belum diberi peran, jadi belum bisa membuka modul apa pun |
| Pembeli | Pihak luar | Membayar tagihan lewat tautan publik `/bayar/:token` tanpa akun |
| Layanan penilaian | Sistem | Layanan Python yang menghitung skor lokasi (`POST /score`) |
| Google News | Sistem | Sumber berita sekitar lokasi (RSS) |
| Jaringan Ethereum Sepolia | Sistem | Tempat transaksi pembayaran on-chain diverifikasi lewat node RPC |

## Hak akses per peran

`T` = baca dan ubah, `B` = baca saja, `-` = tidak bisa membuka. Hak ini dijaga di
backend oleh middleware akses, dan di frontend tombol ubah disembunyikan untuk
peran yang hanya bisa membaca.

| Modul | Admin | Manajer Proyek | Keuangan | Pengadaan | Penjualan | SDM | Pengamat |
|---|---|---|---|---|---|---|---|
| Proyek | T | T | B | B | B | B | B |
| Analisis Lokasi | T | T | - | - | - | - | B |
| Keuangan | T | B | T | - | - | - | B |
| Tagihan | T | - | T | - | B | - | B |
| Penjualan | T | - | B | - | T | - | B |
| Pengadaan | T | B | B | T | - | - | B |
| Inventaris | T | T | - | T | - | - | B |
| SDM | T | - | - | - | - | T | B |
| Data Master | T | B | B | B | B | B | B |
| Laporan | T | T | T | B | B | B | B |
| Pengguna | T | - | - | - | - | - | - |

## Gambaran umum

Use case dikelompokkan per modul. Garis dari aktor ke modul berarti aktor itu
bisa mengubah data di modul tersebut; akses baca saja ada di tabel di atas.

```mermaid
flowchart LR
    admin["Admin"]
    proc["Pengadaan"]
    pm["Manajer Proyek"]
    fin["Keuangan"]
    sales["Penjualan"]
    hr["SDM"]
    obs["Pengamat"]
    buyer["Pembeli"]

    subgraph app["Sistem BanguninAja"]
        direction TB
        UCusers(["Mengelola pengguna"])
        UCmaster(["Mengelola data master"])
        UCproc(["Mengelola pengadaan"])
        UCinv(["Mengelola inventaris"])
        UCproj(["Mengelola proyek"])
        UCloc(["Menganalisis lokasi"])
        UCrep(["Membuat laporan"])
        UCfin(["Mengelola keuangan"])
        UCbill(["Mengelola tagihan"])
        UCsales(["Mengelola penjualan"])
        UChr(["Mengelola SDM"])
        UCview(["Melihat semua modul"])
        UCpay(["Membayar tagihan on-chain"])
    end

    admin --- UCusers
    admin --- UCmaster
    proc --- UCproc
    proc --- UCinv
    pm --- UCinv
    pm --- UCproj
    pm --- UCloc
    pm --- UCrep
    fin --- UCrep
    fin --- UCfin
    fin --- UCbill
    sales --- UCsales
    hr --- UChr
    obs --- UCview
    buyer --- UCpay
```

Admin juga terhubung ke semua use case lain; garisnya tidak digambar supaya
diagram tetap terbaca.

## Autentikasi dan akun

```mermaid
flowchart LR
    guest["Pendaftar baru"]
    user["Pengguna internal"]
    admin["Admin"]

    subgraph auth["Autentikasi dan akun"]
        direction TB
        reg(["Mendaftar akun"])
        login(["Masuk"])
        refresh(["Memperbarui sesi otomatis"])
        logout(["Keluar"])
        profile(["Mengubah nama"])
        pass(["Mengganti kata sandi"])
        assign(["Memberi peran"])
        toggle(["Menonaktifkan atau mengaktifkan pengguna"])
    end

    guest --- reg
    guest --- login
    user --- login
    user --- logout
    user --- profile
    user --- pass
    admin --- assign
    admin --- toggle
    login -. "include" .-> refresh
    toggle -. "include" .-> revoke(["Mencabut sesi aktif"])
```

Aturan yang terlihat di use case ini:

- Backend menolak perubahan peran atau status yang membuat tidak ada lagi Admin
  aktif (`last_admin`). Tombol Nonaktifkan juga tidak ditampilkan di baris akun
  sendiri.
- Menonaktifkan pengguna mencabut semua refresh token-nya, jadi sesi yang sedang
  terbuka tidak bisa diperpanjang lagi.

## Analisis lokasi

Use case inti produk: menilai kelayakan satu titik untuk satu profil bangunan.

```mermaid
flowchart LR
    pm["Manajer Proyek"]
    obs["Pengamat"]
    scoring["Layanan penilaian"]
    news["Google News"]

    subgraph loc["Analisis lokasi"]
        direction TB
        eval(["Mengevaluasi lokasi"])
        score(["Menghitung skor 5 dimensi"])
        reg(["Melihat regulasi zonasi"])
        newsuc(["Melihat berita sekitar"])
        save(["Menyimpan lokasi kandidat"])
        cmp(["Membandingkan lokasi"])
        map(["Melihat peta risiko"])
        report(["Mengunduh laporan lokasi"])
        weights(["Mengatur bobot profil"])
        list(["Melihat lokasi tersimpan"])
    end

    pm --- eval
    pm --- cmp
    pm --- report
    pm --- weights
    obs --- list
    obs --- map
    eval -. "include" .-> score
    eval -. "include" .-> save
    eval -. "include" .-> reg
    eval -. "include" .-> newsuc
    score --- scoring
    map --- scoring
    newsuc --- news
```

Regulasi dan berita ditampilkan sebagai konteks dan **tidak ikut dihitung** dalam
skor. Regulasi adalah syarat sah atau tidak, sehingga tidak boleh dikompensasi
oleh skor pasar yang tinggi.

## Proyek dan keuangan

```mermaid
flowchart LR
    pm["Manajer Proyek"]
    fin["Keuangan"]

    subgraph proj["Proyek"]
        direction TB
        p1(["Membuat dan mengubah proyek"])
        p2(["Mengubah status proyek"])
        p3(["Mencatat progres tahap"])
        p4(["Menyusun RAB"])
        p5(["Mencatat izin"])
        p6(["Mengatur anggota tim proyek"])
    end

    subgraph money["Keuangan"]
        direction TB
        f1(["Menetapkan anggaran tahunan"])
        f2(["Mencatat transaksi kas"])
        f3(["Membuat jurnal umum"])
        f4(["Melihat buku besar"])
        f5(["Melihat arus kas per bulan"])
        f6(["Membuat jurnal otomatis"])
    end

    pm --- p1
    pm --- p2
    pm --- p3
    pm --- p4
    pm --- p5
    pm --- p6
    fin --- f1
    fin --- f2
    fin --- f3
    fin --- f4
    fin --- f5
    f2 -. "include" .-> f6
```

Progres proyek dihitung server dari rata-rata progres tahap, dan total RAB dari
volume kali harga satuan. Pengguna tidak mengisi angka turunan ini sendiri.

## Tagihan dan pembayaran on-chain

```mermaid
flowchart LR
    fin["Keuangan"]
    buyer["Pembeli"]
    chain["Jaringan Ethereum Sepolia"]

    subgraph bill["Tagihan"]
        direction TB
        b1(["Membuat invoice"])
        b2(["Mencatat piutang dan hutang"])
        b3(["Mencatat pembayaran manual"])
        b4(["Membuat permintaan bayar on-chain"])
        b5(["Membatalkan permintaan bayar"])
        b6(["Membuka halaman bayar"])
        b7(["Mengirim hash transaksi"])
        b8(["Memverifikasi transaksi"])
        b9(["Mencatat pembayaran otomatis"])
    end

    fin --- b1
    fin --- b2
    fin --- b3
    fin --- b4
    fin --- b5
    buyer --- b6
    buyer --- b7
    b7 -. "include" .-> b8
    b8 -. "include" .-> b9
    b8 --- chain
```

## Operasional

```mermaid
flowchart LR
    sales["Penjualan"]
    proc["Pengadaan"]
    pm["Manajer Proyek"]
    hr["SDM"]

    subgraph sell["Penjualan"]
        direction TB
        s1(["Mengelola unit"])
        s2(["Mencatat prospek"])
        s3(["Mengubah prospek menjadi pelanggan"])
        s4(["Membuat kontrak"])
        s5(["Mencatat cicilan"])
    end

    subgraph buy["Pengadaan"]
        direction TB
        o1(["Mengelola vendor"])
        o2(["Mengajukan permintaan pembelian"])
        o3(["Menyetujui atau menolak permintaan"])
        o4(["Membuat dan mengirim pesanan"])
        o5(["Mencatat penerimaan barang"])
        o6(["Menambah stok otomatis"])
    end

    subgraph stock["Inventaris"]
        direction TB
        i1(["Mengelola material dan gudang"])
        i2(["Melihat stok menipis"])
        i3(["Mengelola aset dan peralatan"])
    end

    subgraph people["SDM"]
        direction TB
        h1(["Mengelola karyawan"])
        h2(["Mencatat absensi"])
        h3(["Menyusun dan membayar gaji"])
    end

    sales --- s1
    sales --- s2
    sales --- s3
    sales --- s4
    sales --- s5
    proc --- o1
    proc --- o2
    proc --- o3
    proc --- o4
    proc --- o5
    o5 -. "include" .-> o6
    proc --- i1
    pm --- i1
    pm --- i2
    pm --- i3
    hr --- h1
    hr --- h2
    hr --- h3
```
