package config

import (
	"os"
	"strconv"
)

// Config holds server configuration (local disk by default).
type Config struct {
	UploadDir string
	Port      string
	MaxSize   int64
	CORSOrigin string
}

// Load reads configuration from environment variables.
func Load() Config {
	cfg := Config{
		UploadDir:  envOr("UPLOAD_DIR", "./uploads"),
		Port:       envOr("PORT", "8080"),
		CORSOrigin: envOr("CORS_ORIGIN", "http://localhost:5173"),
		MaxSize:    100 * 1024 * 1024,
	}
	if v := os.Getenv("MAX_UPLOAD_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MaxSize = n
		}
	}
	return cfg
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
