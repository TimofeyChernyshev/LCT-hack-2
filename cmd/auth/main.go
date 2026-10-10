package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/config"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/email"
	authhttp "github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/http"
	authpostgres "github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/postgres"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/security"
	authmigrations "github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/migrations"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/logger"
	pkgpostgres "github.com/TimofeyChernyshev/LCT-hack-2/pkg/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}
	log := logger.New(cfg.Env, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---------- DB ----------
	pool, err := authpostgres.NewPool(ctx, cfg.DSN())
	if err != nil {
		slog.Error("db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := authpostgres.RunMigrations(ctx, cfg.DSN(), authmigrations.FS); err != nil {
			slog.Error("migrations", "err", err)
			os.Exit(1)
		}
	}

	// ---------- Repositories ----------
	users := authpostgres.NewUserRepo(pool)
	verif := authpostgres.NewVerificationTokenRepo(pool)
	reset := authpostgres.NewPasswordResetRepo(pool)
	refresh := authpostgres.NewRefreshTokenRepo(pool)
	consents := authpostgres.NewConsentRepo(pool)
	audit := authpostgres.NewAuditRepo(pool)

	// ---------- Infrastructure ----------
	txm := pkgpostgres.NewTransactionManager(pool)
	hasher := security.NewBcryptHasher(0)
	mailer := email.NewSMTPSender(
		cfg.SMTPHost, cfg.SMTPPort,
		cfg.SMTPUser, cfg.SMTPPassword,
		cfg.SMTPFrom,
		email.TLSMode(cfg.SMTPMode),
	)
	signer := jwtx.NewHS256Signer(cfg.JWTSecret, cfg.JWTIssuer)

	// ---------- Application ----------
	svc := application.NewService(
		users, verif, reset, refresh, consents, audit, txm,
		mailer, hasher, signer,
		application.Config{
			Issuer:               cfg.JWTIssuer,
			AccessTTL:            cfg.JWTAccessTTL,
			RefreshTTL:           cfg.JWTRefreshTTL,
			EmailVerificationTTL: cfg.EmailVerificationTTL,
			PasswordResetTTL:     cfg.PasswordResetTTL,
			AppBaseURL:           cfg.AppBaseURL,
			MaxFailedLogins:      cfg.MaxFailedLogins,
			LockDuration:         cfg.LockDuration,
			ConsentVersionPDN:    cfg.ConsentVersionPDN,
			ConsentVersionPub:    cfg.ConsentVersionPub,
			MinPasswordLen:       cfg.MinPasswordLen,
		},
	)

	// ---------- HTTP ----------
	handlers := authhttp.NewHandlers(svc)
	router := authhttp.NewRouter(handlers, signer, log, authhttp.RouterConfig{
		Env:                cfg.Env,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		RateLimiterRPS:     cfg.RateLimiter.RPS,
		RateLimiterBurst:   cfg.RateLimiter.BURST,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("auth service started", "port", cfg.HTTPPort, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		slog.Error("http server error", "err", err)
	}

	shCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	slog.Info("auth service stopped")
}
