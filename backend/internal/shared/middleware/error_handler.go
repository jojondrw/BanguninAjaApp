package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/apperror"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/httpresponse"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		failure := apperror.From(c.Errors.Last().Err)
		if failure.Status >= http.StatusInternalServerError {
			slog.Error("request failed",
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.String("error", failure.Error()),
			)
		}

		httpresponse.Failure(c, failure.Status, failure.Code, failure.Message)
	}
}
