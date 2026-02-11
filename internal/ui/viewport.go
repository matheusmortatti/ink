package ui

import (
	"strings"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
)

// Viewport displays a scrollable, centered document composed of rendered blocks.
type Viewport struct {
	layout       Layout
	lines        []string // pre-composed content lines with centering
	scrollOffset int
	totalLines   int
	blocks       []block.Block
	renderer     *render.Renderer
	cache        *render.RenderCache
}

// NewViewport creates a viewport with the given terminal dimensions.
func NewViewport(width, height int) *Viewport {
	return &Viewport{
		layout: NewLayout(width, height),
	}
}

// SetContent renders blocks and composes them into the viewport.
func (v *Viewport) SetContent(blocks []block.Block, renderer *render.Renderer, cache *render.RenderCache) error {
	v.blocks = blocks
	v.renderer = renderer
	v.cache = cache
	if err := renderer.SetWidth(v.layout.ColumnWidth); err != nil {
		return err
	}
	return v.composeBlocks()
}

// Resize recalculates layout for new terminal dimensions and recomposes content.
func (v *Viewport) Resize(width, height int) error {
	v.layout = NewLayout(width, height)
	if v.renderer != nil && v.cache != nil {
		if err := v.renderer.SetWidth(v.layout.ColumnWidth); err != nil {
			return err
		}
		v.cache.InvalidateAll()
		if err := v.composeBlocks(); err != nil {
			return err
		}
	}
	v.clampScroll()
	return nil
}

// View returns the visible portion of the content as a string.
func (v *Viewport) View() string {
	if v.totalLines == 0 {
		return ""
	}
	end := v.scrollOffset + v.layout.TerminalHeight
	if end > v.totalLines {
		end = v.totalLines
	}
	if v.scrollOffset >= v.totalLines {
		return ""
	}
	visible := v.lines[v.scrollOffset:end]
	return strings.Join(visible, "\n")
}

// ScrollDown moves the viewport down by the given number of lines.
func (v *Viewport) ScrollDown(lines int) {
	v.scrollOffset += lines
	v.clampScroll()
}

// ScrollUp moves the viewport up by the given number of lines.
func (v *Viewport) ScrollUp(lines int) {
	v.scrollOffset -= lines
	v.clampScroll()
}

// ScrollToTop jumps to the beginning of the document.
func (v *Viewport) ScrollToTop() {
	v.scrollOffset = 0
}

// ScrollToBottom jumps to the end of the document.
func (v *Viewport) ScrollToBottom() {
	v.scrollOffset = v.maxScroll()
}

// ContentHeight returns the total number of content lines.
func (v *Viewport) ContentHeight() int {
	return v.totalLines
}

// ScrollOffset returns the current scroll position.
func (v *Viewport) ScrollOffset() int {
	return v.scrollOffset
}

// ViewportHeight returns the visible area height.
func (v *Viewport) ViewportHeight() int {
	return v.layout.TerminalHeight
}

// composeBlocks renders each block, pads and centers lines, and joins with blank separators.
func (v *Viewport) composeBlocks() error {
	if len(v.blocks) == 0 {
		v.lines = nil
		v.totalLines = 0
		return nil
	}

	margin := strings.Repeat(" ", v.layout.LeftMargin)
	var allLines []string

	for i, b := range v.blocks {
		rendered, err := v.renderer.RenderCached(b, v.cache)
		if err != nil {
			return err
		}

		blockLines := strings.Split(rendered, "\n")
		for _, line := range blockLines {
			allLines = append(allLines, margin+line)
		}

		// Add blank line separator between blocks (not after the last)
		if i < len(v.blocks)-1 {
			allLines = append(allLines, margin)
		}
	}

	v.lines = allLines
	v.totalLines = len(allLines)
	v.clampScroll()
	return nil
}

// clampScroll ensures scroll offset stays within valid bounds.
func (v *Viewport) clampScroll() {
	max := v.maxScroll()
	if v.scrollOffset > max {
		v.scrollOffset = max
	}
	if v.scrollOffset < 0 {
		v.scrollOffset = 0
	}
}

// maxScroll returns the maximum valid scroll offset.
func (v *Viewport) maxScroll() int {
	max := v.totalLines - v.layout.TerminalHeight
	if max < 0 {
		return 0
	}
	return max
}
