# Scoring service

Layanan Python (FastAPI) yang menghitung skor kelayakan lokasi dari data GIS
nyata. Dipanggil oleh slice `site` di backend Go lewat `POST /score`; tidak
diakses langsung oleh frontend.

## Menjalankan

```bash
cd scoring
python -m venv .venv
.venv/Scripts/pip install -r requirements.txt        # Linux/macOS: .venv/bin/pip
DATA_DIR="/path/ke/BanguninAja/data/raw" .venv/Scripts/uvicorn banguninaja_scoring.app:app --port 8090
```

Lalu di `.env` backend:

```
SCORE_SERVICE_URL=http://host.docker.internal:8090
SCORE_SERVICE_TIMEOUT=15s
```

`host.docker.internal` dipakai karena backend jalan di Docker sedangkan layanan
ini jalan di host. Kalau `SCORE_SERVICE_URL` kosong atau layanan mati, backend
diam-diam memakai skor stub — cek penjelasan dimensi: stub selalu diakhiri
"(data simulasi)", skor asli tidak.

`DATA_DIR` adalah folder `data/raw` repo data BanguninAja (atau Google Drive
"AOL SWE"). Berkas yang dibaca:

| Berkas | Dipakai untuk |
|---|---|
| `inarisk_wcs/inarisk_{multi,banjir,gempabumi,longsor,tsunami,likuefaksi,gunungapi}_indonesia.tif` | Dimensi fisik + risk flag |
| `dem_copernicus30m_indonesia_2026.tif` | Kemiringan lahan |
| `worldpop_idn_2020.tif` | Dimensi demografi |
| `jaringan_transportasi_indonesia_2026.gpkg` | Jarak jalan & simpul transit |
| `poi_kompetitor_indonesia_2026.gpkg` | Pesaing sejenis, fasilitas harian, rumah sakit |
| `znt_atrbpn_indonesia.gpkg` | Harga tanah (kalau tidak ada: query live ke ATR/BPN per permintaan) |
| `gadm41_indonesia.gpkg`, `bps_ikk_indonesia_2025.csv` | Indeks kemahalan konstruksi per kab/kota |

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
DATA_DIR=... .venv/Scripts/python train.py              # sampel baru + ekstraksi (~4 menit)
.venv/Scripts/python train.py --reuse-points            # pakai ulang model/training_points.csv
```

Menulis `model/model.json` dan menyalin bobot ke `backend/cmd/seed/data/scoring.json`;
jalankan ulang `go run ./cmd/seed` supaya tabel `weight` ikut terbarui. Training
tidak pernah memakai query ZNT live (dulu scraping massal pernah menjatuhkan
server ATR/BPN) — tanpa berkas ZNT lokal, harga tanah dilewati.

## Test

```bash
.venv/Scripts/python -m pytest
```
