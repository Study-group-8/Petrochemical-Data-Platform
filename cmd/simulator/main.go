package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/app/bootstrap"
	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/service"

	"go.uber.org/zap"
)

func main() {
	var (
		configPath = flag.String("config", "", "Путь до config.yaml")
		companies  = flag.String("companies", "", "CompanyID через запятую (по умолчанию будут использованы встроенные)")
		months     = flag.Int("months", 60, "Длина исторического ряда в месяцах")
		delay      = flag.Duration("delay", 500*time.Millisecond, "Пауза между компаниями, чтобы не перегружать ClickHouse")
	)
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() //nolint:errcheck

	cfg, err := bootstrap.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	clickRepo, err := bootstrap.InitClickHouseRepository(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to init ClickHouse repository", zap.Error(err))
	}
	defer clickRepo.Close()

	telemetrySvc := service.NewTelemetryService(clickRepo, logger)
	targetCompanies := selectCompanyIDs(*companies)
	if len(targetCompanies) == 0 {
		targetCompanies = service.DefaultCompanyIDs()
	}

	end := time.Now().UTC()
	start := end.AddDate(0, -*months, 0)

	logger.Info("Starting telemetry simulation",
		zap.Int("companies", len(targetCompanies)),
		zap.Int("months", *months),
		zap.Time("start", start),
		zap.Time("end", end))

	ctx := context.Background()
	for _, companyID := range targetCompanies {
		if err := telemetrySvc.SeedSyntheticTelemetry(ctx, companyID, start, end); err != nil {
			logger.Error("Failed to seed telemetry", zap.String("company_id", companyID), zap.Error(err))
		} else {
			logger.Info("Telemetry seeded", zap.String("company_id", companyID))
		}

		if *delay > 0 {
			time.Sleep(*delay)
		}
	}

	logger.Info("Simulation completed")
}

func selectCompanyIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			ids = append(ids, strings.ToUpper(trimmed))
		}
	}
	return ids
}
