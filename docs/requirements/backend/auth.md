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
| GET | `/me` | ya | Membalas profil pemilik access token |

### POST /api/auth/register

```json
{ "name": "Jonathan", "email": "jonathan@example.com", "password": "rahasia123" }
```

Balasan 201 berisi `id`, `name`, `email`, `createdAt`.

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
    "user": { "id": "...", "name": "...", "email": "...", "createdAt": "..." }
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
| `email_already_used` | 409 | Email sudah dipakai akun lain |
| `internal_error` | 500 | Kesalahan tak terduga di server |

Email yang tidak terdaftar dan kata sandi yang salah sengaja dibalas dengan kode
dan pesan yang sama, supaya halaman login tidak bisa dipakai memeriksa email mana
saja yang punya akun.

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
