# Eksplorasi Zero-Knowledge Proof untuk BanguninAja

Pertanyaannya: adakah masalah nyata di BanguninAja yang paling tepat diselesaikan dengan zero-knowledge proof (ZK), bukan sekadar teknologi yang dicari-cari masalahnya?

## Syarat sebuah kasus layak memakai ZK

ZK membuktikan sebuah pernyataan tentang data tanpa membuka datanya. Jadi kasusnya baru layak kalau keempat syarat ini terpenuhi:

1. **Ada pihak luar yang perlu yakin.** Kalau hanya pengembang sendiri yang melihat, cukup laporan biasa.
2. **Datanya rahasia bagi pihak luar itu.** Kalau boleh dibuka, cukup dibuka.
3. **Datanya sudah terkunci sebelum dibuktikan.** ZK atas data yang bisa diubah pengembang kapan saja tidak membuktikan apa-apa. Syarat ini sekarang terpenuhi oleh jejak bukti blockchain (slice `evidence`): sidik jari data tercatat di Sepolia sebelum bukti ZK dibuat.
4. **Pernyataannya sederhana dan berguna**, misalnya "semua nilai ≥ X" atau "jumlahnya ≥ Y", supaya sirkuitnya masuk akal dibuat.

## Kandidat yang dinilai

| Kandidat | Pihak luar | Rahasia? | Terkunci? | Hasil |
|---|---|---|---|---|
| **Upah pekerja proyek ≥ upah minimum** | Pemberi kerja, auditor, pekerja | Ya, gaji tiap orang adalah data pribadi | Bisa, penggajian ikut dicatat | **Layak, direkomendasikan** |
| Dana pembeli unit cukup untuk menyelesaikan proyek | Pembeli unit | Ya, pembukuan lengkap dan data pembeli lain | Bisa, pembayaran sudah tercatat | Layak, tapi berat; tahap berikutnya |
| Progres proyek ≥ syarat tertentu | Pembeli, pemberi kerja | Tidak, progres boleh dibuka | Sudah | Tidak perlu ZK, cukup jejak bukti |
| Kelayakan pembeli rumah subsidi (penghasilan, rumah pertama) | Bank, pemerintah | Ya | **Tidak**: sumber datanya harus ditandatangani instansi, infrastrukturnya belum ada | Tidak layak sekarang |
| Lokasi lahan tanpa membuka koordinat | Investor | Kadang | Tidak relevan | Masalahnya tidak ada; lokasi justru dibuka saat perizinan |

## Rekomendasi: bukti upah minimum tanpa membuka gaji

### Masalahnya nyata

- Kontraktor wajib membayar pekerja paling sedikit upah minimum kabupaten/kota (UMK) yang ditetapkan tiap tahun (PP 36/2021 tentang Pengupahan, terakhir diubah PP 49/2025).
- Pemberi kerja seperti yayasan rumah sakit di demo RSIA Cirebon punya kepentingan memastikannya, karena pelanggaran upah di proyeknya ikut mencoreng nama mereka.
- Tapi gaji tiap pekerja adalah data pribadi (UU 27/2022 tentang Pelindungan Data Pribadi), dan pengembang juga tidak ingin struktur upahnya terbuka ke pemberi kerja. Hari ini pilihannya hanya membuka slip gaji atau tidak sama sekali.

### Rancangannya

1. **Daun Merkle penggajian.** Tiap baris penggajian proyek (pekerja, periode, upah) menjadi daun dengan *salt* acak, lalu akarnya ikut dicatat ke blockchain lewat slice `evidence`, sama seperti bukti lain.
2. **Bukti ZK untuk pemberi kerja.** Pengembang membuat bukti: "semua N daun di akar R untuk proyek P periode T punya upah ≥ UMK kota proyek". Pemberi kerja memverifikasinya di halaman verifikasi yang sama dengan tagihan termin, tanpa melihat satu angka gaji pun.
3. **Kuitansi untuk tiap pekerja.** Tiap pekerja menerima bukti keanggotaan Merkle (Merkle proof) untuk barisnya sendiri. Ia bisa memastikan gajinya yang sebenarnya ikut dihitung dalam akar yang tercatat.

Langkah 3 menutup celah terbesar ZK di sini, yaitu "data yang dibuktikan adalah data karangan pengembang": kalau pengembang mencatat upah lebih tinggi dari yang dibayar, pekerjanya sendiri yang bisa membantah.

### Kelayakan teknis

- Sirkuitnya sederhana: N perbandingan ≥ dan satu perhitungan akar Merkle. Untuk puluhan pekerja per proyek, ukurannya kecil.
- Pustaka: **gnark** (Go, cocok dengan backend) untuk membuat bukti Groth16. Verifikasi bisa di server untuk demo, lalu di browser atau kontrak Sepolia supaya tidak bergantung pada server pengembang.
- Hash di dalam sirkuit perlu yang ramah sirkuit (Poseidon/MiMC), bukan keccak, jadi akar penggajian dihitung terpisah dari akar bukti lain lalu dimasukkan sebagai satu daun.

### Batasannya (ditulis jujur)

- ZK membuktikan pernyataan tentang data yang tercatat, bukan bahwa data itu benar. Kebenarannya bergantung pada kuitansi pekerja (langkah 3).
- Daftar pekerja tetap dari pengembang. Pekerja yang sama sekali tidak dicatat tidak tertangkap oleh bukti ini; itu ranah absensi dan pengawasan lapangan.
- Butuh tabel UMK per kota per tahun sebagai data publik di sirkuit.

## Kandidat kedua: dana pembeli mencukupi penyelesaian proyek

Masalahnya juga nyata: pembeli unit membayar uang muka dan cicilan sebelum rumahnya selesai, dan kasus perumahan mangkrak membuat mereka menanggung risikonya. Buktinya akan berbunyi "cicilan yang diterima untuk proyek P sudah dipakai untuk biaya konstruksi P minimal sekian persen", tanpa membuka pembukuan dan data pembeli lain.

Tapi pernyataan ini butuh definisi akuntansi yang disepakati (kas mana yang dihitung milik proyek, biaya mana yang sah) dan sirkuit penjumlahan atas ratusan transaksi. Ini lebih tepat setelah bukti upah minimum berjalan.

## Status: prototipe sudah berjalan (18 Oktober 2026)

Slice `internal/wageproof`, kartu **Bukti upah minimum** di tab Keuangan proyek, halaman publik `/verifikasi-upah/:token` (pemberi kerja) dan `/kuitansi-upah/:token` (pekerja).

| Bagian | Isi |
|---|---|
| Sirkuit | gnark v0.16, Groth16 di kurva BN254, paling banyak 16 pekerja per bukti. Publik: akar Merkle, UMK, jumlah pekerja. Rahasia: tiap upah, hari kerja, kunci pekerja, salt |
| Aturan yang dibuktikan | `upah × 25 ≥ UMK × hari kerja` untuk tiap pekerja (bulanan dihitung 25 hari, harian memakai hari hadir dari absensi), hari kerja 1–31 |
| Akar Merkle | MiMC (ramah sirkuit), 16 daun, kedalaman 4. Akar ikut dicatat ke blockchain sebagai bukti jenis "Akar penggajian" |
| Kuitansi | Tiap pekerja dapat tautan berisi upahnya, salt, dan jalur Merkle; server memeriksa jalurnya sampai ke akar |
| Data UMK 2026 | Kota Bandung Rp4.737.678, Kota Cirebon Rp2.878.646 (Kepgub Jabar 561/Kep.862-Kesra/2025); Kota Depok dan Kota Makassar memakai angka yang lebih tinggi dari dua sumber yang berbeda, supaya tidak ada proyek yang lolos karena salah ketik |

Hasil di data demo:

- RSIA Cirebon September 2026: bukti dibuat untuk 3 pekerja (ukuran bukti 164 byte) dan terverifikasi.
- Griya Asri Bandung: **ditolak**. Satu pekerja harian dibayar Rp3.795.000 untuk 23 hari, padahal UMK mewajibkan paling sedikit Rp4.358.664. Hanya pengembang yang melihat nama dan angkanya.

Batasan prototipe:

- **Trusted setup.** Kunci Groth16 dibuat sekali oleh server pengembang dan disimpan di database. Siapa pun yang memegang sisa acak dari proses itu secara teori bisa memalsukan bukti, dan di prototipe itu adalah pengembang sendiri. Untuk dipakai sungguhan, kunci harus dibuat lewat upacara multi-pihak (MPC), misalnya dengan pemberi kerja ikut berkontribusi, atau beralih ke PLONK dengan SRS publik.
- Verifikasi dijalankan di server. Data bukti dan kunci verifikasi ditampilkan di halaman supaya bisa diperiksa mandiri, tetapi verifikasi di browser atau lewat kontrak di Sepolia belum dibuat.
- Kuitansi dibagikan pengembang. Kalau pengembang tidak memberikan kuitansi ke pekerja, perlindungan dari pekerja tidak berjalan.

## Langkah berikutnya

1. Kunci dari upacara multi-pihak atau PLONK dengan SRS publik.
2. Verifikasi di browser atau kontrak Sepolia, supaya tidak bergantung pada server pengembang.
3. Kandidat kedua: dana pembeli mencukupi penyelesaian proyek.
