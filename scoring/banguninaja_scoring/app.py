import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.concurrency import run_in_threadpool
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse, Response
from pydantic import BaseModel, Field

from . import config, tiles
from .model import Model, score

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")


class ScoreRequest(BaseModel):
    latitude: float = Field(ge=-90, le=90)
    longitude: float = Field(ge=-180, le=180)
    building_profile_code: str


def error(status: int, code: str, message: str) -> JSONResponse:
    return JSONResponse(status_code=status, content={"error": {"code": code, "message": message}})


def create_app(extractor=None, model=None) -> FastAPI:
    state = {}

    @asynccontextmanager
    async def lifespan(_: FastAPI):
        if extractor is None:
            from .features import FeatureExtractor

            state["extractor"] = FeatureExtractor()
        else:
            state["extractor"] = extractor
        state["model"] = model or Model.load()
        yield

    app = FastAPI(title="BanguninAja scoring", lifespan=lifespan)

    @app.exception_handler(RequestValidationError)
    async def invalid_request(_: Request, exc: RequestValidationError):
        return error(400, "invalid_request", str(exc.errors()[:1]))

    @app.get("/health")
    def health():
        return {"status": "ok", "model": state["model"].meta}

    @app.post("/score")
    async def score_site(request: ScoreRequest):
        if request.building_profile_code not in config.PROFILES:
            return error(400, "unknown_building_profile", f"Profil {request.building_profile_code!r} tidak dikenal")
        features = await run_in_threadpool(state["extractor"].extract, request.longitude, request.latitude)
        return score(state["model"], features, request.building_profile_code)

    @app.get("/tiles/{layer}/{z}/{x}/{y}.png")
    async def map_tile(layer: str, z: int, x: int, y: int):
        if layer not in tiles.LAYERS:
            return error(404, "unknown_layer", f"Lapisan {layer!r} tidak dikenal")
        if not tiles.valid_tile(z, x, y):
            return error(400, "invalid_tile", "Koordinat tile di luar jangkauan")
        content = await run_in_threadpool(tiles.render, layer, z, x, y)
        return Response(content, media_type="image/png", headers={"Cache-Control": "public, max-age=86400"})

    return app


app = create_app()
