package file

import (
	"errors"
	"os"
	"path/filepath"
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
