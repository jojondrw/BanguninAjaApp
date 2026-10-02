import logging

from . import config
from .layers import (
    ConstructionCost,
    Elevation,
    LandPrice,
    Population,
    Raster,
    load_pois,
    load_transit,
)

log = logging.getLogger(__name__)


class FeatureExtractor:
    def __init__(self):
        bundle = config.bundle_path
        log.info("loading data bundle from %s", config.BUNDLE_DIR)
        hazard_scale = 1 / config.HAZARD_SCALE
        self.hazard_multi = Raster(bundle("hazard_multi.tif"), scale=hazard_scale)
        self.hazards = {name: Raster(bundle(f"hazard_{name}.tif"), scale=hazard_scale) for name in config.HAZARDS}
        self.elevation = Elevation(bundle("dem.tif"))
        self.population = Population(bundle("population.tif"))
        self.roads = Raster(bundle("road_distance.tif"))
        self.land = LandPrice(bundle("land_price.tif"))
        self.transit = load_transit(bundle("transit.gpkg"))
        self.pois = load_pois(bundle("poi.gpkg"))
        self.cost = ConstructionCost(bundle("regions.gpkg"), bundle("ikk.csv"))
        log.info("data bundle ready")

    def extract(self, lon: float, lat: float) -> dict:
        features = {"lon": lon, "lat": lat}
        features.update(self._physical(lon, lat))
        features.update(self._access(lon, lat))
        features.update(self._people_and_market(lon, lat))
        features.update(self._finance(lon, lat))
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
            "road_dist_m": self.roads.value_at(lon, lat),
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

    def _finance(self, lon: float, lat: float) -> dict:
        ikk, region = self.cost.lookup(lon, lat)
        price = self.land.price_at(lon, lat)
        return {"ikk": ikk, "region": region, "land_price": price, "land_source": "ZNT" if price else None}
