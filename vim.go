package main

import (
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the Vim editing mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeVisualLine
	ModeCommand
	ModeReplace // single-char replace (r)
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeVisualLine:
		return "V-LINE"
	case ModeCommand:
		return "COMMAND"
	case ModeReplace:
		return "REPLACE"
	default:
		return "UNKNOWN"
	}
}

// PendingOp represents an operator waiting for a motion.
type PendingOp int

const (
	OpNone PendingOp = iota
	OpDelete
	OpChange
	OpYank
)

// VimState manages modal editing state.
type VimState struct {
	Mode          Mode
	PendingOp     PendingOp
	Count         int
	KeyBuffer     string
	VisualStart   Position
	Registers     map[byte]RegisterEntry
	CommandLine   string
	CommandPrefix byte // ':' for ex commands, '/' for forward search, '?' for backward search
	CmdCursor     int
	SearchQuery   string
	SearchDir     int // 1 = forward, -1 = backward
	StatusMsg     string
	LastFindCh    byte
	LastFindDir   int // 1 = f/t forward, -1 = F/T backward
	LastFindTil   bool
}

// RegisterEntry holds yanked/deleted content.
type RegisterEntry struct {
	Text   string
	IsLine bool // true for linewise yank/delete
}

// NewVimState creates a new vim state in normal mode.
func NewVimState() *VimState {
	return &VimState{
		Mode:      ModeNormal,
		Registers: make(map[byte]RegisterEntry),
		SearchDir: 1,
	}
}

// getCount returns the accumulated count or 1 if none specified.
func (v *VimState) getCount() int {
	if v.Count == 0 {
		return 1
	}
	return v.Count
}

// HandleKey processes a key event and returns a tea.Cmd if any.
func (v *VimState) HandleKey(msg tea.KeyMsg, buf *Buffer) tea.Cmd {
	switch v.Mode {
	case ModeNormal:
		return v.handleNormal(msg, buf)
	case ModeInsert:
		return v.handleInsert(msg, buf)
	case ModeVisual, ModeVisualLine:
		return v.handleVisual(msg, buf)
	case ModeCommand:
		return v.handleCommand(msg, buf)
	case ModeReplace:
		return v.handleReplace(msg, buf)
	}
	return nil
}

func (v *VimState) handleNormal(msg tea.KeyMsg, buf *Buffer) tea.Cmd {
	key := msg.String()

	// Handle pending key sequences first
	if v.KeyBuffer != "" {
		return v.handleKeySequence(key, buf)
	}

	// Accumulate count prefix
	if key >= "1" && key <= "9" && v.Count >= 0 {
		v.Count = v.Count*10 + int(key[0]-'0')
		return nil
	}
	if key == "0" && v.Count > 0 {
		v.Count = v.Count*10
		return nil
	}

	count := v.getCount()

	switch key {
	// Movement
	case "h", "left":
		for i := 0; i < count; i++ {
			if buf.cursor.Col > 0 {
				buf.cursor.Col--
			}
		}
	case "l", "right":
		for i := 0; i < count; i++ {
			maxCol := len(buf.Line(buf.cursor.Row)) - 1
			if maxCol < 0 {
				maxCol = 0
			}
			if buf.cursor.Col < maxCol {
				buf.cursor.Col++
			}
		}
	case "j", "down":
		for i := 0; i < count; i++ {
			if buf.cursor.Row < buf.LineCount()-1 {
				buf.cursor.Row++
			}
		}
		buf.cursor.Col = clampInt(buf.cursor.Col, 0, max(0, len(buf.Line(buf.cursor.Row))-1))
	case "k", "up":
		for i := 0; i < count; i++ {
			if buf.cursor.Row > 0 {
				buf.cursor.Row--
			}
		}
		buf.cursor.Col = clampInt(buf.cursor.Col, 0, max(0, len(buf.Line(buf.cursor.Row))-1))
	case "0":
		buf.cursor.Col = 0
	case "$":
		lineLen := len(buf.Line(buf.cursor.Row))
		if lineLen > 0 {
			buf.cursor.Col = lineLen - 1
		} else {
			buf.cursor.Col = 0
		}
	case "^":
		buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)

	// Word motions
	case "w":
		if v.PendingOp != OpNone {
			return v.execOperatorMotion(buf, func() Position {
				pos := buf.cursor
				for i := 0; i < count; i++ {
					pos = buf.WordForwardPosVim(pos.Row, pos.Col)
				}
				return pos
			})
		}
		for i := 0; i < count; i++ {
			p := buf.WordForwardPosVim(buf.cursor.Row, buf.cursor.Col)
			buf.cursor = p
		}
	case "b":
		if v.PendingOp != OpNone {
			return v.execOperatorMotion(buf, func() Position {
				pos := buf.cursor
				for i := 0; i < count; i++ {
					pos = buf.WordBackwardPosVim(pos.Row, pos.Col)
				}
				return pos
			})
		}
		for i := 0; i < count; i++ {
			p := buf.WordBackwardPosVim(buf.cursor.Row, buf.cursor.Col)
			buf.cursor = p
		}
	case "e":
		if v.PendingOp != OpNone {
			return v.execOperatorMotion(buf, func() Position {
				pos := buf.cursor
				for i := 0; i < count; i++ {
					pos = buf.WordEndPosVim(pos.Row, pos.Col)
				}
				// For operators, end motion is inclusive (add 1 to col for delete range)
				pos.Col++
				return pos
			})
		}
		for i := 0; i < count; i++ {
			p := buf.WordEndPosVim(buf.cursor.Row, buf.cursor.Col)
			buf.cursor = p
		}

	// Mode entry
	case "i":
		v.enterInsert(buf)
	case "I":
		buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
		v.enterInsert(buf)
	case "a":
		lineLen := len(buf.Line(buf.cursor.Row))
		if lineLen > 0 {
			buf.cursor.Col++
		}
		v.enterInsert(buf)
	case "A":
		buf.cursor.Col = len(buf.Line(buf.cursor.Row))
		v.enterInsert(buf)
	case "o":
		buf.SaveSnapshot()
		buf.InsertLineAfter(buf.cursor.Row, "")
		buf.cursor.Row++
		buf.cursor.Col = 0
		v.Mode = ModeInsert
	case "O":
		buf.SaveSnapshot()
		buf.InsertLineBefore(buf.cursor.Row, "")
		buf.cursor.Col = 0
		v.Mode = ModeInsert

	// Editing
	case "x":
		buf.SaveSnapshot()
		for i := 0; i < count; i++ {
			if buf.cursor.Col < len(buf.Line(buf.cursor.Row)) {
				ch := string(buf.Line(buf.cursor.Row)[buf.cursor.Col])
				v.Registers['"'] = RegisterEntry{Text: ch}
				buf.DeleteCharAtCursor()
			}
		}
		// Clamp cursor
		lineLen := len(buf.Line(buf.cursor.Row))
		if lineLen > 0 && buf.cursor.Col >= lineLen {
			buf.cursor.Col = lineLen - 1
		}
	case "r":
		v.Mode = ModeReplace
	case "J":
		buf.SaveSnapshot()
		for i := 0; i < count; i++ {
			buf.JoinLine(buf.cursor.Row)
		}

	// Operators
	case "d":
		if v.PendingOp == OpDelete {
			// dd - delete line
			buf.SaveSnapshot()
			lines := make([]string, 0, count)
			for i := 0; i < count; i++ {
				if buf.cursor.Row < buf.LineCount() {
					lines = append(lines, buf.DeleteLine(buf.cursor.Row))
				}
			}
			v.Registers['"'] = RegisterEntry{Text: strings.Join(lines, "\n"), IsLine: true}
			// Position cursor at first non-blank
			buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
			v.PendingOp = OpNone
			v.Count = 0
			return nil
		}
		v.PendingOp = OpDelete
		return nil
	case "c":
		if v.PendingOp == OpChange {
			// cc - change entire line
			buf.SaveSnapshot()
			line := buf.Line(buf.cursor.Row)
			v.Registers['"'] = RegisterEntry{Text: line}
			indent := ""
			for _, ch := range line {
				if ch == ' ' || ch == '\t' {
					indent += string(ch)
				} else {
					break
				}
			}
			buf.lines[buf.cursor.Row] = indent
			buf.cursor.Col = len(indent)
			v.PendingOp = OpNone
			v.Count = 0
			v.Mode = ModeInsert
			return nil
		}
		v.PendingOp = OpChange
		return nil
	case "y":
		if v.PendingOp == OpYank {
			// yy - yank line
			lines := make([]string, 0, count)
			for i := 0; i < count; i++ {
				row := buf.cursor.Row + i
				if row < buf.LineCount() {
					lines = append(lines, buf.Line(row))
				}
			}
			v.Registers['"'] = RegisterEntry{Text: strings.Join(lines, "\n"), IsLine: true}
			v.StatusMsg = "yanked " + strconv.Itoa(len(lines)) + " line(s)"
			v.PendingOp = OpNone
			v.Count = 0
			return nil
		}
		v.PendingOp = OpYank
		return nil

	// Paste
	case "p":
		reg := v.Registers['"']
		if reg.Text != "" {
			buf.SaveSnapshot()
			if reg.IsLine {
				lines := strings.Split(reg.Text, "\n")
				for i := 0; i < count; i++ {
					buf.InsertLinesAfter(buf.cursor.Row, lines)
				}
				buf.cursor.Row++
				buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
			} else {
				for i := 0; i < count; i++ {
					line := buf.lines[buf.cursor.Row]
					col := buf.cursor.Col + 1
					if col > len(line) {
						col = len(line)
					}
					buf.lines[buf.cursor.Row] = line[:col] + reg.Text + line[col:]
					buf.cursor.Col = col + len(reg.Text) - 1
				}
				buf.dirty = true
			}
		}
	case "P":
		reg := v.Registers['"']
		if reg.Text != "" {
			buf.SaveSnapshot()
			if reg.IsLine {
				lines := strings.Split(reg.Text, "\n")
				for i := 0; i < count; i++ {
					buf.InsertLinesBefore(buf.cursor.Row, lines)
				}
				buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
			} else {
				for i := 0; i < count; i++ {
					line := buf.lines[buf.cursor.Row]
					col := buf.cursor.Col
					buf.lines[buf.cursor.Row] = line[:col] + reg.Text + line[col:]
					buf.cursor.Col = col + len(reg.Text) - 1
				}
				buf.dirty = true
			}
		}

	// Undo/Redo
	case "u":
		buf.Undo()
	case "ctrl+r":
		buf.Redo()

	// Visual mode
	case "v":
		v.Mode = ModeVisual
		v.VisualStart = buf.cursor
	case "V":
		v.Mode = ModeVisualLine
		v.VisualStart = buf.cursor

	// Command mode
	case ":":
		v.Mode = ModeCommand
		v.CommandLine = ""
		v.CommandPrefix = ':'
		v.CmdCursor = 0

	// Search
	case "/":
		v.Mode = ModeCommand
		v.CommandLine = ""
		v.CommandPrefix = '/'
		v.CmdCursor = 0
		v.SearchDir = 1
		return nil
	case "?":
		v.Mode = ModeCommand
		v.CommandLine = ""
		v.CommandPrefix = '?'
		v.CmdCursor = 0
		v.SearchDir = -1
		return nil
	case "n":
		if v.SearchQuery != "" {
			for i := 0; i < count; i++ {
				v.searchNext(buf, v.SearchDir)
			}
		}
	case "N":
		if v.SearchQuery != "" {
			for i := 0; i < count; i++ {
				v.searchNext(buf, -v.SearchDir)
			}
		}

	// Find char
	case "f", "F", "t", "T":
		v.KeyBuffer = key
		return nil

	// Go-to sequences
	case "g":
		v.KeyBuffer = "g"
		return nil

	// Misc
	case "D":
		// Delete to end of line
		buf.SaveSnapshot()
		line := buf.Line(buf.cursor.Row)
		if buf.cursor.Col < len(line) {
			deleted := line[buf.cursor.Col:]
			v.Registers['"'] = RegisterEntry{Text: deleted}
			buf.lines[buf.cursor.Row] = line[:buf.cursor.Col]
			buf.dirty = true
		}
		if buf.cursor.Col > 0 {
			buf.cursor.Col--
		}
	case "C":
		// Change to end of line
		buf.SaveSnapshot()
		line := buf.Line(buf.cursor.Row)
		if buf.cursor.Col < len(line) {
			deleted := line[buf.cursor.Col:]
			v.Registers['"'] = RegisterEntry{Text: deleted}
			buf.lines[buf.cursor.Row] = line[:buf.cursor.Col]
			buf.dirty = true
		}
		v.Mode = ModeInsert

	// gg/G handling in key sequence
	case "G":
		if v.Count > 0 {
			buf.cursor.Row = clampInt(v.Count-1, 0, buf.LineCount()-1)
		} else {
			buf.cursor.Row = buf.LineCount() - 1
		}
		buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)

	// Half-page scrolling
	case "ctrl+d":
		// handled by model for viewport awareness
		return nil
	case "ctrl+u":
		return nil

	// Pending operator motions: $ 0 ^ h l j k for operators
	default:
		if v.PendingOp != OpNone {
			return v.handlePendingMotion(key, buf)
		}
	}

	// Handle pending operator motions for movement keys already dispatched
	if v.PendingOp != OpNone && (key == "h" || key == "l" || key == "j" || key == "k" ||
		key == "$" || key == "0" || key == "^" || key == "G") {
		// These were already handled above as both motion and op
		// But we need to handle them as motions for operators too
	}

	// Reset count and pending after non-accumulating keys
	if key < "0" || key > "9" {
		if v.PendingOp == OpNone {
			v.Count = 0
		}
	}

	return nil
}

func (v *VimState) handlePendingMotion(key string, buf *Buffer) tea.Cmd {
	count := v.getCount()

	switch key {
	case "$":
		return v.execOperatorMotion(buf, func() Position {
			return Position{buf.cursor.Row, len(buf.Line(buf.cursor.Row))}
		})
	case "0":
		return v.execOperatorMotion(buf, func() Position {
			return Position{buf.cursor.Row, 0}
		})
	case "^":
		return v.execOperatorMotion(buf, func() Position {
			return Position{buf.cursor.Row, buf.FirstNonBlank(buf.cursor.Row)}
		})
	case "h", "left":
		return v.execOperatorMotion(buf, func() Position {
			col := buf.cursor.Col - count
			if col < 0 {
				col = 0
			}
			return Position{buf.cursor.Row, col}
		})
	case "l", "right":
		return v.execOperatorMotion(buf, func() Position {
			col := buf.cursor.Col + count
			lineLen := len(buf.Line(buf.cursor.Row))
			if col > lineLen {
				col = lineLen
			}
			return Position{buf.cursor.Row, col}
		})
	case "j", "down":
		// Linewise motion
		return v.execOperatorLinewise(buf, buf.cursor.Row, buf.cursor.Row+count)
	case "k", "up":
		return v.execOperatorLinewise(buf, buf.cursor.Row-count, buf.cursor.Row)
	case "G":
		targetRow := buf.LineCount() - 1
		if v.Count > 0 {
			targetRow = clampInt(v.Count-1, 0, buf.LineCount()-1)
		}
		startRow := buf.cursor.Row
		endRow := targetRow
		if startRow > endRow {
			startRow, endRow = endRow, startRow
		}
		return v.execOperatorLinewise(buf, startRow, endRow)
	case "g":
		v.KeyBuffer = "g"
		return nil
	case "f", "F", "t", "T":
		v.KeyBuffer = key
		return nil
	default:
		// Unknown motion, cancel
		v.PendingOp = OpNone
		v.Count = 0
	}
	return nil
}

func (v *VimState) handleKeySequence(key string, buf *Buffer) tea.Cmd {
	seq := v.KeyBuffer + key
	v.KeyBuffer = ""

	count := v.getCount()

	switch seq {
	case "gg":
		if v.PendingOp != OpNone {
			targetRow := 0
			if v.Count > 0 {
				targetRow = clampInt(v.Count-1, 0, buf.LineCount()-1)
			}
			startRow := buf.cursor.Row
			endRow := targetRow
			if startRow > endRow {
				startRow, endRow = endRow, startRow
			}
			return v.execOperatorLinewise(buf, startRow, endRow)
		}
		if v.Count > 0 {
			buf.cursor.Row = clampInt(v.Count-1, 0, buf.LineCount()-1)
		} else {
			buf.cursor.Row = 0
		}
		buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
	default:
		if len(seq) == 2 {
			prefix := seq[0]
			ch := seq[1]
			switch prefix {
			case 'f':
				if v.PendingOp != OpNone {
					pos := buf.FindCharForward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						return v.execOperatorMotion(buf, func() Position {
							return Position{buf.cursor.Row, pos + 1} // inclusive
						})
					}
					v.PendingOp = OpNone
					v.Count = 0
					return nil
				}
				for i := 0; i < count; i++ {
					pos := buf.FindCharForward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						buf.cursor.Col = pos
					}
				}
				v.LastFindCh = ch
				v.LastFindDir = 1
				v.LastFindTil = false
			case 'F':
				if v.PendingOp != OpNone {
					pos := buf.FindCharBackward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						return v.execOperatorMotion(buf, func() Position {
							return Position{buf.cursor.Row, pos}
						})
					}
					v.PendingOp = OpNone
					v.Count = 0
					return nil
				}
				for i := 0; i < count; i++ {
					pos := buf.FindCharBackward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						buf.cursor.Col = pos
					}
				}
				v.LastFindCh = ch
				v.LastFindDir = -1
				v.LastFindTil = false
			case 't':
				if v.PendingOp != OpNone {
					pos := buf.FindCharForward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						return v.execOperatorMotion(buf, func() Position {
							return Position{buf.cursor.Row, pos} // til = up to but not including
						})
					}
					v.PendingOp = OpNone
					v.Count = 0
					return nil
				}
				for i := 0; i < count; i++ {
					pos := buf.FindCharForward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos > 0 {
						buf.cursor.Col = pos - 1
					}
				}
				v.LastFindCh = ch
				v.LastFindDir = 1
				v.LastFindTil = true
			case 'T':
				if v.PendingOp != OpNone {
					pos := buf.FindCharBackward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						return v.execOperatorMotion(buf, func() Position {
							return Position{buf.cursor.Row, pos + 1}
						})
					}
					v.PendingOp = OpNone
					v.Count = 0
					return nil
				}
				for i := 0; i < count; i++ {
					pos := buf.FindCharBackward(buf.cursor.Row, buf.cursor.Col, ch)
					if pos >= 0 {
						buf.cursor.Col = pos + 1
					}
				}
				v.LastFindCh = ch
				v.LastFindDir = -1
				v.LastFindTil = true
			}
		}
	}

	v.Count = 0
	return nil
}

func (v *VimState) execOperatorMotion(buf *Buffer, motionFn func() Position) tea.Cmd {
	start := buf.cursor
	end := motionFn()
	op := v.PendingOp
	v.PendingOp = OpNone
	v.Count = 0

	if start.Row == end.Row && start.Col == end.Col {
		return nil
	}

	// Normalize: start before end
	if start.Row > end.Row || (start.Row == end.Row && start.Col > end.Col) {
		start, end = end, start
	}

	switch op {
	case OpDelete:
		buf.SaveSnapshot()
		deleted := buf.DeleteRange(start, end)
		v.Registers['"'] = RegisterEntry{Text: deleted}
		buf.cursor = start
		// Clamp cursor
		lineLen := len(buf.Line(buf.cursor.Row))
		if lineLen > 0 && buf.cursor.Col >= lineLen {
			buf.cursor.Col = lineLen - 1
		}
	case OpChange:
		buf.SaveSnapshot()
		deleted := buf.DeleteRange(start, end)
		v.Registers['"'] = RegisterEntry{Text: deleted}
		buf.cursor = start
		v.Mode = ModeInsert
	case OpYank:
		// Don't modify buffer
		var sb strings.Builder
		if start.Row == end.Row {
			line := buf.Line(start.Row)
			sc := clampInt(start.Col, 0, len(line))
			ec := clampInt(end.Col, 0, len(line))
			sb.WriteString(line[sc:ec])
		} else {
			sb.WriteString(buf.Line(start.Row)[start.Col:])
			for r := start.Row + 1; r < end.Row; r++ {
				sb.WriteByte('\n')
				sb.WriteString(buf.Line(r))
			}
			sb.WriteByte('\n')
			ec := clampInt(end.Col, 0, len(buf.Line(end.Row)))
			sb.WriteString(buf.Line(end.Row)[:ec])
		}
		v.Registers['"'] = RegisterEntry{Text: sb.String()}
		v.StatusMsg = "yanked"
	}
	return nil
}

func (v *VimState) execOperatorLinewise(buf *Buffer, startRow, endRow int) tea.Cmd {
	op := v.PendingOp
	v.PendingOp = OpNone
	v.Count = 0

	startRow = clampInt(startRow, 0, buf.LineCount()-1)
	endRow = clampInt(endRow, 0, buf.LineCount()-1)
	if startRow > endRow {
		startRow, endRow = endRow, startRow
	}

	switch op {
	case OpDelete:
		buf.SaveSnapshot()
		deleted := buf.DeleteLines(startRow, endRow)
		v.Registers['"'] = RegisterEntry{Text: strings.Join(deleted, "\n"), IsLine: true}
		buf.cursor.Row = clampInt(startRow, 0, buf.LineCount()-1)
		buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
	case OpChange:
		buf.SaveSnapshot()
		deleted := buf.DeleteLines(startRow, endRow)
		v.Registers['"'] = RegisterEntry{Text: strings.Join(deleted, "\n"), IsLine: true}
		// Insert empty line for editing
		if startRow >= buf.LineCount() {
			buf.InsertLineAfter(buf.LineCount()-1, "")
			buf.cursor.Row = buf.LineCount() - 1
		} else {
			buf.InsertLineBefore(startRow, "")
			buf.cursor.Row = startRow
		}
		buf.cursor.Col = 0
		v.Mode = ModeInsert
	case OpYank:
		lines := make([]string, 0, endRow-startRow+1)
		for r := startRow; r <= endRow; r++ {
			lines = append(lines, buf.Line(r))
		}
		v.Registers['"'] = RegisterEntry{Text: strings.Join(lines, "\n"), IsLine: true}
		v.StatusMsg = "yanked " + strconv.Itoa(len(lines)) + " line(s)"
	}
	return nil
}

func (v *VimState) enterInsert(buf *Buffer) {
	buf.SaveSnapshot()
	v.Mode = ModeInsert
}

func (v *VimState) handleInsert(msg tea.KeyMsg, buf *Buffer) tea.Cmd {
	switch msg.String() {
	case "esc":
		v.Mode = ModeNormal
		// Move cursor back one if possible (vim behavior)
		if buf.cursor.Col > 0 {
			buf.cursor.Col--
		}
		return nil
	case "backspace":
		buf.Backspace()
	case "enter":
		buf.InsertNewline()
	case "tab":
		buf.InsertChar('\t')
	case "left":
		if buf.cursor.Col > 0 {
			buf.cursor.Col--
		}
	case "right":
		if buf.cursor.Col < len(buf.Line(buf.cursor.Row)) {
			buf.cursor.Col++
		}
	case "up":
		if buf.cursor.Row > 0 {
			buf.cursor.Row--
			buf.cursor.Col = clampInt(buf.cursor.Col, 0, len(buf.Line(buf.cursor.Row)))
		}
	case "down":
		if buf.cursor.Row < buf.LineCount()-1 {
			buf.cursor.Row++
			buf.cursor.Col = clampInt(buf.cursor.Col, 0, len(buf.Line(buf.cursor.Row)))
		}
	case "ctrl+w":
		// Delete word backward
		if buf.cursor.Col > 0 {
			line := buf.Line(buf.cursor.Row)
			col := buf.cursor.Col
			// Skip spaces
			for col > 0 && unicode.IsSpace(rune(line[col-1])) {
				col--
			}
			// Skip word
			for col > 0 && !unicode.IsSpace(rune(line[col-1])) {
				col--
			}
			buf.lines[buf.cursor.Row] = line[:col] + line[buf.cursor.Col:]
			buf.cursor.Col = col
			buf.dirty = true
		}
	case "ctrl+u":
		// Delete to beginning of line
		line := buf.Line(buf.cursor.Row)
		buf.lines[buf.cursor.Row] = line[buf.cursor.Col:]
		buf.cursor.Col = 0
		buf.dirty = true
	default:
		// Regular character input
		if len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				buf.InsertChar(r)
			}
		}
	}
	return nil
}

func (v *VimState) handleVisual(msg tea.KeyMsg, buf *Buffer) tea.Cmd {
	key := msg.String()

	switch key {
	case "esc":
		v.Mode = ModeNormal
		return nil
	case "v":
		if v.Mode == ModeVisual {
			v.Mode = ModeNormal
			return nil
		}
		v.Mode = ModeVisual
		v.VisualStart = buf.cursor
	case "V":
		if v.Mode == ModeVisualLine {
			v.Mode = ModeNormal
			return nil
		}
		v.Mode = ModeVisualLine
		v.VisualStart = buf.cursor

	// Movement (same as normal mode)
	case "h", "left":
		if buf.cursor.Col > 0 {
			buf.cursor.Col--
		}
	case "l", "right":
		maxCol := len(buf.Line(buf.cursor.Row)) - 1
		if maxCol < 0 {
			maxCol = 0
		}
		if buf.cursor.Col < maxCol {
			buf.cursor.Col++
		}
	case "j", "down":
		if buf.cursor.Row < buf.LineCount()-1 {
			buf.cursor.Row++
		}
		buf.cursor.Col = clampInt(buf.cursor.Col, 0, max(0, len(buf.Line(buf.cursor.Row))-1))
	case "k", "up":
		if buf.cursor.Row > 0 {
			buf.cursor.Row--
		}
		buf.cursor.Col = clampInt(buf.cursor.Col, 0, max(0, len(buf.Line(buf.cursor.Row))-1))
	case "w":
		p := buf.WordForwardPosVim(buf.cursor.Row, buf.cursor.Col)
		buf.cursor = p
	case "b":
		p := buf.WordBackwardPosVim(buf.cursor.Row, buf.cursor.Col)
		buf.cursor = p
	case "e":
		p := buf.WordEndPosVim(buf.cursor.Row, buf.cursor.Col)
		buf.cursor = p
	case "0":
		buf.cursor.Col = 0
	case "$":
		lineLen := len(buf.Line(buf.cursor.Row))
		if lineLen > 0 {
			buf.cursor.Col = lineLen - 1
		}
	case "G":
		buf.cursor.Row = buf.LineCount() - 1
		buf.cursor.Col = clampInt(buf.cursor.Col, 0, max(0, len(buf.Line(buf.cursor.Row))-1))
	case "g":
		v.KeyBuffer = "g"
		return nil

	// Operations on selection
	case "d", "x":
		v.deleteVisualSelection(buf)
		v.Mode = ModeNormal
	case "c":
		v.deleteVisualSelection(buf)
		v.Mode = ModeInsert
	case "y":
		v.yankVisualSelection(buf)
		v.Mode = ModeNormal
	}

	// Handle gg in visual mode
	if v.KeyBuffer == "g" && key != "g" {
		v.KeyBuffer = ""
	}
	if v.KeyBuffer == "" && key == "g" {
		// Already set above
	}

	return nil
}

func (v *VimState) deleteVisualSelection(buf *Buffer) {
	buf.SaveSnapshot()
	start, end := v.visualRange(buf)

	if v.Mode == ModeVisualLine {
		deleted := buf.DeleteLines(start.Row, end.Row)
		v.Registers['"'] = RegisterEntry{Text: strings.Join(deleted, "\n"), IsLine: true}
		buf.cursor.Row = clampInt(start.Row, 0, buf.LineCount()-1)
		buf.cursor.Col = buf.FirstNonBlank(buf.cursor.Row)
	} else {
		// Character-wise: end is inclusive
		endPos := Position{end.Row, end.Col + 1}
		if endPos.Col > len(buf.Line(endPos.Row)) {
			endPos.Col = len(buf.Line(endPos.Row))
		}
		deleted := buf.DeleteRange(start, endPos)
		v.Registers['"'] = RegisterEntry{Text: deleted}
		buf.cursor = start
		lineLen := len(buf.Line(buf.cursor.Row))
		if lineLen > 0 && buf.cursor.Col >= lineLen {
			buf.cursor.Col = lineLen - 1
		}
	}
}

func (v *VimState) yankVisualSelection(buf *Buffer) {
	start, end := v.visualRange(buf)

	if v.Mode == ModeVisualLine {
		lines := make([]string, 0, end.Row-start.Row+1)
		for r := start.Row; r <= end.Row; r++ {
			lines = append(lines, buf.Line(r))
		}
		v.Registers['"'] = RegisterEntry{Text: strings.Join(lines, "\n"), IsLine: true}
	} else {
		var sb strings.Builder
		if start.Row == end.Row {
			line := buf.Line(start.Row)
			sc := clampInt(start.Col, 0, len(line))
			ec := clampInt(end.Col+1, 0, len(line))
			sb.WriteString(line[sc:ec])
		} else {
			sb.WriteString(buf.Line(start.Row)[start.Col:])
			for r := start.Row + 1; r < end.Row; r++ {
				sb.WriteByte('\n')
				sb.WriteString(buf.Line(r))
			}
			sb.WriteByte('\n')
			ec := clampInt(end.Col+1, 0, len(buf.Line(end.Row)))
			sb.WriteString(buf.Line(end.Row)[:ec])
		}
		v.Registers['"'] = RegisterEntry{Text: sb.String()}
	}
	buf.cursor = start
	v.StatusMsg = "yanked"
}

// visualRange returns the normalized start and end positions of the visual selection.
func (v *VimState) visualRange(buf *Buffer) (Position, Position) {
	start := v.VisualStart
	end := buf.cursor
	if start.Row > end.Row || (start.Row == end.Row && start.Col > end.Col) {
		start, end = end, start
	}
	if v.Mode == ModeVisualLine {
		start.Col = 0
		end.Col = len(buf.Line(end.Row))
	}
	return start, end
}

func (v *VimState) handleCommand(msg tea.KeyMsg, buf *Buffer) tea.Cmd {
	key := msg.String()

	switch key {
	case "esc":
		v.Mode = ModeNormal
		v.CommandLine = ""
		return nil
	case "enter":
		cmd := v.CommandLine
		v.Mode = ModeNormal
		v.CommandLine = ""
		return v.executeCommand(cmd, buf)
	case "backspace":
		if len(v.CommandLine) > 0 {
			v.CommandLine = v.CommandLine[:len(v.CommandLine)-1]
		} else {
			v.Mode = ModeNormal
		}
	default:
		if len(msg.Runes) > 0 {
			for _, r := range msg.Runes {
				v.CommandLine += string(r)
			}
		}
	}
	return nil
}

func (v *VimState) handleReplace(msg tea.KeyMsg, buf *Buffer) tea.Cmd {
	v.Mode = ModeNormal
	key := msg.String()
	if key == "esc" {
		return nil
	}
	if len(msg.Runes) > 0 {
		buf.SaveSnapshot()
		buf.ReplaceCharAtCursor(msg.Runes[0])
	}
	return nil
}

func (v *VimState) executeCommand(cmd string, buf *Buffer) tea.Cmd {
	cmd = strings.TrimSpace(cmd)

	// Check if this was a search command (entered via / or ?)
	if v.CommandPrefix == '/' || v.CommandPrefix == '?' {
		v.SearchQuery = cmd
		dir := 1
		if v.CommandPrefix == '?' {
			dir = -1
		}
		v.SearchDir = dir
		v.searchNext(buf, dir)
		return nil
	}

	return executeExCommand(cmd, buf, v)
}

func (v *VimState) searchNext(buf *Buffer, dir int) {
	if v.SearchQuery == "" {
		return
	}
	query := strings.ToLower(v.SearchQuery)
	startRow := buf.cursor.Row
	startCol := buf.cursor.Col

	if dir > 0 {
		// Forward search
		for i := 0; i < buf.LineCount(); i++ {
			row := (startRow + i) % buf.LineCount()
			line := strings.ToLower(buf.Line(row))
			searchFrom := 0
			if row == startRow {
				searchFrom = startCol + 1
			}
			idx := strings.Index(line[searchFrom:], query)
			if idx >= 0 {
				buf.cursor.Row = row
				buf.cursor.Col = searchFrom + idx
				return
			}
		}
	} else {
		// Backward search
		for i := 0; i < buf.LineCount(); i++ {
			row := (startRow - i + buf.LineCount()) % buf.LineCount()
			line := strings.ToLower(buf.Line(row))
			searchTo := len(line)
			if row == startRow {
				searchTo = startCol
			}
			idx := strings.LastIndex(line[:searchTo], query)
			if idx >= 0 {
				buf.cursor.Row = row
				buf.cursor.Col = idx
				return
			}
		}
	}
	v.StatusMsg = "Pattern not found: " + v.SearchQuery
}
