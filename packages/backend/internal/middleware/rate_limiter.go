package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter maneja el rate limiting por IP
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getVisitor obtiene o crea un rate limiter para una IP específica
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter
}

// cleanupVisitors limpia visitantes antiguos (llamar periódicamente)
func (rl *RateLimiter) cleanupVisitors() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Limpiar visitantes que no han hecho requests en la última hora
	for ip, limiter := range rl.visitors {
		if limiter.Allow() {
			// Si el limiter permite una request, significa que no está en burst
			// y podemos considerar limpiarlo si no se ha usado recientemente
			delete(rl.visitors, ip)
		}
	}
}

// StartCleanup inicia la limpieza periódica de visitantes
func (rl *RateLimiter) StartCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rl.cleanupVisitors()
		}
	}()
}

// getIP extrae la IP real del cliente considerando proxies
func getIP(r *http.Request) string {
	// Verificar headers de proxy
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For puede contener múltiples IPs, tomar la primera
		if host, _, err := net.SplitHostPort(ip); err == nil {
			return host
		}
		return ip
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Usar RemoteAddr como fallback
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// Middleware retorna un middleware HTTP que aplica rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := rl.getVisitor(ip)

		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// CreateAuthRateLimiter crea un rate limiter específico para endpoints de autenticación
// Permite 5 requests por minuto por IP
func CreateAuthRateLimiter() *RateLimiter {
	rl := NewRateLimiter(rate.Every(12*time.Second), 5) // 5 requests per minute
	rl.StartCleanup(10 * time.Minute)                   // Limpiar cada 10 minutos
	return rl
}
