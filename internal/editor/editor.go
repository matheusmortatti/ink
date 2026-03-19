package editor

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/file"
	"github.com/matheusmortatti/ink/internal/render"
	"github.com/matheusmortatti/ink/internal/ui"
	"github.com/matheusmortatti/ink/internal/vim"
	"github.com/muesli/termenv"
)

// ErrorDismissMsg is sent by a tea.Tick timer to auto-dismiss an error message.
// The ID must match the StatusBar's current errorID to prevent stale dismissals.
type ErrorDismissMsg struct {
	ID uint64
}

// AutoSaveMsg is sent by a tea.Tick timer to trigger auto-save.
// The ID prevents stale timer triggers when the timer was reset by a later keystroke.
type AutoSaveMsg struct {
	ID uint64
}

// autoSaveDelay is the debounce duration for auto-save after the last keystroke.
const autoSaveDelay = 1 * time.Second

// EditorModel is the single Bubbletea model. Components are fields.
type EditorModel struct {
	blocks      []block.Block
	viewport    *ui.Viewport
	statusBar   *ui.StatusBar
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
	activeBlockIdx   int              // index of block being edited, -1 when none
	activeBuffer     *block.GapBuffer // gap buffer for active editing block
	undoManager      *UndoManager     // undo/redo history for the active block
	blockPendingCommit bool           // true when block is pending commit (between Esc and navigation)

	// Command mode state
	commandBuf string // accumulated command text (without the leading ':')

	// Auto-save state
	autoSaveID uint64

	// Save-as prompt state
	savePromptActive bool

	// Syntax dimming
	dimStyle lipgloss.Style

	// hasDark is the terminal background detection result, captured once in
	// NewEditor() before bubbletea starts managing stdin.
	hasDark bool
}

// NewEditor creates an EditorModel with the given file path and parsed blocks.
// Must be called before tea.NewProgram().Run(): it detects the terminal
// background (via termenv) and locks that result into lipgloss so neither
// glamour nor lipgloss ever queries the terminal again once bubbletea owns
// stdin. Querying inside bubbletea's input loop causes the terminal's
// response to arrive as garbage KeyPressMsg events.
func NewEditor(filePath string, blocks []block.Block) *EditorModel {
	hasDark := termenv.HasDarkBackground()
	lipgloss.SetHasDarkBackground(hasDark)
	return &EditorModel{
		filePath:       filePath,
		blocks:         blocks,
		modeHandler:    vim.NewNormalHandler(),
		activeBlockIdx: -1,
		hasDark:        hasDark,
		undoManager:    NewUndoManager(100),
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

// Serialize returns the current document content as bytes.
// Safe to call from a defer recover() in main — uses the same pointer that
// Bubbletea mutates throughout the session.
func (e *EditorModel) Serialize() []byte {
	return e.serializeDocument()
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
			sbRows := statusBarRows(e.height)
			_ = e.viewport.Resize(e.width, e.height-sbRows)
			if e.statusBar != nil {
				e.statusBar.Resize(e.width)
			}
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
		// Save-as prompt intercepts keys directly — not routed through vim mode handler
		if e.savePromptActive {
			return e.handleSavePromptKey(msg)
		}
		// Provide next-char context to InsertHandler for skip-over detection.
		if e.modeHandler.Mode() == vim.Insert && e.activeBuffer != nil {
			if ih, ok := e.modeHandler.(*vim.InsertHandler); ok {
				content := e.activeBuffer.Content()
				pos := e.activeBuffer.CursorPos()
				runes := []rune(content)
				var nextChar, nextNextChar rune
				if pos < len(runes) {
					nextChar = runes[pos]
				}
				if pos+1 < len(runes) {
					nextNextChar = runes[pos+1]
				}
				ih.SetNextChars(nextChar, nextNextChar)
			}
		}
		action := e.modeHandler.HandleKey(msg.String())
		return e.applyAction(action)

	case ErrorDismissMsg:
		if e.statusBar != nil {
			e.statusBar.ClearError(msg.ID)
			e.refreshStatusBar()
		}
		return e, nil

	case AutoSaveMsg:
		if msg.ID != e.autoSaveID {
			return e, nil // stale timer — a later keystroke already reset it
		}
		if e.filePath == "" {
			return e, nil // unnamed buffer — auto-save only works on named files
		}
		if err := file.WriteFile(e.filePath, e.serializeDocument()); err != nil {
			return e, e.setErrorWithTimer("E: Cannot save: " + err.Error())
		}
		return e, nil
	}

	return e, nil
}

func (e *EditorModel) View() tea.View {
	if !e.ready {
		return tea.NewView("loading...")
	}

	viewContent := e.viewport.View()
	sbRows := statusBarRows(e.height)
	var content string
	if sbRows > 0 && e.statusBar != nil {
		isDimmed := e.modeHandler.Mode() == vim.Insert && !e.blockPendingCommit
		sbLine := e.statusBar.View(isDimmed)
		if sbRows == 2 {
			content = viewContent + "\n\n" + sbLine
		} else { // sbRows == 1
			content = viewContent + "\n" + sbLine
		}
	} else {
		content = viewContent
	}

	v := tea.NewView(content)
	v.AltScreen = true

	if e.savePromptActive && statusBarRows(e.height) > 0 {
		// Save-as prompt: cursor at end of "Save as: {buf}" in the status bar row
		text := "Save as: " + e.statusBar.SavePromptBuf()
		padLeft := (e.width - len([]rune(text))) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		cursorCol := padLeft + len([]rune(text))
		sbRow := e.height - 1
		v.Cursor = tea.NewCursor(cursorCol, sbRow)
		v.Cursor.Shape = tea.CursorBar
	} else if e.modeHandler.Mode() == vim.Command && statusBarRows(e.height) > 0 {
		// Command mode: cursor appears right after ":commandBuf" in the status bar row
		text := ":" + e.commandBuf
		padLeft := (e.width - len([]rune(text))) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		cursorCol := padLeft + len([]rune(text))
		sbRow := e.height - 1 // status bar is always the last row (0-indexed)
		v.Cursor = tea.NewCursor(cursorCol, sbRow)
		v.Cursor.Shape = tea.CursorBar
	} else if e.activeBlockIdx >= 0 && e.activeBuffer != nil {
		// Insert mode (or pending-commit normal mode): cursor is within the active block
		bufLine, bufCol := e.activeBuffer.CursorLineCol()
		screenRow, screenCol := e.viewport.BufferToScreenPos(bufLine, bufCol)
		screenRow -= e.viewport.ScrollOffset()

		v.Cursor = tea.NewCursor(screenCol, screenRow)
		if e.blockPendingCommit {
			v.Cursor.Shape = tea.CursorBlock // normal mode cursor shape while pending
		} else {
			v.Cursor.Shape = tea.CursorBar
		}
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
	var autoSaveCmd tea.Cmd

	switch a := action.(type) {
	case vim.QuitAction:
		if e.blockPendingCommit {
			e.commitPendingBlock()
		}
		// Same logic as :q — prompt for save if unsaved unnamed buffer
		if e.filePath == "" && e.hasContent() {
			e.activateSavePrompt(true)
			e.refreshStatusBar()
			return e, nil
		}
		return e, tea.Quit

	case vim.ChangeModeAction:
		switch {
		case a.Mode == vim.Insert && e.modeHandler.Mode() == vim.Normal:
			if e.blockPendingCommit {
				// Re-enter insert mode in the same block — reuse existing buffer
				e.blockPendingCommit = false
				switch a.Variant {
				case "a":
					e.activeBuffer.MoveRight()
				case "o":
					e.activeBuffer.MoveToLineEnd()
					e.activeBuffer.Insert('\n')
				case "O":
					curLine, _ := e.activeBuffer.CursorLineCol()
					e.activeBuffer.SetCursorLineCol(curLine, 0)
					e.activeBuffer.MoveToLineStart()
					e.activeBuffer.Insert('\n')
					e.activeBuffer.SetCursorLineCol(curLine, 0)
				}
				if a.Variant == "o" || a.Variant == "O" {
					e.updateActiveBlockDisplay()
				}
				e.modeHandler = vim.NewInsertHandler()
			} else {
				e.enterInsertMode(a.Variant)
			}
		case a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Insert:
			if e.activeBlockIdx >= 0 && e.activeBuffer != nil {
				// First Esc: switch to normal mode but keep block alive for undo/redo
				e.blockPendingCommit = true
				e.modeHandler = vim.NewNormalHandler()
			} else {
				e.exitInsertMode()
			}
		case a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Normal:
			// Second Esc (or esc in normal mode): commit pending block if any
			if e.blockPendingCommit {
				e.commitPendingBlock()
			}
		case a.Mode == vim.Command && e.modeHandler.Mode() == vim.Normal:
			if e.blockPendingCommit {
				e.commitPendingBlock()
			}
			e.enterCommandMode()
		case a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Command:
			e.exitCommandMode()
		}

	case vim.ExecuteCommandAction:
		return e.executeCommand()

	case vim.UndoAction:
		if e.activeBuffer != nil && e.blockPendingCommit {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			if entry, ok := e.undoManager.Undo(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol); ok {
				e.activeBuffer = block.NewGapBuffer(entry.Content)
				e.activeBuffer.SetCursorPos(entry.CursorPos)
				e.updateActiveBlockDisplay()
			}
		}

	case vim.RedoAction:
		if e.activeBuffer != nil && e.blockPendingCommit {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			if entry, ok := e.undoManager.Redo(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol); ok {
				e.activeBuffer = block.NewGapBuffer(entry.Content)
				e.activeBuffer.SetCursorPos(entry.CursorPos)
				e.updateActiveBlockDisplay()
			}
		}

	case vim.InsertCharAction:
		if e.modeHandler.Mode() == vim.Command {
			e.commandBuf += string(a.Char)
		} else if e.activeBuffer != nil {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "insert")
			e.activeBuffer.Insert(a.Char)
			e.updateActiveBlockDisplay()
			autoSaveCmd = e.startAutoSaveTimer()
		}

	case vim.BackspaceAction:
		if e.modeHandler.Mode() == vim.Command {
			runes := []rune(e.commandBuf)
			if len(runes) > 0 {
				e.commandBuf = string(runes[:len(runes)-1])
			}
		} else if e.activeBuffer != nil {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "delete")
			e.activeBuffer.Backspace()
			e.updateActiveBlockDisplay()
			autoSaveCmd = e.startAutoSaveTimer()
		}

	case vim.DeleteCharAction:
		if e.activeBuffer != nil {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "delete")
			e.activeBuffer.Delete()
			e.updateActiveBlockDisplay()
			autoSaveCmd = e.startAutoSaveTimer()
		}

	case vim.InsertNewlineAction:
		if e.activeBuffer != nil {
			// Check for double-Enter at end of block to trigger block split
			content := e.activeBuffer.Content()
			cursorPos := e.activeBuffer.CursorPos()
			if cursorPos == len([]rune(content)) && strings.HasSuffix(content, "\n") {
				e.splitActiveBlock()
				e.refreshStatusBar()
				return e, e.startAutoSaveTimer()
			}
			curLine, curCol := e.activeBuffer.CursorLineCol()
			e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "newline")
			e.activeBuffer.Insert('\n')
			e.updateActiveBlockDisplay()
			autoSaveCmd = e.startAutoSaveTimer()
		}

	case vim.InsertTabAction:
		if e.activeBuffer != nil {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "insert")
			e.activeBuffer.Insert('\t')
			e.updateActiveBlockDisplay()
			autoSaveCmd = e.startAutoSaveTimer()
		}

	case vim.MoveCursorAction:
		if e.blockPendingCommit {
			e.commitPendingBlock()
		}
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
		if e.blockPendingCommit {
			e.commitPendingBlock()
		}
		e.applyWordMotion(a.Forward)
		e.desiredCol = e.cursorCol
		e.ensureCursorVisible()

	case vim.ScrollAction:
		if e.blockPendingCommit {
			e.commitPendingBlock()
		}
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
		if e.blockPendingCommit {
			e.commitPendingBlock()
		}
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

	case vim.InsertPairAction:
		if e.activeBuffer != nil {
			curLine, curCol := e.activeBuffer.CursorLineCol()
			e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "insert")
			for _, r := range a.Opening {
				e.activeBuffer.Insert(r)
			}
			for _, r := range a.Closing {
				e.activeBuffer.Insert(r)
			}
			for range a.Closing {
				e.activeBuffer.MoveLeft()
			}
			e.updateActiveBlockDisplay()
			autoSaveCmd = e.startAutoSaveTimer()
		}

	case vim.SkipClosingAction:
		if e.activeBuffer != nil {
			content := e.activeBuffer.Content()
			pos := e.activeBuffer.CursorPos()
			runes := []rune(content)
			if pos+a.Count <= len(runes) {
				for i := 0; i < a.Count; i++ {
					e.activeBuffer.MoveRight()
				}
				e.updateActiveBlockDisplay()
				// No undo record — cursor movement only, no content change.
				// No auto-save — no content change.
			}
			// If bounds check fails, no-op (chars unexpectedly missing).
		}

	case vim.MultiAction:
		var cmds []tea.Cmd
		for _, sub := range a.Actions {
			model, cmd := e.applyAction(sub)
			e = model.(*EditorModel)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return e, tea.Batch(cmds...)

	case vim.NoOpAction:
		// Nothing to do
	}

	e.refreshStatusBar()
	return e, autoSaveCmd
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
	// Record initial state so the user can undo back to the original block content
	curLine, curCol := e.activeBuffer.CursorLineCol()
	e.undoManager.Record(e.activeBuffer.Content(), e.activeBuffer.CursorPos(), curLine, curCol, "insert")
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

// commitPendingBlock commits the pending block: runs exitInsertMode(), clears undo history,
// and resets the pending-commit flag. Called when navigating away from a pending block.
func (e *EditorModel) commitPendingBlock() {
	e.exitInsertMode()
	e.undoManager.Clear()
	e.blockPendingCommit = false
}

// enterCommandMode activates command mode, clearing the command buffer.
func (e *EditorModel) enterCommandMode() {
	e.commandBuf = ""
	e.modeHandler = vim.NewCommandHandler()
}

// exitCommandMode clears the command buffer and returns to normal mode.
func (e *EditorModel) exitCommandMode() {
	e.commandBuf = ""
	e.modeHandler = vim.NewNormalHandler()
}

// executeCommand executes the current commandBuf and returns to normal mode.
// Error paths use setErrorWithTimer() which returns a tea.Cmd for the
// auto-dismiss timer, bypassing refreshStatusBar() to preserve the error.
func (e *EditorModel) executeCommand() (tea.Model, tea.Cmd) {
	cmd := e.commandBuf
	e.exitCommandMode()
	switch {
	case cmd == "q":
		// Quit: if unsaved content with no file path, show save-as prompt
		if e.filePath == "" && e.hasContent() {
			e.activateSavePrompt(true)
			return e, nil
		}
		return e, tea.Quit

	case cmd == "wq":
		// Write + quit
		if e.filePath != "" {
			if err := file.WriteFile(e.filePath, e.serializeDocument()); err != nil {
				return e, e.setErrorWithTimer("E: " + err.Error())
			}
			return e, tea.Quit
		}
		// No file path — prompt for save-as then quit
		e.activateSavePrompt(true)
		return e, nil

	case strings.HasPrefix(cmd, "w "):
		// Write to specific path: :w <path>
		path := strings.TrimSpace(cmd[2:])
		if path == "" {
			return e, e.setErrorWithTimer("E: No path specified")
		}
		if err := file.ValidatePath(path); err != nil {
			return e, e.setErrorWithTimer("E: Only .md files supported")
		}
		if err := file.WriteFile(path, e.serializeDocument()); err != nil {
			return e, e.setErrorWithTimer("E: " + err.Error())
		}
		e.filePath = path
		e.refreshStatusBar()
		return e, nil

	case cmd == "w":
		// Write only
		if e.filePath != "" {
			if err := file.WriteFile(e.filePath, e.serializeDocument()); err != nil {
				return e, e.setErrorWithTimer("E: " + err.Error())
			}
			e.refreshStatusBar()
			return e, nil
		}
		// No file path — prompt for save-as
		e.activateSavePrompt(false)
		return e, nil

	case cmd == "":
		// Empty command — silently return to normal mode
		e.refreshStatusBar()

	default:
		if e.statusBar != nil {
			return e, e.setErrorWithTimer("E: Not a command: " + cmd)
		}
	}
	return e, nil
}

// startAutoSaveTimer resets the debounce timer and returns the timer tea.Cmd.
// Each call increments autoSaveID, invalidating any pending stale timers.
func (e *EditorModel) startAutoSaveTimer() tea.Cmd {
	e.autoSaveID++
	id := e.autoSaveID
	return tea.Tick(autoSaveDelay, func(_ time.Time) tea.Msg {
		return AutoSaveMsg{ID: id}
	})
}

// setErrorWithTimer displays an error in the status bar and returns a tea.Cmd
// that auto-dismisses it after 3 seconds.
func (e *EditorModel) setErrorWithTimer(msg string) tea.Cmd {
	if e.statusBar == nil {
		return nil
	}
	e.statusBar.SetError(msg)
	id := e.statusBar.ErrorID()
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return ErrorDismissMsg{ID: id}
	})
}

// activateSavePrompt activates the save-as prompt in the status bar.
// When quitAfter is true, a successful save exits the editor.
func (e *EditorModel) activateSavePrompt(quitAfter bool) {
	e.savePromptActive = true
	if e.statusBar != nil {
		e.statusBar.SetSavePrompt(quitAfter)
	}
}

// cancelSavePrompt deactivates the save-as prompt and returns to normal mode.
func (e *EditorModel) cancelSavePrompt() {
	e.savePromptActive = false
	if e.statusBar != nil {
		e.statusBar.ClearSavePrompt()
	}
	e.refreshStatusBar()
}

// handleSavePromptKey processes keyboard input while the save-as prompt is active.
func (e *EditorModel) handleSavePromptKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Code == tea.KeyEscape:
		e.cancelSavePrompt()
		return e, nil

	case msg.Code == tea.KeyEnter:
		return e.attemptSaveAs()

	case msg.Code == tea.KeyBackspace:
		if e.statusBar != nil {
			e.statusBar.BackspaceSavePrompt()
		}
		return e, nil

	default:
		// Printable characters — ignore while error overlay is covering the prompt
		if msg.Code >= 32 && msg.Code != tea.KeyDelete {
			if e.statusBar != nil && !e.statusBar.HasError() {
				e.statusBar.AppendSavePrompt(rune(msg.Code))
			}
		}
		return e, nil
	}
}

// attemptSaveAs validates the path and writes the document.
func (e *EditorModel) attemptSaveAs() (tea.Model, tea.Cmd) {
	if e.statusBar == nil {
		return e, nil
	}
	path := e.statusBar.SavePromptBuf()
	quitAfter := e.statusBar.SavePromptQuitAfter()

	if err := file.ValidatePath(path); err != nil {
		return e, e.setErrorWithTimer("E: Invalid path: " + path)
	}

	if err := file.WriteFile(path, e.serializeDocument()); err != nil {
		return e, e.setErrorWithTimer("E: " + err.Error())
	}

	// Success: update file path, clear prompt
	e.filePath = path
	e.cancelSavePrompt()

	if quitAfter {
		return e, tea.Quit
	}
	return e, nil
}

// serializeDocument joins all block raw content with double-newline separators
// and appends a trailing newline (POSIX convention).
func (e *EditorModel) serializeDocument() []byte {
	parts := make([]string, 0, len(e.blocks))
	for i, b := range e.blocks {
		raw := b.Raw
		if i == e.activeBlockIdx && e.activeBuffer != nil {
			raw = e.activeBuffer.Content()
		}
		parts = append(parts, raw)
	}
	return []byte(strings.Join(parts, "\n\n") + "\n")
}

// hasContent reports whether any block contains non-empty raw content.
func (e *EditorModel) hasContent() bool {
	for _, b := range e.blocks {
		if b.Raw != "" {
			return true
		}
	}
	return false
}

// splitActiveBlock splits the current block at the double-Enter point,
// rendering the current block and creating a new empty block below.
func (e *EditorModel) splitActiveBlock() {
	if e.activeBlockIdx < 0 || e.activeBuffer == nil {
		return
	}

	// Get content and trim exactly one trailing newline (the first Enter that triggered detection)
	content := e.activeBuffer.Content()
	trimmed := strings.TrimSuffix(content, "\n")

	// Update the current block's raw content
	oldBlock := e.blocks[e.activeBlockIdx]
	e.cache.InvalidateBlock(oldBlock)
	e.blocks[e.activeBlockIdx].Raw = trimmed

	// Create new empty paragraph block
	newBlock := block.Block{Type: block.Paragraph, Raw: ""}

	// Insert into blocks slice at activeBlockIdx + 1
	idx := e.activeBlockIdx + 1
	e.blocks = append(e.blocks, block.Block{}) // grow
	copy(e.blocks[idx+1:], e.blocks[idx:])     // shift right
	e.blocks[idx] = newBlock                    // insert

	// Clear active block state and recompose viewport
	_ = e.viewport.ClearActiveBlock()
	_ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)

	// Clear undo history from previous block before entering the new one
	e.undoManager.Clear()

	// Enter insert mode on the new block
	e.activeBlockIdx = idx
	e.activeBuffer = block.NewGapBuffer("")
	if err := e.viewport.SetActiveBlock(idx, ""); err != nil {
		e.activeBlockIdx = -1
		e.activeBuffer = nil
		return
	}
	// Record initial state so user can undo back to empty content in the new block
	e.undoManager.Record("", 0, 0, 0, "insert")
	e.modeHandler = vim.NewInsertHandler()
	e.ensureInsertCursorVisible()
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

// statusBarRows returns how many terminal rows the status bar occupies,
// based on terminal height. Returns 2 (content + blank separator + status bar)
// for 10+ rows, 1 (status bar only, no separator) for 5-9 rows, 0 (hidden) for <5 rows.
func statusBarRows(terminalHeight int) int {
	if terminalHeight >= 10 {
		return 2
	}
	if terminalHeight >= 5 {
		return 1
	}
	return 0
}

// computeDocumentCounts returns the total word and character counts across all blocks.
// For the active editing block, it uses the live gap buffer content to reflect uncommitted edits.
func (e *EditorModel) computeDocumentCounts() (words, chars int) {
	for i, b := range e.blocks {
		raw := b.Raw
		if i == e.activeBlockIdx && e.activeBuffer != nil {
			raw = e.activeBuffer.Content()
		}
		chars += len([]rune(raw))
		words += len(strings.Fields(raw))
	}
	return
}

// refreshStatusBar updates the status bar with the current mode and document counts.
// In command mode, shows the command buffer instead of mode/counts.
// When the save-as prompt is active, skips the update (save prompt manages its own display).
// No-ops if e.statusBar is nil.
func (e *EditorModel) refreshStatusBar() {
	if e.statusBar == nil {
		return
	}
	if e.savePromptActive {
		return
	}
	if e.modeHandler.Mode() == vim.Command {
		e.statusBar.SetCommand(e.commandBuf)
		return
	}
	words, chars := e.computeDocumentCounts()
	e.statusBar.Set(e.modeHandler.Mode().String(), words, chars)
}

// initViewport creates renderer, cache, viewport on first WindowSizeMsg.
func (e *EditorModel) initViewport() {
	colWidth := ui.CalculateColumnWidth(e.width)
	if colWidth <= 0 {
		colWidth = 1
	}

	sbRows := statusBarRows(e.height)

	r, err := render.NewRenderer(colWidth, e.hasDark)
	if err != nil {
		e.ready = true
		e.viewport = ui.NewViewport(e.width, e.height-sbRows)
		e.statusBar = ui.NewStatusBar(e.width)
		return
	}
	e.renderer = r
	e.cache = render.NewRenderCache()

	_ = r.PreRenderAll(e.blocks, e.cache)

	e.dimStyle = render.DimStyle(render.SyntaxDimPercent)

	e.viewport = ui.NewViewport(e.width, e.height-sbRows)
	e.viewport.SetDimFunc(func(s string) string { return e.dimStyle.Render(s) })
	_ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)
	e.statusBar = ui.NewStatusBar(e.width)

	e.ready = true

	// Context-aware startup: blank/empty content → insert mode.
	// Bypasses enterInsertMode() intentionally: the block is empty so cursor
	// mapping is a no-op, and we avoid the overhead of MapRenderedToRaw on a
	// zero-content block. If enterInsertMode() gains new responsibilities,
	// this path must be updated to match.
	if e.isContentEmpty() {
		if len(e.blocks) == 0 {
			e.blocks = []block.Block{{Type: block.Paragraph, Raw: ""}}
			_ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)
		}
		e.activeBlockIdx = 0
		e.activeBuffer = block.NewGapBuffer("")
		_ = e.viewport.SetActiveBlock(0, "")
		// Record initial state so user can undo back to empty content
		e.undoManager.Record("", 0, 0, 0, "insert")
		e.modeHandler = vim.NewInsertHandler()
		e.ensureInsertCursorVisible()
	}

	e.refreshStatusBar()
}

// isContentEmpty returns true if the document has no meaningful content.
func (e *EditorModel) isContentEmpty() bool {
	if len(e.blocks) == 0 {
		return true
	}
	for _, b := range e.blocks {
		if b.Raw != "" {
			return false
		}
	}
	return true
}

