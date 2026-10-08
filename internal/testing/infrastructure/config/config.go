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
	ShutdownTimeout time.Duration `env:"TESTING_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	RunMigrations   bool          `env:"RUN_MIGRATIONS" envDefault:"true"`

	DBUser     string `env:"POSTGRES_USER" envDefault:"fsp"`
	DBPassword string `env:"POSTGRES_PASSWORD" envDefault:"fsp"`
	DBHost     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	DBPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	DBName     string `env:"TESTING_DB_NAME" envDefault:"testing_db"`
	DBSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`

	JWTSecret string `env:"JWT_SECRET" envDefault:"dev-secret-change-me-please-32-chars-min"`
	JWTIssuer string `env:"JWT_ISSUER" envDefault:"fsp-platform"`

	AuthURL      string `env:"AUTH_SERVICE_URL" envDefault:"http://auth:8080"`
	DictURL      string `env:"DICT_SERVICE_URL" envDefault:"http://dict:8080"`
	CandidateURL string `env:"CANDIDATE_SERVICE_URL" envDefault:"http://candidate:8080"`

	// бизнес-правила тестирования
	GradeChangeCooldownDays int     `env:"GRADE_CHANGE_COOLDOWN_DAYS" envDefault:"90"`
	ItemsPerSession         int     `env:"TESTING_ITEMS_PER_SESSION" envDefault:"20"`
	PassThresholdRatio      float64 `env:"TESTING_PASS_THRESHOLD_RATIO" envDefault:"0.7"`
	UpgradeThresholdRatio   float64 `env:"TESTING_UPGRADE_THRESHOLD_RATIO" envDefault:"0.9"`

	HTTPTimeout time.Duration `env:"TESTING_HTTP_TIMEOUT" envDefault:"5s"`
	HTTPRetries int           `env:"TESTING_HTTP_RETRIES" envDefault:"3"`

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
		return nil, fmt.Errorf("parse testing config: %w", err)
	}
	if cfg.PassThresholdRatio <= 0 || cfg.PassThresholdRatio > 1 {
		return nil, fmt.Errorf("TESTING_PASS_THRESHOLD_RATIO must be in (0;1]")
	}
	if cfg.UpgradeThresholdRatio < cfg.PassThresholdRatio || cfg.UpgradeThresholdRatio > 1 {
		return nil, fmt.Errorf("TESTING_UPGRADE_THRESHOLD_RATIO must be in [pass;1]")
	}
	return &cfg, nil
}
