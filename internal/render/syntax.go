package render

import (
	"regexp"
	"strings"

	"github.com/matheusmortatti/ink/internal/block"
)

// Regex patterns for markdown syntax detection
var (
	headingRe    = regexp.MustCompile(`^(#{1,6}\s)`)
	ulMarkerRe   = regexp.MustCompile(`^([-*+]\s)`)
	olMarkerRe   = regexp.MustCompile(`^(\d+\.\s)`)
	blockquoteRe = regexp.MustCompile(`^(>\s?)`)
	codeFenceRe  = regexp.MustCompile("^(`{3,})")
	hrRe         = regexp.MustCompile(`^([-*_]{3,})\s*$`)
)

// DimSyntax returns the raw content with syntax characters wrapped in dim styling
// via the provided dimFunc. The content text remains unstyled (full brightness).
// This is a display-path-only function — it does not modify the underlying content.
func DimSyntax(rawContent string, blockType block.BlockType, dimFunc func(string) string) string {
	if rawContent == "" {
		return rawContent
	}

	// Block-level special cases
	switch blockType {
	case block.HorizontalRule:
		return dimFunc(rawContent)
	case block.CodeFence:
		return dimCodeFence(rawContent, dimFunc)
	}

	lines := strings.Split(rawContent, "\n")
	for i, line := range lines {
		lines[i] = dimLine(line, dimFunc)
	}
	return strings.Join(lines, "\n")
}

// dimCodeFence dims the opening and closing fence lines while keeping content bright.
func dimCodeFence(content string, dimFunc func(string) string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	for i, line := range lines {
		if codeFenceRe.MatchString(line) {
			lines[i] = dimFunc(line)
		}
		// Content lines stay bright (unchanged)
	}
	return strings.Join(lines, "\n")
}

// dimLine applies syntax dimming to a single line, handling both line-level
// and inline markdown syntax.
func dimLine(line string, dimFunc func(string) string) string {
	if line == "" {
		return line
	}

	// Check for horizontal rule line
	if hrRe.MatchString(line) {
		return dimFunc(line)
	}

	var prefix string
	rest := line

	// Check line-level patterns and extract the prefix to dim
	if loc := headingRe.FindStringIndex(line); loc != nil {
		prefix = line[loc[0]:loc[1]]
		rest = line[loc[1]:]
	} else if loc := blockquoteRe.FindStringIndex(line); loc != nil {
		prefix = line[loc[0]:loc[1]]
		rest = line[loc[1]:]
	} else if loc := olMarkerRe.FindStringIndex(line); loc != nil {
		prefix = line[loc[0]:loc[1]]
		rest = line[loc[1]:]
	} else if loc := ulMarkerRe.FindStringIndex(line); loc != nil {
		prefix = line[loc[0]:loc[1]]
		rest = line[loc[1]:]
	} else if codeFenceRe.MatchString(line) {
		return dimFunc(line)
	}

	// Apply inline dimming to the rest of the line
	rest = dimInline(rest, dimFunc)

	if prefix != "" {
		return dimFunc(prefix) + rest
	}
	return rest
}

// dimInline processes inline markdown syntax: bold, italic, code spans, and links.
func dimInline(text string, dimFunc func(string) string) string {
	if text == "" {
		return text
	}

	runes := []rune(text)
	var result strings.Builder
	result.Grow(len(text) * 2) // pre-allocate for styling overhead
	i := 0

	for i < len(runes) {
		// Escaped character: pass through both backslash and next char undimmed
		if runes[i] == '\\' && i+1 < len(runes) {
			result.WriteRune(runes[i])
			result.WriteRune(runes[i+1])
			i += 2
			continue
		}

		// Image: ![alt](url)
		if i+1 < len(runes) && runes[i] == '!' && runes[i+1] == '[' {
			end, altEnd, urlStart := findLink(runes, i+1)
			if end >= 0 {
				result.WriteString(dimFunc("!["))
				result.WriteString(string(runes[i+2 : altEnd]))
				result.WriteString(dimFunc("]("))
				result.WriteString(dimFunc(string(runes[urlStart:end])))
				result.WriteString(dimFunc(")"))
				i = end + 1
				continue
			}
		}

		// Link: [text](url)
		if runes[i] == '[' {
			end, altEnd, urlStart := findLink(runes, i)
			if end >= 0 {
				result.WriteString(dimFunc("["))
				result.WriteString(string(runes[i+1 : altEnd]))
				result.WriteString(dimFunc("]("))
				result.WriteString(dimFunc(string(runes[urlStart:end])))
				result.WriteString(dimFunc(")"))
				i = end + 1
				continue
			}
		}

		// Code span: `code`
		if runes[i] == '`' {
			end := findClosingDelimiter(runes, i+1, '`', 1)
			if end >= 0 {
				result.WriteString(dimFunc("`"))
				result.WriteString(string(runes[i+1 : end]))
				result.WriteString(dimFunc("`"))
				i = end + 1
				continue
			}
		}

		// Bold: **text**
		if i+1 < len(runes) && runes[i] == '*' && runes[i+1] == '*' {
			end := findClosingDelimiter(runes, i+2, '*', 2)
			if end >= 0 {
				result.WriteString(dimFunc("**"))
				// Recursively process content for nested italic
				inner := dimInline(string(runes[i+2:end]), dimFunc)
				result.WriteString(inner)
				result.WriteString(dimFunc("**"))
				i = end + 2
				continue
			}
		}

		// Bold: __text__
		if i+1 < len(runes) && runes[i] == '_' && runes[i+1] == '_' {
			end := findClosingDelimiter(runes, i+2, '_', 2)
			if end >= 0 {
				result.WriteString(dimFunc("__"))
				inner := dimInline(string(runes[i+2:end]), dimFunc)
				result.WriteString(inner)
				result.WriteString(dimFunc("__"))
				i = end + 2
				continue
			}
		}

		// Italic: *text* (single, not **)
		if runes[i] == '*' && (i+1 >= len(runes) || runes[i+1] != '*') {
			end := findClosingDelimiter(runes, i+1, '*', 1)
			if end >= 0 {
				result.WriteString(dimFunc("*"))
				result.WriteString(string(runes[i+1 : end]))
				result.WriteString(dimFunc("*"))
				i = end + 1
				continue
			}
		}

		// Italic: _text_ (single, not __)
		if runes[i] == '_' && (i+1 >= len(runes) || runes[i+1] != '_') {
			end := findClosingDelimiter(runes, i+1, '_', 1)
			if end >= 0 {
				result.WriteString(dimFunc("_"))
				result.WriteString(string(runes[i+1 : end]))
				result.WriteString(dimFunc("_"))
				i = end + 1
				continue
			}
		}

		// Regular character — pass through
		result.WriteRune(runes[i])
		i++
	}

	return result.String()
}

// findClosingDelimiter finds the position of a closing delimiter starting from pos.
// For count=2, it looks for two consecutive chars (e.g., **). Returns -1 if not found.
// Skips escaped delimiters (preceded by backslash).
func findClosingDelimiter(runes []rune, start int, char rune, count int) int {
	for i := start; i <= len(runes)-count; i++ {
		// Skip escaped characters
		if i > 0 && runes[i-1] == '\\' {
			continue
		}
		match := true
		for j := 0; j < count; j++ {
			if runes[i+j] != char {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// findLink finds a markdown link starting with '[' at pos.
// Returns (endParenIdx, closeBracketIdx, urlStartIdx) or (-1, -1, -1) if not found.
// Pattern: [text](url)
func findLink(runes []rune, pos int) (int, int, int) {
	if pos >= len(runes) || runes[pos] != '[' {
		return -1, -1, -1
	}

	// Find closing ]
	bracketEnd := -1
	for i := pos + 1; i < len(runes); i++ {
		if runes[i] == ']' {
			bracketEnd = i
			break
		}
	}
	if bracketEnd < 0 {
		return -1, -1, -1
	}

	// Must be followed by (
	if bracketEnd+1 >= len(runes) || runes[bracketEnd+1] != '(' {
		return -1, -1, -1
	}

	// Find closing )
	parenEnd := -1
	for i := bracketEnd + 2; i < len(runes); i++ {
		if runes[i] == ')' {
			parenEnd = i
			break
		}
	}
	if parenEnd < 0 {
		return -1, -1, -1
	}

	return parenEnd, bracketEnd, bracketEnd + 2
}
