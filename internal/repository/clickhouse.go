package repository

import (
	"context"
	"fmt"
	"time"

	"petrochemical-data-platform/internal/domain"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// ClickHouseRepository обрабатывает операции ClickHouse для данных телеметрии
type ClickHouseRepository struct {
	conn   clickhouse.Conn
	logger *zap.Logger
}

// TelemetryData представляет данные производства/продаж для ClickHouse
type TelemetryData struct {
	CompanyID   string    `ch:"company_id"`
	ProductName string    `ch:"product_name"`
	Value       float64   `ch:"value"`
	Unit        string    `ch:"unit"`
	Timestamp   time.Time `ch:"timestamp"`
	Quality     uint16    `ch:"quality"`
}

// NewClickHouseRepository создает новый репозиторий ClickHouse
func NewClickHouseRepository(addr, database, username, password string, logger *zap.Logger) (*ClickHouseRepository, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &ClickHouseRepository{
		conn:   conn,
		logger: logger,
	}, nil
}

// Close закрывает соединение с ClickHouse
func (r *ClickHouseRepository) Close() error {
	return r.conn.Close()
}

// SaveTelemetryData сохраняет данные производства/продаж в ClickHouse
func (r *ClickHouseRepository) SaveTelemetryData(ctx context.Context, data TelemetryData) error {
	query := `
		INSERT INTO petrochemical.telemetry
		(company_id, product_name, value, unit, timestamp, quality)
		VALUES (?, ?, ?, ?, ?, ?)`

	err := r.conn.Exec(ctx, query, data.CompanyID, data.ProductName, data.Value, data.Unit, data.Timestamp, data.Quality)
	if err != nil {
		return fmt.Errorf("failed to save telemetry data: %w", err)
	}

	r.logger.Debug("Saved telemetry data",
		zap.String("company_id", data.CompanyID),
		zap.String("product", data.ProductName),
		zap.Float64("value", data.Value),
		zap.String("unit", data.Unit))

	return nil
}

// GetTelemetryData получает данные производства/продаж для компании в заданном временном диапазоне
func (r *ClickHouseRepository) GetTelemetryData(ctx context.Context, companyID string, start, end time.Time) ([]domain.TelemetryData, error) {
	query := `
		SELECT company_id, product_name, value, unit, timestamp, quality
		FROM petrochemical.telemetry
		WHERE company_id = ? AND timestamp >= ? AND timestamp <= ?
		ORDER BY timestamp DESC
		LIMIT 1000`

	rows, err := r.conn.Query(ctx, query, companyID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query telemetry data: %w", err)
	}
	defer rows.Close()

	var results []domain.TelemetryData
	for rows.Next() {
		var data domain.TelemetryData
		err := rows.Scan(&data.CompanyID, &data.ProductName, &data.Value, &data.Unit, &data.Timestamp, &data.Quality)
		if err != nil {
			return nil, fmt.Errorf("failed to scan telemetry data: %w", err)
		}
		results = append(results, data)
	}

	if len(results) == 0 {
		// Return mock data if no real data found (for development)
		r.logger.Info("No telemetry data found, returning mock data",
			zap.String("company_id", companyID))
		results = []domain.TelemetryData{
			{
				CompanyID:   companyID,
				ProductName: "Полипропилен",
				Value:       125.5,
				Unit:        "т/час",
				Timestamp:   time.Now(),
				Quality:     1,
			},
		}
	}

	return results, nil
}
