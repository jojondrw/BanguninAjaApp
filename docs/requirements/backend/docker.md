# Docker Backend

Ada dua Dockerfile dengan tujuan berbeda. Yang lama, `backend/Dockerfile`,
dihapus karena masih berisi contoh dari tutorial Docker dan menghasilkan binary
bernama `docker-gs-ping`.

| Berkas | Dipakai saat | Isi image |
|---|---|---|
| `Dockerfile.dev` | Ngoding di laptop | Toolchain Go lengkap plus air |
| `Dockerfile.prod` | Deploy | Hanya binary hasil build, tanpa compiler dan tanpa shell |

## Development

`Dockerfile.dev` memasang [air](https://github.com/air-verse/air). Air mengawasi
berkas `.go`, membangun ulang, lalu menyalakan ulang server. Jadi setiap kali
berkas disimpan, perubahannya langsung jalan tanpa perlu menghentikan container.

Folder `backend` dipasang sebagai bind mount ke `/app`, jadi berkas yang diedit
di Windows langsung terlihat oleh container.

### Kenapa polling dinyalakan

Berkasnya ada di Windows, containernya Linux. Notifikasi perubahan berkas dari
Windows tidak diteruskan ke dalam container, jadi air tidak pernah menerima
kabar apa pun dan diam saja meski berkas sudah diubah.

Karena itu `.air.toml` memakai:

```toml
poll = true
poll_interval = 500
```

Dengan setelan itu air memeriksa sendiri tiap setengah detik, bukan menunggu
dikabari. Sedikit lebih boros CPU, tapi ini satu-satunya cara yang bekerja dari
bind mount Windows.

Kalau suatu saat proyek ini digarap di Linux atau macOS, `poll` boleh dimatikan
supaya lebih hemat.

### Cache

Dua named volume dipakai supaya rebuild tidak lambat:

| Volume | Isi |
|---|---|
| `go_modules` | Dependency hasil `go mod download` |
| `go_build_cache` | Hasil kompilasi paket yang tidak berubah |

Tanpa keduanya, setiap restart container mengunduh ulang seluruh dependency.

### Menjalankan

```bash
docker compose up -d
docker compose logs -f backend
```

Migrasi dijalankan lewat container yang sama:

```bash
docker compose exec backend go run ./cmd/migrate
```

Backend menunggu database benar-benar siap, bukan sekadar hidup. Itu diatur oleh
`healthcheck` pada service `database` dan `condition: service_healthy` pada
service `backend`. Tanpa itu backend sering gagal di detik pertama karena
Postgres masih menyiapkan diri.

## Production

`Dockerfile.prod` memakai multi stage build:

1. Stage `build` memakai image `golang` lengkap untuk mengompilasi.
2. Stage `runtime` memakai `gcr.io/distroless/static-debian12:nonroot` dan hanya
   menerima hasil binary dari stage sebelumnya.

Image akhir tidak berisi compiler Go, tidak berisi package manager, bahkan tidak
berisi shell. Kalau ada penyerang berhasil masuk, tidak ada alat apa pun di
dalam yang bisa dipakai.

Binary dibangun dengan:

| Flag | Gunanya |
|---|---|
| `CGO_ENABLED=0` | Binary berdiri sendiri, tidak bergantung pustaka sistem |
| `-trimpath` | Path folder di laptop tidak ikut tertanam di binary |
| `-ldflags="-s -w"` | Membuang tabel debug, ukuran binary mengecil |

Container berjalan sebagai user `nonroot`, bukan root.

Dua binary ikut disalin: `/api` untuk server dan `/migrate` untuk migrasi. Jadi
image yang sama bisa dipakai menjalankan migrasi sebelum deploy.

```bash
docker build -f backend/Dockerfile.prod -t banguninaja-backend:latest backend
```

## .dockerignore

`backend/.dockerignore` mencegah `tmp` dan berkas `.env` ikut masuk image.
Berkas `.env` berisi kunci rahasia, dan tidak boleh ikut ke image yang nanti
dikirim ke server.
