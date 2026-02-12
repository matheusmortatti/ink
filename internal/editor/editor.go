package editor

import (
	tea "charm.land/bubbletea/v2"
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
}

// NewEditor creates an EditorModel with the given file path and parsed blocks.
func NewEditor(filePath string, blocks []block.Block) *EditorModel {
	return &EditorModel{
		filePath:    filePath,
		blocks:      blocks,
		modeHandler: vim.NewNormalHandler(),
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

	screenRow := e.cursorLine - e.viewport.ScrollOffset()
	screenCol := e.leftMargin() + e.cursorCol

	v.Cursor = tea.NewCursor(screenCol, screenRow)
	v.Cursor.Shape = tea.CursorBlock

	return v
}

func (e *EditorModel) applyAction(action vim.Action) (tea.Model, tea.Cmd) {
	switch a := action.(type) {
	case vim.QuitAction:
		return e, tea.Quit

	case vim.MoveCursorAction:
		if a.Relative {
			if a.Line != 0 {
				e.cursorLine += a.Line
				e.clampCursorLine()
				// Restore desired column (sticky column for j/k)
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

	e.viewport = ui.NewViewport(e.width, e.height)
	_ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)

	e.ready = true
}

