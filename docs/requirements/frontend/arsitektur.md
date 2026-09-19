# Arsitektur Frontend

React 19 dengan Vite dan TypeScript. Susunannya MVC sederhana: tiga folder, satu
arah ketergantungan, tanpa lapisan tambahan yang tidak dipakai.

## Struktur folder

```
frontend/src/
  models/          data dan cara mengambilnya
    auth.ts        bentuk DTO
    authApi.ts     pemanggilan endpoint
  controllers/     penghubung data dan tampilan
    sesiStore.ts   penyimpan sesi di memori
    useAuth.ts     hook TanStack Query
  views/           yang dilihat pengguna
    MasukPage.tsx
    DaftarPage.tsx
    DasborPage.tsx
    komponen/      potongan yang dipakai berulang
  shared/          dipakai lintas bagian
    apiClient.ts
    pesanKesalahan.ts
  App.tsx          rute dan provider
```

## Pembagian tugas

| Lapisan | Isinya | Dilarang |
|---|---|---|
| Model | Tipe DTO dan fungsi pemanggil API | JSX, state React |
| Controller | Hook query dan mutation, penyimpan sesi | JSX, kelas Tailwind |
| View | Komponen React dan tata letak | `fetch` langsung, olah data mentah |

View tidak pernah memanggil API sendiri. View memakai hook dari controller, dan
controller memakai fungsi dari model. Kalau sebuah komponen butuh data baru,
yang ditambah hooknya, bukan `fetch` di dalam komponen.

## DTO

Bentuk DTO mengikuti persis apa yang dikirim backend, termasuk nama fieldnya.
Tidak ada penerjemahan nama di tengah jalan, supaya kalau ada yang tidak cocok
langsung kelihatan dari TypeScript.

```ts
export interface Pengguna {
  id: string
  name: string
  email: string
  createdAt: string
}
```

Semua respons backend dibungkus amplop `{ data }` atau `{ error }`. Pembukaan
amplop itu dikerjakan sekali di `apiClient`, jadi model dan view hanya berurusan
dengan isinya.

## TanStack Query

Dipakai supaya tidak perlu menulis hook sendiri untuk memanggil API. Keadaan
memuat, gagal, dan berhasil sudah disediakan, begitu juga penyimpanan sementara
dan pengambilan ulang.

| Kebutuhan | Cara |
|---|---|
| Membaca data | `useQuery` |
| Mengubah data | `useMutation` |
| Membuang semua data setelah keluar | `queryClient.clear()` |

Aturannya: jangan menulis `useState` plus `useEffect` untuk memanggil API. Itu
yang mau dihindari.

## Penyimpanan sesi

Zustand dipakai hanya untuk sesi, bukan untuk semua state. Data dari server
tetap dipegang TanStack Query.

Store sengaja **tidak** memakai middleware `persist`. Kalau dipersist, tokennya
mendarat di localStorage, dan itu justru yang dilarang.

## Alur token

1. Login berhasil, backend mengirim access token di body dan refresh token di
   cookie HttpOnly.
2. Access token disimpan di memori lewat zustand. Begitu tab ditutup, hilang.
3. Setiap permintaan menyertakan `Authorization: Bearer ...` dan
   `credentials: 'include'`.
4. Kalau backend menjawab 401, `apiClient` memanggil `/auth/refresh` sekali,
   memperbarui token di memori, lalu mengulang permintaan tadi. Kalau refresh
   ikut gagal, sesi dihapus dan pengguna diarahkan ke halaman masuk.
5. Saat aplikasi pertama dibuka, `/auth/refresh` dipanggil sekali. Ini yang
   membuat pengguna tetap masuk setelah halaman dimuat ulang, padahal access
   tokennya hilang bersama memori.

Selama pemeriksaan itu berjalan, aplikasi menampilkan keterangan singkat, bukan
halaman masuk yang berkedip sebentar lalu hilang.

## Larangan penyimpanan

| Tempat | Boleh menyimpan token? |
|---|---|
| Memori aplikasi, zustand | Ya, untuk access token |
| Cookie HttpOnly | Ya, untuk refresh token, diatur backend |
| localStorage | **Tidak** |
| sessionStorage | **Tidak** |
| Cookie biasa yang bisa dibaca JavaScript | **Tidak** |

Alasannya: apa pun yang bisa dibaca JavaScript juga bisa dibaca script asing
yang menyusup ke halaman, misalnya lewat dependency yang disusupi. Cookie
HttpOnly tidak bisa dibaca JavaScript sama sekali.

Sudah diperiksa di browser setelah login: `localStorage` kosong,
`sessionStorage` kosong, dan `document.cookie` kosong.

## Alamat backend

Diambil dari `VITE_API_URL`, dengan nilai bawaan `http://localhost:8080/api`.
Backend mengizinkan origin `http://localhost:5173` berikut credentials, jadi
cookie ikut terkirim walaupun port keduanya berbeda.

## Styling

Tailwind CSS v4, dipasang lewat plugin Vite. Tidak ada berkas
`tailwind.config.js`, karena v4 mengatur token lewat blok `@theme` di
`src/index.css`.

Warna merek diambil dari logo: navy `#0b2b6b` dan amber `#fbbb16`.

Aturan kelas: tulis langsung di komponen. Kalau satu rangkaian kelas dipakai di
lebih dari dua tempat, jadikan komponen di `views/komponen`, bukan `@apply`.
