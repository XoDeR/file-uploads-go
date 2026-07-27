package chunked

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/file-uploads-go/backend/pkg/upload/validation"
)

// Manager handles the lifecycle of chunked uploads.
type Manager struct {
	uploads   map[string]*ChunkedUpload
	uploadDir string
	mu        sync.RWMutex
}

// NewManager creates a new manager for chunked uploads.
func NewManager(uploadDir string) *Manager {
	return &Manager{
		uploads:   make(map[string]*ChunkedUpload),
		uploadDir: uploadDir,
	}
}

// InitiateUpload creates a new chunked upload session.
func (m *Manager) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename  string `json:"filename"`
		TotalSize int64  `json:"total_size"`
		ChunkSize int64  `json:"chunk_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	totalChunks, err := TotalChunks(req.TotalSize, req.ChunkSize)
	if err != nil {
		http.Error(w, "Invalid size parameters", http.StatusBadRequest)
		return
	}

	uploadID := generateUploadID()
	filename := validation.SanitizeFilename(req.Filename)

	upload := &ChunkedUpload{
		ID:             uploadID,
		Filename:       filename,
		TotalSize:      req.TotalSize,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: make(map[int]bool),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	uploadPath := filepath.Join(m.uploadDir, "chunks", uploadID)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}

	m.mu.Lock()
	m.uploads[uploadID] = upload
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(upload)
}

// UploadChunk handles the upload of a single chunk.
func (m *Manager) UploadChunk(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")
	chunkNumber := 0
	if _, err := fmt.Sscanf(r.URL.Query().Get("chunk"), "%d", &chunkNumber); err != nil {
		http.Error(w, "Invalid chunk number", http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	upload, exists := m.uploads[uploadID]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "Upload session not found", http.StatusNotFound)
		return
	}

	if chunkNumber < 0 || chunkNumber >= upload.TotalChunks {
		http.Error(w, "Invalid chunk number", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	if upload.UploadedChunks[chunkNumber] {
		m.mu.Unlock()
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Chunk already uploaded")
		return
	}
	m.mu.Unlock()

	chunkPath := filepath.Join(m.uploadDir, "chunks", uploadID,
		fmt.Sprintf("chunk_%d", chunkNumber))

	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		http.Error(w, "Error creating chunk file", http.StatusInternalServerError)
		return
	}
	defer chunkFile.Close()

	written, err := io.Copy(chunkFile, r.Body)
	if err != nil {
		os.Remove(chunkPath)
		http.Error(w, "Error writing chunk", http.StatusInternalServerError)
		return
	}

	expectedSize := ExpectedChunkSize(upload.TotalSize, upload.ChunkSize, chunkNumber, upload.TotalChunks)
	if written != expectedSize {
		os.Remove(chunkPath)
		http.Error(w, "Chunk size mismatch", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	upload.UploadedChunks[chunkNumber] = true
	uploadedCount := len(upload.UploadedChunks)
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"chunk":    chunkNumber,
		"uploaded": uploadedCount,
		"total":    upload.TotalChunks,
	})
}

// CompleteUpload assembles all chunks into the final file.
func (m *Manager) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")

	m.mu.RLock()
	upload, exists := m.uploads[uploadID]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "Upload session not found", http.StatusNotFound)
		return
	}

	m.mu.RLock()
	uploadedCount := len(upload.UploadedChunks)
	m.mu.RUnlock()
	if uploadedCount != upload.TotalChunks {
		http.Error(w, fmt.Sprintf("Missing chunks: %d/%d uploaded",
			uploadedCount, upload.TotalChunks), http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(m.uploadDir, 0755); err != nil {
		http.Error(w, "Error creating upload directory", http.StatusInternalServerError)
		return
	}

	finalPath := filepath.Join(m.uploadDir, upload.Filename)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		http.Error(w, "Error creating final file", http.StatusInternalServerError)
		return
	}
	defer finalFile.Close()

	chunksDir := filepath.Join(m.uploadDir, "chunks", uploadID)
	for i := 0; i < upload.TotalChunks; i++ {
		chunkPath := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d", i))

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			http.Error(w, "Error reading chunk", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(finalFile, chunkFile); err != nil {
			chunkFile.Close()
			http.Error(w, "Error assembling file", http.StatusInternalServerError)
			return
		}
		chunkFile.Close()
	}

	os.RemoveAll(chunksDir)

	m.mu.Lock()
	delete(m.uploads, uploadID)
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"filename": upload.Filename,
		"size":     upload.TotalSize,
		"status":   "complete",
	})
}

// GetUploadStatus returns the current state of a chunked upload.
func (m *Manager) GetUploadStatus(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")

	m.mu.RLock()
	upload, exists := m.uploads[uploadID]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "Upload session not found", http.StatusNotFound)
		return
	}

	missingChunks := make([]int, 0)
	m.mu.RLock()
	for i := 0; i < upload.TotalChunks; i++ {
		if !upload.UploadedChunks[i] {
			missingChunks = append(missingChunks, i)
		}
	}
	uploadedCount := len(upload.UploadedChunks)
	m.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"upload_id":       upload.ID,
		"filename":        upload.Filename,
		"uploaded_chunks": uploadedCount,
		"total_chunks":    upload.TotalChunks,
		"missing_chunks":  missingChunks,
		"expires_at":      upload.ExpiresAt,
	})
}

func generateUploadID() string {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		panic(fmt.Errorf("failed to generate upload ID: %w", err))
	}
	return hex.EncodeToString(data)
}
