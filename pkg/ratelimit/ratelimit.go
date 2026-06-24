package ratelimit

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	rate   float64 // tokens per second
	burst  int
	mu     sync.Mutex
	ips    map[string]*client
	stopCh chan struct{}
}

type client struct {
	tokens    float64
	lastCheck time.Time
}

func NewLimiter(rate float64, burst int) *Limiter {
	l := &Limiter{
		rate:   rate,
		burst:  burst,
		ips:    make(map[string]*client),
		stopCh: make(chan struct{}),
	}
	go l.janitor(1 * time.Minute)
	return l
}

func (l *Limiter) Close() {
	close(l.stopCh)
}

func (l *Limiter) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for ip, c := range l.ips {
				// Evict if not seen in last 5 minutes
				if now.Sub(c.lastCheck) > 5*time.Minute {
					delete(l.ips, ip)
				}
			}
			l.mu.Unlock()
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		l.mu.Lock()
		c, exists := l.ips[ip]
		now := time.Now()
		if !exists {
			c = &client{
				tokens:    float64(l.burst),
				lastCheck: now,
			}
			l.ips[ip] = c
		} else {
			elapsed := now.Sub(c.lastCheck).Seconds()
			c.tokens += elapsed * l.rate
			if c.tokens > float64(l.burst) {
				c.tokens = float64(l.burst)
			}
			c.lastCheck = now
		}

		if c.tokens >= 1 {
			c.tokens -= 1
			l.mu.Unlock()
			next.ServeHTTP(w, r)
		} else {
			l.mu.Unlock()
			w.Header().Set("Retry-After", "1")
			http.Error(w, "Rate limit exceeded. Too many requests.", http.StatusTooManyRequests)
		}
	})
}

