# Ekstensi PostgreSQL

Image database diganti dari `postgres:18` menjadi `postgis/postgis:18-3.6`.
Postgres polos tidak bisa memasang PostGIS hanya dengan `CREATE EXTENSION`,
karena pustakanya memang tidak ada di dalam image. Aplikasi ini inti kerjanya
memilih lokasi, jadi database tanpa dukungan spasial tidak bisa dipakai.

Ekstensi dipasang oleh `cmd/migrate` lewat `database.EnsureExtensions`, dan
dijalankan sebelum AutoMigrate. Perintahnya memakai `CREATE EXTENSION IF NOT
EXISTS`, jadi aman dijalankan berulang kali.

## Yang dipasang

| Ekstensi | Untuk apa |
|---|---|
| `postgis` | Tipe data `geometry` dan `geography`, plus fungsi spasial seperti `ST_Contains`, `ST_DWithin`, `ST_Area` |
| `pg_trgm` | Pencarian teks yang tahan salah ketik dan tahan potongan kata, dipakai untuk kolom nama dan kode |
| `btree_gist` | Menggabungkan kolom biasa dengan kolom rentang di satu index GiST, dipakai untuk constraint anti tumpang tindih |

## PostGIS

Dipakai untuk kolom lokasi. Titik disimpan sebagai `geometry(Point, 4326)`,
yaitu koordinat lintang bujur yang sama dengan yang dipakai peta web dan dataset
BPN maupun BNPB.

Yang dimungkinkan olehnya:

1. Mencari lahan di dalam sebuah batas kecamatan, bukan sekadar membandingkan
   angka koordinat.
2. Menghitung jarak sebenarnya ke fasilitas terdekat.
3. Menyaring kandidat di dalam kotak peta yang sedang dilihat pengguna.

Tanpa PostGIS, ketiganya harus dihitung di Go setelah semua baris ditarik, dan
itu berarti menarik jutaan baris ke aplikasi hanya untuk membuang hampir
semuanya.

## pg_trgm dan pilihan GiST atau GIN

Trigram memotong teks jadi kepingan tiga huruf. Kata `menteng` menjadi
`men`, `ent`, `nte`, `ten`, `eng`. Karena yang dicocokkan kepingan, pencarian
`%ntengan%` atau salah ketik ringan tetap ketemu.

Ini penting karena index biasa sama sekali tidak menolong pencarian dengan pola
`ILIKE '%kata%'`. Tanpa trigram, Postgres membaca seluruh tabel.

Untuk trigram ada dua jenis index:

| Jenis | Kelebihan | Kekurangan |
|---|---|---|
| GIN | Pencarian lebih cepat, ukuran lebih kecil | Penulisan lebih lambat |
| GiST | Penulisan lebih ringan, bisa dipakai untuk urutan kemiripan dan pencarian tetangga terdekat | Pencarian lebih lambat |

Di aplikasi ini kolom nama dan kode jauh lebih sering dibaca daripada ditulis,
jadi **GIN yang dipakai** untuk pencarian teks. GiST tetap dipakai untuk kolom
spasial, karena di sana GIN memang tidak berlaku.

Catatan untuk tim: Moses menyebut GiST untuk trigram. Secara teknis keduanya
bisa, tapi untuk beban baca berat GIN lebih cepat. Kalau nanti ada kebutuhan
mengurutkan hasil berdasarkan kemiripan, misalnya saran "mungkin maksud Anda",
kolom itu boleh pindah ke GiST.

## btree_gist

Ekstensi ini menyatukan index biasa dengan index GiST dalam satu index. Gunanya
untuk aturan anti tumpang tindih, misalnya satu alat berat tidak boleh dipesan
untuk dua proyek pada rentang tanggal yang beririsan.

Aturan seperti itu ditulis sebagai `EXCLUDE` constraint, dan `EXCLUDE` dengan
campuran kolom id ditambah rentang tanggal hanya bisa dibuat kalau `btree_gist`
terpasang. Ekstensinya dipasang sekarang supaya constraint itu bisa ditambahkan
belakangan tanpa mengubah setelan database lagi.

## Memeriksa hasilnya

```bash
docker compose exec database psql -U $POSTGRES_USER -d $POSTGRES_DB -c "\dx"
```
