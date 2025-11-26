package service

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"petrochemical-data-platform/internal/domain"
	"petrochemical-data-platform/internal/repository"

	"go.uber.org/zap"
)

const assetListCacheTTL = 15 * time.Minute

var (
	defaultSeedRangeEnd = time.Date(2024, time.December, 1, 0, 0, 0, 0, time.UTC)
)

// AssetService обрабатывает бизнес-логику, связанную с активами
type AssetService struct {
	repo   *repository.PostgresRepository
	cache  *repository.RedisRepository
	logger *zap.Logger
}

// NewAssetService создает новый сервис активов
func NewAssetService(repo *repository.PostgresRepository, cache *repository.RedisRepository, logger *zap.Logger) *AssetService {
	return &AssetService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

// GetAssets получает все активы с кэшированием
func (s *AssetService) GetAssets(ctx context.Context) ([]domain.Asset, error) {
	if s.cache != nil {
		cachedAssets, err := s.cache.GetCachedAssets(ctx)
		if err != nil {
			s.logger.Warn("Failed to read assets from cache", zap.Error(err))
		} else if cachedAssets != nil {
			return cachedAssets, nil
		}
	}

	dbAssets, err := s.repo.GetAssets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get assets from database: %w", err)
	}

	assets := make([]domain.Asset, 0, len(dbAssets))
	for _, dbAsset := range dbAssets {
		assets = append(assets, domain.Asset{
			ID:        dbAsset.ID,
			Name:      dbAsset.Name,
			Type:      dbAsset.Type,
			Location:  dbAsset.Location,
			CreatedAt: dbAsset.CreatedAt,
			UpdatedAt: dbAsset.UpdatedAt,
		})
	}

	if s.cache != nil {
		if err := s.cache.CacheAssets(ctx, assets, assetListCacheTTL); err != nil {
			s.logger.Warn("Failed to cache assets list", zap.Error(err))
		}
	}

	return assets, nil
}

// CreateAsset создает новый актив
func (s *AssetService) CreateAsset(ctx context.Context, asset domain.Asset) error {
	repoAsset := repository.Asset{
		ID:        asset.ID,
		Name:      asset.Name,
		Type:      asset.Type,
		Location:  asset.Location,
		CreatedAt: asset.CreatedAt,
		UpdatedAt: asset.UpdatedAt,
	}

	if err := s.repo.SaveAsset(ctx, repoAsset); err != nil {
		return fmt.Errorf("failed to save asset: %w", err)
	}

	// Cache the asset
	if s.cache != nil {
		if err := s.cache.CacheAsset(ctx, repoAsset, time.Hour); err != nil {
			s.logger.Warn("Failed to cache asset", zap.Error(err), zap.String("asset_id", asset.ID))
		}
		if err := s.cache.InvalidateAssetsCache(ctx); err != nil {
			s.logger.Warn("Failed to invalidate assets cache", zap.Error(err))
		}
	}

	return nil
}

// TelemetryService обрабатывает операции с данными телеметрии
type TelemetryService struct {
	repo   *repository.ClickHouseRepository
	logger *zap.Logger
}

// NewTelemetryService создает новый сервис телеметрии
func NewTelemetryService(repo *repository.ClickHouseRepository, logger *zap.Logger) *TelemetryService {
	return &TelemetryService{
		repo:   repo,
		logger: logger,
	}
}

// GetTelemetry получает данные производства/продаж для компании
func (s *TelemetryService) GetTelemetry(ctx context.Context, companyID string, start, end time.Time) ([]domain.TelemetryData, error) {
	if err := s.ensureBaseTelemetry(ctx, companyID); err != nil {
		return nil, err
	}

	data, err := s.repo.GetTelemetryData(ctx, companyID, start, end)
	if err != nil {
		return nil, err
	}

	return data, nil
}

const (
	minSyntheticMonths = 24
	maxSyntheticMonths = 120
)

// ensureBaseTelemetry гарантирует, что для компании есть базовый набор данных в БД
func (s *TelemetryService) ensureBaseTelemetry(ctx context.Context, companyID string) error {
	hasData, err := s.repo.HasTelemetryData(ctx, companyID)
	if err != nil {
		return err
	}
	if hasData {
		return nil
	}

	start, end := defaultSeedRange()
	return s.SeedSyntheticTelemetry(ctx, companyID, start, end)
}

// SeedSyntheticTelemetry generates deterministic synthetic telemetry for the provided time range.
func (s *TelemetryService) SeedSyntheticTelemetry(ctx context.Context, companyID string, start, end time.Time) error {
	nStart, nEnd := normalizeTimeRange(start, end)
	generated := generateSyntheticTelemetry(companyID, nStart, nEnd)
	if len(generated) == 0 {
		return fmt.Errorf("failed to generate telemetry data for %s", companyID)
	}

	if err := s.repo.SaveTelemetryBatch(ctx, generated); err != nil {
		return fmt.Errorf("failed to persist generated telemetry: %w", err)
	}

	s.logger.Info("Synthetic telemetry generated",
		zap.String("company_id", companyID),
		zap.Time("start", nStart),
		zap.Time("end", nEnd),
		zap.Int("records", len(generated)))

	return nil
}

func normalizeTimeRange(start, end time.Time) (time.Time, time.Time) {
	if end.IsZero() {
		end = defaultSeedRangeEnd
	} else {
		end = end.UTC()
	}

	if start.IsZero() || start.After(end) {
		start = end.AddDate(0, -minSyntheticMonths+1, 0)
	} else {
		start = start.UTC()
	}

	months := monthsBetween(start, end)
	if months < minSyntheticMonths {
		start = end.AddDate(0, -minSyntheticMonths+1, 0)
	}
	if months > maxSyntheticMonths {
		start = end.AddDate(0, -maxSyntheticMonths+1, 0)
	}

	start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	return start, end
}

func monthsBetween(start, end time.Time) int {
	yearDiff := end.Year() - start.Year()
	monthDiff := int(end.Month()) - int(start.Month())
	totalMonths := yearDiff*12 + monthDiff + 1
	if totalMonths < 0 {
		return 0
	}
	return totalMonths
}

func generateSyntheticTelemetry(companyID string, start, end time.Time) []domain.TelemetryData {
	seed := deterministicSeed(companyID)
	rnd := rand.New(rand.NewSource(seed))

	products := productsForCompany(companyID)
	if len(products) == 0 {
		products = []string{"Полипропилен"}
	}

	results := make([]domain.TelemetryData, 0)
	current := start

	basePrice := 45000 + rnd.Float64()*30000
	unit := "₽/т"

	for !current.After(end) {
		monthProgress := float64(monthsBetween(start, current)) / float64(max(1, monthsBetween(start, end)))
		trend := 1 + (rnd.Float64()-0.5)*0.1
		for _, product := range products {
			price := basePrice * (0.85 + rnd.Float64()*0.3) * (1 + trend*0.05*monthProgress)
			price = math.Round(price*100) / 100
			results = append(results, domain.TelemetryData{
				CompanyID:   companyID,
				ProductName: product,
				Value:       price,
				Unit:        unit,
				Timestamp:   current,
				Quality:     uint16(80 + rnd.Intn(20)),
			})
		}
		current = current.AddDate(0, 1, 0)
	}

	return results
}

func deterministicSeed(companyID string) int64 {
	hash := sha1.Sum([]byte(companyID))
	return int64(binary.BigEndian.Uint64(hash[:8]))
}

func productsForCompany(companyID string) []string {
	if products, ok := companyProducts[companyID]; ok {
		return products
	}
	return []string{"Полипропилен", "Полиэтилен"}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var companyProducts = map[string][]string{
	"SIBUR_TOBOLSK_POLYMER": {"Полипропилен", "Полиэтилен"},
	"ZAPSIBNEFTEKHIM":       {"Полипропилен", "Полиэтилен"},
	"ROSNEFT_OMSK_REFINERY": {"Бензин", "Дизель"},
	"GAZPROMNEFT_MOSCOW_REFINERY": {
		"Автобензин", "Авиакеросин",
	},
	"LUKOIL_VOLGOGRAD_REFINERY": {"Автобензин", "Битум"},
	"TATNEFT_ROMASHKINO_FIELD":  {"Нефть сырая"},
	"NOVATEK_YAMAL_LNG":         {"СПГ", "Газовый конденсат"},
}

// DefaultCompanyIDs exposes the built-in companies that have detailed product maps.
func DefaultCompanyIDs() []string {
	ids := make([]string, 0, len(companyProducts))
	for id := range companyProducts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// DefaultTelemetryWindow возвращает временной диапазон для дефолтных запросов API
func DefaultTelemetryWindow() (time.Time, time.Time) {
	end := defaultSeedRangeEnd
	start := end.AddDate(0, -36+1, 0)
	return start, end
}

func defaultSeedRange() (time.Time, time.Time) {
	end := defaultSeedRangeEnd
	start := end.AddDate(0, -maxSyntheticMonths+1, 0)
	return start, end
}

// ControlService обрабатывает команды управления
type ControlService struct {
	logger *zap.Logger
	// Would include MQTT client for sending commands
}

// NewControlService создает новый сервис управления
func NewControlService(logger *zap.Logger) *ControlService {
	return &ControlService{
		logger: logger,
	}
}

// SendControlCommand отправляет команду управления
func (s *ControlService) SendControlCommand(ctx context.Context, cmd domain.ControlCommand) error {
	s.logger.Info("Sending control command",
		zap.String("command_id", cmd.ID),
		zap.String("equipment_id", cmd.EquipmentID),
		zap.String("command", cmd.Command))

	// Placeholder - would send via MQTT or other protocol
	// In real implementation, this would publish to MQTT topic

	return nil
}
