# PROJECT HANDOFF — BanguninAja

**Last updated:** 2026-10-17 (P1 construction features, security, docs) · **Canonical copy:** this file, in `BanguninAjaApp`. The copy in the `BanguninAja` data repo is outdated — ignore it.

> Written to be handed to a teammate **and their own AI agent**. It states decisions and current state; the repo plus this doc should be enough for an agent to derive concrete tasks.

---

## TL;DR

- **Site intelligence is real end to end.** `POST /api/site/evaluate` scores a point from actual GIS data through the Python service in `scoring/`. The data is a 267 MB bundle, 239 MB zipped (not the 18 GB raw data), so the service can be hosted.
- **Every ERP module has a screen and the modules are linked** (G1–G6 in §5): projects, finance, billing, sales, procurement, inventory, HR, reports.
- **Since 2026-10-15:** role-based access for 7 roles on every module (#44, #51), report export to PDF/XLSX/CSV (#45), realistic demo data via `cmd/seed-demo` (#47), project teams (#48), master data page (#49), location PDF report (#50), on-chain payment simulation on Sepolia (#52).
- **Construction depth** (P1 list in `docs/analisis/kelengkapan-domain.md`, all done): progress weighted by RAB value plus an S-curve (#61), daily field reports (#65), a contractor contract per project with progress-based payment terms and retention that become payables in billing (#71), and change orders approved by Admin with cumulative term amounts (#72).
- **Quality gates:** 5 CI jobs on every PR — Go unit tests, integration tests against real PostgreSQL + PostGIS (#66), `govulncheck` (#74), frontend lint/build/tests plus `npm audit` of production dependencies, and the Python scoring tests.
- **Docs for the AoL criteria:** UML and ERD (`docs/desain`), test plan and results at four levels (`docs/pengujian`), risks, OWASP and ethics (`docs/analisis/risiko-keamanan.md`), SDLC model, Gantt plan vs actual and WBS (`docs/proses`, PR #73), SE trends and professional ethics (`docs/analisis/tren-rpl.md`, PR #75).
- **Left:** the final presentation, contributions from every member through their own PRs, and the items in §5.

---

## 1. Current State (`main`, 2026-10-17)

### Backend (Go / Gin / GORM / PostGIS)
| Area | Status |
|---|---|
| Auth, 7-role access matrix checked on every request, last-admin guard | ✅ #44, #51; password length and login rate limit #62, #68 |
| 16 slices (`asset`, `auth`, `billing`, `finance`, `hr`, `inventory`, `location`, `master`, `news`, `procurement`, `project`, `regulation`, `reporting`, `sales`, `scoring`, `site`) | ✅ vertical slices, see `docs/requirements/backend/arsitektur.md` |
| `site` → `POST /api/site/evaluate` | ✅ predictive + descriptive; news fetched in parallel with a 2 s budget and a 15-minute cache (#58); fallback scores flagged `is_simulated` (#64) |
| `site` → `GET /api/site/tiles/:layer/:z/:x/:y` | ✅ auth-protected proxy to the scoring service's map tiles |
| `project` → weighted progress, progress history, S-curve, daily reports | ✅ #61, #65 |
| `project` → contractor contract, payment terms, retention, change orders | ✅ #71, #72. Rules and error codes: `docs/requirements/backend/slice-bisnis.md` |
| `billing` → on-chain payment requests on Sepolia | ✅ #52 |
| `cmd/migrate`, `cmd/seed`, `cmd/seed-demo` (`--reset-demo`), `cmd/import-rdtr` | ✅ |
| Integration tests (`-tags integration`, DB name must end in `_test`) | ✅ 9 tests, also in CI |

### Scoring service (`scoring/`, Python / FastAPI)
- `POST /score` → overall score, 5 dimension scores, risk flags, region, and `facts` (flood index, earthquake index, ZNT price).
- `GET /tiles/{bahaya|banjir|gempabumi|longsor|penduduk|harga_tanah}/{z}/{x}/{y}.png` → XYZ tiles from the bundle (min zoom 7, land price 10).
- Reads the **data bundle** (`BUNDLE_DIR`), built once by `build_bundle.py`. Trained on 1,489 sites. See `scoring/README.md`.

### Frontend (React / Vite / Tailwind / MapLibre)
| Route | Page |
|---|---|
| `/` | Overview: KPIs, 12-month cash-flow chart, "Perlu perhatian", projects |
| `/proyek`, `/proyek/:id` | Projects list + workspace with tabs Ringkasan (phases, RAB, permits, S-curve), Lapangan (daily reports), Tim, Analisis Lokasi, Keuangan (budget, cash, contractor terms, change orders) |
| `/lokasi` | Evaluasi (map + layers + score + regulation + news) and Bandingkan (2–4 saved sites), PDF export |
| `/keuangan` | Ringkasan, Anggaran, Kas, Jurnal, Buku besar |
| `/tagihan` | Faktur, Piutang, Utang with payment history and on-chain payment requests |
| `/penjualan` | Unit, Prospek, Pelanggan, Kontrak (+ installment payments) |
| `/pengadaan` | Pesanan pembelian, Penerimaan barang, Vendor, Permintaan pembelian |
| `/inventaris` | Stok dan mutasi, Material dan gudang, Aset dan alat |
| `/sdm` | Employees, attendance, payroll |
| `/laporan` | Reports with preview and PDF/XLSX/CSV export |
| `/data-master` | Chart of accounts, regions, units of measure |
| `/pengguna` | Users and roles (Admin only) |
| `/akun` | Own profile and password |
| `/bayar/:token` | Public on-chain payment page for buyers |

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

After every pull of `main`:

```bash
docker compose exec backend go run ./cmd/migrate
docker compose exec -e DEMO_USER_PASSWORD=<demo password> backend go run ./cmd/seed-demo
docker compose build backend    # once, since #74 pins Go 1.27.2 in the images
```

Integration tests need a database whose name ends in `_test`:

```bash
docker compose exec database createdb -U $POSTGRES_USER bangunin_aja_test
docker compose exec -e POSTGRES_DB=bangunin_aja_test backend go test -tags integration ./internal/integration/
```

Gotchas seen on Derick's laptop:
- A Windows PostgreSQL service on 5432 collides with the compose database — stop it or remap the port.
- Some networks cut Go's TLS 1.3 handshake (`go mod download` → `EOF`). Workaround: a local module mirror (`GOPROXY=file://...`).
- Another app on `[::1]:5173` hijacked `localhost:5173`; use `127.0.0.1` and add it to `CORS_ALLOWED_ORIGINS`.

**Demo check**: fresh clone → migrate + seed → register → create a project → evaluate a point on the map → the explanations must **not** end with "(data simulasi)" (that means the stub answered).

---

## 5. What's Left

### Next

| # | Task | Notes | Priority |
|---|---|---|---|
| L1 | Final presentation (slides or video) | Problem, solution, design, testing, team contributions, ethics, SE trends. Material: `docs/proses`, `docs/pengujian`, `docs/analisis` | high |
| L2 | Every member contributes through their own branch and PR | The VCS criterion grades consistent contributions from all members; most PRs since 2026-10-02 come from one account. See `docs/proses/jadwal.md` §3 | high |
| L3 | Branch protection on `main` | Repo owner (`jojondrw`): require PR, the 5 CI jobs, and one approval | high |
| L4 | Run acceptance scenario AT-13 (on-chain payment with a wallet on Sepolia) | AT-10 passed through the UI on 2026-10-17. `docs/pengujian/acceptance.md` | medium |
| L5 | Project-level access (OWASP A01) | Project roles in `project_member` are not enforced yet; needs a team decision on who sees which project | medium |
| L6 | Audit trail for role and status changes (A09) | Touches `auth/*` (see §6), so agree on it first | medium |
| L7 | `COOKIE_SECURE=true` for the HTTPS demo, CSP headers on Vercel, NIK encryption | `docs/analisis/risiko-keamanan.md` follow-up table | low |
| L8 | Billing toward the employer (receivable terms) and posting billing payments to cash | Current terms pay the contractor; billing payments do not create cash entries yet | low |

### Done

| # | Task | Notes | Priority |
|---|---|---|---|
| N3 | ~~On-chain payment simulation~~ | ✅ branch `feat/onchain-payment` — Sepolia testnet. Staff create a payment request for a customer invoice or receivable in `/tagihan` (QR with an EIP-681 URI, public link `/bayar/<token>`, amount in IDR and SepoliaETH at a *simulation* rate); the buyer pays from MetaMask or a phone wallet; `POST /api/pay/:token/confirm` verifies the tx on Sepolia itself over JSON-RPC (chain, receiver, value, receipt status, confirmations, block time inside the request window, hash never reused) and records one `billing_payment` row (method `onchain`, reference = tx hash) through the G6 path in the same DB transaction, idempotently. No private keys anywhere: the company only receives. Env `ONCHAIN_*` in `.env.example` (empty `ONCHAIN_RECEIVER_ADDRESS` = feature off); `cmd/migrate` adds `onchain_payment_request` and widens `billing_payment.reference` to 80. Details: `docs/requirements/backend/slice-bisnis.md` → Billing → Pembayaran on-chain. Left: no block scanning, so a QR payment from a phone still needs its tx hash pasted on the pay page | done |
| G1 | ~~Receiving goods doesn't move stock~~ | ✅ #37 — a goods receipt posts `in` stock movements in the same transaction; lines whose order unit ≠ stock unit are skipped and reported (no conversion data) | done |
| G2 | ~~Cash transactions don't post journal entries~~ | ✅ #38 — every cash transaction keeps one `KAS-<year>-<n>` journal in sync; `cmd/migrate` backfills old ones | done |
| G3 | ~~Receivables/payables unlinked; invoice + receivable double-count~~ | ✅ #36 — receivables link to project/contract/invoice (one receivable per invoice), payables to project/PO | done |
| G4 | ~~Leads not linked to their customer~~ | ✅ #35 — `POST /api/sales/leads/:id/convert` | done |
| G5 | ~~Dropdowns capped at 100 records~~ | ✅ #39 — `SearchSelect` (server-side search) in 45 places, backend names in procurement lists. #41 — account selects via `postable=true` leaf-account search, unit-of-measure via `SearchSelect`/`LookupName`, backend account/project names in cash, budget and journal responses. #42 — KPI summary endpoints (projects, POs, contracts, budgets) instead of summing the first 100 rows. Polish round — billing contract/PO/invoice choices via `SearchSelect`, contract installment totals (`/sales/contracts/:id/installments/summary`), project site KPIs (`/locations/saved/summary`), backend unit codes on stock lists. Left: nothing material | done |
| G6 | ~~Invoice payments and the linked receivable's payments aren't synced~~ | ✅ `billing_payment` history (date, amount, method, reference, note, who), `GET .../payments`, `lastPaidAt`; a linked invoice + receivable share one payment row and both balances move together (invoice locked first); linking rule + idempotent backfill in `cmd/migrate`. Billing tables now use backend names and `SearchSelect` filters. Pairs whose totals already differed before G6 keep separate rows and need a manual fix; `GET /api/billing/reconciliation` lists them and `/tagihan` shows a warning banner with the list | done |

---

## 6. Don't Touch

- `auth/*` and the login/session frontend. The only changes since 2026-10-03 were small security fixes (#62 password length, #68 login rate limit and security headers), each with tests.
- The `site` slice's stub fallback — it keeps the demo alive if the scoring service is down.
- `scoring/train.py`'s refusal to bulk-query the live ATR/BPN server.
- The RDTR scraper's throttling in `BanguninAja/scripts/ambil_rdtr.py`.

---

## 7. Open Questions

- Final presentation date and format (slides or video).
- Who owns L4 to L6, and through which PRs.
- Has `scoring-bundle.zip` been uploaded to the team Drive?
