package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Global rate limits (per IP)
	GlobalRequestsPerMinute int
	GlobalBurstSize         int

	// Authenticated user rate limits
	UserRequestsPerMinute int
	UserBurstSize         int

	// Endpoint-specific rate limits
	AuthEndpointRequestsPerMinute int
	AuthEndpointBurstSize         int

	// Redis client for distributed rate limiting
	RedisClient *redis.Client
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig(redisClient *redis.Client) *RateLimitConfig {
	return &RateLimitConfig{
		GlobalRequestsPerMinute:       100,
		GlobalBurstSize:               20,
		UserRequestsPerMinute:         300,
		UserBurstSize:                 50,
		AuthEndpointRequestsPerMinute: 10,
		AuthEndpointBurstSize:         5,
		RedisClient:                   redisClient,
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config *RateLimitConfig) func(next http.Handler) http.Handler {
	if config.RedisClient == nil {
		// If no Redis client, skip rate limiting
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Determine rate limit parameters based on request
			key, limit, burst := getRateLimitParams(r, config)

			// Check rate limit
			allowed, resetTime, err := checkRateLimit(config.RedisClient, key, limit, burst)
			if err != nil {
				// Log error but don't block request if Redis is down
				fmt.Printf("Rate limit check failed: %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				// Set rate limit headers
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
				w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(resetTime).Seconds())))

				// Return rate limit error
				sendRateLimitError(w)
				return
			}

			// Set rate limit headers for successful requests
			remaining, _ := getRemainingRequests(config.RedisClient, key, limit)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			next.ServeHTTP(w, r)
		})
	}
}

// getRateLimitParams determines the rate limit key and parameters for a request
func getRateLimitParams(r *http.Request, config *RateLimitConfig) (string, int, int) {
	clientIP := getClientIP(r)

	// Check if this is an auth endpoint
	if isAuthEndpoint(r.URL.Path) {
		key := fmt.Sprintf("rate_limit:auth:%s", clientIP)
		return key, config.AuthEndpointRequestsPerMinute, config.AuthEndpointBurstSize
	}

	// Check if user is authenticated
	if userID, ok := GetUserID(r); ok {
		key := fmt.Sprintf("rate_limit:user:%s", userID.String())
		return key, config.UserRequestsPerMinute, config.UserBurstSize
	}

	// Default to IP-based rate limiting
	key := fmt.Sprintf("rate_limit:ip:%s", clientIP)
	return key, config.GlobalRequestsPerMinute, config.GlobalBurstSize
}

// isAuthEndpoint checks if the request is to an authentication endpoint
func isAuthEndpoint(path string) bool {
	authPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
	}

	for _, authPath := range authPaths {
		if path == authPath {
			return true
		}
	}

	return false
}

// checkRateLimit implements sliding window rate limiting using Redis
func checkRateLimit(redisClient *redis.Client, key string, limit, burst int) (bool, time.Time, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-time.Minute)

	// Redis Lua script for atomic rate limiting check
	script := `
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local burst = tonumber(ARGV[4])
		
		-- Remove expired entries
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)
		
		-- Count current requests in window
		local current = redis.call('ZCARD', key)
		
		-- Check if we can make the request
		if current < limit then
			-- Add current request
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, 60)
			return {1, current + 1}
		else
			return {0, current}
		end
	`

	result, err := redisClient.Eval(ctx, script, []string{key},
		windowStart.Unix(), now.Unix(), limit, burst).Result()
	if err != nil {
		return false, now.Add(time.Minute), err
	}

	resultSlice := result.([]interface{})
	allowed := resultSlice[0].(int64) == 1

	// Calculate reset time (next minute boundary)
	resetTime := now.Truncate(time.Minute).Add(time.Minute)

	return allowed, resetTime, nil
}

// getRemainingRequests gets the number of remaining requests in the current window
func getRemainingRequests(redisClient *redis.Client, key string, limit int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	current, err := redisClient.ZCard(ctx, key).Result()
	if err != nil {
		return limit, err
	}

	remaining := limit - int(current)
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// sendRateLimitError sends a rate limit exceeded error response
func sendRateLimitError(w http.ResponseWriter) {
	response := map[string]interface{}{
		"status":  "error",
		"code":    429,
		"message": "Rate limit exceeded. Please try again later.",
		"error":   "Too many requests",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(response)
}

// IPWhitelistMiddleware creates middleware for IP whitelisting
func IPWhitelistMiddleware(whitelistedIPs []string) func(next http.Handler) http.Handler {
	if len(whitelistedIPs) == 0 {
		// If no whitelist, allow all
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	ipMap := make(map[string]bool)
	for _, ip := range whitelistedIPs {
		ipMap[ip] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !ipMap[clientIP] {
				response := map[string]interface{}{
					"status":  "error",
					"code":    403,
					"message": "Access forbidden from this IP address",
					"error":   "IP not whitelisted",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DDoSProtectionMiddleware creates basic DDoS protection middleware
func DDoSProtectionMiddleware(redisClient *redis.Client) func(next http.Handler) http.Handler {
	if redisClient == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			// Check for suspicious patterns
			if isSuspiciousRequest(redisClient, clientIP, r) {
				response := map[string]interface{}{
					"status":  "error",
					"code":    429,
					"message": "Suspicious activity detected. Access temporarily blocked.",
					"error":   "DDoS protection triggered",
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isSuspiciousRequest checks for suspicious request patterns
func isSuspiciousRequest(redisClient *redis.Client, clientIP string, r *http.Request) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Check request frequency (more than 20 requests in 10 seconds)
	key := fmt.Sprintf("ddos_protection:%s", clientIP)
	current, err := redisClient.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	if current == 1 {
		redisClient.Expire(ctx, key, 10*time.Second)
	}

	// If more than 20 requests in 10 seconds, consider suspicious
	if current > 20 {
		// Extend the block time
		redisClient.Expire(ctx, key, 60*time.Second)
		return true
	}

	return false
}
