package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/interaction/infrastructure/config"
	interactionhttp "github.com/TimofeyChernyshev/LCT-hack-2/internal/interaction/infrastructure/http"
	interactionmigrations "github.com/TimofeyChernyshev/LCT-hack-2/internal/interaction/migrations"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/logger"
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
	if err := migrate(ctx, cfg.DSN()); err != nil {
		log.Error("migrations", "err", err)
		os.Exit(1)
	}
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Error("db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	token := os.Getenv("INTERNAL_TOKEN")
	if token == "" {
		token = "dev-internal-token"
	}
	router := interactionhttp.NewRouter(interactionhttp.NewServer(pool, cfg.CandidateURL, token), jwtx.NewHS256Signer(cfg.JWTSecret, cfg.JWTIssuer), cfg.CORSAllowedOrigins, cfg.Env)
	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info("interaction service started", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	sh, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	_ = srv.Shutdown(sh)
}

func migrate(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(interactionmigrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.UpContext(ctx, db, ".")
}
