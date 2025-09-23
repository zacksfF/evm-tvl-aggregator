// internal/api/middleware.go
package api

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "golang.org/x/time/rate"
)

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(rps int) func(http.Handler) http.Handler {
    limiters := make(map[string]*rate.Limiter)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ip := r.RemoteAddr
            
            limiter, exists := limiters[ip]
            if !exists {
                limiter = rate.NewLimiter(rate.Limit(rps), rps)
                limiters[ip] = limiter
            }
            
            if !limiter.Allow() {
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Log request
        fmt.Printf("[%s] %s %s\n", 
            start.Format("2006-01-02 15:04:05"),
            r.Method,
            r.URL.Path,
        )
        
        next.ServeHTTP(w, r)
        
        // Log response time
        fmt.Printf("  └─ Completed in %v\n", time.Since(start))
    })
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            r = r.WithContext(ctx)
            next.ServeHTTP(w, r)
        })
    }
}