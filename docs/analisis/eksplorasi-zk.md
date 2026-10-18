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

- Kontraktor wajib membayar pekerja paling sedikit upah minimum kabupaten/kota (UMK) yang ditetapkan tiap tahun (PP 36/2021 tentang Pengupahan).
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

## Langkah berikutnya

1. Mencatat penggajian ke jejak bukti (daun dengan salt), termasuk halaman kuitansi untuk pekerja.
2. Prototipe sirkuit gnark "semua upah ≥ UMK" untuk satu proyek demo, dengan verifikasi di server.
3. Menampilkan hasil verifikasi di halaman verifikasi pemberi kerja.
