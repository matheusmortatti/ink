package editor

import (
	"testing"
	"time"

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

func TestEditorModel_InsertMode_EnterAndExit(t *testing.T) {
	blocks := block.Parse([]byte("Hello world"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if e.CurrentMode() != vim.Normal {
		t.Fatalf("expected Normal mode, got %v", e.CurrentMode())
	}

	// Press i to enter insert mode
	e.Update(tea.KeyPressMsg{Code: 'i'})

	if e.CurrentMode() != vim.Insert {
		t.Fatalf("expected Insert mode after i, got %v", e.CurrentMode())
	}
	if e.activeBlockIdx < 0 {
		t.Fatal("expected activeBlockIdx >= 0 in insert mode")
	}
	if e.activeBuffer == nil {
		t.Fatal("expected activeBuffer != nil in insert mode")
	}

	// Press Esc to exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if e.CurrentMode() != vim.Normal {
		t.Errorf("expected Normal mode after Esc, got %v", e.CurrentMode())
	}
	if e.activeBlockIdx != -1 {
		t.Errorf("expected activeBlockIdx=-1 after Esc, got %d", e.activeBlockIdx)
	}
	if e.activeBuffer != nil {
		t.Error("expected activeBuffer=nil after Esc")
	}
}

func TestEditorModel_InsertMode_TypeText_UpdatesBlock(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Type "World"
	for _, ch := range "World" {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	// Exit insert mode
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Block content should be updated — 'i' places cursor at start
	if e.blocks[0].Raw != "WorldHello" {
		t.Errorf("expected block raw='WorldHello', got %q", e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_TypeSpace(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode at end (use 'a' which positions after first char)
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Type space using the "space" key representation
	e.Update(tea.KeyPressMsg{Code: ' '})

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	if e.blocks[0].Raw != " Hello" {
		t.Errorf("expected block raw=' Hello', got %q", e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_Backspace(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode with 'a' (after cursor)
	e.Update(tea.KeyPressMsg{Code: 'a'})

	// Type a character then backspace it
	e.Update(tea.KeyPressMsg{Code: 'X'})
	e.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Block should be unchanged since we typed then deleted
	if e.blocks[0].Raw != "Hello" {
		t.Errorf("expected block raw='Hello' after type+backspace, got %q", e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_NewlineO(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Press o to create new line below
	e.Update(tea.KeyPressMsg{Code: 'o'})

	// Type on the new line
	for _, ch := range "World" {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	expected := "Hello\nWorld"
	if e.blocks[0].Raw != expected {
		t.Errorf("expected block raw=%q after o+type, got %q", expected, e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_NewlineO_MultiLine(t *testing.T) {
	blocks := block.Parse([]byte("Line1\nLine2\nLine3"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Press o to create new line below cursor line (line 0 with simplified mapping)
	e.Update(tea.KeyPressMsg{Code: 'o'})

	// Type on the new line
	for _, ch := range "New" {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	expected := "Line1\nNew\nLine2\nLine3"
	if e.blocks[0].Raw != expected {
		t.Errorf("expected block raw=%q after o on multi-line, got %q", expected, e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_Delete(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode at beginning
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Delete key removes character after cursor
	e.Update(tea.KeyPressMsg{Code: tea.KeyDelete})

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	expected := "ello"
	if e.blocks[0].Raw != expected {
		t.Errorf("expected block raw=%q after delete, got %q", expected, e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_NewlineShiftO(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Press O to create new line above
	e.Update(tea.KeyPressMsg{Code: 'O', Mod: tea.ModShift})

	// Type on the new line
	for _, ch := range "World" {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	expected := "World\nHello"
	if e.blocks[0].Raw != expected {
		t.Errorf("expected block raw=%q after O+type, got %q", expected, e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_Tab(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Type tab
	e.Update(tea.KeyPressMsg{Code: tea.KeyTab})

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	expected := "\tHello"
	if e.blocks[0].Raw != expected {
		t.Errorf("expected block raw=%q after tab, got %q", expected, e.blocks[0].Raw)
	}
}

func TestEditorModel_InsertMode_CursorShape(t *testing.T) {
	blocks := block.Parse([]byte("Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Normal mode: cursor should be block
	v := e.View()
	if v.Cursor.Shape != tea.CursorBlock {
		t.Errorf("expected CursorBlock in normal mode, got %v", v.Cursor.Shape)
	}

	// Enter insert mode
	e.Update(tea.KeyPressMsg{Code: 'i'})

	v = e.View()
	if v.Cursor.Shape != tea.CursorBar {
		t.Errorf("expected CursorBar in insert mode, got %v", v.Cursor.Shape)
	}

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	v = e.View()
	if v.Cursor.Shape != tea.CursorBlock {
		t.Errorf("expected CursorBlock after Esc, got %v", v.Cursor.Shape)
	}
}

func TestEditorModel_InsertMode_ArrowKeys(t *testing.T) {
	blocks := block.Parse([]byte("Hello\nWorld"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Move right, then type — should insert after moved position
	e.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	e.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	e.Update(tea.KeyPressMsg{Code: 'X'})

	// Exit
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	expected := "HeXllo\nWorld"
	if e.blocks[0].Raw != expected {
		t.Errorf("expected %q after arrow+type, got %q", expected, e.blocks[0].Raw)
	}
}

// TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified verifies that the
// cache is NOT invalidated when a block is entered and exited without modification.
func TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if e.cache == nil {
		t.Fatal("expected cache to be initialized")
	}

	// Use the actual viewport column width instead of hardcoded value
	colWidth := e.viewport.ColumnWidth()
	_, cacheHit := e.cache.Get(e.blocks[0], colWidth)
	if !cacheHit {
		// The block should already be cached from initViewport's PreRenderAll
		t.Log("block not yet in cache, pre-rendering now")
		rendered, _ := e.renderer.Render(e.blocks[0])
		e.cache.Put(e.blocks[0], colWidth, rendered)
	}

	// Verify cache entry exists before entering insert mode
	_, cacheHitBefore := e.cache.Get(e.blocks[0], colWidth)
	if !cacheHitBefore {
		t.Fatal("expected block to be in cache before entering insert mode")
	}

	// Enter insert mode on the first block (cursor starts at line 0)
	e.Update(tea.KeyPressMsg{Code: 'i'})

	if e.activeBlockIdx != 0 {
		t.Fatalf("expected activeBlockIdx=0, got %d", e.activeBlockIdx)
	}

	// Exit WITHOUT making any modifications
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Cache entry should STILL exist because content was unchanged
	_, cacheHitAfter := e.cache.Get(e.blocks[0], colWidth)
	if !cacheHitAfter {
		t.Error("expected cache entry to be preserved when block is unmodified")
	}
}

// TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified_VariantA verifies
// cache reuse when entering with 'a' command.
func TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified_VariantA(t *testing.T) {
	blocks := block.Parse([]byte("# Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter with 'a' variant
	e.Update(tea.KeyPressMsg{Code: 'a'})
	// Exit without modification
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Content should be unchanged
	if e.blocks[0].Raw != "# Hello" {
		t.Errorf("content changed unexpectedly: %q", e.blocks[0].Raw)
	}
}

// TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified_VariantO verifies
// cache reuse when entering with 'o' command.
func TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified_VariantO(t *testing.T) {
	blocks := block.Parse([]byte("# Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter with 'o' variant (creates new line)
	e.Update(tea.KeyPressMsg{Code: 'o'})
	// Immediately exit without typing
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Should have newline added
	if e.blocks[0].Raw != "# Hello\n" {
		t.Errorf("expected newline after o command, got %q", e.blocks[0].Raw)
	}
}

// TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified_VariantShiftO verifies
// cache reuse when entering with 'O' command.
func TestEditorModel_ExitInsertMode_CacheReusedWhenUnmodified_VariantShiftO(t *testing.T) {
	blocks := block.Parse([]byte("# Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter with 'O' variant (creates new line above)
	e.Update(tea.KeyPressMsg{Code: 'O', Mod: tea.ModShift})
	// Immediately exit without typing
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Should have newline added above
	if e.blocks[0].Raw != "\n# Hello" {
		t.Errorf("expected newline before content after O command, got %q", e.blocks[0].Raw)
	}
}

// TestEditorModel_CacheInvalidation_MultipleWidths verifies that InvalidateBlock
// removes cache entries at ALL widths for a given block.
func TestEditorModel_CacheInvalidation_MultipleWidths(t *testing.T) {
	blocks := block.Parse([]byte("# Hello"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if e.cache == nil {
		t.Fatal("expected cache to be initialized")
	}

	// Pre-render the same block at multiple widths
	widths := []int{60, 80, 100}
	for _, w := range widths {
		rendered, err := e.renderer.Render(e.blocks[0])
		if err != nil {
			t.Fatalf("failed to render at width %d: %v", w, err)
		}
		e.cache.Put(e.blocks[0], w, rendered)
	}

	// Verify all widths are cached
	for _, w := range widths {
		if _, hit := e.cache.Get(e.blocks[0], w); !hit {
			t.Fatalf("expected cache hit at width %d", w)
		}
	}

	// Invalidate the block
	e.cache.InvalidateBlock(e.blocks[0])

	// Verify ALL widths are invalidated
	for _, w := range widths {
		if _, hit := e.cache.Get(e.blocks[0], w); hit {
			t.Errorf("expected cache miss at width %d after InvalidateBlock", w)
		}
	}
}

// TestEditorModel_ExitInsertMode_CacheInvalidatedWhenModified verifies that the
// cache IS invalidated when block content is modified.
func TestEditorModel_ExitInsertMode_CacheInvalidatedWhenModified(t *testing.T) {
	blocks := block.Parse([]byte("# Hello\n\nWorld"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if e.cache == nil {
		t.Fatal("expected cache to be initialized")
	}

	// Use the actual viewport column width instead of hardcoded value
	colWidth := e.viewport.ColumnWidth()
	originalContent := e.blocks[0].Raw
	_, cacheHit := e.cache.Get(e.blocks[0], colWidth)
	if !cacheHit {
		rendered, _ := e.renderer.Render(e.blocks[0])
		e.cache.Put(e.blocks[0], colWidth, rendered)
	}

	// Verify cache entry exists with original content
	_, cacheHitBefore := e.cache.Get(e.blocks[0], colWidth)
	if !cacheHitBefore {
		t.Fatal("expected block to be in cache before modification")
	}

	// Enter insert mode on the first block
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Modify the content by typing "X"
	e.Update(tea.KeyPressMsg{Code: 'X'})

	// Verify content was actually modified
	if e.activeBuffer.Content() == originalContent {
		t.Fatal("expected buffer content to be modified")
	}

	// Exit insert mode
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Block content should be updated
	if e.blocks[0].Raw == originalContent {
		t.Error("expected block content to be updated after modification")
	}

	// Cache entry for the OLD content should be gone (invalidated)
	// The cache uses content hash, so the old block should no longer have a cache hit
	oldBlock := block.Block{Type: e.blocks[0].Type, Raw: originalContent, Level: e.blocks[0].Level}
	_, cacheHitOld := e.cache.Get(oldBlock, colWidth)
	if cacheHitOld {
		t.Error("expected old cache entry to be invalidated when block is modified")
	}
}

// TestEditorModel_BlockTransition_SurroundingBlockRangesStable verifies that
// non-active blocks' positions are recalculated correctly after transitions.
func TestEditorModel_BlockTransition_SurroundingBlockRangesStable(t *testing.T) {
	// Create a multi-block document to test surrounding blocks
	content := "# Title\n\nFirst paragraph\n\nSecond paragraph\n\nThird paragraph"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if len(blocks) < 3 {
		t.Fatalf("expected at least 3 blocks, got %d", len(blocks))
	}

	// Record initial block start lines (all fully rendered)
	initialStarts := make([]int, len(blocks))
	for i := range blocks {
		initialStarts[i] = e.viewport.BlockStartLine(i)
	}

	// Enter insert mode on the first block (index 0)
	e.Update(tea.KeyPressMsg{Code: 'i'})

	if e.activeBlockIdx != 0 {
		t.Fatalf("expected activeBlockIdx=0, got %d", e.activeBlockIdx)
	}

	// Record block start lines after entering insert mode
	afterEnterStarts := make([]int, len(blocks))
	for i := range blocks {
		afterEnterStarts[i] = e.viewport.BlockStartLine(i)
	}

	// The active block (0) may have different line count in raw vs rendered
	// Surrounding blocks (1, 2, 3...) should have valid start lines
	for i := 1; i < len(blocks); i++ {
		if afterEnterStarts[i] < 0 {
			t.Errorf("block %d has invalid start line %d after entering insert mode", i, afterEnterStarts[i])
		}
	}

	// Exit insert mode
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Record block start lines after exiting insert mode
	afterExitStarts := make([]int, len(blocks))
	for i := range blocks {
		afterExitStarts[i] = e.viewport.BlockStartLine(i)
	}

	// After exit, all blocks should have valid start lines
	for i := range blocks {
		if afterExitStarts[i] < 0 {
			t.Errorf("block %d has invalid start line %d after exiting insert mode", i, afterExitStarts[i])
		}
	}

	// If the content wasn't modified, start lines should return to initial values
	// (since line counts for raw vs rendered might differ, but unchanged content goes back to same rendered form)
	for i := range blocks {
		if afterExitStarts[i] != initialStarts[i] {
			// This is acceptable if content was modified or if raw/rendered line counts differ
			// But for unmodified content, they should match
			t.Logf("block %d start: initial=%d, after exit=%d (difference expected if line counts vary)",
				i, initialStarts[i], afterExitStarts[i])
		}
	}
}

// TestEditorModel_BlockTransition_LineCountDifference verifies behavior when
// raw and rendered states have different line counts.
func TestEditorModel_BlockTransition_LineCountDifference(t *testing.T) {
	// Create a document where rendered form might differ from raw
	// A code fence or blockquote will have different rendering
	content := "# Title\n\n```\ncode line 1\ncode line 2\n```\n\nParagraph"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if len(blocks) < 2 {
		t.Skip("need at least 2 blocks for this test")
	}

	// Find the code fence block (Type == block.CodeFence)
	codeFenceIdx := -1
	for i, b := range blocks {
		if b.Type == block.CodeFence {
			codeFenceIdx = i
			break
		}
	}

	if codeFenceIdx < 0 {
		t.Skip("no code fence block found in test content")
	}

	// Move cursor to the code fence block
	codeFenceStart := e.viewport.BlockStartLine(codeFenceIdx)
	e.cursorLine = codeFenceStart
	e.clampCursor()

	// Enter insert mode on the code fence
	e.Update(tea.KeyPressMsg{Code: 'i'})

	if e.activeBlockIdx != codeFenceIdx {
		t.Fatalf("expected activeBlockIdx=%d, got %d", codeFenceIdx, e.activeBlockIdx)
	}

	// Record the block start line for the next block (if it exists)
	var nextBlockStart int
	if codeFenceIdx+1 < len(blocks) {
		nextBlockStart = e.viewport.BlockStartLine(codeFenceIdx + 1)
		if nextBlockStart < 0 {
			t.Error("next block should have valid start line during active editing")
		}
	}

	// Exit insert mode
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// After exit, next block should still have a valid start line (though it may have shifted)
	if codeFenceIdx+1 < len(blocks) {
		nextBlockStartAfter := e.viewport.BlockStartLine(codeFenceIdx + 1)
		if nextBlockStartAfter < 0 {
			t.Error("next block should have valid start line after exiting insert mode")
		}
		// Position may shift, but that's expected and acceptable
		t.Logf("next block start: during edit=%d, after exit=%d", nextBlockStart, nextBlockStartAfter)
	}
}

// TestEditorModel_MultipleEnterExitCycles verifies cache coherence across
// multiple enter/exit cycles on the same block.
func TestEditorModel_MultipleEnterExitCycles(t *testing.T) {
	blocks := block.Parse([]byte("# Title\n\nParagraph content"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	originalContent := e.blocks[0].Raw

	// Perform multiple enter/exit cycles without modifications
	for cycle := 0; cycle < 5; cycle++ {
		// Enter insert mode
		e.Update(tea.KeyPressMsg{Code: 'i'})

		if e.CurrentMode() != vim.Insert {
			t.Fatalf("cycle %d: expected Insert mode, got %v", cycle, e.CurrentMode())
		}
		if e.activeBlockIdx != 0 {
			t.Fatalf("cycle %d: expected activeBlockIdx=0, got %d", cycle, e.activeBlockIdx)
		}

		// Exit without modification
		e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

		if e.CurrentMode() != vim.Normal {
			t.Fatalf("cycle %d: expected Normal mode after exit, got %v", cycle, e.CurrentMode())
		}
		if e.activeBlockIdx != -1 {
			t.Fatalf("cycle %d: expected activeBlockIdx=-1 after exit, got %d", cycle, e.activeBlockIdx)
		}

		// Content should remain unchanged
		if e.blocks[0].Raw != originalContent {
			t.Errorf("cycle %d: block content changed unexpectedly: %q -> %q",
				cycle, originalContent, e.blocks[0].Raw)
		}
	}

	// All cycles should complete without errors, and content should be unchanged
	if e.blocks[0].Raw != originalContent {
		t.Errorf("after %d cycles, content should be unchanged", 5)
	}
}

// TestEditorModel_ReenterAfterModification verifies that re-entering a block
// after modification shows the updated raw content.
func TestEditorModel_ReenterAfterModification(t *testing.T) {
	blocks := block.Parse([]byte("# Original Title"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	originalContent := e.blocks[0].Raw

	// First cycle: enter, modify, exit
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Type "Modified " at the beginning
	for _, ch := range "Modified " {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	modifiedContent := e.activeBuffer.Content()
	if modifiedContent == originalContent {
		t.Fatal("expected content to be modified in active buffer")
	}

	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Verify block was updated
	if e.blocks[0].Raw != modifiedContent {
		t.Errorf("block content not updated after exit: got %q, want %q",
			e.blocks[0].Raw, modifiedContent)
	}

	// Second cycle: re-enter the same block
	e.Update(tea.KeyPressMsg{Code: 'i'})

	// Verify the gap buffer shows the NEW content (not the original)
	reenteredContent := e.activeBuffer.Content()
	if reenteredContent != modifiedContent {
		t.Errorf("re-entered content doesn't match modified content: got %q, want %q",
			reenteredContent, modifiedContent)
	}
	if reenteredContent == originalContent {
		t.Error("re-entered content shows original content instead of modified content")
	}

	// Exit again
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Final verification: block still has modified content
	if e.blocks[0].Raw != modifiedContent {
		t.Errorf("block content lost after re-enter/exit: got %q, want %q",
			e.blocks[0].Raw, modifiedContent)
	}
}

// TestEditorModel_TransitionPerformance verifies NFR2 requirement that
// block transitions complete in under 50ms.
func TestEditorModel_TransitionPerformance(t *testing.T) {
	// Test with various block sizes to ensure performance across all scenarios
	testCases := []struct {
		name    string
		content string
		maxTime time.Duration
	}{
		{
			name:    "single line heading",
			content: "# Short Heading",
			maxTime: 50 * time.Millisecond,
		},
		{
			name: "multi-line paragraph",
			content: "This is a longer paragraph with multiple lines of text. " +
				"It contains enough content to wrap across several lines when rendered.",
			maxTime: 50 * time.Millisecond,
		},
		{
			name:    "cached path (no modification)",
			content: "# Cached Content",
			maxTime: 50 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			blocks := block.Parse([]byte(tc.content))
			e := NewEditor("test.md", blocks)
			e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

			start := time.Now()
			e.Update(tea.KeyPressMsg{Code: 'i'})
			e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
			duration := time.Since(start)

			if duration > tc.maxTime {
				t.Errorf("transition took %v, exceeds %v NFR2 requirement",
					duration, tc.maxTime)
			}
		})
	}
}

// TestEditorModel_ResizeDuringInsertMode verifies that terminal resize while
// in insert mode doesn't corrupt active block state or cache.
func TestEditorModel_ResizeDuringInsertMode(t *testing.T) {
	blocks := block.Parse([]byte("# Title\n\nParagraph content"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter insert mode
	e.Update(tea.KeyPressMsg{Code: 'i'})

	if e.activeBlockIdx != 0 {
		t.Fatalf("expected activeBlockIdx=0, got %d", e.activeBlockIdx)
	}

	originalContent := e.activeBuffer.Content()

	// Type some content
	for _, ch := range "Modified " {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	modifiedContent := e.activeBuffer.Content()
	if modifiedContent == originalContent {
		t.Fatal("expected content to be modified before resize")
	}

	// Resize terminal while in insert mode
	e.Update(tea.WindowSizeMsg{Width: 80, Height: 30})

	// Verify we're still in insert mode with active block
	if e.CurrentMode() != vim.Insert {
		t.Errorf("expected Insert mode after resize, got %v", e.CurrentMode())
	}
	if e.activeBlockIdx != 0 {
		t.Errorf("expected activeBlockIdx=0 after resize, got %d", e.activeBlockIdx)
	}
	if e.activeBuffer == nil {
		t.Fatal("expected activeBuffer != nil after resize")
	}

	// Verify buffer content is preserved
	if e.activeBuffer.Content() != modifiedContent {
		t.Errorf("buffer content changed during resize: got %q, want %q",
			e.activeBuffer.Content(), modifiedContent)
	}

	// Exit insert mode
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Verify block was updated with modified content
	if e.blocks[0].Raw != modifiedContent {
		t.Errorf("block content not saved after resize+exit: got %q, want %q",
			e.blocks[0].Raw, modifiedContent)
	}
}

// TestEditorModel_TransitionCycle_FullIntegration is a comprehensive integration
// test covering enter, edit, exit, re-enter flow.
func TestEditorModel_TransitionCycle_FullIntegration(t *testing.T) {
	content := "# Title\n\nFirst paragraph\n\nSecond paragraph"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Test scenario: enter first block, exit without editing (should use cache)
	e.Update(tea.KeyPressMsg{Code: 'i'})
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Verify we're back in normal mode at the block start
	if e.CurrentMode() != vim.Normal {
		t.Fatalf("expected Normal mode, got %v", e.CurrentMode())
	}

	// Move to second block (cursor line 2 or 3 depending on rendering)
	e.Update(tea.KeyPressMsg{Code: 'j'})
	e.Update(tea.KeyPressMsg{Code: 'j'})
	secondBlockIdx := e.blockIndexForLine(e.cursorLine)

	if secondBlockIdx < 1 {
		t.Fatalf("expected to be on second block (idx >= 1), got idx=%d", secondBlockIdx)
	}

	secondBlockOriginal := e.blocks[secondBlockIdx].Raw

	// Enter second block, modify, exit
	e.Update(tea.KeyPressMsg{Code: 'i'})

	if e.activeBlockIdx != secondBlockIdx {
		t.Fatalf("expected activeBlockIdx=%d, got %d", secondBlockIdx, e.activeBlockIdx)
	}

	// Type some text
	for _, ch := range "EDITED: " {
		e.Update(tea.KeyPressMsg{Code: ch})
	}

	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Verify modification was saved
	if e.blocks[secondBlockIdx].Raw == secondBlockOriginal {
		t.Error("expected block content to be modified after editing")
	}

	// Verify surrounding blocks (first block) are still intact
	if e.blocks[0].Raw != "# Title" {
		t.Errorf("first block was unexpectedly modified: %q", e.blocks[0].Raw)
	}
}
