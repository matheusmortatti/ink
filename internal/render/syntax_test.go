package render

import (
	"testing"

	"github.com/matheusmortatti/ink/internal/block"
)

// markerDim wraps text in markers for deterministic test assertions.
// This replaces lipgloss styling which may not produce ANSI in test environments.
func markerDim(s string) string {
	return "\u00ab" + s + "\u00bb"
}

func TestDimSyntax_Heading_DimsPoundSigns(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"H1", "# Hello", "\u00ab# \u00bbHello"},
		{"H2", "## Hello World", "\u00ab## \u00bbHello World"},
		{"H3", "### Third", "\u00ab### \u00bbThird"},
		{"H4", "#### Fourth", "\u00ab#### \u00bbFourth"},
		{"H5", "##### Fifth", "\u00ab##### \u00bbFifth"},
		{"H6", "###### Sixth", "\u00ab###### \u00bbSixth"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.Heading, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_Bold_DimsDelimiters(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"asterisks", "**bold text**", "\u00ab**\u00bbbold text\u00ab**\u00bb"},
		{"underscores", "__bold text__", "\u00ab__\u00bbbold text\u00ab__\u00bb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.Paragraph, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_Italic_DimsDelimiters(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"asterisk", "*italic*", "\u00ab*\u00bbitalic\u00ab*\u00bb"},
		{"underscore", "_italic_", "\u00ab_\u00bbitalic\u00ab_\u00bb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.Paragraph, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_CodeSpan_DimsBackticks(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "`code`", "\u00ab`\u00bbcode\u00ab`\u00bb"},
		{"with_text", "run `go test` now", "run \u00ab`\u00bbgo test\u00ab`\u00bb now"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.Paragraph, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_Link_DimsBracketsAndURL(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"basic", "[click here](https://example.com)", "\u00ab[\u00bbclick here\u00ab](\u00bb\u00abhttps://example.com\u00bb\u00ab)\u00bb"},
		{"with_text", "see [docs](http://doc.go) for info", "see \u00ab[\u00bbdocs\u00ab](\u00bb\u00abhttp://doc.go\u00bb\u00ab)\u00bb for info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.Paragraph, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_CodeFence_DimsFenceLines(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			"basic",
			"```\nfmt.Println(\"hello\")\n```",
			"\u00ab```\u00bb\nfmt.Println(\"hello\")\n\u00ab```\u00bb",
		},
		{
			"with_lang",
			"```go\npackage main\n```",
			"\u00ab```go\u00bb\npackage main\n\u00ab```\u00bb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.CodeFence, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_ListMarker_DimsMarker(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"dash", "- item one", "\u00ab- \u00bbitem one"},
		{"ordered", "1. first item", "\u00ab1. \u00bbfirst item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.List, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestDimSyntax_Blockquote_DimsMarker(t *testing.T) {
	result := DimSyntax("> quoted text", block.BlockQuote, markerDim)
	expect := "\u00ab> \u00bbquoted text"
	if result != expect {
		t.Errorf("DimSyntax blockquote = %q, want %q", result, expect)
	}
}

func TestDimSyntax_HorizontalRule_DimsAll(t *testing.T) {
	result := DimSyntax("---", block.HorizontalRule, markerDim)
	expect := "\u00ab---\u00bb"
	if result != expect {
		t.Errorf("DimSyntax horizontal rule = %q, want %q", result, expect)
	}
}

func TestDimSyntax_PlainText_Unchanged(t *testing.T) {
	input := "Just plain text with no markdown"
	result := DimSyntax(input, block.Paragraph, markerDim)
	if result != input {
		t.Errorf("Plain text should be unchanged, got %q", result)
	}
}

func TestDimSyntax_EmptyString_Unchanged(t *testing.T) {
	result := DimSyntax("", block.Paragraph, markerDim)
	if result != "" {
		t.Errorf("Empty string should return empty, got %q", result)
	}
}

func TestDimSyntax_MultiLine(t *testing.T) {
	input := "# Title\nSome **bold** text\n- a list item"
	result := DimSyntax(input, block.Paragraph, markerDim)
	expect := "\u00ab# \u00bbTitle\nSome \u00ab**\u00bbbold\u00ab**\u00bb text\n\u00ab- \u00bba list item"
	if result != expect {
		t.Errorf("MultiLine dimming = %q, want %q", result, expect)
	}
}

func TestDimSyntax_Image_DimsMarkup(t *testing.T) {
	input := "![alt text](image.png)"
	result := DimSyntax(input, block.Paragraph, markerDim)
	expect := "\u00ab![\u00bbalt text\u00ab](\u00bb\u00abimage.png\u00bb\u00ab)\u00bb"
	if result != expect {
		t.Errorf("Image dimming = %q, want %q", result, expect)
	}
}

func TestDimSyntax_EscapedDelimiters_NotDimmed(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"escaped_star", "this is \\*not italic\\*", "this is \\*not italic\\*"},
		{"escaped_underscore", "no \\_emphasis\\_ here", "no \\_emphasis\\_ here"},
		{"escaped_backtick", "not \\`code\\`", "not \\`code\\`"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DimSyntax(tt.input, block.Paragraph, markerDim)
			if result != tt.expect {
				t.Errorf("DimSyntax(%q) = %q, want %q", tt.input, result, tt.expect)
			}
		})
	}
}

func TestFindClosingDelimiter(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		start  int
		char   rune
		count  int
		expect int
	}{
		{"single_star", "*hello*", 1, '*', 1, 6},
		{"double_star", "**hello**", 2, '*', 2, 7},
		{"no_close", "*hello", 1, '*', 1, -1},
		{"backtick", "`code`", 1, '`', 1, 5},
		{"escaped_close", "*hello\\**", 1, '*', 1, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.input)
			got := findClosingDelimiter(runes, tt.start, tt.char, tt.count)
			if got != tt.expect {
				t.Errorf("findClosingDelimiter = %d, want %d", got, tt.expect)
			}
		})
	}
}

func TestFindLink(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		pos       int
		expectEnd int
	}{
		{"basic", "[text](url)", 0, 10},
		{"no_paren", "[text]no paren", 0, -1},
		{"no_close", "[text(url)", 0, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.input)
			end, _, _ := findLink(runes, tt.pos)
			if end != tt.expectEnd {
				t.Errorf("findLink end = %d, want %d", end, tt.expectEnd)
			}
		})
	}
}
