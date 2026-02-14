package ui

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
)

var ansiEscRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// blockRange tracks the line range of a block within the composed output.
type blockRange struct {
	start int // first line index (inclusive)
	end   int // last line index (exclusive)
}

// Viewport displays a scrollable, centered document composed of rendered blocks.
type Viewport struct {
	layout       Layout
	lines        []string // pre-composed content lines with centering
	scrollOffset int
	totalLines   int
	blocks       []block.Block
	renderer     *render.Renderer
	cache        *render.RenderCache
	blockRanges  []blockRange // line ranges for each block
	activeBlock      int                    // index of active editing block, -1 when none
	activeRawContent string                 // raw content of the active block
	dimFunc          func(string) string    // function for dimming syntax characters
}

// NewViewport creates a viewport with the given terminal dimensions.
func NewViewport(width, height int) *Viewport {
	return &Viewport{
		layout:      NewLayout(width, height),
		activeBlock: -1,
	}
}

// SetDimFunc sets the function used for dimming syntax characters in the active block.
func (v *Viewport) SetDimFunc(dimFunc func(string) string) {
	v.dimFunc = dimFunc
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

// Line returns a specific content line by index. Returns empty string for out-of-range.
func (v *Viewport) Line(n int) string {
	if n < 0 || n >= v.totalLines {
		return ""
	}
	return v.lines[n]
}

// SetScrollOffset sets the scroll position directly (clamped to valid bounds).
func (v *Viewport) SetScrollOffset(offset int) {
	v.scrollOffset = offset
	v.clampScroll()
}

// composeBlocks renders each block, pads and centers lines, and joins with blank separators.
func (v *Viewport) composeBlocks() error {
	if len(v.blocks) == 0 {
		v.lines = nil
		v.totalLines = 0
		v.blockRanges = nil
		return nil
	}

	margin := strings.Repeat(" ", v.layout.LeftMargin)
	var allLines []string
	v.blockRanges = make([]blockRange, len(v.blocks))

	for i, b := range v.blocks {
		start := len(allLines)

		// Use raw content if this is the active editing block
		if i == v.activeBlock {
			displayContent := v.activeRawContent
			if v.dimFunc != nil {
				displayContent = render.DimSyntax(displayContent, b.Type, v.dimFunc)
			}
			rawLines := strings.Split(displayContent, "\n")
			for _, line := range rawLines {
				wrapped := wrapLine(line, v.layout.ColumnWidth)
				for _, wl := range wrapped {
					allLines = append(allLines, margin+wl)
				}
			}
		} else {
			rendered, err := v.renderer.RenderCached(b, v.cache)
			if err != nil {
				return err
			}
			blockLines := strings.Split(rendered, "\n")
			for _, line := range blockLines {
				allLines = append(allLines, margin+line)
			}
		}

		v.blockRanges[i] = blockRange{start: start, end: len(allLines)}

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

// SetActiveBlock replaces a block's rendered output with raw markdown text for editing.
func (v *Viewport) SetActiveBlock(blockIdx int, rawContent string) error {
	v.activeBlock = blockIdx
	v.activeRawContent = rawContent
	return v.composeBlocks()
}

// ClearActiveBlock restores full rendered display after editing.
func (v *Viewport) ClearActiveBlock() error {
	v.activeBlock = -1
	v.activeRawContent = ""
	return v.composeBlocks()
}

// UpdateActiveBlockContent live-updates the raw block content as the user types.
// Optimized to only rebuild the active block's lines instead of recomposing all blocks.
func (v *Viewport) UpdateActiveBlockContent(rawContent string) {
	if v.activeBlock < 0 || v.activeBlock >= len(v.blockRanges) {
		return
	}
	v.activeRawContent = rawContent

	displayContent := rawContent
	if v.dimFunc != nil && v.activeBlock >= 0 && v.activeBlock < len(v.blocks) {
		displayContent = render.DimSyntax(displayContent, v.blocks[v.activeBlock].Type, v.dimFunc)
	}

	margin := strings.Repeat(" ", v.layout.LeftMargin)
	newRawLines := strings.Split(displayContent, "\n")

	var newBlockLines []string
	for _, line := range newRawLines {
		wrapped := wrapLine(line, v.layout.ColumnWidth)
		for _, wl := range wrapped {
			newBlockLines = append(newBlockLines, margin+wl)
		}
	}

	oldRange := v.blockRanges[v.activeBlock]
	oldLen := oldRange.end - oldRange.start
	newLen := len(newBlockLines)
	delta := newLen - oldLen

	if delta == 0 {
		copy(v.lines[oldRange.start:oldRange.end], newBlockLines)
	} else {
		newLines := make([]string, 0, v.totalLines+delta)
		newLines = append(newLines, v.lines[:oldRange.start]...)
		newLines = append(newLines, newBlockLines...)
		newLines = append(newLines, v.lines[oldRange.end:]...)
		v.lines = newLines
		v.totalLines = len(newLines)

		v.blockRanges[v.activeBlock] = blockRange{start: oldRange.start, end: oldRange.start + newLen}
		for i := v.activeBlock + 1; i < len(v.blockRanges); i++ {
			v.blockRanges[i].start += delta
			v.blockRanges[i].end += delta
		}
	}

	v.clampScroll()
}

// BlockStartLine returns the first line index for the given block, or -1 if invalid.
func (v *Viewport) BlockStartLine(blockIdx int) int {
	if blockIdx < 0 || blockIdx >= len(v.blockRanges) {
		return -1
	}
	return v.blockRanges[blockIdx].start
}

// BufferToScreenPos converts a gap buffer (line, col) position to an absolute
// screen (row, col) position within the active block, accounting for line wrapping.
func (v *Viewport) BufferToScreenPos(bufLine, bufCol int) (screenRow, screenCol int) {
	if v.activeBlock < 0 || v.activeBlock >= len(v.blockRanges) {
		return 0, 0
	}

	blockStart := v.blockRanges[v.activeBlock].start
	rawLines := strings.Split(v.activeRawContent, "\n")

	screenRow = blockStart
	for l := 0; l < bufLine && l < len(rawLines); l++ {
		lineRuneLen := len([]rune(rawLines[l]))
		if v.layout.ColumnWidth > 0 && lineRuneLen > v.layout.ColumnWidth {
			screenRow += (lineRuneLen + v.layout.ColumnWidth - 1) / v.layout.ColumnWidth
		} else {
			screenRow++
		}
	}

	if v.layout.ColumnWidth > 0 && bufCol >= v.layout.ColumnWidth {
		screenRow += bufCol / v.layout.ColumnWidth
		screenCol = v.layout.LeftMargin + (bufCol % v.layout.ColumnWidth)
	} else {
		screenCol = v.layout.LeftMargin + bufCol
	}

	return screenRow, screenCol
}

// wrapLine splits a line into multiple lines at width visible characters.
// ANSI escape sequences are not counted as visible characters but are preserved
// in the output. Returns the original line in a slice if it fits within width.
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}

	// Fast path: if no ANSI codes present, use simple rune-based wrapping
	if !strings.Contains(line, "\x1b[") {
		runes := []rune(line)
		if len(runes) <= width {
			return []string{line}
		}
		var result []string
		for len(runes) > width {
			result = append(result, string(runes[:width]))
			runes = runes[width:]
		}
		result = append(result, string(runes))
		return result
	}

	// ANSI-aware wrapping: count only visible characters for width
	return wrapLineANSI(line, width)
}

// wrapLineANSI wraps a line containing ANSI escape sequences at width visible characters.
// Tracks active ANSI styling state and carries it across wrap boundaries so that
// styled regions that span a wrap point continue on the next line.
func wrapLineANSI(line string, width int) []string {
	matches := ansiEscRe.FindAllStringIndex(line, -1)
	ansiSet := make(map[int]int) // start byte -> end byte for ANSI sequences
	for _, m := range matches {
		ansiSet[m[0]] = m[1]
	}

	var result []string
	var current strings.Builder
	var activeANSI string // tracks current ANSI styling state
	visibleCount := 0
	i := 0
	bytes := []byte(line)

	for i < len(bytes) {
		// Check if this position starts an ANSI escape sequence
		if end, ok := ansiSet[i]; ok {
			seq := string(bytes[i:end])
			current.WriteString(seq)
			// Track ANSI state: reset clears, anything else sets
			if seq == "\x1b[0m" || seq == "\x1b[m" {
				activeANSI = ""
			} else {
				activeANSI = seq
			}
			i = end
			continue
		}

		// Regular visible character (may be multi-byte UTF-8)
		if visibleCount >= width {
			// Close styling on current line if active
			if activeANSI != "" {
				current.WriteString("\x1b[0m")
			}
			result = append(result, current.String())
			current.Reset()
			// Restore styling on new line
			if activeANSI != "" {
				current.WriteString(activeANSI)
			}
			visibleCount = 0
		}

		// Read one UTF-8 character
		_, size := utf8.DecodeRune(bytes[i:])
		current.Write(bytes[i : i+size])
		visibleCount++
		i += size
	}

	result = append(result, current.String())
	return result
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
