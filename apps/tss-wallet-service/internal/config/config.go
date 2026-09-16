package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is deliberately environment-driven so local development and Docker
// deployment use the same executable.
type Config struct {
	HTTPAddr          string
	ShutdownTimeout   time.Duration
	StaticDir         string
	MySQLDSN          string
	MySQLMaxOpenConns int
	MySQLMaxIdleConns int
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
}

func Load() Config {
	return Config{
		HTTPAddr:          env("HTTP_ADDR", ":8080"),
		ShutdownTimeout:   envDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		StaticDir:         env("STATIC_DIR", "./web/admin-console"),
		MySQLDSN:          strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		MySQLMaxOpenConns: envInt("MYSQL_MAX_OPEN_CONNS", 10),
		MySQLMaxIdleConns: envInt("MYSQL_MAX_IDLE_CONNS", 5),
		RedisAddr:         strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           envInt("REDIS_DB", 0),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err == nil {
		return value
	}
	return fallback
}
func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(strings.TrimSpace(os.Getenv(key)))
	if err == nil {
		return value
	}
	return fallback
}
