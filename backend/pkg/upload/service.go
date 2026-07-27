package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/file-uploads-go/backend/pkg/upload/chunked"
	"github.com/file-uploads-go/backend/pkg/upload/progress"
	"github.com/file-uploads-go/backend/pkg/upload/ratelimit"
	"github.com/file-uploads-go/backend/pkg/upload/storage"
	"github.com/file-uploads-go/backend/pkg/upload/validation"
)

// Service combines upload handling capabilities.
type Service struct {
	validator       *validation.FileValidator
	progressTracker *progress.Tracker
	chunkedManager  *chunked.Manager
	rateLimiter     *ratelimit.Limiter
	store           storage.Storage
	uploadDir       string
}

// NewService creates a new upload service.
func NewService(opts Options) (*Service, error) {
	cfg := opts.Config
	if cfg.UploadDir == "" {
		cfg.UploadDir = "./uploads"
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 * 1024 * 1024
	}

	validator := opts.Validator
	if validator == nil {
		validator = validation.DefaultValidator()
		validator.MaxFileSize = cfg.MaxSize
	}

	store := opts.Storage
	if store == nil {
		store = storage.NewLocal(cfg.UploadDir)
	}

	return &Service{
		validator:       validator,
		progressTracker: progress.NewTracker(),
		chunkedManager:  chunked.NewManager(cfg.UploadDir),
		rateLimiter:     ratelimit.New(100, time.Hour),
		store:           store,
		uploadDir:       cfg.UploadDir,
	}, nil
}

// ProgressTracker returns the progress tracker.
func (s *Service) ProgressTracker() *progress.Tracker {
	return s.progressTracker
}

// ChunkedManager returns the chunked upload manager.
func (s *Service) ChunkedManager() *chunked.Manager {
	return s.chunkedManager
}

// RateLimiter returns the rate limiter.
func (s *Service) RateLimiter() *ratelimit.Limiter {
	return s.rateLimiter
}

// Validator returns the file validator.
func (s *Service) Validator() *validation.FileValidator {
	return s.validator
}

// HandleStream processes a streaming multipart file upload.
func (s *Service) HandleStream(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.validator.MaxFileSize)

	uploadID := r.Header.Get("X-Upload-ID")
	if uploadID == "" {
		uploadID = generateID()
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "Invalid multipart request", http.StatusBadRequest)
		return
	}

	var uploaded []map[string]string

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "Error reading request", http.StatusBadRequest)
			return
		}

		if part.FileName() == "" {
			continue
		}

		filename, err := s.processUploadPart(r.Context(), uploadID, part)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		uploaded = append(uploaded, map[string]string{
			"filename":  filename,
			"upload_id": uploadID,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Upload-ID", uploadID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"files":     uploaded,
		"upload_id": uploadID,
		"message":   fmt.Sprintf("Successfully uploaded %d file(s)", len(uploaded)),
	})
}

func (s *Service) processUploadPart(ctx context.Context, uploadID string, part *multipart.Part) (string, error) {
	filename := validation.SanitizeFilename(part.FileName())

	header := make([]byte, 512)
	n, err := io.ReadFull(part, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("error reading file header: %w", err)
	}
	header = header[:n]

	if err := s.validator.ValidateHeaderBytes(filename, header); err != nil {
		return "", err
	}

	fullReader := io.MultiReader(bytes.NewReader(header), part)

	var contentLength int64
	if cl := part.Header.Get("Content-Length"); cl != "" {
		fmt.Sscanf(cl, "%d", &contentLength)
	}

	prog := s.progressTracker.StartTracking(uploadID, filename, contentLength)
	progressReader := progress.NewReader(fullReader, contentLength, func(bytesRead, totalBytes int64, _ float64) {
		prog.Update(bytesRead)
		_ = totalBytes
	})

	contentType := part.Header.Get("Content-Type")
	path, err := s.store.Save(ctx, filename, progressReader, contentType)
	if err != nil {
		prog.Fail()
		return "", err
	}

	prog.Complete()
	return path, nil
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
