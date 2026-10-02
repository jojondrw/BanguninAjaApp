import zlib

import numpy as np
import pytest
import rasterio
from fastapi.testclient import TestClient
from rasterio.transform import from_origin

from banguninaja_scoring import config, tiles
from banguninaja_scoring.app import create_app
from test_scoring import FakeExtractor, linear_model


def test_tile_bounds_cover_the_world_at_zoom_zero():
    west, south, east, north = tiles.tile_bounds(0, 0, 0)
    assert west == pytest.approx(-tiles.WEB_MERCATOR_HALF)
    assert east == pytest.approx(tiles.WEB_MERCATOR_HALF)
    assert north == pytest.approx(tiles.WEB_MERCATOR_HALF)
    assert south == pytest.approx(-tiles.WEB_MERCATOR_HALF)


def test_png_is_well_formed_rgba():
    rgba = np.zeros((4, 3, 4), dtype="uint8")
    rgba[..., 0] = 255
    png = tiles.encode_png(rgba)
    assert png.startswith(b"\x89PNG\r\n\x1a\n")
    width, height = int.from_bytes(png[16:20], "big"), int.from_bytes(png[20:24], "big")
    assert (width, height) == (3, 4)
    idat_start = png.index(b"IDAT") + 4
    idat_length = int.from_bytes(png[idat_start - 8 : idat_start - 4], "big")
    raw = zlib.decompress(png[idat_start : idat_start + idat_length])
    assert len(raw) == 4 * (1 + 3 * 4)


def test_ramp_keeps_invalid_cells_transparent():
    values = np.array([[0.0, 0.5, 1.0]])
    valid = np.array([[False, True, True]])
    rgba = tiles.HAZARD_RAMP.apply(values, valid)
    assert rgba[0, 0, 3] == 0
    assert rgba[0, 1, 3] > 0
    assert rgba[0, 2, 0] > rgba[0, 1, 0] or rgba[0, 2, 1] < rgba[0, 1, 1]


def test_valid_tile_range():
    assert tiles.valid_tile(3, 7, 7)
    assert not tiles.valid_tile(3, 8, 0)
    assert not tiles.valid_tile(-1, 0, 0)


@pytest.fixture
def bundle(tmp_path, monkeypatch):
    data = np.full((200, 200), 200, dtype="uint8")
    profile = dict(driver="GTiff", width=200, height=200, count=1, dtype="uint8", crs="EPSG:4326",
                   transform=from_origin(106.0, -6.0, 0.005, 0.005), nodata=255)
    with rasterio.open(tmp_path / "hazard_multi.tif", "w", **profile) as dataset:
        dataset.write(data, 1)
    monkeypatch.setattr(config, "BUNDLE_DIR", tmp_path)
    tiles.render.cache_clear()
    yield tmp_path
    tiles.render.cache_clear()


def test_render_paints_covered_tile_and_skips_low_zoom(bundle):
    covered = tiles.render("bahaya", 9, 406, 265)
    assert covered != tiles.EMPTY_TILE
    assert tiles.render("bahaya", 3, 6, 4) == tiles.EMPTY_TILE
    assert tiles.render("bahaya", 9, 0, 0) == tiles.EMPTY_TILE


def test_tile_endpoint(bundle):
    with TestClient(create_app(FakeExtractor(), linear_model())) as client:
        response = client.get("/tiles/bahaya/9/406/265.png")
        assert response.status_code == 200
        assert response.headers["content-type"] == "image/png"
        assert "max-age" in response.headers["cache-control"]
        assert client.get("/tiles/rahasia/9/406/265.png").status_code == 404
        assert client.get("/tiles/bahaya/2/9/0.png").status_code == 400
