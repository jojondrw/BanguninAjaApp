package cookie

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

const clearMaxAge = -1

type RefreshWriter struct {
	settings config.Cookie
}

func NewRefreshWriter(settings config.Cookie) RefreshWriter {
	return RefreshWriter{settings: settings}
}

func (w RefreshWriter) Set(c *gin.Context, value string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	w.write(c, value, maxAge)
}

func (w RefreshWriter) Clear(c *gin.Context) {
	w.write(c, "", clearMaxAge)
}

func (w RefreshWriter) Read(c *gin.Context) string {
	value, err := c.Cookie(w.settings.Name)
	if err != nil {
		return ""
	}
	return value
}

func (w RefreshWriter) write(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(sameSiteMode(w.settings.SameSite))
	c.SetCookie(w.settings.Name, value, maxAge, w.settings.Path, w.settings.Domain, w.settings.Secure, true)
}

func sameSiteMode(name string) http.SameSite {
	switch strings.ToLower(name) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
