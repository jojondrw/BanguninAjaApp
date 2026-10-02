import os
from pathlib import Path

PACKAGE_DIR = Path(__file__).resolve().parent
MODEL_PATH = Path(os.environ.get("SCORING_MODEL", PACKAGE_DIR.parent / "model" / "model.json"))
DATA_DIR = Path(os.environ.get("DATA_DIR", ""))

PROFILES = ("housing", "hospital", "mall", "entertainment")
DIMENSIONS = (
    "fisik_lingkungan",
    "infrastruktur",
    "demografi_sosial",
    "pasar_kompetisi",
    "finansial_proyek",
)
HAZARDS = ("banjir", "gempabumi", "longsor", "tsunami", "likuefaksi", "gunungapi")
HAZARD_LABELS = {
    "banjir": "banjir",
    "gempabumi": "gempa bumi",
    "longsor": "tanah longsor",
    "tsunami": "tsunami",
    "likuefaksi": "likuefaksi",
    "gunungapi": "letusan gunung api",
}

POPULATION_RADIUS_M = 2000
COMPETITOR_RADIUS_M = 3000
AMENITY_RADIUS_M = 2000
TARGET_RADIUS_M = 5000

ZNT_MAX_VALID = 500_000_000
ZNT_LIVE_URL = "https://atlas.atrbpn.go.id/geoserver/ows"


def data_path(*parts: str) -> Path:
    return DATA_DIR.joinpath(*parts)
