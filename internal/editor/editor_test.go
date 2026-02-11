package editor

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matheusmortatti/ink/internal/block"
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

	// First WindowSizeMsg to initialize
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	// Second WindowSizeMsg to trigger resize
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

func TestEditorModel_KeyPressMsg_ScrollDown(t *testing.T) {
	// Create enough content to overflow a small viewport
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three\n\nParagraph four\n\nParagraph five\n\nParagraph six\n\nParagraph seven\n\nParagraph eight"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	// Initialize viewport with small height to ensure overflow
	msg := tea.WindowSizeMsg{Width: 120, Height: 3}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	if m.viewport.ContentHeight() <= m.viewport.ViewportHeight() {
		t.Skip("content does not overflow viewport, cannot test scroll")
	}

	offsetBefore := m.viewport.ScrollOffset()
	jMsg := tea.KeyPressMsg{Code: 'j'}
	updated2, _ := m.Update(jMsg)
	m2 := updated2.(*EditorModel)

	if m2.viewport.ScrollOffset() <= offsetBefore {
		t.Errorf("expected scroll offset to increase after j key, got %d (was %d)", m2.viewport.ScrollOffset(), offsetBefore)
	}
}

func TestEditorModel_KeyPressMsg_ScrollUp(t *testing.T) {
	content := "# Title\n\nParagraph one\n\nParagraph two\n\nParagraph three\n\nParagraph four\n\nParagraph five\n\nParagraph six\n\nParagraph seven\n\nParagraph eight"
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)

	// Initialize viewport with small height
	msg := tea.WindowSizeMsg{Width: 120, Height: 3}
	updated, _ := e.Update(msg)
	m := updated.(*EditorModel)

	// Scroll down first
	jMsg := tea.KeyPressMsg{Code: 'j'}
	updated2, _ := m.Update(jMsg)
	m2 := updated2.(*EditorModel)
	offsetAfterDown := m2.viewport.ScrollOffset()

	// Scroll back up
	kMsg := tea.KeyPressMsg{Code: 'k'}
	updated3, _ := m2.Update(kMsg)
	m3 := updated3.(*EditorModel)

	if m3.viewport.ScrollOffset() >= offsetAfterDown {
		t.Errorf("expected scroll offset to decrease after k key, got %d (was %d)", m3.viewport.ScrollOffset(), offsetAfterDown)
	}
}

func TestEditorModel_View_NotReady(t *testing.T) {
	e := NewEditor("test.md", nil)
	v := e.View()

	// Should return a view with content (loading placeholder), no panic
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
	// Generate 10,000+ words (700 paragraphs × ~15 words each)
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
	// Should not crash, AltScreen should be set
	if !v.AltScreen {
		t.Error("expected AltScreen = true even with empty blocks")
	}
}
