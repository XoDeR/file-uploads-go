package chunked

import "testing"

func TestTotalChunks(t *testing.T) {
	tests := []struct {
		name      string
		totalSize int64
		chunkSize int64
		want      int
		wantErr   bool
	}{
		{name: "exact multiple", totalSize: 15, chunkSize: 5, want: 3},
		{name: "remainder last chunk", totalSize: 12, chunkSize: 5, want: 3},
		{name: "single chunk exact", totalSize: 5, chunkSize: 5, want: 1},
		{name: "single chunk smaller", totalSize: 3, chunkSize: 5, want: 1},
		{name: "large file", totalSize: 100*1024*1024 + 1, chunkSize: 5 * 1024 * 1024, want: 21},
		{name: "zero total", totalSize: 0, chunkSize: 5, wantErr: true},
		{name: "negative total", totalSize: -1, chunkSize: 5, wantErr: true},
		{name: "zero chunk", totalSize: 10, chunkSize: 0, wantErr: true},
		{name: "negative chunk", totalSize: 10, chunkSize: -5, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TotalChunks(tt.totalSize, tt.chunkSize)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got chunks=%d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("TotalChunks(%d, %d) = %d, want %d", tt.totalSize, tt.chunkSize, got, tt.want)
			}
		})
	}
}

func TestExpectedChunkSize(t *testing.T) {
	tests := []struct {
		name        string
		totalSize   int64
		chunkSize   int64
		chunkNumber int
		totalChunks int
		want        int64
	}{
		{name: "middle chunk", totalSize: 12, chunkSize: 5, chunkNumber: 0, totalChunks: 3, want: 5},
		{name: "second chunk", totalSize: 12, chunkSize: 5, chunkNumber: 1, totalChunks: 3, want: 5},
		{name: "last remainder", totalSize: 12, chunkSize: 5, chunkNumber: 2, totalChunks: 3, want: 2},
		{name: "exact last", totalSize: 15, chunkSize: 5, chunkNumber: 2, totalChunks: 3, want: 5},
		{name: "single chunk", totalSize: 3, chunkSize: 5, chunkNumber: 0, totalChunks: 1, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpectedChunkSize(tt.totalSize, tt.chunkSize, tt.chunkNumber, tt.totalChunks)
			if got != tt.want {
				t.Fatalf("ExpectedChunkSize(...) = %d, want %d", got, tt.want)
			}
		})
	}
}
