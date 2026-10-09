package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------- net/http ----------

type ctxKey string

const (
	ctxRequestID ctxKey = "request_id"
	ctxLogger    ctxKey = "logger"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

func Logger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID, _ := r.Context().Value(ctxRequestID).(string)
			l := base.With("request_id", reqID)
			ww := &wrapWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(ww, r.WithContext(withLogger(r.Context(), l)))
			l.Info("http", "method", r.Method, "path", r.URL.Path, "status", ww.status,
				"dur_ms", time.Since(start).Milliseconds())
		})
	}
}

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic", "err", rec, "stack", string(debug.Stack()))
				WriteError(w, http.StatusInternalServerError, "internal_error", "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type wrapWriter struct {
	http.ResponseWriter
	status int
}

func (w *wrapWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// ---------- gin ----------

func GinRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		c.Header("X-Request-ID", id)
		c.Set(string(ctxRequestID), id)
		c.Next()
	}
}

func GinLogger(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID, _ := c.Get(string(ctxRequestID))
		l, _ := reqID.(string)
		logger := base.With("request_id", l)
		c.Request = c.Request.WithContext(withLogger(c.Request.Context(), logger))

		c.Next()

		logger.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"dur_ms", time.Since(start).Milliseconds(),
		)
	}
}

func GinRecoverer() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, rec any) {
		slog.Error("panic", "err", rec)
		GinError(c, http.StatusInternalServerError, "internal_error", "internal error")
	})
}

// ---------- helpers ----------

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxRequestID, id)
}

func withLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger, l)
}
