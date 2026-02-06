package main

import (
	"os"
	"strings"
	"unicode"
)

// Position represents a cursor position in source coordinates.
type Position struct {
	Row int
	Col int
}

// snapshot holds buffer state for undo/redo.
type snapshot struct {
	lines  []string
	cursor Position
}

// Buffer is the source of truth for document content.
type Buffer struct {
	lines     []string
	cursor    Position
	dirty     bool
	filePath  string
	undoStack []snapshot
	redoStack []snapshot
}

// NewBuffer creates a buffer with a single empty line.
func NewBuffer() *Buffer {
	return &Buffer{
		lines: []string{""},
	}
}

// NewBufferFromFile loads a file into a buffer.
func NewBufferFromFile(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := &Buffer{filePath: path}
	content := string(data)
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	if content == "" {
		b.lines = []string{""}
	} else {
		b.lines = strings.Split(content, "\n")
		// Remove trailing empty line from Split if file ended with newline
		if len(b.lines) > 1 && b.lines[len(b.lines)-1] == "" {
			b.lines = b.lines[:len(b.lines)-1]
		}
	}
	return b, nil
}

// Content returns the full text content of the buffer.
func (b *Buffer) Content() string {
	return strings.Join(b.lines, "\n")
}

// LineCount returns the number of lines.
func (b *Buffer) LineCount() int {
	return len(b.lines)
}

// Line returns the content of line i, or "" if out of range.
func (b *Buffer) Line(i int) string {
	if i < 0 || i >= len(b.lines) {
		return ""
	}
	return b.lines[i]
}

// Lines returns all lines.
func (b *Buffer) Lines() []string {
	return b.lines
}

// CursorRow returns the cursor row.
func (b *Buffer) CursorRow() int {
	return b.cursor.Row
}

// CursorCol returns the cursor col.
func (b *Buffer) CursorCol() int {
	return b.cursor.Col
}

// SetCursor sets the cursor position, clamping to valid range.
func (b *Buffer) SetCursor(row, col int) {
	b.cursor.Row = b.clampRow(row)
	b.cursor.Col = b.clampCol(b.cursor.Row, col)
}

// SetCursorRow sets only the row, clamping.
func (b *Buffer) SetCursorRow(row int) {
	b.cursor.Row = b.clampRow(row)
	b.cursor.Col = b.clampCol(b.cursor.Row, b.cursor.Col)
}

// SetCursorCol sets only the col, clamping.
func (b *Buffer) SetCursorCol(col int) {
	b.cursor.Col = b.clampCol(b.cursor.Row, col)
}

// SaveSnapshot pushes the current state onto the undo stack.
func (b *Buffer) SaveSnapshot() {
	linesCopy := make([]string, len(b.lines))
	copy(linesCopy, b.lines)
	b.undoStack = append(b.undoStack, snapshot{
		lines:  linesCopy,
		cursor: b.cursor,
	})
	// Clear redo stack on new edit
	b.redoStack = nil
}

// Undo restores the previous state.
func (b *Buffer) Undo() bool {
	if len(b.undoStack) == 0 {
		return false
	}
	// Push current state to redo
	linesCopy := make([]string, len(b.lines))
	copy(linesCopy, b.lines)
	b.redoStack = append(b.redoStack, snapshot{
		lines:  linesCopy,
		cursor: b.cursor,
	})
	// Pop undo
	s := b.undoStack[len(b.undoStack)-1]
	b.undoStack = b.undoStack[:len(b.undoStack)-1]
	b.lines = s.lines
	b.cursor = s.cursor
	b.dirty = true
	return true
}

// Redo re-applies the last undone change.
func (b *Buffer) Redo() bool {
	if len(b.redoStack) == 0 {
		return false
	}
	// Push current to undo
	linesCopy := make([]string, len(b.lines))
	copy(linesCopy, b.lines)
	b.undoStack = append(b.undoStack, snapshot{
		lines:  linesCopy,
		cursor: b.cursor,
	})
	s := b.redoStack[len(b.redoStack)-1]
	b.redoStack = b.redoStack[:len(b.redoStack)-1]
	b.lines = s.lines
	b.cursor = s.cursor
	b.dirty = true
	return true
}

// InsertChar inserts a character at the cursor position.
func (b *Buffer) InsertChar(ch rune) {
	line := b.lines[b.cursor.Row]
	col := b.cursor.Col
	if col > len(line) {
		col = len(line)
	}
	b.lines[b.cursor.Row] = line[:col] + string(ch) + line[col:]
	b.cursor.Col = col + 1
	b.dirty = true
}

// InsertNewline splits the current line at the cursor.
func (b *Buffer) InsertNewline() {
	line := b.lines[b.cursor.Row]
	col := b.cursor.Col
	if col > len(line) {
		col = len(line)
	}
	before := line[:col]
	after := line[col:]
	b.lines[b.cursor.Row] = before
	// Insert new line after
	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:b.cursor.Row+1]...)
	newLines = append(newLines, after)
	newLines = append(newLines, b.lines[b.cursor.Row+1:]...)
	b.lines = newLines
	b.cursor.Row++
	b.cursor.Col = 0
	b.dirty = true
}

// Backspace deletes the character before the cursor.
func (b *Buffer) Backspace() {
	if b.cursor.Col > 0 {
		line := b.lines[b.cursor.Row]
		col := b.cursor.Col
		if col > len(line) {
			col = len(line)
		}
		b.lines[b.cursor.Row] = line[:col-1] + line[col:]
		b.cursor.Col = col - 1
		b.dirty = true
	} else if b.cursor.Row > 0 {
		// Join with previous line
		prevLine := b.lines[b.cursor.Row-1]
		curLine := b.lines[b.cursor.Row]
		b.lines[b.cursor.Row-1] = prevLine + curLine
		// Remove current line
		b.lines = append(b.lines[:b.cursor.Row], b.lines[b.cursor.Row+1:]...)
		b.cursor.Row--
		b.cursor.Col = len(prevLine)
		b.dirty = true
	}
}

// DeleteCharAtCursor deletes the character under the cursor.
func (b *Buffer) DeleteCharAtCursor() {
	line := b.lines[b.cursor.Row]
	if b.cursor.Col < len(line) {
		b.lines[b.cursor.Row] = line[:b.cursor.Col] + line[b.cursor.Col+1:]
		b.dirty = true
	} else if b.cursor.Row < len(b.lines)-1 {
		// Join with next line
		nextLine := b.lines[b.cursor.Row+1]
		b.lines[b.cursor.Row] = line + nextLine
		b.lines = append(b.lines[:b.cursor.Row+1], b.lines[b.cursor.Row+2:]...)
		b.dirty = true
	}
}

// DeleteLine deletes line at given row and returns its content.
func (b *Buffer) DeleteLine(row int) string {
	if row < 0 || row >= len(b.lines) {
		return ""
	}
	content := b.lines[row]
	if len(b.lines) == 1 {
		b.lines[0] = ""
	} else {
		b.lines = append(b.lines[:row], b.lines[row+1:]...)
	}
	if b.cursor.Row >= len(b.lines) {
		b.cursor.Row = len(b.lines) - 1
	}
	b.cursor.Col = b.clampCol(b.cursor.Row, b.cursor.Col)
	b.dirty = true
	return content
}

// DeleteLines deletes lines from startRow to endRow inclusive.
func (b *Buffer) DeleteLines(startRow, endRow int) []string {
	if startRow < 0 {
		startRow = 0
	}
	if endRow >= len(b.lines) {
		endRow = len(b.lines) - 1
	}
	if startRow > endRow {
		return nil
	}
	deleted := make([]string, endRow-startRow+1)
	copy(deleted, b.lines[startRow:endRow+1])
	if startRow == 0 && endRow == len(b.lines)-1 {
		b.lines = []string{""}
	} else {
		b.lines = append(b.lines[:startRow], b.lines[endRow+1:]...)
	}
	if b.cursor.Row >= len(b.lines) {
		b.cursor.Row = len(b.lines) - 1
	}
	b.cursor.Col = b.clampCol(b.cursor.Row, b.cursor.Col)
	b.dirty = true
	return deleted
}

// InsertLineAfter inserts text as a new line after row.
func (b *Buffer) InsertLineAfter(row int, text string) {
	if row < 0 {
		row = 0
	}
	if row >= len(b.lines) {
		row = len(b.lines) - 1
	}
	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:row+1]...)
	newLines = append(newLines, text)
	newLines = append(newLines, b.lines[row+1:]...)
	b.lines = newLines
	b.dirty = true
}

// InsertLineBefore inserts text as a new line before row.
func (b *Buffer) InsertLineBefore(row int, text string) {
	if row < 0 {
		row = 0
	}
	if row > len(b.lines) {
		row = len(b.lines)
	}
	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:row]...)
	newLines = append(newLines, text)
	newLines = append(newLines, b.lines[row:]...)
	b.lines = newLines
	b.dirty = true
}

// InsertLinesAfter inserts multiple lines after row.
func (b *Buffer) InsertLinesAfter(row int, lines []string) {
	if row < 0 {
		row = 0
	}
	if row >= len(b.lines) {
		row = len(b.lines) - 1
	}
	newLines := make([]string, 0, len(b.lines)+len(lines))
	newLines = append(newLines, b.lines[:row+1]...)
	newLines = append(newLines, lines...)
	newLines = append(newLines, b.lines[row+1:]...)
	b.lines = newLines
	b.dirty = true
}

// InsertLinesBefore inserts multiple lines before row.
func (b *Buffer) InsertLinesBefore(row int, lines []string) {
	if row < 0 {
		row = 0
	}
	if row > len(b.lines) {
		row = len(b.lines)
	}
	newLines := make([]string, 0, len(b.lines)+len(lines))
	newLines = append(newLines, b.lines[:row]...)
	newLines = append(newLines, lines...)
	newLines = append(newLines, b.lines[row:]...)
	b.lines = newLines
	b.dirty = true
}

// DeleteRange deletes text from start to end position (end exclusive).
func (b *Buffer) DeleteRange(start, end Position) string {
	if start.Row == end.Row {
		line := b.lines[start.Row]
		sc := clampInt(start.Col, 0, len(line))
		ec := clampInt(end.Col, 0, len(line))
		if sc > ec {
			sc, ec = ec, sc
		}
		deleted := line[sc:ec]
		b.lines[start.Row] = line[:sc] + line[ec:]
		b.dirty = true
		return deleted
	}
	// Multi-line delete
	if start.Row > end.Row {
		start, end = end, start
	}
	firstLine := b.lines[start.Row]
	lastLine := b.lines[end.Row]
	sc := clampInt(start.Col, 0, len(firstLine))
	ec := clampInt(end.Col, 0, len(lastLine))

	var sb strings.Builder
	sb.WriteString(firstLine[sc:])
	for r := start.Row + 1; r < end.Row; r++ {
		sb.WriteByte('\n')
		sb.WriteString(b.lines[r])
	}
	sb.WriteByte('\n')
	sb.WriteString(lastLine[:ec])

	b.lines[start.Row] = firstLine[:sc] + lastLine[ec:]
	if end.Row > start.Row {
		b.lines = append(b.lines[:start.Row+1], b.lines[end.Row+1:]...)
	}
	b.dirty = true
	return sb.String()
}

// JoinLine joins the current line with the next one.
func (b *Buffer) JoinLine(row int) {
	if row < 0 || row >= len(b.lines)-1 {
		return
	}
	cur := b.lines[row]
	next := strings.TrimLeft(b.lines[row+1], " \t")
	sep := " "
	if cur == "" || next == "" {
		sep = ""
	}
	joinCol := len(cur)
	b.lines[row] = cur + sep + next
	b.lines = append(b.lines[:row+1], b.lines[row+2:]...)
	b.cursor.Col = joinCol
	b.dirty = true
}

// ReplaceCharAtCursor replaces the character under the cursor.
func (b *Buffer) ReplaceCharAtCursor(ch rune) {
	line := b.lines[b.cursor.Row]
	if b.cursor.Col >= len(line) {
		return
	}
	b.lines[b.cursor.Row] = line[:b.cursor.Col] + string(ch) + line[b.cursor.Col+1:]
	b.dirty = true
}

// WordForwardPos returns the position of the start of the next word.
func (b *Buffer) WordForwardPos(row, col int) Position {
	line := b.lines[row]

	// Skip current word characters
	for col < len(line) && !unicode.IsSpace(rune(line[col])) {
		col++
	}
	// Skip whitespace
	for col < len(line) && unicode.IsSpace(rune(line[col])) {
		col++
	}
	// If we ran off the end, move to next line
	if col >= len(line) && row < len(b.lines)-1 {
		row++
		col = 0
		// Skip leading whitespace on new line
		line = b.lines[row]
		for col < len(line) && unicode.IsSpace(rune(line[col])) {
			col++
		}
	}
	if col > len(b.lines[row]) {
		col = len(b.lines[row])
	}
	return Position{row, col}
}

// WordForwardPosVim returns the next word position using Vim-style word detection.
// Words are sequences of word chars (alnum + _) or sequences of non-word non-space chars.
func (b *Buffer) WordForwardPosVim(row, col int) Position {
	line := b.lines[row]
	if len(line) == 0 {
		if row < len(b.lines)-1 {
			return Position{row + 1, 0}
		}
		return Position{row, 0}
	}

	if col >= len(line) {
		if row < len(b.lines)-1 {
			return Position{row + 1, 0}
		}
		return Position{row, len(line) - 1}
	}

	ch := rune(line[col])
	if isWordChar(ch) {
		// Skip word chars
		for col < len(line) && isWordChar(rune(line[col])) {
			col++
		}
	} else if !unicode.IsSpace(ch) {
		// Skip non-word non-space
		for col < len(line) && !isWordChar(rune(line[col])) && !unicode.IsSpace(rune(line[col])) {
			col++
		}
	}

	// Skip whitespace
	for col < len(line) && unicode.IsSpace(rune(line[col])) {
		col++
	}

	if col >= len(line) {
		if row < len(b.lines)-1 {
			row++
			col = 0
			line = b.lines[row]
			for col < len(line) && unicode.IsSpace(rune(line[col])) {
				col++
			}
		} else {
			col = len(line) - 1
			if col < 0 {
				col = 0
			}
		}
	}
	return Position{row, col}
}

// WordBackwardPosVim returns the previous word start position.
func (b *Buffer) WordBackwardPosVim(row, col int) Position {
	if col == 0 {
		if row > 0 {
			row--
			col = len(b.lines[row])
			if col > 0 {
				col--
			}
		} else {
			return Position{0, 0}
		}
	} else {
		col--
	}

	line := b.lines[row]
	if len(line) == 0 {
		return Position{row, 0}
	}
	if col >= len(line) {
		col = len(line) - 1
	}

	// Skip whitespace backward
	for col > 0 && unicode.IsSpace(rune(line[col])) {
		col--
	}
	if col == 0 && unicode.IsSpace(rune(line[col])) {
		if row > 0 {
			return b.WordBackwardPosVim(row, 0)
		}
		return Position{row, 0}
	}

	ch := rune(line[col])
	if isWordChar(ch) {
		for col > 0 && isWordChar(rune(line[col-1])) {
			col--
		}
	} else {
		for col > 0 && !isWordChar(rune(line[col-1])) && !unicode.IsSpace(rune(line[col-1])) {
			col--
		}
	}
	return Position{row, col}
}

// WordEndPosVim returns the position of the end of the current/next word.
func (b *Buffer) WordEndPosVim(row, col int) Position {
	line := b.lines[row]

	// Move forward one
	col++
	if col >= len(line) {
		if row < len(b.lines)-1 {
			row++
			col = 0
			line = b.lines[row]
		} else {
			return Position{row, len(line) - 1}
		}
	}

	// Skip whitespace
	for {
		if len(line) == 0 {
			if row < len(b.lines)-1 {
				row++
				col = 0
				line = b.lines[row]
				continue
			}
			return Position{row, 0}
		}
		if col < len(line) && unicode.IsSpace(rune(line[col])) {
			col++
			if col >= len(line) {
				if row < len(b.lines)-1 {
					row++
					col = 0
					line = b.lines[row]
					continue
				}
				return Position{row, len(line) - 1}
			}
			continue
		}
		break
	}

	if col >= len(line) {
		return Position{row, max(0, len(line)-1)}
	}

	ch := rune(line[col])
	if isWordChar(ch) {
		for col < len(line)-1 && isWordChar(rune(line[col+1])) {
			col++
		}
	} else {
		for col < len(line)-1 && !isWordChar(rune(line[col+1])) && !unicode.IsSpace(rune(line[col+1])) {
			col++
		}
	}
	return Position{row, col}
}

// FindCharForward finds the next occurrence of ch on the current line after col.
func (b *Buffer) FindCharForward(row, col int, ch byte) int {
	line := b.lines[row]
	for i := col + 1; i < len(line); i++ {
		if line[i] == ch {
			return i
		}
	}
	return -1
}

// FindCharBackward finds the previous occurrence of ch on the current line before col.
func (b *Buffer) FindCharBackward(row, col int, ch byte) int {
	line := b.lines[row]
	for i := col - 1; i >= 0; i-- {
		if line[i] == ch {
			return i
		}
	}
	return -1
}

// SaveFile writes the buffer to its file path.
func (b *Buffer) SaveFile() error {
	if b.filePath == "" {
		return nil
	}
	content := strings.Join(b.lines, "\n") + "\n"
	err := os.WriteFile(b.filePath, []byte(content), 0644)
	if err != nil {
		return err
	}
	b.dirty = false
	return nil
}

func (b *Buffer) clampRow(row int) int {
	return clampInt(row, 0, len(b.lines)-1)
}

func (b *Buffer) clampCol(row, col int) int {
	maxCol := len(b.lines[row])
	if maxCol > 0 {
		maxCol-- // In normal mode, can't be past last char
	}
	return clampInt(col, 0, maxCol)
}

// ClampColInsert clamps column for insert mode (can be at end of line).
func (b *Buffer) ClampColInsert(row, col int) int {
	return clampInt(col, 0, len(b.lines[row]))
}

// FirstNonBlank returns the column of the first non-whitespace character on a line.
func (b *Buffer) FirstNonBlank(row int) int {
	line := b.lines[row]
	for i, ch := range line {
		if !unicode.IsSpace(ch) {
			return i
		}
	}
	return 0
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func isWordChar(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}
