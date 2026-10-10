package http

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/candidate/infrastructure/http/api"
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
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID", "X-Internal-Token"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           5 * time.Minute,
	}))

	wrapper := &api.ServerInterfaceWrapper{
		Handler: si,
		ErrorHandler: func(c *gin.Context, err error, code int) {
			c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
		},
	}

	r.GET("/healthz", wrapper.Healthz)

	// Публичные (внутри — только с JWT).
	auth := r.Group("")
	auth.Use(httpx.GinJWTAuth(signer))
	auth.Use(httpx.GinRequireVerifiedEmail())
	{
		auth.GET("/me/profile", wrapper.GetMyProfile)
		auth.PATCH("/me/profile", wrapper.UpdateMyProfile)

		auth.GET("/me/contacts", wrapper.GetMyContacts)
		auth.PUT("/me/contacts", wrapper.ReplaceMyContacts)

		auth.GET("/me/visibility", wrapper.GetMyVisibility)
		auth.PUT("/me/visibility", wrapper.UpdateMyVisibility)

		auth.GET("/me/resumes", wrapper.ListResumes)
		auth.POST("/me/resumes", wrapper.CreateResume)
		auth.POST("/me/resumes/upload", wrapper.UploadResumePDF)

		auth.GET("/me/experiences", wrapper.ListExperiences)
		auth.POST("/me/experiences", wrapper.AddExperience)
		if h, ok := si.(*Handlers); ok {
			auth.DELETE("/me/experiences/:experienceId", h.DeleteExperience)
			auth.GET("/me/technologies", h.ListMyTechnologies)
		}

		auth.PUT("/me/technologies", wrapper.ReplaceTechnologies)
		auth.GET("/me/category", wrapper.GetMyCategory)
		if h, ok := si.(*Handlers); ok {
			auth.PUT("/me/category", h.AssignMyCategory)
		}
		auth.GET("/me/fsp", wrapper.GetMyFSP)
		auth.PUT("/me/fsp", wrapper.LinkFSP)
	}

	// Работодательский доступ.
	employer := r.Group("")
	employer.Use(httpx.GinJWTAuth(signer))
	employer.Use(httpx.GinRequireRole("employer", "admin"))
	{
		employer.GET("/candidates/:userId", wrapper.GetCandidateForEmployer)
	}

	// Внутренние маршруты (закрыты X-Internal-Token в хендлере).
	internal := r.Group("/internal")
	{
		internal.POST("/reveals", wrapper.InternalRecordReveal)
	}

	return r
}
