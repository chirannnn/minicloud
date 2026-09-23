package config

import (
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	Environment   string
	Host          string
	Port          string
	ConsoleOrigin string
	OTLPEndpoint  string
	ServiceName   string
	DatabaseURL   string
	Dependencies  map[string]string
}

func Load() (Config, error) {
	c := Config{
		Environment:   env("APP_ENV", "development"),
		Host:          env("API_HOST", "0.0.0.0"),
		Port:          env("API_PORT", "8080"),
		ConsoleOrigin: env("CONSOLE_ORIGIN", "http://localhost:3000"),
		OTLPEndpoint:  env("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		ServiceName:   env("OTEL_SERVICE_NAME", "minicloud-api"),
		DatabaseURL:   fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", env("POSTGRES_USER", "minicloud"), env("POSTGRES_PASSWORD", "minicloud_dev_password"), env("POSTGRES_HOST", "localhost"), env("POSTGRES_PORT", "5432"), env("POSTGRES_DATABASE", "minicloud")),
		Dependencies: map[string]string{
			"postgres": net.JoinHostPort(env("POSTGRES_HOST", "localhost"), env("POSTGRES_PORT", "5432")),
			"redis":    net.JoinHostPort(env("REDIS_HOST", "localhost"), env("REDIS_PORT", "6379")),
			"kafka":    firstBroker(env("KAFKA_BROKERS", "localhost:9092")),
			"minio":    env("MINIO_ENDPOINT", "localhost:9000"),
		},
	}
	if _, err := net.LookupPort("tcp", c.Port); err != nil {
		return Config{}, fmt.Errorf("API_PORT must be a valid TCP port: %w", err)
	}
	if c.ConsoleOrigin == "" {
		return Config{}, fmt.Errorf("CONSOLE_ORIGIN must not be empty")
	}
	return c, nil
}

func (c Config) Address() string { return net.JoinHostPort(c.Host, c.Port) }
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func firstBroker(value string) string { return strings.TrimSpace(strings.Split(value, ",")[0]) }
