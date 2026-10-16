package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

func testSettings(sameSite string) config.Cookie {
	return config.Cookie{Name: "refresh_token", Path: "/api/auth", Secure: true, SameSite: sameSite}
}

func newTestContext(request *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = request
	return ctx, recorder
}

func onlyCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	return cookies[0]
}

func TestSetWritesAnHttpOnlySecureCookieThatExpiresWithTheToken(t *testing.T) {
	ctx, recorder := newTestContext(httptest.NewRequest(http.MethodPost, "/api/auth/login", nil))

	NewRefreshWriter(testSettings("strict")).Set(ctx, "nilai-refresh", time.Now().Add(time.Hour))

	written := onlyCookie(t, recorder)
	if written.Name != "refresh_token" || written.Value != "nilai-refresh" || written.Path != "/api/auth" {
		t.Fatalf("cookie = %s=%s path %s", written.Name, written.Value, written.Path)
	}
	if !written.HttpOnly {
		t.Fatal("refresh cookie is readable from JavaScript")
	}
	if !written.Secure {
		t.Fatal("refresh cookie can be sent over plain HTTP")
	}
	if written.SameSite != http.SameSiteStrictMode {
		t.Fatalf("SameSite = %v, want Strict", written.SameSite)
	}
	if written.MaxAge < 3590 || written.MaxAge > 3600 {
		t.Fatalf("Max-Age = %d, want about one hour", written.MaxAge)
	}
}

func TestClearExpiresTheCookieImmediately(t *testing.T) {
	ctx, recorder := newTestContext(httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil))

	NewRefreshWriter(testSettings("lax")).Clear(ctx)

	written := onlyCookie(t, recorder)
	if written.Value != "" {
		t.Fatalf("cleared cookie still holds %q", written.Value)
	}
	if written.MaxAge >= 0 {
		t.Fatalf("MaxAge = %d, want a negative value that deletes the cookie", written.MaxAge)
	}
	if !written.HttpOnly || written.Path != "/api/auth" {
		t.Fatal("clearing must use the same path and flags as setting, or the browser keeps the old cookie")
	}
}

func TestReadReturnsTheRefreshCookieOrEmpty(t *testing.T) {
	writer := NewRefreshWriter(testSettings("lax"))

	withCookie := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	withCookie.AddCookie(&http.Cookie{Name: "refresh_token", Value: "dari-browser"})
	ctx, _ := newTestContext(withCookie)
	if got := writer.Read(ctx); got != "dari-browser" {
		t.Fatalf("Read = %q, want dari-browser", got)
	}

	withoutCookie := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	withoutCookie.AddCookie(&http.Cookie{Name: "cookie_lain", Value: "abaikan"})
	ctx, _ = newTestContext(withoutCookie)
	if got := writer.Read(ctx); got != "" {
		t.Fatalf("Read = %q without a refresh cookie, want empty", got)
	}
}

func TestSameSiteSettingMapsToTheBrowserMode(t *testing.T) {
	cases := map[string]http.SameSite{
		"strict":  http.SameSiteStrictMode,
		"Strict":  http.SameSiteStrictMode,
		"none":    http.SameSiteNoneMode,
		"lax":     http.SameSiteLaxMode,
		"":        http.SameSiteLaxMode,
		"unknown": http.SameSiteLaxMode,
	}

	for setting, want := range cases {
		if got := sameSiteMode(setting); got != want {
			t.Fatalf("sameSiteMode(%q) = %v, want %v", setting, got, want)
		}
	}
}
