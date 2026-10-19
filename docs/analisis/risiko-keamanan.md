# Analisis Risiko dan Keamanan

Dokumen ini berisi daftar risiko proyek dengan rencana RMMM (mitigasi,
monitoring, dan manajemen), analisis keamanan mengikuti OWASP Top 10 (2021),
serta pertimbangan etika dan hukum. Setiap pernyataan merujuk ke kode atau
pull request supaya bisa diperiksa.

## 1. Daftar risiko

Probabilitas adalah perkiraan kemungkinan terjadi. Dampak memakai skala
Pressman: **1** katastrofik, **2** kritis, **3** marginal, **4** dapat
diabaikan. Tabel diurutkan dari probabilitas tertinggi dan dampak terberat.
Kolom **Terjadi** mencatat risiko yang sudah benar-benar terjadi selama
proyek.

| ID | Risiko | Kategori | Prob. | Dampak | Terjadi |
|---|---|---|---|---|---|
| R1 | Kontribusi anggota tim timpang: sebagian besar commit berasal dari satu orang | Staf | 90% | 2 | Ya, sekitar 92% commit di `main` dari satu orang (dua akun) |
| R2 | Demo online bergantung pada satu laptop; laptop mati, tidur, atau disk penuh membuat demo mati | Lingkungan pengembangan | 70% | 2 | Ya, 15 Oktober: disk C penuh, disk virtual Docker rusak, backend mati |
| R3 | Sumber data RDTR (zonasi) tidak tersedia secara nasional | Teknologi | 100% | 3 | Ya, sumber satu-satunya mati sejak September |
| R4 | Data simulasi (regulasi, skor cadangan) terbaca sebagai data asli | Dampak bisnis | 40% | 2 | Hampir: skor cadangan sempat tersimpan tanpa tanda (diperbaiki #64) |
| R5 | Layanan luar (Google News, node RPC Sepolia, petak OpenStreetMap) lambat atau mati | Teknologi | 60% | 3 | Ya, Google News kadang galat; evaluasi sempat tertahan sampai 10 detik (diperbaiki #58) |
| R6 | Kehilangan data karena tidak ada backup database terjadwal | Lingkungan pengembangan | 30% | 1 | Hampir: volume Docker selamat setelah disk penuh |
| R7 | Lingkup bertambah sementara tenggat tetap | Proses | 60% | 3 | Ya, modul ERP P2 ikut dibangun |
| R8 | Pengetahuan sistem terpusat di satu orang (bus factor 1) | Staf | 70% | 2 | Berkaitan dengan R1 |
| R9 | Pelanggaran lisensi data (GADM, ZNT ATR/BPN, OSM) | Dampak bisnis | 20% | 2 | Tidak; bundle tidak diunggah ke repo publik |
| R10 | Kebocoran data pribadi pelanggan (NIK, kontak) | Pelanggan | 15% | 1 | Tidak |
| R11 | Kesalahan perhitungan yang memengaruhi keputusan (progres, skor, jatuh tempo) | Ukuran produk | 40% | 2 | Ya: progres rata-rata biasa (#61), tanggal UTC (#63) |
| R12 | Perubahan yang belum lulus test masuk ke `main` | Proses | 30% | 3 | Tidak; branch `main` belum diproteksi |

### RMMM

| ID | Mitigasi (mencegah) | Monitoring (tanda awal) | Manajemen (kalau terjadi) |
|---|---|---|---|
| R1 | Pembagian kerja per vertical di `PROJECT_HANDOFF.md`; setiap tugas lewat branch dan PR | Jumlah commit dan PR per anggota per minggu (`git shortlog`) | Bagi ulang tugas kecil yang jelas ke anggota yang belum berkontribusi; pasangkan dengan anggota yang paham kodenya |
| R2 | Disk Docker dipindah ke drive dengan ruang lebih; skrip `start-online.ps1` menyalakan ulang semua layanan | Ruang kosong drive C dan status `docker info` sebelum demo | Hentikan Docker dan WSL, kosongkan disk, nyalakan ulang; kalau disk virtual rusak, purge lalu migrate dan seed ulang (data demo bisa dibuat ulang) |
| R3 | Regulasi dipisah dari skor; data simulasi diberi `is_simulated` dan label di UI | Proporsi titik yang jatuh ke data simulasi | Pakai data asli 11 kota yang sudah diambil; jelaskan keterbatasan di laporan dan presentasi |
| R4 | Tanda simulasi dari API sampai laporan PDF (#64); kontrak API tidak menyembunyikan data simulasi | Review setiap jalur yang mengembalikan nilai cadangan | Evaluasi ulang lokasi bertanda simulasi setelah layanan pulih |
| R5 | Batas waktu per layanan, berita paralel dan di-cache (#58), skor cadangan bertanda | Waktu respons evaluasi di system test (batas 3 detik) | Kartu berita kosong tanpa menggagalkan evaluasi; ganti node RPC lewat `ONCHAIN_RPC_URL` |
| R6 | Volume Docker terpisah dari container | Tanggal backup terakhir | Jadwalkan `pg_dump` ke penyimpanan terpisah; data demo bisa dibuat ulang dengan `seed-demo` |
| R7 | Prioritas P1 dan P2 di `docs/analisis/kelengkapan-domain.md`; fitur P2 tidak menunda P1 | Daftar PR terbuka dan sisa hari | Tunda P2; fokus ke alur yang didemokan |
| R8 | Dokumen desain, pengujian, dan handoff di repo; PR dengan deskripsi lengkap | Siapa saja yang bisa menjalankan aplikasi dari README | Sesi pengenalan kode untuk anggota lain sebelum presentasi |
| R9 | Bundle GIS hanya lewat Google Drive tim; atribusi OSM di peta | Isi repo publik sebelum push | Hapus berkas dari riwayat Git dan ganti dengan petunjuk unduh |
| R10 | Akses data pelanggan hanya untuk peran Penjualan (tulis) serta Keuangan dan Pengamat (baca); HTTPS di demo | Log akses dan daftar pengguna aktif | Nonaktifkan akun terkait (sesi langsung dicabut), ganti `JWT_ACCESS_SECRET`, beri tahu pihak terdampak |
| R11 | Aturan perhitungan di service dengan unit test; integration test untuk yang bergantung pada database | CI di setiap PR; system test sebelum demo | Perbaiki, tambah test yang mereproduksi, hitung ulang data lewat backfill migrasi |
| R12 | CI lima job di setiap PR | Status CI sebelum merge | Pemilik repo mengaktifkan branch protection: wajib PR dan wajib kelima job lulus |

## 2. Analisis keamanan (OWASP Top 10 2021)

Status: **Terkendali** (kontrol ada dan diuji), **Sebagian** (ada celah yang
diketahui), **Diperbaiki** (celah ditemukan dan ditutup pada siklus ini).

| Kategori | Status | Kontrol yang ada | Celah yang tersisa |
|---|---|---|---|
| A01 Broken Access Control | Sebagian | Matriks peran per modul diperiksa backend di setiap request (`shared/access`, middleware `Authorization`), termasuk status aktif; frontend menyembunyikan tombol ubah; rute publik hanya `/auth` dan `/pay/:token` dengan token acak. Di modul Proyek, peran selain Admin, Pengamat, dan Keuangan hanya bisa membuka proyek yang timnya memuat mereka (#82), dengan middleware yang menolak peran tak dikenal | Modul lain yang menyaring per proyek (keuangan, tagihan, pengadaan, laporan) belum mengikuti keanggotaan tim |
| A02 Cryptographic Failures | Sebagian | Kata sandi bcrypt; refresh token 32 byte acak, disimpan sebagai hash SHA-256; JWT HS256 dengan rahasia dari env; HTTPS di demo lewat Vercel dan Cloudflare | NIK pelanggan dan karyawan masih disimpan tanpa enkripsi di database (tampilannya sudah tersamar untuk peran yang hanya membaca) |
| A03 Injection | Terkendali | Query lewat GORM dengan parameter; tidak ada SQL yang disusun dari input pengguna; React meng-escape keluaran | |
| A04 Insecure Design | Terkendali | Peta transisi status di service; pembayaran on-chain: hash unik, baris dikunci, chain id dan batas waktu diperiksa; data simulasi ditandai (#64). Pembayaran QRIS: webhook Xendit diverifikasi dengan `x-callback-token` (perbandingan waktu-konstan), lalu server tetap menanyakan status dan nominal ke Xendit sebelum mencatat; baris dikunci sehingga webhook ganda tidak mencatat dua kali; simulasi hanya aktif dengan kunci uji | |
| A05 Security Misconfiguration | Diperbaiki | CORS dari daftar asal di env; galat ke klien tanpa detail internal; header keamanan API (#68); CSP dan header keamanan frontend di Vercel dari `frontend/vercel.json` | Trusted proxies belum diatur |
| A06 Vulnerable and Outdated Components | Sebagian | Job CI `Security (govulncheck)` gagal kalau ada kerentanan Go yang terjangkau kode; `npm audit --omit=dev --audit-level=high` jalan di job frontend. Pemeriksaan pertama menemukan 10 kerentanan terjangkau (standard library Go 1.27.1, `golang.org/x/net`, `quic-go`), semua ditutup dengan Go 1.27.2 dan upgrade modul | 1 kerentanan *high* di `source-map-js` (alat build dan test, tidak ikut ke produksi). Image Docker yang sudah terlanjur dibangun perlu dibangun ulang |
| A07 Identification and Authentication Failures | Diperbaiki | Batas 10 percobaan login per email per 15 menit (#68); panjang kata sandi 8 sampai 72 byte (#62); rotasi refresh token dengan deteksi pemakaian ulang (diuji integration test); akun nonaktif ditolak di request berikutnya | Batas per email bisa dipakai orang lain untuk mengunci satu akun selama 15 menit |
| A08 Software and Data Integrity Failures | Sebagian | CI lima job di setiap PR, termasuk integration test PostGIS dan pemeriksaan kerentanan | Branch `main` belum diproteksi, jadi merge tanpa CI hijau masih mungkin |
| A09 Security Logging and Monitoring Failures | Sebagian | Log terstruktur (`slog`); pembayaran mencatat `created_by`; laporan harian, pengajuan termin, dan keputusan adendum mencatat pelakunya; perubahan peran dan status pengguna tercatat di `user_audit` dalam transaksi yang sama dan tampil di halaman Pengguna (#83) | Belum ada peringatan otomatis untuk kejadian mencurigakan, misalnya banyak login gagal |
| A10 Server-Side Request Forgery | Terkendali | Permintaan keluar hanya ke alamat dari konfigurasi (layanan penilaian, node RPC, Google News); proxy tile memvalidasi nama layer dengan regex dan koordinat sebagai angka | |

### Pemetaan ke OWASP Top 10:2025

Analisis di atas memakai edisi 2021. Edisi 2025 menata ulang urutannya,
melebur SSRF ke dalam A01, memperluas komponen rentan menjadi rantai pasok
perangkat lunak, dan menambah kategori penanganan kondisi luar biasa.

| Kategori 2025 | Padanan 2021 | Status | Catatan |
|---|---|---|---|
| A01:2025 Broken Access Control | A01, ditambah A10 SSRF | Sebagian | Hak akses per proyek di modul Proyek sudah ditegakkan; modul lain belum. SSRF terkendali |
| A02:2025 Security Misconfiguration | A05 | Diperbaiki | |
| A03:2025 Software Supply Chain Failures | A06, diperluas | Sebagian | `go.sum` dan `package-lock.json` mengunci versi dan checksum; `govulncheck` dan `npm audit` di CI (#74). Belum ada SBOM dan branch `main` belum diproteksi |
| A04:2025 Cryptographic Failures | A02 | Sebagian | |
| A05:2025 Injection | A03 | Terkendali | |
| A06:2025 Insecure Design | A04 | Terkendali | Aturan bisnis keuangan dijaga di service dan constraint database, misalnya termin yang tidak bisa diajukan sebelum progresnya tercapai |
| A07:2025 Authentication Failures | A07 | Diperbaiki | |
| A08:2025 Software or Data Integrity Failures | A08 | Sebagian | |
| A09:2025 Security Logging and Alerting Failures | A09 | Sebagian | Jejak audit akses sudah ada; peringatan (*alerting*) belum |
| A10:2025 Mishandling of Exceptional Conditions | Baru | Terkendali | Galat ditranslasi ke kode dan pesan aman tanpa detail internal (`apperror`, `database.Translate`). Perubahan berlapis berjalan dalam transaksi, jadi galat di tengah jalan membatalkan semuanya. Layanan luar yang gagal (penilaian, berita) tidak menggagalkan evaluasi lokasi, dan hasil cadangannya ditandai simulasi. Frontend punya error boundary |

### Rencana tindak lanjut keamanan

| Prioritas | Tindakan | Menutup |
|---|---|---|
| Tinggi | Aktifkan branch protection di `main` (pemilik repo) | A08 |
| Sedang | Perluas hak akses per proyek ke modul keuangan, tagihan, pengadaan, dan laporan | A01 |
| Sedang | Peringatan otomatis untuk login gagal beruntun dan perubahan peran Admin | A09 |
| Rendah | Enkripsi kolom NIK di database; tampilannya sudah tersamar | A02 |
| Rendah | Trusted proxies di backend | A05 |

## 3. Etika dan hukum

| Aspek | Pertimbangan | Penerapan |
|---|---|---|
| Pelindungan data pribadi (UU No. 27 Tahun 2022) | NIK, kontak, dan alamat pelanggan adalah data pribadi | Hanya peran yang perlu yang bisa membuka modul Penjualan; NIK hanya utuh untuk peran yang boleh mengubah data itu, peran lain dan laporan gaji melihat empat digit terakhir; demo memakai data fiktif; enkripsi NIK di database masih rencana |
| Lisensi data | GADM melarang redistribusi tanpa izin; ZNT milik ATR/BPN; OSM berlisensi ODbL | Bundle GIS tidak diunggah ke repo publik; atribusi OpenStreetMap tampil di peta |
| Kejujuran data | Data simulasi dan skor cadangan bisa menyesatkan keputusan investasi | Ditandai `is_simulated` dari API sampai UI dan PDF; regulasi tidak ikut dihitung dalam skor |
| Transparansi model | Bobot skor berasal dari regresi, bukan pilihan bebas | Bobot per profil ditampilkan; tiap dimensi punya penjelasan satu kalimat; laporan PDF menjelaskan cara membaca skor |
| Sistem pendukung keputusan, bukan pengambil keputusan | Skor tidak mencakup faktor sosial, politik, dan legal penuh | Laporan dan UI menyebut skor sebagai bahan pertimbangan; regulasi ditampilkan terpisah sebagai syarat |
| Pembayaran on-chain | Risiko kehilangan dana nyata | Hanya jaringan uji Sepolia dengan kurs tetap simulasi; backend tidak memegang kunci privat |
| Pembayaran QRIS lewat gerbang pembayaran | Kunci rahasia gerbang pembayaran bisa dipakai menarik atau mengembalikan dana | Kunci hanya di env backend, tidak pernah dikirim ke frontend; demo memakai kunci mode uji Xendit atau simulator lokal; batas Rp 10 juta per transaksi mengikuti aturan QRIS |

## 4. Perbaikan selama siklus ini

| PR | Perbaikan | Kategori |
|---|---|---|
| #57 | Unit test paket token, cookie, dan kata sandi (0% menjadi lebih dari 93%) | A02, A07 |
| #58 | Batas waktu dan cache untuk layanan berita | R5 |
| #61 | Progres proyek berbobot nilai RAB | R11 |
| #62 | Kata sandi lebih dari 72 byte ditolak 422, bukan 500 | A07 |
| #63 | Tanggal hari ini dihitung dalam WIB | R11 |
| #64 | Skor cadangan ditandai simulasi sampai laporan PDF | A04, R4 |
| #66 | Integration test ke PostGIS sungguhan di CI | R11, R12 |
| #68 | Batas percobaan login dan header keamanan API | A05, A07 |
| #95 | NIK tersamar untuk peran yang hanya membaca dan di laporan gaji, pencarian NIK hanya untuk peran penulis; CSP dan header keamanan frontend di Vercel | A01, A02, A05 |
