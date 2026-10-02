# PROJECT HANDOFF — BanguninAja

**Last updated:** 2026-10-02 · **Canonical copy:** this file, in `BanguninAjaApp`. The copy in the `BanguninAja` data repo is outdated — ignore it.

> Written to be handed to a teammate **and their own AI agent**. It states decisions and current state; the repo plus this doc should be enough for an agent to derive concrete tasks.

---

## TL;DR

- **The core feature is now real.** `POST /api/site/evaluate` scores a site from actual GIS data (InaRISK hazards, DEM, WorldPop, roads/transit, POI, ZNT land value, construction cost index) through the Python service in `scoring/`. The stub only kicks in if that service isn't running.
- **A fresh database is usable**: `go run ./cmd/seed` loads regions, units, a chart of accounts, and the scoring tables.
- **`/lokasi` has a map**: click to pick a point, see saved sites as markers, link each evaluation to a project.
- Verified end to end on 2026-10-02 (register → login → create project → evaluate on the map → score + risk flags + regulation, saved under the project).
- Left: merge PR #10 (news outside RDTR cities) after a browser check, payment simulation, polish (§5).

---

## 1. Current State (`main`, 2026-10-02)

### Backend (Go / Gin / GORM / PostGIS)
| Area | Status |
|---|---|
| Auth | ✅ |
| 13 business slices (CRUD, services, tests) | ✅ |
| `regulation` + `cmd/import-rdtr` | ✅ real RDTR for 11 imported cities, simulated (`is_simulated`) elsewhere |
| `news` (Google News RSS) | ✅ but only fetched when the regulation lookup returns a district — see §5 |
| `site` → `POST /api/site/evaluate` | ✅ predictive + descriptive in one response; saves `project_id` |
| `cmd/seed` | ✅ 531 regions, 14 units, 26 accounts, 5 dimensions, 4 profiles, trained weights |
| `saved_location.project_id`, `GET /api/locations/saved?projectId=` | ✅ |
| On-chain payment simulation | ❌ not started |

### Scoring service (`scoring/`, Python / FastAPI)
✅ `POST /score` from real data, trained on 1,489 sites. See `scoring/README.md` for how it works, how to run it, and the honest caveat about the learned weights.

### Frontend (React / Vite / Tailwind / MapLibre)
| Route | Page | Status |
|---|---|---|
| `/` | Overview (cross-project KPIs) | ✅ |
| `/proyek` | Projects list + create form | ✅ |
| `/keuangan` | Financial | ✅ |
| `/lokasi` | Site evaluation with map, project selector, score, risk flags, regulation, news | ✅ |

### Who built what (git authors)
- **Jonathan Andrew Saleh** — auth, DB schema, Docker, ERP frontend shell + Overview/Financial
- **nathanaelmosesES (Moses)** — service/controller layer for all 13 slices
- **AnthonyBudiarto** — regulation, news, `site/evaluate`, Location page *(which teammate owns this account is still unconfirmed)*
- **draxmit (Derick)** — seed data, scoring service + training, map + project link, RDTR importer fix

---

## 2. Data Status

5 of 6 layers are nationally complete, shared via Google Drive (`AOL SWE`). Regulasi/Zonasi is real for 11 RDTR cities and simulated elsewhere — its source (`gistaru.atrbpn.go.id`) has been unreachable since mid-September, so it is **never part of the score**, only shown as descriptive context.

Data preprocessing done for the seed: GADM coded Kalimantan Utara as 64 (Kaltim's code) → fixed to 65 with BPS codes for its 5 regencies; "Danau Limboto" (a lake listed as a regency) dropped. ZNT values above Rp500 jt/m² (the int32-overflow artifact) are ignored by the scoring service.

---

## 3. Scoring Design (frozen, implemented)

- **Predictive** = weighted sum of 5 dimensions (0–100 each, percentile against a national sample). Weights learned per building profile by non-negative regression on `log1p(pop_5km / (same_profile_facilities_5km + 1))`, then shrunk halfway toward equal weights (each dimension ≥ 10%) because that target shares population with the demography dimension. Current weights are in `backend/cmd/seed/data/scoring.json`.
- **Descriptive** = regulation + news, shown beside the score, never scored.
- Risk flags (high/medium) for flood, earthquake, landslide, tsunami, liquefaction, volcano, steep slope, and out-of-coverage points.

```text
POST /api/site/evaluate   (auth required)
  in:  project_id?, latitude, longitude, building_profile_id, name
  out: saved_location_id,
       predictive:  { overall_score, dimension_scores[{dimension_code, value, explanation}], risk_flags[{code, severity, message}] },
       descriptive: { regulasi: {zona, kdb, klb, is_simulated}, news: [{title, url, source, published_at}] }

POST /score   (internal, scoring/ service)
  in:  latitude, longitude, building_profile_code   (housing | hospital | mall | entertainment)
  out: overall_score, dimension_scores[], risk_flags[]
dimension_code ∈ fisik_lingkungan | infrastruktur | demografi_sosial | pasar_kompetisi | finansial_proyek
```

---

## 4. Running Everything Locally

```bash
cp .env.example .env                                    # set JWT_ACCESS_SECRET
docker compose up -d
docker compose exec backend go run ./cmd/migrate
docker compose exec backend go run ./cmd/seed
docker compose run --rm -v "/path/to/BanguninAja/data/raw/rdtr:/rdtr" backend go run ./cmd/import-rdtr -dir /rdtr   # optional
# scoring service (needs the GIS data, ~18 GB, from the Drive):
cd scoring && python -m venv .venv && .venv/Scripts/pip install -r requirements.txt
DATA_DIR=/path/to/BanguninAja/data/raw .venv/Scripts/uvicorn banguninaja_scoring.app:app --port 8090
# in .env: SCORE_SERVICE_URL=http://host.docker.internal:8090  SCORE_SERVICE_TIMEOUT=15s
cd frontend && npm install && npm run dev
```

Gotchas seen on Derick's laptop:
- A Windows PostgreSQL service on port 5432 collides with the compose database. Stop it or remap the compose port.
- Some networks cut Go's TLS 1.3 handshake (`go mod download` → `EOF` while `curl` works). Workaround used: a local module mirror (`GOPROXY=file://...`) built from `go.sum`.
- Another app on `[::1]:5173` hijacked `localhost:5173`; use `127.0.0.1` and add it to `CORS_ALLOWED_ORIGINS`.

---

## 5. What's Left

| # | Task | Why | Priority |
|---|---|---|---|
| B1 | `finance` budget query calls `make_date(bigint, …)` → every `GET /api/finance/budgets` returns 500; Overview's "Serapan anggaran" card always fails | Visible on the dashboard | ✅ fixed (#8) |
| B2 | Two concurrent `/auth/refresh` calls (React StrictMode, or parallel 401 retries) — the second presents a just-rotated token and logs the user out on reload | Users get bounced to the login page | ✅ fixed (#8) |
| N1 | News is only fetched when the regulation lookup returns a district, and simulated regulation has none — so the news panel is empty almost everywhere | Descriptive panel looks broken in the demo | PR #10 open — needs a browser check, then merge |
| N2 | `/proyek` has no "create project" form (projects can only be created via the API) | Demo needs a project to link sites to | ✅ done (#9) |
| N3 | On-chain payment simulation (billing: QR request → buyer pays from own wallet on testnet → verify via block-explorer API) | Independent, Derick's | P2 |
| N4 | Per-project workspace (Overview / Site / Financial as tabs inside one project) | Polish | P3 |
| N5 | The scoring service needs the ~18 GB dataset next to it, so it only runs on a machine that has the Drive data | Deployment beyond a laptop | P3 |

**Demo check**: on a fresh clone, migrate + seed → register → create a project → evaluate a point on the map → the explanations must **not** end with "(data simulasi)" (that means the stub answered, i.e. the scoring service isn't reachable).

---

## 6. Don't Touch

- `auth/*` and the login/session frontend (except bug B2).
- The `site` slice's stub fallback — it keeps the demo alive if the scoring service is down.
- `scoring/train.py`'s refusal to bulk-query the live ATR/BPN server.
- The RDTR scraper's throttling in `BanguninAja/scripts/ambil_rdtr.py`.

---

## 7. Open Questions

- **Who is `AnthonyBudiarto`?** Confirm which teammate owns this git account.
- New deadline (the 2026-09-27 target has passed).
- Sales / Procurement / HR / Inventory screens: backend exists, no frontend — cut or later?
