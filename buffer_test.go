package main

import (
	"testing"
)

func TestNewBuffer(t *testing.T) {
	b := NewBuffer()
	if b.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", b.LineCount())
	}
	if b.Line(0) != "" {
		t.Errorf("expected empty line, got %q", b.Line(0))
	}
}

func TestInsertChar(t *testing.T) {
	b := NewBuffer()
	b.InsertChar('H')
	b.InsertChar('i')
	if b.Line(0) != "Hi" {
		t.Errorf("expected 'Hi', got %q", b.Line(0))
	}
	if b.cursor.Col != 2 {
		t.Errorf("expected col 2, got %d", b.cursor.Col)
	}
}

func TestInsertNewline(t *testing.T) {
	b := NewBuffer()
	b.InsertChar('A')
	b.InsertChar('B')
	b.cursor.Col = 1
	b.InsertNewline()
	if b.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", b.LineCount())
	}
	if b.Line(0) != "A" {
		t.Errorf("expected 'A', got %q", b.Line(0))
	}
	if b.Line(1) != "B" {
		t.Errorf("expected 'B', got %q", b.Line(1))
	}
}

func TestBackspace(t *testing.T) {
	b := NewBuffer()
	b.InsertChar('A')
	b.InsertChar('B')
	b.InsertChar('C')
	b.Backspace()
	if b.Line(0) != "AB" {
		t.Errorf("expected 'AB', got %q", b.Line(0))
	}

	// Backspace at beginning of line joins with previous
	b.InsertNewline()
	b.InsertChar('D')
	b.cursor.Col = 0
	b.Backspace()
	if b.LineCount() != 1 {
		t.Errorf("expected 1 line after join, got %d", b.LineCount())
	}
	if b.Line(0) != "ABD" {
		t.Errorf("expected 'ABD', got %q", b.Line(0))
	}
}

func TestDeleteCharAtCursor(t *testing.T) {
	b := NewBuffer()
	b.InsertChar('A')
	b.InsertChar('B')
	b.InsertChar('C')
	b.cursor.Col = 1
	b.DeleteCharAtCursor()
	if b.Line(0) != "AC" {
		t.Errorf("expected 'AC', got %q", b.Line(0))
	}
}

func TestDeleteLine(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"line1", "line2", "line3"}
	deleted := b.DeleteLine(1)
	if deleted != "line2" {
		t.Errorf("expected 'line2', got %q", deleted)
	}
	if b.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", b.LineCount())
	}
}

func TestDeleteLines(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"a", "b", "c", "d", "e"}
	deleted := b.DeleteLines(1, 3)
	if len(deleted) != 3 {
		t.Errorf("expected 3 deleted lines, got %d", len(deleted))
	}
	if b.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", b.LineCount())
	}
	if b.Line(0) != "a" || b.Line(1) != "e" {
		t.Errorf("expected [a, e], got [%s, %s]", b.Line(0), b.Line(1))
	}
}

func TestUndoRedo(t *testing.T) {
	b := NewBuffer()
	b.InsertChar('A')
	b.SaveSnapshot()
	b.InsertChar('B')

	if b.Line(0) != "AB" {
		t.Errorf("expected 'AB', got %q", b.Line(0))
	}

	b.Undo()
	if b.Line(0) != "A" {
		t.Errorf("after undo expected 'A', got %q", b.Line(0))
	}

	b.Redo()
	// Redo restores to the state when undo was called, which was "AB"
	// Actually: SaveSnapshot saved state "A" (col=1). Then we typed "B" -> "AB".
	// Undo pops to snapshot "A". Redo goes back to "AB" state.
	if b.Line(0) != "AB" {
		t.Errorf("after redo expected 'AB', got %q", b.Line(0))
	}
}

func TestWordForwardPosVim(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"hello world foo"}
	pos := b.WordForwardPosVim(0, 0)
	if pos.Col != 6 {
		t.Errorf("expected col 6, got %d", pos.Col)
	}
	pos = b.WordForwardPosVim(0, 6)
	if pos.Col != 12 {
		t.Errorf("expected col 12, got %d", pos.Col)
	}
}

func TestWordBackwardPosVim(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"hello world foo"}
	pos := b.WordBackwardPosVim(0, 12)
	if pos.Col != 6 {
		t.Errorf("expected col 6, got %d", pos.Col)
	}
	pos = b.WordBackwardPosVim(0, 6)
	if pos.Col != 0 {
		t.Errorf("expected col 0, got %d", pos.Col)
	}
}

func TestWordEndPosVim(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"hello world"}
	pos := b.WordEndPosVim(0, 0)
	if pos.Col != 4 {
		t.Errorf("expected col 4, got %d", pos.Col)
	}
}

func TestDeleteRange(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"hello world"}
	deleted := b.DeleteRange(Position{0, 0}, Position{0, 5})
	if deleted != "hello" {
		t.Errorf("expected 'hello', got %q", deleted)
	}
	if b.Line(0) != " world" {
		t.Errorf("expected ' world', got %q", b.Line(0))
	}
}

func TestJoinLine(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"hello", "world"}
	b.JoinLine(0)
	if b.LineCount() != 1 {
		t.Errorf("expected 1 line, got %d", b.LineCount())
	}
	if b.Line(0) != "hello world" {
		t.Errorf("expected 'hello world', got %q", b.Line(0))
	}
}

func TestFindChar(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"hello world"}
	pos := b.FindCharForward(0, 0, 'o')
	if pos != 4 {
		t.Errorf("expected 4, got %d", pos)
	}
	pos = b.FindCharBackward(0, 10, 'o')
	if pos != 7 {
		t.Errorf("expected 7, got %d", pos)
	}
}

func TestFirstNonBlank(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"  hello"}
	col := b.FirstNonBlank(0)
	if col != 2 {
		t.Errorf("expected 2, got %d", col)
	}
}

func TestInsertLineAfterBefore(t *testing.T) {
	b := NewBuffer()
	b.lines = []string{"a", "c"}
	b.InsertLineAfter(0, "b")
	if b.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", b.LineCount())
	}
	if b.Line(1) != "b" {
		t.Errorf("expected 'b' at line 1, got %q", b.Line(1))
	}

	b.InsertLineBefore(0, "z")
	if b.Line(0) != "z" {
		t.Errorf("expected 'z' at line 0, got %q", b.Line(0))
	}
}
