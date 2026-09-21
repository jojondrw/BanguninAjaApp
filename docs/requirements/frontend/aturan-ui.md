# Aturan UI

Acuannya dua daftar yang sudah mapan: delapan aturan emas Shneiderman dan
sepuluh heuristik Nielsen. Keduanya banyak bertumpuk, jadi di sini digabung
menjadi satu daftar kerja. Isinya bukan teori, tapi bentuk nyatanya di aplikasi.

| Aturan | Asal | Bentuknya di sini |
|---|---|---|
| Konsisten | Shneiderman 1, Nielsen 4 | Semua tombol lewat `Tombol`, semua isian lewat `Kolom`. Tidak ada tombol yang ditulis sendiri di satu halaman |
| Jalan pintas untuk yang sudah terbiasa | Shneiderman 2 | Urutan Tab mengikuti urutan isian, Enter mengirim formulir, `autocomplete` diisi supaya pengisi otomatis peramban bekerja |
| Umpan balik yang jelas | Shneiderman 3, Nielsen 1 | Tombol berubah jadi "Sedang masuk" dan terkunci selama proses. Setiap keadaan memuat punya keterangan, bukan layar diam |
| Alur punya penutup | Shneiderman 4 | Selesai mendaftar, pengguna diantar ke halaman masuk lengkap dengan pesan "Akun berhasil dibuat" |
| Cegah kesalahan sebelum terjadi | Shneiderman 5, Nielsen 5 | Isian wajib ditandai, panjang sandi minimal dijaga, tombol kirim terkunci selama proses supaya tidak terkirim dua kali |
| Mudah dibatalkan | Shneiderman 6, Nielsen 3 | Setiap halaman punya jalan keluar: tautan ke halaman lain, dan tombol keluar di dasbor |
| Pengguna yang memegang kendali | Shneiderman 7, Nielsen 3 | Tidak ada pengalihan halaman mendadak kecuali setelah tindakan yang pengguna mulai sendiri |
| Ringankan ingatan | Shneiderman 8, Nielsen 6 | Formulir paling banyak tiga isian. Syarat sandi ditulis di bawah kolomnya, bukan disembunyikan sampai gagal |
| Bahasa yang dikenal pengguna | Nielsen 2 | Semua teks bahasa Indonesia. Tanggal ditulis "19 September 2026", bukan format ISO |
| Tampilan minimal | Nielsen 8 | Satu warna aksen, tanpa hiasan, tanpa ikon yang tidak berarti |
| Pesan kesalahan yang menolong | Nielsen 9 | Pesan datang dari backend dan berbentuk kalimat, misalnya "Email atau kata sandi salah", bukan kode status |
| Bantuan saat dibutuhkan | Nielsen 10 | Keterangan pendek menempel di kolomnya, muncul saat relevan, misalnya "Kurang 3 karakter lagi" |

## Penerapan yang wajib diikuti komponen baru

1. Setiap tombol yang memicu permintaan ke server wajib punya keadaan sedang
   proses, dan wajib terkunci selama itu.
2. Setiap kolom isian wajib punya `label` yang terhubung lewat `htmlFor`, bukan
   placeholder sebagai pengganti label. Placeholder hilang begitu diketik.
3. Pesan kesalahan memakai `role="alert"`, pesan keberhasilan memakai
   `role="status"`, supaya pembaca layar ikut membacakannya.
4. Jangan memakai warna sebagai satu satunya penanda. Kesalahan memakai warna
   merah **dan** kalimat.
5. Fokus keyboard harus terlihat. Gaya `:focus-visible` diatur sekali di
   `index.css` dan tidak boleh dimatikan.
6. Teks tombol memakai kata kerja yang menjelaskan akibatnya: "Daftar", "Masuk",
   "Keluar", bukan "OK" atau "Kirim".

## Yang dihindari

| Jangan | Alasan |
|---|---|
| Emoji sebagai ikon | Bentuknya berbeda di tiap sistem dan dibacakan aneh oleh pembaca layar |
| Efek kaca buram dan gradasi berlebihan | Menurunkan keterbacaan, dan tidak menambah informasi apa pun |
| Animasi masuk di setiap elemen | Membuat halaman terasa lambat |
| Teks abu abu muda di atas putih | Kontrasnya tidak cukup untuk dibaca |
| Memakai `alert()` atau `confirm()` | Tidak bisa digaya, dan memblokir seluruh halaman |
