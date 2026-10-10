package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/employer/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
)

type RouterConfig struct {
	Env                string
	CORSAllowedOrigins []string
	RateLimiterRPS     float64
	RateLimiterBurst   int
}

func NewRouter(
	si api.ServerInterface,
	signer jwtx.Signer,
	logger *slog.Logger,
	cfg RouterConfig,
) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(httpx.GinRecoverer())
	r.Use(httpx.GinRequestID())
	r.Use(httpx.GinLogger(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           5 * time.Minute,
	}))

	wrapper := &api.ServerInterfaceWrapper{
		Handler: si,
		ErrorHandler: func(c *gin.Context, err error, code int) {
			httpx.GinError(c, code, "bad_request", err.Error())
		},
	}

	r.GET("/healthz", wrapper.Healthz)

	// Публичные.
	r.GET("/vacancies", wrapper.ListPublicVacancies)
	r.GET("/vacancies/:id", wrapper.GetPublicVacancy)

	// Работодательские (JWT + role=employer).
	emp := r.Group("")
	emp.Use(httpx.GinJWTAuth(signer))
	emp.Use(httpx.GinRequireVerifiedEmail())
	emp.Use(httpx.GinRequireRole("employer", "admin"))
	{
		emp.GET("/me/company", wrapper.GetMyCompany)
		emp.PUT("/me/company", wrapper.UpsertMyCompany)

		emp.GET("/me/needs", wrapper.ListMyNeeds)
		emp.POST("/me/needs", wrapper.CreateNeed)
		emp.PATCH("/me/needs/:id", wrapper.UpdateNeed)
		emp.DELETE("/me/needs/:id", wrapper.DeleteNeed)
		emp.GET("/me/needs/:id/matches", wrapper.GetNeedMatches)

		emp.GET("/me/vacancies", wrapper.ListMyVacancies)
		emp.POST("/me/vacancies", wrapper.CreateVacancy)
		emp.PATCH("/me/vacancies/:id", wrapper.UpdateVacancy)
		emp.DELETE("/me/vacancies/:id", wrapper.DeleteVacancy)
		emp.POST("/me/vacancies/:id/publish", wrapper.PublishVacancy)

		emp.GET("/candidates", wrapper.SearchCandidates)
		emp.GET("/candidates/:userId", wrapper.GetCandidateCard)
	}

	_ = http.StatusOK
	return r
}
