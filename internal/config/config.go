package config

import (
	"log"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort string `mapstructure:"APP_PORT"`

	// Database
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBPassword string `mapstructure:"DB_PASSWORD"`
	DBName     string `mapstructure:"DB_NAME"`

	// Redis
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	// Auth
	JWTSecret     string        `mapstructure:"JWT_SECRET"`
	JWTExpire     time.Duration // Will be parsed manually
	RefreshExpire time.Duration // Will be parsed manually
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "gonotes")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("JWT_SECRET", "supersecretkey")
	viper.SetDefault("JWT_EXPIRE", "15m")
	viper.SetDefault("REFRESH_EXPIRE", "7d")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Parse durations manually
	jwtExpire, err := time.ParseDuration(viper.GetString("JWT_EXPIRE"))
	if err != nil {
		cfg.JWTExpire = 15 * time.Minute
	} else {
		cfg.JWTExpire = jwtExpire
	}

	refreshExpireStr := viper.GetString("REFRESH_EXPIRE")
	refreshExpire, err := parseDurationWithDays(refreshExpireStr)
	if err != nil {
		cfg.RefreshExpire = 7 * 24 * time.Hour
	} else {
		cfg.RefreshExpire = refreshExpire
	}

	return &cfg, nil
}

// parseDurationWithDays parses duration string with support for "d" (days) unit
func parseDurationWithDays(s string) (time.Duration, error) {
	// Handle "d" suffix for days
	if len(s) > 1 && s[len(s)-1] == 'd' {
		daysStr := s[:len(s)-1]
		days, err := strconv.ParseInt(daysStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	// For other units, use standard time.ParseDuration
	return time.ParseDuration(s)
}
