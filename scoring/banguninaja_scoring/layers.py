import csv
import difflib
import math
import re
import threading
from dataclasses import dataclass

import numpy as np
import pyogrio
import rasterio
import requests
import shapely
from pyproj import Transformer
from rasterio.windows import Window

from . import config
from .geo import METRES_PER_DEG_LAT, PointIndex, metres_per_deg_lon, min_distance_m

TAG_PATTERN = re.compile(r'"([^"]+)"=>"([^"]*)"')

PROFILE_TAGS = {
    "hospital": {"amenity": {"hospital", "clinic"}, "healthcare": {"hospital", "clinic", "centre"}, "building": {"hospital"}},
    "mall": {"shop": {"mall", "department_store", "supermarket"}},
    "entertainment": {
        "amenity": {"bar", "cinema", "nightclub", "theatre", "casino", "karaoke"},
        "tourism": {"attraction", "theme_park", "zoo"},
        "leisure": {"amusement_arcade", "water_park", "bowling_alley"},
    },
    "housing": {"amenity": {"school"}},
}
DAILY_AMENITY_TAGS = {
    "amenity": {"school", "bank", "marketplace", "restaurant", "cafe", "fast_food"},
    "shop": {"convenience", "supermarket", "bakery"},
}


def parse_tags(text) -> dict:
    return dict(TAG_PATTERN.findall(text)) if isinstance(text, str) else {}


def matches(tags: dict, wanted: dict) -> bool:
    for key, values in wanted.items():
        for value in (tags.get(key) or "").split(";"):
            if value.strip() in values:
                return True
    return False


class Raster:
    def __init__(self, path):
        self.path = str(path)
        self.local = threading.local()
        with rasterio.open(self.path) as dataset:
            self.transform = dataset.transform
            self.nodata = dataset.nodata
            self.res_x = abs(dataset.res[0])
            self.geographic = dataset.crs.is_geographic
            self.project = None if self.geographic else Transformer.from_crs(4326, dataset.crs, always_xy=True).transform

    def dataset(self):
        handle = getattr(self.local, "dataset", None)
        if handle is None:
            handle = rasterio.open(self.path)
            self.local.dataset = handle
        return handle

    def window(self, lon: float, lat: float, half: int):
        x, y = (lon, lat) if self.project is None else self.project(lon, lat)
        row, col = rasterio.transform.rowcol(self.transform, x, y)
        size = 2 * half + 1
        fill = self.nodata if self.nodata is not None else np.nan
        values = self.dataset().read(
            1, window=Window(col - half, row - half, size, size), boundless=True, fill_value=fill
        ).astype("float64")
        valid = np.isfinite(values) & (values > -1e30)
        if self.nodata is not None:
            valid &= values != self.nodata
        return values, valid

    def mean_and_max(self, lon: float, lat: float):
        values, valid = self.window(lon, lat, 1)
        if not valid.any():
            return None, None
        return float(values[valid].mean()), float(values[valid].max())


class Population(Raster):
    def sum_within(self, lon: float, lat: float, radius_m: float) -> float | None:
        cell_x = self.res_x * metres_per_deg_lon(lat)
        cell_y = self.res_x * METRES_PER_DEG_LAT
        half = math.ceil(radius_m / min(cell_x, cell_y))
        values, valid = self.window(lon, lat, half)
        if not valid.any():
            return None
        offsets = np.arange(-half, half + 1)
        dy, dx = np.meshgrid(offsets * cell_y, offsets * cell_x, indexing="ij")
        inside = valid & (np.hypot(dx, dy) <= radius_m) & (values > 0)
        return float(values[inside].sum())


class Elevation(Raster):
    def slope_deg(self, lon: float, lat: float) -> float | None:
        values, valid = self.window(lon, lat, 1)
        if not valid.all():
            return None
        dx = self.res_x * metres_per_deg_lon(lat)
        dy = self.res_x * METRES_PER_DEG_LAT
        z = values
        gx = ((z[0, 2] + 2 * z[1, 2] + z[2, 2]) - (z[0, 0] + 2 * z[1, 0] + z[2, 0])) / (8 * dx)
        gy = ((z[2, 0] + 2 * z[2, 1] + z[2, 2]) - (z[0, 0] + 2 * z[0, 1] + z[0, 2])) / (8 * dy)
        return float(math.degrees(math.atan(math.hypot(gx, gy))))


class LineLayer:
    def __init__(self, path, layer):
        self.path = str(path)
        self.layer = layer

    def nearest_m(self, lon: float, lat: float, start_deg=0.02, max_deg=0.32) -> float | None:
        half = start_deg
        while half <= max_deg:
            frame = pyogrio.read_dataframe(
                self.path, layer=self.layer, bbox=(lon - half, lat - half, lon + half, lat + half), columns=[]
            )
            distance = min_distance_m(frame.geometry.values, lon, lat)
            if distance is not None and distance <= half * METRES_PER_DEG_LAT * 0.95:
                return distance
            half *= 4
        return None


@dataclass
class PoiIndexes:
    competitors: dict
    daily_amenities: PointIndex
    hospitals: PointIndex


def load_pois(path) -> PoiIndexes:
    points = pyogrio.read_dataframe(str(path), layer="poi_points", columns=["other_tags"])
    areas = pyogrio.read_dataframe(str(path), layer="poi_areas", columns=["other_tags"])
    areas["geometry"] = areas.geometry.representative_point()
    lons = np.concatenate([points.geometry.x.values, areas.geometry.x.values])
    lats = np.concatenate([points.geometry.y.values, areas.geometry.y.values])
    tags = [parse_tags(text) for text in list(points["other_tags"]) + list(areas["other_tags"])]

    def index_for(wanted):
        mask = np.array([matches(item, wanted) for item in tags], dtype=bool)
        return PointIndex(lons[mask], lats[mask])

    competitors = {profile: index_for(wanted) for profile, wanted in PROFILE_TAGS.items()}
    return PoiIndexes(competitors, index_for(DAILY_AMENITY_TAGS), competitors["hospital"])


def load_transit(path) -> PointIndex:
    frame = pyogrio.read_dataframe(str(path), layer="titik_transit", columns=[])
    return PointIndex(frame.geometry.x.values, frame.geometry.y.values)


def normalize_name(name: str) -> str:
    return re.sub(r"\s+", " ", (name or "").lower().replace("kab.", "").replace("kabupaten", "")).strip()


class ConstructionCost:
    def __init__(self, gadm_path, ikk_path):
        self.regions = pyogrio.read_dataframe(str(gadm_path), layer="ADM_ADM_2", columns=["NAME_2", "TYPE_2"])
        self.regions = self.regions[self.regions["TYPE_2"] != "Water Body"].reset_index(drop=True)
        self.index = self.regions.sindex
        with open(ikk_path, encoding="utf-8") as handle:
            self.ikk = {normalize_name(row["kabupaten_kota"]): float(row["ikk_2025"]) for row in csv.DictReader(handle)}
        self.national_median = float(np.median(list(self.ikk.values())))

    def region_at(self, lon: float, lat: float):
        hits = self.index.query(shapely.Point(lon, lat), predicate="intersects")
        if len(hits) == 0:
            return None
        return self.regions.iloc[int(hits[0])]

    def lookup(self, lon: float, lat: float):
        region = self.region_at(lon, lat)
        if region is None:
            return None, None
        name = normalize_name(region["NAME_2"])
        keys = [f"kota {name}", name] if region["TYPE_2"] == "Kota" else [name, f"kota {name}"]
        for key in keys:
            if key in self.ikk:
                return self.ikk[key], region["NAME_2"]
        close = difflib.get_close_matches(keys[0], self.ikk.keys(), n=1, cutoff=0.85)
        return (self.ikk[close[0]], region["NAME_2"]) if close else (None, region["NAME_2"])


class LandValue:
    def __init__(self, gpkg_path):
        self.path = str(gpkg_path)
        self.local_available = gpkg_path.exists() and gpkg_path.stat().st_size > 1_000_000_000

    def price_at(self, lon: float, lat: float):
        if self.local_available:
            try:
                return self._from_local(lon, lat), "ZNT"
            except Exception:
                pass
        return self._from_live(lon, lat), "ZNT (live)"

    def _from_local(self, lon: float, lat: float) -> float | None:
        for half in (0.002, 0.01):
            frame = pyogrio.read_dataframe(
                self.path, layer="znt", bbox=(lon - half, lat - half, lon + half, lat + half), columns=["NILAI"]
            )
            price = pick_price(frame.geometry.values, frame["NILAI"].values, lon, lat)
            if price is not None:
                return price
        return None

    def _from_live(self, lon: float, lat: float) -> float | None:
        half = 0.003
        response = requests.get(
            config.ZNT_LIVE_URL,
            params={
                "service": "WFS",
                "version": "2.0.0",
                "request": "GetFeature",
                "typeName": "petabpn:ZONANILAITANAH",
                "outputFormat": "application/json",
                "count": 50,
                "bbox": f"{lat - half},{lon - half},{lat + half},{lon + half},urn:ogc:def:crs:EPSG::4326",
            },
            headers={"User-Agent": "BanguninAja scoring (kuliah COMP6100001)"},
            timeout=4,
        )
        response.raise_for_status()
        features = response.json().get("features", [])
        geometries = [shapely.geometry.shape(item["geometry"]) for item in features if item.get("geometry")]
        prices = [item["properties"].get("NILAI") for item in features if item.get("geometry")]
        return pick_price(np.array(geometries, dtype=object), np.array(prices, dtype=float), lon, lat)


def pick_price(geometries, prices, lon: float, lat: float) -> float | None:
    if len(geometries) == 0:
        return None
    prices = np.asarray(prices, dtype=float)
    usable = np.isfinite(prices) & (prices > 0) & (prices < config.ZNT_MAX_VALID)
    if not usable.any():
        return None
    point = shapely.Point(lon, lat)
    containing = usable & shapely.contains(np.asarray(geometries), point)
    if containing.any():
        return float(np.median(prices[containing]))
    return float(np.median(prices[usable]))
