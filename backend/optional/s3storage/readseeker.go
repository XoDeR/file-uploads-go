package s3storage

import (
	"fmt"
	"io"
)

// readSeeker wraps a byte slice to satisfy io.ReadSeeker for S3 SDK.
type readSeeker struct {
	data   []byte
	offset int64
}

func (r *readSeeker) Read(p []byte) (n int, err error) {
	if r.offset >= int64(len(r.data)) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.offset:])
	r.offset += int64(n)
	return n, nil
}

func (r *readSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		r.offset = offset
	case io.SeekCurrent:
		r.offset += offset
	case io.SeekEnd:
		r.offset = int64(len(r.data)) + offset
	}
	if r.offset < 0 || r.offset > int64(len(r.data)) {
		return 0, fmt.Errorf("invalid seek offset")
	}
	return r.offset, nil
}
