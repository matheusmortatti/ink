package main

// Viewport handles scrolling for the editor.
type Viewport struct {
	YOffset      int
	Width        int
	Height       int
	ScrollMargin int
}

// NewViewport creates a viewport with default scroll margin.
func NewViewport(width, height int) *Viewport {
	return &Viewport{
		Width:        width,
		Height:       height,
		ScrollMargin: 5,
	}
}

// EnsureCursorVisible adjusts YOffset so the cursor display line is visible
// with scroll margin padding.
func (vp *Viewport) EnsureCursorVisible(cursorDisplayLine, totalDisplayLines int) {
	if vp.Height <= 0 {
		return
	}

	margin := vp.ScrollMargin
	if margin > vp.Height/2 {
		margin = vp.Height / 2
	}

	// Cursor above visible area
	if cursorDisplayLine < vp.YOffset+margin {
		vp.YOffset = cursorDisplayLine - margin
	}

	// Cursor below visible area
	if cursorDisplayLine >= vp.YOffset+vp.Height-margin {
		vp.YOffset = cursorDisplayLine - vp.Height + margin + 1
	}

	// Clamp
	maxOffset := totalDisplayLines - vp.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.YOffset > maxOffset {
		vp.YOffset = maxOffset
	}
	if vp.YOffset < 0 {
		vp.YOffset = 0
	}
}

// VisibleRange returns the first and last visible display lines.
func (vp *Viewport) VisibleRange() (int, int) {
	return vp.YOffset, vp.YOffset + vp.Height - 1
}

// ScrollDown scrolls down by n lines.
func (vp *Viewport) ScrollDown(n, totalLines int) {
	vp.YOffset += n
	maxOffset := totalLines - vp.Height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if vp.YOffset > maxOffset {
		vp.YOffset = maxOffset
	}
}

// ScrollUp scrolls up by n lines.
func (vp *Viewport) ScrollUp(n int) {
	vp.YOffset -= n
	if vp.YOffset < 0 {
		vp.YOffset = 0
	}
}
