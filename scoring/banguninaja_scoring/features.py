import logging

from . import config
from .layers import (
    ConstructionCost,
    Elevation,
    LandValue,
    LineLayer,
    Population,
    Raster,
    load_pois,
    load_transit,
)

log = logging.getLogger(__name__)


class FeatureExtractor:
    def __init__(self):
        raw = config.data_path
        log.info("loading GIS layers from %s", config.DATA_DIR)
        self.hazard_multi = Raster(raw("inarisk_wcs", "inarisk_multi_indonesia.tif"))
        self.hazards = {
            name: Raster(raw("inarisk_wcs", f"inarisk_{name}_indonesia.tif")) for name in config.HAZARDS
        }
        self.elevation = Elevation(raw("dem_copernicus30m_indonesia_2026.tif"))
        self.population = Population(raw("worldpop_idn_2020.tif"))
        self.roads = LineLayer(raw("jaringan_transportasi_indonesia_2026.gpkg"), "jalan")
        self.transit = load_transit(raw("jaringan_transportasi_indonesia_2026.gpkg"))
        self.pois = load_pois(raw("poi_kompetitor_indonesia_2026.gpkg"))
        self.cost = ConstructionCost(raw("gadm41_indonesia.gpkg"), raw("bps_ikk_indonesia_2025.csv"))
        self.land = LandValue(raw("znt_atrbpn_indonesia.gpkg"))
        log.info("GIS layers ready (ZNT local=%s)", self.land.local_available)

    def extract(self, lon: float, lat: float, with_land_price: bool = True) -> dict:
        features = {"lon": lon, "lat": lat}
        features.update(self._physical(lon, lat))
        features.update(self._access(lon, lat))
        features.update(self._people_and_market(lon, lat))
        features.update(self._finance(lon, lat, with_land_price))
        return features

    def _physical(self, lon: float, lat: float) -> dict:
        multi, _ = self.hazard_multi.mean_and_max(lon, lat)
        result = {"hazard_multi": multi, "slope_deg": self.elevation.slope_deg(lon, lat)}
        for name, raster in self.hazards.items():
            _, peak = raster.mean_and_max(lon, lat)
            result[f"hazard_{name}"] = peak
        return result

    def _access(self, lon: float, lat: float) -> dict:
        return {
            "road_dist_m": self.roads.nearest_m(lon, lat),
            "transit_dist_m": self.transit.nearest_m(lon, lat),
            "hospital_dist_m": self.pois.hospitals.nearest_m(lon, lat),
        }

    def _people_and_market(self, lon: float, lat: float) -> dict:
        result = {
            "pop_2km": self.population.sum_within(lon, lat, config.POPULATION_RADIUS_M),
            "pop_5km": self.population.sum_within(lon, lat, config.TARGET_RADIUS_M),
            "amenities_2km": self.pois.daily_amenities.count_within(lon, lat, config.AMENITY_RADIUS_M),
        }
        for profile, index in self.pois.competitors.items():
            result[f"competitors_3km_{profile}"] = index.count_within(lon, lat, config.COMPETITOR_RADIUS_M)
            result[f"competitors_5km_{profile}"] = index.count_within(lon, lat, config.TARGET_RADIUS_M)
        return result

    def _finance(self, lon: float, lat: float, with_land_price: bool) -> dict:
        ikk, region = self.cost.lookup(lon, lat)
        result = {"ikk": ikk, "region": region, "land_price": None, "land_source": None}
        if not with_land_price:
            return result
        try:
            result["land_price"], result["land_source"] = self.land.price_at(lon, lat)
        except Exception as error:
            log.warning("land value lookup failed at %.5f,%.5f: %s", lat, lon, error)
        return result
