package block

import "testing"

func TestClassifyRunes_PlainText(t *testing.T) {
	vis := classifyRunes("hello world")
	for i, v := range vis {
		if !v {
			t.Errorf("position %d should be visible", i)
		}
	}
}

func TestClassifyRunes_Bold(t *testing.T) {
	// **bold** → markers at 0,1 and 6,7
	vis := classifyRunes("**bold**")
	expected := []bool{false, false, true, true, true, true, false, false}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_Italic(t *testing.T) {
	// *italic* → markers at 0 and 7
	vis := classifyRunes("*italic*")
	expected := []bool{false, true, true, true, true, true, true, false}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_InlineCode(t *testing.T) {
	// `code` → markers at 0 and 5
	vis := classifyRunes("`code`")
	expected := []bool{false, true, true, true, true, false}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_Link(t *testing.T) {
	// [text](url) → '[' hidden, 'text' visible, '](url)' hidden
	vis := classifyRunes("[text](url)")
	//                    [  t    e    x    t    ]    (    u    r    l    )
	expected := []bool{false, true, true, true, true, false, false, false, false, false, false}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_NestedBoldItalic(t *testing.T) {
	// **bold *italic* more**
	vis := classifyRunes("**bold *italic* more**")
	// Positions:
	// 0,1: ** (hidden)
	// 2-6: "bold " (visible)
	// 7: * (hidden)
	// 8-13: "italic" (visible)
	// 14: * (hidden)
	// 15-19: " more" (visible)
	// 20,21: ** (hidden)
	expected := []bool{
		false, false, // **
		true, true, true, true, true, // "bold "
		false,                                    // *
		true, true, true, true, true, true,       // "italic"
		false,                                    // *
		true, true, true, true, true,             // " more"
		false, false, // **
	}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_Escaped(t *testing.T) {
	// \*not italic\* → backslashes hidden, asterisks visible
	vis := classifyRunes(`\*not italic\*`)
	if vis[0] != false { // backslash
		t.Error("position 0 (backslash) should be hidden")
	}
	if vis[1] != true { // escaped *
		t.Error("position 1 (escaped *) should be visible")
	}
}

func TestRawColFromVisualCol_PlainText(t *testing.T) {
	tests := []struct {
		raw       string
		visualCol int
		wantRaw   int
	}{
		{"hello world", 0, 0},
		{"hello world", 5, 5},
		{"hello world", 10, 10},
	}
	for _, tt := range tests {
		got := rawColFromVisualCol(tt.raw, tt.visualCol)
		if got != tt.wantRaw {
			t.Errorf("rawColFromVisualCol(%q, %d) = %d, want %d", tt.raw, tt.visualCol, got, tt.wantRaw)
		}
	}
}

func TestRawColFromVisualCol_Bold(t *testing.T) {
	// **bold** text → visual "bold text"
	tests := []struct {
		visualCol int
		wantRaw   int
	}{
		{0, 2},  // 'b' → after **
		{3, 5},  // 'd' → position 5 in raw
		{4, 8},  // ' ' → after **bold**
		{5, 9},  // 't' in "text"
	}
	for _, tt := range tests {
		got := rawColFromVisualCol("**bold** text", tt.visualCol)
		if got != tt.wantRaw {
			t.Errorf("rawColFromVisualCol(bold, %d) = %d, want %d", tt.visualCol, got, tt.wantRaw)
		}
	}
}

func TestRawColFromVisualCol_Link(t *testing.T) {
	// [click here](https://example.com) → visual "click here"
	tests := []struct {
		visualCol int
		wantRaw   int
	}{
		{0, 1},  // 'c' → after [
		{5, 6},  // ' ' in "click here"
		{9, 10}, // 'e' in "here"
	}
	for _, tt := range tests {
		got := rawColFromVisualCol("[click here](https://example.com)", tt.visualCol)
		if got != tt.wantRaw {
			t.Errorf("rawColFromVisualCol(link, %d) = %d, want %d", tt.visualCol, got, tt.wantRaw)
		}
	}
}

func TestVisualColFromRawCol_Bold(t *testing.T) {
	// **bold** text → visual "bold text"
	tests := []struct {
		rawCol  int
		wantVis int
	}{
		{0, 0},  // on first *
		{1, 0},  // on second *
		{2, 0},  // 'b' → visual 0
		{5, 3},  // 'd' → visual 3
		{6, 4},  // first closing *
		{8, 4},  // ' ' after **bold**
		{9, 5},  // 't' in "text"
	}
	for _, tt := range tests {
		got := visualColFromRawCol("**bold** text", tt.rawCol)
		if got != tt.wantVis {
			t.Errorf("visualColFromRawCol(bold, %d) = %d, want %d", tt.rawCol, got, tt.wantVis)
		}
	}
}

func TestMapRenderedToRaw_Paragraph(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "hello world"}
	tests := []struct {
		rLine, rCol int
		wLine, wCol int
	}{
		{0, 0, 0, 0},
		{0, 5, 0, 5},
		{0, 10, 0, 10},
	}
	for _, tt := range tests {
		rl, rc := MapRenderedToRaw(b, tt.rLine, tt.rCol)
		if rl != tt.wLine || rc != tt.wCol {
			t.Errorf("MapRenderedToRaw(para, %d, %d) = (%d, %d), want (%d, %d)",
				tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
		}
	}
}

func TestMapRenderedToRaw_ParagraphWithBold(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "**bold** text"}
	tests := []struct {
		rLine, rCol int
		wLine, wCol int
	}{
		{0, 0, 0, 2},  // 'b' → after **
		{0, 4, 0, 8},  // ' ' → after **bold**
		{0, 5, 0, 9},  // 't' in "text"
	}
	for _, tt := range tests {
		rl, rc := MapRenderedToRaw(b, tt.rLine, tt.rCol)
		if rl != tt.wLine || rc != tt.wCol {
			t.Errorf("MapRenderedToRaw(bold para, %d, %d) = (%d, %d), want (%d, %d)",
				tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
		}
	}
}

func TestMapRenderedToRaw_Heading(t *testing.T) {
	b := Block{Type: Heading, Raw: "## My Title", Level: 2}
	tests := []struct {
		name        string
		rLine, rCol int
		wLine, wCol int
	}{
		{"start", 0, 0, 0, 3},       // 'M' → after "## "
		{"middle", 0, 3, 0, 6},      // 'T' → position 6
		{"end", 0, 7, 0, 10},        // 'e' → position 10
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl, rc := MapRenderedToRaw(b, tt.rLine, tt.rCol)
			if rl != tt.wLine || rc != tt.wCol {
				t.Errorf("MapRenderedToRaw(heading, %d, %d) = (%d, %d), want (%d, %d)",
					tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
			}
		})
	}
}

func TestMapRenderedToRaw_HeadingLevel1(t *testing.T) {
	b := Block{Type: Heading, Raw: "# Title", Level: 1}
	rl, rc := MapRenderedToRaw(b, 0, 0)
	if rl != 0 || rc != 2 {
		t.Errorf("MapRenderedToRaw(h1, 0, 0) = (%d, %d), want (0, 2)", rl, rc)
	}
}

func TestMapRenderedToRaw_HeadingWithBold(t *testing.T) {
	b := Block{Type: Heading, Raw: "## **Bold Title**", Level: 2}
	// rendered: "Bold Title" (no ## prefix, no ** markers)
	rl, rc := MapRenderedToRaw(b, 0, 0) // 'B' → after "## " and "**"
	if rl != 0 || rc != 5 {
		t.Errorf("MapRenderedToRaw(bold heading, 0, 0) = (%d, %d), want (0, 5)", rl, rc)
	}
}

func TestMapRenderedToRaw_CodeFence(t *testing.T) {
	b := Block{Type: CodeFence, Raw: "```go\nfmt.Println(\"hello\")\n```"}
	tests := []struct {
		rLine, rCol int
		wLine, wCol int
	}{
		{1, 0, 1, 0},   // first content line (rendered line 1 due to Glamour padding) → raw line 1
		{1, 4, 1, 4},   // within first content line
	}
	for _, tt := range tests {
		rl, rc := MapRenderedToRaw(b, tt.rLine, tt.rCol)
		if rl != tt.wLine || rc != tt.wCol {
			t.Errorf("MapRenderedToRaw(codefence, %d, %d) = (%d, %d), want (%d, %d)",
				tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
		}
	}
}

func TestMapRenderedToRaw_BlockQuote(t *testing.T) {
	b := Block{Type: BlockQuote, Raw: "> quoted text"}
	rl, rc := MapRenderedToRaw(b, 0, 0) // 'q' → after "> "
	if rl != 0 || rc != 2 {
		t.Errorf("MapRenderedToRaw(blockquote, 0, 0) = (%d, %d), want (0, 2)", rl, rc)
	}
	rl, rc = MapRenderedToRaw(b, 0, 7) // 't' in "text" → col 9 in raw
	if rl != 0 || rc != 9 {
		t.Errorf("MapRenderedToRaw(blockquote, 0, 7) = (%d, %d), want (0, 9)", rl, rc)
	}
}

func TestMapRenderedToRaw_BlockQuoteMultiLine(t *testing.T) {
	b := Block{Type: BlockQuote, Raw: "> line one\n> line two"}
	rl, rc := MapRenderedToRaw(b, 1, 0)
	if rl != 1 || rc != 2 {
		t.Errorf("MapRenderedToRaw(bq line2, 1, 0) = (%d, %d), want (1, 2)", rl, rc)
	}
}

func TestMapRenderedToRaw_List(t *testing.T) {
	b := Block{Type: List, Raw: "- item one\n- item two"}
	tests := []struct {
		rLine, rCol int
		wLine, wCol int
	}{
		{0, 0, 0, 2},  // 'i' in "item one" → after "- "
		{0, 4, 0, 6},  // 'o' in "one"
		{1, 0, 1, 2},  // 'i' in "item two"
	}
	for _, tt := range tests {
		rl, rc := MapRenderedToRaw(b, tt.rLine, tt.rCol)
		if rl != tt.wLine || rc != tt.wCol {
			t.Errorf("MapRenderedToRaw(list, %d, %d) = (%d, %d), want (%d, %d)",
				tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
		}
	}
}

func TestMapRenderedToRaw_OrderedList(t *testing.T) {
	b := Block{Type: List, Raw: "1. first\n2. second"}
	rl, rc := MapRenderedToRaw(b, 0, 0) // 'f' → after "1. "
	if rl != 0 || rc != 3 {
		t.Errorf("MapRenderedToRaw(ol, 0, 0) = (%d, %d), want (0, 3)", rl, rc)
	}
}

func TestMapRenderedToRaw_Table(t *testing.T) {
	b := Block{Type: Table, Raw: "| a | b |\n| --- | --- |\n| 1 | 2 |"}
	tests := []struct {
		name        string
		rLine, rCol int
		wLine, wCol int
	}{
		{"header", 0, 0, 0, 2},      // header row, first cell
		{"data row", 1, 0, 2, 2},     // first data row → raw line 2 (skipping separator)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl, rc := MapRenderedToRaw(b, tt.rLine, tt.rCol)
			if rl != tt.wLine || rc != tt.wCol {
				t.Errorf("MapRenderedToRaw(table, %d, %d) = (%d, %d), want (%d, %d)",
					tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
			}
		})
	}
}

func TestMapRawToRendered_Table(t *testing.T) {
	b := Block{Type: Table, Raw: "| a | b |\n| --- | --- |\n| 1 | 2 |"}
	tests := []struct {
		name        string
		rLine, rCol int
		wLine, wCol int
	}{
		{"header", 0, 2, 0, 0},           // raw header → rendered line 0
		{"separator", 1, 0, 0, 0},        // separator row → maps to header (rendered 0)
		{"data row", 2, 2, 1, 0},         // raw line 2 → rendered line 1 (skipping separator)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl, rc := MapRawToRendered(b, tt.rLine, tt.rCol)
			if rl != tt.wLine || rc != tt.wCol {
				t.Errorf("MapRawToRendered(table, %d, %d) = (%d, %d), want (%d, %d)",
					tt.rLine, tt.rCol, rl, rc, tt.wLine, tt.wCol)
			}
		})
	}
}

func TestMapRawToRendered_Paragraph(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "hello world"}
	rl, rc := MapRawToRendered(b, 0, 5)
	if rl != 0 || rc != 5 {
		t.Errorf("MapRawToRendered(para, 0, 5) = (%d, %d), want (0, 5)", rl, rc)
	}
}

func TestMapRawToRendered_ParagraphWithBold(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "**bold** text"}
	tests := []struct {
		rawCol  int
		wantCol int
	}{
		{0, 0},  // on first **
		{2, 0},  // 'b' → visual 0
		{5, 3},  // 'd' → visual 3
		{8, 4},  // ' ' after **bold** → visual 4
		{9, 5},  // 't' in text → visual 5
	}
	for _, tt := range tests {
		rl, rc := MapRawToRendered(b, 0, tt.rawCol)
		if rl != 0 || rc != tt.wantCol {
			t.Errorf("MapRawToRendered(bold para, 0, %d) = (%d, %d), want (0, %d)",
				tt.rawCol, rl, rc, tt.wantCol)
		}
	}
}

func TestMapRawToRendered_Heading(t *testing.T) {
	b := Block{Type: Heading, Raw: "## My Title", Level: 2}
	tests := []struct {
		rawCol  int
		wantCol int
	}{
		{0, 0},  // on '#', maps to rendered 0
		{2, 0},  // on space after '##', maps to rendered 0
		{3, 0},  // 'M' → visual 0
		{6, 3},  // 'T' → visual 3
	}
	for _, tt := range tests {
		rl, rc := MapRawToRendered(b, 0, tt.rawCol)
		if rl != 0 || rc != tt.wantCol {
			t.Errorf("MapRawToRendered(heading, 0, %d) = (%d, %d), want (0, %d)",
				tt.rawCol, rl, rc, tt.wantCol)
		}
	}
}

func TestMapRawToRendered_CodeFence(t *testing.T) {
	b := Block{Type: CodeFence, Raw: "```go\nfmt.Println()\n```"}
	tests := []struct {
		rawLine  int
		wantLine int
	}{
		{0, 0},  // opening ``` → rendered 0 (padding line)
		{1, 1},  // content line → rendered 1 (1:1 mapping)
		{2, 0},  // closing ``` → rendered 0 (delimiter)
	}
	for _, tt := range tests {
		rl, _ := MapRawToRendered(b, tt.rawLine, 0)
		if rl != tt.wantLine {
			t.Errorf("MapRawToRendered(codefence, %d, 0) line = %d, want %d",
				tt.rawLine, rl, tt.wantLine)
		}
	}
}

func TestMapRawToRendered_BlockQuote(t *testing.T) {
	b := Block{Type: BlockQuote, Raw: "> some text"}
	tests := []struct {
		rawCol  int
		wantCol int
	}{
		{0, 0},  // '>' → rendered 0
		{1, 0},  // ' ' after '>' → rendered 0
		{2, 0},  // 's' → rendered 0
		{5, 3},  // 'e' → rendered 3
	}
	for _, tt := range tests {
		rl, rc := MapRawToRendered(b, 0, tt.rawCol)
		if rl != 0 || rc != tt.wantCol {
			t.Errorf("MapRawToRendered(bq, 0, %d) = (%d, %d), want (0, %d)",
				tt.rawCol, rl, rc, tt.wantCol)
		}
	}
}

func TestMapRawToRendered_List(t *testing.T) {
	b := Block{Type: List, Raw: "- item text"}
	tests := []struct {
		rawCol  int
		wantCol int
	}{
		{0, 0},  // '-' → rendered 0
		{1, 0},  // ' ' → rendered 0
		{2, 0},  // 'i' → rendered 0
		{6, 4},  // ' ' before "text" → rendered 4
	}
	for _, tt := range tests {
		rl, rc := MapRawToRendered(b, 0, tt.rawCol)
		if rl != 0 || rc != tt.wantCol {
			t.Errorf("MapRawToRendered(list, 0, %d) = (%d, %d), want (0, %d)",
				tt.rawCol, rl, rc, tt.wantCol)
		}
	}
}

// Bidirectional consistency tests
func TestBidirectional_Paragraph(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "hello world"}
	for col := 0; col <= 10; col++ {
		rl, rc := MapRenderedToRaw(b, 0, col)
		rl2, rc2 := MapRawToRendered(b, rl, rc)
		if rl2 != 0 || rc2 != col {
			t.Errorf("round-trip para col %d: got (%d, %d) → (%d, %d) → (%d, %d)",
				col, rl, rc, rl2, rc2, 0, col)
		}
	}
}

func TestBidirectional_Heading(t *testing.T) {
	b := Block{Type: Heading, Raw: "## My Title", Level: 2}
	// Test round-trip for visual positions (rendered → raw → rendered)
	for col := 0; col < 8; col++ {
		rl, rc := MapRenderedToRaw(b, 0, col)
		rl2, rc2 := MapRawToRendered(b, rl, rc)
		if rl2 != 0 || rc2 != col {
			t.Errorf("round-trip heading col %d: rendered→raw=(%d,%d) raw→rendered=(%d,%d) want (0,%d)",
				col, rl, rc, rl2, rc2, col)
		}
	}
}

func TestBidirectional_Bold(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "**bold** text"}
	// Visual: "bold text" (9 chars)
	for col := 0; col < 9; col++ {
		rl, rc := MapRenderedToRaw(b, 0, col)
		rl2, rc2 := MapRawToRendered(b, rl, rc)
		if rl2 != 0 || rc2 != col {
			t.Errorf("round-trip bold col %d: rendered→raw=(%d,%d) raw→rendered=(%d,%d) want (0,%d)",
				col, rl, rc, rl2, rc2, col)
		}
	}
}

func TestBidirectional_CodeFence(t *testing.T) {
	b := Block{Type: CodeFence, Raw: "```\nline1\nline2\n```"}
	// Glamour adds padding line at rendered 0; content starts at rendered 1.
	// rendered line 1 = raw line 1, rendered line 2 = raw line 2
	for rLine := 1; rLine <= 2; rLine++ {
		rl, rc := MapRenderedToRaw(b, rLine, 0)
		rl2, rc2 := MapRawToRendered(b, rl, rc)
		if rl2 != rLine || rc2 != 0 {
			t.Errorf("round-trip codefence line %d: raw=(%d,%d) back=(%d,%d) want (%d,0)",
				rLine, rl, rc, rl2, rc2, rLine)
		}
	}
}

// Edge cases
func TestMapRenderedToRaw_EmptyBlock(t *testing.T) {
	b := Block{Type: Paragraph, Raw: ""}
	rl, rc := MapRenderedToRaw(b, 0, 0)
	if rl != 0 || rc != 0 {
		t.Errorf("MapRenderedToRaw(empty, 0, 0) = (%d, %d), want (0, 0)", rl, rc)
	}
}

func TestMapRenderedToRaw_SingleChar(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "x"}
	rl, rc := MapRenderedToRaw(b, 0, 0)
	if rl != 0 || rc != 0 {
		t.Errorf("MapRenderedToRaw(single, 0, 0) = (%d, %d), want (0, 0)", rl, rc)
	}
}

func TestMapRenderedToRaw_CursorAtEndOfLine(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "hello"}
	rl, rc := MapRenderedToRaw(b, 0, 5)
	if rl != 0 || rc != 5 {
		t.Errorf("MapRenderedToRaw(end, 0, 5) = (%d, %d), want (0, 5)", rl, rc)
	}
}

func TestMapRenderedToRaw_CursorBeyondContent(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "hi"}
	rl, rc := MapRenderedToRaw(b, 0, 100)
	// Should clamp to end of line
	if rl != 0 || rc != 2 {
		t.Errorf("MapRenderedToRaw(beyond, 0, 100) = (%d, %d), want (0, 2)", rl, rc)
	}
}

func TestMapRenderedToRaw_LineBeyondContent(t *testing.T) {
	b := Block{Type: Paragraph, Raw: "single line"}
	rl, _ := MapRenderedToRaw(b, 5, 0)
	// Should clamp to last line
	if rl != 0 {
		t.Errorf("MapRenderedToRaw(line beyond, 5, 0) line = %d, want 0", rl)
	}
}

func TestMapRawToRendered_EmptyBlock(t *testing.T) {
	b := Block{Type: Paragraph, Raw: ""}
	rl, rc := MapRawToRendered(b, 0, 0)
	if rl != 0 || rc != 0 {
		t.Errorf("MapRawToRendered(empty, 0, 0) = (%d, %d), want (0, 0)", rl, rc)
	}
}

// Inline marker edge cases
func TestClassifyRunes_AdjacentBoldItalic(t *testing.T) {
	// **bold***italic*
	vis := classifyRunes("**bold***italic*")
	// **bold** → 0,1 hidden, 2-5 visible, 6,7 hidden
	// *italic* → 8 hidden, 9-14 visible, 15 hidden
	// Wait, this is tricky: ***italic* could be **  then *italic*
	// Let me verify: the parser should see **bold** first (findClosingDelim for ** finds position 6)
	// Then *italic* (findClosingDelim for * finds position 15)
	if !vis[2] { // 'b' should be visible
		t.Error("position 2 ('b') should be visible")
	}
	if vis[0] { // first * should be hidden
		t.Error("position 0 ('*') should be hidden")
	}
}

func TestClassifyRunes_BoldAtLineStart(t *testing.T) {
	vis := classifyRunes("**text**")
	if vis[0] || vis[1] {
		t.Error("opening ** should be hidden")
	}
	if !vis[2] || !vis[3] || !vis[4] || !vis[5] {
		t.Error("text content should be visible")
	}
	if vis[6] || vis[7] {
		t.Error("closing ** should be hidden")
	}
}

func TestClassifyRunes_TripleAsterisk(t *testing.T) {
	// ***bold italic*** → all 6 asterisks hidden, "bold italic" visible
	vis := classifyRunes("***bold italic***")
	// Positions: 0,1,2 = *** (hidden), 3-14 = "bold italic" (visible), 15,16,17 = *** (hidden) -- wait
	// ***bold italic*** = 18 chars
	// 0,1,2: *** hidden
	// 3-13: "bold italic" visible (11 chars)
	// 14,15,16: *** hidden
	expected := []bool{
		false, false, false, // ***
		true, true, true, true, true, true, true, true, true, true, true, // "bold italic"
		false, false, false, // ***
	}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_Image(t *testing.T) {
	// ![alt text](image.png) → '!' hidden, '[' hidden, 'alt text' visible, '](image.png)' hidden
	vis := classifyRunes("![alt text](image.png)")
	//                    !  [  a    l    t    _    t    e    x    t    ]    (    i    m    a    g    e    .    p    n    g    )
	expected := []bool{false, false, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false, false, false}
	if len(vis) != len(expected) {
		t.Fatalf("length mismatch: got %d, want %d", len(vis), len(expected))
	}
	for i := range vis {
		if vis[i] != expected[i] {
			t.Errorf("position %d: got %v, want %v", i, vis[i], expected[i])
		}
	}
}

func TestClassifyRunes_UnmatchedBold(t *testing.T) {
	// **unclosed bold → no closing **, all chars stay visible
	vis := classifyRunes("**unclosed bold")
	for i, v := range vis {
		if !v {
			t.Errorf("position %d should be visible for unmatched marker", i)
		}
	}
}

func TestClassifyRunes_EscapedBackslash(t *testing.T) {
	// \\*italic* → backslash escapes backslash, *italic* is italic
	vis := classifyRunes(`\\*italic*`)
	// \(0) \(1) *(2) i(3) t(4) a(5) l(6) i(7) c(8) *(9)
	// Position 0: backslash, hidden (escaping pos 1)
	// Position 1: escaped backslash, visible
	// Position 2: italic opener, hidden
	// Position 3-8: "italic", visible
	// Position 9: italic closer, hidden
	if vis[0] != false {
		t.Error("position 0 (first backslash) should be hidden")
	}
	if vis[1] != true {
		t.Error("position 1 (escaped backslash) should be visible")
	}
	if vis[2] != false {
		t.Error("position 2 (italic *) should be hidden")
	}
	if vis[9] != false {
		t.Error("position 9 (closing *) should be hidden")
	}
}

func TestClassifyRunes_BoldAtLineEnd(t *testing.T) {
	vis := classifyRunes("abc **text**")
	if !vis[0] || !vis[1] || !vis[2] || !vis[3] { // "abc "
		t.Error("leading text should be visible")
	}
	if vis[4] || vis[5] { // opening **
		t.Error("opening ** should be hidden")
	}
}

// Helper functions tests
func TestBlockQuotePrefixLen(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"> text", 2},
		{">text", 1},
		{"> > nested", 4},
		{"> > > deep", 6},
		{"text", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := blockQuotePrefixLen(tt.line)
		if got != tt.want {
			t.Errorf("blockQuotePrefixLen(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

func TestListPrefixLen(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"- item", 2},
		{"* item", 2},
		{"+ item", 2},
		{"1. item", 3},
		{"10. item", 4},
		{"  - nested", 4},
		{"text", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := listPrefixLen(tt.line)
		if got != tt.want {
			t.Errorf("listPrefixLen(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}
