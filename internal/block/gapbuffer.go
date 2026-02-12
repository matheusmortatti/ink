package block

const defaultGapSize = 64

// GapBuffer provides O(1) insert and delete at the cursor position.
// It maintains a gap in the backing buffer at the cursor location,
// allowing constant-time character insertion and deletion.
type GapBuffer struct {
	buf      []rune
	gapStart int
	gapEnd   int
}

// NewGapBuffer creates a new GapBuffer initialized with the given content.
// The cursor is positioned at the end of the content.
func NewGapBuffer(content string) *GapBuffer {
	runes := []rune(content)
	buf := make([]rune, len(runes)+defaultGapSize)
	copy(buf, runes)
	return &GapBuffer{
		buf:      buf,
		gapStart: len(runes),
		gapEnd:   len(buf),
	}
}

// Insert inserts a rune at the current cursor position.
func (g *GapBuffer) Insert(r rune) {
	if g.gapStart == g.gapEnd {
		g.grow()
	}
	g.buf[g.gapStart] = r
	g.gapStart++
}

// InsertString inserts a string at the current cursor position.
func (g *GapBuffer) InsertString(s string) {
	for _, r := range s {
		g.Insert(r)
	}
}

// Backspace deletes the rune before the cursor. Returns false if at start.
func (g *GapBuffer) Backspace() bool {
	if g.gapStart == 0 {
		return false
	}
	g.gapStart--
	return true
}

// Delete deletes the rune after the cursor. Returns false if at end.
func (g *GapBuffer) Delete() bool {
	if g.gapEnd == len(g.buf) {
		return false
	}
	g.gapEnd++
	return true
}

// Content returns the full text content by concatenating pre-gap and post-gap regions.
func (g *GapBuffer) Content() string {
	return string(g.buf[:g.gapStart]) + string(g.buf[g.gapEnd:])
}

// MoveLeft moves the cursor one position to the left. Returns false if at start.
func (g *GapBuffer) MoveLeft() bool {
	if g.gapStart == 0 {
		return false
	}
	g.gapEnd--
	g.buf[g.gapEnd] = g.buf[g.gapStart-1]
	g.gapStart--
	return true
}

// MoveRight moves the cursor one position to the right. Returns false if at end.
func (g *GapBuffer) MoveRight() bool {
	if g.gapEnd == len(g.buf) {
		return false
	}
	g.buf[g.gapStart] = g.buf[g.gapEnd]
	g.gapStart++
	g.gapEnd++
	return true
}

// MoveToStart moves the cursor to the beginning of the content.
func (g *GapBuffer) MoveToStart() {
	for g.gapStart > 0 {
		g.gapEnd--
		g.buf[g.gapEnd] = g.buf[g.gapStart-1]
		g.gapStart--
	}
}

// MoveToEnd moves the cursor to the end of the content.
func (g *GapBuffer) MoveToEnd() {
	for g.gapEnd < len(g.buf) {
		g.buf[g.gapStart] = g.buf[g.gapEnd]
		g.gapStart++
		g.gapEnd++
	}
}

// CursorPos returns the current cursor position as a rune offset in the content.
func (g *GapBuffer) CursorPos() int {
	return g.gapStart
}

// SetCursorPos moves the cursor to an absolute rune position, clamped to valid range.
func (g *GapBuffer) SetCursorPos(pos int) {
	length := g.Length()
	if pos < 0 {
		pos = 0
	}
	if pos > length {
		pos = length
	}
	for g.gapStart > pos {
		g.gapEnd--
		g.buf[g.gapEnd] = g.buf[g.gapStart-1]
		g.gapStart--
	}
	for g.gapStart < pos {
		g.buf[g.gapStart] = g.buf[g.gapEnd]
		g.gapStart++
		g.gapEnd++
	}
}

// Length returns the total content length in runes (excludes gap).
func (g *GapBuffer) Length() int {
	return len(g.buf) - (g.gapEnd - g.gapStart)
}

// LineCount returns the number of lines in the content.
func (g *GapBuffer) LineCount() int {
	count := 1
	for i := 0; i < g.gapStart; i++ {
		if g.buf[i] == '\n' {
			count++
		}
	}
	for i := g.gapEnd; i < len(g.buf); i++ {
		if g.buf[i] == '\n' {
			count++
		}
	}
	return count
}

// CursorLineCol returns the cursor position as (line, col), both 0-indexed.
func (g *GapBuffer) CursorLineCol() (int, int) {
	line := 0
	col := 0
	for i := 0; i < g.gapStart; i++ {
		if g.buf[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}

// SetCursorLineCol moves the cursor to the specified (line, col) position, both 0-indexed.
// If the line is beyond content, clamps to last line. If col is beyond line length, clamps to end of line.
func (g *GapBuffer) SetCursorLineCol(line, col int) {
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}

	// Move cursor to start to scan from beginning
	g.MoveToStart()

	currentLine := 0
	currentCol := 0

	for currentLine < line {
		if g.gapEnd == len(g.buf) {
			// Reached end of content before target line; stay at end
			break
		}
		if g.buf[g.gapEnd] == '\n' {
			currentLine++
		}
		g.buf[g.gapStart] = g.buf[g.gapEnd]
		g.gapStart++
		g.gapEnd++
	}

	// Now on the target line (or last line if line was out of range), move to target col
	for currentCol < col {
		if g.gapEnd == len(g.buf) {
			break
		}
		if g.buf[g.gapEnd] == '\n' {
			// Don't cross line boundary
			break
		}
		g.buf[g.gapStart] = g.buf[g.gapEnd]
		g.gapStart++
		g.gapEnd++
		currentCol++
	}
}

// MoveUp moves the cursor to the same column on the previous line.
// Returns false if already on the first line.
func (g *GapBuffer) MoveUp() bool {
	line, col := g.CursorLineCol()
	if line == 0 {
		return false
	}
	g.SetCursorLineCol(line-1, col)
	return true
}

// MoveDown moves the cursor to the same column on the next line.
// Returns false if already on the last line.
func (g *GapBuffer) MoveDown() bool {
	line, col := g.CursorLineCol()
	// Check post-gap for a '\n' instead of scanning the entire buffer via LineCount()
	hasNextLine := false
	for i := g.gapEnd; i < len(g.buf); i++ {
		if g.buf[i] == '\n' {
			hasNextLine = true
			break
		}
	}
	if !hasNextLine {
		return false
	}
	g.SetCursorLineCol(line+1, col)
	return true
}

// MoveToLineStart moves the cursor to the beginning of the current line.
func (g *GapBuffer) MoveToLineStart() {
	for g.gapStart > 0 && g.buf[g.gapStart-1] != '\n' {
		g.gapEnd--
		g.buf[g.gapEnd] = g.buf[g.gapStart-1]
		g.gapStart--
	}
}

// MoveToLineEnd moves the cursor to the end of the current line.
func (g *GapBuffer) MoveToLineEnd() {
	for g.gapEnd < len(g.buf) && g.buf[g.gapEnd] != '\n' {
		g.buf[g.gapStart] = g.buf[g.gapEnd]
		g.gapStart++
		g.gapEnd++
	}
}

// grow expands the backing buffer when the gap is exhausted.
func (g *GapBuffer) grow() {
	newSize := len(g.buf) * 2
	if newSize < len(g.buf)+defaultGapSize {
		newSize = len(g.buf) + defaultGapSize
	}
	newBuf := make([]rune, newSize)
	copy(newBuf, g.buf[:g.gapStart])
	newGapEnd := newSize - (len(g.buf) - g.gapEnd)
	copy(newBuf[newGapEnd:], g.buf[g.gapEnd:])
	g.buf = newBuf
	g.gapEnd = newGapEnd
}
