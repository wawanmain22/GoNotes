package utils

import (
	"context"
	"fmt"
	"time"

	"gonotes/internal/config"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis establishes a connection to Redis
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password:     cfg.RedisPassword, // Use password from config
		DB:           0,                 // Default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}

// SetSession stores session data in Redis
func SetSession(rdb *redis.Client, key string, value interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return rdb.Set(ctx, key, value, expiration).Err()
}

// GetSession retrieves session data from Redis
func GetSession(rdb *redis.Client, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	}
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	return val, nil
}

// DeleteSession removes session data from Redis
func DeleteSession(rdb *redis.Client, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return rdb.Del(ctx, key).Err()
}

// SessionExists checks if a session exists in Redis
func SessionExists(rdb *redis.Client, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// SetRefreshToken stores refresh token with user ID mapping
func SetRefreshToken(rdb *redis.Client, tokenID, userID string, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s", tokenID)
	return rdb.Set(ctx, key, userID, expiration).Err()
}

// GetUserIDFromRefreshToken retrieves user ID from refresh token
func GetUserIDFromRefreshToken(rdb *redis.Client, tokenID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s", tokenID)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Token does not exist or expired
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user ID from refresh token: %w", err)
	}

	return val, nil
}

// InvalidateRefreshToken removes refresh token from Redis
func InvalidateRefreshToken(rdb *redis.Client, tokenID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("refresh_token:%s", tokenID)
	return rdb.Del(ctx, key).Err()
}

// SetProfileCache stores user profile in Redis cache
func SetProfileCache(rdb *redis.Client, userID string, profileData interface{}, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("profile:%s", userID)
	return rdb.Set(ctx, key, profileData, expiration).Err()
}

// GetProfileCache retrieves user profile from Redis cache
func GetProfileCache(rdb *redis.Client, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("profile:%s", userID)
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Profile not cached
	}
	if err != nil {
		return "", fmt.Errorf("failed to get profile from cache: %w", err)
	}

	return val, nil
}

// InvalidateProfileCache removes user profile from Redis cache
func InvalidateProfileCache(rdb *redis.Client, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("profile:%s", userID)
	return rdb.Del(ctx, key).Err()
}
