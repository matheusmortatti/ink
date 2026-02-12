package vim

import (
	"regexp"
	"unicode"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// VisibleLength returns the visible character count (excluding ANSI codes).
func VisibleLength(s string) int {
	return len([]rune(StripANSI(s)))
}

// charClass returns the character class for word motion boundary detection.
// 0 = whitespace, 1 = word (alphanumeric/underscore), 2 = punctuation
func charClass(r rune) int {
	if unicode.IsSpace(r) {
		return 0
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return 1
	}
	return 2
}

// NextWordStart finds the next word boundary position from a given position in text.
// Returns the new position (may equal pos if already at end).
func NextWordStart(text []rune, pos int) int {
	if len(text) == 0 {
		return 0
	}
	if pos >= len(text)-1 {
		return len(text) - 1
	}

	i := pos
	startClass := charClass(text[i])

	// Skip past current word (same character class)
	for i < len(text) && charClass(text[i]) == startClass {
		i++
	}

	// Skip whitespace
	for i < len(text) && unicode.IsSpace(text[i]) {
		i++
	}

	if i >= len(text) {
		return len(text) - 1
	}
	return i
}

// PrevWordStart finds the previous word boundary position from a given position in text.
// Returns the new position (may equal pos if already at beginning).
func PrevWordStart(text []rune, pos int) int {
	if pos <= 0 || len(text) == 0 {
		return 0
	}

	i := pos - 1

	// Skip whitespace backwards
	for i > 0 && unicode.IsSpace(text[i]) {
		i--
	}

	// Find start of current word (same character class)
	cls := charClass(text[i])
	for i > 0 && charClass(text[i-1]) == cls {
		i--
	}

	return i
}
