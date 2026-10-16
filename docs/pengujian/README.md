# Rencana Pengujian

Dokumen ini menjelaskan apa yang diuji, di level mana, dengan alat apa, dan
kapan sebuah perubahan dianggap siap digabung. Hasil terakhirnya ada di
[laporan hasil](laporan-hasil.md), skenario penerimaan pengguna di
[acceptance test](acceptance.md).

## Tujuan

1. Setiap aturan bisnis yang tertulis di kode punya test yang gagal kalau
   aturan itu rusak, misalnya transisi status, perhitungan progres berbobot,
   dan pencegahan pembayaran ganda.
2. Perilaku yang hanya muncul di database sungguhan (constraint, foreign key,
   query spasial, transaksi) diuji ke PostgreSQL + PostGIS, bukan ke tiruan.
3. Alur utama pengguna dibuktikan berjalan dari ujung ke ujung pada aplikasi
   yang hidup.
4. Tidak ada perubahan yang masuk `main` tanpa melewati CI.

## Level pengujian

| Level | Yang diuji | Alat | Jumlah | Kapan berjalan |
|---|---|---|---|---|
| Unit | Fungsi, service, dan repository satu per satu. Database dan layanan luar diganti tiruan | Go `testing` + `httptest` + driver SQL palsu (`shared/databasetest`); Vitest + Testing Library; pytest | 473 Go, 143 frontend, 15 Python | Setiap PR (CI) dan saat menulis kode |
| Integration | Service dan repository bersama PostgreSQL + PostGIS sungguhan: migrasi, constraint, CASCADE dan SET NULL, `ST_DWithin`, rotasi token | Go `testing` dengan build tag `integration`, database berakhiran `_test` | 6 | Setiap PR (CI, container `postgis/postgis:18-3.6`) |
| System | Alur utama pada aplikasi yang hidup lewat API sungguhan: autentikasi, proyek dan progres berbobot, laporan harian, evaluasi lokasi dan laporan PDF | `tests/system/run_system_tests.py` (Python, pustaka standar) | 5 skenario, 27 langkah | Sebelum demo dan sebelum rilis, di laptop dengan seluruh layanan menyala |
| Acceptance | Apakah sistem memenuhi kebutuhan pengguna per peran | Skenario Given/When/Then di [acceptance.md](acceptance.md), dijalankan manual di browser | 14 skenario | Sebelum demo; diulang oleh pengguna (anggota tim di luar pengembang) |

Pengujian tampilan pendukung yang dilakukan selama pengembangan: pemeriksaan
visual di browser pada 1440 px dan 390 px, dan pengukuran performa API.

## Lingkungan

| Lingkungan | Isi | Dipakai untuk |
|---|---|---|
| CI GitHub Actions | Ubuntu, Go dari `go.mod`, Node 22, Python 3.13, service PostGIS | Unit dan integration test di setiap PR |
| Laptop pengembang | Docker Desktop: PostGIS 18, backend dengan hot reload; Vite; layanan penilaian | System dan acceptance test, pemeriksaan browser |
| Database `bangunin_aja_test` | Database kosong terpisah, dimigrasi oleh test | Integration test lokal |

Integration test menolak berjalan kalau nama database tidak berakhiran
`_test`, supaya database kerja dan data demo tidak pernah tersentuh.

## Kriteria

**Siap digabung (exit criteria per PR)**

- Keempat job CI hijau: Backend (Go), Integration (PostGIS), Frontend (React), Scoring (Python)
- `go vet`, `oxlint`, dan `tsc` tanpa peringatan
- Perubahan perilaku punya test baru yang gagal sebelum perubahan
- Perubahan tampilan diperiksa di browser pada lebar desktop dan ponsel

**Siap demo atau rilis**

- System test lulus semua langkah pada aplikasi yang akan didemokan
- Skenario acceptance untuk fitur yang didemokan lulus

## Cara menjalankan

```bash
docker compose exec backend go test ./...
docker compose exec -e POSTGRES_DB=bangunin_aja_test backend go test -tags integration -v ./internal/integration/
cd frontend && npm test
cd scoring && python -m pytest
BANGUNINAJA_EMAIL=... BANGUNINAJA_PASSWORD=... python tests/system/run_system_tests.py
```

## Pelaporan cacat

Cacat yang ditemukan saat pengujian dicatat di deskripsi pull request yang
memperbaikinya, lengkap dengan langkah untuk memunculkannya, dampaknya, dan
bukti setelah diperbaiki. Daftar cacat yang ditemukan lewat pengujian ada di
[laporan hasil](laporan-hasil.md#cacat-yang-ditemukan-lewat-pengujian).
