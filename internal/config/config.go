package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Application Settings
	Host     string
	Port     string
	LogLevel string
	Debug    bool

	// Kafka Configuration
	KafkaBrokers       string
	KafkaConsumerGroup string
	KafkaTopics        KafkaTopics

	// ClickHouse Configuration
	ClickHouse ClickHouseConfig

	// Redis Configuration
	Redis RedisConfig

	// Monitoring and Metrics
	EnableMetrics bool
	EnableTracing bool
	MetricsPort   string

	// Parser Configuration
	Parser ParserConfig

	// Security
	JWTSecret string
	APIKey    string

	// Development Settings
	HotReload       bool
	EnableProfiling bool
}

// KafkaTopics holds Kafka topic names
type KafkaTopics struct {
	Events  string
	Errors  string
	Metrics string
}

// ClickHouseConfig holds ClickHouse database configuration
type ClickHouseConfig struct {
	Host           string
	Port           int
	HTTPPort       int
	Database       string
	User           string
	Password       string
	MaxConnections int
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ParserConfig holds parser-specific configuration
type ParserConfig struct {
	MaxInputSize int64
	Timeout      time.Duration
	Workers      int
}

// Load returns the application configuration loaded from environment variables
func Load() *Config {
	return &Config{
		// Application Settings
		Host:     getEnv("HOST", "localhost"),
		Port:     getEnv("PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Debug:    getBoolEnv("DEBUG", false),

		// Kafka Configuration
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "hoe_parser_group"),
		KafkaTopics: KafkaTopics{
			Events:  getEnv("KAFKA_TOPICS_EVENTS", "events"),
			Errors:  getEnv("KAFKA_TOPICS_ERRORS", "errors"),
			Metrics: getEnv("KAFKA_TOPICS_METRICS", "metrics"),
		},

		// ClickHouse Configuration
		ClickHouse: ClickHouseConfig{
			Host:           getEnv("CLICKHOUSE_HOST", "localhost"),
			Port:           getIntEnv("CLICKHOUSE_PORT", 9000),
			HTTPPort:       getIntEnv("CLICKHOUSE_HTTP_PORT", 8123),
			Database:       getEnv("CLICKHOUSE_DATABASE", "hoe_parser"),
			User:           getEnv("CLICKHOUSE_USER", "admin"),
			Password:       getEnv("CLICKHOUSE_PASSWORD", "password"),
			MaxConnections: getIntEnv("CLICKHOUSE_MAX_CONNECTIONS", 10),
		},

		// Redis Configuration
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getIntEnv("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", "redispassword"),
			DB:       getIntEnv("REDIS_DB", 0),
		},

		// Monitoring and Metrics
		EnableMetrics: getBoolEnv("ENABLE_METRICS", true),
		EnableTracing: getBoolEnv("ENABLE_TRACING", false),
		MetricsPort:   getEnv("METRICS_PORT", "9090"),

		// Parser Configuration
		Parser: ParserConfig{
			MaxInputSize: getInt64Env("PARSER_MAX_INPUT_SIZE", 1048576),
			Timeout:      getDurationEnv("PARSER_TIMEOUT", 60*time.Second),
			Workers:      getIntEnv("PARSER_WORKERS", 4),
		},

		// Security
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		APIKey:    getEnv("API_KEY", "your-api-key-here"),

		// Development Settings
		HotReload:       getBoolEnv("HOT_RELOAD", false),
		EnableProfiling: getBoolEnv("ENABLE_PROFILING", false),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getBoolEnv gets a boolean environment variable with a fallback value
func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return fallback
}

// getIntEnv gets an integer environment variable with a fallback value
func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

// getInt64Env gets an int64 environment variable with a fallback value
func getInt64Env(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return fallback
}

// getDurationEnv gets a duration environment variable with a fallback value
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return fallback
}
