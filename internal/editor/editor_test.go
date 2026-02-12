package editor

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/vim"
)

func TestNewEditor_WithBlocks(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	if e.filePath != "test.md" {
		t.Errorf("filePath = %q, want %q", e.filePath, "test.md")
	}
	if len(e.blocks) != len(blocks) {
		t.Errorf("blocks length = %d, want %d", len(e.blocks), len(blocks))
	}
	if e.ready {
		t.Error("expected ready = false before WindowSizeMsg")
	}
	if e.viewport != nil {
		t.Error("expected viewport = nil before WindowSizeMsg")
	}
}

func TestNewEditor_WithoutBlocks(t *testing.T) {
	e := NewEditor("empty.md", nil)

	if e.filePath != "empty.md" {
		t.Errorf("filePath = %q, want %q", e.filePath, "empty.md")
	}
	if e.blocks != nil {
		t.Errorf("blocks = %v, want nil", e.blocks)
	}
	if e.ready {
		t.Error("expected ready = false")
	}
}

func TestEditorModel_WindowSizeMsg_InitializesViewport(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	if !m.ready {
		t.Fatal("expected editor to be ready after WindowSizeMsg")
	}
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
	if m.viewport == nil {
		t.Fatal("expected viewport to be initialized")
	}
	if m.renderer == nil {
		t.Fatal("expected renderer to be initialized")
	}
	if m.cache == nil {
		t.Fatal("expected cache to be initialized")
	}
}

func TestEditorModel_WindowSizeMsg_Resize(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	msg2 := tea.WindowSizeMsg{Width: 80, Height: 30}
	updated2, _ := m.Update(msg2)
	m2 := updated2.(*EditorModel)

	if m2.width != 80 {
		t.Errorf("width after resize = %d, want 80", m2.width)
	}
	if m2.height != 30 {
		t.Errorf("height after resize = %d, want 30", m2.height)
	}
}

func TestEditorModel_KeyPressMsg_CtrlC(t *testing.T) {
	e := NewEditor("test.md", nil)

	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	_, cmd := e.Update(msg)

	if cmd == nil {
		t.Fatal("expected quit command from ctrl+c")
	}
}

func TestEditorModel_KeyPressMsg_CtrlC_WhenReady(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	_, cmd := e.Update(msg)

	if cmd == nil {
		t.Fatal("expected quit command from ctrl+c when ready")
	}
}

func TestEditorModel_JKey_MovesCursorDown(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if e.CursorLine() != 0 {
		t.Fatalf("expected initial cursorLine=0, got %d", e.CursorLine())
	}

	e.Update(tea.KeyPressMsg{Code: 'j'})

	if e.CursorLine() != 1 {
		t.Errorf("expected cursorLine=1 after j, got %d", e.CursorLine())
	}
}

func TestEditorModel_KKey_MovesCursorUp(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move down first
	e.Update(tea.KeyPressMsg{Code: 'j'})
	e.Update(tea.KeyPressMsg{Code: 'j'})

	if e.CursorLine() != 2 {
		t.Fatalf("expected cursorLine=2, got %d", e.CursorLine())
	}

	e.Update(tea.KeyPressMsg{Code: 'k'})

	if e.CursorLine() != 1 {
		t.Errorf("expected cursorLine=1 after k, got %d", e.CursorLine())
	}
}

func TestEditorModel_JKey_ClampsAtBottom(t *testing.T) {
	blocks := block.Parse([]byte("# Title\n\nParagraph"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	maxLine := e.maxLine()
	// Press j more times than there are lines
	for i := 0; i < maxLine+10; i++ {
		e.Update(tea.KeyPressMsg{Code: 'j'})
	}

	if e.CursorLine() != maxLine {
		t.Errorf("expected cursorLine clamped to %d, got %d", maxLine, e.CursorLine())
	}
}

func TestEditorModel_KKey_ClampsAtTop(t *testing.T) {
	blocks := block.Parse([]byte("# Title\n\nParagraph"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Try to move up from line 0
	e.Update(tea.KeyPressMsg{Code: 'k'})

	if e.CursorLine() != 0 {
		t.Errorf("expected cursorLine=0 at top, got %d", e.CursorLine())
	}
}

func TestEditorModel_HKey_MovesCursorLeft(t *testing.T) {
	blocks := block.Parse([]byte("# Hello World"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move right first
	e.Update(tea.KeyPressMsg{Code: 'l'})
	e.Update(tea.KeyPressMsg{Code: 'l'})

	if e.CursorCol() < 1 {
		t.Fatalf("expected cursorCol > 0 after l, got %d", e.CursorCol())
	}

	e.Update(tea.KeyPressMsg{Code: 'h'})
	colAfterH := e.CursorCol()
	if colAfterH != 1 {
		t.Errorf("expected cursorCol=1 after h, got %d", colAfterH)
	}
}

func TestEditorModel_LKey_MovesCursorRight(t *testing.T) {
	blocks := block.Parse([]byte("# Hello World"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	e.Update(tea.KeyPressMsg{Code: 'l'})

	if e.CursorCol() != 1 {
		t.Errorf("expected cursorCol=1 after l, got %d", e.CursorCol())
	}
}

func TestEditorModel_HKey_ClampsAtZero(t *testing.T) {
	blocks := block.Parse([]byte("# Hello"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	e.Update(tea.KeyPressMsg{Code: 'h'})

	if e.CursorCol() != 0 {
		t.Errorf("expected cursorCol=0 at start, got %d", e.CursorCol())
	}
}

func TestEditorModel_LKey_ClampsAtLineEnd(t *testing.T) {
	blocks := block.Parse([]byte("# Hi"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Press l many times past end of line
	for i := 0; i < 100; i++ {
		e.Update(tea.KeyPressMsg{Code: 'l'})
	}

	lineLen := e.visibleLineLength(0)
	if lineLen > 0 && e.CursorCol() >= lineLen {
		t.Errorf("cursorCol=%d should be < lineLen=%d", e.CursorCol(), lineLen)
	}
}

func TestEditorModel_GKey_JumpsToBottom(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	e.Update(tea.KeyPressMsg{Code: 'G', Mod: tea.ModShift})

	if e.CursorLine() != e.maxLine() {
		t.Errorf("expected cursorLine=%d after G, got %d", e.maxLine(), e.CursorLine())
	}
}

func TestEditorModel_GGKey_JumpsToTop(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move down first
	for i := 0; i < 3; i++ {
		e.Update(tea.KeyPressMsg{Code: 'j'})
	}

	if e.CursorLine() == 0 {
		t.Fatal("expected cursor moved away from 0 before testing gg")
	}

	// Press gg
	e.Update(tea.KeyPressMsg{Code: 'g'})
	e.Update(tea.KeyPressMsg{Code: 'g'})

	if e.CursorLine() != 0 {
		t.Errorf("expected cursorLine=0 after gg, got %d", e.CursorLine())
	}
}

func TestEditorModel_CtrlD_ScrollsHalfPageDown(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three\n\nParagraph four\n\nParagraph five\n\nParagraph six\n\nParagraph seven\n\nParagraph eight"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 10})

	lineBefore := e.CursorLine()
	e.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	lineAfter := e.CursorLine()

	if lineAfter <= lineBefore {
		t.Errorf("expected cursorLine to increase after ctrl+d, got before=%d after=%d", lineBefore, lineAfter)
	}
}

func TestEditorModel_CtrlU_ScrollsHalfPageUp(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three\n\nParagraph four\n\nParagraph five\n\nParagraph six\n\nParagraph seven\n\nParagraph eight"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 10})

	// Move down first
	e.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	lineAfterDown := e.CursorLine()

	e.Update(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	lineAfterUp := e.CursorLine()

	if lineAfterUp >= lineAfterDown {
		t.Errorf("expected cursorLine to decrease after ctrl+u, got down=%d up=%d", lineAfterDown, lineAfterUp)
	}
}

func TestEditorModel_EnsureCursorVisible_ScrollsDown(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three\n\nParagraph four\n\nParagraph five\n\nParagraph six\n\nParagraph seven\n\nParagraph eight"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 3})

	if e.viewport.ContentHeight() <= e.viewport.ViewportHeight() {
		t.Skip("content does not overflow viewport")
	}

	// Move cursor past visible area
	for i := 0; i < 5; i++ {
		e.Update(tea.KeyPressMsg{Code: 'j'})
	}

	offset := e.viewport.ScrollOffset()
	if offset == 0 && e.CursorLine() >= 3 {
		t.Errorf("expected scroll offset > 0 when cursor at line %d with vpHeight=3", e.CursorLine())
	}
}

func TestEditorModel_EnsureCursorVisible_ScrollsUp(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three\n\nParagraph four\n\nParagraph five"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 3})

	// Move down then back up
	for i := 0; i < 5; i++ {
		e.Update(tea.KeyPressMsg{Code: 'j'})
	}
	for i := 0; i < 5; i++ {
		e.Update(tea.KeyPressMsg{Code: 'k'})
	}

	if e.CursorLine() != 0 {
		t.Errorf("expected cursorLine=0 after returning to top, got %d", e.CursorLine())
	}
	if e.viewport.ScrollOffset() != 0 {
		t.Errorf("expected scrollOffset=0 when cursor at top, got %d", e.viewport.ScrollOffset())
	}
}

func TestEditorModel_EnsureCursorVisible_NoChangeWhenInView(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	e.Update(tea.KeyPressMsg{Code: 'j'})

	if e.viewport.ScrollOffset() != 0 {
		t.Errorf("expected scrollOffset=0 when cursor in view, got %d", e.viewport.ScrollOffset())
	}
}

func TestEditorModel_StickyColumn_PreservedOnShorterLine(t *testing.T) {
	// First line is long, second line is shorter, third line is long again
	content := "This is a long line with many characters\n\nX\n\nAnother long line with many characters"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move right several characters on first line
	for i := 0; i < 10; i++ {
		e.Update(tea.KeyPressMsg{Code: 'l'})
	}
	colOnLongLine := e.CursorCol()
	if colOnLongLine == 0 {
		t.Fatal("expected cursor to move right on long line")
	}

	// Move down to shorter line — col should clamp to shorter line length
	e.Update(tea.KeyPressMsg{Code: 'j'})
	colOnShortLine := e.CursorCol()
	shortLineLen := e.visibleLineLength(e.CursorLine())
	if shortLineLen > 0 && colOnShortLine >= shortLineLen {
		t.Errorf("cursor col %d should be clamped to short line length %d", colOnShortLine, shortLineLen)
	}

	// Move down past short line(s) to another long line — should restore desired col
	e.Update(tea.KeyPressMsg{Code: 'j'})
	colRestored := e.CursorCol()
	restoredLineLen := e.visibleLineLength(e.CursorLine())

	// If the restored line is long enough, desiredCol should be restored
	if restoredLineLen > colOnLongLine && colRestored != colOnLongLine {
		t.Errorf("expected desiredCol restored to %d on long line, got %d", colOnLongLine, colRestored)
	}
}

func TestEditorModel_HLKey_ResetsDesiredCol(t *testing.T) {
	content := "# Hello World"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move right
	e.Update(tea.KeyPressMsg{Code: 'l'})
	e.Update(tea.KeyPressMsg{Code: 'l'})
	e.Update(tea.KeyPressMsg{Code: 'l'})

	if e.desiredCol != e.CursorCol() {
		t.Errorf("desiredCol=%d should equal cursorCol=%d after h/l", e.desiredCol, e.CursorCol())
	}
}

func TestEditorModel_WKey_MovesForwardByWord(t *testing.T) {
	content := "Hello World Foo"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	colBefore := e.CursorCol()
	e.Update(tea.KeyPressMsg{Code: 'w'})
	colAfter := e.CursorCol()

	if colAfter <= colBefore {
		t.Errorf("expected cursorCol to advance after w, got before=%d after=%d", colBefore, colAfter)
	}
}

func TestEditorModel_BKey_MovesBackwardByWord(t *testing.T) {
	content := "Hello World Foo"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move forward first
	e.Update(tea.KeyPressMsg{Code: 'w'})
	colAfterW := e.CursorCol()

	e.Update(tea.KeyPressMsg{Code: 'b'})
	colAfterB := e.CursorCol()

	if colAfterB >= colAfterW {
		t.Errorf("expected cursorCol to decrease after b, got w=%d b=%d", colAfterW, colAfterB)
	}
}

func TestEditorModel_View_CursorPosition(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	v := e.View()
	if v.Cursor == nil {
		t.Fatal("expected cursor to be set in View")
	}

	// Cursor should be at screen position accounting for margin
	expectedX := e.leftMargin()
	if v.Cursor.X != expectedX {
		t.Errorf("cursor X = %d, want %d (leftMargin)", v.Cursor.X, expectedX)
	}
	if v.Cursor.Y != 0 {
		t.Errorf("cursor Y = %d, want 0", v.Cursor.Y)
	}
}

func TestEditorModel_View_CursorAfterMove(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	e.Update(tea.KeyPressMsg{Code: 'j'})

	v := e.View()
	if v.Cursor == nil {
		t.Fatal("expected cursor to be set")
	}
	if v.Cursor.Y != 1 {
		t.Errorf("cursor Y = %d, want 1 after j", v.Cursor.Y)
	}
}

func TestEditorModel_View_NotReady(t *testing.T) {
	e := NewEditor("test.md", nil)
	v := e.View()

	if v.Content == nil {
		t.Error("expected non-nil content in placeholder view")
	}
}

func TestEditorModel_View_ReadyWithBlocks(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	v := m.View()
	if !v.AltScreen {
		t.Error("expected AltScreen = true")
	}
	if v.Content == nil {
		t.Error("expected non-nil content with blocks")
	}
}

func TestEditorModel_LargeDocument(t *testing.T) {
	var content []byte
	content = append(content, []byte("# Large Document\n\n")...)
	for i := 0; i < 700; i++ {
		content = append(content, []byte("This is a paragraph with enough words to make it substantial and meaningful for performance validation testing.\n\n")...)
	}
	blocks := block.Parse(content)
	e := NewEditor("large.md", blocks)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	if !m.ready {
		t.Fatal("expected editor to be ready with large document")
	}
	if m.viewport == nil {
		t.Fatal("expected viewport initialized for large document")
	}
	if m.viewport.ContentHeight() == 0 {
		t.Error("expected non-zero content height for large document")
	}
}

func TestEditorModel_View_ReadyEmptyBlocks(t *testing.T) {
	e := NewEditor("empty.md", nil)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	v := m.View()
	if !v.AltScreen {
		t.Error("expected AltScreen = true even with empty blocks")
	}
}

func TestEditorModel_CurrentMode(t *testing.T) {
	e := NewEditor("test.md", nil)
	if e.CurrentMode() != vim.Normal {
		t.Errorf("expected Normal mode, got %v", e.CurrentMode())
	}
}

func TestEditorModel_WKey_CrossesLineBoundary(t *testing.T) {
	// Two short lines — w at end of first line should cross to second
	content := "Hi\n\nWorld"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	lineBefore := e.CursorLine()

	// Keep pressing w until we cross to a new line
	for i := 0; i < 10; i++ {
		e.Update(tea.KeyPressMsg{Code: 'w'})
		if e.CursorLine() > lineBefore {
			return // success — crossed line boundary
		}
	}
	t.Error("expected w to cross line boundary, but cursor stayed on line 0")
}

func TestEditorModel_BKey_CrossesLineBoundary(t *testing.T) {
	content := "Hello\n\nWorld"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Move to a later line first
	e.Update(tea.KeyPressMsg{Code: 'G', Mod: tea.ModShift})
	lineAtBottom := e.CursorLine()
	if lineAtBottom == 0 {
		t.Skip("document too short to test cross-line b")
	}

	// Keep pressing b until we cross back to an earlier line
	for i := 0; i < 10; i++ {
		e.Update(tea.KeyPressMsg{Code: 'b'})
		if e.CursorLine() < lineAtBottom {
			return // success — crossed line boundary backward
		}
	}
	t.Error("expected b to cross line boundary backward")
}

func TestEditorModel_Resize_ClampsCursorPosition(t *testing.T) {
	content := "This is a line with enough characters to test column clamping after resize"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	// Initialize with wide terminal
	e.Update(tea.WindowSizeMsg{Width: 200, Height: 40})

	// Move cursor right to a high column
	for i := 0; i < 30; i++ {
		e.Update(tea.KeyPressMsg{Code: 'l'})
	}
	colBeforeResize := e.CursorCol()
	if colBeforeResize == 0 {
		t.Fatal("expected cursor to have moved right")
	}

	// Resize to narrow terminal — cursor col should be clamped
	e.Update(tea.WindowSizeMsg{Width: 40, Height: 40})

	lineLen := e.visibleLineLength(e.CursorLine())
	if lineLen > 0 && e.CursorCol() >= lineLen {
		t.Errorf("after resize: cursorCol=%d should be < lineLen=%d", e.CursorCol(), lineLen)
	}
}

func TestEditorModel_EmptyDocument_NoNavigationCrash(t *testing.T) {
	e := NewEditor("empty.md", nil)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// All these should not panic
	e.Update(tea.KeyPressMsg{Code: 'j'})
	e.Update(tea.KeyPressMsg{Code: 'k'})
	e.Update(tea.KeyPressMsg{Code: 'h'})
	e.Update(tea.KeyPressMsg{Code: 'l'})
	e.Update(tea.KeyPressMsg{Code: 'w'})
	e.Update(tea.KeyPressMsg{Code: 'b'})
	e.Update(tea.KeyPressMsg{Code: 'G', Mod: tea.ModShift})
	e.Update(tea.KeyPressMsg{Code: 'g'})
	e.Update(tea.KeyPressMsg{Code: 'g'})
	e.Update(tea.KeyPressMsg{Code: 'd', Mod: tea.ModCtrl})
	e.Update(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
}
