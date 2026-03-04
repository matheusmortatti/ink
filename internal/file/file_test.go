package file

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Valid extensions
		{"lowercase .md", "readme.md", false},
		{"uppercase .MD", "README.MD", false},
		{"mixed .Md", "file.Md", false},
		{"mixed .mD", "file.mD", false},

		// Invalid extensions
		{"txt extension", "file.txt", true},
		{"html extension", "file.html", true},
		{"no extension", "file", true},
		{"empty string", "", true},
		{"markdown extension", "file.markdown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath_ErrorWrapping(t *testing.T) {
	err := ValidatePath("file.txt")
	if err == nil {
		t.Fatal("expected error for non-.md file")
	}
	if !errors.Is(err, ErrNotMarkdown) {
		t.Errorf("expected error to wrap ErrNotMarkdown, got %v", err)
	}
}

func TestReadFile_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	want := []byte("# Hello\n\nWorld")
	if err := os.WriteFile(path, want, 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) unexpected error: %v", path, err)
	}
	if string(got) != string(want) {
		t.Errorf("ReadFile(%q) = %q, want %q", path, got, want)
	}
}

func TestReadFile_MissingFile(t *testing.T) {
	_, err := ReadFile("/nonexistent/path/file.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !errors.Is(err, ErrFileNotFound) {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestWriteFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.md")
	want := []byte("# Hello\n\nWorld")
	if err := WriteFile(path, want); err != nil {
		t.Fatalf("WriteFile unexpected error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile after WriteFile: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("WriteFile content = %q, want %q", got, want)
	}
}

func TestWriteFile_AtomicWrite(t *testing.T) {
	// Verify the target file appears after rename, not during write.
	dir := t.TempDir()
	path := filepath.Join(dir, "atomic.md")
	content := []byte("atomic content")
	if err := WriteFile(path, content); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// No leftover temp files should exist.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".ink-save-") {
			t.Errorf("leftover temp file after successful write: %s", e.Name())
		}
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", got, content)
	}
}

func TestWriteFile_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; permission checks don't apply")
	}
	dir := t.TempDir()
	// Make directory read-only so CreateTemp fails.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	path := filepath.Join(dir, "nope.md")
	err := WriteFile(path, []byte("data"))
	if err == nil {
		t.Fatal("expected error for read-only directory, got nil")
	}
}

func TestWriteFile_InvalidDirectory(t *testing.T) {
	path := filepath.Join("/nonexistent/dir", "file.md")
	err := WriteFile(path, []byte("data"))
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}

func TestWriteFile_Permissions(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; permission checks don't apply")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.md")
	if err := WriteFile(path, []byte("data")); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("file permissions = %o, want 0644", perm)
	}
}

func TestReadFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) unexpected error: %v", path, err)
	}
	if len(got) != 0 {
		t.Errorf("ReadFile(%q) = %q, want empty", path, got)
	}
}
