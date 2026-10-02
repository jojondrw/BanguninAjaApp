package site

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

const maxTileZoom = 22

var tileLayerPattern = regexp.MustCompile(`^[a-z_]{2,32}$`)

type TileProxy struct {
	baseURL string
	client  *http.Client
}

func NewTileProxy(baseURL string, timeout time.Duration) *TileProxy {
	return &TileProxy{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *TileProxy) Tile(ctx *gin.Context) {
	if p.baseURL == "" {
		httpresponse.Failure(ctx, http.StatusServiceUnavailable, "tiles_unavailable", "Layanan peta belum tersambung")
		return
	}

	layer := ctx.Param("layer")
	z, x, y, ok := parseTile(ctx.Param("z"), ctx.Param("x"), ctx.Param("y"))
	if !tileLayerPattern.MatchString(layer) || !ok {
		httpresponse.Failure(ctx, http.StatusBadRequest, "invalid_tile", "Alamat tile tidak valid")
		return
	}

	url := fmt.Sprintf("%s/tiles/%s/%d/%d/%d.png", p.baseURL, layer, z, x, y)
	request, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodGet, url, nil)
	if err != nil {
		httpresponse.Failure(ctx, http.StatusInternalServerError, "internal_error", "Tile gagal diminta")
		return
	}

	response, err := p.client.Do(request)
	if err != nil {
		httpresponse.Failure(ctx, http.StatusBadGateway, "tiles_unreachable", "Layanan peta tidak bisa dihubungi")
		return
	}
	defer func() { _ = response.Body.Close() }()

	switch response.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		httpresponse.Failure(ctx, http.StatusNotFound, "unknown_layer", "Lapisan peta tidak dikenal")
		return
	case http.StatusBadRequest:
		httpresponse.Failure(ctx, http.StatusBadRequest, "invalid_tile", "Alamat tile tidak valid")
		return
	default:
		httpresponse.Failure(ctx, http.StatusBadGateway, "tiles_failed", "Layanan peta gagal membuat tile")
		return
	}

	if cacheControl := response.Header.Get("Cache-Control"); cacheControl != "" {
		ctx.Header("Cache-Control", cacheControl)
	}
	ctx.DataFromReader(http.StatusOK, response.ContentLength, "image/png", response.Body, nil)
}

func parseTile(rawZ, rawX, rawY string) (int, int, int, bool) {
	z, errZ := strconv.Atoi(rawZ)
	x, errX := strconv.Atoi(rawX)
	y, errY := strconv.Atoi(rawY)
	if errZ != nil || errX != nil || errY != nil || z < 0 || z > maxTileZoom {
		return 0, 0, 0, false
	}
	limit := 1 << z
	if x < 0 || y < 0 || x >= limit || y >= limit {
		return 0, 0, 0, false
	}
	return z, x, y, true
}
