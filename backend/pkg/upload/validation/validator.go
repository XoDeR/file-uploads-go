package validation

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	ErrFileTooLarge     = errors.New("file exceeds maximum size limit")
	ErrInvalidFileType  = errors.New("file type not allowed")
	ErrMaliciousContent = errors.New("file contains potentially malicious content")
)

// FileValidator provides comprehensive file validation.
type FileValidator struct {
	MaxFileSize       int64
	AllowedTypes      []string
	AllowedExtensions []string
	BlockedExtensions []string
}

// DefaultValidator returns a validator with sensible defaults.
func DefaultValidator() *FileValidator {
	return &FileValidator{
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"application/pdf",
			"text/plain",
			"application/zip",
			"application/octet-stream",
		},
		// Empty AllowedExtensions = allow any extension not in BlockedExtensions (POC-friendly).
		AllowedExtensions: nil,
		BlockedExtensions: []string{
			".exe", ".bat", ".cmd", ".sh", ".php",
			".js", ".vbs", ".ps1", ".msi",
		},
	}
}

// ValidateFile performs comprehensive validation on an uploaded file.
func (v *FileValidator) ValidateFile(header *multipart.FileHeader, file io.ReadSeeker) error {
	if header.Size > v.MaxFileSize {
		return fmt.Errorf("%w: %d bytes (max: %d)",
			ErrFileTooLarge, header.Size, v.MaxFileSize)
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if err := v.ValidateExtension(ext); err != nil {
		return err
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error reading file: %w", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error resetting file position: %w", err)
	}

	contentType := http.DetectContentType(buf[:n])
	if err := v.ValidateContentType(contentType, ext); err != nil {
		return err
	}

	return v.ScanForMaliciousContent(buf[:n])
}

// ValidateExtension checks blocked and allowed extensions.
func (v *FileValidator) ValidateExtension(ext string) error {
	ext = strings.ToLower(ext)
	for _, blocked := range v.BlockedExtensions {
		if ext == blocked {
			return fmt.Errorf("%w: %s extension is blocked", ErrInvalidFileType, ext)
		}
	}

	if len(v.AllowedExtensions) == 0 {
		return nil
	}

	for _, allowed := range v.AllowedExtensions {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("%w: %s extension not allowed", ErrInvalidFileType, ext)
}

// ValidateHeaderBytes validates extension and content from the first bytes of a stream.
func (v *FileValidator) ValidateHeaderBytes(filename string, header []byte) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if err := v.ValidateExtension(ext); err != nil {
		return err
	}
	if len(header) == 0 {
		return nil
	}
	contentType := http.DetectContentType(header)
	if err := v.ValidateContentType(contentType, ext); err != nil {
		return err
	}
	return v.ScanForMaliciousContent(header)
}

// ValidateContentType ensures the detected content type matches the extension.
func (v *FileValidator) ValidateContentType(contentType, ext string) error {
	expectedTypes := map[string][]string{
		".jpg":  {"image/jpeg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".webp": {"image/webp"},
		".pdf":  {"application/pdf"},
		".txt":  {"text/plain", "text/plain; charset=utf-8"},
		".zip":  {"application/zip", "application/x-zip-compressed"},
	}

	expected, exists := expectedTypes[ext]
	if !exists {
		return nil
	}

	for _, exp := range expected {
		if strings.HasPrefix(contentType, strings.Split(exp, ";")[0]) {
			return nil
		}
	}

	return fmt.Errorf("%w: content type %s doesn't match extension %s",
		ErrInvalidFileType, contentType, ext)
}

// ScanForMaliciousContent looks for dangerous patterns in the file header.
func (v *FileValidator) ScanForMaliciousContent(header []byte) error {
	dangerousPatterns := [][]byte{
		[]byte("<?php"),
		[]byte("<script"),
		[]byte("javascript:"),
		[]byte("vbscript:"),
	}

	lower := bytes.ToLower(header)
	for _, pattern := range dangerousPatterns {
		if bytes.Contains(lower, bytes.ToLower(pattern)) {
			return fmt.Errorf("%w: dangerous pattern detected", ErrMaliciousContent)
		}
	}

	return nil
}

// SanitizeFilename removes or replaces dangerous characters from filenames.
func SanitizeFilename(filename string) string {
	// Normalize separators without filepath.Base (which treats "a:b" as a Windows drive).
	filename = strings.ReplaceAll(filename, "\\", "/")
	if i := strings.LastIndex(filename, "/"); i >= 0 {
		filename = filename[i+1:]
	}

	replacer := strings.NewReplacer(
		"..", "",
		"/", "_",
		"\x00", "",
		"<", "",
		">", "",
		":", "",
		"\"", "",
		"|", "",
		"?", "",
		"*", "",
	)

	filename = replacer.Replace(filename)

	if filename == "" || filename == "." {
		filename = "unnamed_file"
	}

	return filename
}
