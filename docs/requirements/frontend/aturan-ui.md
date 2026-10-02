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
| Efek kaca buram dan gradasi di luar tempat yang disebut di bawah | Menurunkan keterbacaan, dan tidak menambah informasi apa pun |
| Animasi masuk di setiap elemen | Membuat halaman terasa lambat. Yang beranimasi hanya isi halaman saat dibuka, panel tab, formulir yang dibuka, dan palet perintah |
| Teks abu abu muda di atas putih | Kontrasnya tidak cukup untuk dibaca |
| Memakai `alert()` atau `confirm()` | Tidak bisa digaya, dan memblokir seluruh halaman |

## Bahasa visual

Diperbarui Oktober 2026 mengikuti gaya aplikasi Apple dan aplikasi kerja
seperti Linear dan Mercury: tenang, padat informasi, angka mudah dibaca.

| Unsur | Ketentuan |
|---|---|
| Huruf | `-apple-system` (SF Pro di perangkat Apple), cadangan Inter yang di-host sendiri. Judul diberi tracking negatif, angka memakai `tabular-nums` |
| Warna | Abu netral (`slate` di `index.css` sudah diganti ke abu netral ala Apple), satu aksen navy merek, kuning merek hanya untuk logo dan penanda kecil |
| Permukaan | Kanvas abu, panel konten putih melayang, kartu putih dengan `shadow-hairline` (garis 1px lembut), bukan border tebal |
| Teks sekunder | Minimal `slate-500` di atas putih atau abu muda. `slate-400` hanya untuk ikon dan placeholder |
| Status | Chip berbentuk pil dengan titik warna (`CHIP_CLASS`), selalu disertai kata, tidak hanya warna |
| Pindah bagian halaman | Segmented control (`SectionTabs`/`Tabs`) dengan indikator yang meluncur |
| Saring cepat | Chip filter bulat (`FilterChips`), yang aktif berwarna gelap |
| Tombol | `Button` atau konstanta `BUTTON_*`; mengecil sedikit saat ditekan (`active:scale-[0.97]`) |
| Tabel | Header kalimat biasa (bukan huruf kapital semua), baris disorot saat hover, di ponsel tidak dilipat tapi digulir ke samping |
| Memuat | Skeleton abu, bukan teks "Memuat..." saja |

Tempat yang boleh tembus pandang: header lengket (garis bawahnya baru muncul
saat halaman digulir) dan latar palet perintah. Gradasi hanya di panel merek
halaman Masuk/Daftar.

Gerak: kurva `ease-fluid` (`cubic-bezier(0.32, 0.72, 0, 1)`), 150–300 ms, dan
selalu lewat varian `motion-safe:` supaya hilang kalau pengguna memilih
mengurangi gerakan.

Pintasan: `Ctrl K` / `⌘K` membuka palet perintah untuk lompat ke halaman atau
proyek. Tautan "Lewati ke konten" adalah elemen pertama yang bisa difokus.
