"""XYZ map tiles (256 px, Web Mercator) rendered from the scoring bundle rasters."""

import struct
import zlib
from dataclasses import dataclass
from functools import lru_cache

import numpy as np
import rasterio
from rasterio.enums import Resampling
from rasterio.transform import from_bounds
from rasterio.warp import reproject

from . import config

TILE_SIZE = 256
WEB_MERCATOR_HALF = 20037508.342789244
CELL_KM2_POPULATION = 0.275 * 0.275
NODATA = -1.0


@dataclass(frozen=True)
class Ramp:
    stops: tuple
    colors: tuple
    log: bool = False

    def apply(self, values: np.ndarray, valid: np.ndarray) -> np.ndarray:
        data = np.where(valid, values, np.nan)
        positions = np.array(self.stops, dtype="float64")
        if self.log:
            data = np.log10(np.clip(data, positions[0], None))
            positions = np.log10(positions)
        rgba = np.zeros(values.shape + (4,), dtype="uint8")
        colors = np.array(self.colors, dtype="float64")
        for channel in range(4):
            mapped = np.interp(np.nan_to_num(data, nan=positions[0]), positions, colors[:, channel])
            rgba[..., channel] = np.where(valid, np.rint(mapped), 0).astype("uint8")
        return rgba


HAZARD_RAMP = Ramp(
    stops=(0.0, 0.15, 0.4, 0.66, 1.0),
    colors=((46, 157, 85, 0), (46, 157, 85, 90), (245, 197, 24, 140), (240, 138, 36, 170), (215, 38, 61, 190)),
)
POPULATION_RAMP = Ramp(
    stops=(300, 1000, 4000, 15000, 40000),
    colors=((237, 233, 254, 0), (196, 181, 253, 80), (139, 92, 246, 130), (91, 33, 182, 175), (46, 16, 101, 210)),
    log=True,
)
LAND_PRICE_RAMP = Ramp(
    stops=(100_000, 1_000_000, 5_000_000, 20_000_000, 60_000_000),
    colors=((253, 230, 138, 120), (245, 158, 11, 150), (217, 119, 6, 170), (180, 83, 9, 190), (124, 45, 18, 210)),
    log=True,
)


@dataclass(frozen=True)
class TileLayer:
    file: str
    ramp: Ramp
    min_zoom: int
    resampling: Resampling
    scale: float = 1.0
    per_km2: float | None = None


LAYERS = {
    "bahaya": TileLayer("hazard_multi.tif", HAZARD_RAMP, 7, Resampling.average, 1 / config.HAZARD_SCALE),
    "banjir": TileLayer("hazard_banjir.tif", HAZARD_RAMP, 7, Resampling.average, 1 / config.HAZARD_SCALE),
    "gempabumi": TileLayer("hazard_gempabumi.tif", HAZARD_RAMP, 7, Resampling.average, 1 / config.HAZARD_SCALE),
    "longsor": TileLayer("hazard_longsor.tif", HAZARD_RAMP, 7, Resampling.average, 1 / config.HAZARD_SCALE),
    "penduduk": TileLayer("population.tif", POPULATION_RAMP, 7, Resampling.average, per_km2=CELL_KM2_POPULATION),
    "harga_tanah": TileLayer("land_price.tif", LAND_PRICE_RAMP, 10, Resampling.nearest),
}


def tile_bounds(z: int, x: int, y: int) -> tuple[float, float, float, float]:
    size = 2 * WEB_MERCATOR_HALF / (2**z)
    west = -WEB_MERCATOR_HALF + x * size
    north = WEB_MERCATOR_HALF - y * size
    return west, north - size, west + size, north


def encode_png(rgba: np.ndarray) -> bytes:
    height, width, _ = rgba.shape
    raw = b"".join(b"\x00" + rgba[row].tobytes() for row in range(height))

    def chunk(tag: bytes, data: bytes) -> bytes:
        return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)

    header = struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0)
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(b"IDAT", zlib.compress(raw, 6)) + chunk(b"IEND", b"")


EMPTY_TILE = encode_png(np.zeros((TILE_SIZE, TILE_SIZE, 4), dtype="uint8"))


def valid_tile(z: int, x: int, y: int) -> bool:
    return 0 <= z <= 22 and 0 <= x < 2**z and 0 <= y < 2**z


def read_tile(path, layer: TileLayer, z: int, x: int, y: int) -> tuple[np.ndarray, np.ndarray]:
    destination = np.full((TILE_SIZE, TILE_SIZE), NODATA, dtype="float32")
    with rasterio.open(path) as source:
        reproject(
            source=rasterio.band(source, 1),
            destination=destination,
            src_nodata=source.nodata,
            dst_transform=from_bounds(*tile_bounds(z, x, y), TILE_SIZE, TILE_SIZE),
            dst_crs="EPSG:3857",
            dst_nodata=NODATA,
            resampling=layer.resampling,
        )
    valid = (destination != NODATA) & np.isfinite(destination) & (destination > 0)
    values = destination.astype("float64") * layer.scale
    if layer.per_km2:
        values = values / layer.per_km2
    return values, valid


@lru_cache(maxsize=4096)
def render(name: str, z: int, x: int, y: int) -> bytes:
    layer = LAYERS[name]
    if z < layer.min_zoom:
        return EMPTY_TILE
    path = config.bundle_path(layer.file)
    if not path.exists():
        return EMPTY_TILE
    values, valid = read_tile(path, layer, z, x, y)
    if not valid.any():
        return EMPTY_TILE
    return encode_png(layer.ramp.apply(values, valid))
