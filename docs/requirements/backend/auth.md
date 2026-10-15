# Slice Auth

## Dua jenis token

| Token | Umur | Bentuk | Disimpan di client |
|---|---|---|---|
| Access token | 15 menit | JWT HS256, subject berisi id user | Memori aplikasi frontend |
| Refresh token | 7 hari | String acak 32 byte, base64url | Cookie HttpOnly |

Access token dipakai di setiap permintaan lewat header `Authorization: Bearer ...`.
Karena umurnya pendek, kebocorannya berdampak terbatas.

Refresh token sengaja bukan JWT. Token ini harus bisa dicabut sewaktu-waktu, dan
JWT tidak bisa dicabut tanpa menyimpan daftar di server. Karena tetap perlu
disimpan di database, memakai string acak lebih sederhana dan lebih aman.

## Penyimpanan refresh token

Di database hanya disimpan hasil SHA-256 dari refresh token, bukan nilai aslinya.
Kalau isi database bocor, token di dalamnya tidak bisa dipakai masuk.

Tabel `refresh_tokens` menyimpan `user_id`, `token_hash`, `expires_at`, dan
`revoked_at`. Token dianggap sah hanya kalau `revoked_at` masih kosong dan
`expires_at` belum lewat.

## Rotasi dan deteksi pemakaian ulang

Setiap kali `/api/auth/refresh` dipanggil, token lama dicabut dan diganti token
baru. Satu refresh token hanya berlaku sekali pakai.

Kalau ada permintaan memakai token yang sudah dicabut, itu tanda token pernah
dicuri. Dalam keadaan itu semua refresh token milik user tersebut ikut dicabut,
sehingga pencuri dan pemilik sama-sama terlempar keluar dan pemilik harus masuk
lagi.

## Aturan cookie

Nama cookie diatur lewat `REFRESH_COOKIE_NAME`.

| Atribut | Nilai | Alasan |
|---|---|---|
| HttpOnly | selalu true | JavaScript tidak bisa membacanya, jadi script yang disusupkan ke halaman tidak bisa mencuri token |
| Path | `/api/auth` | Cookie tidak ikut terkirim di permintaan lain, jadi permukaan serangan lebih kecil |
| SameSite | `lax` saat development | Frontend dan backend sama-sama di localhost, jadi masih dihitung satu site |
| Secure | `false` saat development, `true` saat production | Di laptop masih HTTP polos, di server wajib HTTPS |

Refresh token dilarang disimpan di localStorage maupun sessionStorage. Keduanya
bisa dibaca JavaScript mana pun yang berjalan di halaman.

## CORS

Frontend berjalan di `http://localhost:5173`, backend di port lain, jadi browser
menganggapnya beda origin. Karena alur login memakai cookie, dua hal wajib ada:

1. Backend mengirim `Access-Control-Allow-Credentials: true` dan memantulkan
   origin yang diizinkan satu per satu. Wildcard `*` tidak boleh dipakai bersama
   credentials.
2. Frontend mengirim permintaan dengan `credentials: 'include'`.

Daftar origin yang diizinkan diatur lewat `CORS_ALLOWED_ORIGINS`, dipisah koma.

## Endpoint

Semua endpoint berada di bawah `/api/auth`.

| Method | Path | Butuh access token | Keterangan |
|---|---|---|---|
| POST | `/register` | tidak | Membuat akun. Membalas data user, tidak langsung membuat sesi |
| POST | `/login` | tidak | Membalas access token di body dan menaruh refresh token di cookie |
| POST | `/refresh` | tidak | Membaca cookie, mencabut token lama, membalas access token baru |
| POST | `/logout` | tidak | Mencabut refresh token dan menghapus cookie |
| GET | `/me` | ya | Membalas profil pemilik access token, termasuk peran dan hak akses |

### POST /api/auth/register

```json
{ "name": "Jonathan", "email": "jonathan@example.com", "password": "rahasia123" }
```

Balasan 201 berisi `id`, `name`, `email`, `active`, `role`, `permissions`, `createdAt`.
Pendaftar pertama di sistem kosong otomatis menjadi `admin`. Pendaftar berikutnya
**belum punya peran** (`role: null`, `permissions: {}`) sampai admin memberinya
peran lewat `PATCH /api/users/:id/role`.

Alasannya: tautan demo bersifat publik, jadi siapa pun bisa mendaftar. Dulu
pendaftar baru menjadi `pengamat`, dan pengamat bisa membaca semua modul termasuk
gaji dan kas. Sekarang akun tanpa peran:

1. tetap bisa `login`, `refresh`, `logout`, dan membaca `/auth/me`;
2. ditolak 403 `forbidden` di semua modul oleh `middleware.Authorization`,
   karena `access.Allows` tidak mengenal peran kosong;
3. di frontend melihat layar "Akun menunggu persetujuan" (email, tombol
   "Periksa lagi" yang mengambil ulang `/auth/me`, dan "Keluar"), bukan aplikasi.

Admin melihat pendaftar yang menunggu di halaman Pengguna (saringan
"Menunggu peran", `GET /api/users?role=none`).

### POST /api/auth/login

```json
{ "email": "jonathan@example.com", "password": "rahasia123" }
```

Balasan 200:

```json
{
  "data": {
    "accessToken": "...",
    "accessTokenExpiresAt": "2026-09-18T10:15:00Z",
    "user": {
      "id": "...", "name": "...", "email": "...", "active": true,
      "role": { "code": "keuangan", "name": "Keuangan" },
      "permissions": { "finance": "write", "billing": "write", "projects": "read" },
      "createdAt": "..."
    }
  }
}
```

Refresh token tidak muncul di body. Token itu dikirim sebagai cookie.

### POST /api/auth/refresh

Tidak ada body. Browser mengirim cookie secara otomatis. Balasannya sama dengan
login. Kalau cookie tidak ada, sudah kedaluwarsa, atau sudah dipakai, balasannya
401 dan cookie dihapus.

### POST /api/auth/logout

Tidak ada body. Selalu 204, termasuk kalau cookie sudah tidak ada, supaya tidak
bisa dipakai menebak status sesi orang lain.

## Kode error

| Kode | Status | Muncul saat |
|---|---|---|
| `invalid_payload` | 400 | Body tidak lengkap atau formatnya salah |
| `invalid_credential` | 401 | Email tidak terdaftar atau kata sandi salah |
| `invalid_refresh_token` | 401 | Cookie hilang, kedaluwarsa, atau sudah dicabut |
| `missing_access_token` | 401 | Header Authorization kosong |
| `invalid_access_token` | 401 | Access token rusak atau kedaluwarsa |
| `account_inactive` | 401 / 403 | Akun dinonaktifkan admin. 403 saat login, 401 di refresh, `/me`, dan semua endpoint modul |
| `user_not_found` | 401 | Pemilik access token sudah tidak ada |
| `forbidden` | 403 | Peran tidak punya akses baca atau ubah ke modul yang diminta |
| `email_already_used` | 409 | Email sudah dipakai akun lain |
| `unknown_role` | 422 | Kode peran di `PATCH /api/users/:id/role` tidak ada di matriks |
| `last_admin` | 422 | Perubahan akan meninggalkan sistem tanpa admin aktif |
| `internal_error` | 500 | Kesalahan tak terduga di server |

Email yang tidak terdaftar dan kata sandi yang salah sengaja dibalas dengan kode
dan pesan yang sama, supaya halaman login tidak bisa dipakai memeriksa email mana
saja yang punya akun.

## Peran dan hak akses

Hak akses ditentukan per peran, bukan per orang. Matriksnya hanya ada di satu
tempat, `backend/internal/shared/access/matrix.go`, dan dipakai bersama oleh
middleware, `/api/auth/me`, seeding peran, dan tes. Frontend membaca hasilnya dari
`permissions` di profil, jadi tidak menyimpan salinan matriks.

Setiap modul punya tiga tingkat: tidak ada, `read` (GET/HEAD/OPTIONS), dan
`write` (POST/PUT/PATCH/DELETE, sudah termasuk read).

| Modul | Prefix API | admin | manajer_proyek | keuangan | pengadaan | penjualan | sdm | pengamat |
|---|---|---|---|---|---|---|---|---|
| `projects` | `/projects` | write | write | read | read | read | read | read |
| `locations` | `/locations`, `/site`, `/scoring`, `/regulation`, `/news` | write | write | - | - | - | - | read |
| `finance` | `/finance` | write | read | write | - | - | - | read |
| `billing` | `/billing` | write | - | write | - | read | - | read |
| `sales` | `/sales` | write | - | read | - | write | - | read |
| `procurement` | `/procurement` | write | read | read | write | - | - | read |
| `inventory` | `/inventory`, `/assets`, `/equipment` | write | write | - | write | - | - | read |
| `hr` | `/hr` | write | - | - | - | - | write | read |
| `master` | `/master` | write | read | read | read | read | read | read |
| `reports` | `/reports` | write | write | write | read | read | read | read |
| `users` | `/users` | write | - | - | - | - | - | - |

Keputusan yang melebihi daftar awal:

1. `projects: read` diberikan ke semua peran karena proyek adalah dimensi bersama.
   Hampir setiap formulir (karyawan, gudang, PO, unit, faktur) memilih proyek lewat
   `SearchSelect`, dan tanpa akses baca pilihan itu gagal.
2. `master: read` diberikan ke semua peran karena satuan, akun, dan wilayah dipakai
   sebagai pilihan di banyak modul. Mengubah data master hanya untuk admin.
3. `manajer_proyek` mendapat `finance: read` supaya tab Keuangan di ruang kerja
   proyek (anggaran dan realisasi per proyek) dan Ringkasan bisa dibuka.
4. `penjualan` mendapat `billing: read` untuk seluruh modul Tagihan, karena izin
   dibaca per prefix dan piutang berada di bawah `/billing`.
5. `reports`: peran yang mengelola uang dan proyek boleh membuat laporan, peran
   lain hanya membaca.

### Penegakan

`middleware.Authorization` dipasang sekali di grup `/api` pada `cmd/api/main.go`.
Urutannya:

1. Ambil segmen pertama setelah `/api` dari rute yang cocok (`c.FullPath()`).
   Segmen `auth` dilewati tanpa pemeriksaan, jadi login, refresh, logout, dan
   `/auth/me` tetap seperti sebelumnya.
2. Baca access token. Kalau kosong atau rusak, balas 401 seperti middleware
   autentikasi. User id disimpan di context, sehingga `middleware.Authentication`
   milik modul tidak mengurai token dua kali.
3. Ambil peran dan status aktif user lewat satu query berindeks
   (`users` primary key, `LEFT JOIN role` primary key). Hasilnya disimpan 30 detik
   di memori per user id (`access.Directory`) dan langsung dihapus saat admin
   mengubah peran atau status aktif user itu.
4. User nonaktif dibalas 401 `account_inactive`. User tanpa peran (pendaftar yang
   menunggu persetujuan) tidak punya akses ke modul mana pun dan dibalas 403.
5. Kalau peran tidak punya tingkat yang dibutuhkan, balas 403 `forbidden`
   dengan pesan seperti "Peran Anda tidak punya akses mengubah data modul Keuangan".

Prefix baru yang belum terdaftar di matriks hanya bisa dipakai admin sampai
didaftarkan. Ini sengaja, supaya modul baru tidak terbuka untuk semua orang.

### Kelola pengguna (khusus admin)

| Method | Path | Body | Keterangan |
|---|---|---|---|
| GET | `/api/users` | - | Query `search` (nama atau email), `role` (kode peran, atau `none` untuk akun yang belum punya peran), `active`, `page`, `pageSize` |
| PATCH | `/api/users/:id/role` | `{ "role": "keuangan" }` | Mengganti peran |
| PATCH | `/api/users/:id/active` | `{ "active": false }` | Menonaktifkan atau mengaktifkan lagi |

Kedua PATCH membalas baris pengguna yang sudah berubah: `id`, `name`, `email`,
`active`, `role`, `createdAt`. Menonaktifkan akun langsung mencabut semua refresh
tokennya, jadi akun itu tidak bisa login maupun refresh lagi.

Perubahan yang akan meninggalkan sistem tanpa admin aktif ditolak 422
`last_admin`, misalnya admin terakhir menurunkan perannya sendiri atau
menonaktifkan dirinya. Pemeriksaan dan perubahannya berjalan dalam satu transaksi
dengan advisory lock, jadi dua admin yang saling menurunkan pada saat bersamaan
tidak bisa membuat sistem kehilangan admin.

### Bootstrap dan migrasi

`cmd/migrate` dan `cmd/seed` meng-upsert ketujuh peran berdasarkan `code`.
Setelah itu `cmd/migrate` memberi peran `admin` ke user yang `role_id`-nya masih
kosong, **hanya kalau belum ada satu user pun yang punya peran**. Itu keadaan
basis data dari sebelum fitur peran ada, dan tim yang sudah terdaftar saat itu
tetap mendapat akses penuh. Begitu ada user berperan, langkah ini tidak mengubah
apa pun, jadi pendaftar baru yang sengaja dibiarkan tanpa peran tidak ikut
menjadi admin setiap kali migrasi dijalankan. Keduanya idempoten: menjalankan
migrasi dua kali tidak mengubah apa pun pada putaran kedua.

Pendaftaran baru menghitung jumlah user di dalam transaksi yang memegang advisory
lock. Sistem kosong menghasilkan `admin`, selebihnya tanpa peran. Kalau tabel
`role` belum diisi saat pendaftar pertama masuk, pendaftaran gagal dengan 500
supaya migrasi yang terlewat langsung ketahuan.

`cmd/seed-demo` tetap membuat akun demo menjadi admin: akun itu didaftarkan lewat
`Register` (tanpa peran kalau sudah ada user lain), lalu langsung dinaikkan
dengan `ChangeRole` ke `admin`.

## Kata sandi

Kata sandi di-hash memakai bcrypt dengan cost bawaan. Panjang minimal 8 karakter
dan maksimal 72 karakter, karena bcrypt memotong masukan setelah 72 byte.

Kata sandi tidak pernah dicatat di log dan tidak pernah ikut di response.

## Environment variable

| Nama | Contoh | Keterangan |
|---|---|---|
| `JWT_ACCESS_SECRET` | string acak 48 byte | Wajib diisi, aplikasi menolak jalan tanpa ini |
| `JWT_ACCESS_TTL` | `15m` | Umur access token |
| `JWT_REFRESH_TTL` | `168h` | Umur refresh token |
| `JWT_ISSUER` | `bangunin-aja` | Ikut diperiksa saat token dibaca |
| `REFRESH_COOKIE_SAMESITE` | `lax` atau `none` | `none` dipakai kalau frontend dan backend beda domain |
| `COOKIE_SECURE` | `false` di laptop, `true` di server | Cookie hanya dikirim lewat HTTPS |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Dipisah koma kalau lebih dari satu |
