"""Train the scoring model: sample sites across Indonesia, learn per-profile weights.

Usage:
    DATA_DIR=/path/to/BanguninAja/data/raw python train.py [--per-region 3] [--seed 42]

Writes model/model.json (percentile tables + weights) and model/training_points.csv,
and copies the learned weights into backend/cmd/seed/data/scoring.json.
"""

import argparse
import json
import logging
import time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path

import numpy as np
import pandas as pd
import pyogrio
import shapely
from sklearn.linear_model import LinearRegression

from banguninaja_scoring import config
from banguninaja_scoring.features import FeatureExtractor
from banguninaja_scoring.model import QUANTILE_FEATURES, Model, dimension_values

HERE = Path(__file__).resolve().parent
SEED_FILE = HERE.parent / "backend" / "cmd" / "seed" / "data" / "scoring.json"
WEIGHT_FLOOR = 10
LEARNED_SHARE = 100 - WEIGHT_FLOOR * len(config.DIMENSIONS)

log = logging.getLogger("train")


def sample_points(per_region: int, seed: int) -> list:
    regions = pyogrio.read_dataframe(str(config.data_path("gadm41_indonesia.gpkg")), layer="ADM_ADM_2", columns=["TYPE_2"])
    regions = regions[regions["TYPE_2"] != "Water Body"]
    rng = np.random.default_rng(seed)
    points = []
    for geometry in regions.geometry:
        minx, miny, maxx, maxy = geometry.bounds
        found, tries = 0, 0
        while found < per_region and tries < per_region * 60:
            tries += 1
            lon, lat = rng.uniform(minx, maxx), rng.uniform(miny, maxy)
            if geometry.contains(shapely.Point(lon, lat)):
                points.append((lon, lat))
                found += 1
    return points


def extract_all(extractor: FeatureExtractor, points: list) -> pd.DataFrame:
    started = time.time()
    use_land_price = extractor.land.local_available
    if not use_land_price:
        log.warning("local ZNT file missing: training without land price (never bulk-query the live ATR/BPN server)")

    def work(point):
        try:
            return extractor.extract(point[0], point[1], with_land_price=use_land_price)
        except Exception as exc:
            log.warning("skip %s: %s", point, exc)
            return None

    rows = []
    with ThreadPoolExecutor(6) as pool:
        for index, row in enumerate(pool.map(work, points), start=1):
            if row is not None:
                rows.append(row)
            if index % 100 == 0:
                log.info("%d/%d points (%.0fs)", index, len(points), time.time() - started)
    return pd.DataFrame(rows)


def build_quantiles(frame: pd.DataFrame) -> dict:
    quantiles = {}
    for name in QUANTILE_FEATURES:
        values = pd.to_numeric(frame[name], errors="coerce").dropna().to_numpy()
        quantiles[name] = np.quantile(values, np.linspace(0, 1, 101)).round(4).tolist()
    return quantiles


def target(frame: pd.DataFrame, profile: str) -> np.ndarray:
    return np.log1p(frame["pop_5km"].to_numpy() / (frame[f"competitors_5km_{profile}"].to_numpy() + 1))


def to_weights(coefficients: np.ndarray) -> dict:
    total = coefficients.sum()
    shares = coefficients / total if total > 0 else np.full(len(coefficients), 1 / len(coefficients))
    raw = WEIGHT_FLOOR + LEARNED_SHARE * shares
    weights = np.floor(raw).astype(int)
    for index in np.argsort(-(raw - weights))[: 100 - weights.sum()]:
        weights[index] += 1
    return dict(zip(config.DIMENSIONS, (int(w) for w in weights)))


def fit_profile(model: Model, frame: pd.DataFrame, profile: str):
    records = frame.to_dict("records")
    matrix = np.array([[dimension_values(model, row, profile)[code][0] for code in config.DIMENSIONS] for row in records])
    y = target(frame, profile)
    regression = LinearRegression(positive=True).fit(matrix, y)
    fit = {
        "r2": round(float(regression.score(matrix, y)), 4),
        "coefficients": dict(zip(config.DIMENSIONS, (round(float(c), 6) for c in regression.coef_))),
        "target": f"log1p(pop_5km / (competitors_5km_{profile} + 1))",
    }
    return to_weights(regression.coef_), fit


def update_seed(weights: dict):
    seed = json.loads(SEED_FILE.read_text(encoding="utf-8"))
    seed["weights"] = weights
    SEED_FILE.write_text(json.dumps(seed, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--per-region", type=int, default=3)
    parser.add_argument("--seed", type=int, default=42)
    parser.add_argument("--reuse-points", action="store_true", help="refit from model/training_points.csv")
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")

    points_file = HERE / "model" / "training_points.csv"
    if args.reuse_points:
        frame = pd.read_csv(points_file)
        land_source = "ZNT local"
    else:
        extractor = FeatureExtractor()
        land_source = "ZNT local" if extractor.land.local_available else "none"
        points = sample_points(args.per_region, args.seed)
        log.info("sampled %d candidate points", len(points))
        frame = extract_all(extractor, points)
        frame = frame[frame["pop_2km"].notna() & frame["hazard_multi"].notna() & (frame["pop_5km"] > 0)].reset_index(drop=True)
        log.info("%d usable points after dropping water/no-data", len(frame))
        frame.to_csv(points_file, index=False)

    model = Model(quantiles={k: np.asarray(v) for k, v in build_quantiles(frame).items()}, weights={}, meta={})
    weights, fits = {}, {}
    for profile in config.PROFILES:
        weights[profile], fits[profile] = fit_profile(model, frame, profile)
        log.info("%s: R2=%.3f weights=%s", profile, fits[profile]["r2"], weights[profile])

    payload = {
        "meta": {
            "trained_at": time.strftime("%Y-%m-%dT%H:%M:%S"),
            "points": int(len(frame)),
            "land_price_source": land_source,
            "weighting": f"each dimension gets {WEIGHT_FLOOR}% fixed plus a learned share of {LEARNED_SHARE}% "
            "(shrinkage toward equal weights, because the target shares population with demografi_sosial)",
            "fits": fits,
        },
        "quantiles": {k: list(map(float, v)) for k, v in model.quantiles.items()},
        "weights": weights,
    }
    (HERE / "model" / "model.json").write_text(json.dumps(payload, indent=2), encoding="utf-8")
    update_seed(weights)
    log.info("wrote model/model.json and updated %s", SEED_FILE)


if __name__ == "__main__":
    main()
