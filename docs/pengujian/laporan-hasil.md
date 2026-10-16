# Laporan Hasil Pengujian

Pengujian dijalankan 16 sampai 17 Oktober 2026 pada `main` setelah PR #64 dan
#65, ditambah integration test dari PR #66. Rencana dan cara menjalankannya ada
di [rencana pengujian](README.md).

## Ringkasan

| Level | Jumlah | Lulus | Gagal | Catatan |
|---|---|---|---|---|
| Unit, backend Go | 473 test (+60 subtest) | 473 | 0 | Coverage pernyataan 47,1% |
| Unit, frontend | 143 test di 23 berkas | 143 | 0 | Coverage belum diukur (lihat keterbatasan) |
| Unit, layanan penilaian Python | 15 | 15 | 0 | |
| Integration (PostgreSQL + PostGIS) | 6 | 6 | 0 | 5,3 detik |
| System | 5 skenario, 27 langkah | 27 | 0 | 3,2 detik |
| Acceptance | 14 skenario | 12 | 0 | 8 lewat UI, 4 lewat API; 2 belum dijalankan (pesanan pembelian, pembayaran on-chain). Persetujuan pengguna di [acceptance.md](acceptance.md) |

### Pembaruan setelah termin, retensi, dan adendum

Diukur ulang pada branch PR #71 (termin dan retensi), yang sudah memuat PR #67 sampai
#70. Angka di bagian lain laporan ini tetap angka pengukuran pertama.

| Level | Jumlah | Lulus | Gagal | Catatan |
|---|---|---|---|---|
| Unit, backend Go | 500 test (+62 subtest) | 500 | 0 | Coverage pernyataan 47,8%. Paket `project` 72 test, coverage 62,1% |
| Unit, frontend | 156 test di 24 berkas | 156 | 0 | Ditambah 13 test kartu dan model termin serta adendum |
| Integration (PostgreSQL + PostGIS) | 9 | 9 | 0 | Ditambah pengajuan termin menjadi hutang sungguhan, retensi yang hanya bisa dilepas sekali, dan adendum yang mengubah nilai pengajuan berikutnya |
| Acceptance | 17 skenario | 15 | 0 | Ditambah AT-15 dan AT-17 (lewat UI) serta AT-16 (lewat test render dan aturan akses); AT-10 dan AT-13 masih belum dijalankan |

Coverage `billing` turun tipis dari 73,4% ke 72,8% karena `payable_ledger.go`
diuji lewat integration test, bukan unit test paket itu sendiri.

## Unit test backend per paket

Diurutkan dari jumlah test terbanyak. Coverage adalah persentase pernyataan Go
yang dijalankan oleh test paket itu sendiri.

| Paket | Test | Coverage |
|---|---|---|
| `billing` | 84 | 73,4% |
| `project` | 50 | 57,8% |
| `auth` | 35 | 61,5% |
| `reporting` | 34 | 87,0% |
| `finance` | 27 | 39,5% |
| `procurement` | 24 | 27,7% |
| `site` | 24 | 66,9% |
| `location` | 22 | 61,7% |
| `sales` | 19 | 15,9% |
| `shared/ethereum` | 19 | 88,8% |
| `master` | 16 | 34,5% |
| `hr` | 15 | 34,1% |
| `shared/access` | 15 | 92,6% |
| `cmd/seed-demo` | 14 | 30,3% |
| `inventory` | 12 | 20,1% |
| `shared/middleware` | 11 | 74,5% |
| `reporting/export` | 9 | 90,0% |
| `news` | 5 | 28,2% |
| `shared/security` | 5 | 100% |
| `shared/token` | 4 | 94,9% |
| `shared/cookie` | 4 | 93,8% |
| `cmd/api` | 4 | 46,9% |
| `asset` | 4 | 7,3% |
| `scoring` | 3 | 11,0% |
| `shared/clock` | 3 | 100% |
| `shared/duedate` | 3 | 100% |
| `shared/config` | 3 | 44,3% |
| `shared/pagination` | 3 | 69,2% |
| `shared/database` | 2 | 22,4% |

Paket tanpa unit test: `regulation` dan `migration` (diuji lewat integration
test), program `cmd/migrate`, `cmd/seed`, `cmd/import-rdtr`, serta paket kecil
pembungkus (`apperror`, `entity`, `geo`, `httprequest`, `httpresponse`,
`databasetest`).

Paket keamanan `token`, `cookie`, dan `security` naik dari 0% ke lebih dari 93%
selama siklus ini.

## Integration test

| Test | Hasil |
|---|---|
| Migrasi dapat dijalankan ulang dan constraint baru ada | Lulus |
| Progres berbobot RAB tersimpan, riwayat harian satu baris, hapus tahap mengosongkan `phase_id` (SET NULL) | Lulus |
| Laporan harian kedua di tanggal yang sama ditolak constraint unik (409) | Lulus |
| Hapus proyek ikut menghapus laporan harian dan riwayat progres (CASCADE) | Lulus |
| `ST_DWithin` memilih zona regulasi terdekat dalam 150 m, di luar radius kembali ke data simulasi | Lulus |
| Refresh token yang dipakai ulang mencabut semua sesi pengguna | Lulus |

## System test

Keluaran `tests/system/run_system_tests.py` pada aplikasi lokal (backend,
PostGIS, dan layanan penilaian menyala), 17 Oktober 2026 pukul 00.56 WIB:

```text
ST-01 Autentikasi
  LULUS  Masuk dengan kredensial benar                    HTTP 200
  LULUS  Profil pengguna terbaca                          HTTP 200
  LULUS  Kata sandi salah ditolak                         HTTP 401 invalid_credential
  LULUS  Tanpa token ditolak                              HTTP 401
  LULUS  Token palsu ditolak                              HTTP 401

ST-02 Proyek dan progres berbobot
  LULUS  Membuat proyek                                   HTTP 201
  LULUS  Menambah tahap Persiapan                         HTTP 201
  LULUS  Menambah tahap Struktur                          HTTP 201
  LULUS  Membaca satuan dari data master                  14 satuan
  LULUS  Item RAB 1.1 dikaitkan ke tahap                  HTTP 201
  LULUS  Item RAB 2.1 dikaitkan ke tahap                  HTTP 201
  LULUS  Progres berbobot RAB: 10% x 100 + 90% x 20 = 28  progres 28
  LULUS  Kurva S berbobot RAB dengan titik rencana        weighting budget, 54 titik
  LULUS  Deviasi hari ini dihitung                        deviasi 9.6

ST-03 Laporan harian
  LULUS  Menulis laporan hari ini                         HTTP 201
  LULUS  Laporan kedua di tanggal yang sama ditolak       HTTP 409 daily_report_date_used
  LULUS  Laporan untuk besok ditolak                      HTTP 422 daily_report_in_future
  LULUS  Ringkasan 30 hari menghitung hari hujan          1 laporan, 1 hujan

ST-04 Evaluasi lokasi
  LULUS  Profil bangunan tersedia                         4 profil
  LULUS  Mengevaluasi titik di Bandung                    HTTP 201, 2107 ms
  LULUS  Skor 0 sampai 100 dengan 5 dimensi               skor 77
  LULUS  Skor berasal dari model, bukan cadangan          is_simulated False
  LULUS  Evaluasi selesai di bawah 3 detik                2107 ms
  LULUS  Lokasi tersimpan dan terkait ke proyek           HTTP 200
  LULUS  Laporan PDF lokasi bisa diunduh                  HTTP 200, 9261 byte

ST-05 Pembersihan
  LULUS  DELETE locations                                 HTTP 204
  LULUS  DELETE projects                                  HTTP 204

27/27 langkah lulus dalam 3.2 detik
```

Percobaan pertama menghasilkan 20/21 karena skrip mengharapkan 200 untuk
evaluasi lokasi, padahal API menjawab 201 Created. Yang salah adalah ekspektasi
skrip (dan sequence diagram di `docs/desain`), bukan aplikasinya; keduanya
diperbaiki.

## Performa

Diukur di laptop pengembang dengan data demo.

| Yang diukur | Hasil |
|---|---|
| 22 endpoint baca halaman utama | Median 8 sampai 25 ms |
| Perhitungan skor di layanan penilaian | Median 125 ms |
| Evaluasi lokasi ujung ke ujung, sebelum perbaikan | Median 3,4 s, bisa sampai 10 s |
| Evaluasi lokasi, sesudah perbaikan (PR #58) | Paling lama sekitar 2 s; 0,23 s untuk wilayah yang sudah dievaluasi |
| Bundle awal frontend | 111 kB gzip; 15 halaman dimuat terpisah, peta 1 MB hanya di halaman peta |

## Cacat yang ditemukan lewat pengujian

| Cacat | Ditemukan oleh | Dampak | Perbaikan |
|---|---|---|---|
| Dua test PayPage gagal saat mesin sibuk | Suite frontend di laptop yang sedang berat | CI bisa merah tanpa ada bug | #56: hash ditempel, bukan diketik |
| Kata sandi 72 karakter beraksen (lebih dari 72 byte) menyebabkan galat 500 | Unit test baru paket `security` | Pendaftaran dan ganti kata sandi gagal dengan galat server | #62: 422 `password_too_long` |
| Evaluasi lokasi menunggu Google News sampai 10 s | Pengukuran performa | Pengguna menunggu lama untuk kartu yang tidak ikut dihitung | #58: berita paralel, batas 2 s, cache |
| Tanggal "hari ini" mundur sehari antara 00.00 dan 07.00 WIB | Pemeriksaan data setelah migrasi PR #61 | Status jatuh tempo dan riwayat progres salah tanggal | #63: zona waktu `Asia/Jakarta` |
| Skor cadangan tersimpan seperti skor model | Penelusuran kode untuk sequence diagram | Keputusan investasi bisa memakai angka semu | #64: tanda simulasi sampai laporan PDF |
| Progres proyek rata-rata biasa, tidak berbobot | Penilaian domain | Progres terlihat lebih tinggi dari nilai pekerjaan sebenarnya | #61: progres berbobot RAB |

## Keterbatasan

| Keterbatasan | Rencana |
|---|---|
| Coverage frontend belum diukur karena `@vitest/coverage-v8` belum dipasang | Tambah dependensi dan laporkan di CI |
| Belum ada test otomatis lewat browser (end-to-end) | System test API ditambah beberapa skenario Playwright untuk alur paling penting |
| Coverage beberapa slice ERP masih rendah (`sales` 15,9%, `asset` 7,3%, `inventory` 20,1%) | Prioritaskan aturan bisnis yang belum dites: transisi kontrak, stok keluar |
| Pengukuran performa memakai data demo yang kecil | Uji beban dengan data sintetis ribuan transaksi di database terpisah |
| Acceptance test baru dijalankan tim pengembang | Diulang oleh anggota tim di luar pengembang dan dosen pembimbing |
