package upload

import (
	"github.com/file-uploads-go/backend/pkg/upload/storage"
	"github.com/file-uploads-go/backend/pkg/upload/validation"
)

// Config holds upload service configuration.
type Config struct {
	UploadDir string
	MaxSize   int64
}

// DefaultConfig returns sensible defaults for local development.
func DefaultConfig() Config {
	return Config{
		UploadDir: "./uploads",
		MaxSize:   100 * 1024 * 1024,
	}
}

// Options configures an UploadService.
type Options struct {
	Config    Config
	Storage   storage.Storage
	Validator *validation.FileValidator
}
