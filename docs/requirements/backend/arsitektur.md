# Arsitektur Backend

## Ringkasan

Backend memakai Go dengan Gin sebagai HTTP interface dan GORM sebagai ORM.
Struktur folder mengikuti layout standar Go: `cmd` untuk entry point, `internal`
untuk kode yang tidak boleh diimpor dari luar modul.

Arsitektur yang dipakai adalah **Vertical Slice Architecture (VSA)**. Kode
dikelompokkan per fitur, bukan per jenis file. Satu fitur berarti satu folder di
`internal`, dan di dalamnya ada semua lapisan yang dibutuhkan fitur itu.

## Struktur folder

```
backend/
  cmd/
    api/main.go          menjalankan HTTP server
    migrate/main.go      menjalankan AutoMigrate
  internal/
    auth/                satu slice
      entity.go          model GORM milik slice ini
      dto.go             request dan response
      repository.go      akses database
      service.go         aturan bisnis
      controller.go      handler HTTP
      module.go          perakitan dan pendaftaran route
    shared/              dipakai lintas slice
      config/
      cookie/
      database/
      apperror/
      httpresponse/
      middleware/
      security/
      token/
```

## Aturan slice

1. Satu folder di `internal` sama dengan satu fitur. Nama folder memakai kata
   benda tunggal dalam huruf kecil, contoh `auth`, `project`, `procurement`.
2. Setiap slice wajib punya pemisahan lapisan controller, service, repository.
   Boleh ada file tambahan selama masih milik fitur itu.
3. Slice tidak boleh mengimpor slice lain secara langsung. Kalau dua slice butuh
   hal yang sama, hal itu naik ke `shared`. Kalau yang dibutuhkan adalah data
   milik slice lain, minta lewat interface yang dideklarasikan di slice pemakai.
4. Setiap slice mengekspor `Entities()` yang mengembalikan daftar model GORM
   miliknya. `cmd/migrate` hanya mengumpulkan daftar itu, jadi menambah slice
   baru tidak mengubah isi file migrasi.
5. Perakitan dependency ada di `module.go` milik slice, bukan di `main.go`.
   `main.go` cukup memanggil `NewModule(...)` lalu `RegisterRoutes(...)`.

## Aturan lapisan

| Lapisan | Boleh tahu soal | Dilarang |
|---|---|---|
| Controller | Gin, DTO, service | Query database, GORM |
| Service | Repository interface, DTO, entity, paket shared | Gin, `*gin.Context`, HTTP status |
| Repository | GORM, entity | DTO, Gin, aturan bisnis |

Arah ketergantungan selalu satu arah: controller memanggil service, service
memanggil repository. Tidak ada yang memanggil balik ke atas.

Repository dideklarasikan sebagai interface di slice-nya, implementasinya
memakai GORM. Service menerima interface itu, bukan `*gorm.DB`, supaya service
bisa diuji tanpa database.

## Aturan folder shared

Semua yang dipakai lebih dari satu slice hanya boleh tinggal di `internal/shared`.
Dilarang membuat folder `utils`, `helpers`, `common`, atau `pkg` di tempat lain.

| Paket | Isi |
|---|---|
| `config` | Pembacaan environment variable dan nilai bawaan |
| `apperror` | Custom exception. Satu tipe `Error` yang membawa status HTTP, kode mesin, dan pesan untuk pengguna |
| `httpresponse` | Bentuk baku body response sukses dan gagal |
| `middleware` | CORS, penanganan error, autentikasi |
| `database` | Koneksi GORM dan setelan connection pool |
| `token` | Pembuatan dan pembacaan access token serta refresh token |
| `security` | Hashing kata sandi |
| `cookie` | Penulisan dan penghapusan cookie refresh token |

Isi `shared` harus bebas dari aturan bisnis. Kalau sebuah fungsi hanya masuk akal
untuk satu fitur, fungsi itu milik slice, bukan `shared`.

## Penanganan error

Service mengembalikan `*apperror.Error` untuk kesalahan yang boleh dilihat
pengguna, dan membungkus kesalahan tak terduga dengan `apperror.Internal(err)`.
Controller tidak menulis status HTTP sendiri, cukup `ctx.Error(err)` lalu
middleware `ErrorHandler` yang menerjemahkan menjadi response.

Pesan untuk pengguna ditulis dalam bahasa Indonesia. Kode error ditulis dalam
snake_case bahasa Inggris supaya stabil dipakai frontend, contoh
`invalid_credential`.

Detail kesalahan teknis tidak pernah dikirim ke client. Error dengan status 500
dicatat lewat `slog` lalu dibalas dengan pesan umum.

## Entry point

`cmd/api` hanya merakit: baca config, buka database, bangun router, jalankan
server, lalu matikan dengan rapi saat menerima sinyal berhenti.

`cmd/migrate` hanya menjalankan AutoMigrate atas daftar entity dari semua slice.
Dijalankan terpisah dari server supaya migrasi bisa dijalankan sendiri di CI
atau sebelum deploy.
