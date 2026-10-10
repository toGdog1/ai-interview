package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application runtime configuration.
type Config struct {
	ServerAddr      string
	MaxUploadBytes  int64
	AIProvider      string // "rule" | "openai" (extensible)
	OpenAIAPIKey    string
	OpenAIBaseURL   string
	OpenAIModel     string
	RequestTimeout  time.Duration
	EnableCORS      bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		ServerAddr:     getEnv("SERVER_ADDR", ":8080"),
		MaxUploadBytes: getEnvInt64("MAX_UPLOAD_BYTES", 10*1024*1024),
		AIProvider:     getEnv("AI_PROVIDER", "rule"),
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:  getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:    getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		RequestTimeout: time.Duration(getEnvInt64("REQUEST_TIMEOUT_SEC", 60)) * time.Second,
		EnableCORS:     getEnv("ENABLE_CORS", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}
