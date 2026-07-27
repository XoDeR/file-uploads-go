package ratelimit

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limiter implements a fixed-window rate limiter for uploads.
type Limiter struct {
	requests    map[string]*clientLimit
	maxRequests int
	windowSize  time.Duration
	mu          sync.Mutex
}

type clientLimit struct {
	count     int
	windowEnd time.Time
}

// New creates a new rate limiter.
func New(maxRequests int, windowSize time.Duration) *Limiter {
	rl := &Limiter{
		requests:    make(map[string]*clientLimit),
		maxRequests: maxRequests,
		windowSize:  windowSize,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given IP should be allowed.
func (rl *Limiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.requests[ip]

	if !exists || now.After(limit.windowEnd) {
		rl.requests[ip] = &clientLimit{
			count:     1,
			windowEnd: now.Add(rl.windowSize),
		}
		return true
	}

	if limit.count >= rl.maxRequests {
		return false
	}

	limit.count++
	return true
}

func (rl *Limiter) cleanup() {
	ticker := time.NewTicker(rl.windowSize)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, limit := range rl.requests {
			if now.After(limit.windowEnd) {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware wraps an HTTP handler with rate limiting.
func (rl *Limiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ip = strings.TrimSpace(strings.Split(xff, ",")[0])
		}

		if !rl.Allow(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
