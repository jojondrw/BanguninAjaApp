# BanguninAja

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

Dijalankan sekali di awal, dan diulang setiap ada entity baru. Aman dijalankan
berkali-kali.

### 4. Nyalakan tampilannya

```bash
cd frontend
npm install
npm run dev
```

Buka `http://localhost:5173`, buat akun, lalu masuk.

## Memeriksa isi database

```bash
docker compose exec database psql -U $POSTGRES_USER -d $POSTGRES_DB
```

## Dokumentasi

Alasan di balik tiap keputusan ada di `docs/requirements`.

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
| `frontend/aturan-ui.md` | Aturan UI gabungan Shneiderman dan Nielsen |

## Keadaan sekarang

Sudah jalan:

- Daftar, masuk, keluar, dan sesi yang bertahan setelah halaman dimuat ulang
- Skema database lengkap untuk semua modul, 47 tabel berikut constraintnya
- API untuk data master (wilayah, satuan, akun), proyek (tahapan, RAB, izin), dan
  persediaan (material, gudang, stok, mutasi stok)
- API untuk pengadaan, aset, penjualan, keuangan, tagihan, SDM, penilaian,
  lokasi tersimpan, dan daftar laporan. Rinciannya ada di
  `backend/slice-bisnis.md`

Belum dikerjakan:

- Pembatasan akses berdasarkan peran
- Penerimaan barang yang langsung menambah stok gudang
- Pembuatan berkas laporan PDF, XLSX, dan CSV
- Layar modul proyek, pengadaan, penjualan, keuangan, SDM, dan peta lokasi
