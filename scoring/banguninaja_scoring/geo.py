import math

import numpy as np
import shapely
from scipy.spatial import cKDTree

EARTH_RADIUS_M = 6_371_000
METRES_PER_DEG_LAT = 110_540
METRES_PER_DEG_LON_EQUATOR = 111_320


def metres_per_deg_lon(lat: float) -> float:
    return METRES_PER_DEG_LON_EQUATOR * math.cos(math.radians(lat))


def to_xyz(lons, lats) -> np.ndarray:
    lon = np.radians(np.asarray(lons, dtype=float))
    lat = np.radians(np.asarray(lats, dtype=float))
    return np.column_stack(
        [np.cos(lat) * np.cos(lon), np.cos(lat) * np.sin(lon), np.sin(lat)]
    ) * EARTH_RADIUS_M


class PointIndex:
    def __init__(self, lons, lats):
        self.size = len(lons)
        self.tree = cKDTree(to_xyz(lons, lats)) if self.size else None

    def count_within(self, lon: float, lat: float, radius_m: float) -> int:
        if self.tree is None:
            return 0
        return len(self.tree.query_ball_point(to_xyz([lon], [lat])[0], radius_m))

    def nearest_m(self, lon: float, lat: float) -> float | None:
        if self.tree is None:
            return None
        distance, _ = self.tree.query(to_xyz([lon], [lat])[0])
        return float(distance)


def min_distance_m(geometries, lon: float, lat: float) -> float | None:
    if len(geometries) == 0:
        return None
    kx, ky = metres_per_deg_lon(lat), METRES_PER_DEG_LAT
    local = shapely.transform(
        np.asarray(geometries), lambda xy: (xy - np.array([lon, lat])) * np.array([kx, ky])
    )
    return float(shapely.distance(local, shapely.Point(0, 0)).min())
