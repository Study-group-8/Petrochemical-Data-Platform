package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/domain"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const assetsCacheKey = "assets:all"

// RedisRepository обрабатывает операции Redis для кэширования
type RedisRepository struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisRepository создает новый репозиторий Redis
func NewRedisRepository(addr, password string, db int, logger *zap.Logger) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis connection
func (r *RedisRepository) Close() {
	r.client.Close()
}

// CacheDataPoint caches a telemetry data point with expiration
func (r *RedisRepository) CacheDataPoint(ctx context.Context, data domain.TelemetryData, expiration time.Duration) error {
	key := fmt.Sprintf("company:%s:product:%s:latest", data.CompanyID, data.ProductName)

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data point: %w", err)
	}

	if err := r.client.Set(ctx, key, dataJSON, expiration).Err(); err != nil {
		r.logger.Error("Failed to cache data point", zap.Error(err),
			zap.String("company_id", data.CompanyID),
			zap.String("product", data.ProductName))
		return err
	}

	return nil
}

// GetCachedDataPoint retrieves a cached telemetry data point
func (r *RedisRepository) GetCachedDataPoint(ctx context.Context, companyID, productName string) (*domain.TelemetryData, error) {
	key := fmt.Sprintf("company:%s:product:%s:latest", companyID, productName)

	dataJSON, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Not found
	} else if err != nil {
		return nil, err
	}

	var data domain.TelemetryData
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return &data, nil
}
func (r *RedisRepository) CacheAsset(ctx context.Context, asset Asset, expiration time.Duration) error {
	key := fmt.Sprintf("asset:%s", asset.ID)

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset: %w", err)
	}

	if err := r.client.Set(ctx, key, assetJSON, expiration).Err(); err != nil {
		r.logger.Error("Failed to cache asset", zap.Error(err), zap.String("asset_id", asset.ID))
		return err
	}

	return nil
}

// CacheAssets stores the full list of assets for faster repeated reads
func (r *RedisRepository) CacheAssets(ctx context.Context, assets []domain.Asset, expiration time.Duration) error {
	data, err := json.Marshal(assets)
	if err != nil {
		return fmt.Errorf("failed to marshal assets list: %w", err)
	}

	if err := r.client.Set(ctx, assetsCacheKey, data, expiration).Err(); err != nil {
		r.logger.Error("Failed to cache assets list", zap.Error(err))
		return err
	}
	return nil
}

// GetCachedAssets returns cached list of assets if present
func (r *RedisRepository) GetCachedAssets(ctx context.Context) ([]domain.Asset, error) {
	dataJSON, err := r.client.Get(ctx, assetsCacheKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var assets []domain.Asset
	if err := json.Unmarshal([]byte(dataJSON), &assets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached assets: %w", err)
	}
	return assets, nil
}

// InvalidateAssetsCache removes the cached list of assets
func (r *RedisRepository) InvalidateAssetsCache(ctx context.Context) error {
	return r.client.Del(ctx, assetsCacheKey).Err()
}

// GetCachedAsset retrieves a cached asset
func (r *RedisRepository) GetCachedAsset(ctx context.Context, assetID string) (*Asset, error) {
	key := fmt.Sprintf("asset:%s", assetID)

	assetJSON, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Not found
	} else if err != nil {
		return nil, err
	}

	var asset Asset
	if err := json.Unmarshal([]byte(assetJSON), &asset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached asset: %w", err)
	}

	return &asset, nil
}

// PublishData publishes telemetry data to Redis pub/sub channel
func (r *RedisRepository) PublishData(ctx context.Context, channel string, data domain.TelemetryData) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for publish: %w", err)
	}

	if err := r.client.Publish(ctx, channel, dataJSON).Err(); err != nil {
		r.logger.Error("Failed to publish data", zap.Error(err), zap.String("channel", channel))
		return err
	}

	return nil
}
