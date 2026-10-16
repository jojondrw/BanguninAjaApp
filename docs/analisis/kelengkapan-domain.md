# Kelengkapan Domain: Seberapa Dekat dengan Perusahaan Konstruksi Sungguhan

Penilaian ini menjawab dua pertanyaan: apakah fitur BanguninAja sudah cukup
untuk developer properti atau kontraktor yang sungguhan, dan kalau fitur
ditambah, bagaimana UI tetap sederhana.

## Jawaban singkat

**Belum ideal, tetapi kekurangannya terarah.** BanguninAja sudah **luas**:
hampir semua fungsi kantor (back office) ada, dari proyek, keuangan, tagihan,
penjualan, pengadaan, inventaris, sampai SDM. Yang masih **dangkal** adalah
bagian yang membuat perusahaan konstruksi berbeda dari perusahaan dagang biasa:
**kendali pelaksanaan di lapangan dan kendali biaya proyek**.

Empat celah paling penting, semuanya bisa ditambahkan tanpa menu baru:

1. Progres proyek dihitung sebagai rata-rata biasa tahap, bukan berbobot nilai pekerjaan
2. Belum ada laporan harian lapangan
3. Belum ada termin pembayaran berdasarkan progres dan retensi
4. Belum ada pekerjaan tambah-kurang (change order / adendum)

### Kemajuan sejak penilaian ini

| Celah | Keadaan |
|---|---|
| Ruang lega di seluruh halaman | Selesai (#60) |
| Progres berbobot RAB dan kurva S | Selesai (#61) |
| Laporan harian lapangan | Selesai (#65), tab Lapangan |
| Termin dan retensi | Selesai (#71), kartu Termin kontraktor di tab Keuangan proyek |
| Pekerjaan tambah-kurang | Selesai (#72), kartu di bawah Termin kontraktor. Pindah dari rencana awal di tab Ringkasan karena adendum mengubah kontrak kontraktor yang ada di tab Keuangan |

## 1. Peta fitur

`Ada` = sudah dipakai sehari-hari, `Dangkal` = ada tapi belum seperti praktik
lapangan, `Belum` = tidak ada.

| Area | Yang dipakai perusahaan nyata | BanguninAja | Status |
|---|---|---|---|
| **Pemilihan lokasi** | Studi kelayakan lahan sebelum membeli | Skor 5 dimensi dari data GIS nasional, regulasi, berita, perbandingan, laporan PDF | Ada (pembeda utama) |
| **Portofolio** | Ringkasan semua proyek untuk direksi | Ringkasan KPI, arus kas 12 bulan, daftar perlu perhatian | Ada |
| **Jadwal** | Tahap pekerjaan, kurva S rencana vs realisasi, baseline | Tahap dengan tanggal dan progres | Dangkal |
| **Progres fisik** | Progres berbobot nilai pekerjaan, dari opname lapangan | Rata-rata biasa progres tahap | Dangkal |
| **Laporan lapangan** | Laporan harian: cuaca, tenaga kerja, alat, pekerjaan, kendala, foto | - | Belum |
| **RAB** | Item pekerjaan, volume, harga satuan, realisasi per item | Item, volume, harga satuan; realisasi hanya per proyek | Dangkal |
| **Perubahan kontrak** | Pekerjaan tambah-kurang, adendum nilai dan waktu | Nilai kontrak bisa diedit langsung, tanpa jejak | Belum |
| **Termin dan retensi** | Penagihan atau pembayaran per termin sesuai progres, retensi 5% sampai masa pemeliharaan selesai | Hutang dan piutang manual, tidak terhubung ke progres | Belum |
| **Perizinan** | PKKPR, PBG, SLF, AMDAL dengan masa berlaku | Ada, dengan status dan tanggal | Ada |
| **Keuangan** | Anggaran, kas, jurnal, buku besar | Ada, jurnal otomatis dari kas | Ada |
| **Pengadaan** | Permintaan, penawaran vendor, pesanan, penerimaan | Permintaan, pesanan, penerimaan dengan stok otomatis; belum ada perbandingan penawaran | Ada |
| **Inventaris dan alat** | Stok per gudang proyek, alat dan perawatan | Ada | Ada |
| **SDM** | Karyawan, absensi, penggajian, pekerja harian dan subkon | Ada, termasuk jenis pekerja harian dan subkontraktor | Ada |
| **Penjualan properti** | Unit, prospek, kontrak, cicilan, KPR | Ada | Ada |
| **Serah terima unit** | BAST, daftar perbaikan, garansi, komplain | - | Belum |
| **Mutu dan K3** | Inspeksi mutu, catatan kecelakaan kerja | - | Belum |
| **Dokumen** | Gambar kerja berversi, RFI | - | Belum |

## 2. Celah prioritas

### P1. Progres berbobot dan kurva S

**Masalah.** Progres proyek sekarang adalah rata-rata biasa semua tahap. Di
proyek RSIA, "Pekerjaan persiapan" (biaya kecil) dihitung sama berat dengan
"Struktur lantai 1-4" (biaya besar), sehingga progres terlihat 53% padahal
sebagian besar nilai pekerjaan belum selesai. Direksi dan bank memakai angka
progres untuk mencairkan dana, jadi angka yang terlalu tinggi berbahaya.

**Praktik lapangan.** Bobot tiap pekerjaan = nilai pekerjaan / total RAB.
Progres proyek = jumlah (bobot x persen selesai). Kurva S membandingkan progres
rencana (dari jadwal) dengan progres nyata per minggu.

**Rancangan minimal.**
- Item RAB dikaitkan ke tahap. Bobot tahap dihitung otomatis dari RAB-nya,
  tidak perlu diisi manual
- Kalau belum ada RAB yang terkait, sistem kembali ke rata-rata biasa dan
  memberi tahu bahwa bobot belum dari RAB
- Setiap perubahan progres disimpan sebagai riwayat, sehingga kurva realisasi
  bisa digambar

**Tempat di UI.** Satu grafik kurva S di tab Ringkasan proyek, menggantikan
ruang kosong di bawah KPI. Tidak ada halaman baru.

### P1. Laporan harian lapangan

**Masalah.** Kantor tidak tahu apa yang terjadi di lokasi proyek kecuali lewat
pesan singkat. Padahal laporan harian adalah dokumen kontrak: dipakai saat
klaim keterlambatan karena hujan, dan sebagai dasar opname progres.

**Rancangan minimal.** Satu laporan per proyek per hari: cuaca pagi dan siang,
jumlah tenaga kerja, alat yang bekerja, pekerjaan yang dikerjakan, kendala.
Foto menyusul setelah ada penyimpanan berkas.

**Tempat di UI.** Satu tab baru "Lapangan" di ruang kerja proyek, berisi daftar
per tanggal. Form terbuka hanya saat menekan "Tulis laporan hari ini".

### P1. Termin dan retensi

**Masalah.** Developer membayar kontraktor bertahap sesuai progres, dan menahan
retensi (umumnya 5%) sampai masa pemeliharaan selesai. Saat ini hutang dicatat
manual tanpa hubungan ke progres, sehingga mudah membayar lebih dulu dari
pekerjaannya.

**Rancangan minimal.** Termin per proyek: persen progres yang disyaratkan,
persen nilai kontrak, retensi. Termin hanya bisa diajukan kalau progres
berbobot sudah mencapai syaratnya, lalu menjadi hutang di modul Tagihan dengan
potongan retensi.

**Tempat di UI.** Satu kartu "Termin" di tab Keuangan proyek.

### P1. Pekerjaan tambah-kurang

**Masalah.** Nilai kontrak dan RAB berubah hampir di setiap proyek, tetapi
sekarang nilainya diedit langsung sehingga tidak ada jejak siapa mengubah,
kapan, dan kenapa.

**Rancangan minimal.** Catatan perubahan berisi alasan, selisih nilai, selisih
waktu, dan status disetujui. Nilai kontrak berjalan = nilai awal + perubahan
yang disetujui.

**Tempat di UI.** Bagian kecil di tab Ringkasan proyek, di bawah informasi
kontrak.

### P2, setelah P1 selesai

- Realisasi biaya per item RAB (transaksi kas dikaitkan ke item RAB)
- Perbandingan penawaran vendor sebelum pesanan
- Serah terima unit (BAST) dan daftar perbaikan
- Inspeksi mutu dan catatan K3
- Dokumen proyek berversi

### Sengaja tidak dibangun

| Fitur | Alasan |
|---|---|
| Penjadwalan CPM penuh dengan dependensi | Pengguna sasaran memakai tahap besar, bukan ribuan aktivitas; alat khusus seperti MS Project lebih tepat |
| BIM atau model 3D | Di luar lingkup sistem informasi manajemen |
| Akuntansi pajak lengkap | Biasanya ditangani konsultan pajak dan perangkat lunak akuntansi |

## 3. Menjaga UI tetap sederhana

Kekhawatiran yang wajar: kalau fitur bertambah, layar jadi ramai. Aturan yang
dipakai supaya itu tidak terjadi:

| Aturan | Penerapan |
|---|---|
| Satu proyek, satu ruang kerja | Fitur proyek baru masuk ke tab ruang kerja proyek, bukan menu samping baru. Menu samping tetap 12 item |
| Maksimal empat KPI per halaman | Sudah dijaga oleh komponen `KpiRow` |
| Form tertutup sampai dibutuhkan | Tombol "Tambah ..." membuka form; daftar tetap jadi tampilan utama |
| Satu grafik per tab | Grafik hanya untuk hal yang dibaca sekilas: arus kas, kurva S |
| Detail di belakang klik | Tabel menampilkan kolom inti; rincian dibuka per baris |
| Ruang lega | Jarak antarbagian, padding kartu, dan lebar maksimum konten diperbesar supaya mata tidak lelah |

Dengan aturan ini, empat fitur P1 menambah **satu tab**, **satu grafik**, dan
**dua kartu**, tanpa menu baru.

## 4. Urutan pengerjaan

1. Ruang lega di seluruh halaman (jarak, padding, lebar maksimum)
2. Progres berbobot RAB dan kurva S
3. Laporan harian lapangan
4. Termin dan retensi
5. Pekerjaan tambah-kurang
