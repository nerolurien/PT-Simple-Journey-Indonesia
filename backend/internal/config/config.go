package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Env                string
	DatabaseURL        string
	JWTSecret          string
	CORSAllowedOrigins []string
	RateLimitRPS       float64
	RateLimitBurst     int
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := getEnv("DB_HOST", "localhost")
		dbPort := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		pass := os.Getenv("DB_PASSWORD")
		name := getEnv("DB_NAME", "simple_journey_helpdesk")
		ssl := getEnv("DB_SSLMODE", "disable")

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			url.QueryEscape(user),
			url.QueryEscape(pass),
			host,
			dbPort,
			name,
			ssl,
		)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" && env == "production" {
		return nil, fmt.Errorf("JWT_SECRET is required in production")
	}
	if jwtSecret == "" {
		jwtSecret = "dev-insecure-jwt-secret-key-change-in-production"
	}

	corsRaw := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000")
	origins := strings.Split(corsRaw, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	rateRPS, err := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "5"), 64)
	if err != nil {
		rateRPS = 5.0
	}

	rateBurst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "10"))
	if err != nil {
		rateBurst = 10
	}

	return &Config{
		Port:               port,
		Env:                env,
		DatabaseURL:        dbURL,
		JWTSecret:          jwtSecret,
		CORSAllowedOrigins: origins,
		RateLimitRPS:       rateRPS,
		RateLimitBurst:     rateBurst,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
