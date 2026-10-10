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

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/clients"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/config"
	employerhttp "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/http"
	employerpostgres "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/postgres"
	employermigrations "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/migrations"
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

	pool, err := employerpostgres.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Error("db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := employerpostgres.RunMigrations(ctx, cfg.DSN(), employermigrations.FS, log); err != nil {
			log.Error("migrations", "err", err)
			os.Exit(1)
		}
	}

	// repos
	companies := employerpostgres.NewCompanyRepo(pool)
	needs := employerpostgres.NewNeedRepo(pool)
	vacancies := employerpostgres.NewVacancyRepo(pool)

	// clients
	searchClient := clients.NewSearchClient(cfg.SearchURL, cfg.HTTPTimeout)
	candidateClient := clients.NewCandidateClient(cfg.CandidateURL, cfg.HTTPTimeout)

	txm := pkgpostgres.NewTransactionManager(pool)

	svc := application.NewService(
		companies, needs, vacancies,
		searchClient, candidateClient,
		txm,
	)

	handlers := employerhttp.NewHandlers(svc)
	router := employerhttp.NewRouter(handlers,
		jwtx.NewHS256Signer(cfg.JWTSecret, cfg.JWTIssuer),
		log,
		employerhttp.RouterConfig{
			Env:                cfg.Env,
			CORSAllowedOrigins: cfg.CORSAllowedOrigins,
			RateLimiterRPS:     cfg.RateLimiter.RPS,
			RateLimiterBurst:   cfg.RateLimiter.Burst,
		},
	)

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
		log.Info("employer service started", "port", cfg.HTTPPort, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-serverErr:
		log.Error("http server error", "err", err)
	}

	shCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		log.Error("shutdown", "err", err)
	}
	log.Info("employer service stopped")
}
