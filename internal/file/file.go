package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrNotMarkdown  = errors.New("not a markdown file")
	ErrFileNotFound = errors.New("file not found")
)

// ValidatePath checks that the path has a .md extension.
// Case-insensitive: .md, .MD, .Md all accepted.
func ValidatePath(path string) error {
	ext := filepath.Ext(path)
	if !strings.EqualFold(ext, ".md") {
		return fmt.Errorf("%w: %s", ErrNotMarkdown, path)
	}
	return nil
}

// WriteFile atomically writes content to path using a temp file + rename pattern.
// This ensures the target file is never partially written.
func WriteFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ink-save-*")
	if err != nil {
		return fmt.Errorf("cannot save: %w", err)
	}
	tmpPath := tmp.Name()

	_, err = tmp.Write(content)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot save: %w", err)
	}

	// Set standard file permissions before rename (CreateTemp uses 0600).
	if err := os.Chmod(tmpPath, 0644); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot save: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cannot save: %w", err)
	}
	return nil
}

// ReadFile reads the contents of a file at the given path.
// Returns ErrFileNotFound if the file does not exist.
func ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrFileNotFound, path)
		}
		return nil, err
	}
	return content, nil
}
