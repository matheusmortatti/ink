// package block_test uses an external test package so it can import
// internal/render without creating an import cycle (render imports block).
package block_test

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
)

// Run with -update to generate or refresh fixture files:
//
//	go test ./internal/block/ -run TestGlamourFixtures -update
var update = flag.Bool("update", false, "update golden fixture files")

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

// glamourRender renders b through an 80-column light-mode Glamour renderer and
// returns the ANSI-stripped visual text. Width 80 and light style are fixed so
// fixtures are stable across terminal environments.
func glamourRender(t *testing.T, b block.Block) string {
	t.Helper()
	r, err := render.NewRenderer(80, false)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	out, err := r.Render(b)
	if err != nil {
		t.Fatalf("Render(%q): %v", b.Raw, err)
	}
	return stripANSI(out)
}

const fixtureDir = "testdata/glamour_output"

func fixturePath(name string) string {
	return filepath.Join(fixtureDir, name+".txt")
}

// checkFixture verifies that the actual Glamour output for b matches the stored
// fixture named name. With -update it writes/updates the fixture instead.
// Returns the fixture text (use this as the ground-truth rendered output in
// subsequent cursor-mapping assertions).
func checkFixture(t *testing.T, name string, b block.Block) string {
	t.Helper()
	actual := glamourRender(t, b)

	if *update {
		if err := os.MkdirAll(fixtureDir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", fixtureDir, err)
		}
		if err := os.WriteFile(fixturePath(name), []byte(actual), 0644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
		t.Logf("wrote fixture %s", fixturePath(name))
		return actual
	}

	stored, err := os.ReadFile(fixturePath(name))
	if err != nil {
		t.Fatalf("fixture %q not found — run:\n  go test ./internal/block/ -run TestGlamourFixtures -update\nerror: %v", name, err)
	}

	if actual != string(stored) {
		t.Errorf("Glamour output for block %q changed\ngot:\n%s\nwant:\n%s\n(run with -update to regenerate)", name, actual, string(stored))
	}
	return string(stored)
}

// visualLines splits the fixture text into lines.
func visualLines(fixture string) []string {
	return strings.Split(fixture, "\n")
}

// findVisualCol returns the visual (rune) column of the first occurrence of
// char in line (0-indexed). Returns -1 if not found.
// Uses rune index, not byte index, to handle multi-byte Unicode characters
// (e.g. │ U+2502 in blockquote renders, • U+2022 in list renders).
func findVisualCol(line, char string) int {
	idx := strings.Index(line, char)
	if idx < 0 {
		return -1
	}
	return len([]rune(line[:idx]))
}

// TestGlamourFixtures generates or verifies golden fixture files for every
// block type. Run once with -update to create testdata/glamour_output/*.txt;
// subsequent CI runs verify Glamour has not changed its output.
func TestGlamourFixtures(t *testing.T) {
	fixtures := []struct {
		name  string
		block block.Block
	}{
		// Paragraphs
		{"paragraph_plain", block.Block{Type: block.Paragraph, Raw: "hello world"}},
		{"paragraph_bold", block.Block{Type: block.Paragraph, Raw: "**bold** text"}},
		{"paragraph_italic", block.Block{Type: block.Paragraph, Raw: "*italic* text"}},
		{"paragraph_bold_italic", block.Block{Type: block.Paragraph, Raw: "***bold italic*** text"}},
		{"paragraph_code_span", block.Block{Type: block.Paragraph, Raw: "use `fmt.Println` here"}},
		{"paragraph_link", block.Block{Type: block.Paragraph, Raw: "[click here](https://example.com)"}},
		{"paragraph_escaped_backslash", block.Block{Type: block.Paragraph, Raw: `\\*not italic*`}},

		// Headings
		{"heading_h1", block.Block{Type: block.Heading, Raw: "# My Title", Level: 1}},
		{"heading_h2", block.Block{Type: block.Heading, Raw: "## My Title", Level: 2}},
		{"heading_h3", block.Block{Type: block.Heading, Raw: "### My Title", Level: 3}},
		{"heading_h4", block.Block{Type: block.Heading, Raw: "#### My Title", Level: 4}},
		{"heading_h5", block.Block{Type: block.Heading, Raw: "##### My Title", Level: 5}},
		{"heading_h6", block.Block{Type: block.Heading, Raw: "###### My Title", Level: 6}},
		{"heading_h2_with_bold", block.Block{Type: block.Heading, Raw: "## **Bold Title**", Level: 2}},

		// Code fence
		{"codefence_go", block.Block{Type: block.CodeFence, Raw: "```go\nfmt.Println(\"hello\")\n```"}},
		{"codefence_plain", block.Block{Type: block.CodeFence, Raw: "```\nline one\nline two\n```"}},

		// Block quotes
		{"blockquote_simple", block.Block{Type: block.BlockQuote, Raw: "> quoted text"}},
		// Consecutive "> lines" are joined by Glamour into one paragraph.
		// Use a blank "> " separator so the two lines render as distinct paragraphs.
		{"blockquote_multiline", block.Block{Type: block.BlockQuote, Raw: "> line one\n>\n> line two"}},
		{"blockquote_nested", block.Block{Type: block.BlockQuote, Raw: "> outer\n> > inner"}},

		// Lists
		{"list_unordered", block.Block{Type: block.List, Raw: "- item one\n- item two\n- item three"}},
		{"list_ordered", block.Block{Type: block.List, Raw: "1. first\n2. second\n3. third"}},
		{"list_bold_item", block.Block{Type: block.List, Raw: "- **bold** item\n- plain item"}},

		// Tables
		{"table_simple", block.Block{Type: block.Table, Raw: "| a | b |\n| --- | --- |\n| 1 | 2 |"}},
		{"table_wider", block.Block{Type: block.Table, Raw: "| Name | Value |\n| --- | --- |\n| foo | bar |"}},
	}

	for _, tt := range fixtures {
		t.Run(tt.name, func(t *testing.T) {
			checkFixture(t, tt.name, tt.block)
		})
	}
}

// ---------------------------------------------------------------------------
// Cursor mapping golden tests
//
// Each test below:
//   1. Verifies Glamour output matches the stored fixture.
//   2. Locates specific characters in the visual output.
//   3. Asserts MapRenderedToRaw and MapRawToRendered produce the correct
//      positions given the *actual* rendered layout.
// ---------------------------------------------------------------------------

func TestGlamourCursorMapping_PlainParagraph(t *testing.T) {
	b := block.Block{Type: block.Paragraph, Raw: "hello world"}
	fixture := checkFixture(t, "paragraph_plain", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}
	col := findVisualCol(lines[0], "h")
	if col < 0 {
		t.Fatalf("'h' not found in rendered line %q", lines[0])
	}

	// 'h' in rendered matches 'h' in raw at same column (no prefix stripped).
	rl, rc := block.MapRenderedToRaw(b, 0, col)
	if rl != 0 || rc != col {
		t.Errorf("MapRenderedToRaw(plain para, 0, %d) = (%d, %d), want (0, %d)", col, rl, rc, col)
	}
	// Round-trip
	rl2, rc2 := block.MapRawToRendered(b, rl, rc)
	if rl2 != 0 || rc2 != col {
		t.Errorf("MapRawToRendered round-trip failed: got (%d, %d), want (0, %d)", rl2, rc2, col)
	}
}

func TestGlamourCursorMapping_BoldParagraph(t *testing.T) {
	b := block.Block{Type: block.Paragraph, Raw: "**bold** text"}
	fixture := checkFixture(t, "paragraph_bold", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	// Glamour strips the ** markers: the visual output should NOT contain "**".
	if strings.Contains(lines[0], "**") {
		t.Errorf("rendered line contains literal **: %q — Glamour should strip bold markers", lines[0])
	}

	// 'b' is the first visible character of "bold".
	// In the raw source "**bold** text", 'b' is at raw col 2 (after the opening **).
	bCol := findVisualCol(lines[0], "b")
	if bCol < 0 {
		t.Fatalf("'b' not found in rendered line %q", lines[0])
	}

	rl, rc := block.MapRenderedToRaw(b, 0, bCol)
	if rl != 0 || rc != 2 {
		t.Errorf("MapRenderedToRaw(bold para, 0, %d ['b']) = (%d, %d), want (0, 2)", bCol, rl, rc)
	}

	// 't' in "text" — find it after "bold ".
	tIdx := strings.Index(lines[0], "text")
	if tIdx < 0 {
		t.Fatalf("'text' not found in rendered line %q", lines[0])
	}

	rl, rc = block.MapRenderedToRaw(b, 0, tIdx)
	if rl != 0 || rc != 9 {
		// Raw: **bold** text → 't' is at index 9
		t.Errorf("MapRenderedToRaw(bold para, 0, %d ['t']) = (%d, %d), want (0, 9)", tIdx, rl, rc)
	}
}

func TestGlamourCursorMapping_ItalicParagraph(t *testing.T) {
	b := block.Block{Type: block.Paragraph, Raw: "*italic* text"}
	fixture := checkFixture(t, "paragraph_italic", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	if strings.Contains(lines[0], "*italic*") {
		t.Errorf("rendered line contains raw italic markers: %q", lines[0])
	}

	// 'i' is the first char of "italic"; raw col 1 (after opening *).
	iCol := findVisualCol(lines[0], "i")
	if iCol < 0 {
		t.Fatalf("'i' not found in rendered line %q", lines[0])
	}

	rl, rc := block.MapRenderedToRaw(b, 0, iCol)
	if rl != 0 || rc != 1 {
		t.Errorf("MapRenderedToRaw(italic para, 0, %d) = (%d, %d), want (0, 1)", iCol, rl, rc)
	}
}

// TestGlamourCursorMapping_BoldItalic covers the triple-asterisk case that
// caused a bug (caught in code review during Story 2.5). Glamour renders
// ***text*** without the *** delimiters; the cursor map must skip all 3.
func TestGlamourCursorMapping_BoldItalic(t *testing.T) {
	b := block.Block{Type: block.Paragraph, Raw: "***bold italic*** text"}
	fixture := checkFixture(t, "paragraph_bold_italic", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	// Rendered output must not contain "***".
	if strings.Contains(lines[0], "***") {
		t.Errorf("rendered line still contains *** delimiters: %q", lines[0])
	}

	// 'b' is the first visible character; raw col 3 (after opening ***).
	bCol := findVisualCol(lines[0], "b")
	if bCol < 0 {
		t.Fatalf("'b' not found in rendered line %q", lines[0])
	}

	rl, rc := block.MapRenderedToRaw(b, 0, bCol)
	if rl != 0 || rc != 3 {
		t.Errorf("MapRenderedToRaw(bolditalic, 0, %d ['b']) = (%d, %d), want (0, 3)", bCol, rl, rc)
	}

	// 't' in trailing " text"; raw: ***bold italic*** text → 't' at index 18.
	tIdx := strings.Index(lines[0], " text")
	if tIdx < 0 {
		t.Fatalf("' text' not found in rendered line %q", lines[0])
	}
	tCol := tIdx + 1 // skip the space, land on 't'

	rl, rc = block.MapRenderedToRaw(b, 0, tCol)
	if rl != 0 || rc != 18 {
		// Raw: ***bold italic*** text → after *** (3) + "bold italic" (11) + *** (3) + " " (1) = 18
		t.Errorf("MapRenderedToRaw(bolditalic, 0, %d ['t' in text]) = (%d, %d), want (0, 18)", tCol, rl, rc)
	}
}

func TestGlamourCursorMapping_CodeSpan(t *testing.T) {
	b := block.Block{Type: block.Paragraph, Raw: "use `fmt.Println` here"}
	fixture := checkFixture(t, "paragraph_code_span", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	// The backtick delimiters should not appear as content.
	if strings.Contains(lines[0], "`fmt.Println`") {
		t.Errorf("rendered line contains raw backtick markers: %q", lines[0])
	}

	// 'f' in "fmt.Println" — raw col 5 (after "use `").
	fIdx := strings.Index(lines[0], "fmt")
	if fIdx < 0 {
		t.Fatalf("'fmt' not found in rendered line %q", lines[0])
	}

	rl, rc := block.MapRenderedToRaw(b, 0, fIdx)
	if rl != 0 || rc != 5 {
		// Raw: "use `fmt.Println` here" → 'f' is at index 5
		t.Errorf("MapRenderedToRaw(codespan, 0, %d ['f']) = (%d, %d), want (0, 5)", fIdx, rl, rc)
	}
}

// TestGlamourCursorMapping_EscapedBackslash covers the \\* case that caused a
// classification error (caught in code review during Story 2.5).
// Raw: \\*not italic* — the \\ escapes the second \, then * starts an italic span.
func TestGlamourCursorMapping_EscapedBackslash(t *testing.T) {
	b := block.Block{Type: block.Paragraph, Raw: `\\*not italic*`}
	fixture := checkFixture(t, "paragraph_escaped_backslash", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	// Rendered line must contain a visible backslash character.
	if !strings.Contains(lines[0], `\`) {
		t.Errorf("rendered line missing escaped backslash: %q", lines[0])
	}
}

// TestGlamourCursorMapping_Headings verifies Glamour heading layout and cursor mapping.
// H1: Glamour strips "# " (rendered starts with content at col 0).
// H2+: Glamour keeps "## ", "### " etc. as visible text (1:1 mapping in prefix zone).
func TestGlamourCursorMapping_Headings(t *testing.T) {
	tests := []struct {
		name       string
		raw        string
		level      int
		fixture    string
		prefixKept bool // true if Glamour keeps the ## prefix in rendered output
		// wantRawCol is the raw column of 'M' (first char of "My Title").
		wantRawCol int
	}{
		{"h1", "# My Title", 1, "heading_h1", false, 2},  // H1: prefix stripped, 'M' at rendered 0 → raw 2
		{"h2", "## My Title", 2, "heading_h2", true, 3},  // H2+: prefix kept, 'M' at rendered 3 → raw 3
		{"h3", "### My Title", 3, "heading_h3", true, 4},
		{"h4", "#### My Title", 4, "heading_h4", true, 5},
		{"h5", "##### My Title", 5, "heading_h5", true, 6},
		{"h6", "###### My Title", 6, "heading_h6", true, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := block.Block{Type: block.Heading, Raw: tt.raw, Level: tt.level}
			fixture := checkFixture(t, tt.fixture, b)
			lines := visualLines(fixture)

			if len(lines) == 0 {
				t.Fatal("fixture is empty")
			}

			rawPrefix := strings.Repeat("#", tt.level) + " "
			if tt.prefixKept {
				// H2+: rendered line must start with the raw prefix.
				if !strings.HasPrefix(lines[0], rawPrefix) {
					t.Errorf("expected rendered H%d to start with %q, got %q",
						tt.level, rawPrefix, lines[0])
				}
			} else {
				// H1: rendered line must NOT contain the raw prefix.
				if strings.HasPrefix(lines[0], rawPrefix) {
					t.Errorf("H1 rendered line still has raw prefix %q: %q", rawPrefix, lines[0])
				}
			}

			// 'M' is the first char of "My Title"; find its visual column.
			mCol := findVisualCol(lines[0], "M")
			if mCol < 0 {
				t.Fatalf("'M' not found in rendered line %q", lines[0])
			}

			// MapRenderedToRaw: visual col of 'M' → raw col of 'M'.
			rl, rc := block.MapRenderedToRaw(b, 0, mCol)
			if rl != 0 || rc != tt.wantRawCol {
				t.Errorf("MapRenderedToRaw(%s, 0, %d ['M']) = (%d, %d), want (0, %d)",
					tt.name, mCol, rl, rc, tt.wantRawCol)
			}

			// Round-trip.
			rl2, rc2 := block.MapRawToRendered(b, 0, tt.wantRawCol)
			if rl2 != 0 || rc2 != mCol {
				t.Errorf("MapRawToRendered(%s, 0, %d) = (%d, %d), want (0, %d)",
					tt.name, tt.wantRawCol, rl2, rc2, mCol)
			}
		})
	}
}

func TestGlamourCursorMapping_HeadingWithBold(t *testing.T) {
	// H2 with bold: Glamour keeps "## " visible, strips "**" bold markers.
	// Rendered: "## Bold Title" — col 0='#', col 3='B'.
	b := block.Block{Type: block.Heading, Raw: "## **Bold Title**", Level: 2}
	fixture := checkFixture(t, "heading_h2_with_bold", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	// "## " prefix must be present; "**" bold markers must be stripped.
	if !strings.HasPrefix(lines[0], "## ") {
		t.Errorf("expected rendered H2 to start with '## ', got %q", lines[0])
	}
	if strings.Contains(lines[0], "**") {
		t.Errorf("rendered line still contains ** bold markers: %q", lines[0])
	}

	bCol := findVisualCol(lines[0], "B")
	if bCol < 0 {
		t.Fatalf("'B' not found in rendered line %q", lines[0])
	}

	// 'B' is at rendered col 3 (after "## ").
	// Raw: "## **Bold Title**" → 'B' is at raw col 5 (after "## **").
	rl, rc := block.MapRenderedToRaw(b, 0, bCol)
	if rl != 0 || rc != 5 {
		t.Errorf("MapRenderedToRaw(bold heading, 0, %d ['B']) = (%d, %d), want (0, 5)", bCol, rl, rc)
	}
}

// TestGlamourCursorMapping_CodeFence verifies the Glamour code-fence layout
// (empty padding line and 1:1 content mapping) that the cursor map relies on.
func TestGlamourCursorMapping_CodeFence(t *testing.T) {
	b := block.Block{Type: block.CodeFence, Raw: "```go\nfmt.Println(\"hello\")\n```"}
	fixture := checkFixture(t, "codefence_go", b)
	lines := visualLines(fixture)

	// Find which rendered line contains the code content.
	codeLineIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "fmt.Println") {
			codeLineIdx = i
			break
		}
	}
	if codeLineIdx < 0 {
		t.Fatalf("code content not found in fixture:\n%s", fixture)
	}

	// MapRenderedToRaw: the code content line in rendered maps to raw line 1
	// (raw line 0 = ```go, raw line 1 = fmt.Println(...), raw line 2 = ```).
	rl, rc := block.MapRenderedToRaw(b, codeLineIdx, 0)
	if rl != 1 {
		t.Errorf("MapRenderedToRaw(codefence, rendered line %d, 0).line = %d, want 1 (code content is raw line 1)", codeLineIdx, rl)
	}
	_ = rc

	// MapRawToRendered: raw line 1 (content) should map back to the same
	// rendered line that contains the code.
	rl2, _ := block.MapRawToRendered(b, 1, 0)
	if rl2 != codeLineIdx {
		t.Errorf("MapRawToRendered(codefence, 1, 0).line = %d, want %d", rl2, codeLineIdx)
	}

	// Delimiter lines (raw 0 and raw 2) must not place the cursor in the
	// middle of code content.
	openLine, _ := block.MapRawToRendered(b, 0, 0)
	closeLine, _ := block.MapRawToRendered(b, 2, 0)
	if openLine == codeLineIdx {
		t.Errorf("opening ``` (raw line 0) mapped to rendered code line %d, want a delimiter/padding line", codeLineIdx)
	}
	if closeLine == codeLineIdx {
		t.Errorf("closing ``` (raw line 2) mapped to rendered code line %d, want a delimiter/padding line", codeLineIdx)
	}
}

func TestGlamourCursorMapping_BlockQuote(t *testing.T) {
	b := block.Block{Type: block.BlockQuote, Raw: "> quoted text"}
	fixture := checkFixture(t, "blockquote_simple", b)
	lines := visualLines(fixture)

	if len(lines) == 0 {
		t.Fatal("fixture is empty")
	}

	// Find rendered line containing "quoted text".
	quoteLineIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "quoted") {
			quoteLineIdx = i
			break
		}
	}
	if quoteLineIdx < 0 {
		t.Fatalf("'quoted' not found in fixture:\n%s", fixture)
	}

	qCol := findVisualCol(lines[quoteLineIdx], "q")
	if qCol < 0 {
		t.Fatalf("'q' not found in rendered line %q", lines[quoteLineIdx])
	}

	// 'q' in the rendered output maps to raw col 2 (after "> ").
	rl, rc := block.MapRenderedToRaw(b, quoteLineIdx, qCol)
	if rl != 0 || rc != 2 {
		t.Errorf("MapRenderedToRaw(blockquote, %d, %d ['q']) = (%d, %d), want (0, 2)", quoteLineIdx, qCol, rl, rc)
	}
}

func TestGlamourCursorMapping_BlockQuoteMultiLine(t *testing.T) {
	// Use blank-separated blockquote to force distinct rendered lines.
	// Raw: "> line one\n>\n> line two" → raw[0]="line one", raw[1]="" (blank), raw[2]="line two"
	b := block.Block{Type: block.BlockQuote, Raw: "> line one\n>\n> line two"}
	fixture := checkFixture(t, "blockquote_multiline", b)
	lines := visualLines(fixture)

	// Find rendered lines for "line one" and "line two".
	oneIdx, twoIdx := -1, -1
	for i, line := range lines {
		if strings.Contains(line, "line one") && oneIdx < 0 {
			oneIdx = i
		}
		if strings.Contains(line, "line two") && twoIdx < 0 {
			twoIdx = i
		}
	}
	if oneIdx < 0 || twoIdx < 0 {
		t.Fatalf("could not find both quote lines in fixture:\n%s", fixture)
	}
	if oneIdx == twoIdx {
		t.Fatalf("both lines on same rendered line %d — Glamour joined them (blank separator not working)", oneIdx)
	}

	// "line two" is raw line 2. 'l' at rendered col 2 (after "│ ") → raw (2, 2).
	lCol := findVisualCol(lines[twoIdx], "l")
	if lCol < 0 {
		t.Fatalf("'l' not found in rendered line %q", lines[twoIdx])
	}

	rl, rc := block.MapRenderedToRaw(b, twoIdx, lCol)
	if rl != 2 || rc != 2 {
		t.Errorf("MapRenderedToRaw(bq multiline, %d, %d ['l']) = (%d, %d), want (2, 2)", twoIdx, lCol, rl, rc)
	}
}

func TestGlamourCursorMapping_UnorderedList(t *testing.T) {
	// Use the same 3-item block as the list_unordered fixture.
	b := block.Block{Type: block.List, Raw: "- item one\n- item two\n- item three"}
	fixture := checkFixture(t, "list_unordered", b)
	lines := visualLines(fixture)

	// Find the rendered line for "item one".
	oneIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "item one") {
			oneIdx = i
			break
		}
	}
	if oneIdx < 0 {
		t.Fatalf("'item one' not found in fixture:\n%s", fixture)
	}

	iCol := findVisualCol(lines[oneIdx], "i")
	if iCol < 0 {
		t.Fatalf("'i' not found in rendered line %q", lines[oneIdx])
	}

	// 'i' in "item one" → raw col 2 (after "- ").
	rl, rc := block.MapRenderedToRaw(b, oneIdx, iCol)
	if rl != 0 || rc != 2 {
		t.Errorf("MapRenderedToRaw(list, %d, %d ['i']) = (%d, %d), want (0, 2)", oneIdx, iCol, rl, rc)
	}
}

func TestGlamourCursorMapping_OrderedList(t *testing.T) {
	// Use the same 3-item block as the list_ordered fixture.
	b := block.Block{Type: block.List, Raw: "1. first\n2. second\n3. third"}
	fixture := checkFixture(t, "list_ordered", b)
	lines := visualLines(fixture)

	firstIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "first") {
			firstIdx = i
			break
		}
	}
	if firstIdx < 0 {
		t.Fatalf("'first' not found in fixture:\n%s", fixture)
	}

	fCol := findVisualCol(lines[firstIdx], "f")
	if fCol < 0 {
		t.Fatalf("'f' not found in rendered line %q", lines[firstIdx])
	}

	// 'f' in "first" → raw col 3 (after "1. ").
	rl, rc := block.MapRenderedToRaw(b, firstIdx, fCol)
	if rl != 0 || rc != 3 {
		t.Errorf("MapRenderedToRaw(ordered list, %d, %d ['f']) = (%d, %d), want (0, 3)", firstIdx, fCol, rl, rc)
	}
}

// TestGlamourCursorMapping_Table covers the separator-row bug: the raw
// "| --- | --- |" row must not appear as a rendered line, so rendered line 1
// maps to raw line 2 (the first data row), skipping the separator.
func TestGlamourCursorMapping_Table(t *testing.T) {
	b := block.Block{Type: block.Table, Raw: "| a | b |\n| --- | --- |\n| 1 | 2 |"}
	fixture := checkFixture(t, "table_simple", b)
	lines := visualLines(fixture)

	// The separator row "| --- | --- |" must not appear in rendered output.
	for _, line := range lines {
		clean := strings.TrimSpace(line)
		if clean == "| --- | --- |" || strings.Contains(clean, "---") {
			t.Errorf("separator row leaked into rendered output: %q", line)
		}
	}

	// Find rendered lines containing header cell 'a' and data cell '1'.
	headerIdx, dataIdx := -1, -1
	for i, line := range lines {
		if strings.Contains(line, " a ") || (strings.Contains(line, "a") && strings.Contains(line, "b")) {
			headerIdx = i
		}
		if strings.Contains(line, " 1 ") || (strings.Contains(line, "1") && strings.Contains(line, "2")) {
			dataIdx = i
		}
	}
	if headerIdx < 0 || dataIdx < 0 {
		t.Fatalf("could not find both table rows in fixture (headerIdx=%d, dataIdx=%d):\n%s", headerIdx, dataIdx, fixture)
	}

	// Header row: rendered line headerIdx → raw line 0
	rl, _ := block.MapRenderedToRaw(b, headerIdx, 0)
	if rl != 0 {
		t.Errorf("MapRenderedToRaw(table header, %d, 0).line = %d, want 0", headerIdx, rl)
	}

	// Data row: rendered line dataIdx → raw line 2 (separator at raw line 1 is skipped)
	rl, _ = block.MapRenderedToRaw(b, dataIdx, 0)
	if rl != 2 {
		t.Errorf("MapRenderedToRaw(table data, %d, 0).line = %d, want 2 (separator skipped)", dataIdx, rl)
	}

	// Inverse: raw line 2 → rendered line dataIdx
	rl2, _ := block.MapRawToRendered(b, 2, 0)
	if rl2 != dataIdx {
		t.Errorf("MapRawToRendered(table, 2, 0).line = %d, want %d", rl2, dataIdx)
	}

	// Separator row (raw line 1) must not map to a line that shows data.
	sepLine, _ := block.MapRawToRendered(b, 1, 0)
	if sepLine == dataIdx {
		t.Errorf("separator row (raw line 1) mapped to data row rendered line %d", dataIdx)
	}
}
