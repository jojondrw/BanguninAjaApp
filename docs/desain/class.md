# Class Diagram

Go tidak punya kelas, tetapi `struct` dengan method dan `interface` berperan
sama, jadi digambar dengan notasi class diagram. Diagram di sini menunjukkan
**struktur kode**, bukan struktur data. Struktur data ada di [ERD](erd.md).

## 1. Pola vertical slice di backend

Setiap fitur backend adalah satu paket di `backend/internal/<slice>` dengan
lapisan yang sama. Contoh di bawah adalah slice `site` (evaluasi lokasi), karena
slice ini memakai semua lapisan sekaligus: repository, layanan luar, dan dua
slice lain.

```mermaid
classDiagram
    direction LR

    class Module {
        -controller *Controller
        -tiles *TileProxy
        +NewModule(db, tokens, scoreConfig) *Module
        +RegisterRoutes(router)
    }
    class Controller {
        -service Service
        +Evaluate(ctx *gin.Context)
    }
    class Service {
        <<interface>>
        +Evaluate(ctx, userID, EvaluateRequest) EvaluateResponse
    }
    class service {
        -repository Repository
        -scores ScoreClient
        -regulationService regulation.Service
        -newsService news.Service
        +Evaluate(ctx, userID, EvaluateRequest) EvaluateResponse
        -persist(...) SavedRecord
        -describe(...) Descriptive
        -fetchNews(...) News[]
    }
    class Repository {
        <<interface>>
        +Transaction(ctx, work) error
        +FindBuildingProfile(ctx, id) BuildingProfile
        +DimensionIDByCode(ctx) map
        +ProjectExists(ctx, id) bool
        +CreateSavedLocation(ctx, saved) error
    }
    class gormRepository {
        -db *gorm.DB
    }
    class ScoreClient {
        <<interface>>
        +Score(ctx, ScoreInput) ScoreResult
    }
    class httpScoreClient {
        -baseURL string
        -client *http.Client
        -stub stubScoreClient
    }
    class stubScoreClient {
        +Score(ctx, ScoreInput) ScoreResult
    }
    class EvaluateRequest {
        +ProjectID *uuid
        +Latitude *float64
        +Longitude *float64
        +BuildingProfileID uuid
        +Name string
    }
    class EvaluateResponse {
        +SavedLocationID uuid
        +Predictive Predictive
        +Descriptive *Descriptive
    }

    Module *-- Controller
    Controller --> Service
    service ..|> Service
    service --> Repository
    service --> ScoreClient
    gormRepository ..|> Repository
    httpScoreClient ..|> ScoreClient
    stubScoreClient ..|> ScoreClient
    httpScoreClient *-- stubScoreClient : cadangan
    Controller ..> EvaluateRequest : bind JSON
    Controller ..> EvaluateResponse : kirim JSON
```

| Lapisan | Tanggung jawab | Tidak boleh |
|---|---|---|
| `module.go` | Merakit dependensi dan mendaftarkan rute | Berisi logika bisnis |
| `controller.go` | Membaca request, validasi bentuk (`binding`), memanggil service, mengirim respons | Menyentuh database |
| `service.go` | Aturan bisnis: transisi status, perhitungan, transaksi | Tahu soal HTTP |
| `repository.go` | Query GORM dan SQL PostGIS | Membuat keputusan bisnis |
| `dto.go` | Bentuk JSON masuk dan keluar | Dipakai sebagai tabel |
| `entity.go` | Bentuk tabel dan migrasi | Dikirim langsung ke klien |

Karena `Service`, `Repository`, dan `ScoreClient` berupa interface, unit test
bisa mengganti implementasinya dengan versi palsu. Test repository memakai
driver SQL palsu (`shared/databasetest`) yang merekam query dan mengembalikan
hasil buatan, sehingga ke-70 berkas test Go berjalan tanpa database atau
layanan luar.

## 2. Infrastruktur bersama

Paket di `backend/internal/shared` dipakai semua slice.

```mermaid
classDiagram
    direction LR

    class Error {
        +Status int
        +Code string
        +Message string
        -cause error
        +Error() string
    }
    class ErrorMap {
        +NotFound *Error
        +Duplicate *Error
        +Referenced *Error
        +Invalid *Error
        +Resolve(err) error
    }
    class Manager {
        +IssueAccess(userID) Access
        +ParseAccess(value) uuid
        +IssueRefresh() Refresh
    }
    class RefreshWriter {
        +Set(ctx, value, expiresAt)
        +Read(ctx) string
        +Clear(ctx)
    }
    class Role {
        +Code string
        +Name string
        +Unrestricted bool
        +Grants map~Module, Level~
    }
    class Level {
        <<enumeration>>
        LevelNone
        LevelRead
        LevelWrite
    }
    class SubjectSource {
        <<interface>>
        +Subject(ctx, userID) Subject
    }
    class Authorization {
        <<middleware>>
        +ModuleForPath(path) Module
        +LevelForMethod(method) Level
        +Allows(role, module, level) bool
    }
    class Page~T~ {
        +Items T[]
        +Page int
        +PageSize int
        +TotalItems int64
        +TotalPages int
    }

    ErrorMap --> Error : memetakan galat GORM
    Authorization --> Manager : verifikasi JWT
    Authorization --> SubjectSource : peran dan status aktif
    Authorization --> Role : cek hak
    Role --> Level
```

| Paket | Isi |
|---|---|
| `apperror` | Satu tipe galat dengan status HTTP, kode mesin, dan pesan berbahasa Indonesia. Semua respons galat berbentuk sama |
| `database` | `ErrorMap` mengubah galat Postgres (data tidak ada, FK dilanggar, unik dilanggar) menjadi `apperror` milik slice |
| `token` | Menerbitkan dan memverifikasi JWT, membuat dan meng-hash refresh token |
| `cookie` | Menulis refresh token sebagai cookie `HttpOnly` |
| `access` | Matriks peran dan modul, satu sumber untuk backend dan nama modul di frontend |
| `middleware` | Autentikasi, otorisasi, CORS, dan pengubah galat menjadi JSON |
| `pagination` | Struktur halaman yang sama untuk semua daftar |

## 3. Frontend

Frontend memakai pembagian model, controller, dan view (lihat
`docs/requirements/frontend/arsitektur.md`). Contoh di bawah adalah alur evaluasi
lokasi dari halaman sampai request HTTP.

```mermaid
classDiagram
    direction LR

    class LocationPage {
        <<view>>
        SiteMap (Leaflet)
        panel skor dan konteks
    }
    class useSiteEvaluate {
        <<controller>>
        useMutation(siteApi.evaluate)
        invalidate daftar lokasi
    }
    class siteApi {
        <<model>>
        +evaluate(SiteEvaluateRequest) SiteEvaluateResponse
    }
    class apiClient {
        <<shared>>
        +request~T~(path, options) T
        +refreshSession() bool
        -pendingRefresh Promise
    }
    class useSessionStore {
        <<controller>>
        accessToken string
        user User
        +setSession(token, user)
        +clearSession()
    }
    class useCan {
        <<controller>>
        +canAccess(module, level) bool
    }

    LocationPage --> useSiteEvaluate
    LocationPage --> useCan : sembunyikan tombol
    useSiteEvaluate --> siteApi
    siteApi --> apiClient
    apiClient --> useSessionStore : baca token
```

| Folder | Isi | Contoh |
|---|---|---|
| `src/models` | Tipe data dan fungsi pemanggil API (`*Api.ts`) | `salesApi.ts`, `onchainApi.ts` |
| `src/controllers` | Hook TanStack Query dan store sesi | `useSales.ts`, `sessionStore.ts`, `useAccess.ts` |
| `src/views` | Halaman dan komponen | `SalesPage.tsx`, `views/components/Data.tsx` |
| `src/shared` | Klien HTTP, format angka dan tanggal, pesan galat | `apiClient.ts`, `format.ts` |

## 4. Layanan penilaian (Python)

```mermaid
classDiagram
    direction LR

    class App {
        <<FastAPI>>
        +GET /health
        +POST /score
        +GET /tiles/layer/z/x/y.png
    }
    class FeatureExtractor {
        +extract(lon, lat) dict
        -_physical(lon, lat) dict
        -_access(lon, lat) dict
        -_people_and_market(lon, lat) dict
        -_finance(lon, lat) dict
    }
    class Model {
        +quantiles dict
        +weights dict
        +meta dict
        +load(path) Model
        +percentile(name, value) float
    }
    class Raster {
        +value_at(lon, lat)
        +mean_and_max(lon, lat)
    }
    class Population {
        +sum_within(lon, lat, radius_m)
    }
    class Elevation {
        +slope_deg(lon, lat)
    }
    class LandPrice {
        +price_at(lon, lat)
    }
    class ConstructionCost {
        IKK per kabupaten/kota
    }
    class PointIndex {
        +count_within(lon, lat, radius_m) int
        +nearest_m(lon, lat) float
    }
    class score {
        <<function>>
        score(model, features, profile) dict
    }

    App --> FeatureExtractor
    App --> score
    score --> Model
    Population --|> Raster
    Elevation --|> Raster
    LandPrice --|> Raster
    FeatureExtractor --> Raster : InaRISK, jarak jalan
    FeatureExtractor --> Population : WorldPop
    FeatureExtractor --> Elevation : DEM
    FeatureExtractor --> LandPrice : ZNT
    FeatureExtractor --> PointIndex : POI, transit
    FeatureExtractor --> ConstructionCost
```

`FeatureExtractor` membaca bundle data GIS sekali saat layanan dinyalakan.
`Model` berisi kuantil tiap fitur (untuk mengubah angka mentah menjadi skor 0
sampai 100) dan bobot per profil hasil `train.py`. Bobot yang sama di-seed ke
tabel `weight` supaya terlihat di aplikasi.
