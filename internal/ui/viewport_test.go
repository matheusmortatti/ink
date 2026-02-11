package ui

import (
	"strings"
	"testing"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
)

func TestNewViewport(t *testing.T) {
	vp := NewViewport(120, 40)
	if vp == nil {
		t.Fatal("NewViewport returned nil")
	}
	if vp.layout.TerminalWidth != 120 {
		t.Errorf("TerminalWidth = %d, want 120", vp.layout.TerminalWidth)
	}
	if vp.layout.TerminalHeight != 40 {
		t.Errorf("TerminalHeight = %d, want 40", vp.layout.TerminalHeight)
	}
	if vp.layout.ColumnWidth != 80 {
		t.Errorf("ColumnWidth = %d, want 80", vp.layout.ColumnWidth)
	}
	if vp.layout.LeftMargin != 20 {
		t.Errorf("LeftMargin = %d, want 20", vp.layout.LeftMargin)
	}
}

func TestViewport_SetContent_BlockCompositing(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	blocks := []block.Block{
		{Type: block.Paragraph, Raw: "Hello world"},
		{Type: block.Paragraph, Raw: "Second paragraph"},
	}

	vp := NewViewport(120, 40)
	if err := vp.SetContent(blocks, r, cache); err != nil {
		t.Fatalf("SetContent: %v", err)
	}

	view := vp.View()
	if view == "" {
		t.Fatal("View returned empty string")
	}

	// Check content height is positive
	if vp.ContentHeight() == 0 {
		t.Error("ContentHeight is 0 after SetContent")
	}
}

func TestViewport_SetContent_BlocksSeparatedByBlankLine(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	blocks := []block.Block{
		{Type: block.Paragraph, Raw: "Block one"},
		{Type: block.Paragraph, Raw: "Block two"},
		{Type: block.Paragraph, Raw: "Block three"},
	}

	vp := NewViewport(120, 100) // tall enough to show all content
	if err := vp.SetContent(blocks, r, cache); err != nil {
		t.Fatalf("SetContent: %v", err)
	}

	// The composed content should have blank line separators between blocks
	view := vp.View()
	lines := strings.Split(view, "\n")

	// Find blank lines (may have margin spaces)
	blankCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			blankCount++
		}
	}

	// With 3 blocks, we expect at least 2 blank separators
	if blankCount < 2 {
		t.Errorf("Expected at least 2 blank line separators, got %d", blankCount)
	}
}

func TestViewport_ContentCentering(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	blocks := []block.Block{
		{Type: block.Paragraph, Raw: "Centered text"},
	}

	vp := NewViewport(120, 40) // margin should be 20
	if err := vp.SetContent(blocks, r, cache); err != nil {
		t.Fatalf("SetContent: %v", err)
	}

	view := vp.View()
	lines := strings.Split(view, "\n")

	// Non-blank lines should be indented by the left margin (20 spaces)
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, strings.Repeat(" ", 20)) {
			t.Errorf("Line not centered with 20-space margin: %q", line)
		}
	}
}

func TestViewport_NoCentering_SmallTerminal(t *testing.T) {
	// Verify that the layout margin is 0 for small terminals.
	// Glamour itself may add formatting indentation, but our centering margin should be 0.
	vp := NewViewport(30, 20) // margin should be 0
	if vp.layout.LeftMargin != 0 {
		t.Errorf("LeftMargin for 30-wide terminal = %d, want 0", vp.layout.LeftMargin)
	}
	if vp.layout.ColumnWidth != 30 {
		t.Errorf("ColumnWidth for 30-wide terminal = %d, want 30", vp.layout.ColumnWidth)
	}
}

func TestViewport_ViewportWindowing(t *testing.T) {
	vp := NewViewport(120, 5) // only 5 lines visible
	// Manually set lines to test windowing
	vp.lines = make([]string, 20)
	for i := range vp.lines {
		vp.lines[i] = strings.Repeat(" ", 20) + "line " + string(rune('A'+i))
	}
	vp.totalLines = len(vp.lines)

	view := vp.View()
	viewLines := strings.Split(view, "\n")

	if len(viewLines) != 5 {
		t.Errorf("View returned %d lines, want 5", len(viewLines))
	}
}

func TestViewport_ScrollDown_Basic(t *testing.T) {
	vp := NewViewport(120, 5)
	vp.lines = make([]string, 20)
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 20

	vp.ScrollDown(3)
	if vp.ScrollOffset() != 3 {
		t.Errorf("ScrollOffset after ScrollDown(3) = %d, want 3", vp.ScrollOffset())
	}
}

func TestViewport_ScrollDown_ClampsAtBottom(t *testing.T) {
	vp := NewViewport(120, 5)
	vp.lines = make([]string, 10)
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 10

	vp.ScrollDown(100) // try to scroll way past
	// Max offset: totalLines - viewportHeight = 10 - 5 = 5
	if vp.ScrollOffset() != 5 {
		t.Errorf("ScrollOffset after large ScrollDown = %d, want 5", vp.ScrollOffset())
	}
}

func TestViewport_ScrollUp_Basic(t *testing.T) {
	vp := NewViewport(120, 5)
	vp.lines = make([]string, 20)
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 20
	vp.scrollOffset = 10

	vp.ScrollUp(3)
	if vp.ScrollOffset() != 7 {
		t.Errorf("ScrollOffset after ScrollUp(3) from 10 = %d, want 7", vp.ScrollOffset())
	}
}

func TestViewport_ScrollUp_ClampsAtTop(t *testing.T) {
	vp := NewViewport(120, 5)
	vp.lines = make([]string, 20)
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 20
	vp.scrollOffset = 2

	vp.ScrollUp(10)
	if vp.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset after large ScrollUp from 2 = %d, want 0", vp.ScrollOffset())
	}
}

func TestViewport_ScrollToTop(t *testing.T) {
	vp := NewViewport(120, 5)
	vp.lines = make([]string, 20)
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 20
	vp.scrollOffset = 10

	vp.ScrollToTop()
	if vp.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset after ScrollToTop = %d, want 0", vp.ScrollOffset())
	}
}

func TestViewport_ScrollToBottom(t *testing.T) {
	vp := NewViewport(120, 5)
	vp.lines = make([]string, 20)
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 20

	vp.ScrollToBottom()
	// Max offset: 20 - 5 = 15
	if vp.ScrollOffset() != 15 {
		t.Errorf("ScrollOffset after ScrollToBottom = %d, want 15", vp.ScrollOffset())
	}
}

func TestViewport_Resize(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	blocks := []block.Block{
		{Type: block.Paragraph, Raw: "Test resize"},
	}

	vp := NewViewport(120, 40)
	if err := vp.SetContent(blocks, r, cache); err != nil {
		t.Fatalf("SetContent: %v", err)
	}

	originalMargin := vp.layout.LeftMargin

	// Resize to a different width
	if err := vp.Resize(200, 50); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	if vp.layout.TerminalWidth != 200 {
		t.Errorf("After Resize, TerminalWidth = %d, want 200", vp.layout.TerminalWidth)
	}
	if vp.layout.TerminalHeight != 50 {
		t.Errorf("After Resize, TerminalHeight = %d, want 50", vp.layout.TerminalHeight)
	}

	// Margin should change: (200-80)/2 = 60
	if vp.layout.LeftMargin == originalMargin {
		t.Error("Margin did not change after resize")
	}
	if vp.layout.LeftMargin != 60 {
		t.Errorf("After Resize(200, 50), LeftMargin = %d, want 60", vp.layout.LeftMargin)
	}
}

func TestViewport_Resize_RecomposesContent(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	blocks := []block.Block{
		{Type: block.Paragraph, Raw: "Resize test content"},
	}

	vp := NewViewport(120, 40)
	if err := vp.SetContent(blocks, r, cache); err != nil {
		t.Fatalf("SetContent: %v", err)
	}

	viewBefore := vp.View()

	// Resize to smaller — different column width, different margin
	if err := vp.Resize(60, 20); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	viewAfter := vp.View()

	// Content should be different (different centering)
	if viewBefore == viewAfter {
		t.Error("View did not change after resize to different width")
	}
}

func TestViewport_EmptyDocument(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	vp := NewViewport(120, 40)
	if err := vp.SetContent(nil, r, cache); err != nil {
		t.Fatalf("SetContent with nil blocks: %v", err)
	}

	if vp.ContentHeight() != 0 {
		t.Errorf("ContentHeight for empty document = %d, want 0", vp.ContentHeight())
	}

	view := vp.View()
	if view != "" {
		t.Errorf("View for empty document = %q, want empty string", view)
	}
}

func TestViewport_EmptyBlocks(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	vp := NewViewport(120, 40)
	if err := vp.SetContent([]block.Block{}, r, cache); err != nil {
		t.Fatalf("SetContent with empty blocks: %v", err)
	}

	if vp.ContentHeight() != 0 {
		t.Errorf("ContentHeight for empty blocks = %d, want 0", vp.ContentHeight())
	}
}

func TestViewport_ViewportHeight(t *testing.T) {
	vp := NewViewport(120, 40)
	if vp.ViewportHeight() != 40 {
		t.Errorf("ViewportHeight = %d, want 40", vp.ViewportHeight())
	}
}

func TestViewport_ContentHeight_AfterSetContent(t *testing.T) {
	r, err := render.NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := render.NewRenderCache()

	blocks := []block.Block{
		{Type: block.Paragraph, Raw: "Line 1"},
		{Type: block.Paragraph, Raw: "Line 2"},
	}

	vp := NewViewport(120, 40)
	if err := vp.SetContent(blocks, r, cache); err != nil {
		t.Fatalf("SetContent: %v", err)
	}

	// Should have at least 2 lines (one per block) plus separator
	if vp.ContentHeight() < 3 {
		t.Errorf("ContentHeight = %d, want at least 3", vp.ContentHeight())
	}
}

func TestViewport_ScrollDown_ContentShorterThanViewport(t *testing.T) {
	vp := NewViewport(120, 40) // viewport is 40 lines
	vp.lines = make([]string, 5) // only 5 lines of content
	for i := range vp.lines {
		vp.lines[i] = "line"
	}
	vp.totalLines = 5

	vp.ScrollDown(10)
	// Content is shorter than viewport, can't scroll
	if vp.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset when content shorter than viewport = %d, want 0", vp.ScrollOffset())
	}
}

func TestViewport_ViewOnlyVisibleLines(t *testing.T) {
	vp := NewViewport(120, 3) // only 3 lines visible
	vp.lines = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	vp.totalLines = 10

	// At offset 0, should see A, B, C
	view := vp.View()
	if view != "A\nB\nC" {
		t.Errorf("View at offset 0 = %q, want %q", view, "A\nB\nC")
	}

	// Scroll down 3
	vp.ScrollDown(3)
	view = vp.View()
	if view != "D\nE\nF" {
		t.Errorf("View at offset 3 = %q, want %q", view, "D\nE\nF")
	}
}
