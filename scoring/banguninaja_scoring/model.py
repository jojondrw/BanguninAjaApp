import json
from dataclasses import dataclass

import numpy as np

from . import config

QUANTILE_FEATURES = (
    "hazard_multi",
    "slope_deg",
    "road_dist_m",
    "transit_dist_m",
    "hospital_dist_m",
    "pop_2km",
    "amenities_2km",
    "land_price",
    "ikk",
) + tuple(f"competitors_3km_{profile}" for profile in config.PROFILES)

COMPETITOR_LABELS = {
    "hospital": "rumah sakit/klinik",
    "mall": "pusat perbelanjaan/supermarket",
    "entertainment": "tempat hiburan",
}


@dataclass
class Model:
    quantiles: dict
    weights: dict
    meta: dict

    @classmethod
    def load(cls, path=config.MODEL_PATH):
        with open(path, encoding="utf-8") as handle:
            payload = json.load(handle)
        quantiles = {name: np.asarray(values, dtype=float) for name, values in payload["quantiles"].items()}
        return cls(quantiles, payload["weights"], payload.get("meta", {}))

    def percentile(self, name: str, value) -> float | None:
        if value is None or name not in self.quantiles:
            return None
        sample = self.quantiles[name]
        left = np.searchsorted(sample, value, side="left")
        right = np.searchsorted(sample, value, side="right")
        return float((left + right) / (2 * len(sample)) * 100)


def combine(parts) -> float:
    available = [(score, weight) for score, weight in parts if score is not None]
    if not available:
        return 50.0
    total = sum(weight for _, weight in available)
    return sum(score * weight for score, weight in available) / total


def inverse(score):
    return None if score is None else 100 - score


def fmt_int(value) -> str:
    return f"{value:,.0f}".replace(",", ".")


def fmt_decimal(value, digits=2) -> str:
    return f"{value:.{digits}f}".replace(".", ",")


def fmt_distance(metres) -> str:
    if metres is None:
        return "tidak diketahui"
    if metres < 1000:
        return f"{fmt_int(round(metres, -1))} m"
    return f"{fmt_decimal(metres / 1000, 1)} km"


def hazard_level(value) -> str:
    if value >= 0.66:
        return "tinggi"
    if value >= 0.33:
        return "sedang"
    return "rendah"


def physical(model: Model, f: dict):
    hazard, slope = f.get("hazard_multi"), f.get("slope_deg")
    if hazard is None:
        return 50.0, "Indeks bahaya InaRISK tidak tersedia di titik ini (di luar cakupan data)."
    value = combine([
        (inverse(model.percentile("hazard_multi", hazard)), 0.7),
        (inverse(model.percentile("slope_deg", slope)), 0.3),
    ])
    hazard_text = "tidak tersedia" if hazard is None else f"{fmt_decimal(hazard)} ({hazard_level(hazard)})"
    slope_text = "tidak tersedia" if slope is None else f"sekitar {fmt_decimal(slope, 1)}°"
    return value, f"Indeks multi-bahaya InaRISK {hazard_text}; kemiringan lahan {slope_text}."


def access(model: Model, f: dict):
    value = combine([
        (inverse(model.percentile("road_dist_m", f.get("road_dist_m"))), 0.4),
        (inverse(model.percentile("transit_dist_m", f.get("transit_dist_m"))), 0.3),
        (inverse(model.percentile("hospital_dist_m", f.get("hospital_dist_m"))), 0.3),
    ])
    text = (
        f"Jalan terdekat {fmt_distance(f.get('road_dist_m'))}, simpul transit {fmt_distance(f.get('transit_dist_m'))}, "
        f"rumah sakit/klinik {fmt_distance(f.get('hospital_dist_m'))}."
    )
    return value, text


def demography(model: Model, f: dict):
    population = f.get("pop_2km")
    value = combine([(model.percentile("pop_2km", population), 1.0)])
    if population is None:
        return value, "Data kepadatan penduduk tidak tersedia di titik ini."
    return value, f"Sekitar {fmt_int(population)} penduduk dalam radius 2 km (WorldPop 2020)."


def market(model: Model, f: dict, profile: str):
    amenities = f.get("amenities_2km")
    amenity_score = model.percentile("amenities_2km", amenities)
    if profile == "housing":
        value = combine([(amenity_score, 1.0)])
        return value, f"{fmt_int(amenities or 0)} fasilitas layanan harian (sekolah, toko, bank, kuliner) dalam radius 2 km."
    competitors = f.get(f"competitors_3km_{profile}")
    value = combine([
        (inverse(model.percentile(f"competitors_3km_{profile}", competitors)), 0.6),
        (amenity_score, 0.4),
    ])
    text = (
        f"{fmt_int(competitors or 0)} {COMPETITOR_LABELS[profile]} sejenis dalam radius 3 km; "
        f"{fmt_int(amenities or 0)} fasilitas komersial dan layanan harian dalam 2 km."
    )
    return value, text


def finance(model: Model, f: dict):
    price, ikk = f.get("land_price"), f.get("ikk")
    value = combine([
        (inverse(model.percentile("land_price", price)), 0.6),
        (inverse(model.percentile("ikk", ikk)), 0.4),
    ])
    price_text = "Harga tanah ZNT tidak tersedia di titik ini" if price is None else f"Harga tanah (ZNT) sekitar Rp{fmt_int(price)}/m²"
    region = f.get("region") or "wilayah ini"
    ikk_text = "indeks kemahalan konstruksi tidak tersedia" if ikk is None else f"indeks kemahalan konstruksi {region} {fmt_decimal(ikk, 1)}"
    return value, f"{price_text}; {ikk_text}."


def risk_flags(f: dict) -> list:
    flags = []
    for name in config.HAZARDS:
        value = f.get(f"hazard_{name}")
        if value is None or value < 0.4:
            continue
        severity = "high" if value >= 0.66 else "medium"
        label = config.HAZARD_LABELS[name]
        flags.append({
            "code": f"rawan_{name}",
            "severity": severity,
            "message": f"Indeks bahaya {label} {hazard_level(value)} ({fmt_decimal(value)}).",
        })
    slope = f.get("slope_deg")
    if slope is not None and slope >= 15:
        flags.append({
            "code": "lereng_curam",
            "severity": "high" if slope >= 25 else "medium",
            "message": f"Kemiringan lahan sekitar {fmt_decimal(slope, 1)}°, perlu kajian geoteknik.",
        })
    if f.get("hazard_multi") is None and f.get("pop_2km") is None:
        flags.append({
            "code": "di_luar_cakupan_data",
            "severity": "high",
            "message": "Titik ini berada di luar cakupan data (kemungkinan perairan).",
        })
    return flags


def dimension_values(model: Model, f: dict, profile: str) -> dict:
    return {
        "fisik_lingkungan": physical(model, f),
        "infrastruktur": access(model, f),
        "demografi_sosial": demography(model, f),
        "pasar_kompetisi": market(model, f, profile),
        "finansial_proyek": finance(model, f),
    }


def score(model: Model, f: dict, profile: str) -> dict:
    dimensions = dimension_values(model, f, profile)
    weights = model.weights[profile]
    overall = sum(weights[code] * dimensions[code][0] for code in config.DIMENSIONS) / 100
    return {
        "overall_score": int(round(overall)),
        "dimension_scores": [
            {"dimension_code": code, "value": int(round(dimensions[code][0])), "explanation": dimensions[code][1]}
            for code in config.DIMENSIONS
        ],
        "risk_flags": risk_flags(f),
        "region": f.get("region") or None,
    }
