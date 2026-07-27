package progress

import (
	"io"
	"sync"
	"time"
)

// Callback is called with upload progress updates.
type Callback func(bytesRead, totalBytes int64, percentage float64)

// Reader wraps an io.Reader to track read progress.
type Reader struct {
	reader      io.Reader
	totalBytes  int64
	bytesRead   int64
	callback    Callback
	mu          sync.Mutex
	lastUpdate  time.Time
	minInterval time.Duration
}

// NewReader creates a new progress-tracking reader.
func NewReader(reader io.Reader, totalBytes int64, callback Callback) *Reader {
	return &Reader{
		reader:      reader,
		totalBytes:  totalBytes,
		callback:    callback,
		minInterval: 100 * time.Millisecond,
	}
}

// Read implements io.Reader with progress tracking.
func (pr *Reader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.mu.Lock()
		pr.bytesRead += int64(n)

		now := time.Now()
		if now.Sub(pr.lastUpdate) >= pr.minInterval {
			pr.lastUpdate = now
			bytesRead := pr.bytesRead
			pr.mu.Unlock()

			percentage := float64(0)
			if pr.totalBytes > 0 {
				percentage = float64(bytesRead) / float64(pr.totalBytes) * 100
			}
			if pr.callback != nil {
				pr.callback(bytesRead, pr.totalBytes, percentage)
			}
		} else {
			pr.mu.Unlock()
		}
	}
	return n, err
}

// BytesRead returns the current number of bytes read.
func (pr *Reader) BytesRead() int64 {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	return pr.bytesRead
}
