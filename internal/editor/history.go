package editor

import "time"

// UndoEntry holds a gap buffer content snapshot and cursor position for undo/redo.
type UndoEntry struct {
	Content    string
	CursorPos  int
	CursorLine int
	CursorCol  int
}

// UndoManager manages per-block undo/redo stacks with keystroke grouping.
// It uses time-based debouncing and action-type boundaries to group related
// keystrokes into single undo entries, similar to vim's undo granularity.
type UndoManager struct {
	undoStack      []UndoEntry
	redoStack      []UndoEntry
	maxSize        int
	lastRecordTime time.Time
	lastActionType string
}

// NewUndoManager creates a new UndoManager with the given maximum stack size.
func NewUndoManager(maxSize int) *UndoManager {
	return &UndoManager{maxSize: maxSize}
}

// Record pushes the current state to the undo stack if grouping rules require a new entry.
// A new entry is created when:
//   - no previous entry exists
//   - >500ms has elapsed since the last recorded entry
//   - the action type changed (e.g., insert → delete)
//   - actionType is "newline" (always creates a boundary)
//
// Clears the redo stack on every call. Enforces maxSize cap.
func (m *UndoManager) Record(content string, cursorPos, cursorLine, cursorCol int, actionType string) {
	now := time.Now()
	shouldRecord := m.lastRecordTime.IsZero() ||
		now.Sub(m.lastRecordTime) > 500*time.Millisecond ||
		m.lastActionType != actionType ||
		actionType == "newline"

	// Always clear redo stack when a mutation occurs
	m.redoStack = m.redoStack[:0]

	if !shouldRecord {
		return
	}

	m.lastRecordTime = now
	m.lastActionType = actionType

	m.undoStack = append(m.undoStack, UndoEntry{
		Content:    content,
		CursorPos:  cursorPos,
		CursorLine: cursorLine,
		CursorCol:  cursorCol,
	})

	if len(m.undoStack) > m.maxSize {
		m.undoStack = m.undoStack[len(m.undoStack)-m.maxSize:]
	}
}

// forceRecord bypasses debounce grouping and pushes the entry directly.
// Used in tests to set up precise undo stack states.
func (m *UndoManager) forceRecord(entry UndoEntry) {
	m.redoStack = m.redoStack[:0]
	m.undoStack = append(m.undoStack, entry)
	if len(m.undoStack) > m.maxSize {
		m.undoStack = m.undoStack[len(m.undoStack)-m.maxSize:]
	}
}

// Undo saves the current state to the redo stack and pops from the undo stack.
// Returns the restored entry and true, or zero value and false if stack is empty.
func (m *UndoManager) Undo(currentContent string, currentCursorPos, currentCursorLine, currentCursorCol int) (UndoEntry, bool) {
	if len(m.undoStack) == 0 {
		return UndoEntry{}, false
	}
	m.redoStack = append(m.redoStack, UndoEntry{
		Content:    currentContent,
		CursorPos:  currentCursorPos,
		CursorLine: currentCursorLine,
		CursorCol:  currentCursorCol,
	})
	n := len(m.undoStack)
	entry := m.undoStack[n-1]
	m.undoStack = m.undoStack[:n-1]
	return entry, true
}

// Redo saves the current state to the undo stack and pops from the redo stack.
// Returns the restored entry and true, or zero value and false if stack is empty.
func (m *UndoManager) Redo(currentContent string, currentCursorPos, currentCursorLine, currentCursorCol int) (UndoEntry, bool) {
	if len(m.redoStack) == 0 {
		return UndoEntry{}, false
	}
	m.undoStack = append(m.undoStack, UndoEntry{
		Content:    currentContent,
		CursorPos:  currentCursorPos,
		CursorLine: currentCursorLine,
		CursorCol:  currentCursorCol,
	})
	n := len(m.redoStack)
	entry := m.redoStack[n-1]
	m.redoStack = m.redoStack[:n-1]
	return entry, true
}

// Clear resets both stacks and debounce state. Called when a block is committed.
func (m *UndoManager) Clear() {
	m.undoStack = m.undoStack[:0]
	m.redoStack = m.redoStack[:0]
	m.lastRecordTime = time.Time{}
	m.lastActionType = ""
}

// CanUndo returns true if there are entries available to undo.
func (m *UndoManager) CanUndo() bool {
	return len(m.undoStack) > 0
}

// CanRedo returns true if there are entries available to redo.
func (m *UndoManager) CanRedo() bool {
	return len(m.redoStack) > 0
}
