# Scoring service

Layanan Python (FastAPI) yang menghitung skor kelayakan lokasi dari data GIS
nyata. Dipanggil oleh slice `site` di backend Go lewat `POST /score`; tidak
diakses langsung oleh frontend.

## Kontrak `POST /score`

```text
in:  latitude, longitude, building_profile_code
out: overall_score, dimension_scores[{dimension_code, value, explanation}], risk_flags[{code, severity, message}],
     region   (nama kab/kota GADM NAME_2, mis. "Kota Bandung"; null kalau titik di luar wilayah GADM)
```

`region` hanya informasi deskriptif: backend memakainya sebagai kata kunci berita
kalau data RDTR tidak punya kecamatan. Nilainya tidak pernah ikut dihitung dalam skor.

## Data bundle

Layanan membaca **bundle data ringkas** (267 MB, zip 239 MB), bukan 18 GB data mentah.

Cara cepat: unduh `scoring-bundle.zip` dari Google Drive "AOL SWE", ekstrak ke
`scoring/` sehingga isinya ada di `scoring/bundle/` (folder hasil ekstrak bernama
`scoring-bundle`, ganti namanya jadi `bundle`). Bundle tidak di-commit dan tidak
diunggah ke GitHub: isinya turunan GADM (lisensi melarang redistribusi tanpa izin)
dan data ZNT ATR/BPN, sedangkan repo ini publik.

Membangun ulang dari data mentah (`data/raw` repo BanguninAja). Langkah `roads`
(±20 menit) dan `land` (±40 menit) paling lama; jalankan di proses terpisah dengan
`--only` supaya paralel:

```bash
DATA_DIR="/path/ke/BanguninAja/data/raw" .venv/Scripts/python build_bundle.py --out bundle
.venv/Scripts/python build_bundle.py --out bundle --only roads land   # sebagian langkah saja
```

| Berkas bundle | Dari data mentah | Dipakai untuk |
|---|---|---|
| `hazard_{multi,banjir,gempabumi,longsor,tsunami,likuefaksi,gunungapi}.tif` | InaRISK WCS 250 m, disimpan uint8 (indeks × 250) | Dimensi fisik + risk flag |
| `dem.tif` | Copernicus DEM | Kemiringan lahan |
| `population.tif` | WorldPop 90 m dijumlahkan 3×3 (~275 m, total penduduk tetap) | Dimensi demografi |
| `road_distance.tif` | Jaringan jalan (2 GB) → jarak ke jalan terdekat, grid 0,001° (~111 m), maks 7 km | Infrastruktur |
| `land_price.tif` | ZNT (7,6 GB) → Rp/m² per sel 0,001°, 0 = tanpa zona | Finansial |
| `poi.gpkg`, `transit.gpkg` | POI OSM, simpul transit | Pesaing, fasilitas harian, RS, transit |
| `regions.gpkg`, `ikk.csv` | GADM kab/kota (disederhanakan), IKK BPS | Indeks kemahalan konstruksi, nama wilayah |

Tidak ada lagi query live ke server ATR/BPN saat runtime.

## Menjalankan

Lokal:

```bash
cd scoring
python -m venv .venv
.venv/Scripts/pip install -r requirements.txt        # Linux/macOS: .venv/bin/pip
BUNDLE_DIR=bundle .venv/Scripts/uvicorn banguninaja_scoring.app:app --port 8090
```

Docker (dari root repo): `docker compose --profile scoring up -d scoring` — bundle
dibaca dari `scoring/bundle` (atau `SCORING_BUNDLE_DIR` di `.env`).

Lalu di `.env` backend:

```
SCORE_SERVICE_URL=http://scoring:8090              # kalau scoring jalan via compose
SCORE_SERVICE_URL=http://host.docker.internal:8090 # kalau scoring jalan di host
SCORE_SERVICE_TIMEOUT=15s
```

Kalau `SCORE_SERVICE_URL` kosong atau layanan mati, backend diam-diam memakai skor
stub — cek penjelasan dimensi: stub selalu diakhiri "(data simulasi)", skor asli tidak.

## Cara menghitung

1. **Fitur mentah** per titik: indeks bahaya, kemiringan, jarak ke jalan/transit/RS,
   penduduk radius 2 km, jumlah pesaing sejenis radius 3 km, fasilitas harian radius
   2 km, harga tanah ZNT, IKK.
2. **Skor dimensi 0–100** = persentil fitur terhadap 1.489 titik sampel se-Indonesia
   (3 titik acak per kab/kota), arahnya disesuaikan (bahaya/jarak/harga makin rendah
   makin baik). Data kosong di satu titik dianggap netral (50).
3. **Skor total** = jumlah berbobot 5 dimensi. Regulasi sengaja tidak ikut — itu
   batasan ya/tidak, ditampilkan terpisah di backend.
4. **Bobot dipelajari** (`train.py`) per profil bangunan dengan regresi linear
   koefisien non-negatif, target `log1p(penduduk_5km / (fasilitas_sejenis_5km + 1))`
   — proksi "permintaan belum terlayani". Untuk perumahan, "fasilitas sejenis" =
   sekolah (proksi layanan permukiman, karena tidak ada data perumahan eksisting).

### Catatan jujur soal bobot

Target berbagi komponen penduduk dengan dimensi demografi, jadi R² tinggi
(0,73–0,94) sebagian bersifat mekanis dan regresi murni memberi bahaya bencana
serta infrastruktur bobot 0. Karena itu bobot akhir = 10% tetap per dimensi + 50%
dibagi menurut koefisien hasil belajar (penyusutan ke bobot rata). Bahaya tinggi
tetap selalu muncul sebagai risk flag terpisah, apa pun bobotnya.

## Melatih ulang

```bash
BUNDLE_DIR=bundle .venv/Scripts/python train.py         # sampel baru + ekstraksi dari bundle
.venv/Scripts/python train.py --reuse-points            # pakai ulang model/training_points.csv
```

Menulis `model/model.json` dan menyalin bobot ke `backend/cmd/seed/data/scoring.json`;
jalankan ulang `go run ./cmd/seed` supaya tabel `weight` ikut terbarui. Training
dan layanan memakai bundle yang sama, jadi persentil di model konsisten dengan
nilai yang dibaca saat runtime.

## Test

```bash
.venv/Scripts/python -m pytest
```
