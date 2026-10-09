package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/TimofeyChernyshev/LCT-hack-2/pkg/jwtx"

	"github.com/gin-gonic/gin"
)

type ctxKey string

const (
	ctxUserID       ctxKey = "auth.user_id"
	ctxUserRole     ctxKey = "auth.role"
	ctxUserEmail    ctxKey = "auth.email"
	ctxUserVerified ctxKey = "auth.email_verified"
)

type UserInfo struct {
	ID            string
	Role          string
	Email         string
	EmailVerified bool
}

func UserFromContext(ctx context.Context) (UserInfo, bool) {
	id, _ := ctx.Value(ctxUserID).(string)
	if id == "" {
		return UserInfo{}, false
	}
	role, _ := ctx.Value(ctxUserRole).(string)
	email, _ := ctx.Value(ctxUserEmail).(string)
	verified, _ := ctx.Value(ctxUserVerified).(bool)
	return UserInfo{ID: id, Role: role, Email: email, EmailVerified: verified}, true
}

// ---------- RequestID ----------

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// ---------- Logger ----------

func Logger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		base.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"dur_ms", time.Since(start).Milliseconds(),
		)
	}
}

// ---------- JWT Auth ----------

func JWTAuth(signer jwtx.Signer) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "missing bearer token"})
			return
		}
		claims, err := signer.Parse(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "unauthorized", "message": "invalid token"})
			return
		}
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, ctxUserID, claims.Subject)
		ctx = context.WithValue(ctx, ctxUserRole, claims.Role)
		ctx = context.WithValue(ctx, ctxUserEmail, claims.Email)
		ctx = context.WithValue(ctx, ctxUserVerified, claims.EmailVerified)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// ---------- RequireVerifiedEmail ----------

func RequireVerifiedEmail() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := UserFromContext(c.Request.Context())
		if !ok || !u.EmailVerified {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "email_not_verified", "message": "email not verified"})
			return
		}
		c.Next()
	}
}

// ---------- Rate limiter ----------

type rateLimiter struct {
	mu      sync.Mutex
	rps     float64
	burst   int
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = 20
	}
	if burst <= 0 {
		burst = 40
	}
	rl := &rateLimiter{rps: rps, burst: burst, buckets: make(map[string]*bucket)}
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "rate_limited", "message": "too many requests"})
			return
		}
		c.Next()
	}
}

func (l *rateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: float64(l.burst) - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = minf(float64(l.burst), b.tokens+elapsed*l.rps)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
