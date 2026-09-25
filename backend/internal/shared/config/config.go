package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	environmentProduction = "production"
	dotenvSearchDepth     = 4
)

type Config struct {
	App      App
	Database Database
	Token    Token
	Cookie   Cookie
	CORS     CORS
	Score    Score
}

type App struct {
	Environment string
	Port        string
}

type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

type Token struct {
	AccessSecret string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	Issuer       string
}

type Cookie struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite string
}

type CORS struct {
	AllowedOrigins []string
}

// Score configures the internal Python scoring service (Moses's /score, contract §8).
// BaseURL empty means the service is not wired yet; the site slice falls back to a
// deterministic stub so /api/site/evaluate still works end to end for the demo.
type Score struct {
	BaseURL string
	Timeout time.Duration
}

func (a App) IsProduction() bool {
	return a.Environment == environmentProduction
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.TimeZone,
	)
}

func Load() (Config, error) {
	loadDotenv()

	accessSecret, err := requiredEnv("JWT_ACCESS_SECRET")
	if err != nil {
		return Config{}, err
	}

	databasePassword, err := requiredEnv("POSTGRES_PASSWORD")
	if err != nil {
		return Config{}, err
	}

	accessTTL, err := durationEnv("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	refreshTTL, err := durationEnv("JWT_REFRESH_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	secureCookie, err := boolEnv("COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}

	scoreTimeout, err := durationEnv("SCORE_SERVICE_TIMEOUT", 5*time.Second)
	if err != nil {
		return Config{}, err
	}

	return Config{
		App: App{
			Environment: env("APP_ENV", "development"),
			Port:        env("APP_PORT", "8080"),
		},
		Database: Database{
			Host:     env("POSTGRES_HOST", "localhost"),
			Port:     env("POSTGRES_PORT", "5432"),
			User:     env("POSTGRES_USER", "postgres"),
			Password: databasePassword,
			Name:     env("POSTGRES_DB", "bangunin_aja"),
			SSLMode:  env("POSTGRES_SSLMODE", "disable"),
			TimeZone: env("POSTGRES_TIMEZONE", "Asia/Jakarta"),
		},
		Token: Token{
			AccessSecret: accessSecret,
			AccessTTL:    accessTTL,
			RefreshTTL:   refreshTTL,
			Issuer:       env("JWT_ISSUER", "bangunin-aja"),
		},
		Cookie: Cookie{
			Name:     env("REFRESH_COOKIE_NAME", "bangunin_refresh_token"),
			Path:     env("REFRESH_COOKIE_PATH", "/api/auth"),
			Domain:   env("REFRESH_COOKIE_DOMAIN", ""),
			Secure:   secureCookie,
			SameSite: env("REFRESH_COOKIE_SAMESITE", "lax"),
		},
		CORS: CORS{
			AllowedOrigins: listEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Score: Score{
			BaseURL: env("SCORE_SERVICE_URL", ""),
			Timeout: scoreTimeout,
		},
	}, nil
}

func loadDotenv() {
	path := ".env"
	for depth := 0; depth < dotenvSearchDepth; depth++ {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			return
		}
		path = "../" + path
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func requiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("environment variable %s is required", key)
	}
	return value, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s must be a duration such as 15m: %w", key, err)
	}
	return parsed, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("environment variable %s must be true or false: %w", key, err)
	}
	return parsed, nil
}

func listEnv(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}

	if len(values) == 0 {
		return fallback
	}
	return values
}
