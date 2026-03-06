package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// emergencySaveTo writes content to a recovery file in dir.
// It creates dir if it does not exist (AC #4). Uses direct os.WriteFile
// rather than atomic write — emergency context, simpler is safer (AC #1–3).
// Returns the full path on success, or ("", err) on failure.
func emergencySaveTo(dir string, content []byte) (string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	now := time.Now()
	filename := "recovery-" + now.Format("20060102-150405") + fmt.Sprintf("-%d", now.UnixNano()%1e6) + ".md"
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", err
	}
	return path, nil
}

// userStateDir returns the XDG state home directory (Linux: $XDG_STATE_HOME or ~/.local/state).
// Falls back to os.TempDir() if the home directory cannot be determined.
func userStateDir() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return xdg
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
	}
	return filepath.Join(home, ".local", "state")
}

// EmergencySave writes content to ~/.local/state/ink/recovery-{timestamp}.md.
// Falls back to os.TempDir()/ink if the state directory cannot be determined.
// Returns the recovery file path on success, or ("", err) on failure.
func EmergencySave(content []byte) (string, error) {
	return emergencySaveTo(filepath.Join(userStateDir(), "ink"), content)
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
