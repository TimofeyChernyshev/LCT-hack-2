package http

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	api "github.com/TimofeyChernyshev/LCT-hack-2/internal/dict/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/httpx"
)

func NewRouter(si api.ServerInterface, logger *slog.Logger, env string, origins []string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(httpx.GinRecoverer(), httpx.GinRequestID(), httpx.GinLogger(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins: origins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"},
		MaxAge:       5 * time.Minute,
	}))
	wrapper := &api.ServerInterfaceWrapper{
		Handler: si,
		ErrorHandler: func(c *gin.Context, err error, code int) {
			c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
		},
	}
	r.GET("/healthz", wrapper.Healthz)
	r.GET("/industries", wrapper.ListIndustries)
	r.GET("/specializations", wrapper.ListSpecializations)
	r.GET("/grades", wrapper.ListGrades)
	r.GET("/categories", wrapper.ListCategories)
	r.GET("/categories/:id", wrapper.GetCategory)
	r.GET("/technologies", wrapper.ListTechnologies)
	r.GET("/internal/dictionaries/snapshot", wrapper.Snapshot)
	return r
}
