package http

import (
	"log/slog"
	"time"

	apiauth "github.com/TimofeyChernyshev/LCT-hack-2/internal/auth/infrastructure/http/api"
	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	Env                string
	CORSAllowedOrigins []string
	RateLimiterRPS     float64
	RateLimiterBurst   int
}

func NewRouter(
	si apiauth.ServerInterface,
	signer jwtx.Signer,
	logger *slog.Logger,
	cfg RouterConfig,
) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestID())
	r.Use(Logger(logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           5 * time.Minute,
	}))

	wrapper := &apiauth.ServerInterfaceWrapper{
		Handler: si,
		ErrorHandler: func(c *gin.Context, err error, code int) {
			c.AbortWithStatusJSON(code, gin.H{"code": "bad_request", "message": err.Error()})
		},
	}

	// Health.
	r.GET("/healthz", wrapper.Healthz)

	// Публичные маршруты + rate limiter.
	public := r.Group("")
	public.Use(RateLimiter(cfg.RateLimiterRPS, cfg.RateLimiterBurst))
	{
		public.POST("/auth/register", wrapper.Register)
		public.POST("/auth/confirm-email", wrapper.ConfirmEmail)
		public.POST("/auth/resend-confirmation", wrapper.ResendConfirmation)
		public.POST("/auth/login", wrapper.Login)
		public.POST("/auth/refresh", wrapper.Refresh)
		public.POST("/auth/forgot-password", wrapper.ForgotPassword)
		public.POST("/auth/reset-password", wrapper.ResetPassword)
		public.POST("/auth/external/:provider/callback", wrapper.ExternalCallback)
	}

	// Требуется аутентификация.
	protected := r.Group("")
	protected.Use(JWTAuth(signer))
	{
		protected.POST("/auth/logout", wrapper.Logout)
		protected.GET("/auth/me", wrapper.Me)
		protected.DELETE("/auth/me", wrapper.DeleteMe)
	}

	// Требуется подтверждённый email (на будущее — когда вернём consents).
	verified := r.Group("")
	verified.Use(JWTAuth(signer))
	verified.Use(RequireVerifiedEmail())
	{
		verified.GET("/auth/consents", wrapper.ListConsents)
		verified.POST("/auth/consents/:type/grant", wrapper.GrantConsent)
		verified.POST("/auth/consents/:type/revoke", wrapper.RevokeConsent)
	}

	// Внутренние маршруты — для других сервисов (в prod закрываются на gateway).
	internal := r.Group("/internal")
	{
		internal.GET("/users/:userId", wrapper.InternalGetUser)
	}

	return r
}
