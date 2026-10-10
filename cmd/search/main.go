package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/config"
	searchhttp "github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/http"
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
	dsn := strings.Replace(cfg.DSN(), "/"+cfg.DBName, "/candidate_db", 1)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Error("db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	router := searchhttp.NewRouter(searchhttp.NewServer(pool), jwtx.NewHS256Signer(cfg.JWTSecret, cfg.JWTIssuer), cfg.CORSAllowedOrigins, cfg.Env)
	srv := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info("search service started", "port", cfg.HTTPPort)
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
