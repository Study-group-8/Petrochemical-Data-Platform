package bootstrap

import (
	"fmt"

	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/config"
	"github.com/Study-group-8/Petrochemical-Data-Platform/internal/repository"
	"go.uber.org/zap"
)

// LoadConfig loads configuration from the provided path (empty -> default)
func LoadConfig(path string) (*config.Config, error) {
	if path == "" {
		return config.Load()
	}
	return config.LoadFromPath(path)
}

// InitPostgresRepository initializes a Postgres repository from config
func InitPostgresRepository(cfg *config.Config, logger *zap.Logger) (*repository.PostgresRepository, error) {
	pg := cfg.Database.PostgreSQL
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pg.User, pg.Password, pg.Host, pg.Port, pg.DBName, pg.SSLMode)
	return repository.NewPostgresRepository(connStr, logger)
}

// InitClickHouseRepository initializes ClickHouse repository from config
func InitClickHouseRepository(cfg *config.Config, logger *zap.Logger) (*repository.ClickHouseRepository, error) {
	ch := cfg.Database.ClickHouse
	addr := fmt.Sprintf("%s:%s", ch.Host, ch.Port)
	return repository.NewClickHouseRepository(addr, ch.Database, ch.User, ch.Password, logger)
}

// InitRedisRepository initializes Redis repository from config
func InitRedisRepository(cfg *config.Config, logger *zap.Logger) (*repository.RedisRepository, error) {
	r := cfg.Redis
	addr := fmt.Sprintf("%s:%s", r.Host, r.Port)
	return repository.NewRedisRepository(addr, r.Password, r.DB, logger)
}
