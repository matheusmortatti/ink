package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// reparseMsg signals the document should be re-parsed.
type reparseMsg struct{}

// debounceTickMsg is sent after a debounce interval.
type debounceTickMsg struct {
	id int
}

const maxContentWidth = 100

// Model is the top-level bubbletea model.
type Model struct {
	buf      *Buffer
	vim      *VimState
	viewport *Viewport
	renderer *HybridRenderer
	doc      *Document

	width        int
	height       int
	contentWidth int
	leftMargin   int

	// Debounce state for re-parsing during insert mode
	needsReparse bool
	debounceID   int

	// Markdown mode enabled
	markdownMode bool
}

// NewModel creates a new editor model.
func NewModel(filePath string) Model {
	var buf *Buffer
	var err error

	if filePath != "" {
		buf, err = NewBufferFromFile(filePath)
		if err != nil {
			buf = NewBuffer()
			buf.filePath = filePath
		}
	} else {
		buf = NewBuffer()
	}

	m := Model{
		buf:          buf,
		vim:          NewVimState(),
		viewport:     NewViewport(80, 24),
		markdownMode: isMarkdownFile(filePath),
	}

	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle(m.windowTitle()),
	)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2 // Reserve 2 lines: status bar + command line
		m.contentWidth = min(m.width, maxContentWidth)
		m.leftMargin = (m.width - m.contentWidth) / 2
		m.viewport.Width = m.contentWidth
		m.viewport.Height = m.height
		if m.renderer == nil || m.renderer.width != m.contentWidth {
			m.renderer = NewHybridRenderer(m.contentWidth)
		}
		// Parse document on first size message
		if m.markdownMode && m.doc == nil {
			m.reparseDocument()
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case debounceTickMsg:
		if msg.id == m.debounceID && m.needsReparse {
			m.reparseDocument()
			m.needsReparse = false
		}
		return m, nil

	case SaveMsg:
		if msg.Err != nil {
			m.vim.StatusMsg = "Error saving: " + msg.Err.Error()
		} else {
			m.vim.StatusMsg = "Written: " + m.buf.filePath
		}
		return m, nil

	case QuitMsg:
		return m, tea.Quit

	case OpenFileMsg:
		newBuf, err := NewBufferFromFile(msg.Path)
		if err != nil {
			m.vim.StatusMsg = "Error: " + err.Error()
			return m, nil
		}
		m.buf = newBuf
		m.markdownMode = isMarkdownFile(msg.Path)
		if m.markdownMode {
			m.reparseDocument()
		} else {
			m.doc = nil
		}
		m.renderer.cache.Invalidate()
		return m, nil

	case StatusMsgMsg:
		m.vim.StatusMsg = msg.Text
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prevMode := m.vim.Mode

	// Handle ctrl+d / ctrl+u for half-page scrolling in normal mode
	if m.vim.Mode == ModeNormal {
		key := msg.String()
		if key == "ctrl+d" {
			halfPage := m.height / 2
			m.buf.cursor.Row += halfPage
			if m.buf.cursor.Row >= m.buf.LineCount() {
				m.buf.cursor.Row = m.buf.LineCount() - 1
			}
			m.buf.cursor.Col = clampInt(m.buf.cursor.Col, 0, max(0, len(m.buf.Line(m.buf.cursor.Row))-1))
			return m, nil
		}
		if key == "ctrl+u" {
			halfPage := m.height / 2
			m.buf.cursor.Row -= halfPage
			if m.buf.cursor.Row < 0 {
				m.buf.cursor.Row = 0
			}
			m.buf.cursor.Col = clampInt(m.buf.cursor.Col, 0, max(0, len(m.buf.Line(m.buf.cursor.Row))-1))
			return m, nil
		}
	}

	// Handle the key in vim state
	cmd := m.vim.HandleKey(msg, m.buf)

	// Clear status message on any key press (except in command mode)
	if m.vim.Mode != ModeCommand && prevMode != ModeCommand {
		if msg.String() != ":" && msg.String() != "/" && msg.String() != "?" {
			m.vim.StatusMsg = ""
		}
	}

	// Handle search commands: if we entered command mode via / or ?,
	// the executeCommand needs to know it's a search
	if prevMode == ModeCommand && m.vim.Mode == ModeNormal {
		// Command was executed, check if it was a search
		if m.vim.SearchDir != 0 && m.vim.CommandLine == "" {
			// Search was handled
		}
		// Reset search direction for next command mode entry
		// (only if we entered via :, not / or ?)
	}

	// Markdown re-parse logic
	if m.markdownMode {
		if m.vim.Mode == ModeInsert {
			// Debounced re-parse during insert mode
			if m.buf.dirty {
				m.needsReparse = true
				m.debounceID++
				id := m.debounceID
				debounceCmd := tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
					return debounceTickMsg{id: id}
				})
				if cmd != nil {
					return m, tea.Batch(cmd, debounceCmd)
				}
				return m, debounceCmd
			}
		} else if prevMode == ModeInsert && m.vim.Mode == ModeNormal {
			// Immediate re-parse on leaving insert mode
			m.reparseDocument()
			m.needsReparse = false
		} else if m.vim.Mode == ModeNormal && m.buf.dirty {
			// Immediate re-parse after normal mode mutations (dd, x, p, etc.)
			m.reparseDocument()
		}
	}

	return m, cmd
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		// Convert display coordinates to source coordinates
		displayRow := msg.Y + m.viewport.YOffset
		displayCol := msg.X - m.leftMargin
		if displayCol < 0 {
			displayCol = 0
		}

		if m.markdownMode && m.doc != nil {
			// Find which line map entry this display line falls in
			// For active blocks, we can map directly
			// For simplicity in plain text mode or when no linemap, use direct mapping
			sourceRow := displayRow
			if sourceRow >= m.buf.LineCount() {
				sourceRow = m.buf.LineCount() - 1
			}
			if sourceRow < 0 {
				sourceRow = 0
			}
			m.buf.cursor.Row = sourceRow
		} else {
			sourceRow := displayRow
			if sourceRow >= m.buf.LineCount() {
				sourceRow = m.buf.LineCount() - 1
			}
			if sourceRow < 0 {
				sourceRow = 0
			}
			m.buf.cursor.Row = sourceRow
		}

		lineLen := len(m.buf.Line(m.buf.cursor.Row))
		if m.vim.Mode == ModeInsert {
			m.buf.cursor.Col = clampInt(displayCol, 0, lineLen)
		} else {
			maxCol := lineLen - 1
			if maxCol < 0 {
				maxCol = 0
			}
			m.buf.cursor.Col = clampInt(displayCol, 0, maxCol)
		}

	case tea.MouseButtonWheelUp:
		m.viewport.ScrollUp(3)
	case tea.MouseButtonWheelDown:
		m.viewport.ScrollDown(3, m.buf.LineCount())
	}
	return m, nil
}

func (m *Model) reparseDocument() {
	content := m.buf.Content()
	m.doc = NewDocument(content)
	m.renderer.cache.Invalidate()
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	if m.renderer == nil {
		m.renderer = NewHybridRenderer(m.contentWidth)
	}

	var displayLines []string
	var cursorDisplayLine int

	if m.markdownMode && m.doc != nil {
		// Hybrid markdown rendering
		activeIdx := m.doc.BlockIndexAt(m.buf.cursor.Row)
		lineMap := m.renderer.Render(m.doc, m.buf, m.vim, activeIdx)
		displayLines = lineMap.AllDisplayLines()
		cursorDisplayLine = lineMap.SourceToDisplay(m.buf.cursor.Row)
	} else {
		// Plain text rendering
		displayLines = m.renderPlainText()
		cursorDisplayLine = m.buf.cursor.Row
	}

	// Ensure we have at least enough lines to fill the viewport
	for len(displayLines) < m.height {
		displayLines = append(displayLines, tildeLine(m.contentWidth))
	}

	// Adjust viewport
	m.viewport.EnsureCursorVisible(cursorDisplayLine, len(displayLines))

	// Slice visible lines
	start, end := m.viewport.VisibleRange()
	if end >= len(displayLines) {
		end = len(displayLines) - 1
	}
	visibleLines := displayLines[start : end+1]

	// Pad to viewport height
	for len(visibleLines) < m.height {
		visibleLines = append(visibleLines, tildeLine(m.contentWidth))
	}

	// Center content lines with margins
	margin := strings.Repeat(" ", m.leftMargin)
	rightPad := m.width - m.leftMargin - m.contentWidth
	if rightPad < 0 {
		rightPad = 0
	}
	rpad := strings.Repeat(" ", rightPad)

	// Build final output
	var sb strings.Builder
	for _, line := range visibleLines {
		sb.WriteString(margin)
		sb.WriteString(line)
		sb.WriteString(rpad)
		sb.WriteByte('\n')
	}

	// Status bar
	sb.WriteString(RenderStatusBar(m.vim, m.buf, m.width))
	sb.WriteByte('\n')

	// Command line
	sb.WriteString(RenderCommandLine(m.vim, m.width))

	return sb.String()
}

// renderPlainText renders the buffer as plain text with cursor.
func (m Model) renderPlainText() []string {
	lines := make([]string, m.buf.LineCount())
	for i := 0; i < m.buf.LineCount(); i++ {
		text := m.buf.Line(i)
		isCursorLine := i == m.buf.cursor.Row

		if isCursorLine {
			lines[i] = m.renderCursorLine(text)
		} else if m.vim.Mode == ModeVisual || m.vim.Mode == ModeVisualLine {
			start, end := m.vim.visualRange(m.buf)
			if i >= start.Row && i <= end.Row {
				lines[i] = m.renderVisualLine(text, i, start, end)
			} else {
				lines[i] = text
			}
		} else {
			lines[i] = text
		}
	}
	return lines
}

func (m Model) renderCursorLine(text string) string {
	col := m.buf.cursor.Col

	if m.vim.Mode == ModeVisual || m.vim.Mode == ModeVisualLine {
		// Render with both visual selection and cursor
		start, end := m.vim.visualRange(m.buf)
		row := m.buf.cursor.Row
		if row >= start.Row && row <= end.Row {
			if m.vim.Mode == ModeVisualLine {
				if len(text) == 0 {
					return cursorStyle.Render(" ")
				}
				if col >= len(text) {
					col = len(text) - 1
				}
				before := text[:col]
				ch := string(text[col])
				after := ""
				if col+1 < len(text) {
					after = text[col+1:]
				}
				return visualHighlightStyle.Render(before) +
					cursorStyle.Render(ch) +
					visualHighlightStyle.Render(after)
			}

			selStart := 0
			selEnd := len(text)
			if row == start.Row {
				selStart = start.Col
			}
			if row == end.Row {
				selEnd = end.Col + 1
			}
			if selEnd > len(text) {
				selEnd = len(text)
			}
			if selStart > len(text) {
				selStart = len(text)
			}

			// Build the line piece by piece
			var sb strings.Builder
			for ci := 0; ci < len(text); ci++ {
				ch := string(text[ci])
				inSel := ci >= selStart && ci < selEnd
				if ci == col {
					sb.WriteString(cursorStyle.Render(ch))
				} else if inSel {
					sb.WriteString(visualHighlightStyle.Render(ch))
				} else {
					sb.WriteString(ch)
				}
			}
			return sb.String()
		}
	}

	if m.vim.Mode == ModeInsert {
		if col >= len(text) {
			return text + cursorStyle.Render(" ")
		}
		before := text[:col]
		ch := string(text[col])
		after := text[col+1:]
		return before + cursorStyle.Render(ch) + after
	}

	// Normal mode
	if len(text) == 0 {
		return cursorStyle.Render(" ")
	}
	if col >= len(text) {
		col = len(text) - 1
	}
	if col < 0 {
		col = 0
	}
	before := text[:col]
	ch := string(text[col])
	after := ""
	if col+1 < len(text) {
		after = text[col+1:]
	}
	return before + cursorStyle.Render(ch) + after
}

func (m Model) renderVisualLine(text string, row int, start, end Position) string {
	if m.vim.Mode == ModeVisualLine {
		if len(text) == 0 {
			return visualHighlightStyle.Render(" ")
		}
		return visualHighlightStyle.Render(text)
	}

	// Character-wise visual
	selStart := 0
	selEnd := len(text)
	if row == start.Row {
		selStart = start.Col
	}
	if row == end.Row {
		selEnd = end.Col + 1
	}
	if selEnd > len(text) {
		selEnd = len(text)
	}
	if selStart > len(text) {
		selStart = len(text)
	}

	before := text[:selStart]
	selected := text[selStart:selEnd]
	after := text[selEnd:]
	return before + visualHighlightStyle.Render(selected) + after
}

func tildeLine(width int) string {
	if width <= 0 {
		return "~"
	}
	line := "~"
	if width > 1 {
		line += strings.Repeat(" ", width-1)
	}
	return line
}

func (m Model) windowTitle() string {
	if m.buf.filePath != "" {
		return "markdown-term - " + m.buf.filePath
	}
	return "markdown-term"
}

func isMarkdownFile(path string) bool {
	if path == "" {
		return false
	}
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".md") ||
		strings.HasSuffix(lower, ".markdown") ||
		strings.HasSuffix(lower, ".mkd") ||
		strings.HasSuffix(lower, ".mdx")
}
