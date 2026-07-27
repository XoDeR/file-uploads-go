package progress

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Tracker manages progress updates for multiple uploads.
type Tracker struct {
	uploads map[string]*UploadProgress
	mu      sync.RWMutex
}

// UploadProgress represents the progress of a single upload.
type UploadProgress struct {
	UploadID    string    `json:"upload_id"`
	Filename    string    `json:"filename"`
	BytesRead   int64     `json:"bytes_read"`
	TotalBytes  int64     `json:"total_bytes"`
	Percentage  float64   `json:"percentage"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	subscribers []chan *UploadProgress
	mu          sync.Mutex
}

// NewTracker creates a new progress tracker.
func NewTracker() *Tracker {
	return &Tracker{
		uploads: make(map[string]*UploadProgress),
	}
}

// StartTracking begins tracking a new upload.
func (pt *Tracker) StartTracking(uploadID, filename string, totalBytes int64) *UploadProgress {
	progress := &UploadProgress{
		UploadID:   uploadID,
		Filename:   filename,
		TotalBytes: totalBytes,
		Status:     "in_progress",
		StartedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	pt.mu.Lock()
	pt.uploads[uploadID] = progress
	pt.mu.Unlock()

	return progress
}

// Get returns progress for an upload ID.
func (pt *Tracker) Get(uploadID string) (*UploadProgress, bool) {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	p, ok := pt.uploads[uploadID]
	return p, ok
}

// Complete marks an upload as completed.
func (up *UploadProgress) Complete() {
	up.mu.Lock()
	up.Status = "completed"
	up.Percentage = 100
	up.BytesRead = up.TotalBytes
	up.UpdatedAt = time.Now()
	subs := make([]chan *UploadProgress, len(up.subscribers))
	copy(subs, up.subscribers)
	up.mu.Unlock()

	for _, sub := range subs {
		select {
		case sub <- up:
		default:
		}
	}
}

// Fail marks an upload as failed.
func (up *UploadProgress) Fail() {
	up.mu.Lock()
	up.Status = "error"
	up.UpdatedAt = time.Now()
	subs := make([]chan *UploadProgress, len(up.subscribers))
	copy(subs, up.subscribers)
	up.mu.Unlock()

	for _, sub := range subs {
		select {
		case sub <- up:
		default:
		}
	}
}

// Update updates the progress of an upload and notifies subscribers.
func (up *UploadProgress) Update(bytesRead int64) {
	up.mu.Lock()
	up.BytesRead = bytesRead
	if up.TotalBytes > 0 {
		up.Percentage = float64(bytesRead) / float64(up.TotalBytes) * 100
	}
	up.UpdatedAt = time.Now()

	subs := make([]chan *UploadProgress, len(up.subscribers))
	copy(subs, up.subscribers)
	up.mu.Unlock()

	for _, sub := range subs {
		select {
		case sub <- up:
		default:
		}
	}
}

// Subscribe adds a subscriber for progress updates.
func (up *UploadProgress) Subscribe() chan *UploadProgress {
	ch := make(chan *UploadProgress, 10)
	up.mu.Lock()
	up.subscribers = append(up.subscribers, ch)
	up.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber.
func (up *UploadProgress) Unsubscribe(ch chan *UploadProgress) {
	up.mu.Lock()
	defer up.mu.Unlock()

	for i, sub := range up.subscribers {
		if sub == ch {
			up.subscribers = append(up.subscribers[:i], up.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

// SSEHandler handles Server-Sent Events for upload progress.
func (pt *Tracker) SSEHandler(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")

	pt.mu.RLock()
	progress, exists := pt.uploads[uploadID]
	pt.mu.RUnlock()

	if !exists {
		http.Error(w, "Upload not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	updates := progress.Subscribe()
	defer progress.Unsubscribe(updates)

	// Send current state immediately
	progress.mu.Lock()
	snapshot := *progress
	progress.mu.Unlock()
	data, _ := json.Marshal(&snapshot)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}

			data, _ := json.Marshal(update)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()

			if update.Status == "completed" || update.Status == "error" {
				return
			}
		}
	}
}
