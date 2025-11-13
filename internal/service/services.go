package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"petrochemical-data-platform/internal/domain"
	"petrochemical-data-platform/internal/pkg/simulator"
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

// InitializeAssetsFromSensors создает активы на основе конфигураций датчиков
func (s *AssetService) InitializeAssetsFromSensors(ctx context.Context, sensors []simulator.SensorConfig) error {
	for _, sensor := range sensors {
		// Extract company, location, and equipment info from sensor ID and description
		company, location, equipment := s.extractAssetInfo(sensor)

		assetID := fmt.Sprintf("%s_%s", company, location)
		assetName := fmt.Sprintf("%s - %s", company, location)

		asset := domain.Asset{
			ID:       assetID,
			Name:     assetName,
			Type:     "Производственное оборудование",
			Location: location,
			Status:   "active",
			Specifications: map[string]interface{}{
				"company":   company,
				"equipment": equipment,
				"products":  s.getCompanyProducts(company),
			},
			Sensors:   []string{sensor.ID},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.CreateAsset(ctx, asset); err != nil {
			s.logger.Warn("Failed to create asset", zap.Error(err), zap.String("asset_id", assetID))
			// Continue with other assets
		}
	}

	return nil
}

// extractAssetInfo извлекает компанию, местоположение и оборудование из описания датчика
func (s *AssetService) extractAssetInfo(sensor simulator.SensorConfig) (company, location, equipment string) {
	description := sensor.Description

	// Extract company name
	if strings.Contains(description, "СИБУР") {
		company = "СИБУР Холдинг"
		if strings.Contains(description, "Тобольск") {
			location = "Тюменская обл."
		}
	} else if strings.Contains(description, "Нижнекамскнефтехим") {
		company = "Нижнекамскнефтехим"
		location = "Татарстан"
	} else if strings.Contains(description, "Ангарская НХК") {
		company = "Ангарская НХК (Роснефть)"
		location = "Восточная Сибирь"
	} else if strings.Contains(description, "ЗапСибНефтехим") {
		company = "ЗапСибНефтехим"
		location = "Тюменская обл."
	} else if strings.Contains(description, "Новокуйбышевская НХК") {
		company = "Новокуйбышевская НХК"
		location = "Самарская обл."
	} else if strings.Contains(description, "Ставролен") {
		company = "Ставролен"
		location = "Ставропольский край"
	} else if strings.Contains(description, "Балтийский ХК") {
		company = "Балтийский Химический Комплекс"
		location = "Ленинградская обл."
	} else if strings.Contains(description, "Стерлитамакский НХЗ") {
		company = "Стерлитамакский НХЗ"
		location = "Башкортостан"
	} else if strings.Contains(description, "Кемеровский КХЗ") {
		company = "Кемеровский КХЗ"
		location = "Кемеровская обл."
	}

	// Extract equipment type
	if strings.Contains(description, "Реактор") {
		equipment = "Реактор"
	} else if strings.Contains(description, "Колонна") {
		equipment = "Колонна"
	} else if strings.Contains(description, "Линия") {
		equipment = "Производственная линия"
	} else if strings.Contains(description, "Резервуар") {
		equipment = "Резервуар"
	} else if strings.Contains(description, "Силос") {
		equipment = "Силос"
	} else if strings.Contains(description, "Батарея") {
		equipment = "Коксовая батарея"
	} else {
		equipment = "Оборудование"
	}

	return company, location, equipment
}

// getCompanyProducts возвращает продукты для данной компании
func (s *AssetService) getCompanyProducts(company string) []string {
	products := map[string][]string{
		"СИБУР Холдинг":                  {"Полипропилен", "Полиэтилен", "МТБЭ", "Бутадиен"},
		"Нижнекамскнефтехим":             {"Синтетические каучуки", "Полиэтилен", "Стирол"},
		"Ангарская НХК (Роснефть)":       {"Нефтепродукты", "Полимеры", "Ароматика"},
		"ЗапСибНефтехим":                 {"Полипропилен", "Полиэтилен"},
		"Новокуйбышевская НХК":           {"Полипропилен", "Бутиловые каучуки", "МТБЭ"},
		"Ставролен":                      {"Полипропилен", "Полиэтилен"},
		"Балтийский Химический Комплекс": {"ПЭВД", "Полистирол"},
		"Стерлитамакский НХЗ":            {"Каучуки", "Антиоксиданты", "МТБЭ"},
		"Кемеровский КХЗ":                {"Кокс", "Бензол", "Нафталин"},
	}

	if prods, exists := products[company]; exists {
		return prods
	}
	return []string{}
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

// GetTelemetry получает данные телеметрии для датчика
func (s *TelemetryService) GetTelemetry(ctx context.Context, sensorID string, start, end time.Time) ([]domain.TelemetryData, error) {
	return s.repo.GetTelemetryData(ctx, sensorID, start, end)
}

// getUnitForSensor определяет подходящую единицу измерения на основе ID датчика
func (s *TelemetryService) getUnitForSensor(sensorID string) string {
	// All sensors show production volume in tons
	return "тонны"
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
