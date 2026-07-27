package validation

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "normal", input: "photo.jpg", want: "photo.jpg"},
		{name: "path traversal unix", input: "../../etc/passwd", want: "passwd"},
		{name: "path traversal nested", input: "foo/../../bar.txt", want: "bar.txt"},
		{name: "dangerous chars", input: `a:b*c?d|e<f>g"h.txt`, want: "abcdefgh.txt"},
		{name: "null byte", input: "file\x00name.txt", want: "filename.txt"},
		{name: "dots only after sanitize", input: "..", want: "unnamed_file"},
		{name: "empty", input: "", want: "unnamed_file"},
		{name: "dot", input: ".", want: "unnamed_file"},
		{name: "backslash path", input: `..\..\secret.txt`, want: "secret.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFilename(tt.input)
			if got != tt.want {
				t.Fatalf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
