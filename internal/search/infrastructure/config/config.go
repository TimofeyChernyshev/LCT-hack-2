package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPPort        string        `env:"HTTP_PORT" envDefault:"8080"`
	Env             string        `env:"APP_ENV" envDefault:"development"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout time.Duration `env:"SEARCH_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	RunMigrations   bool          `env:"RUN_MIGRATIONS" envDefault:"true"`

	DBUser     string `env:"POSTGRES_USER" envDefault:"fsp"`
	DBPassword string `env:"POSTGRES_PASSWORD" envDefault:"fsp"`
	DBHost     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	DBPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	DBName     string `env:"SEARCH_DB_NAME" envDefault:"search_db"`
	DBSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`

	JWTSecret string `env:"JWT_SECRET" envDefault:"dev-secret-change-me-please-32-chars-min"`
	JWTIssuer string `env:"JWT_ISSUER" envDefault:"fsp-platform"`

	AuthURL      string `env:"AUTH_SERVICE_URL" envDefault:"http://auth:8080"`
	DictURL      string `env:"DICT_SERVICE_URL" envDefault:"http://dict:8080"`
	CandidateURL string `env:"CANDIDATE_SERVICE_URL" envDefault:"http://candidate:8080"`
	TestingURL   string `env:"TESTING_SERVICE_URL" envDefault:"http://testing:8080"`

	// синхронизация read-model
	SyncBatchSize int           `env:"SEARCH_SYNC_BATCH_SIZE" envDefault:"100"`
	SyncInterval  time.Duration `env:"SEARCH_SYNC_INTERVAL" envDefault:"30s"`

	DefaultLimit int `env:"SEARCH_DEFAULT_LIMIT" envDefault:"20"`
	MaxLimit     int `env:"SEARCH_MAX_LIMIT" envDefault:"100"`

	HTTPTimeout time.Duration `env:"SEARCH_HTTP_TIMEOUT" envDefault:"5s"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword),
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse search config: %w", err)
	}
	if cfg.DefaultLimit < 1 || cfg.DefaultLimit > cfg.MaxLimit {
		return nil, fmt.Errorf("SEARCH_DEFAULT_LIMIT must be in [1;SEARCH_MAX_LIMIT]")
	}
	return &cfg, nil
}
