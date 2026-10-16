# Model SDLC dan Proses Pengembangan

Dokumen ini menjelaskan cara BanguninAja benar-benar dibangun. Semua tanggal dan
angka diambil dari riwayat git dua repositori: repositori data `BanguninAja`
(skrip pengumpulan data dan catatan arsitektur) dan repositori aplikasi
`BanguninAjaApp`. Jadwalnya ada di [jadwal dan pembagian kerja](jadwal.md).

## 1. Model yang dipakai: iteratif dan inkremental

BanguninAja dibangun dalam beberapa **inkremen**. Setiap inkremen menambah fitur
di atas aplikasi yang sudah berjalan, dan sebelum inkremen pertama ada satu
putaran eksplorasi untuk risiko terbesar proyek, yaitu ketersediaan data.

| Ciri model | Bukti di repositori |
|---|---|
| Risiko terbesar dikerjakan lebih dulu | Fitur utama (analisis lokasi) bergantung pada data GIS nasional yang belum pasti ada. Minggu pertama (7 sampai 13 September) dipakai untuk mencari sumber data: 34 commit di repositori data. Data RDTR dengan KDB dan KLB baru ditemukan bisa diambil pada 12 September |
| Setiap inkremen menghasilkan perangkat lunak yang berjalan | Sejak 2 Oktober, perubahan ke `main` masuk lewat pull request (satu pengecualian: pembaruan dokumen handoff pada hari itu). Setiap PR menjalankan CI dan digabung setelah hijau, jadi `main` selalu bisa dijalankan. Pada fase fondasi masih ada 14 commit langsung ke `main` |
| Kebutuhan bertambah di tengah jalan | Analisis [kelengkapan domain](../analisis/kelengkapan-domain.md) (16 Oktober) menambah empat fitur P1, yaitu progres berbobot, laporan harian, termin, dan adendum. Keempatnya selesai di inkremen 3 |
| Umpan balik pemilik produk masuk ke inkremen berikutnya | Permintaan UI yang lebih lega dan isi domain yang lebih realistis menjadi #60 sampai #72 |

### Kenapa bukan model lain

| Model | Alasan tidak dipakai utuh |
|---|---|
| Waterfall | Kebutuhan tidak bisa dibekukan sebelum datanya pasti ada. Ketersediaan harga tanah, RDTR, dan data bencana baru terjawab pada 11 dan 12 September. Pekerjaan juga terpotong jadwal kuliah menjadi beberapa periode pendek, bukan satu alur panjang |
| Scrum murni | Tim mahasiswa tidak punya sprint berdurasi tetap dan rapat harian yang tercatat. Unsur Scrum yang dipakai: backlog tugas (T1 sampai T19 dan daftar G1 sampai G6 di `PROJECT_HANDOFF.md`), satu pemilik produk, inkremen yang bisa didemokan, dan tinjauan lewat pull request |
| Spiral | Putaran pertamanya dipakai (eksplorasi risiko data), tetapi analisis risiko formal per putaran baru ditulis di [risiko dan keamanan](../analisis/risiko-keamanan.md) pada 16 Oktober |

## 2. Tahapan SDLC dan artefaknya

Tahapan tidak berjalan berurutan sekali jalan. Analisis, desain, implementasi,
dan pengujian berulang di setiap inkremen. Tabel ini mencatat kapan tiap tahapan
pertama kali dikerjakan dan di mana hasilnya disimpan.

| Tahapan | Mulai | Kegiatan | Artefak |
|---|---|---|---|
| Perencanaan dan studi kelayakan | 7 September | Mencari dan menguji sumber data (BPS, InaRISK, ZNT, RDTR, Pusiknas), memilih arsitektur, membagi pekerjaan | Repositori data: tracker dataset, desain data, catatan arsitektur, `PROJECT_HANDOFF.md` dengan pembagian kerja dan target |
| Analisis kebutuhan | 18 September | Aturan arsitektur, skema database, aturan UI, peran dan hak akses, celah domain | [requirements](../requirements), [use case](../desain/use-case.md), [kelengkapan domain](../analisis/kelengkapan-domain.md) |
| Desain | 18 September (skema), 16 Oktober (UML lengkap) | ERD, class, sequence, state dan activity, deployment | [desain sistem](../desain/README.md) |
| Implementasi | 17 September | Empat inkremen, lihat bagian 3 | `backend/`, `frontend/`, `scoring/` |
| Pengujian | 23 September (unit test pertama) | Unit, integration, system, acceptance. CI sejak 2 Oktober | [pengujian](../pengujian/README.md) |
| Deployment | 2 Oktober | Docker Compose untuk pengembangan, demo online lewat Vercel dan Cloudflare tunnel | [deployment](../desain/deployment.md) |
| Pemeliharaan | 2 Oktober | Perbaikan lewat branch `fix/*`, misalnya #70: bug pencarian zona yang ditemukan integration test di CI | Pull request berawalan `fix` |

## 3. Inkremen

| Inkremen | Tanggal | Isi | PR |
|---|---|---|---|
| 0. Fondasi | 17 sampai 27 September | Auth dengan access dan refresh token, skema database, Docker, lapisan service dan controller semua slice, regulasi, berita, evaluasi lokasi, kerangka peta | #1 sampai #3 dan commit langsung |
| 1. MVP lengkap | 2 sampai 3 Oktober | Penilaian lokasi dari data GIS sungguhan, lapisan peta, ruang kerja proyek, semua layar ERP, modul yang saling terhubung (G1 sampai G6), CI | #4 sampai #43 |
| 2. Siap dipakai banyak peran | 15 Oktober | Hak akses 7 peran, laporan PDF, XLSX, dan CSV, data demo, tim proyek, data master, pembayaran on-chain, tinjauan desain | #44 sampai #53 |
| 3. Kualitas dan kedalaman domain | 16 sampai 17 Oktober | Dokumen UML, test keamanan dan integration, performa evaluasi lokasi, ruang lega, progres berbobot, laporan harian, termin dan retensi, adendum, perbaikan keamanan dan kerentanan dependensi | #54 sampai #74 |

## 4. Praktik yang dijalankan di setiap perubahan

| Praktik | Penerapan |
|---|---|
| Satu tugas, satu branch | Nama branch mengikuti jenis perubahan: `feat/`, `fix/`, `docs/`, `test/`, `perf/`, `style/`, `refactor/`, `ci/` |
| Pesan commit konvensional | `jenis(cakupan): ringkasan`, misalnya `fix(regulation): pilih zona terdekat`. Badan commit menjelaskan alasan, bukan mengulang isi diff |
| Pull request untuk setiap perubahan ke `main` | 72 PR sampai 17 Oktober. Deskripsi PR memuat apa yang berubah, cara mengecek, dan keterbatasan |
| Continuous integration | GitHub Actions sejak #23 (2 Oktober), sekarang lima job: backend Go, integration test ke PostgreSQL dan PostGIS (sejak #66), pemeriksaan kerentanan dependensi dengan `govulncheck` (sejak #74), frontend React dengan audit npm, layanan penilaian Python. PR digabung setelah semua job hijau. Aturan ini masih kesepakatan, belum dipaksa oleh branch protection |
| Merge commit | PR digabung dengan merge commit, jadi riwayat menunjukkan batas setiap PR |
| Definisi selesai | Test lulus di CI, dokumen requirements atau desain ikut diperbarui, perubahan UI diperiksa di browser termasuk lebar ponsel, data uji berawalan `UJI-` dihapus setelah pemeriksaan |

## 5. Kelemahan proses yang tercatat

Bagian ini sengaja dicatat supaya bisa diperbaiki, bukan disembunyikan.

| Kelemahan | Dampak | Perbaikan yang disarankan |
|---|---|---|
| Sebagian besar PR sejak 2 Oktober dibuat dan digabung oleh satu orang | Tinjauan kode oleh orang kedua hampir tidak ada | Setiap anggota memegang tugas lewat branch dan PR sendiri, dan setiap PR ditinjau minimal satu anggota lain |
| Branch `main` belum dilindungi | Push langsung ke `main` masih mungkin | Pemilik repositori (`jojondrw`) mengaktifkan branch protection: wajib PR, wajib CI hijau, wajib satu persetujuan |
| Target 27 September terlewat | MVP terintegrasi baru tercapai 3 Oktober, enam hari dari target | Lihat analisis rencana dan realisasi di [jadwal](jadwal.md) |
| Dokumen UML ditulis setelah sebagian besar kode jadi | Desain lebih banyak mendokumentasikan daripada memandu | Fitur P1 (inkremen 3) sudah memakai urutan analisis, desain, implementasi |
