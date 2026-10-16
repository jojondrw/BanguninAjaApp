# Tren Rekayasa Perangkat Lunak dan Etika Profesi

Topik AoL mata kuliah ini adalah *Emerging Trends in Software Engineering*.
Dokumen ini memetakan tren yang diterapkan BanguninAja beserta buktinya di
repositori, tren yang sengaja tidak diikuti, dan etika profesi yang dipegang
selama pengerjaan. Rujukan ada di bagian akhir.

## 1. Tren yang diterapkan

| Tren | Intinya | Penerapan di BanguninAja | Bukti |
|---|---|---|---|
| DevOps dan CI/CD | Perubahan kecil yang sering, diperiksa otomatis oleh pipeline. Kinerja pengiriman diukur dengan metrik DORA [2], misalnya lead time perubahan dan frekuensi deployment | Lima job CI berjalan di setiap pull request. Ada 72 PR dalam empat minggu, dan `main` selalu bisa dijalankan | #23, #66, #74, [model SDLC](../proses/sdlc.md) |
| DevSecOps dan keamanan rantai pasok perangkat lunak | Keamanan diperiksa sejak PR, bukan menjelang rilis. OWASP Top 10:2025 kini memuat kategori baru A03 *Software Supply Chain Failures* [1] | `govulncheck` [9] dan `npm audit` menjadi gerbang CI. Pemeriksaan pertama menemukan dan menutup 10 kerentanan dependensi. Header keamanan API dan batas percobaan login ditambahkan | #68, #74, [risiko dan keamanan](risiko-keamanan.md) |
| Container dan konfigurasi lewat environment | Lingkungan yang sama untuk semua pengembang. Konfigurasi disimpan di environment, bukan di kode (faktor III Twelve-Factor App [3]) | Docker Compose untuk database dan backend; rahasia dan alamat layanan dibaca dari `.env` yang tidak masuk Git; Dockerfile produksi terpisah | `docker-compose.yml`, `backend/Dockerfile.prod` |
| Monolit modular dengan *vertical slice* | Mulai dari monolit dan memecahnya hanya kalau perlu [4]; kode disusun per fitur, bukan per lapisan teknis [5] | Backend Go terdiri dari 16 slice, masing-masing punya controller, service, dan repository sendiri. Satu-satunya layanan terpisah adalah penilaian lokasi (Python), karena kebutuhan pustaka GIS dan datanya memang berbeda | [arsitektur backend](../requirements/backend/arsitektur.md) |
| *Docs as code* dan *diagrams as code* | Dokumen dan diagram ditulis sebagai teks, disimpan bersama kode, dan ditinjau di PR yang sama | 37 diagram UML dan ERD ditulis dalam Mermaid [10] dan diperiksa parse serta render-nya sebelum digabung | #55, [desain sistem](../desain/README.md) |
| Piramida test dengan dependensi sungguhan | Banyak test unit yang cepat, lebih sedikit test integration, dan sedikit test ujung ke ujung [6] | 500 test unit Go, 156 test frontend, 9 integration test ke PostgreSQL dan PostGIS sungguhan di CI, 27 langkah system test, 17 skenario acceptance. Integration test menemukan bug nyata yang tidak tertangkap test unit | #66, #67, #70, [laporan pengujian](../pengujian/laporan-hasil.md) |
| Branch berumur pendek dan commit konvensional | Integrasi ke cabang utama setidaknya setiap hari [8]; pesan commit yang terbaca mesin dan manusia [7] | Satu tugas satu branch, digabung lewat PR dalam hitungan jam. Commit memakai format `jenis(cakupan): ringkasan` | [model SDLC](../proses/sdlc.md) |
| *Location intelligence* | Data geospasial dan model penilaian dipakai langsung dalam keputusan bisnis | Skor kelayakan lokasi lima dimensi dari data GIS nasional, dipakai sebelum lahan dibeli dan dilanjutkan ke ruang kerja proyek | #7, #18, #28 |
| Pembayaran berbasis blockchain (eksperimen) | Pembayaran yang bisa diverifikasi publik tanpa perantara | Simulasi di testnet Sepolia. Server hanya menerima dan memverifikasi transaksi, tanpa menyimpan kunci privat | #52 |
| Pengembangan perangkat lunak berbantuan AI | Asisten AI mempercepat penulisan kode; dalam satu eksperimen terkendali, tugas selesai 55,8% lebih cepat [11]. Tanggung jawab atas hasilnya tetap pada pengembang | Aturan proyek berlaku sama untuk semua kode, siapa pun atau apa pun yang menulisnya: lewat PR, lulus test dan CI, dan dokumen ikut diperbarui | Bagian 3 dokumen ini |

## 2. Tren yang sengaja tidak diikuti

Mengikuti tren bukan tujuan. Pilihan berikut diambil karena biayanya lebih besar
daripada manfaatnya untuk tim dan skala proyek ini.

| Tren | Alasan tidak dipakai |
|---|---|
| Microservices dan Kubernetes | Tim kecil dengan satu database relasional. Memecah layanan menambah biaya jaringan, deployment, dan konsistensi data tanpa masalah nyata yang diselesaikan [4] |
| Backend serverless | Query spasial PostGIS butuh koneksi database yang bertahan lama, dan layanan penilaian memuat bundle data sekitar 239 MB |
| GraphQL | Layar mengikuti sumber daya REST yang jelas per slice; REST lebih sederhana dicache dan diamankan per modul |
| Aplikasi mobile native | Aplikasi web responsif yang diuji di lebar 375 px sudah memenuhi kebutuhan staf kantor dan lapangan |

## 3. Etika profesi

Pegangan utama: *Software Engineering Code of Ethics and Professional Practice*
dari IEEE-CS dan ACM [13], yang berisi delapan prinsip, serta *ACM Code of Ethics
and Professional Conduct* [12]. Pertimbangan hukum dan lisensi data ada di
[risiko dan keamanan](risiko-keamanan.md) bagian 3.

| Prinsip | Penerapan |
|---|---|
| Masyarakat (*Public*) | Skor dari data cadangan ditandai "simulasi" mulai dari respons API sampai laporan PDF (#64), supaya tidak dibaca sebagai hasil analisis sungguhan saat orang memutuskan membeli lahan. Berita hanya menjadi konteks, tidak masuk ke skor |
| Klien dan pemberi kerja (*Client and Employer*) | Uang tidak bisa keluar tanpa dasar. Termin hanya bisa diajukan kalau progres berbobot sudah tercapai; yang mengajukan (Manajer Proyek) berbeda dari yang membayar (Keuangan); adendum disetujui Admin. Siapa dan kapan dicatat di setiap langkah |
| Produk (*Product*) | Empat level pengujian, pemeriksaan kerentanan di CI, dan keterbatasan yang ditulis terbuka di setiap PR |
| Penilaian (*Judgment*) | Analisis keamanan menandai kategori yang belum aman sebagai "Sebagian", bukan dibulatkan menjadi aman. Contohnya: hak akses per proyek dan enkripsi NIK |
| Manajemen (*Management*) | Rencana dan realisasi jadwal dibandingkan apa adanya, termasuk keterlambatan enam hari dan kelemahan proses ([jadwal](../proses/jadwal.md)) |
| Profesi (*Profession*) | Konvensi kode, dokumentasi, dan atribusi sumber data (BPS, BNPB InaRISK, ATR/BPN, OpenStreetMap) mengikuti lisensinya |
| Rekan kerja (*Colleagues*) | Setiap PR menjelaskan apa yang berubah dan cara mengeceknya. README memuat langkah menjalankan migrasi dan data demo supaya anggota lain bisa mengikuti |
| Diri sendiri (*Self*) | Praktik diperbarui mengikuti standar terbaru, misalnya pemetaan keamanan ke OWASP Top 10:2025 |

### Data pribadi

Aplikasi menyimpan data pribadi pelanggan dan karyawan (nama, NIK, kontak, gaji),
sehingga tunduk pada UU Pelindungan Data Pribadi [14]. Kontrol yang ada dan celah
yang tersisa, misalnya NIK yang belum dienkripsi, dicatat di
[risiko dan keamanan](risiko-keamanan.md).

### Pemakaian alat AI

Kalau tim memakai asisten AI, berlaku tiga aturan:

1. **Tanggung jawab tetap pada pengembang.** Kode buatan AI melewati PR, test,
   dan CI yang sama dengan kode lain.
2. **Jangan mengirim rahasia atau data pribadi ke layanan AI.** Ini mencakup isi
   `.env`, kata sandi, dan data pelanggan.
3. **Bersikap terbuka kepada dosen dan tim** tentang alat yang dipakai, sesuai
   aturan kejujuran akademik yang berlaku.

## Rujukan

1. OWASP Foundation. *OWASP Top 10:2025*. https://top10.owasp.org/2025/
2. DORA. *DORA's software delivery performance metrics*. https://dora.dev/guides/dora-metrics-four-keys/
3. Wiggins, A. *The Twelve-Factor App*. https://12factor.net/
4. Fowler, M. (2015). *Monolith First*. https://martinfowler.com/bliki/MonolithFirst.html
5. Bogard, J. (2018). *Vertical Slice Architecture*. https://www.jimmybogard.com/vertical-slice-architecture/
6. Vocke, H. (2018). *The Practical Test Pyramid*. https://martinfowler.com/articles/practical-test-pyramid.html
7. *Conventional Commits 1.0.0*. https://www.conventionalcommits.org/en/v1.0.0/
8. *Trunk Based Development*. https://trunkbaseddevelopment.com/
9. The Go Project. *Go Vulnerability Management*. https://go.dev/security/vuln/
10. *Mermaid*. https://mermaid.js.org/
11. Peng, S., Kalliamvakou, E., Cihon, P., & Demirer, M. (2023). *The Impact of AI on Developer Productivity: Evidence from GitHub Copilot*. arXiv:2302.06590.
12. ACM. (2018). *ACM Code of Ethics and Professional Conduct*. https://www.acm.org/code-of-ethics
13. IEEE-CS/ACM Joint Task Force on Software Engineering Ethics and Professional Practices. (1999). *Software Engineering Code of Ethics and Professional Practice*. https://www.acm.org/code-of-ethics/software-engineering-code
14. Republik Indonesia. (2022). *Undang-Undang Nomor 27 Tahun 2022 tentang Pelindungan Data Pribadi*.
