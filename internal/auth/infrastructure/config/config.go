package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	minJWTSecretLen = 32
	minPasswordLen  = 6
)

type Config struct {
	// HTTP
	HTTPPort        string        `env:"HTTP_PORT" envDefault:"8080"`
	Env             string        `env:"APP_ENV" envDefault:"development"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout time.Duration `env:"AUTH_SHUTDOWN_TIMEOUT" envDefault:"30s"`
	RunMigrations   bool          `env:"RUN_MIGRATIONS" envDefault:"true"`

	// DB
	DBUser     string `env:"POSTGRES_USER" envDefault:"fsp"`
	DBPassword string `env:"POSTGRES_PASSWORD" envDefault:"fsp"`
	DBHost     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	DBPort     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	DBName     string `env:"AUTH_DB_NAME" envDefault:"auth_db"`
	DBSSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`

	// JWT
	JWTSecret     string        `env:"JWT_SECRET" envDefault:"dev-secret-change-me-please-32-chars-min"`
	JWTIssuer     string        `env:"JWT_ISSUER" envDefault:"fsp-platform"`
	JWTAccessTTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"720h"`

	EmailVerificationTTL time.Duration `env:"EMAIL_VERIFICATION_TTL" envDefault:"24h"`
	PasswordResetTTL     time.Duration `env:"PASSWORD_RESET_TTL" envDefault:"1h"`

	MaxFailedLogins int           `env:"MAX_FAILED_LOGINS" envDefault:"5"`
	LockDuration    time.Duration `env:"LOCK_DURATION" envDefault:"15m"`
	MinPasswordLen  int           `env:"MIN_PASSWORD_LEN" envDefault:"8"`

	ConsentVersionPDN string `env:"CONSENT_VERSION_PDN" envDefault:"v1"`
	ConsentVersionPub string `env:"CONSENT_VERSION_PROFILE_PUBLICATION" envDefault:"v1"`

	// SMTP
	SMTPHost     string `env:"SMTP_HOST" envDefault:"mailhog"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"1025"`
	SMTPUser     string `env:"SMTP_USER"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPFrom     string `env:"SMTP_FROM" envDefault:"noreply@fsp.local"`
	SMTPMode     string `env:"SMTP_MODE" envDefault:"none"`

	AppBaseURL         string   `env:"APP_BASE_URL" envDefault:"http://localhost:3000"`
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`

	RateLimiter RateLimiterConfig `envPrefix:"AUTH_RATE_LIMITER_"`
}

type RateLimiterConfig struct {
	RPS   float64 `env:"RPS" envDefault:"20"`
	BURST int     `env:"BURST" envDefault:"40"`
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword),
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse auth config: %w", err)
	}
	if len(cfg.JWTSecret) < minJWTSecretLen {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d chars", minJWTSecretLen)
	}
	if cfg.MinPasswordLen < minPasswordLen {
		return nil, fmt.Errorf("MIN_PASSWORD_LEN must be at least 6")
	}
	return &cfg, nil
}
