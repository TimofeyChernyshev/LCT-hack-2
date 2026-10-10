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

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/application"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/config"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/files"
	candidatehttp "github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/http"
	candidatepostgres "github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/postgres"
	candidatemigrations "github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/migrations"
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
	pool, err := candidatepostgres.NewPool(ctx, cfg.DSN())
	if err != nil {
		log.Error("db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.RunMigrations {
		if err := candidatepostgres.RunMigrations(ctx, cfg.DSN(), candidatemigrations.FS, log); err != nil {
			log.Error("migrations", "err", err)
			os.Exit(1)
		}
	}

	// ---------- Repositories ----------
	profiles := candidatepostgres.NewProfileRepo(pool)
	contacts := candidatepostgres.NewContactsRepo(pool)
	visibility := candidatepostgres.NewVisibilityRepo(pool)
	reveals := candidatepostgres.NewRevealRepo(pool)
	resumes := candidatepostgres.NewResumeRepo(pool)
	experiences := candidatepostgres.NewExperienceRepo(pool)
	technologies := candidatepostgres.NewTechnologyRepo(pool)
	fspRepo := candidatepostgres.NewFSPRepo(pool)
	categoryRepo := candidatepostgres.NewCategoryRepo(pool)

	// ---------- Infrastructure ----------
	txm := pkgpostgres.NewTransactionManager(pool)
	filesStorage := files.NewLocalStorage(cfg.UploadDir)
	signer := jwtx.NewHS256Signer(cfg.JWTSecret, cfg.JWTIssuer)

	// ---------- Application ----------
	svc := application.NewService(
		profiles,
		contacts,
		visibility,
		reveals,
		resumes,
		experiences,
		technologies,
		fspRepo,
		categoryRepo,
		filesStorage,
		txm,
		application.Config{
			MaxUploadBytes: cfg.MaxUploadBytes,
		},
	)

	// ---------- HTTP ----------
	handlers := candidatehttp.NewHandlers(svc, cfg.InternalToken)
	router := candidatehttp.NewRouter(handlers, signer, log, candidatehttp.RouterConfig{
		Env:                cfg.Env,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		RateLimiterRPS:     cfg.RateLimiter.RPS,
		RateLimiterBurst:   cfg.RateLimiter.Burst,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second, // чуть больше — PDF upload
		IdleTimeout:       60 * time.Second,
	}

	// ---------- Run ----------
	serverErr := make(chan error, 1)
	go func() {
		log.Info("candidate service started", "port", cfg.HTTPPort, "env", cfg.Env)
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
	log.Info("candidate service stopped")
}
