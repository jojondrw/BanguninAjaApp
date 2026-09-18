package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

const preflightMaxAgeSeconds = "600"

var (
	allowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions}
	allowedHeaders = []string{"Authorization", "Content-Type", "Accept"}
)

func CORS(settings config.CORS) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" || !slices.Contains(settings.AllowedOrigins, origin) {
			c.Next()
			return
		}

		header := c.Writer.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Set("Access-Control-Allow-Credentials", "true")
		header.Add("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			header.Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			header.Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			header.Set("Access-Control-Max-Age", preflightMaxAgeSeconds)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
