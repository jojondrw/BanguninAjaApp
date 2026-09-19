# Aturan Index

## Prinsip

Index mempercepat pembacaan dan memperlambat penulisan. Setiap index menambah
pekerjaan pada setiap INSERT, UPDATE, dan DELETE, serta memakan ruang disk.
Jadi index dipasang karena ada alasan, bukan karena kolomnya kelihatan penting.

Sebuah kolom diberi index kalau memenuhi salah satu dari ini:

1. Dipakai di `WHERE` pada query yang sering dijalankan.
2. Dipakai sebagai kunci `JOIN`, termasuk semua foreign key.
3. Dipakai di `ORDER BY` pada daftar yang berhalaman.
4. Dipakai untuk menjamin keunikan.

Sebuah kolom tidak diberi index kalau:

1. Tabelnya kecil dan jarang tumbuh, misalnya tabel master satuan atau peran.
   Postgres lebih cepat membaca seluruh tabel kecil daripada lewat index.
2. Isinya cuma sedikit ragam dan tersebar rata, misalnya kolom boolean yang
   separuhnya true. Kecuali dipakai sebagai index parsial.
3. Hanya dipakai untuk ditampilkan, tidak pernah dicari maupun difilter.

## Foreign key wajib diberi index

GORM membuat constraint foreign key, tapi **tidak** otomatis membuat index untuk
kolomnya. Padahal kolom itu hampir selalu dipakai untuk join dan filter, dan
tanpa index, menghapus baris induk memaksa Postgres membaca seluruh tabel anak.

Jadi setiap kolom `*_id` di entity wajib ditulis dengan tag `index`.

## Urutan kolom pada index gabungan

Kolom yang dibandingkan dengan sama dengan diletakkan lebih dulu, baru kolom
rentang. Index `(project_id, status)` bisa melayani query yang menyaring
`project_id` saja, tapi index `(status, project_id)` tidak bisa melayani query yang
hanya menyaring `project_id`.

Kalau sebuah daftar selalu disaring dengan satu kolom lalu diurutkan dengan kolom
lain, keduanya digabung dalam satu index dengan kolom penyaring di depan.

## Index parsial

Kalau query hampir selalu menyaring sebagian kecil baris, index cukup dibuat
untuk bagian itu saja. Index parsial jauh lebih kecil dan lebih cepat.

Contoh yang dipakai di sini, pada tabel `refresh_tokens`:

```sql
CREATE INDEX idx_refresh_tokens_active
  ON refresh_tokens (user_id)
  WHERE revoked_at IS NULL;
```

Token yang sudah dicabut tidak pernah dicari lagi, jadi tidak perlu ikut masuk
index.

## Pencarian nama dan kode

Kolom nama dan kode dicari dengan pola `ILIKE '%kata%'`. Index biasa tidak
terpakai untuk pola yang diawali persen, jadi kolom seperti itu memakai index
trigram:

```sql
CREATE INDEX idx_project_name_trgm
  ON project USING gin (name gin_trgm_ops);
```

Alasan memilih GIN dan bukan GiST ada di `ekstensi.md`.

Kolom kode yang selalu dicari dari awal, misalnya `PRJ-01`, cukup memakai index
biasa dengan `text_pattern_ops` karena polanya `code LIKE 'PRJ%'`. Kalau kodenya
juga dicari dari tengah, barulah ikut trigram.

## Kolom spasial

Kolom `geometry` memakai index GiST:

```sql
CREATE INDEX idx_saved_location_point ON saved_location USING gist (point);
```

Tanpa index ini, penyaringan berdasarkan kotak peta membaca seluruh tabel, dan
itu fatal untuk data lahan yang jumlahnya jutaan baris.

## Rencana index per modul

Daftar ini menjadi acuan saat entity dibuat. Kolom disebut memakai nama yang
nanti dipakai di entity.

| Tabel | Index | Alasan |
|---|---|---|
| `users` | unik pada `email` | Dipakai setiap login |
| `refresh_tokens` | unik pada `token_hash`, parsial pada `user_id` saat aktif | Dicari di setiap refresh |
| `project` | unik `code`, biasa `status`, trigram `name` | Daftar proyek disaring status, dicari lewat nama |
| `unit` | gabungan `(project_id, status)` | Halaman unit selalu menyaring proyek lalu status |
| `purchase_order` | `vendor_id`, gabungan `(status, date)` | Daftar pengadaan disaring status lalu diurut tanggal |
| `stock_movement` | gabungan `(material_id, date)` | Kartu stok dibaca per material urut tanggal |
| `installment` | gabungan `(status, due_date)` | Halaman tagihan mencari yang belum lunas dan hampir jatuh tempo |
| `receivable`, `payable` | gabungan `(status, due_date)` | Sama seperti cicilan |
| `attendance` | gabungan `(employee_id, date)`, unik `(employee_id, date)` | Satu orang satu baris per hari |
| `journal_entry` | `date`, `account_id` | Buku besar dibaca per rentang tanggal dan per akun |
| `saved_location` | GiST pada `point`, `user_id` | Peta menyaring berdasarkan area tampilan |
| `employee` | trigram `name`, `status` | Daftar karyawan dicari lewat nama |

## Cara memeriksa apakah index terpakai

Jangan menebak. Jalankan:

```sql
EXPLAIN ANALYZE SELECT ... ;
```

Kalau hasilnya berisi `Seq Scan` pada tabel besar padahal ada penyaring, berarti
index yang dibuat belum tepat atau query-nya menghalangi index, misalnya karena
kolomnya dibungkus fungsi.

Index yang tidak pernah terpakai bisa dilihat dari:

```sql
SELECT relname, indexrelname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan;
```

Index dengan `idx_scan` nol setelah aplikasi dipakai sebaiknya dihapus.

## Aturan menulis query di repository

1. Jangan membungkus kolom dengan fungsi di sisi kiri perbandingan. Tulis
   `date >= ? AND date < ?`, bukan `DATE(date) = ?`, karena bentuk
   kedua membuat index tidak terpakai.
2. Ambil kolom seperlunya untuk daftar yang panjang, jangan selalu `SELECT *`.
3. Daftar yang panjang wajib memakai `LIMIT` dan `OFFSET` atau kursor.
4. Hindari query di dalam perulangan. Ambil sekaligus dengan `IN` atau
   `Preload`, supaya tidak terjadi satu query per baris.
