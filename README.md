# BanguninAja

[![CI](https://github.com/jojondrw/BanguninAjaApp/actions/workflows/ci.yml/badge.svg)](https://github.com/jojondrw/BanguninAjaApp/actions/workflows/ci.yml)

Sistem pendukung keputusan pemilihan lokasi bangunan, digabung dengan modul ERP
untuk mengelola proyek yang sudah berjalan.

| Bagian | Teknologi |
|---|---|
| Frontend | React 19, Vite, TypeScript, Tailwind v4, TanStack Query, Zustand |
| Backend | Go 1.27, Gin, GORM, arsitektur vertical slice |
| Database | PostgreSQL 18 dengan PostGIS |

## Menjalankan di laptop

Yang perlu terpasang: Docker Desktop dan Node.js. Go tidak perlu, karena
jalannya di dalam container.

### 1. Siapkan berkas .env

Berkas ini berisi kunci rahasia, jadi sengaja tidak ikut masuk Git. Salin
contohnya:

```bash
cp .env.example .env
```

Lalu isi `JWT_ACCESS_SECRET` dengan teks acak. Cara cepat membuatnya:

```bash
node -e "console.log(require('crypto').randomBytes(48).toString('base64url'))"
```

Tempelkan hasilnya ke baris `JWT_ACCESS_SECRET` di `.env`. Nilai yang lain boleh
dibiarkan apa adanya untuk keperluan ngoding.

### 2. Nyalakan database dan backend

```bash
docker compose up -d
```

Perintah ini menyalakan PostgreSQL berikut PostGIS, lalu backend dengan hot
reload. Setiap berkas Go yang disimpan langsung dibangun ulang otomatis.

### 3. Buat tabelnya

```bash
docker compose exec backend go run ./cmd/migrate
```

Dijalankan sekali di awal, dan diulang setiap kali menarik perubahan baru.
Aman dijalankan berkali-kali.

### 3a. Isi data awal

```bash
docker compose exec backend go run ./cmd/seed
```

Mengisi wilayah, satuan, bagan akun, dimensi dan bobot penilaian. Aman diulang.

### Data demo (opsional, untuk presentasi)

```bash
docker compose exec -e DEMO_USER_PASSWORD=<kata-sandi-demo> backend go run ./cmd/seed-demo
```

Dijalankan setelah `cmd/migrate` dan `cmd/seed`. Mengisi cerita pengembang
properti yang utuh: 4 proyek (Griya Asri Bandung berjalan, Ruko Sentra Makassar
perencanaan, RSIA Cirebon berjalan, Cluster Lama Depok selesai) lengkap dengan
tahap, RAB, izin, anggaran, kas 10 bulan terakhir (tiap transaksi punya jurnal
`KAS-…`), unit, prospek, kontrak dan cicilan, pengadaan sampai penerimaan barang
dan stok, tagihan/piutang/hutang dengan riwayat pembayaran, karyawan dengan
absensi dan penggajian, aset, alat, serta lokasi tersimpan dan satu perbandingan.
Semua kode berawalan `DEMO-`, tanggal dihitung dari hari ini, dan aman diulang
(data yang sudah ada dilewati).

- Kontrak kontraktor dan termin ada di RSIA Cirebon, Griya Asri, dan Cluster
  Lama Depok. Termin diajukan lewat service supaya hutangnya tercatat di
  Tagihan seperti pemakaian biasa, lalu tanggal pengajuannya dimundurkan ke
  hari progres mencapai syarat (dari riwayat progres). Hutang yang jatuh
  temponya sudah lewat dibuat lunas.
- Cluster Lama Depok sudah selesai, semua terminnya lunas, dan retensinya
  sengaja belum dilepas supaya tombol Lepas retensi bisa diperagakan. Kontrak
  dan terminnya ditulis langsung ke database karena proyek yang sudah selesai
  tidak bisa diubah lewat service.
- RSIA punya satu adendum disetujui (setelah termin ketiga diajukan, jadi termin
  lama tetap memakai nilai awal) dan satu usulan yang menunggu Admin. Griya Asri
  punya satu adendum ditolak. Adendum dicek per proyek dan uraian, jadi database
  yang sudah punya kontrak demo tetap mendapatkannya.
- Laporan harian memakai pola cuaca tetap musim hujan (pagi cerah atau berawan,
  hujan siang atau sore) supaya hasilnya sama setiap dibuat ulang.

- Lokasi tersimpan dan permintaan pembelian dimiliki satu akun. Urutannya:
  `--owner email@anda` kalau diisi; kalau tidak, akun `demo@banguninaja.id`
  (dibuat otomatis bila `DEMO_USER_PASSWORD` diisi); kalau tidak juga, akun
  pertama yang terdaftar.
- Kalau `SCORE_SERVICE_URL` bisa dihubungi, lokasi dinilai lewat layanan
  penilaian sungguhan. Kalau tidak, nilainya contoh dan catatannya diawali
  "Nilai contoh".
- Hapus semua data demo (hanya yang berawalan `DEMO-`, data lain tidak
  disentuh):

  ```bash
  docker compose exec backend go run ./cmd/seed-demo --reset-demo
  ```

  Jalankan lagi perintah pertama sesudahnya supaya tanggal demo kembali segar.

### 3b. Isi data RDTR (opsional)

Berkas `rdtr_*.json` tidak ada di repo ini — ada di folder `data/raw/rdtr` repo
data BanguninAja (atau Google Drive tim). Pasang foldernya ke container:

```bash
docker compose run --rm -v "/path/ke/BanguninAja/data/raw/rdtr:/rdtr" backend go run ./cmd/import-rdtr -dir /rdtr
```

Tanpa langkah ini, nilai regulasi tetap muncul tapi ditandai simulasi.

### 3c. Nyalakan layanan penilaian lokasi (disarankan)

Unduh `scoring-bundle.zip` (239 MB) dari Google Drive tim "AOL SWE", ekstrak ke
`scoring/bundle`, lalu:

```bash
docker compose --profile scoring up -d scoring
```

dan isi `SCORE_SERVICE_URL=http://scoring:8090` serta `SCORE_SERVICE_TIMEOUT=15s`
di `.env`. Tanpa layanan ini skor tetap muncul tapi dari data simulasi
(penjelasannya diakhiri "(data simulasi)"), dan lapisan peta tidak tersedia.
Cara lain dan cara membangun bundle ada di `scoring/README.md`.

### 4. Nyalakan tampilannya

```bash
cd frontend
npm install
npm run dev
```

Buka `http://localhost:5173`, buat akun, lalu masuk. Pakai nama host yang sama
untuk tampilan dan API (`localhost` dengan `localhost`, atau `127.0.0.1` dengan
`VITE_API_URL=http://127.0.0.1:8080/api`): cookie sesi memakai SameSite=Lax, jadi
kalau hostnya beda sesi hilang setiap halaman dimuat ulang.

## Memeriksa isi database

```bash
docker compose exec database psql -U $POSTGRES_USER -d $POSTGRES_DB
```

## Menjalankan test

| Test | Perintah | Butuh |
|---|---|---|
| Unit test backend | `docker compose exec backend go test ./...` | Container backend |
| Integration test backend | lihat di bawah | Database `bangunin_aja_test` |
| Test frontend | `cd frontend && npm test` | Node.js |
| Test layanan penilaian | `cd scoring && python -m pytest` | Python dan dependensinya |

Integration test berjalan ke PostgreSQL + PostGIS sungguhan dan menolak jalan
kalau nama database tidak berakhiran `_test`, supaya database kerja tidak
tersentuh. Buat database-nya sekali, lalu jalankan:

```bash
docker compose exec database createdb -U $POSTGRES_USER bangunin_aja_test
docker compose exec -e POSTGRES_DB=bangunin_aja_test backend go test -tags integration -v ./internal/integration/
```

Di CI, keempatnya berjalan otomatis di setiap pull request, ditambah pemeriksaan
kerentanan dependensi Go (`govulncheck`) dan npm (dependensi produksi).

## Dokumentasi

Desain sistem (UML, ERD, deployment) ada di `docs/desain`, rencana dan hasil
pengujian di `docs/pengujian`. Alasan di balik tiap keputusan ada di
`docs/requirements`.

| Berkas | Isi |
|---|---|
| `backend/arsitektur.md` | Aturan vertical slice, pembagian lapisan, isi folder shared |
| `backend/konvensi-kode.md` | Gaya penulisan, termasuk aturan tanpa komentar |
| `backend/auth.md` | Alur access token dan refresh token, aturan cookie, daftar endpoint |
| `backend/slice-bisnis.md` | Pola service dan controller slice bisnis, endpoint dan aturan bisnis master, project, inventory |
| `backend/docker.md` | Beda Dockerfile dev dan prod, kenapa polling dipakai di Windows |
| `database/ekstensi.md` | PostGIS, pg_trgm, btree_gist |
| `database/index.md` | Kapan sebuah kolom diberi index, dan kapan tidak |
| `database/skema.md` | Daftar entity per slice dan alasan setiap tambahan |
| `frontend/arsitektur.md` | Susunan MVC, alur token, larangan penyimpanan |
| `frontend/aturan-ui.md` | Aturan UI gabungan Shneiderman dan Nielsen, dan bahasa visual |

## Keadaan sekarang

Sudah jalan:

- Daftar, masuk, keluar, dan sesi yang bertahan setelah halaman dimuat ulang
- Skema database lengkap untuk semua modul berikut constraintnya
- Penilaian lokasi dari data GIS nyata (layanan `scoring/`), peta dengan lapisan
  bahaya, kepadatan penduduk, dan harga tanah, serta perbandingan kandidat lokasi
- Layar untuk semua modul: Ringkasan, Proyek (ruang kerja per proyek), Analisis
  Lokasi, Keuangan (anggaran, kas, jurnal, buku besar), Tagihan, Penjualan,
  Pengadaan, Inventaris, dan SDM
- Laporan (arus kas, anggaran, proyek, umur piutang dan utang, stok,
  penggajian, penjualan) dengan pratinjau dan unduhan PDF, XLSX, dan CSV
- Progres berbobot RAB, kurva S, laporan harian lapangan, termin kontraktor
  dengan retensi yang menjadi hutang di Tagihan, dan adendum (pekerjaan
  tambah-kurang) yang disetujui Admin
- Hak akses per modul untuk 7 peran
- CI di GitHub Actions untuk backend, frontend, layanan scoring, dan integration
  test ke PostgreSQL + PostGIS

Belum dikerjakan:

- Hak akses per proyek (sekarang per modul)
- Penagihan termin ke pemberi kerja

Rincian dan sisa pekerjaan lain ada di `PROJECT_HANDOFF.md`.
