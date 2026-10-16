# Desain Sistem BanguninAja

Model arsitektur dan UML untuk BanguninAja. Semua diagram ditulis dengan
[Mermaid](https://mermaid.js.org/) sehingga GitHub langsung menampilkannya
sebagai gambar, dan perubahan diagram tercatat di Git seperti kode.

Diagram diturunkan dari kode yang berjalan, bukan dari rancangan awal: relasi
ERD diambil dari `pg_constraint` database, peran dari matriks akses backend,
dan setiap langkah sequence diagram punya fungsi yang bisa ditelusuri. Kalau
kode berubah, diagram yang terdampak ikut diperbarui di pull request yang sama.

| Dokumen | Isi | Jumlah diagram |
|---|---|---|
| [Use case](use-case.md) | 12 aktor, matriks hak akses 7 peran, use case per modul | 6 |
| [ERD](erd.md) | 50 tabel dan 79 foreign key, dipecah per domain | 7 |
| [Class diagram](class.md) | Pola vertical slice backend, infrastruktur bersama, lapisan frontend, layanan penilaian Python | 4 |
| [Sequence diagram](sequence.md) | Login dan rotasi token, otorisasi, evaluasi lokasi, pembayaran on-chain, penerimaan barang | 5 |
| [State dan activity diagram](state-activity.md) | Siklus status proyek, pengadaan, kontrak, tagihan, on-chain; alur pemilihan lokasi, pengadaan, penjualan | 9 |
| [Deployment dan komponen](deployment.md) | Demo online, pengembangan dan CI, komponen modular monolith, keterbatasan | 3 |

Dokumen kebutuhan dan aturan kode ada di [`docs/requirements`](../requirements).

## Ringkasan arsitektur

| Bagian | Teknologi | Peran |
|---|---|---|
| Frontend | React 19, Vite, TypeScript, Tailwind v4, TanStack Query, MapLibre GL | 17 halaman ERP dan analisis lokasi |
| Backend | Go, Gin, GORM, 16 slice dalam satu proses | 224 endpoint REST, hak akses per peran |
| Database | PostgreSQL 18 + PostGIS 3.6 | Data ERP dan titik lokasi |
| Layanan penilaian | Python, FastAPI, rasterio, geopandas, scikit-learn | Skor kelayakan 5 dimensi dari data GIS nasional |
| Pembayaran | Ethereum Sepolia (jaringan uji) | Pembayaran tagihan on-chain yang diverifikasi backend |

## Cara memperbarui diagram

1. Ubah blok `mermaid` di berkas Markdown yang bersangkutan.
2. Pratinjau di <https://mermaid.live> atau di tab pratinjau pull request GitHub.
3. Untuk label dua baris, pakai `<br>`, bukan `\n`.
