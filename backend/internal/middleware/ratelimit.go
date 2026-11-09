package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	tokens     float64
	lastRefill time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64
	capacity float64
}

func NewIPRateLimiter(rate float64, capacity int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		capacity: float64(capacity),
	}

	go limiter.cleanupStale(3 * time.Minute)
	return limiter
}

func (l *IPRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	v, exists := l.visitors[ip]
	if !exists {
		l.visitors[ip] = &visitor{
			tokens:     l.capacity - 1,
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(v.lastRefill).Seconds()
	v.tokens += elapsed * l.rate
	if v.tokens > l.capacity {
		v.tokens = l.capacity
	}
	v.lastRefill = now

	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

func (l *IPRateLimiter) cleanupStale(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, v := range l.visitors {
			if now.Sub(v.lastRefill) > interval {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func RateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !limiter.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"terlalu banyak permintaan, coba lagi nanti"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
