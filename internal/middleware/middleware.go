package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logger middleware for request logging
func Logger(next http.Handler) http.Handler {
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  log.New(log.Writer(), "", log.LstdFlags),
		NoColor: false,
	})(next)
}

// Recoverer middleware for panic recovery
func Recoverer(next http.Handler) http.Handler {
	return middleware.Recoverer(next)
}

// Custom logger with structured format
func CustomLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			duration := time.Since(start)
			log.Printf(
				"[%s] %s %s %d %s %s",
				r.Method,
				r.URL.Path,
				r.Proto,
				ww.Status(),
				duration,
				r.RemoteAddr,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID(next http.Handler) http.Handler {
	return middleware.RequestID(next)
}

// Timeout middleware for request timeout
func Timeout(duration time.Duration) func(next http.Handler) http.Handler {
	return middleware.Timeout(duration)
}
