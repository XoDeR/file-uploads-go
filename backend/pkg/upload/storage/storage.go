package storage

import (
	"context"
	"io"
)

// Storage abstracts where uploaded bytes are persisted.
type Storage interface {
	Save(ctx context.Context, name string, r io.Reader, contentType string) (path string, err error)
}
