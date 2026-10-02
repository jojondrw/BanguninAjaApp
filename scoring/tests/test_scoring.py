import numpy as np
import pytest
from fastapi.testclient import TestClient

from banguninaja_scoring import config
from banguninaja_scoring.app import create_app
from banguninaja_scoring.model import Model, risk_flags, score
from train import to_weights

FEATURES = {
    "hazard_multi": 0.2,
    "slope_deg": 2.0,
    "hazard_banjir": 0.8,
    "hazard_gempabumi": 0.5,
    "road_dist_m": 50.0,
    "transit_dist_m": 800.0,
    "hospital_dist_m": 1500.0,
    "pop_2km": 40000.0,
    "pop_5km": 200000.0,
    "amenities_2km": 30,
    "competitors_3km_hospital": 1,
    "competitors_3km_mall": 2,
    "competitors_3km_entertainment": 0,
    "competitors_3km_housing": 5,
    "land_price": 3_000_000.0,
    "ikk": 100.0,
    "region": "Kota Contoh",
}


def linear_model(weights=None):
    quantiles = {
        "hazard_multi": np.linspace(0, 1, 101),
        "slope_deg": np.linspace(0, 40, 101),
        "road_dist_m": np.linspace(0, 10_000, 101),
        "transit_dist_m": np.linspace(0, 50_000, 101),
        "hospital_dist_m": np.linspace(0, 50_000, 101),
        "pop_2km": np.linspace(0, 100_000, 101),
        "amenities_2km": np.linspace(0, 100, 101),
        "land_price": np.linspace(0, 10_000_000, 101),
        "ikk": np.linspace(80, 180, 101),
    }
    for profile in config.PROFILES:
        quantiles[f"competitors_3km_{profile}"] = np.linspace(0, 20, 101)
    equal = {code: 20 for code in config.DIMENSIONS}
    return Model(quantiles, weights or {profile: equal for profile in config.PROFILES}, {"test": True})


class FakeExtractor:
    def extract(self, lon, lat):
        return dict(FEATURES, lon=lon, lat=lat)


def test_percentile_handles_ties_and_bounds():
    model = Model({"x": np.array([0, 0, 0, 0, 10])}, {}, {})
    assert model.percentile("x", 0) == pytest.approx(40)
    assert model.percentile("x", -5) == 0
    assert model.percentile("x", 99) == 100
    assert model.percentile("x", None) is None


def test_score_shape_matches_go_contract():
    result = score(linear_model(), FEATURES, "hospital")
    assert isinstance(result["overall_score"], int)
    assert [d["dimension_code"] for d in result["dimension_scores"]] == list(config.DIMENSIONS)
    assert all(isinstance(d["value"], int) and 0 <= d["value"] <= 100 for d in result["dimension_scores"])
    assert all(d["explanation"] for d in result["dimension_scores"])


def test_overall_is_weighted_sum():
    weights = {profile: {"fisik_lingkungan": 100, "infrastruktur": 0, "demografi_sosial": 0, "pasar_kompetisi": 0, "finansial_proyek": 0} for profile in config.PROFILES}
    result = score(linear_model(weights), FEATURES, "mall")
    physical = next(d for d in result["dimension_scores"] if d["dimension_code"] == "fisik_lingkungan")
    assert abs(result["overall_score"] - physical["value"]) <= 1


def test_missing_values_fall_back_to_neutral():
    result = score(linear_model(), {}, "housing")
    assert all(d["value"] == 50 for d in result["dimension_scores"])
    assert any(flag["code"] == "di_luar_cakupan_data" for flag in result["risk_flags"])
    assert result["region"] is None


def test_region_is_reported_but_never_scored():
    with_region = score(linear_model(), FEATURES, "housing")
    without_region = score(linear_model(), dict(FEATURES, region=None), "housing")
    assert with_region["region"] == "Kota Contoh"
    assert without_region["region"] is None
    assert with_region["overall_score"] == without_region["overall_score"]
    assert [d["value"] for d in with_region["dimension_scores"]] == [d["value"] for d in without_region["dimension_scores"]]


def test_risk_flags_for_high_hazard():
    codes = {flag["code"]: flag["severity"] for flag in risk_flags(FEATURES)}
    assert codes["rawan_banjir"] == "high"
    assert codes["rawan_gempabumi"] == "medium"


def test_weights_have_floor_and_sum_to_100():
    weights = to_weights(np.array([0.0, 0.0, 3.0, 1.0, 0.0]))
    assert sum(weights.values()) == 100
    assert min(weights.values()) >= 5
    assert weights["demografi_sosial"] > weights["pasar_kompetisi"] > weights["infrastruktur"]


def test_api_contract():
    with TestClient(create_app(FakeExtractor(), linear_model())) as client:
        ok = client.post("/score", json={"latitude": -6.2, "longitude": 106.8, "building_profile_code": "mall"})
        assert ok.status_code == 200
        assert set(ok.json()) == {"overall_score", "dimension_scores", "risk_flags", "region"}
        assert ok.json()["region"] == "Kota Contoh"

        unknown = client.post("/score", json={"latitude": -6.2, "longitude": 106.8, "building_profile_code": "castle"})
        assert unknown.status_code == 400
        assert unknown.json()["error"]["code"] == "unknown_building_profile"

        invalid = client.post("/score", json={"latitude": 200, "longitude": 106.8, "building_profile_code": "mall"})
        assert invalid.status_code == 400
