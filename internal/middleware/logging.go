package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// LoggingMiddleware creates a structured logging middleware
func LoggingMiddleware() func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{})
}

// StructuredLogger implements the middleware.LogFormatter interface
type StructuredLogger struct{}

// NewLogEntry creates a new log entry for each request
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{
		request: r,
	}

	// Extract user information if available
	if userID, ok := GetUserID(r); ok {
		entry.userID = &userID
	}

	// Extract client information
	entry.userAgent = r.Header.Get("User-Agent")
	entry.clientIP = getClientIP(r)

	return entry
}

// StructuredLoggerEntry represents a single log entry
type StructuredLoggerEntry struct {
	request   *http.Request
	userID    *uuid.UUID
	userAgent string
	clientIP  string
}

// Write logs the completed request
func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	method := l.request.Method
	uri := l.request.RequestURI
	proto := l.request.Proto

	// Determine log level based on status code
	level := "INFO"
	if status >= 400 && status < 500 {
		level = "WARN"
	} else if status >= 500 {
		level = "ERROR"
	}

	// Build log message
	logMsg := fmt.Sprintf("[%s] %s %s %s - %d - %dB - %v",
		level,
		method,
		uri,
		proto,
		status,
		bytes,
		elapsed,
	)

	// Add user context if available
	if l.userID != nil {
		logMsg += fmt.Sprintf(" - UserID: %s", l.userID.String())
	}

	// Add client info
	logMsg += fmt.Sprintf(" - IP: %s", l.clientIP)

	// Add user agent for non-browser requests
	if l.userAgent != "" && !strings.Contains(strings.ToLower(l.userAgent), "browser") {
		logMsg += fmt.Sprintf(" - UA: %s", l.userAgent)
	}

	// Print the log (in production, this would go to a proper logger)
	fmt.Println(logMsg)
}

// Panic logs panic information
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	method := l.request.Method
	uri := l.request.RequestURI

	logMsg := fmt.Sprintf("[PANIC] %s %s - %v",
		method,
		uri,
		v,
	)

	if l.userID != nil {
		logMsg += fmt.Sprintf(" - UserID: %s", l.userID.String())
	}

	logMsg += fmt.Sprintf(" - IP: %s", l.clientIP)

	// Print panic log with stack trace
	fmt.Printf("%s\nStack Trace:\n%s\n", logMsg, string(stack))
}

// AuditLogMiddleware creates an audit logging middleware for sensitive operations
func AuditLogMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only log sensitive operations
			if shouldAuditLog(r) {
				logAuditEvent(r, "REQUEST_START")
			}

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process request
			next.ServeHTTP(ww, r)

			// Log completion for sensitive operations
			if shouldAuditLog(r) {
				logAuditEvent(r, "REQUEST_COMPLETE", map[string]interface{}{
					"status_code": ww.Status(),
					"bytes_sent":  ww.BytesWritten(),
				})
			}
		})
	}
}

// shouldAuditLog determines if a request should be audit logged
func shouldAuditLog(r *http.Request) bool {
	sensitiveEndpoints := []string{
		"/api/v1/auth/",
		"/api/v1/user/profile",
		"/api/v1/user/sessions",
	}

	for _, endpoint := range sensitiveEndpoints {
		if strings.HasPrefix(r.URL.Path, endpoint) {
			return true
		}
	}

	return false
}

// logAuditEvent logs an audit event
func logAuditEvent(r *http.Request, event string, extra ...map[string]interface{}) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	userID := "anonymous"

	if uid, ok := GetUserID(r); ok {
		userID = uid.String()
	}

	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	method := r.Method
	path := r.URL.Path

	auditLog := fmt.Sprintf("[AUDIT] %s - %s - User: %s - %s %s - IP: %s",
		timestamp,
		event,
		userID,
		method,
		path,
		clientIP,
	)

	// Add extra information if provided
	if len(extra) > 0 {
		for key, value := range extra[0] {
			auditLog += fmt.Sprintf(" - %s: %v", key, value)
		}
	}

	// Add user agent for important operations
	if userAgent != "" {
		auditLog += fmt.Sprintf(" - UA: %s", userAgent)
	}

	// Print audit log (in production, this would go to a secure audit log system)
	fmt.Println(auditLog)
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// API-specific headers
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			next.ServeHTTP(w, r)
		})
	}
}
