package chunked

import "fmt"

// TotalChunks returns the number of chunks needed for a file of totalSize
// using chunks of chunkSize. Returns an error for non-positive sizes.
func TotalChunks(totalSize, chunkSize int64) (int, error) {
	if totalSize <= 0 || chunkSize <= 0 {
		return 0, fmt.Errorf("invalid size parameters: total_size=%d chunk_size=%d", totalSize, chunkSize)
	}
	return int((totalSize + chunkSize - 1) / chunkSize), nil
}

// ExpectedChunkSize returns the expected byte length of chunkNumber
// (0-based). The last chunk may be smaller than chunkSize.
func ExpectedChunkSize(totalSize, chunkSize int64, chunkNumber, totalChunks int) int64 {
	if chunkNumber == totalChunks-1 {
		return totalSize - (int64(chunkNumber) * chunkSize)
	}
	return chunkSize
}
