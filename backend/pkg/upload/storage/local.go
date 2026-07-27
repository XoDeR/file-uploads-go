package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Local writes files to a directory on disk.
type Local struct {
	dir string
}

// NewLocal creates a local disk storage backend.
func NewLocal(dir string) *Local {
	return &Local{dir: dir}
}

// Dir returns the upload directory.
func (l *Local) Dir() string {
	return l.dir
}

// Save streams r to a file named name under the upload directory.
func (l *Local) Save(ctx context.Context, name string, r io.Reader, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := os.MkdirAll(l.dir, 0755); err != nil {
		return "", fmt.Errorf("error creating upload directory: %w", err)
	}

	path := filepath.Join(l.dir, name)
	dst, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, r); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("error saving file: %w", err)
	}

	return name, nil
}
