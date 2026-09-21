# Konvensi Kode Backend

## Larangan komentar

Kode di `backend/` ditulis tanpa komentar sama sekali. Kalau sebuah baris butuh
komentar supaya bisa dimengerti, baris itu yang harus diperbaiki, bukan ditambahi
penjelasan.

Cara mengganti komentar:

| Kalau tergoda menulis | Lakukan ini |
|---|---|
| Penjelasan isi blok if yang panjang | Pisahkan jadi fungsi dengan nama yang menjelaskan |
| Penjelasan arti sebuah angka | Jadikan konstanta bernama |
| Penjelasan alur sebuah fungsi | Pecah jadi beberapa fungsi kecil |
| Penjelasan kenapa sebuah keputusan diambil | Tulis di berkas ini atau di dokumen requirements lain |

Pengecualian yang tetap diizinkan: tag struct dan directive resmi Go seperti
`//go:build`, karena keduanya dibaca oleh compiler, bukan manusia.

## Penamaan

1. Nama fungsi memakai kata kerja, contoh `IssueAccess`, `RevokeUserRefreshTokens`.
2. Nama yang mengembalikan boolean memakai bentuk pernyataan, contoh
   `PasswordMatches`, `IsUsable`, `EmailExists`.
3. Hindari singkatan yang tidak umum. `repository` ditulis penuh, bukan `repo`.
4. Nama variabel menjelaskan isinya, bukan tipenya. Pakai `stored`, bukan
   `refreshTokenStruct`.
5. Receiver method cukup satu atau dua huruf dan konsisten dalam satu tipe.
6. Konstanta yang dipakai di satu paket ditulis huruf kecil supaya tidak bocor
   keluar paket.

## Ukuran dan bentuk fungsi

1. Satu fungsi mengerjakan satu hal. Kalau namanya mengandung kata "dan", pecah.
2. Batas wajar satu fungsi sekitar 40 baris. Lebih dari itu biasanya tanda sudah
   waktunya dipecah.
3. Kembalikan lebih awal saat kondisi gagal, jangan menumpuk `else`.
4. Maksimal dua tingkat indentasi di dalam fungsi. Lebih dalam dari itu, pecah.

## Error

1. Error dibungkus dengan konteks memakai `fmt.Errorf("...: %w", err)` supaya
   asalnya bisa dilacak.
2. Error yang dibandingkan memakai `errors.Is` dan `errors.As`, bukan
   perbandingan string.
3. Error milik repository (`errUserNotFound`, `errRefreshTokenNotFound`) tidak
   bocor ke controller. Service yang menerjemahkannya menjadi `apperror`.
4. Nilai error yang sengaja diabaikan ditulis eksplisit dengan `_ =` supaya
   terlihat disengaja.

## DTO

1. Request dan response tinggal di `dto.go` milik slice.
2. Entity tidak pernah dikirim langsung sebagai response. Selalu lewat struct
   response terpisah supaya kolom seperti `PasswordHash` tidak mungkin bocor.
3. Validasi input memakai tag `binding` milik Gin, bukan pengecekan manual di
   dalam service.
4. Nama field JSON memakai camelCase supaya cocok dengan kebiasaan frontend.
5. Field waktu memakai `time.Time` dan dikirim sebagai RFC 3339.

## Bentuk response

Semua response dibungkus amplop yang sama.

Sukses:

```json
{ "data": { "id": "...", "name": "..." } }
```

Gagal:

```json
{ "error": { "code": "invalid_credential", "message": "Email atau kata sandi salah" } }
```

## Database

1. Semua query memakai `WithContext(ctx)` supaya bisa dibatalkan.
2. Dilarang menyusun query dengan penggabungan string. Selalu pakai placeholder
   `?` supaya aman dari SQL injection.
3. Query milik sebuah slice hanya boleh ada di `repository.go` slice itu.
4. Aturan index dan pertimbangan performa ditulis terpisah di
   `docs/requirements/database`.

## Pemeriksaan sebelum commit

```bash
cd backend
gofmt -l .
go vet ./...
go build ./...
```

`gofmt -l .` harus tidak mengeluarkan nama berkas apa pun.
