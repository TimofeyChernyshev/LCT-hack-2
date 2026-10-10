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
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	httpdelivery "github.com/TimofeyChernyshev/LCT-hack-2/internal/search/delivery/http"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/domain"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/config"
	apisearch "github.com/TimofeyChernyshev/LCT-hack-2/internal/search/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/migrations"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/repository"
	"github.com/TimofeyChernyshev/LCT-hack-2/internal/search/service"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/auth"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/ranking"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("[search-service] starting on port :%s (env: %s)...", cfg.HTTPPort, cfg.Env)

	// Connect to Database
	db, err := connectDB(cfg.DSN(), 10, cfg.HTTPTimeout)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	// Run Migrations
	if cfg.RunMigrations {
		log.Println("[search-service] running database migrations...")
		goose.SetBaseFS(migrations.FS)
		if err := goose.SetDialect("postgres"); err != nil {
			log.Fatalf("failed to set goose dialect: %v", err)
		}
		if err := goose.Up(db, "."); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
		log.Println("[search-service] migrations applied successfully")
	}

	repo := repository.NewPostgresRepository(db)
	svc := service.NewSearchService(repo, cfg)

	// Seed demo candidates if index is empty
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	count, err := repo.Count(ctx)
	cancel()
	if err == nil && count == 0 {
		log.Println("[search-service] seeding initial candidate search documents...")
		seedCandidates(repo, svc)
		log.Println("[search-service] candidate search documents seeded")
	}

	handler := httpdelivery.NewHandler(svc)

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(corsMiddleware(cfg.CORSAllowedOrigins))

	// Auth Middleware (optional bearer token validation)
	jwtValidator := auth.NewJWTValidator(cfg.JWTSecret, cfg.JWTIssuer)
	isDev := cfg.Env == "development" || cfg.Env == ""
	router.Use(auth.Middleware(jwtValidator, isDev))

	// Register generated handlers
	apisearch.RegisterHandlers(router, handler)

	// BE2-06: Standardized Candidate PDF Export & PDF Resume Parser
	router.GET("/candidates/:userId/export-pdf", handler.ExportCandidatePDF)
	router.POST("/candidates/export-pdf", handler.ExportCandidatePDF)
	router.POST("/resumes/parse-pdf", handler.ParseResumePDF)
	router.POST("/me/resumes/upload", handler.ParseResumePDF)

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
	log.Println("[search-service] shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("[search-service] stopped cleanly")
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
		log.Printf("[search-service] waiting for db (attempt %d/%d): %v", i, retries, err)
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

func seedCandidates(repo repository.Repository, svc *service.SearchService) {
	ctx := context.Background()

	// Well-known specialization and category IDs
	backendCatID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	frontendCatID := uuid.MustParse("00000000-0000-0000-0000-000000000102")
	specBackendID := uuid.MustParse("00000000-0000-0000-0000-000000000201")
	specFrontendID := uuid.MustParse("00000000-0000-0000-0000-000000000202")

	gradeSeniorID := uuid.MustParse("00000000-0000-0000-0000-000000000303")
	gradeMiddleID := uuid.MustParse("00000000-0000-0000-0000-000000000302")
	gradeJuniorID := uuid.MustParse("00000000-0000-0000-0000-000000000301")

	// Tech stack IDs
	goID := uuid.MustParse("00000000-0000-0000-0000-000000000401")
	pgID := uuid.MustParse("00000000-0000-0000-0000-000000000402")
	redisID := uuid.MustParse("00000000-0000-0000-0000-000000000403")
	pyID := uuid.MustParse("00000000-0000-0000-0000-000000000404")
	reactID := uuid.MustParse("00000000-0000-0000-0000-000000000405")
	tsID := uuid.MustParse("00000000-0000-0000-0000-000000000406")

	p1 := 1
	p2 := 2
	p3 := 3
	now := time.Now()

	score98 := 98.0
	score95 := 95.0
	score88 := 88.0
	score82 := 82.0
	score74 := 74.0

	exp6 := 6.5
	exp4 := 4.0
	exp2 := 2.5
	exp1 := 1.0

	moscow := "г. Москва"
	spb := "г. Санкт-Петербург"
	novosib := "г. Новосибирск"

	seeds := []domain.CandidateSearchDoc{
		{
			UserID:               uuid.MustParse("11111111-1111-1111-1111-111111111101"),
			DisplayName:          "Александр Смирнов",
			CategoryID:           backendCatID,
			SpecializationID:     specBackendID,
			SpecializationName:   "Backend Go / Python",
			GradeID:              gradeSeniorID,
			GradeName:            "Senior",
			GradeRank:            3,
			TestScore:            &score98,
			HasFSP:               true,
			SportsRank:           "Мастер спорта",
			FSPRating:            2480,
			FSPScore:             92.0,
			FSPAchievementsCount: 6,
			FSPBestPlace:         &p2,
			FSPWeightSum:         24,
			FSPHighlights:        []string{"серебряный призёр Кубка ФСП 2025", "победитель Чемпионата Москвы"},
			Stack:                []uuid.UUID{goID, pgID, redisID},
			YearsExperience:      &exp6,
			Location:             &moscow,
			PeriodicTasksSolved:  5,
			LastActiveAt:         now,
		},
		{
			UserID:               uuid.MustParse("11111111-1111-1111-1111-111111111102"),
			DisplayName:          "Дмитрий Ковалев",
			CategoryID:           backendCatID,
			SpecializationID:     specBackendID,
			SpecializationName:   "Backend Python",
			GradeID:              gradeSeniorID,
			GradeName:            "Senior",
			GradeRank:            3,
			TestScore:            &score95,
			HasFSP:               true,
			SportsRank:           "КМС",
			FSPRating:            2250,
			FSPScore:             82.0,
			FSPAchievementsCount: 4,
			FSPBestPlace:         &p1,
			FSPWeightSum:         18,
			FSPHighlights:        []string{"1 место Всероссийского хакатона ФСП 2024"},
			Stack:                []uuid.UUID{pyID, pgID, redisID},
			YearsExperience:      &exp4,
			Location:             &spb,
			PeriodicTasksSolved:  4,
			LastActiveAt:         now,
		},
		{
			UserID:               uuid.MustParse("11111111-1111-1111-1111-111111111103"),
			DisplayName:          "Михаил Романов",
			CategoryID:           backendCatID,
			SpecializationID:     specBackendID,
			SpecializationName:   "Backend Python",
			GradeID:              gradeMiddleID,
			GradeName:            "Middle",
			GradeRank:            2,
			TestScore:            &score88,
			HasFSP:               false,
			FSPScore:             0.0,
			Stack:                []uuid.UUID{pyID, pgID},
			YearsExperience:      &exp2,
			Location:             &moscow,
			PeriodicTasksSolved:  3,
			LastActiveAt:         now,
		},
		{
			UserID:               uuid.MustParse("11111111-1111-1111-1111-111111111104"),
			DisplayName:          "Анна Соколова",
			CategoryID:           frontendCatID,
			SpecializationID:     specFrontendID,
			SpecializationName:   "Frontend React",
			GradeID:              gradeSeniorID,
			GradeName:            "Senior",
			GradeRank:            3,
			TestScore:            &score95,
			HasFSP:               true,
			SportsRank:           "1-й спортивный разряд",
			FSPRating:            2100,
			FSPScore:             78.0,
			FSPAchievementsCount: 3,
			FSPBestPlace:         &p3,
			FSPWeightSum:         14,
			FSPHighlights:        []string{"бронзовый призёр Чемпионата ФСП 2024"},
			Stack:                []uuid.UUID{reactID, tsID},
			YearsExperience:      &exp4,
			Location:             &spb,
			PeriodicTasksSolved:  3,
			LastActiveAt:         now,
		},
		{
			UserID:               uuid.MustParse("11111111-1111-1111-1111-111111111105"),
			DisplayName:          "Сергей Васильев",
			CategoryID:           frontendCatID,
			SpecializationID:     specFrontendID,
			SpecializationName:   "Frontend React",
			GradeID:              gradeMiddleID,
			GradeName:            "Middle",
			GradeRank:            2,
			TestScore:            &score82,
			HasFSP:               false,
			FSPScore:             0.0,
			Stack:                []uuid.UUID{reactID, tsID},
			YearsExperience:      &exp2,
			Location:             &novosib,
			PeriodicTasksSolved:  2,
			LastActiveAt:         now,
		},
		{
			UserID:               uuid.MustParse("11111111-1111-1111-1111-111111111106"),
			DisplayName:          "Павел Орлов",
			CategoryID:           backendCatID,
			SpecializationID:     specBackendID,
			SpecializationName:   "Backend Go",
			GradeID:              gradeJuniorID,
			GradeName:            "Junior",
			GradeRank:            1,
			TestScore:            &score74,
			HasFSP:               false,
			FSPScore:             0.0,
			Stack:                []uuid.UUID{goID},
			YearsExperience:      &exp1,
			Location:             &moscow,
			PeriodicTasksSolved:  1,
			LastActiveAt:         now,
		},
	}

	for _, s := range seeds {
		doc := s
		features := ranking.CandidateFeatures{
			UserID:               doc.UserID,
			DisplayName:          doc.DisplayName,
			CategoryID:           doc.CategoryID,
			SpecializationID:     doc.SpecializationID,
			SpecializationName:   doc.SpecializationName,
			GradeID:              doc.GradeID,
			GradeName:            doc.GradeName,
			GradeRank:            doc.GradeRank,
			TestScore:            *doc.TestScore,
			HasFSP:               doc.HasFSP,
			FSPScore:             doc.FSPScore,
			FSPRating:            doc.FSPRating,
			SportsRank:           doc.SportsRank,
			FSPAchievementsCount: doc.FSPAchievementsCount,
			FSPBestPlace:         doc.FSPBestPlace,
			FSPWeightSum:         doc.FSPWeightSum,
			FSPHighlights:        doc.FSPHighlights,
			CandidateStack:       doc.Stack,
			YearsExperience:      doc.YearsExperience,
			PeriodicTasksSolved:  doc.PeriodicTasksSolved,
			LastActiveAt:         &doc.LastActiveAt,
		}

		ranked := ranking.Calculate(features, ranking.RankingParams{})
		doc.CalculatedScore = float64(ranked.FinalScore)
		doc.ActivityScore = float64(ranked.ActivityScore)
		doc.Explanation = ranked.Explanation
		doc.Reasons = ranked.Reasons

		if err := repo.Upsert(ctx, &doc); err != nil {
			log.Printf("[search-service] error seeding candidate %s: %v", doc.UserID, err)
		}
	}
}
