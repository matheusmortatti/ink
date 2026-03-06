package editor

import (
	"testing"
)

func TestUndoManager_Record_PushesToUndoStack(t *testing.T) {
	m := NewUndoManager(100)
	if m.CanUndo() {
		t.Fatal("expected CanUndo=false before any record")
	}
	m.Record("hello", 5, 0, 5, "insert")
	if !m.CanUndo() {
		t.Fatal("expected CanUndo=true after record")
	}
}

func TestUndoManager_Undo_RestoresPreviousState(t *testing.T) {
	m := NewUndoManager(100)
	m.Record("hello", 5, 0, 5, "insert")
	entry, ok := m.Undo("hello world", 11, 0, 11)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if entry.Content != "hello" {
		t.Errorf("got Content=%q, want %q", entry.Content, "hello")
	}
	if entry.CursorPos != 5 {
		t.Errorf("got CursorPos=%d, want 5", entry.CursorPos)
	}
}

func TestUndoManager_Redo_ReappliesUndoneState(t *testing.T) {
	m := NewUndoManager(100)
	m.Record("hello", 5, 0, 5, "insert")
	_, _ = m.Undo("hello world", 11, 0, 11)
	entry, ok := m.Redo("hello", 5, 0, 5)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if entry.Content != "hello world" {
		t.Errorf("got Content=%q, want %q", entry.Content, "hello world")
	}
}

func TestUndoManager_Undo_EmptyStack_ReturnsFalse(t *testing.T) {
	m := NewUndoManager(100)
	_, ok := m.Undo("content", 7, 0, 7)
	if ok {
		t.Fatal("expected ok=false on empty undo stack")
	}
}

func TestUndoManager_Redo_EmptyStack_ReturnsFalse(t *testing.T) {
	m := NewUndoManager(100)
	_, ok := m.Redo("content", 7, 0, 7)
	if ok {
		t.Fatal("expected ok=false on empty redo stack")
	}
}

func TestUndoManager_Record_ClearsRedoStack(t *testing.T) {
	m := NewUndoManager(100)
	m.Record("a", 1, 0, 1, "insert")
	_, _ = m.Undo("ab", 2, 0, 2)
	if !m.CanRedo() {
		t.Fatal("expected CanRedo=true after undo")
	}
	// Record new state — should clear redo stack
	m.Record("ac", 2, 0, 2, "insert")
	if m.CanRedo() {
		t.Fatal("expected CanRedo=false after new record")
	}
}

func TestUndoManager_Clear_ResetsBothStacks(t *testing.T) {
	m := NewUndoManager(100)
	m.Record("a", 1, 0, 1, "insert")
	_, _ = m.Undo("ab", 2, 0, 2) // populate redo
	m.Clear()
	if m.CanUndo() {
		t.Error("expected CanUndo=false after clear")
	}
	if m.CanRedo() {
		t.Error("expected CanRedo=false after clear")
	}
}

func TestUndoManager_MaxSize_EnforcesLimit(t *testing.T) {
	const maxSize = 5
	m := NewUndoManager(maxSize)
	for i := 0; i < maxSize+10; i++ {
		m.forceRecord(UndoEntry{Content: string(rune('a' + i)), CursorPos: i})
	}
	if len(m.undoStack) > maxSize {
		t.Errorf("undo stack size %d exceeds maxSize %d", len(m.undoStack), maxSize)
	}
}

func TestUndoManager_MultipleUndos_ReverseChronological(t *testing.T) {
	m := NewUndoManager(100)
	// Force-record three entries to bypass debounce
	m.forceRecord(UndoEntry{Content: "A", CursorPos: 1})
	m.forceRecord(UndoEntry{Content: "B", CursorPos: 1})
	m.forceRecord(UndoEntry{Content: "C", CursorPos: 1})

	// Undo should return C, then B, then A
	e1, _ := m.Undo("current", 0, 0, 0)
	if e1.Content != "C" {
		t.Errorf("first undo: got %q, want %q", e1.Content, "C")
	}
	e2, _ := m.Undo(e1.Content, 0, 0, 0)
	if e2.Content != "B" {
		t.Errorf("second undo: got %q, want %q", e2.Content, "B")
	}
	e3, _ := m.Undo(e2.Content, 0, 0, 0)
	if e3.Content != "A" {
		t.Errorf("third undo: got %q, want %q", e3.Content, "A")
	}
	_, ok := m.Undo(e3.Content, 0, 0, 0)
	if ok {
		t.Error("expected ok=false after exhausting stack")
	}
}
