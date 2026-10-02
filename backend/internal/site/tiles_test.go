package site

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func tileRouter(proxy *TileProxy) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tiles/:layer/:z/:x/:y", proxy.Tile)
	return router
}

func serveTile(router *gin.Engine, path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestTileProxyForwardsPNG(t *testing.T) {
	var requested string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		if r.URL.Path == "/tiles/rahasia/9/406/265.png" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write([]byte("png-bytes"))
	}))
	defer upstream.Close()

	router := tileRouter(NewTileProxy(upstream.URL+"/", time.Second))

	recorder := serveTile(router, "/tiles/bahaya/9/406/265")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if requested != "/tiles/bahaya/9/406/265.png" {
		t.Fatalf("upstream path = %q", requested)
	}
	if recorder.Body.String() != "png-bytes" || recorder.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("unexpected body or content type: %q %q", recorder.Body.String(), recorder.Header().Get("Content-Type"))
	}
	if recorder.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Fatalf("cache header not forwarded")
	}

	if code := serveTile(router, "/tiles/rahasia/9/406/265").Code; code != http.StatusNotFound {
		t.Fatalf("unknown layer status = %d, want 404", code)
	}
}

func TestTileProxyRejectsInvalidTiles(t *testing.T) {
	router := tileRouter(NewTileProxy("http://127.0.0.1:1", time.Second))

	for _, path := range []string{
		"/tiles/bahaya/2/4/0",
		"/tiles/bahaya/-1/0/0",
		"/tiles/bahaya/23/0/0",
		"/tiles/bahaya/x/0/0",
		"/tiles/Bahaya!/3/0/0",
	} {
		if code := serveTile(router, path).Code; code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400", path, code)
		}
	}
}

func TestTileProxyWithoutServiceIsUnavailable(t *testing.T) {
	router := tileRouter(NewTileProxy("  ", time.Second))

	if code := serveTile(router, "/tiles/bahaya/9/406/265").Code; code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", code)
	}
}

func TestTileProxyReportsUnreachableService(t *testing.T) {
	router := tileRouter(NewTileProxy("http://127.0.0.1:1", 200*time.Millisecond))

	if code := serveTile(router, "/tiles/bahaya/9/406/265").Code; code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", code)
	}
}
