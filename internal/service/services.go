package service

import (
	"context"
	"fmt"
	"time"

	"petrochemical-data-platform/internal/domain"
	"petrochemical-data-platform/internal/repository"

	"go.uber.org/zap"
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
	// Try cache first
	// Note: This is simplified - in production you'd cache all assets or use a list key
	assets := []domain.Asset{} // Would need to implement proper caching strategy

	// If not in cache, get from database
	if len(assets) == 0 {
		dbAssets, err := s.repo.GetAssets(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get assets from database: %w", err)
		}

		// Convert repository assets to domain assets
		for _, dbAsset := range dbAssets {
			asset := domain.Asset{
				ID:        dbAsset.ID,
				Name:      dbAsset.Name,
				Type:      dbAsset.Type,
				Location:  dbAsset.Location,
				CreatedAt: dbAsset.CreatedAt,
				UpdatedAt: dbAsset.UpdatedAt,
			}
			assets = append(assets, asset)
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
	if err := s.cache.CacheAsset(ctx, repoAsset, time.Hour); err != nil {
		s.logger.Warn("Failed to cache asset", zap.Error(err), zap.String("asset_id", asset.ID))
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
	return s.repo.GetTelemetryData(ctx, companyID, start, end)
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
