package httpx

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// SimpleRateLimiter — token bucket per IP. Для MVP достаточно.
type SimpleRateLimiter struct {
	mu      sync.Mutex
	rps     float64
	burst   int
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewSimpleRateLimiter(rps float64, burst int) *SimpleRateLimiter {
	return &SimpleRateLimiter{rps: rps, burst: burst, buckets: make(map[string]*bucket)}
}

func (l *SimpleRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *SimpleRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{tokens: float64(l.burst) - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens = min(float64(l.burst), b.tokens+elapsed*l.rps)
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func clientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		return h
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
