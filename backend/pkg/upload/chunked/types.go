package chunked

import "time"

// ChunkedUpload represents the state of an in-progress chunked upload.
type ChunkedUpload struct {
	ID             string      `json:"id"`
	Filename       string      `json:"filename"`
	TotalSize      int64       `json:"total_size"`
	ChunkSize      int64       `json:"chunk_size"`
	TotalChunks    int         `json:"total_chunks"`
	UploadedChunks map[int]bool `json:"uploaded_chunks"`
	CreatedAt      time.Time   `json:"created_at"`
	ExpiresAt      time.Time   `json:"expires_at"`
}
