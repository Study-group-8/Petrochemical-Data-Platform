package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PostgresRepository обрабатывает операции PostgreSQL
type PostgresRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPostgresRepository создает новый репозиторий PostgreSQL
func NewPostgresRepository(connString string, logger *zap.Logger) (*PostgresRepository, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{
		pool:   pool,
		logger: logger,
	}, nil
}

// Close закрывает соединение с базой данных
func (r *PostgresRepository) Close() {
	r.pool.Close()
}

// SaveAsset сохраняет метаданные актива
func (r *PostgresRepository) SaveAsset(ctx context.Context, asset Asset) error {
	query := `
		INSERT INTO assets (id, name, type, location, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			location = EXCLUDED.location,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, query,
		asset.ID, asset.Name, asset.Type, asset.Location,
		asset.CreatedAt, asset.UpdatedAt)

	if err != nil {
		r.logger.Error("Failed to save asset", zap.Error(err), zap.String("asset_id", asset.ID))
		return err
	}

	return nil
}

// GetAssets получает все активы
func (r *PostgresRepository) GetAssets(ctx context.Context) ([]Asset, error) {
	query := `SELECT id, name, type, location, created_at, updated_at FROM assets ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []Asset
	for rows.Next() {
		var asset Asset
		err := rows.Scan(&asset.ID, &asset.Name, &asset.Type, &asset.Location, &asset.CreatedAt, &asset.UpdatedAt)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, rows.Err()
}

// Asset represents equipment metadata
type Asset struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	Location  string    `json:"location" db:"location"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
