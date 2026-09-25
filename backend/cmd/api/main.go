package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/asset"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/auth"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/billing"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/finance"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/hr"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/inventory"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/location"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/master"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/procurement"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/project"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/regulation"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/reporting"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/sales"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/scoring"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/cookie"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/database"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/middleware"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/token"
	"github.com/jojondrw/BanguninAjaApp/backend/internal/news"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 15 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := database.Close(db); closeErr != nil {
			slog.Error("close database failed", slog.String("error", closeErr.Error()))
		}
	}()

	tokens := token.NewManager(cfg.Token)
	refreshCookie := cookie.NewRefreshWriter(cfg.Cookie)

	server := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           buildRouter(cfg, db, tokens, refreshCookie),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		slog.Info("server listening", slog.String("port", cfg.App.Port), slog.String("environment", cfg.App.Environment))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listen failed", slog.String("error", err.Error()))
		}
	}()

	waitForShutdownSignal()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return server.Shutdown(ctx)
}

func buildRouter(cfg config.Config, db *gorm.DB, tokens *token.Manager, refreshCookie cookie.RefreshWriter) *gin.Engine {
	if cfg.App.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.CORS(cfg.CORS), middleware.ErrorHandler())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")
	auth.NewModule(db, tokens, refreshCookie).RegisterRoutes(api)
	master.NewModule(db, tokens).RegisterRoutes(api)
	project.NewModule(db, tokens).RegisterRoutes(api)
	inventory.NewModule(db, tokens).RegisterRoutes(api)
	procurement.NewModule(db, tokens).RegisterRoutes(api)
	asset.NewModule(db, tokens).RegisterRoutes(api)
	sales.NewModule(db, tokens).RegisterRoutes(api)
	finance.NewModule(db, tokens).RegisterRoutes(api)
	billing.NewModule(db, tokens).RegisterRoutes(api)
	hr.NewModule(db, tokens).RegisterRoutes(api)
	scoring.NewModule(db, tokens).RegisterRoutes(api)
	location.NewModule(db, tokens).RegisterRoutes(api)
	reporting.NewModule(db, tokens).RegisterRoutes(api)
	regulation.NewModule(db, tokens).RegisterRoutes(api)
	news.NewModule(db, tokens).RegisterRoutes(api)

	return router
}

func waitForShutdownSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")
}
