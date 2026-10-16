# Acceptance Test

Skenario penerimaan ditulis dari sudut pandang pengguna per peran, dengan
format Given/When/Then. Setiap skenario dijalankan di browser pada aplikasi
lokal dengan data demo, kecuali yang ditandai API.

**Status** menunjukkan siapa yang sudah menjalankannya dan bagaimana:

| Status | Arti |
|---|---|
| Lulus (UI) | Dijalankan tim pengembang lewat browser, hasil sesuai |
| Lulus (API) | Hasil sesuai, tetapi baru dibuktikan lewat API atau test otomatis, belum diklik lewat UI |
| Belum | Belum dijalankan pada siklus ini |

Kolom **Disetujui** diisi pengguna di luar tim pengembang (anggota tim lain atau
dosen) setelah mengulang skenario. Skenario baru dianggap diterima kalau sudah
disetujui.

## Skenario

| ID | Peran | Skenario | Status | Bukti | Disetujui |
|---|---|---|---|---|---|
| AT-01 | Semua | Masuk dan keluar | Lulus (UI) | Masuk dengan akun demo, sesi pulih setelah muat ulang | |
| AT-02 | Manajer Proyek | Melihat progres proyek berbobot dan kurva S | Lulus (UI) | RSIA: 46% berbobot RAB, 26 poin di belakang rencana, tooltip kurva | |
| AT-03 | Manajer Proyek | Memperbarui progres satu tahap | Lulus (UI) | Form terbuka hanya untuk satu tahap, tertutup setelah disimpan | |
| AT-04 | Manajer Proyek | Mengaitkan item RAB ke tahap | Lulus (API) | System test ST-02; pilihan Tahap di form terlihat di UI | |
| AT-05 | Manajer Proyek | Menulis laporan harian | Lulus (UI) | Tanggal bawaan hari ini, laporan muncul di atas, ringkasan berubah | |
| AT-06 | Manajer Proyek | Ditolak saat menulis laporan kedua di hari yang sama atau untuk besok | Lulus (API) | System test ST-03 (409 dan 422) | |
| AT-07 | Manajer Proyek | Mengevaluasi titik lokasi dan mengunduh laporan PDF | Lulus (API) | System test ST-04: skor 77, PDF 9 kB | |
| AT-08 | Manajer Proyek | Diberi peringatan saat skor berasal dari cadangan | Lulus (API) | Layanan penilaian dimatikan: `is_simulated` dan `scoreSimulated` bernilai true; test render label | |
| AT-09 | Penjualan | Tidak bisa menghapus unit yang terikat kontrak | Lulus (UI) | Unit Terjual dan Dipesan tanpa tombol Hapus | |
| AT-10 | Pengadaan | Membuat dan mengirim pesanan pembelian | Belum | Baru tombol Buat pesanan yang diperiksa | |
| AT-11 | Keuangan | Membaca transaksi terbaru dan arus kas | Lulus (UI) | Jumlah bertanda dan berwarna, tanggal satu baris | |
| AT-12 | Admin | Tidak bisa menonaktifkan akun sendiri dari halaman Pengguna | Lulus (UI) | Baris (Anda) tanpa Nonaktifkan | |
| AT-13 | Pembeli | Membayar tagihan lewat tautan on-chain | Belum | | |
| AT-14 | Semua | Memakai aplikasi di ponsel | Lulus (UI) | 390 px: sepuluh halaman dan tab Lapangan tanpa geser ke samping | |
| AT-15 | Manajer Proyek | Mengajukan termin kontraktor yang progresnya tercapai | Lulus (UI) | Proyek uji: konfirmasi dua langkah, hutang Rp 190 jt (20% dari Rp 1 M dikurangi retensi 5%) tercatat, jatuh tempo 30 hari, kontrak terkunci; 375 px tanpa geser halaman | |
| AT-16 | Keuangan | Melihat termin tanpa bisa mengajukannya | Lulus (API) | Test render: tanpa tombol Ajukan dan Tambah termin, ada tautan ke Tagihan; endpoint pengajuan butuh hak tulis Proyek | |
| AT-17 | Admin | Menyetujui adendum pekerjaan kurang | Lulus (UI) | RSIA Adendum 2 (−Rp 350 jt): nilai berjalan turun ke Rp 24,4 M, termin 5 dihitung ulang, jejak disetujui tercatat; dikembalikan ke usulan setelah uji | |

## Rincian skenario

### AT-02 Melihat progres proyek berbobot dan kurva S

- **Given** proyek RSIA punya lima tahap dan item RAB yang dikaitkan ke tahapnya
- **When** Manajer Proyek membuka tab Ringkasan proyek
- **Then** KPI Progres proyek menyebut "berbobot nilai RAB" dan posisinya terhadap rencana
- **And** kartu Kurva S menampilkan garis rencana, garis realisasi, penanda hari ini, dan label status jadwal
- **And** setiap tahap menampilkan bobotnya; tahap tanpa item RAB ditandai "bobot 0%, belum ada item RAB"

### AT-03 Memperbarui progres satu tahap

- **Given** tahap Struktur berprogres 55%
- **When** Manajer Proyek menekan Perbarui progres pada tahap itu
- **Then** hanya satu form progres terbuka, dengan nilai saat ini dan kursor di kolomnya
- **When** ia menyimpan nilai baru
- **Then** form tertutup, muncul "Progres disimpan", dan progres proyek serta kurva S ikut diperbarui

### AT-05 Menulis laporan harian

- **Given** proyek sedang berjalan dan belum ada laporan hari ini
- **When** Manajer Proyek membuka tab Lapangan dan menekan Tulis laporan
- **Then** tanggal terisi hari ini dan tanggal setelah hari ini tidak bisa dipilih
- **When** ia mengisi cuaca, tenaga kerja, pekerjaan, dan kendala, lalu menyimpan
- **Then** laporan muncul paling atas, hari hujan ditandai, kendala tampil di bawah pekerjaan, dan ringkasan 30 hari bertambah

### AT-06 Laporan ganda atau untuk hari yang belum terjadi

- **Given** laporan hari ini sudah ada
- **When** laporan lain untuk hari ini dikirim
- **Then** sistem menolak dengan pesan "Laporan untuk tanggal ini sudah ada. Ubah laporan tersebut"
- **When** laporan untuk besok dikirim
- **Then** sistem menolak dengan pesan bahwa tanggal belum terjadi

### AT-07 Mengevaluasi titik lokasi

- **Given** layanan penilaian menyala
- **When** Manajer Proyek memilih titik di Bandung dengan profil Perumahan dan proyek terkait
- **Then** skor 0 sampai 100 dan lima dimensi tampil dalam waktu kurang dari 3 detik
- **And** lokasi tersimpan terkait ke proyek dan laporan PDF-nya bisa diunduh

### AT-08 Peringatan skor cadangan

- **Given** layanan penilaian sedang mati
- **When** Manajer Proyek mengevaluasi titik
- **Then** hasil diberi peringatan "Skor simulasi" dan KPI skor menyebut bukan hasil data GIS
- **And** lokasi itu bertanda Skor simulasi di daftar lokasi proyek, tabel perbandingan, dan laporan PDF

### AT-09 Unit yang terikat kontrak

- **Given** unit berstatus Terjual atau Dipesan
- **When** staf Penjualan membuka daftar unit
- **Then** unit itu tidak punya tombol Ubah atau Hapus, sedangkan unit Tersedia dan Ditahan punya

### AT-12 Akun sendiri di halaman Pengguna

- **Given** Admin membuka halaman Pengguna
- **When** ia melihat baris akunnya sendiri, ditandai (Anda)
- **Then** tidak ada tombol Nonaktifkan di baris itu, sedangkan pengguna lain punya

### AT-13 Membayar tagihan lewat tautan on-chain

- **Given** staf Keuangan membuat permintaan bayar on-chain untuk sebuah invoice
- **When** Pembeli membuka tautan `/bayar/:token` dan membayar dengan dompet di jaringan Sepolia
- **Then** halaman menunggu konfirmasi jaringan, lalu menampilkan Lunas dengan tautan Etherscan
- **And** invoice di modul Tagihan berstatus lunas dengan pembayaran metode onchain

### AT-15 Mengajukan termin kontraktor

- **Given** proyek punya kontrak kontraktor dengan retensi 5% dan termin yang syarat progresnya sudah tercapai
- **When** Manajer Proyek membuka tab Keuangan proyek dan menekan Ajukan pada termin berstatus Siap diajukan
- **Then** muncul konfirmasi yang menyebut nilai yang akan dibayarkan setelah retensi, dan fokus pindah ke Batal
- **When** ia menekan Ya, ajukan
- **Then** termin berstatus Diajukan dengan tanggal dan progres saat diajukan, hutang ke kontraktor tercatat di Tagihan dengan jatuh tempo sesuai termin bayar vendor, dan tombol Ubah kontrak hilang karena kontrak terkunci
