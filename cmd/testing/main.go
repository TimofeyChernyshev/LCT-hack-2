package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	httpdelivery "TimofeyChernyshev/LCT-hack-2/internal/testing/delivery/http"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/config"
	apitesting "TimofeyChernyshev/LCT-hack-2/internal/testing/infrastructure/http"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/migrations"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/repository"
	"TimofeyChernyshev/LCT-hack-2/internal/testing/service"
	"TimofeyChernyshev/LCT-hack-2/pkg/auth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load testing config: %v", err)
	}

	log.Printf("[testing-service] starting on port :%s (env: %s)", cfg.HTTPPort, cfg.Env)

	// Connect to PostgreSQL with retry
	db, err := connectDB(cfg.DSN(), cfg.HTTPRetries, cfg.HTTPTimeout)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// Run migrations
	if cfg.RunMigrations {
		log.Println("[testing-service] running database migrations...")
		goose.SetBaseFS(migrations.FS)
		if err := goose.SetDialect("postgres"); err != nil {
			log.Fatalf("failed to set goose dialect: %v", err)
		}
		if err := goose.Up(db, "."); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
		log.Println("[testing-service] migrations applied successfully")
	}

	// Initialize components
	repo := repository.NewPostgresRepository(db)
	svc := service.NewTestingService(repo, cfg)
	handler := httpdelivery.NewHandler(svc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(corsMiddleware(cfg.CORSAllowedOrigins))

	// Auth Middleware
	jwtValidator := auth.NewJWTValidator(cfg.JWTSecret, cfg.JWTIssuer)
	isDev := cfg.Env == "development" || cfg.Env == ""
	router.Use(auth.Middleware(jwtValidator, isDev))

	// Register generated handlers
	apitesting.RegisterHandlers(router, handler)

	// BE2-02: Questionnaire alias routes
	router.POST("/me/onboarding-questionnaire", handler.SubmitQuestionnaire)
	router.GET("/me/onboarding-questionnaire", handler.GetQuestionnaireState)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[testing-service] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("[testing-service] stopped cleanly")
}

func connectDB(dsn string, retries int, timeout time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 1; i <= retries; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err = db.PingContext(ctx)
			cancel()
			if err == nil {
				db.SetMaxOpenConns(25)
				db.SetMaxIdleConns(5)
				db.SetConnMaxLifetime(5 * time.Minute)
				return db, nil
			}
		}
		log.Printf("[testing-service] waiting for db (attempt %d/%d): %v", i, retries, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to db after %d attempts: %w", retries, err)
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := false
		for _, o := range allowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed || len(allowedOrigins) == 0 {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID, X-User-Role")
			c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
