package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
)

var (
	limiters  sync.Map
	rateEvery = rate.Every(1 * time.Second)
	rateBurst = 2
)

// rateMW throttles requests per client IP.
func rateMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
		l, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rateEvery, rateBurst))
		if !l.(*rate.Limiter).Allow() {
			log.Warn("rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path,
			)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// logMW logs each request once with the trace ID from the surrounding
// OpenTelemetry span; status and duration live on the span itself.
func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc := trace.SpanContextFromContext(r.Context())
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"trace.id", sc.TraceID().String(),
		)
		next.ServeHTTP(w, r)
	})
}
