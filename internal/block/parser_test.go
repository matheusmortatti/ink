package block

import (
	"strings"
	"testing"
)

// --- Paragraph parsing tests ---

func TestParse_Paragraphs(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantTypes  []BlockType
		wantRaw    []string
	}{
		{
			name:      "single paragraph",
			input:     "Hello world",
			wantCount: 1,
			wantTypes: []BlockType{Paragraph},
			wantRaw:   []string{"Hello world"},
		},
		{
			name:      "two paragraphs",
			input:     "Hello\n\nWorld",
			wantCount: 2,
			wantTypes: []BlockType{Paragraph, Paragraph},
			wantRaw:   []string{"Hello", "World"},
		},
		{
			name:      "three paragraphs",
			input:     "One\n\nTwo\n\nThree",
			wantCount: 3,
			wantTypes: []BlockType{Paragraph, Paragraph, Paragraph},
			wantRaw:   []string{"One", "Two", "Three"},
		},
		{
			name:      "multi-line paragraph",
			input:     "Line one\nline two\nline three",
			wantCount: 1,
			wantTypes: []BlockType{Paragraph},
			wantRaw:   []string{"Line one\nline two\nline three"},
		},
		{
			name:      "paragraph with inline formatting",
			input:     "This is **bold** and *italic* text",
			wantCount: 1,
			wantTypes: []BlockType{Paragraph},
			wantRaw:   []string{"This is **bold** and *italic* text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, tt.wantCount)
			assertBlockTypes(t, blocks, tt.wantTypes)
			assertBlockRaw(t, blocks, tt.wantRaw)
		})
	}
}

// --- Heading parsing tests ---

func TestParse_Headings(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantType  BlockType
		wantRaw   string
		wantLevel int
	}{
		{"h1", "# Heading 1", Heading, "# Heading 1", 1},
		{"h2", "## Heading 2", Heading, "## Heading 2", 2},
		{"h3", "### Heading 3", Heading, "### Heading 3", 3},
		{"h4", "#### Heading 4", Heading, "#### Heading 4", 4},
		{"h5", "##### Heading 5", Heading, "##### Heading 5", 5},
		{"h6", "###### Heading 6", Heading, "###### Heading 6", 6},
		{"h1 setext", "Heading\n=======", Heading, "Heading\n=======", 1},
		{"h2 setext", "Heading\n-------", Heading, "Heading\n-------", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, 1)
			if blocks[0].Type != tt.wantType {
				t.Errorf("type: got %s, want %s", blocks[0].Type, tt.wantType)
			}
			if blocks[0].Raw != tt.wantRaw {
				t.Errorf("raw: got %q, want %q", blocks[0].Raw, tt.wantRaw)
			}
			if blocks[0].Level != tt.wantLevel {
				t.Errorf("level: got %d, want %d", blocks[0].Level, tt.wantLevel)
			}
		})
	}
}

func TestParse_HeadingAndParagraph(t *testing.T) {
	blocks := Parse([]byte("# Title\n\nBody text"))
	assertBlockCount(t, blocks, 2)
	assertBlockTypes(t, blocks, []BlockType{Heading, Paragraph})
	if blocks[0].Level != 1 {
		t.Errorf("heading level: got %d, want 1", blocks[0].Level)
	}
	if blocks[0].Raw != "# Title" {
		t.Errorf("heading raw: got %q, want %q", blocks[0].Raw, "# Title")
	}
}

// --- List parsing tests ---

func TestParse_Lists(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{
			name:    "unordered list with dashes",
			input:   "- item 1\n- item 2\n- item 3",
			wantRaw: "- item 1\n- item 2\n- item 3",
		},
		{
			name:    "unordered list with asterisks",
			input:   "* item 1\n* item 2",
			wantRaw: "* item 1\n* item 2",
		},
		{
			name:    "ordered list",
			input:   "1. first\n2. second\n3. third",
			wantRaw: "1. first\n2. second\n3. third",
		},
		{
			name:    "nested list",
			input:   "- parent\n  - child\n  - child2\n- parent2",
			wantRaw: "- parent\n  - child\n  - child2\n- parent2",
		},
		{
			name:    "list with multi-line items",
			input:   "- item one\n  continues here\n- item two",
			wantRaw: "- item one\n  continues here\n- item two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, 1)
			if blocks[0].Type != List {
				t.Errorf("type: got %s, want List", blocks[0].Type)
			}
			if blocks[0].Raw != tt.wantRaw {
				t.Errorf("raw: got %q, want %q", blocks[0].Raw, tt.wantRaw)
			}
		})
	}
}

// --- Code fence parsing tests ---

func TestParse_CodeFence(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{
			name:    "basic code fence",
			input:   "```\ncode here\n```",
			wantRaw: "```\ncode here\n```",
		},
		{
			name:    "code fence with language",
			input:   "```go\nfunc main() {}\n```",
			wantRaw: "```go\nfunc main() {}\n```",
		},
		{
			name:    "code fence with blank lines",
			input:   "```go\nfunc main() {\n\n}\n```",
			wantRaw: "```go\nfunc main() {\n\n}\n```",
		},
		{
			name:    "code fence with multiple blank lines",
			input:   "```\nline1\n\n\nline2\n```",
			wantRaw: "```\nline1\n\n\nline2\n```",
		},
		{
			name:    "tilde code fence",
			input:   "~~~\ncode\n~~~",
			wantRaw: "~~~\ncode\n~~~",
		},
		{
			name:    "empty code fence",
			input:   "```\n```",
			wantRaw: "```\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, 1)
			if blocks[0].Type != CodeFence {
				t.Errorf("type: got %s, want CodeFence", blocks[0].Type)
			}
			if blocks[0].Raw != tt.wantRaw {
				t.Errorf("raw:\ngot:  %q\nwant: %q", blocks[0].Raw, tt.wantRaw)
			}
		})
	}

	// Separate test for empty code fence between blocks (3 blocks total).
	t.Run("empty code fence between blocks", func(t *testing.T) {
		blocks := Parse([]byte("Before\n\n```\n```\n\nAfter"))
		assertBlockCount(t, blocks, 3)
		assertBlockTypes(t, blocks, []BlockType{Paragraph, CodeFence, Paragraph})
		if blocks[1].Raw != "```\n```" {
			t.Errorf("code fence raw:\ngot:  %q\nwant: %q", blocks[1].Raw, "```\n```")
		}
	})
}

// --- Indented code block tests ---

func TestParse_IndentedCodeBlock(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{
			name:    "basic indented code",
			input:   "    code line 1\n    code line 2",
			wantRaw: "    code line 1\n    code line 2",
		},
		{
			name:    "indented code after paragraph",
			input:   "Text\n\n    code here",
			wantRaw: "    code here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			found := false
			for _, b := range blocks {
				if b.Type == CodeBlock {
					found = true
					if b.Raw != tt.wantRaw {
						t.Errorf("raw:\ngot:  %q\nwant: %q", b.Raw, tt.wantRaw)
					}
				}
			}
			if !found {
				t.Errorf("no CodeBlock found in blocks: %v", blockTypes(blocks))
			}
		})
	}
}

// --- Block quote parsing tests ---

func TestParse_BlockQuotes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{
			name:    "single line quote",
			input:   "> quoted text",
			wantRaw: "> quoted text",
		},
		{
			name:    "multi-line quote",
			input:   "> line 1\n> line 2\n> line 3",
			wantRaw: "> line 1\n> line 2\n> line 3",
		},
		{
			name:    "nested quote",
			input:   "> outer\n>> inner",
			wantRaw: "> outer\n>> inner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, 1)
			if blocks[0].Type != BlockQuote {
				t.Errorf("type: got %s, want BlockQuote", blocks[0].Type)
			}
			if blocks[0].Raw != tt.wantRaw {
				t.Errorf("raw:\ngot:  %q\nwant: %q", blocks[0].Raw, tt.wantRaw)
			}
		})
	}
}

// --- Table parsing tests (GFM) ---

func TestParse_Tables(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{
			name:    "basic table",
			input:   "| A | B |\n| --- | --- |\n| 1 | 2 |",
			wantRaw: "| A | B |\n| --- | --- |\n| 1 | 2 |",
		},
		{
			name:    "table with alignment",
			input:   "| Left | Center | Right |\n| :--- | :---: | ---: |\n| a | b | c |",
			wantRaw: "| Left | Center | Right |\n| :--- | :---: | ---: |\n| a | b | c |",
		},
		{
			name:    "multi-row table",
			input:   "| H1 | H2 |\n| --- | --- |\n| r1c1 | r1c2 |\n| r2c1 | r2c2 |\n| r3c1 | r3c2 |",
			wantRaw: "| H1 | H2 |\n| --- | --- |\n| r1c1 | r1c2 |\n| r2c1 | r2c2 |\n| r3c1 | r3c2 |",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, 1)
			if blocks[0].Type != Table {
				t.Errorf("type: got %s, want Table", blocks[0].Type)
			}
			if blocks[0].Raw != tt.wantRaw {
				t.Errorf("raw:\ngot:  %q\nwant: %q", blocks[0].Raw, tt.wantRaw)
			}
		})
	}
}

// --- Horizontal rule parsing tests ---

func TestParse_HorizontalRules(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{"dashes", "Above\n\n---\n\nBelow", "---"},
		{"asterisks", "Above\n\n***\n\nBelow", "***"},
		{"underscores", "Above\n\n___\n\nBelow", "___"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockCount(t, blocks, 3)
			if blocks[1].Type != HorizontalRule {
				t.Errorf("type: got %s, want HorizontalRule", blocks[1].Type)
			}
			if blocks[1].Raw != tt.wantRaw {
				t.Errorf("raw: got %q, want %q", blocks[1].Raw, tt.wantRaw)
			}
		})
	}
}

// --- Mixed block types ---

func TestParse_MixedBlockTypes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTypes []BlockType
	}{
		{
			name:      "heading, paragraph, list",
			input:     "# Title\n\nSome text\n\n- item 1\n- item 2",
			wantTypes: []BlockType{Heading, Paragraph, List},
		},
		{
			name:      "paragraph, code fence, paragraph",
			input:     "Before\n\n```\ncode\n```\n\nAfter",
			wantTypes: []BlockType{Paragraph, CodeFence, Paragraph},
		},
		{
			name:      "heading, blockquote, hr, paragraph",
			input:     "# Title\n\n> quote\n\n---\n\nText",
			wantTypes: []BlockType{Heading, BlockQuote, HorizontalRule, Paragraph},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			assertBlockTypes(t, blocks, tt.wantTypes)
		})
	}
}

// --- Edge cases ---

func TestParse_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
	}{
		{"empty string", "", 0},
		{"nil input", "", 0},
		{"single block no separators", "just plain text", 1},
		{"only whitespace", "   \n   \n   ", 0},
		{"single newline", "\n", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input []byte
			if tt.name == "nil input" {
				input = nil
			} else {
				input = []byte(tt.input)
			}
			blocks := Parse(input)
			if len(blocks) != tt.wantCount {
				t.Errorf("block count: got %d, want %d", len(blocks), tt.wantCount)
			}
		})
	}
}

// --- Invalid/malformed input tests ---

func TestParse_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"binary data", []byte{0x00, 0x01, 0xFF, 0xFE}},
		{"binary with text", append([]byte("hello"), 0x00, 0xFF, 0xFE)},
		{"unclosed code fence", []byte("```\ncode without closing fence")},
		{"deeply nested quotes", []byte(">>>>>>> very nested quote")},
		{"very long line", []byte(strings.Repeat("a", 100000))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Must not panic.
			blocks := Parse(tt.input)
			// Should return some blocks (or none), but never crash.
			_ = blocks
		})
	}
}

// --- Byte position accuracy tests ---

func TestParse_BytePositions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"paragraphs", "Hello\n\nWorld"},
		{"heading and paragraph", "# Title\n\nBody"},
		{"horizontal rule", "Above\n\n---\n\nBelow"},
		{"code fence", "Before\n\n```\ncode\n```\n\nAfter"},
		{"empty code fence", "Before\n\n```\n```\n\nAfter"},
		{"table", "Text\n\n| A | B |\n| --- | --- |\n| 1 | 2 |\n\nMore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			for i, b := range blocks {
				if b.StartByte < 0 || b.EndByte > len(tt.input) || b.StartByte > b.EndByte {
					t.Errorf("block[%d] (%s): invalid byte range [%d:%d] for input len %d",
						i, b.Type, b.StartByte, b.EndByte, len(tt.input))
					continue
				}
				sourceSlice := tt.input[b.StartByte:b.EndByte]
				if sourceSlice != b.Raw {
					t.Errorf("block[%d] (%s): source[%d:%d]=%q != Raw=%q",
						i, b.Type, b.StartByte, b.EndByte, sourceSlice, b.Raw)
				}
			}
		})
	}
}

// --- Trailing whitespace preservation tests ---

func TestParse_TrailingWhitespace(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantRaw string
	}{
		{
			name:    "paragraph with trailing spaces",
			input:   "Hello  ",
			wantRaw: "Hello  ",
		},
		{
			name:    "hard line break preserved",
			input:   "Line one  \nline two",
			wantRaw: "Line one  \nline two",
		},
		{
			name:    "trailing spaces in first of two paragraphs",
			input:   "Hello  \n\nWorld",
			wantRaw: "Hello  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := Parse([]byte(tt.input))
			if len(blocks) == 0 {
				t.Fatal("no blocks parsed")
			}
			if blocks[0].Raw != tt.wantRaw {
				t.Errorf("raw:\ngot:  %q\nwant: %q", blocks[0].Raw, tt.wantRaw)
			}
		})
	}
}

// --- BlockType.String() tests ---

func TestBlockType_String(t *testing.T) {
	tests := []struct {
		bt   BlockType
		want string
	}{
		{Paragraph, "Paragraph"},
		{Heading, "Heading"},
		{List, "List"},
		{CodeFence, "CodeFence"},
		{CodeBlock, "CodeBlock"},
		{BlockQuote, "BlockQuote"},
		{Table, "Table"},
		{HorizontalRule, "HorizontalRule"},
		{BlockType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.bt.String(); got != tt.want {
				t.Errorf("BlockType(%d).String() = %q, want %q", tt.bt, got, tt.want)
			}
		})
	}
}

// --- Performance test ---

func TestParse_Performance_10kWords(t *testing.T) {
	// Build a document with 10,000+ words across many paragraphs.
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		// ~50 words per paragraph
		sb.WriteString("Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur Excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum")
	}

	input := []byte(sb.String())
	words := len(strings.Fields(sb.String()))
	if words < 10000 {
		t.Fatalf("test document has only %d words, need 10,000+", words)
	}

	blocks := Parse(input)
	if len(blocks) != 200 {
		t.Errorf("expected 200 blocks, got %d", len(blocks))
	}
}

func BenchmarkParse_10kWords(b *testing.B) {
	var sb strings.Builder
	for i := 0; i < 200; i++ {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("Lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur")
	}
	input := []byte(sb.String())

	b.ResetTimer()
	for range b.N {
		Parse(input)
	}
}

// --- Test helpers ---

func assertBlockCount(t *testing.T, blocks []Block, want int) {
	t.Helper()
	if len(blocks) != want {
		t.Fatalf("block count: got %d, want %d (types: %v)", len(blocks), want, blockTypes(blocks))
	}
}

func assertBlockTypes(t *testing.T, blocks []Block, want []BlockType) {
	t.Helper()
	if len(blocks) != len(want) {
		t.Fatalf("block count: got %d, want %d (types: %v)", len(blocks), len(want), blockTypes(blocks))
	}
	for i, b := range blocks {
		if b.Type != want[i] {
			t.Errorf("block[%d] type: got %s, want %s", i, b.Type, want[i])
		}
	}
}

func assertBlockRaw(t *testing.T, blocks []Block, want []string) {
	t.Helper()
	if len(blocks) != len(want) {
		t.Fatalf("block count: got %d, want %d", len(blocks), len(want))
	}
	for i, b := range blocks {
		if b.Raw != want[i] {
			t.Errorf("block[%d] raw:\n  got:  %q\n  want: %q", i, b.Raw, want[i])
		}
	}
}

func blockTypes(blocks []Block) []string {
	types := make([]string, len(blocks))
	for i, b := range blocks {
		types[i] = b.Type.String()
	}
	return types
}

// TestParser_InvalidInput_NoPanic documents NFR11: goldmark handles any byte
// sequence gracefully without panicking. The test itself proves resilience —
// if Parse() panics, the test runner catches it as a failure.
func TestParser_InvalidInput_NoPanic(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty bytes", []byte{}},
		{"random binary garbage", []byte{0x00, 0xff, 0xfe, 0x80, 0x01, 0x7f}},
		{"malformed utf-8", []byte{0xc3, 0x28, 0xe2, 0x82, 0xac, 0xf0, 0x28}},
		{"deeply nested bold", []byte("**" + strings.Repeat("**bold**", 50) + "**")},
		{"unclosed code fence", []byte("```go\nfunc main() {}\n")},
		{"unclosed bold", []byte("**unclosed bold")},
		{"null bytes in text", []byte("hello\x00world")},
		{"only newlines", []byte("\n\n\n\n\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No assertion needed: if Parse panics the test runner reports it.
			_ = Parse(tt.input)
		})
	}
}
