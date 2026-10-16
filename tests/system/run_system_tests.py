"""System test BanguninAja.

Menjalankan skenario utama terhadap aplikasi yang sedang berjalan (backend,
database, dan layanan penilaian) lewat API sungguhan, lalu mencetak tabel
hasil. Semua data uji diberi awalan UJI- dan dihapus di akhir, termasuk saat
ada langkah yang gagal.

Env:
  BANGUNINAJA_API       alamat API, bawaan http://127.0.0.1:8080/api
  BANGUNINAJA_EMAIL     email akun dengan hak tulis Proyek dan Analisis Lokasi
  BANGUNINAJA_PASSWORD  kata sandinya

Keluar dengan kode 1 kalau ada langkah yang gagal.
"""
import datetime
import json
import os
import sys
import time
import urllib.error
import urllib.request
import uuid

API = os.environ.get("BANGUNINAJA_API", "http://127.0.0.1:8080/api").rstrip("/")
EMAIL = os.environ.get("BANGUNINAJA_EMAIL", "")
PASSWORD = os.environ.get("BANGUNINAJA_PASSWORD", "")

results = []
cleanup = []


def call(method, path, token=None, body=None, raw=False):
    data = json.dumps(body).encode() if body is not None else None
    request = urllib.request.Request(API + path, data=data, method=method)
    request.add_header("Content-Type", "application/json")
    if token:
        request.add_header("Authorization", "Bearer " + token)
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            payload = response.read()
            if raw:
                return response.status, payload, response.headers.get("Content-Type", "")
            return response.status, json.loads(payload) if payload else None
    except urllib.error.HTTPError as error:
        payload = error.read()
        if raw:
            return error.code, payload, error.headers.get("Content-Type", "")
        try:
            return error.code, json.loads(payload)
        except json.JSONDecodeError:
            return error.code, None


def check(scenario, step, passed, detail=""):
    results.append((scenario, step, bool(passed), detail))
    return bool(passed)


def error_code(body):
    return (body or {}).get("error", {}).get("code", "")


def api_date(day):
    return f"{day.isoformat()}T00:00:00Z"


def scenario_authentication():
    sid = "ST-01 Autentikasi"
    status, body = call("POST", "/auth/login", body={"email": EMAIL, "password": PASSWORD})
    if not check(sid, "Masuk dengan kredensial benar", status == 200, f"HTTP {status}"):
        return None
    token = body["data"]["accessToken"]

    status, body = call("GET", "/auth/me", token)
    check(sid, "Profil pengguna terbaca", status == 200 and body["data"]["email"] == EMAIL, f"HTTP {status}")

    status, body = call("POST", "/auth/login", body={"email": EMAIL, "password": PASSWORD + "-salah"})
    check(sid, "Kata sandi salah ditolak", status == 401 and error_code(body) == "invalid_credential", f"HTTP {status} {error_code(body)}")

    status, _ = call("GET", "/projects")
    check(sid, "Tanpa token ditolak", status == 401, f"HTTP {status}")

    status, _ = call("GET", "/projects", "token-palsu")
    check(sid, "Token palsu ditolak", status == 401, f"HTTP {status}")
    return token


def scenario_project(token):
    sid = "ST-02 Proyek dan progres berbobot"
    today = datetime.date.today()
    status, body = call("POST", "/projects", token, {
        "code": "UJI-" + uuid.uuid4().hex[:6].upper(),
        "name": "UJI System Test",
        "type": "Perumahan",
        "startDate": api_date(today - datetime.timedelta(days=60)),
        "targetEndDate": api_date(today + datetime.timedelta(days=300)),
        "contractValue": 5_000_000_000,
    })
    if not check(sid, "Membuat proyek", status == 201, f"HTTP {status}"):
        return None
    project_id = body["data"]["id"]
    cleanup.append(("DELETE", f"/projects/{project_id}"))

    phases = []
    for order, (name, start, end, progress) in enumerate([
        ("Persiapan", -60, -31, 100),
        ("Struktur", -30, 300, 20),
    ], start=1):
        status, body = call("POST", f"/projects/{project_id}/phases", token, {
            "name": name, "sortOrder": order, "progress": progress,
            "startDate": api_date(today + datetime.timedelta(days=start)),
            "targetEndDate": api_date(today + datetime.timedelta(days=end)),
        })
        check(sid, f"Menambah tahap {name}", status == 201, f"HTTP {status}")
        phases.append(body["data"]["id"] if status == 201 else None)

    status, body = call("GET", "/master/units-of-measure?pageSize=100", token)
    units = (body or {}).get("data", {}).get("items", []) if status == 200 else []
    if not check(sid, "Membaca satuan dari data master", len(units) > 0, f"{len(units)} satuan"):
        return project_id
    unit = units[0]["id"]

    for phase_id, code, price in [(phases[0], "1.1", 100_000), (phases[1], "2.1", 900_000)]:
        status, body = call("POST", f"/projects/{project_id}/budget-items", token, {
            "phaseId": phase_id, "code": code, "description": f"UJI item {code}",
            "volume": 100, "unitOfMeasureId": unit, "unitPrice": price,
        })
        check(sid, f"Item RAB {code} dikaitkan ke tahap", status == 201 and body["data"]["phaseId"] == phase_id, f"HTTP {status}")

    status, body = call("GET", f"/projects/{project_id}", token)
    progress = body["data"]["progress"] if status == 200 else None
    check(sid, "Progres berbobot RAB: 10% x 100 + 90% x 20 = 28", progress == 28, f"progres {progress}")

    status, body = call("GET", f"/projects/{project_id}/progress-curve", token)
    curve = (body or {}).get("data", {})
    check(sid, "Kurva S berbobot RAB dengan titik rencana",
          status == 200 and curve.get("weighting") == "budget" and len(curve.get("points", [])) > 10,
          f"weighting {curve.get('weighting')}, {len(curve.get('points', []))} titik")
    check(sid, "Deviasi hari ini dihitung", curve.get("today", {}).get("deviation") is not None,
          f"deviasi {curve.get('today', {}).get('deviation')}")
    return project_id


def scenario_daily_report(token, project_id):
    sid = "ST-03 Laporan harian"
    today = datetime.date.today()
    report = {
        "reportDate": api_date(today), "weatherMorning": "sunny", "weatherAfternoon": "rain",
        "workers": 30, "equipment": "UJI molen", "workDone": "UJI pengecoran", "issues": "UJI hujan sore",
    }
    status, _ = call("POST", f"/projects/{project_id}/daily-reports", token, report)
    check(sid, "Menulis laporan hari ini", status == 201, f"HTTP {status}")

    status, body = call("POST", f"/projects/{project_id}/daily-reports", token, report)
    check(sid, "Laporan kedua di tanggal yang sama ditolak", status == 409 and error_code(body) == "daily_report_date_used",
          f"HTTP {status} {error_code(body)}")

    future = dict(report, reportDate=api_date(today + datetime.timedelta(days=1)))
    status, body = call("POST", f"/projects/{project_id}/daily-reports", token, future)
    check(sid, "Laporan untuk besok ditolak", status == 422 and error_code(body) == "daily_report_in_future",
          f"HTTP {status} {error_code(body)}")

    status, body = call("GET", f"/projects/{project_id}/daily-reports", token)
    data = (body or {}).get("data", {})
    check(sid, "Ringkasan 30 hari menghitung hari hujan",
          status == 200 and data.get("reportsLast30") == 1 and data.get("rainDaysLast30") == 1,
          f"{data.get('reportsLast30')} laporan, {data.get('rainDaysLast30')} hujan")


def scenario_site(token, project_id):
    sid = "ST-04 Evaluasi lokasi"
    status, body = call("GET", "/scoring/building-profiles", token)
    profiles = body["data"] if status == 200 else []
    profiles = profiles["items"] if isinstance(profiles, dict) else profiles
    housing = next((p for p in profiles if p["code"] == "housing"), None)
    if not check(sid, "Profil bangunan tersedia", housing is not None, f"{len(profiles)} profil"):
        return

    started = time.perf_counter()
    status, body = call("POST", "/site/evaluate", token, {
        "latitude": -6.9175, "longitude": 107.6191, "building_profile_id": housing["id"],
        "name": "UJI System Test Bandung", "project_id": project_id,
    })
    elapsed = (time.perf_counter() - started) * 1000
    result = (body or {}).get("data") or {}
    location_id = result.get("saved_location_id")
    if location_id:
        cleanup.insert(0, ("DELETE", f"/locations/saved/{location_id}"))
    if not check(sid, "Mengevaluasi titik di Bandung", status == 201 and location_id, f"HTTP {status}, {elapsed:.0f} ms"):
        return
    predictive = result["predictive"]

    check(sid, "Skor 0 sampai 100 dengan 5 dimensi",
          0 <= predictive["overall_score"] <= 100 and len(predictive["dimension_scores"]) == 5,
          f"skor {predictive['overall_score']}")
    check(sid, "Skor berasal dari model, bukan cadangan", predictive["is_simulated"] is False,
          f"is_simulated {predictive['is_simulated']}")
    check(sid, "Evaluasi selesai di bawah 3 detik", elapsed < 3000, f"{elapsed:.0f} ms")

    status, body = call("GET", f"/locations/saved/{location_id}", token)
    stored = (body or {}).get("data", {})
    check(sid, "Lokasi tersimpan dan terkait ke proyek",
          status == 200 and stored.get("projectId") == project_id and stored.get("scoreSimulated") is False,
          f"HTTP {status}")

    status, payload, content_type = call("GET", f"/locations/saved/{location_id}/report", token, raw=True)
    check(sid, "Laporan PDF lokasi bisa diunduh",
          status == 200 and content_type.startswith("application/pdf") and payload[:4] == b"%PDF",
          f"HTTP {status}, {len(payload)} byte")


def run_cleanup(token):
    sid = "ST-05 Pembersihan"
    for method, path in cleanup:
        status, _ = call(method, path, token)
        check(sid, f"{method} {path.split('/')[1]}", status in (200, 204), f"HTTP {status}")


def main():
    if not EMAIL or not PASSWORD:
        print("Isi BANGUNINAJA_EMAIL dan BANGUNINAJA_PASSWORD", file=sys.stderr)
        return 2

    started = time.perf_counter()
    token = scenario_authentication()
    if token:
        try:
            project_id = scenario_project(token)
            if project_id:
                scenario_daily_report(token, project_id)
                scenario_site(token, project_id)
        finally:
            run_cleanup(token)

    failed = [result for result in results if not result[2]]
    width = max(len(result[1]) for result in results)
    print(f"System test BanguninAja  {datetime.datetime.now():%Y-%m-%d %H:%M}  API {API}")
    current = None
    for scenario, step, passed, detail in results:
        if scenario != current:
            print(f"\n{scenario}")
            current = scenario
        print(f"  {'LULUS' if passed else 'GAGAL'}  {step.ljust(width)}  {detail}")
    print(f"\n{len(results) - len(failed)}/{len(results)} langkah lulus dalam {time.perf_counter() - started:.1f} detik")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
