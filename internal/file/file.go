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
