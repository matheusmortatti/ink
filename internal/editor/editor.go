package editor

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
	"github.com/matheusmortatti/ink/internal/ui"
	"github.com/matheusmortatti/ink/internal/vim"
)

// EditorModel is the single Bubbletea model. Components are fields.
type EditorModel struct {
	blocks      []block.Block
	viewport    *ui.Viewport
	renderer    *render.Renderer
	cache       *render.RenderCache
	filePath    string
	ready       bool
	width       int
	height      int
	modeHandler vim.ModeHandler
	cursorLine  int
	cursorCol   int
	desiredCol  int

	// Insert mode state
	activeBlockIdx int              // index of block being edited, -1 when none
	activeBuffer   *block.GapBuffer // gap buffer for active editing block

	// Syntax dimming
	dimStyle lipgloss.Style
}

// NewEditor creates an EditorModel with the given file path and parsed blocks.
func NewEditor(filePath string, blocks []block.Block) *EditorModel {
	return &EditorModel{
		filePath:       filePath,
		blocks:         blocks,
		modeHandler:    vim.NewNormalHandler(),
		activeBlockIdx: -1,
	}
}

// CursorLine returns the current cursor line position.
func (e *EditorModel) CursorLine() int {
	return e.cursorLine
}

// CursorCol returns the current cursor column position.
func (e *EditorModel) CursorCol() int {
	return e.cursorCol
}

// CurrentMode returns the current vim mode.
func (e *EditorModel) CurrentMode() vim.Mode {
	return e.modeHandler.Mode()
}

func (e *EditorModel) Init() tea.Cmd {
	return nil
}

func (e *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
		if !e.ready {
			e.initViewport()
		} else {
			_ = e.viewport.Resize(e.width, e.height)
		}
		e.clampCursor()
		return e, nil

	case tea.KeyPressMsg:
		// ctrl+c always works, even before viewport is ready
		if msg.String() == "ctrl+c" {
			return e, tea.Quit
		}
		if !e.ready {
			return e, nil
		}
		action := e.modeHandler.HandleKey(msg.String())
		return e.applyAction(action)
	}

	return e, nil
}

func (e *EditorModel) View() tea.View {
	if !e.ready {
		return tea.NewView("loading...")
	}

	v := tea.NewView(e.viewport.View())
	v.AltScreen = true

	if e.activeBlockIdx >= 0 && e.activeBuffer != nil {
		// Insert mode: cursor is within the active block (accounts for line wrapping)
		bufLine, bufCol := e.activeBuffer.CursorLineCol()
		screenRow, screenCol := e.viewport.BufferToScreenPos(bufLine, bufCol)
		screenRow -= e.viewport.ScrollOffset()

		v.Cursor = tea.NewCursor(screenCol, screenRow)
		v.Cursor.Shape = tea.CursorBar
	} else {
		// Normal mode: cursor is at document level
		screenRow := e.cursorLine - e.viewport.ScrollOffset()
		screenCol := e.leftMargin() + e.cursorCol

		v.Cursor = tea.NewCursor(screenCol, screenRow)
		v.Cursor.Shape = tea.CursorBlock
	}

	return v
}

func (e *EditorModel) applyAction(action vim.Action) (tea.Model, tea.Cmd) {
	switch a := action.(type) {
	case vim.QuitAction:
		return e, tea.Quit

	case vim.ChangeModeAction:
		if a.Mode == vim.Insert && e.modeHandler.Mode() == vim.Normal {
			e.enterInsertMode(a.Variant)
		} else if a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Insert {
			e.exitInsertMode()
		}

	case vim.InsertCharAction:
		if e.activeBuffer != nil {
			e.activeBuffer.Insert(a.Char)
			e.updateActiveBlockDisplay()
		}

	case vim.BackspaceAction:
		if e.activeBuffer != nil {
			e.activeBuffer.Backspace()
			e.updateActiveBlockDisplay()
		}

	case vim.DeleteCharAction:
		if e.activeBuffer != nil {
			e.activeBuffer.Delete()
			e.updateActiveBlockDisplay()
		}

	case vim.InsertNewlineAction:
		if e.activeBuffer != nil {
			e.activeBuffer.Insert('\n')
			e.updateActiveBlockDisplay()
		}

	case vim.InsertTabAction:
		if e.activeBuffer != nil {
			e.activeBuffer.Insert('\t')
			e.updateActiveBlockDisplay()
		}

	case vim.MoveCursorAction:
		if e.activeBlockIdx >= 0 && e.activeBuffer != nil {
			// In insert mode: move within the gap buffer (wrap-aware for up/down)
			if a.Relative {
				if a.Col < 0 {
					e.activeBuffer.MoveLeft()
				} else if a.Col > 0 {
					e.activeBuffer.MoveRight()
				}
				if a.Line < 0 {
					e.moveInsertCursorUp()
				} else if a.Line > 0 {
					e.moveInsertCursorDown()
				}
			}
			e.ensureInsertCursorVisible()
		} else {
			// Normal mode cursor movement
			if a.Relative {
				if a.Line != 0 {
					e.cursorLine += a.Line
					e.clampCursorLine()
					e.cursorCol = e.desiredCol
					e.clampCursorCol()
				}
				if a.Col != 0 {
					e.cursorCol += a.Col
					e.clampCursorCol()
					e.desiredCol = e.cursorCol
				}
			} else {
				e.cursorLine = a.Line
				e.cursorCol = a.Col
				e.clampCursor()
				e.desiredCol = e.cursorCol
			}
			e.ensureCursorVisible()
		}

	case vim.WordMotionAction:
		e.applyWordMotion(a.Forward)
		e.desiredCol = e.cursorCol
		e.ensureCursorVisible()

	case vim.ScrollAction:
		halfPage := e.viewport.ViewportHeight() / 2
		if halfPage < 1 {
			halfPage = 1
		}
		lines := halfPage
		if a.Direction < 0 {
			lines = -halfPage
		}
		if a.MoveCursor {
			e.cursorLine += lines
			e.clampCursorLine()
			e.cursorCol = e.desiredCol
			e.clampCursorCol()
		}
		e.viewport.SetScrollOffset(e.viewport.ScrollOffset() + lines)
		e.ensureCursorVisible()

	case vim.DocumentPositionAction:
		switch a.Position {
		case "top":
			e.cursorLine = 0
			e.cursorCol = 0
			e.desiredCol = 0
		case "bottom":
			e.cursorLine = e.maxLine()
			e.cursorCol = 0
			e.desiredCol = 0
		}
		e.ensureCursorVisible()

	case vim.NoOpAction:
		// Nothing to do
	}

	return e, nil
}

// blockIndexForLine maps a document cursor line to the block index it falls within.
func (e *EditorModel) blockIndexForLine(docLine int) int {
	if e.viewport == nil || len(e.blocks) == 0 {
		return 0
	}
	for i := range e.blocks {
		start := e.viewport.BlockStartLine(i)
		if start < 0 {
			continue
		}
		// Check if next block's start is beyond docLine
		if i+1 < len(e.blocks) {
			nextStart := e.viewport.BlockStartLine(i + 1)
			if nextStart > 0 && docLine < nextStart {
				return i
			}
		} else {
			// Last block
			return i
		}
	}
	return len(e.blocks) - 1
}

// enterInsertMode activates insert mode on the block at the cursor position.
func (e *EditorModel) enterInsertMode(variant string) {
	blockIdx := e.blockIndexForLine(e.cursorLine)
	if blockIdx < 0 || blockIdx >= len(e.blocks) {
		return
	}

	e.activeBlockIdx = blockIdx
	e.activeBuffer = block.NewGapBuffer(e.blocks[blockIdx].Raw)

	// Compute block-relative rendered position and map to raw
	blockStart := e.viewport.BlockStartLine(blockIdx)
	renderedLine := e.cursorLine - blockStart
	renderedCol := e.cursorCol
	rawLine, rawCol := block.MapRenderedToRaw(e.blocks[blockIdx], renderedLine, renderedCol)

	// Position cursor in gap buffer based on variant
	switch variant {
	case "i":
		e.activeBuffer.SetCursorLineCol(rawLine, rawCol)
	case "a":
		e.activeBuffer.SetCursorLineCol(rawLine, rawCol)
		e.activeBuffer.MoveRight()
	case "o":
		e.activeBuffer.SetCursorLineCol(rawLine, 0)
		e.activeBuffer.MoveToLineEnd()
		e.activeBuffer.Insert('\n')
	case "O":
		e.activeBuffer.SetCursorLineCol(rawLine, 0)
		e.activeBuffer.MoveToLineStart()
		e.activeBuffer.Insert('\n')
		e.activeBuffer.SetCursorLineCol(rawLine, 0)
	}

	if err := e.viewport.SetActiveBlock(blockIdx, e.activeBuffer.Content()); err != nil {
		e.activeBlockIdx = -1
		e.activeBuffer = nil
		return
	}
	e.modeHandler = vim.NewInsertHandler()
	e.ensureInsertCursorVisible()
}

// exitInsertMode commits changes and returns to normal mode.
func (e *EditorModel) exitInsertMode() {
	if e.activeBlockIdx < 0 || e.activeBuffer == nil {
		return
	}

	// Capture raw cursor position before clearing
	rawLine, rawCol := e.activeBuffer.CursorLineCol()

	// Only invalidate cache if content actually changed
	newContent := e.activeBuffer.Content()
	if newContent != e.blocks[e.activeBlockIdx].Raw {
		// Invalidate the old content's cache entry before updating
		e.cache.InvalidateBlock(e.blocks[e.activeBlockIdx])
		e.blocks[e.activeBlockIdx].Raw = newContent
	}

	// Map raw cursor position to rendered position using updated block content
	renderedLine, renderedCol := block.MapRawToRendered(e.blocks[e.activeBlockIdx], rawLine, rawCol)

	// Recompose viewport to fully rendered state
	_ = e.viewport.ClearActiveBlock()

	// Place document cursor at the mapped rendered position
	blockStart := e.viewport.BlockStartLine(e.activeBlockIdx)
	if blockStart >= 0 {
		e.cursorLine = blockStart + renderedLine
		e.cursorCol = renderedCol
		e.desiredCol = renderedCol
	}

	e.activeBlockIdx = -1
	e.activeBuffer = nil
	e.modeHandler = vim.NewNormalHandler()
	e.clampCursor()
	e.ensureCursorVisible()
}

// updateActiveBlockDisplay updates the viewport with the current gap buffer content.
func (e *EditorModel) updateActiveBlockDisplay() {
	if e.activeBuffer == nil || e.viewport == nil {
		return
	}
	e.viewport.UpdateActiveBlockContent(e.activeBuffer.Content())
	e.ensureInsertCursorVisible()
}

// ensureInsertCursorVisible ensures the insert mode cursor is visible in the viewport.
func (e *EditorModel) ensureInsertCursorVisible() {
	if e.viewport == nil || e.activeBuffer == nil || e.activeBlockIdx < 0 {
		return
	}
	bufLine, bufCol := e.activeBuffer.CursorLineCol()
	absLine, _ := e.viewport.BufferToScreenPos(bufLine, bufCol)

	offset := e.viewport.ScrollOffset()
	vpHeight := e.viewport.ViewportHeight()

	if absLine < offset {
		e.viewport.SetScrollOffset(absLine)
	}
	if absLine >= offset+vpHeight {
		e.viewport.SetScrollOffset(absLine - vpHeight + 1)
	}
}

// moveInsertCursorDown moves the insert mode cursor down one visual row,
// accounting for line wrapping at column width.
func (e *EditorModel) moveInsertCursorDown() {
	colWidth := ui.CalculateColumnWidth(e.width)
	if colWidth <= 0 {
		e.activeBuffer.MoveDown()
		return
	}

	bufLine, bufCol := e.activeBuffer.CursorLineCol()
	rawLines := strings.Split(e.activeBuffer.Content(), "\n")
	if bufLine >= len(rawLines) {
		return
	}

	currentLineLen := len([]rune(rawLines[bufLine]))
	visualCol := bufCol % colWidth

	// Check if there's a next visual row on the same actual line
	nextVisualRowStart := ((bufCol / colWidth) + 1) * colWidth
	if nextVisualRowStart < currentLineLen {
		// Move to next visual row on same line
		e.activeBuffer.SetCursorLineCol(bufLine, nextVisualRowStart+visualCol)
	} else if bufLine+1 < len(rawLines) {
		// Move to next actual line, first visual row
		e.activeBuffer.SetCursorLineCol(bufLine+1, visualCol)
	}
}

// moveInsertCursorUp moves the insert mode cursor up one visual row,
// accounting for line wrapping at column width.
func (e *EditorModel) moveInsertCursorUp() {
	colWidth := ui.CalculateColumnWidth(e.width)
	if colWidth <= 0 {
		e.activeBuffer.MoveUp()
		return
	}

	bufLine, bufCol := e.activeBuffer.CursorLineCol()
	rawLines := strings.Split(e.activeBuffer.Content(), "\n")
	visualCol := bufCol % colWidth
	currentVisualRow := bufCol / colWidth

	if currentVisualRow > 0 {
		// Move to previous visual row on same line
		e.activeBuffer.SetCursorLineCol(bufLine, (currentVisualRow-1)*colWidth+visualCol)
	} else if bufLine > 0 {
		// Move to previous actual line, last visual row
		prevLineLen := len([]rune(rawLines[bufLine-1]))
		lastVisualRowStart := 0
		if prevLineLen > colWidth {
			lastVisualRowStart = ((prevLineLen - 1) / colWidth) * colWidth
		}
		e.activeBuffer.SetCursorLineCol(bufLine-1, lastVisualRowStart+visualCol)
	}
}

func (e *EditorModel) applyWordMotion(forward bool) {
	if e.viewport == nil {
		return
	}

	contentHeight := e.viewport.ContentHeight()
	if contentHeight == 0 {
		return
	}

	leftMargin := e.leftMargin()

	// Use a bounded window around the cursor to avoid O(document) per keypress.
	const windowRadius = 50
	startLine := e.cursorLine - windowRadius
	if startLine < 0 {
		startLine = 0
	}
	endLine := e.cursorLine + windowRadius
	if endLine > contentHeight {
		endLine = contentHeight
	}

	// Calculate flat position within the window
	flatPos := 0
	for i := startLine; i < e.cursorLine; i++ {
		line := e.visibleLineContent(i, leftMargin)
		flatPos += len([]rune(line)) + 1 // +1 for newline
	}
	flatPos += e.cursorCol

	// Build flat text for the window only
	var runes []rune
	for i := startLine; i < endLine; i++ {
		if i > startLine {
			runes = append(runes, '\n')
		}
		line := e.visibleLineContent(i, leftMargin)
		runes = append(runes, []rune(line)...)
	}

	if len(runes) == 0 {
		return
	}

	if flatPos < 0 {
		flatPos = 0
	}
	if flatPos >= len(runes) {
		flatPos = len(runes) - 1
	}

	var newFlatPos int
	if forward {
		newFlatPos = vim.NextWordStart(runes, flatPos)
	} else {
		newFlatPos = vim.PrevWordStart(runes, flatPos)
	}

	// Convert flat position back to line/col within the window
	pos := 0
	for i := startLine; i < endLine; i++ {
		line := e.visibleLineContent(i, leftMargin)
		lineLen := len([]rune(line))
		if pos+lineLen >= newFlatPos {
			e.cursorLine = i
			e.cursorCol = newFlatPos - pos
			return
		}
		pos += lineLen + 1 // +1 for newline
	}

	// Fallback: end of window
	e.cursorLine = endLine - 1
	lastLineLen := e.visibleLineLength(e.cursorLine)
	if lastLineLen > 0 {
		e.cursorCol = lastLineLen - 1
	} else {
		e.cursorCol = 0
	}
}

func (e *EditorModel) ensureCursorVisible() {
	if e.viewport == nil {
		return
	}
	offset := e.viewport.ScrollOffset()
	vpHeight := e.viewport.ViewportHeight()

	if e.cursorLine < offset {
		e.viewport.SetScrollOffset(e.cursorLine)
	}
	if e.cursorLine >= offset+vpHeight {
		e.viewport.SetScrollOffset(e.cursorLine - vpHeight + 1)
	}
}

func (e *EditorModel) leftMargin() int {
	return ui.CalculateMargin(e.width, ui.CalculateColumnWidth(e.width))
}

func (e *EditorModel) visibleLineContent(line, leftMargin int) string {
	if e.viewport == nil {
		return ""
	}
	raw := e.viewport.Line(line)
	clean := vim.StripANSI(raw)
	// Remove left margin (centering spaces)
	runes := []rune(clean)
	if leftMargin > 0 && len(runes) >= leftMargin {
		return string(runes[leftMargin:])
	}
	return clean
}

func (e *EditorModel) visibleLineLength(line int) int {
	content := e.visibleLineContent(line, e.leftMargin())
	return len([]rune(content))
}

func (e *EditorModel) maxLine() int {
	if e.viewport == nil {
		return 0
	}
	max := e.viewport.ContentHeight() - 1
	if max < 0 {
		return 0
	}
	return max
}

func (e *EditorModel) clampCursor() {
	// TODO: Remove debug logging after performance investigation
	// log.Printf("[DEBUG] clampCursor called")
	e.clampCursorLine()
	e.clampCursorCol()
}

func (e *EditorModel) clampCursorLine() {
	if e.cursorLine < 0 {
		e.cursorLine = 0
	}
	max := e.maxLine()
	if e.cursorLine > max {
		e.cursorLine = max
	}
}

func (e *EditorModel) clampCursorCol() {
	if e.cursorCol < 0 {
		e.cursorCol = 0
	}
	lineLen := e.visibleLineLength(e.cursorLine)
	if lineLen > 0 && e.cursorCol >= lineLen {
		e.cursorCol = lineLen - 1
	}
	if lineLen == 0 {
		e.cursorCol = 0
	}
}

// initViewport creates renderer, cache, viewport on first WindowSizeMsg.
func (e *EditorModel) initViewport() {
	colWidth := ui.CalculateColumnWidth(e.width)
	if colWidth <= 0 {
		colWidth = 1
	}

	r, err := render.NewRenderer(colWidth)
	if err != nil {
		e.ready = true
		e.viewport = ui.NewViewport(e.width, e.height)
		return
	}
	e.renderer = r
	e.cache = render.NewRenderCache()

	_ = r.PreRenderAll(e.blocks, e.cache)

	e.dimStyle = render.DimStyle(render.SyntaxDimPercent)
	// Force terminal background detection now (not lazily on first insert mode)
	_ = e.dimStyle.Render("")

	e.viewport = ui.NewViewport(e.width, e.height)
	e.viewport.SetDimFunc(func(s string) string { return e.dimStyle.Render(s) })
	_ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)

	e.ready = true
}

