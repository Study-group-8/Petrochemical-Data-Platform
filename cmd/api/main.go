package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"petrochemical-data-platform/internal/app/bootstrap"
	"petrochemical-data-platform/internal/handler"
	"petrochemical-data-platform/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const defaultServerPort = "8080"

func main() {
	var (
		configPathFlag = flag.String("config", "", "Путь до config.yaml")
		envFlag        = flag.String("env", "", "Режим работы (dev/prod)")
	)
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to init zap logger: %v", err))
	}
	defer logger.Sync() //nolint:errcheck

	cfg, err := bootstrap.LoadConfig(*configPathFlag)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	if strings.EqualFold(*envFlag, "dev") || strings.EqualFold(os.Getenv("APP_ENV"), "dev") {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	pgRepo, err := bootstrap.InitPostgresRepository(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to init PostgreSQL repository", zap.Error(err))
	}
	defer pgRepo.Close()

	clickRepo, err := bootstrap.InitClickHouseRepository(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to init ClickHouse repository", zap.Error(err))
	}
	defer clickRepo.Close()

	redisRepo, err := bootstrap.InitRedisRepository(cfg, logger)
	if err != nil {
		logger.Warn("Redis cache disabled", zap.Error(err))
	}
	if redisRepo != nil {
		defer redisRepo.Close()
	}

	assetSvc := service.NewAssetService(pgRepo, redisRepo, logger)
	telemetrySvc := service.NewTelemetryService(clickRepo, logger)
	controlSvc := service.NewControlService(logger)

	httpHandler := handler.NewHandler(assetSvc, telemetrySvc, controlSvc, logger)
	router := buildRouter(logger, httpHandler)

	port := strings.TrimSpace(cfg.Server.Port)
	if port == "" {
		port = defaultServerPort
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("API server starting", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server crashed", zap.Error(err))
		}
	}()

	waitForShutdown(logger, server)
}

func buildRouter(logger *zap.Logger, httpHandler *handler.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	handler.SetupRoutes(router, httpHandler)
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

func waitForShutdown(logger *zap.Logger, server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped")
}
