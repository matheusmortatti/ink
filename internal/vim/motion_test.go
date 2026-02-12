package vim

import (
	"testing"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text unchanged", "hello world", "hello world"},
		{"single escape removed", "\x1b[31mhello\x1b[0m", "hello"},
		{"multiple escapes removed", "\x1b[1m\x1b[34mbold blue\x1b[0m", "bold blue"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripANSI(tt.input)
			if got != tt.want {
				t.Errorf("StripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestVisibleLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"plain text", "hello", 5},
		{"ansi styled text", "\x1b[31mhello\x1b[0m", 5},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VisibleLength(tt.input)
			if got != tt.want {
				t.Errorf("VisibleLength(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestNextWordStart(t *testing.T) {
	tests := []struct {
		name string
		text string
		pos  int
		want int
	}{
		{"middle of word", "hello world", 2, 6},
		{"between words", "hello world", 5, 6},
		{"at word start", "hello world", 0, 6},
		{"punctuation boundary", "hello, world", 0, 5},
		{"end of text", "hello", 4, 4},
		{"cross line boundary", "hello\nworld", 3, 6},
		{"empty text", "", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)
			got := NextWordStart(runes, tt.pos)
			if got != tt.want {
				t.Errorf("NextWordStart(%q, %d) = %d, want %d", tt.text, tt.pos, got, tt.want)
			}
		})
	}
}

func TestPrevWordStart(t *testing.T) {
	tests := []struct {
		name string
		text string
		pos  int
		want int
	}{
		{"middle of word", "hello world", 8, 6},
		{"start of second word", "hello world", 6, 0},
		{"beginning of text", "hello world", 0, 0},
		{"from end", "hello world", 11, 6},
		{"cross line boundary", "hello\nworld", 6, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)
			got := PrevWordStart(runes, tt.pos)
			if got != tt.want {
				t.Errorf("PrevWordStart(%q, %d) = %d, want %d", tt.text, tt.pos, got, tt.want)
			}
		})
	}
}
