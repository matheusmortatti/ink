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

func TestEmergencySaveTo_WritesContent(t *testing.T) {
	dir := t.TempDir()
	content := []byte("# Emergency save\n\ntest content")
	path, err := emergencySaveTo(dir, content)
	if err != nil {
		t.Fatalf("emergencySaveTo unexpected error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", path, err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestEmergencySaveTo_CreatesDirectory(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "new", "nested", "dir")
	_, err := emergencySaveTo(dir, []byte("data"))
	if err != nil {
		t.Fatalf("emergencySaveTo unexpected error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected a directory, got a file")
	}
}

func TestEmergencySaveTo_ReturnsRecoveryPath(t *testing.T) {
	dir := t.TempDir()
	path, err := emergencySaveTo(dir, []byte("data"))
	if err != nil {
		t.Fatalf("emergencySaveTo unexpected error: %v", err)
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "recovery-") || !strings.HasSuffix(base, ".md") {
		t.Errorf("returned path %q does not match recovery-*.md pattern", base)
	}
}

func TestEmergencySaveTo_FailsOnUnwritableDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root; permission checks don't apply")
	}
	dir := t.TempDir()
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	_, err := emergencySaveTo(dir, []byte("data"))
	if err == nil {
		t.Fatal("expected error writing to read-only directory, got nil")
	}
}

func TestEmergencySave_ProducesRecoveryFile(t *testing.T) {
	// Point XDG_STATE_HOME to a temp dir so we don't pollute the real state dir
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)

	content := []byte("# Recovery test\n\nSome content here.")
	path, err := EmergencySave(content)
	if err != nil {
		t.Fatalf("EmergencySave unexpected error: %v", err)
	}

	// Verify "ink" is a directory component in the path
	if !strings.Contains(path, filepath.Join("ink", "recovery-")) {
		t.Errorf("expected path to contain ink/recovery-..., got %q", path)
	}

	// Verify the file was actually written with correct content
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", path, err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestUserStateDir_XDGOverride(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/custom/state")
	got := userStateDir()
	if got != "/custom/state" {
		t.Errorf("userStateDir() = %q, want /custom/state", got)
	}
}

func TestUserStateDir_DefaultsToLocalState(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	got := userStateDir()
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".local", "state")
	if got != want {
		t.Errorf("userStateDir() = %q, want %q", got, want)
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
