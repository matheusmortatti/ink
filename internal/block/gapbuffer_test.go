package block

import (
	"strings"
	"testing"
)

// --- Constructor Tests ---

func TestGapBuffer_NewEmpty_ContentEmpty(t *testing.T) {
	g := NewGapBuffer("")
	if got := g.Content(); got != "" {
		t.Errorf("Content() = %q, want %q", got, "")
	}
	if got := g.CursorPos(); got != 0 {
		t.Errorf("CursorPos() = %d, want 0", got)
	}
	if got := g.Length(); got != 0 {
		t.Errorf("Length() = %d, want 0", got)
	}
}

func TestGapBuffer_NewWithContent_ContentPreserved(t *testing.T) {
	g := NewGapBuffer("hello")
	if got := g.Content(); got != "hello" {
		t.Errorf("Content() = %q, want %q", got, "hello")
	}
	if got := g.CursorPos(); got != 5 {
		t.Errorf("CursorPos() = %d, want 5", got)
	}
	if got := g.Length(); got != 5 {
		t.Errorf("Length() = %d, want 5", got)
	}
}

func TestGapBuffer_NewMultiLine_ContentPreserved(t *testing.T) {
	text := "line one\nline two\nline three"
	g := NewGapBuffer(text)
	if got := g.Content(); got != text {
		t.Errorf("Content() = %q, want %q", got, text)
	}
	if got := g.LineCount(); got != 3 {
		t.Errorf("LineCount() = %d, want 3", got)
	}
}

// --- Insert Tests ---

func TestGapBuffer_InsertSingle_ContentReflectsChar(t *testing.T) {
	g := NewGapBuffer("")
	g.Insert('a')
	if got := g.Content(); got != "a" {
		t.Errorf("Content() = %q, want %q", got, "a")
	}
}

func TestGapBuffer_InsertSequential_PreservesOrder(t *testing.T) {
	g := NewGapBuffer("")
	g.Insert('a')
	g.Insert('b')
	g.Insert('c')
	if got := g.Content(); got != "abc" {
		t.Errorf("Content() = %q, want %q", got, "abc")
	}
}

func TestGapBuffer_InsertMidText_InsertsAtCursor(t *testing.T) {
	g := NewGapBuffer("ac")
	g.MoveLeft() // cursor before 'c'
	g.Insert('b')
	if got := g.Content(); got != "abc" {
		t.Errorf("Content() = %q, want %q", got, "abc")
	}
}

func TestGapBuffer_InsertString_ContentReflectsString(t *testing.T) {
	g := NewGapBuffer("")
	g.InsertString("hello")
	if got := g.Content(); got != "hello" {
		t.Errorf("Content() = %q, want %q", got, "hello")
	}
}

func TestGapBuffer_InsertStringMidText_InsertsAtCursor(t *testing.T) {
	g := NewGapBuffer("hd")
	g.MoveLeft()
	g.InsertString("ello worl")
	if got := g.Content(); got != "hello world" {
		t.Errorf("Content() = %q, want %q", got, "hello world")
	}
}

// --- Backspace Tests ---

func TestGapBuffer_BackspaceAtStart_ReturnsFalse(t *testing.T) {
	g := NewGapBuffer("abc")
	g.MoveToStart()
	if got := g.Backspace(); got != false {
		t.Errorf("Backspace() = %v, want false", got)
	}
	if got := g.Content(); got != "abc" {
		t.Errorf("Content() = %q, want %q", got, "abc")
	}
}

func TestGapBuffer_BackspaceMidText_DeletesPreviousChar(t *testing.T) {
	g := NewGapBuffer("abc")
	g.SetCursorPos(2) // cursor after 'b'
	g.Backspace()
	if got := g.Content(); got != "ac" {
		t.Errorf("Content() = %q, want %q", got, "ac")
	}
}

func TestGapBuffer_BackspaceAtEnd_DeletesLastChar(t *testing.T) {
	g := NewGapBuffer("abc")
	g.Backspace()
	if got := g.Content(); got != "ab" {
		t.Errorf("Content() = %q, want %q", got, "ab")
	}
}

// --- Delete Tests ---

func TestGapBuffer_DeleteAtEnd_ReturnsFalse(t *testing.T) {
	g := NewGapBuffer("abc")
	if got := g.Delete(); got != false {
		t.Errorf("Delete() = %v, want false", got)
	}
	if got := g.Content(); got != "abc" {
		t.Errorf("Content() = %q, want %q", got, "abc")
	}
}

func TestGapBuffer_DeleteMidText_DeletesNextChar(t *testing.T) {
	g := NewGapBuffer("abc")
	g.SetCursorPos(1)
	g.Delete()
	if got := g.Content(); got != "ac" {
		t.Errorf("Content() = %q, want %q", got, "ac")
	}
}

func TestGapBuffer_DeleteAtStart_DeletesFirstChar(t *testing.T) {
	g := NewGapBuffer("abc")
	g.MoveToStart()
	g.Delete()
	if got := g.Content(); got != "bc" {
		t.Errorf("Content() = %q, want %q", got, "bc")
	}
}

// --- MoveLeft / MoveRight Tests ---

func TestGapBuffer_MoveLeftAtStart_ReturnsFalse(t *testing.T) {
	g := NewGapBuffer("abc")
	g.MoveToStart()
	if got := g.MoveLeft(); got != false {
		t.Errorf("MoveLeft() = %v, want false", got)
	}
}

func TestGapBuffer_MoveRightAtEnd_ReturnsFalse(t *testing.T) {
	g := NewGapBuffer("abc")
	if got := g.MoveRight(); got != false {
		t.Errorf("MoveRight() = %v, want false", got)
	}
}

func TestGapBuffer_MoveLeft_UpdatesCursorAndInsertPosition(t *testing.T) {
	g := NewGapBuffer("abc")
	g.MoveLeft()
	if got := g.CursorPos(); got != 2 {
		t.Errorf("CursorPos() = %d, want 2", got)
	}
	g.Insert('X')
	if got := g.Content(); got != "abXc" {
		t.Errorf("Content() = %q, want %q", got, "abXc")
	}
}

func TestGapBuffer_MoveRight_UpdatesCursor(t *testing.T) {
	g := NewGapBuffer("abc")
	g.MoveToStart()
	g.MoveRight()
	if got := g.CursorPos(); got != 1 {
		t.Errorf("CursorPos() = %d, want 1", got)
	}
	g.Insert('X')
	if got := g.Content(); got != "aXbc" {
		t.Errorf("Content() = %q, want %q", got, "aXbc")
	}
}

// --- MoveToStart / MoveToEnd Tests ---

func TestGapBuffer_MoveToStart_CursorAtZero(t *testing.T) {
	g := NewGapBuffer("hello")
	g.MoveToStart()
	if got := g.CursorPos(); got != 0 {
		t.Errorf("CursorPos() = %d, want 0", got)
	}
	g.Insert('X')
	if got := g.Content(); got != "Xhello" {
		t.Errorf("Content() = %q, want %q", got, "Xhello")
	}
}

func TestGapBuffer_MoveToEnd_CursorAtLength(t *testing.T) {
	g := NewGapBuffer("hello")
	g.MoveToStart()
	g.MoveToEnd()
	if got := g.CursorPos(); got != 5 {
		t.Errorf("CursorPos() = %d, want 5", got)
	}
	g.Insert('X')
	if got := g.Content(); got != "helloX" {
		t.Errorf("Content() = %q, want %q", got, "helloX")
	}
}

// --- CursorPos / SetCursorPos Tests ---

func TestGapBuffer_SetCursorPos_MovesToPosition(t *testing.T) {
	g := NewGapBuffer("hello")
	g.SetCursorPos(2)
	if got := g.CursorPos(); got != 2 {
		t.Errorf("CursorPos() = %d, want 2", got)
	}
	g.Insert('X')
	if got := g.Content(); got != "heXllo" {
		t.Errorf("Content() = %q, want %q", got, "heXllo")
	}
}

func TestGapBuffer_SetCursorPos_ClampsOutOfRange(t *testing.T) {
	g := NewGapBuffer("hi")
	g.SetCursorPos(-5)
	if got := g.CursorPos(); got != 0 {
		t.Errorf("CursorPos() after -5 = %d, want 0", got)
	}
	g.SetCursorPos(100)
	if got := g.CursorPos(); got != 2 {
		t.Errorf("CursorPos() after 100 = %d, want 2", got)
	}
}

func TestGapBuffer_CursorPosAfterInserts(t *testing.T) {
	g := NewGapBuffer("")
	g.Insert('a')
	if got := g.CursorPos(); got != 1 {
		t.Errorf("CursorPos() = %d, want 1", got)
	}
	g.Insert('b')
	if got := g.CursorPos(); got != 2 {
		t.Errorf("CursorPos() = %d, want 2", got)
	}
}

func TestGapBuffer_CursorPosAfterMoves(t *testing.T) {
	g := NewGapBuffer("abcde")
	g.MoveLeft()
	g.MoveLeft()
	if got := g.CursorPos(); got != 3 {
		t.Errorf("CursorPos() = %d, want 3", got)
	}
	g.MoveRight()
	if got := g.CursorPos(); got != 4 {
		t.Errorf("CursorPos() = %d, want 4", got)
	}
}

// --- MoveUp / MoveDown Tests ---

func TestGapBuffer_MoveUp_FirstLine_ReturnsFalse(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	g.MoveToStart()
	if got := g.MoveUp(); got != false {
		t.Errorf("MoveUp() on first line = %v, want false", got)
	}
}

func TestGapBuffer_MoveDown_LastLine_ReturnsFalse(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	// cursor at end, on last line
	if got := g.MoveDown(); got != false {
		t.Errorf("MoveDown() on last line = %v, want false", got)
	}
}

func TestGapBuffer_MoveDown_SameColumn(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	g.SetCursorLineCol(0, 1) // line 0, col 1 (after 'a')
	g.MoveDown()
	line, col := g.CursorLineCol()
	if line != 1 || col != 1 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 1)", line, col)
	}
}

func TestGapBuffer_MoveUp_SameColumn(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	g.SetCursorLineCol(2, 2) // line 2, col 2 (after 'gh')
	g.MoveUp()
	line, col := g.CursorLineCol()
	if line != 1 || col != 2 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 2)", line, col)
	}
}

func TestGapBuffer_MoveDown_ShorterLineClampsCol(t *testing.T) {
	g := NewGapBuffer("abcdef\nhi\njklmno")
	g.SetCursorLineCol(0, 5) // col 5 on first line
	g.MoveDown()
	line, col := g.CursorLineCol()
	if line != 1 {
		t.Errorf("line = %d, want 1", line)
	}
	if col != 2 {
		t.Errorf("col = %d, want 2 (clamped to end of shorter line)", col)
	}
}

// --- MoveToLineStart / MoveToLineEnd Tests ---

func TestGapBuffer_MoveToLineStart_SingleLine(t *testing.T) {
	g := NewGapBuffer("hello")
	g.MoveToLineStart()
	if got := g.CursorPos(); got != 0 {
		t.Errorf("CursorPos() = %d, want 0", got)
	}
}

func TestGapBuffer_MoveToLineEnd_SingleLine(t *testing.T) {
	g := NewGapBuffer("hello")
	g.MoveToStart()
	g.MoveToLineEnd()
	if got := g.CursorPos(); got != 5 {
		t.Errorf("CursorPos() = %d, want 5", got)
	}
}

func TestGapBuffer_MoveToLineStart_MultiLine(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	g.SetCursorLineCol(1, 2)
	g.MoveToLineStart()
	line, col := g.CursorLineCol()
	if line != 1 || col != 0 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 0)", line, col)
	}
}

func TestGapBuffer_MoveToLineEnd_MultiLine(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	g.SetCursorLineCol(1, 0) // start of second line
	g.MoveToLineEnd()
	line, col := g.CursorLineCol()
	if line != 1 || col != 3 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 3)", line, col)
	}
	// Verify cursor stopped before '\n' by inserting and checking content
	g.Insert('X')
	if got := g.Content(); got != "abc\ndefX\nghi" {
		t.Errorf("Content() = %q, want %q", got, "abc\ndefX\nghi")
	}
}

// --- CursorLineCol / SetCursorLineCol Tests ---

func TestGapBuffer_CursorLineCol_AtStart(t *testing.T) {
	g := NewGapBuffer("abc\ndef")
	g.MoveToStart()
	line, col := g.CursorLineCol()
	if line != 0 || col != 0 {
		t.Errorf("CursorLineCol() = (%d, %d), want (0, 0)", line, col)
	}
}

func TestGapBuffer_CursorLineCol_AtEnd(t *testing.T) {
	g := NewGapBuffer("abc\ndef")
	line, col := g.CursorLineCol()
	if line != 1 || col != 3 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 3)", line, col)
	}
}

func TestGapBuffer_CursorLineCol_MidSecondLine(t *testing.T) {
	g := NewGapBuffer("abc\ndef\nghi")
	g.SetCursorLineCol(1, 1)
	line, col := g.CursorLineCol()
	if line != 1 || col != 1 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 1)", line, col)
	}
}

func TestGapBuffer_SetCursorLineCol_BeyondLastLine(t *testing.T) {
	g := NewGapBuffer("abc\ndef")
	g.SetCursorLineCol(10, 0)
	line, _ := g.CursorLineCol()
	if line != 1 {
		t.Errorf("line = %d, want 1 (clamped to last line)", line)
	}
}

// --- Content Round-Trip Tests ---

func TestGapBuffer_ContentRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"single line", "hello world"},
		{"multi-line paragraph", "First line.\nSecond line.\nThird line."},
		{"code fence", "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```"},
		{"list items", "- item one\n- item two\n- item three"},
		{"table", "| col1 | col2 |\n|------|------|\n| a    | b    |"},
		{"unicode text", "Hello 世界 🌍 café naïve"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGapBuffer(tt.content)
			if got := g.Content(); got != tt.content {
				t.Errorf("Content() = %q, want %q", got, tt.content)
			}
		})
	}
}

// --- Length / LineCount Tests ---

func TestGapBuffer_Length(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"empty", "", 0},
		{"single char", "a", 1},
		{"hello", "hello", 5},
		{"multi-line", "ab\ncd", 5},
		{"emoji", "🌍", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGapBuffer(tt.content)
			if got := g.Length(); got != tt.want {
				t.Errorf("Length() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGapBuffer_LengthAfterInsertDelete(t *testing.T) {
	g := NewGapBuffer("abc")
	if got := g.Length(); got != 3 {
		t.Errorf("Length() = %d, want 3", got)
	}
	g.Insert('d')
	if got := g.Length(); got != 4 {
		t.Errorf("Length() after insert = %d, want 4", got)
	}
	g.Backspace()
	if got := g.Length(); got != 3 {
		t.Errorf("Length() after backspace = %d, want 3", got)
	}
}

func TestGapBuffer_LineCount(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"empty", "", 1},
		{"single line", "hello", 1},
		{"two lines", "ab\ncd", 2},
		{"three lines", "a\nb\nc", 3},
		{"trailing newline", "abc\n", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGapBuffer(tt.content)
			if got := g.LineCount(); got != tt.want {
				t.Errorf("LineCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGapBuffer_LineCountAfterEdits(t *testing.T) {
	g := NewGapBuffer("abc")
	if got := g.LineCount(); got != 1 {
		t.Errorf("LineCount() = %d, want 1", got)
	}
	// Insert a newline — should become 2 lines
	g.Insert('\n')
	if got := g.LineCount(); got != 2 {
		t.Errorf("LineCount() after inserting newline = %d, want 2", got)
	}
	// Insert another newline
	g.Insert('\n')
	if got := g.LineCount(); got != 3 {
		t.Errorf("LineCount() after second newline = %d, want 3", got)
	}
	// Backspace a newline — should go back to 2 lines
	g.Backspace()
	if got := g.LineCount(); got != 2 {
		t.Errorf("LineCount() after backspace newline = %d, want 2", got)
	}
}

// --- Unicode Tests ---

func TestGapBuffer_Unicode_Emoji(t *testing.T) {
	g := NewGapBuffer("a🌍b")
	if got := g.Length(); got != 3 {
		t.Errorf("Length() = %d, want 3", got)
	}
	// Backspace should remove the 'b' (cursor at end)
	g.Backspace()
	if got := g.Content(); got != "a🌍" {
		t.Errorf("Content() after backspace = %q, want %q", got, "a🌍")
	}
	// Backspace should remove the emoji as one unit
	g.Backspace()
	if got := g.Content(); got != "a" {
		t.Errorf("Content() after second backspace = %q, want %q", got, "a")
	}
}

func TestGapBuffer_Unicode_CJK(t *testing.T) {
	g := NewGapBuffer("你好世界")
	if got := g.Length(); got != 4 {
		t.Errorf("Length() = %d, want 4", got)
	}
	g.SetCursorPos(2)
	g.Insert('X')
	if got := g.Content(); got != "你好X世界" {
		t.Errorf("Content() = %q, want %q", got, "你好X世界")
	}
}

func TestGapBuffer_Unicode_Accented(t *testing.T) {
	g := NewGapBuffer("café")
	if got := g.Length(); got != 4 {
		t.Errorf("Length() = %d, want 4", got)
	}
	g.Backspace()
	if got := g.Content(); got != "caf" {
		t.Errorf("Content() = %q, want %q", got, "caf")
	}
}

func TestGapBuffer_Unicode_Mixed(t *testing.T) {
	text := "Hello 世界 🌍 café"
	g := NewGapBuffer(text)
	if got := g.Content(); got != text {
		t.Errorf("Content() = %q, want %q", got, text)
	}
}

// --- Gap Growth Tests ---

func TestGapBuffer_GapGrowth_ExhaustAndContinue(t *testing.T) {
	g := NewGapBuffer("")
	// Insert more than defaultGapSize (64) characters
	for i := 0; i < 200; i++ {
		g.Insert('x')
	}
	if got := g.Length(); got != 200 {
		t.Errorf("Length() = %d, want 200", got)
	}
	if got := g.Content(); got != strings.Repeat("x", 200) {
		t.Errorf("Content() length = %d, want 200", len([]rune(got)))
	}
}

func TestGapBuffer_GapGrowth_ContentIntegrity(t *testing.T) {
	g := NewGapBuffer("start")
	// Insert enough to force multiple growths
	for i := 0; i < 300; i++ {
		g.Insert(rune('a' + (i % 26)))
	}
	content := g.Content()
	if !strings.HasPrefix(content, "start") {
		t.Errorf("Content() lost prefix, got %q...", content[:10])
	}
	if got := g.Length(); got != 305 {
		t.Errorf("Length() = %d, want 305", got)
	}
}

// --- Large Buffer / Performance Tests ---

func TestGapBuffer_LargeBuffer_10000Chars(t *testing.T) {
	g := NewGapBuffer("")
	for i := 0; i < 10000; i++ {
		g.Insert(rune('a' + (i % 26)))
	}
	if got := g.Length(); got != 10000 {
		t.Errorf("Length() = %d, want 10000", got)
	}
	if got := len([]rune(g.Content())); got != 10000 {
		t.Errorf("Content() rune length = %d, want 10000", got)
	}

	// Verify cursor movement works after many inserts
	g.MoveToStart()
	if got := g.CursorPos(); got != 0 {
		t.Errorf("CursorPos() after MoveToStart = %d, want 0", got)
	}
	g.MoveToEnd()
	if got := g.CursorPos(); got != 10000 {
		t.Errorf("CursorPos() after MoveToEnd = %d, want 10000", got)
	}
}

// --- Edge Case Tests ---

func TestGapBuffer_EmptyInit_Operations(t *testing.T) {
	g := NewGapBuffer("")
	if g.Backspace() {
		t.Error("Backspace() on empty buffer should return false")
	}
	if g.Delete() {
		t.Error("Delete() on empty buffer should return false")
	}
	if g.MoveLeft() {
		t.Error("MoveLeft() on empty buffer should return false")
	}
	if g.MoveRight() {
		t.Error("MoveRight() on empty buffer should return false")
	}
	if got := g.Content(); got != "" {
		t.Errorf("Content() = %q, want empty", got)
	}
}

func TestGapBuffer_SingleChar_Operations(t *testing.T) {
	g := NewGapBuffer("a")
	if got := g.Length(); got != 1 {
		t.Errorf("Length() = %d, want 1", got)
	}
	g.Backspace()
	if got := g.Content(); got != "" {
		t.Errorf("Content() after backspace = %q, want empty", got)
	}
	if got := g.Length(); got != 0 {
		t.Errorf("Length() after backspace = %d, want 0", got)
	}
}

func TestGapBuffer_RapidInsertDelete(t *testing.T) {
	g := NewGapBuffer("")
	for i := 0; i < 100; i++ {
		g.Insert('x')
		g.Backspace()
	}
	if got := g.Content(); got != "" {
		t.Errorf("Content() after rapid insert/delete = %q, want empty", got)
	}
	if got := g.Length(); got != 0 {
		t.Errorf("Length() = %d, want 0", got)
	}
}

// --- Multi-line MoveUp/Down with SetCursorLineCol ---

func TestGapBuffer_MoveUpDown_ColumnTracking(t *testing.T) {
	g := NewGapBuffer("abc\ndefgh\nij")
	g.SetCursorLineCol(1, 4) // on second line, col 4 (after 'defg')
	line, col := g.CursorLineCol()
	if line != 1 || col != 4 {
		t.Errorf("After set, CursorLineCol() = (%d, %d), want (1, 4)", line, col)
	}

	g.MoveUp()
	line, col = g.CursorLineCol()
	if line != 0 || col != 3 {
		t.Errorf("After MoveUp, CursorLineCol() = (%d, %d), want (0, 3) (clamped)", line, col)
	}

	g.MoveDown()
	line, col = g.CursorLineCol()
	if line != 1 || col != 3 {
		t.Errorf("After MoveDown, CursorLineCol() = (%d, %d), want (1, 3)", line, col)
	}
}

func TestGapBuffer_SetCursorLineCol_ColBeyondLineLength(t *testing.T) {
	g := NewGapBuffer("ab\ncde")
	g.SetCursorLineCol(0, 10)
	line, col := g.CursorLineCol()
	if line != 0 || col != 2 {
		t.Errorf("CursorLineCol() = (%d, %d), want (0, 2) (clamped)", line, col)
	}
}

func TestGapBuffer_MultiLine_CursorLineColAfterEdit(t *testing.T) {
	g := NewGapBuffer("ab\ncd")
	g.SetCursorLineCol(1, 1) // after 'c' on second line
	g.Insert('X')
	if got := g.Content(); got != "ab\ncXd" {
		t.Errorf("Content() = %q, want %q", got, "ab\ncXd")
	}
	line, col := g.CursorLineCol()
	if line != 1 || col != 2 {
		t.Errorf("CursorLineCol() = (%d, %d), want (1, 2)", line, col)
	}
}
