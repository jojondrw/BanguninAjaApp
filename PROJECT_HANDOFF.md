# PROJECT HANDOFF — BanguninAja

**Last updated:** 2026-10-03 · **Canonical copy:** this file, in `BanguninAjaApp`. The copy in the `BanguninAja` data repo is outdated — ignore it.

> Written to be handed to a teammate **and their own AI agent**. It states decisions and current state; the repo plus this doc should be enough for an agent to derive concrete tasks.

---

## TL;DR

- **Site intelligence is real end to end.** `POST /api/site/evaluate` scores a point from actual GIS data through the Python service in `scoring/`. The data is a 267 MB bundle (not the 18 GB raw data), so the service can be hosted.
- **Every ERP module has a screen**: projects (workspace with phases, permits, RAB, status), finance (budgets, cash, journal, ledger), billing (invoices, receivables, payables), sales (units, leads, customers, contracts, installments), procurement (requests, POs, goods receipts, vendors), inventory, HR.
- **GIS extras**: compare 2–4 candidate sites side by side; map layers for hazard (multi, flood, earthquake, landslide), population density, and ZNT land price, rendered from the same bundle.
- **UI** refreshed in an Apple-like style (SF Pro / Inter, segmented controls, command palette Ctrl/⌘ K). See `docs/requirements/frontend/aturan-ui.md`.
- **CI** runs Go, React, and Python checks on every PR.
- Left: on-chain payment simulation (Derick's), plus the backend gaps in §5.

---

## 1. Current State (`main`, 2026-10-03)

### Backend (Go / Gin / GORM / PostGIS)
| Area | Status |
|---|---|
| Auth, 13 business slices, `regulation`, `news` | ✅ |
| `site` → `POST /api/site/evaluate` | ✅ predictive + descriptive; stores score, dimension scores, flood/earthquake index, ZNT price and the real risk flags (`saved_location.risk_flags`, jsonb) |
| `site` → `GET /api/site/tiles/:layer/:z/:x/:y` | ✅ auth-protected proxy to the scoring service's map tiles |
| `cmd/seed`, `cmd/migrate`, `cmd/import-rdtr` | ✅ |
| Row locking on update/delete (billing, project) | ✅ fixed lost-update races found this round |
| On-chain payment simulation | ❌ not started (Derick) |

### Scoring service (`scoring/`, Python / FastAPI)
- `POST /score` → overall score, 5 dimension scores, risk flags, region, and `facts` (flood index, earthquake index, ZNT price).
- `GET /tiles/{bahaya|banjir|gempabumi|longsor|penduduk|harga_tanah}/{z}/{x}/{y}.png` → XYZ tiles from the bundle (min zoom 7, land price 10).
- Reads the **data bundle** (`BUNDLE_DIR`), built once by `build_bundle.py`. Trained on 1,489 sites. See `scoring/README.md`.

### Frontend (React / Vite / Tailwind / MapLibre)
| Route | Page |
|---|---|
| `/` | Overview: KPIs, 12-month cash-flow chart, "Perlu perhatian" (overdue installments, pending requests, low stock, unpaid payroll), projects |
| `/proyek`, `/proyek/:id` | Projects list + workspace (status transitions, edit/delete, phases, permits, RAB, site analysis, finance tabs) |
| `/lokasi` | Evaluasi (map + layers + score + regulation + news) and Bandingkan (compare 2–4 saved sites, saved comparisons) |
| `/keuangan` | Ringkasan, Anggaran, Kas, Jurnal, Buku besar |
| `/tagihan` | Faktur, Piutang, Utang with payment recording |
| `/penjualan` | Unit, Prospek, Pelanggan, Kontrak (+ installment payments) |
| `/pengadaan` | Pesanan pembelian, Penerimaan barang, Vendor, Permintaan pembelian |
| `/inventaris` | Stok dan mutasi, Material dan gudang, Aset dan alat (full CRUD) |
| `/sdm` | Employees, attendance, payroll (edit/delete, SQL payroll summary) |

Pages are lazy-loaded per route; a route-level error boundary keeps one broken page from blanking the app.

### Who built what (git authors)
- **Jonathan Andrew Saleh** — auth, DB schema, Docker, ERP frontend shell + Overview/Financial (first version)
- **nathanaelmosesES (Moses)** — service/controller layer for all 13 slices
- **AnthonyBudiarto (Ian)** — regulation, news, `site/evaluate`, Location page
- **draxmit (Derick)** — seed data, scoring service + training + data bundle, map + layers, project workspace, all ERP screens added in October, UI refresh, CI

---

## 2. Data Status

5 of 6 layers are nationally complete. Raw data is on Google Drive (`AOL SWE`); the scoring service only needs **`scoring-bundle.zip`** (239 MB) from the same Drive — not committed to GitHub because the repo is public and the bundle contains GADM-derived boundaries (licence forbids redistribution) and ATR/BPN ZNT data.

Regulasi/Zonasi is real for 11 RDTR cities and simulated elsewhere (source unreachable since mid-September), so it is **never scored**, only shown as context.

Preprocessing: Kalimantan Utara recoded to 65 with BPS regency codes; "Danau Limboto" dropped; ZNT values above Rp500 jt/m² (int32 overflow artifact) ignored.

---

## 3. Scoring Design (frozen, implemented)

- **Predictive** = weighted sum of 5 dimensions (0–100 each, percentile against a national sample). Weights learned per building profile by non-negative regression on `log1p(pop_5km / (same_profile_facilities_5km + 1))`, shrunk halfway toward equal weights (each dimension ≥ 10%). Weights are in `backend/cmd/seed/data/scoring.json`.
- **Descriptive** = regulation + news, beside the score, never scored.
- Risk flags (high/medium): flood, earthquake, landslide, tsunami, liquefaction, volcano, steep slope, out of coverage.

```text
POST /api/site/evaluate   (auth)
  in:  project_id?, latitude, longitude, building_profile_id, name
  out: saved_location_id, predictive {overall_score, dimension_scores[], risk_flags[]}, descriptive {regulasi, news[]}

POST /score   (internal scoring service)
  in:  latitude, longitude, building_profile_code   (housing | hospital | mall | entertainment)
  out: overall_score, dimension_scores[], risk_flags[], region, facts {flood_index, earthquake_index, land_price_per_sqm}
```

---

## 4. Running Everything Locally

```bash
cp .env.example .env                                    # set JWT_ACCESS_SECRET
docker compose up -d
docker compose exec backend go run ./cmd/migrate        # run again after every pull
docker compose exec backend go run ./cmd/seed
# scoring: unzip scoring-bundle.zip (Drive) into scoring/bundle, then either
docker compose --profile scoring up -d scoring          # SCORE_SERVICE_URL=http://scoring:8090
# or locally: cd scoring && python -m venv .venv && .venv/Scripts/pip install -r requirements.txt
#             BUNDLE_DIR=bundle .venv/Scripts/uvicorn banguninaja_scoring.app:app --port 8090
# in .env: SCORE_SERVICE_TIMEOUT=15s
cd frontend && npm install && npm run dev
```

Gotchas seen on Derick's laptop:
- A Windows PostgreSQL service on 5432 collides with the compose database — stop it or remap the port.
- Some networks cut Go's TLS 1.3 handshake (`go mod download` → `EOF`). Workaround: a local module mirror (`GOPROXY=file://...`).
- Another app on `[::1]:5173` hijacked `localhost:5173`; use `127.0.0.1` and add it to `CORS_ALLOWED_ORIGINS`.

**Demo check**: fresh clone → migrate + seed → register → create a project → evaluate a point on the map → the explanations must **not** end with "(data simulasi)" (that means the stub answered).

---

## 5. What's Left

| # | Task | Notes | Priority |
|---|---|---|---|
| N3 | On-chain payment simulation | Derick's. Plugs into `/tagihan` (billing). Billing has no per-payment history table yet (only a running paid total) — the simulation will likely need one | P1 |
| G1 | Receiving goods doesn't move stock | `/pengadaan` offers a manual "Catat stok masuk" per receipt; automatic posting would be a backend change in procurement | P3 |
| G2 | Cash transactions don't post journal entries | Cash and the ledger are separate books; the UI says so | P3 |
| G3 | Receivables/payables have no project, contract, or PO link; an invoice and a receivable can record the same debt twice | Billing data model | P3 |
| G4 | Leads aren't linked to the customer created from them | Sales data model | P3 |
| G5 | Dropdowns (materials, warehouses, parent assets, active employees) and some name lookups in procurement/billing tables still load at most 100 records | Stock tables already get names from the backend (#34); the rest needs searchable selects or joins | P3 |

---

## 6. Don't Touch

- `auth/*` and the login/session frontend.
- The `site` slice's stub fallback — it keeps the demo alive if the scoring service is down.
- `scoring/train.py`'s refusal to bulk-query the live ATR/BPN server.
- The RDTR scraper's throttling in `BanguninAja/scripts/ambil_rdtr.py`.

---

## 7. Open Questions

- New deadline (the 2026-09-27 target has passed).
- Has `scoring-bundle.zip` been uploaded to the Drive for the team?
