"""Turn the ~18 GB raw GIS data into a compact bundle the scoring service can ship with.

Usage:
    DATA_DIR=/path/to/BanguninAja/data/raw python build_bundle.py --out bundle/

Output (all EPSG:4326 unless noted, tiled + DEFLATE):
    hazard_<name>.tif   uint8, index * 250, nodata 255 (EPSG:3395, original 250 m grid)
    population.tif      float32 people per cell, WorldPop summed 3x3 (~275 m)
    dem.tif             copied
    road_distance.tif   uint16 metres to the nearest road on a 0.001 deg grid, capped
    land_price.tif      uint32 Rp/m2 from ZNT on a 0.001 deg grid, 0 = no zone
    poi.gpkg, transit.gpkg, regions.gpkg, ikk.csv
"""

import argparse
import logging
import math
import shutil
import time
from pathlib import Path

import numpy as np
import pyogrio
import rasterio
from rasterio.features import rasterize
from rasterio.transform import from_origin
from rasterio.windows import Window
from scipy.ndimage import distance_transform_edt

from banguninaja_scoring import config
from banguninaja_scoring.geo import METRES_PER_DEG_LAT, metres_per_deg_lon

log = logging.getLogger("bundle")

GRID_DEG = 0.001
EXTENT = (94.9, -11.1, 141.1, 6.1)
TILE = 2048
ROAD_MARGIN = 64
ROAD_CAP_M = config.ROAD_DISTANCE_CAP_M
HAZARD_SCALE = config.HAZARD_SCALE
HAZARD_NODATA = 255
CREATE = dict(driver="GTiff", tiled=True, blockxsize=512, blockysize=512, compress="deflate", BIGTIFF="IF_SAFER")


def grid():
    west, south, east, north = EXTENT
    width = math.ceil((east - west) / GRID_DEG)
    height = math.ceil((north - south) / GRID_DEG)
    return from_origin(west, north, GRID_DEG, GRID_DEG), width, height


def tiles(width, height):
    for row in range(0, height, TILE):
        for col in range(0, width, TILE):
            yield Window(col, row, min(TILE, width - col), min(TILE, height - row))


def build_hazards(out: Path):
    for name in ("multi",) + config.HAZARDS:
        source = config.data_path("inarisk_wcs", f"inarisk_{name}_indonesia.tif")
        with rasterio.open(source) as src:
            profile = src.profile | CREATE | dict(dtype="uint8", nodata=HAZARD_NODATA, predictor=2)
            with rasterio.open(out / f"hazard_{name}.tif", "w", **profile) as dst:
                for _, window in src.block_windows(1):
                    values = src.read(1, window=window)
                    valid = np.isfinite(values) & (values > -1e30) & (values != src.nodata)
                    scaled = np.where(valid, np.clip(np.rint(values * HAZARD_SCALE), 0, HAZARD_SCALE), HAZARD_NODATA)
                    dst.write(scaled.astype("uint8"), 1, window=window)
        log.info("hazard %s done", name)


def build_population(out: Path, factor=3, rows_per_chunk=1800):
    with rasterio.open(config.data_path("worldpop_idn_2020.tif")) as src:
        width = src.width // factor
        height = math.ceil(src.height / factor)
        transform = src.transform * src.transform.scale(factor, factor)
        profile = dict(CREATE, height=height, width=width, count=1, dtype="float32", crs=src.crs,
                       transform=transform, nodata=-1.0, predictor=3)
        with rasterio.open(out / "population.tif", "w", **profile) as dst:
            for start in range(0, src.height, rows_per_chunk):
                rows = min(rows_per_chunk, src.height - start)
                values = src.read(1, window=Window(0, start, width * factor, rows))
                values = np.where(values > 0, values, 0).astype("float64")
                pad = (-rows) % factor
                if pad:
                    values = np.vstack([values, np.zeros((pad, values.shape[1]))])
                summed = values.reshape(-1, factor, width, factor).sum(axis=(1, 3)).astype("float32")
                dst.write(summed, 1, window=Window(0, start // factor, width, summed.shape[0]))
    log.info("population done")


def tile_bounds(window: Window, transform, margin=0):
    west = transform.c + (window.col_off - margin) * GRID_DEG
    north = transform.f - (window.row_off - margin) * GRID_DEG
    east = west + (window.width + 2 * margin) * GRID_DEG
    south = north - (window.height + 2 * margin) * GRID_DEG
    return west, south, east, north


def road_tile(path, window, transform):
    west, south, east, north = tile_bounds(window, transform, ROAD_MARGIN)
    shape = (window.height + 2 * ROAD_MARGIN, window.width + 2 * ROAD_MARGIN)
    frame = pyogrio.read_dataframe(path, layer="jalan", bbox=(west, south, east, north), columns=[])
    if frame.empty:
        return np.full((window.height, window.width), ROAD_CAP_M, dtype="uint16")
    burned = rasterize(frame.geometry.values, out_shape=shape, transform=from_origin(west, north, GRID_DEG, GRID_DEG),
                       fill=0, default_value=1, all_touched=True, dtype="uint8")
    lat = (north + south) / 2
    sampling = (GRID_DEG * METRES_PER_DEG_LAT, GRID_DEG * metres_per_deg_lon(lat))
    distance = distance_transform_edt(burned == 0, sampling=sampling)
    inner = distance[ROAD_MARGIN:ROAD_MARGIN + window.height, ROAD_MARGIN:ROAD_MARGIN + window.width]
    return np.minimum(inner, ROAD_CAP_M).astype("uint16")


def land_tile(path, window, transform):
    west, south, east, north = tile_bounds(window, transform)
    frame = pyogrio.read_dataframe(path, layer="znt", bbox=(west, south, east, north), columns=["NILAI"])
    shape = (window.height, window.width)
    if frame.empty:
        return np.zeros(shape, dtype="uint32")
    values = frame["NILAI"].to_numpy(dtype="float64")
    keep = np.isfinite(values) & (values > 0) & (values < config.ZNT_MAX_VALID)
    if not keep.any():
        return np.zeros(shape, dtype="uint32")
    order = np.argsort(values[keep])
    shapes = zip(frame.geometry.values[keep][order], values[keep][order].astype("uint32"))
    return rasterize(shapes, out_shape=shape, transform=from_origin(west, north, GRID_DEG, GRID_DEG),
                     fill=0, dtype="uint32")


def build_tiled(out_file: Path, source: Path, make_tile, dtype, nodata, predictor):
    transform, width, height = grid()
    profile = dict(CREATE, height=height, width=width, count=1, dtype=dtype, crs="EPSG:4326",
                   transform=transform, nodata=nodata, predictor=predictor)
    windows = list(tiles(width, height))
    started = time.time()
    with rasterio.open(out_file, "w", **profile) as dst:
        for index, window in enumerate(windows, start=1):
            dst.write(make_tile(str(source), window, transform), 1, window=window)
            if index % 20 == 0:
                log.info("%s %d/%d tiles (%.0fs)", out_file.name, index, len(windows), time.time() - started)
    log.info("%s done", out_file.name)


def copy_small_files(out: Path):
    shutil.copy(config.data_path("dem_copernicus30m_indonesia_2026.tif"), out / "dem.tif")
    shutil.copy(config.data_path("poi_kompetitor_indonesia_2026.gpkg"), out / "poi.gpkg")
    shutil.copy(config.data_path("bps_ikk_indonesia_2025.csv"), out / "ikk.csv")
    transit = pyogrio.read_dataframe(str(config.data_path("jaringan_transportasi_indonesia_2026.gpkg")), layer="titik_transit", columns=[])
    pyogrio.write_dataframe(transit, out / "transit.gpkg", layer="titik_transit")
    regions = pyogrio.read_dataframe(str(config.data_path("gadm41_indonesia.gpkg")), layer="ADM_ADM_2", columns=["NAME_2", "TYPE_2"])
    regions["geometry"] = regions.geometry.simplify(0.0005, preserve_topology=True)
    pyogrio.write_dataframe(regions, out / "regions.gpkg", layer="ADM_ADM_2")
    log.info("small files done")


STEPS = ("small", "hazards", "population", "roads", "land")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", default="bundle")
    parser.add_argument("--only", choices=STEPS, nargs="*")
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(message)s")
    out = Path(args.out)
    out.mkdir(parents=True, exist_ok=True)
    steps = args.only or STEPS
    if "small" in steps:
        copy_small_files(out)
    if "hazards" in steps:
        build_hazards(out)
    if "population" in steps:
        build_population(out)
    if "roads" in steps:
        build_tiled(out / "road_distance.tif", config.data_path("jaringan_transportasi_indonesia_2026.gpkg"),
                    road_tile, "uint16", None, 2)
    if "land" in steps:
        build_tiled(out / "land_price.tif", config.data_path("znt_atrbpn_indonesia.gpkg"), land_tile, "uint32", 0, 2)


if __name__ == "__main__":
    main()
