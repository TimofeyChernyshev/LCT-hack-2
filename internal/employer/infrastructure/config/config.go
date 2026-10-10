package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
)

const minJWTSecretLen = 32

type Config struct {
	HTTPPort        string        `env:"HTTP_PORT" envDefault:"8080"`
	Env             string        `env:"APP_ENV" envDefault:"development"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout time.Duration `env:"EMPLOYER_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	RunMigrations   bool          `env:"RUN_MIGRATIONS" envDefault:"true"`

	DBUser     string `env:"POSTGRES_USER" envDefault:"fsp"`
	DBPassword string `env:"POSTGRES_PASSWORD" envDefault:"fsp"`
	DBHost     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	DBPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	DBName     string `env:"EMPLOYER_DB_NAME" envDefault:"employer_db"`
	DBSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`

	JWTSecret string `env:"JWT_SECRET" envDefault:"dev-secret-change-me-please-32-chars-min"`
	JWTIssuer string `env:"JWT_ISSUER" envDefault:"fsp-platform"`

	SearchURL    string        `env:"SEARCH_SERVICE_URL" envDefault:"http://search:8080"`
	CandidateURL string        `env:"CANDIDATE_SERVICE_URL" envDefault:"http://candidate:8080"`
	HTTPTimeout  time.Duration `env:"EMPLOYER_HTTP_TIMEOUT" envDefault:"5s"`
	HTTPRetries  int           `env:"EMPLOYER_HTTP_RETRIES" envDefault:"3"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`

	RateLimiter RateLimiterConfig `envPrefix:"EMPLOYER_RATE_LIMITER_"`
}

type RateLimiterConfig struct {
	RPS   float64 `env:"RPS" envDefault:"50"`
	Burst int     `env:"BURST" envDefault:"100"`
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword),
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse employer config: %w", err)
	}
	if len(cfg.JWTSecret) < minJWTSecretLen {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d chars", minJWTSecretLen)
	}
	return &cfg, nil
}
