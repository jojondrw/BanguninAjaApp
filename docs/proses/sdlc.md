# Model SDLC dan Proses Pengembangan

Dokumen ini menjelaskan cara BanguninAja benar-benar dibangun. Semua tanggal dan
angka diambil dari riwayat git dua repositori: repositori data `BanguninAja`
(skrip pengumpulan data dan catatan arsitektur) dan repositori aplikasi
`BanguninAjaApp`. Jadwalnya ada di [jadwal dan pembagian kerja](jadwal.md).

## 1. Model yang dipakai: Scrum dengan prototipe

Tim memakai **Scrum**, kerangka agile yang iteratif dan inkremental: pekerjaan
dipecah menjadi sprint pendek, setiap sprint ditutup dengan inkremen aplikasi
yang bisa dijalankan, dan prioritas berikutnya disesuaikan dari umpan balik
pemilik produk. Tampilan disepakati lebih dulu lewat **prototipe**. Sebelum
sprint pertama ada sprint 0 untuk risiko terbesar proyek, yaitu ketersediaan
data. Model ini sama dengan yang disampaikan di presentasi progres.

| Unsur Scrum | Penerapan di BanguninAja | Bukti |
|---|---|---|
| Pemilik produk | Derick menetapkan prioritas dan menerima hasil setiap sprint | Keputusan fitur P1 dan permintaan UI yang lebih lega menjadi #60 sampai #87 |
| Product backlog | Tugas T1 sampai T19 dan G1 sampai G6 di `PROJECT_HANDOFF.md`; sejak 17 Oktober issue #78 sampai #85 di milestone [AoL akhir](https://github.com/jojondrw/BanguninAjaApp/milestone/1) | `PROJECT_HANDOFF.md`, GitHub Issues |
| Sprint dan inkremen | Sprint satu sampai dua minggu (bagian 3). Setiap perubahan masuk ke `main` lewat pull request dan CI hijau, jadi `main` selalu bisa dijalankan | Sejak 2 Oktober semua perubahan lewat PR, kecuali satu pembaruan dokumen handoff hari itu; fase fondasi masih punya 14 commit langsung |
| Sprint review | Presentasi progres dan demo aplikasi kepada dosen | Tujuh versi deck presentasi progres |
| Definition of done | Test lulus di CI, dokumen ikut diperbarui, perubahan UI dicek di browser | Bagian 4 |
| Prototipe | Tampilan disepakati sebelum dibangun, lalu diperhalus lewat tinjauan desain | Deck progres, tinjauan desain #53 dan #54 |
| Kebutuhan yang berubah | Analisis [kelengkapan domain](../analisis/kelengkapan-domain.md) (16 Oktober) menambah empat fitur P1 ke backlog, dan keempatnya selesai di sprint 3 | #61, #65, #71, #72 |

### Kenapa Scrum, bukan model lain

| Model | Alasan tidak dipilih |
|---|---|
| Waterfall | Kebutuhan tidak bisa dibekukan sebelum datanya pasti ada. Ketersediaan harga tanah, RDTR, dan data bencana baru terjawab pada 11 dan 12 September |
| V-Model | Setiap tahap pengujian terikat pada dokumen tahap yang sudah beku; tidak cocok saat kebutuhan masih berubah |
| Spiral | Analisis risiko formal di setiap putaran terlalu berat untuk tim kecil. Gagasan "risiko terbesar dulu" tetap dipakai sebagai sprint 0 |

## 2. Tahapan SDLC dan artefaknya

Tahapan tidak berjalan berurutan sekali jalan. Analisis, desain, implementasi,
dan pengujian berulang di setiap sprint. Tabel ini mencatat kapan tiap tahapan
pertama kali dikerjakan dan di mana hasilnya disimpan.

| Tahapan | Mulai | Kegiatan | Artefak |
|---|---|---|---|
| Perencanaan dan studi kelayakan | 7 September | Mencari dan menguji sumber data (BPS, InaRISK, ZNT, RDTR, Pusiknas), memilih arsitektur, membagi pekerjaan | Repositori data: tracker dataset, desain data, catatan arsitektur, `PROJECT_HANDOFF.md` dengan pembagian kerja dan target |
| Analisis kebutuhan | 18 September | Aturan arsitektur, skema database, aturan UI, peran dan hak akses, celah domain | [requirements](../requirements), [use case](../desain/use-case.md), [kelengkapan domain](../analisis/kelengkapan-domain.md) |
| Desain | 18 September (skema), 16 Oktober (UML lengkap) | ERD, class, sequence, state dan activity, deployment | [desain sistem](../desain/README.md) |
| Implementasi | 17 September | Tiga sprint pembangunan setelah sprint 0, lihat bagian 3 | `backend/`, `frontend/`, `scoring/` |
| Pengujian | 23 September (unit test pertama) | Unit, integration, system, acceptance. CI sejak 2 Oktober | [pengujian](../pengujian/README.md) |
| Deployment | 2 Oktober | Docker Compose untuk pengembangan, demo online lewat Vercel dan Cloudflare tunnel | [deployment](../desain/deployment.md) |
| Pemeliharaan | 2 Oktober | Perbaikan lewat branch `fix/*`, misalnya #70: bug pencarian zona yang ditemukan integration test di CI | Pull request berawalan `fix` |

### Alat yang mendukung tiap tahapan

| Tahapan | Alat | Cara alat itu membantu |
|---|---|---|
| Perencanaan | Git dan GitHub, Markdown, Google Drive tim "AOL SWE" | Daftar tugas T1 sampai T19, pembagian kerja, dan target tanggal ditulis di `PROJECT_HANDOFF.md` sehingga ikut terversi. Dataset besar yang tidak masuk Git dibagikan lewat Drive |
| Analisis kebutuhan | Markdown di folder `docs/requirements`, Mermaid | Setiap aturan bisnis dan alasannya tersimpan di samping kode dan ikut ditinjau di PR yang sama |
| Desain | Mermaid (diagram sebagai kode), PostgreSQL `pg_constraint` | 37 diagram UML dan ERD ditulis sebagai teks, jadi bisa di-diff dan diperbarui bersama kodenya. ERD diturunkan dari foreign key yang benar-benar ada di database |
| Implementasi | Go, Gin, GORM; React, Vite, TypeScript, Tailwind; Python FastAPI; Docker Compose dengan air | Docker Compose menyamakan lingkungan semua anggota; air membangun ulang backend setiap berkas disimpan |
| Pengujian | `go test`, Vitest dan Testing Library, pytest, test integration ke PostGIS sungguhan, skrip system test, browser untuk acceptance | Empat level pengujian berjalan dengan alat yang sesuai bahasanya. Hasilnya dicatat di [laporan pengujian](../pengujian/laporan-hasil.md) |
| Keamanan | `govulncheck`, `npm audit`, OWASP Top 10 sebagai daftar periksa | Kerentanan dependensi diperiksa otomatis di setiap PR sejak #74 |
| Deployment | Dockerfile produksi, Vercel, Cloudflare Tunnel | Frontend di Vercel, API demo dari laptop lewat tunnel tanpa membuka port |
| Manajemen proyek dan kolaborasi | GitHub Issues dengan milestone dan label prioritas, pull request, GitHub Actions, Gantt Mermaid | Sampai 17 Oktober tugas dilacak di `PROJECT_HANDOFF.md` dan terlihat sebagai PR beserta status CI-nya. Sejak itu sisa pekerjaan menjadi issue #78 sampai #85 di milestone [AoL akhir](https://github.com/jojondrw/BanguninAjaApp/milestone/1), dan PR yang menyelesaikannya menutup issue-nya secara otomatis |

## 3. Sprint

| Sprint | Tanggal | Hasil | PR |
|---|---|---|---|
| 0. Eksplorasi data | 7 sampai 13 September | Sumber data nasional diuji dan diambil, arsitektur dan desain data ditulis | Repositori data |
| 1. Fondasi | 14 sampai 27 September | Auth dengan access dan refresh token, skema database, Docker, lapisan service dan controller semua slice, regulasi, berita, evaluasi lokasi, kerangka peta | #1 sampai #3 dan commit langsung |
| 2. MVP lengkap | 28 September sampai 4 Oktober | Penilaian lokasi dari data GIS sungguhan, lapisan peta, ruang kerja proyek, semua layar ERP, modul yang saling terhubung (G1 sampai G6), CI | #4 sampai #43 |
| 3. Banyak peran, mutu, dan kedalaman domain | 12 sampai 18 Oktober | Hak akses 7 peran dan per proyek, laporan, data demo, pembayaran on-chain, dokumen UML, pengujian empat level, keamanan, progres berbobot, laporan harian, termin dan retensi, adendum, jejak audit | #44 sampai #87 |

Tidak ada commit pada 5 sampai 11 Oktober, sehingga minggu itu tidak dihitung
sebagai sprint.

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
| Daily scrum dan retrospektif tidak tercatat | Kendala antar-anggota sulit ditelusuri, dan perbaikan proses tidak terdokumentasi | Catat ringkasan singkat di issue atau PR setiap akhir sprint |
| Target 27 September terlewat | MVP terintegrasi baru tercapai 3 Oktober, enam hari dari target | Lihat analisis rencana dan realisasi di [jadwal](jadwal.md) |
| Dokumen UML ditulis setelah sebagian besar kode jadi | Desain lebih banyak mendokumentasikan daripada memandu | Fitur P1 (sprint 3) sudah memakai urutan analisis, desain, implementasi |
