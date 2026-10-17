# Jadwal, WBS, dan Pembagian Kerja

Rencana diambil dari `PROJECT_HANDOFF.md` di repositori data (ditulis 21
September). Realisasi diambil dari tanggal commit dan pull request di kedua
repositori. Model prosesnya dijelaskan di [model SDLC](sdlc.md).

## 1. Gantt rencana dan realisasi

```mermaid
gantt
    title Rencana dan realisasi BanguninAja
    dateFormat YYYY-MM-DD
    axisFormat %d %b
    tickInterval 1week

    section Rencana 21 Sep
    Kickoff dan sinkron tim          :milestone, r0, 2026-09-21, 0d
    Pembekuan kontrak API            :milestone, r1, 2026-09-22, 0d
    Pengerjaan vertikal per orang    :r2, 2026-09-21, 2026-09-26
    Integrasi                        :r3, 2026-09-26, 1d
    Semua tugas selesai              :milestone, r4, 2026-09-27, 0d

    section Realisasi
    Sprint 0 eksplorasi data         :done, a0, 2026-09-07, 2026-09-14
    Sprint 1 fondasi                 :done, a1, 2026-09-14, 2026-09-28
    Sprint 2 MVP lengkap             :done, a2, 2026-09-28, 2026-10-05
    MVP terintegrasi                 :milestone, a3, 2026-10-03, 0d
    Sprint 3 peran, mutu, domain     :done, a5, 2026-10-12, 2026-10-19

    section Dokumentasi dan mutu
    Requirements dan aturan          :done, d1, 2026-09-18, 2026-10-18
    CI di setiap PR                  :milestone, d2, 2026-10-02, 0d
    UML dan ERD                      :done, d3, 2026-10-16, 1d
    Pengujian empat level            :done, d4, 2026-10-16, 2026-10-18
```

### Selisih rencana dan realisasi

| Hal | Rencana | Realisasi | Selisih |
|---|---|---|---|
| Aplikasi terintegrasi dan bisa didemokan | 27 September | 3 Oktober (#43) | 6 hari terlambat |
| Cakupan MVP | Tiga tab: Ringkasan, Analisis Lokasi, Keuangan | Semua modul ERP, termasuk Penjualan, Pengadaan, Inventaris, dan SDM yang semula P2 | Cakupan lebih luas dari rencana |
| Pembagian kerja | Lima vertikal paralel, satu orang satu fitur | Sprint 1 dikerjakan empat orang; sprint 2 dan 3 dikerjakan terutama oleh satu orang | Integrasi menumpuk di satu orang |
| Hari tanpa commit | Tidak direncanakan | 28 September sampai 1 Oktober, dan 4 sampai 14 Oktober | Dua jeda, pekerjaan berlanjut dalam beberapa hari yang padat |

## 2. Work Breakdown Structure

| Kode | Paket kerja | Status | Rujukan |
|---|---|---|---|
| **1** | **Data dan analisis lokasi** | | |
| 1.1 | Pengumpulan data nasional: BPS, InaRISK, ZNT, RDTR, Pusiknas, OSM | Selesai | Repositori data, 7 sampai 13 September |
| 1.2 | Layanan penilaian lokasi (Python, lima dimensi, bobot per profil bangunan) | Selesai | #7, #18 |
| 1.3 | Evaluasi lokasi, regulasi, dan berita | Selesai | Fondasi, #10, #58, #64, #70 |
| 1.4 | Peta, lapisan data, perbandingan, laporan PDF | Selesai | #6, #27, #28, #50 |
| **2** | **ERP proyek** | | |
| 2.1 | Proyek: tahap, RAB, izin, tim, progres berbobot, kurva S, laporan harian | Selesai | #13, #26, #48, #61, #65 |
| 2.2 | Kontrak kontraktor: termin, retensi, adendum | Selesai | #71, #72 |
| 2.3 | Keuangan: anggaran, kas, jurnal, buku besar | Selesai | #31, #38, #41 |
| 2.4 | Tagihan: faktur, piutang, hutang, pembayaran, on-chain | Selesai | #25, #36, #43, #52 |
| 2.5 | Penjualan: prospek, unit, kontrak, cicilan | Selesai | #14, #29, #35 |
| 2.6 | Pengadaan, inventaris, aset, dan alat | Selesai | #14, #16, #30, #34, #37 |
| 2.7 | SDM: karyawan, absensi, penggajian | Selesai | #16, #34 |
| 2.8 | Laporan PDF, XLSX, CSV | Selesai | #45 |
| **3** | **Platform** | | |
| 3.1 | Auth, sesi, dan hak akses 7 peran | Selesai | #1, #44, #51, #62, #68 |
| 3.2 | Data master | Selesai | #5, #49 |
| 3.3 | Docker, CI, demo online, pemeriksaan kerentanan | Selesai | Fondasi, #23, #66, #74 |
| **4** | **Dokumentasi dan mutu** | | |
| 4.1 | Requirements | Berjalan terus | [requirements](../requirements) |
| 4.2 | Desain UML dan ERD | Selesai | #55 |
| 4.3 | Pengujian unit, integration, system, acceptance | Selesai, 1 skenario acceptance (AT-13, butuh dompet Sepolia) belum dijalankan | #57, #66, #67 |
| 4.4 | Risiko, keamanan, etika | Selesai | #69 |
| 4.5 | Proses SDLC dan jadwal | Dokumen ini | |
| **5** | **Presentasi akhir** | Berjalan | #78 |

## 3. Peran dan tanggung jawab

Pembagian peran mengikuti rencana 21 September di `PROJECT_HANDOFF.md`
repositori data, dilengkapi bagian yang dikerjakan sejak itu.

| Anggota | Peran dan fokus kerja |
|---|---|
| Derick | Pemilik produk; layanan penilaian lokasi, peta, ruang kerja proyek, layar ERP, CI, dokumentasi |
| Jonathan Andrew Saleh | Pengumpulan data, auth, skema database, Docker, kerangka frontend ERP, konteks deskriptif lokasi |
| Moses | Pipeline layanan penilaian, lapisan service dan controller backend |
| Ian | Site intelligence: regulasi, berita, evaluasi lokasi, halaman Analisis Lokasi |
| Erick | Frontend ERP |

Sisa pekerjaan dilacak sebagai issue di milestone
[AoL akhir](https://github.com/jojondrw/BanguninAjaApp/milestone/1) (#78 sampai
#85). Setiap anggota mengambil issue, mengerjakannya lewat branch dan PR
sendiri, dan setiap PR ditinjau minimal satu anggota lain.

## 4. Aktivitas per minggu

| Minggu | Commit repositori data | Commit repositori aplikasi |
|---|---|---|
| 7 sampai 13 September | 34 | 0 |
| 14 sampai 20 September | 0 | 8 |
| 21 sampai 27 September | 1 | 8 |
| 28 September sampai 4 Oktober | 0 | 107 |
| 5 sampai 11 Oktober | 0 | 0 |
| 12 sampai 18 Oktober | 0 | 82 |

Angka commit tanpa merge commit. Repositori aplikasi juga punya 104 merge commit
dari 72 pull request.
