package vim

// Action represents a command returned by a ModeHandler.
// The editor interprets and applies actions — handlers never mutate state directly.
type Action interface {
	actionTag() // unexported method prevents external implementations
}

// NoOpAction indicates no state change needed.
type NoOpAction struct{}

// MoveCursorAction moves the cursor to an absolute or relative position.
type MoveCursorAction struct {
	Line     int  // Target line (absolute if Relative=false, delta if Relative=true)
	Col      int  // Target column (absolute if Relative=false, delta if Relative=true)
	Relative bool // If true, Line/Col are deltas from current position
}

// ScrollAction scrolls the viewport by half-page increments.
type ScrollAction struct {
	Direction  int  // Positive = down, negative = up (sign only; editor calculates magnitude)
	MoveCursor bool // If true, also move cursor by same amount
}

// DocumentPositionAction jumps to a document-level position.
type DocumentPositionAction struct {
	Position string // "top" or "bottom"
}

// WordMotionAction moves the cursor by word boundaries.
type WordMotionAction struct {
	Forward bool // true = next word (w), false = previous word (b)
}

// QuitAction signals the editor to exit.
type QuitAction struct{}

// ChangeModeAction switches the editor to a different vim mode.
type ChangeModeAction struct {
	Mode    Mode   // Target mode
	Variant string // Entry variant: "i", "a", "o", "O" (for insert mode entry)
}

// InsertCharAction inserts a character at the cursor position.
type InsertCharAction struct {
	Char rune
}

// BackspaceAction deletes the character before the cursor.
type BackspaceAction struct{}

// DeleteCharAction deletes the character after the cursor.
type DeleteCharAction struct{}

// InsertNewlineAction inserts a newline at the cursor position.
type InsertNewlineAction struct{}

// InsertTabAction inserts a tab/indentation at the cursor position.
type InsertTabAction struct{}

// ExecuteCommandAction signals the editor to execute the accumulated command buffer.
type ExecuteCommandAction struct{}

// UndoAction signals the editor to undo the last edit in the active block.
type UndoAction struct{}

// RedoAction signals the editor to redo the last undone edit in the active block.
type RedoAction struct{}

// InsertPairAction inserts an opening+closing pair at the cursor, leaving the cursor between them.
type InsertPairAction struct {
	Opening string
	Closing string
}

// SkipClosingAction moves the cursor right past an existing closing character rather than inserting a duplicate.
type SkipClosingAction struct {
	Count int // number of runes to skip (1 for single-char pairs, 2 for ** and __)
}

// MultiAction groups multiple actions to be applied in sequence.
type MultiAction struct {
	Actions []Action
}

func (NoOpAction) actionTag()             {}
func (MoveCursorAction) actionTag()       {}
func (ScrollAction) actionTag()           {}
func (DocumentPositionAction) actionTag() {}
func (WordMotionAction) actionTag()       {}
func (QuitAction) actionTag()             {}
func (ChangeModeAction) actionTag()       {}
func (InsertCharAction) actionTag()       {}
func (BackspaceAction) actionTag()        {}
func (DeleteCharAction) actionTag()       {}
func (InsertNewlineAction) actionTag()    {}
func (InsertTabAction) actionTag()        {}
func (ExecuteCommandAction) actionTag()   {}
func (UndoAction) actionTag()             {}
func (RedoAction) actionTag()             {}
func (InsertPairAction) actionTag()       {}
func (SkipClosingAction) actionTag()      {}
func (MultiAction) actionTag()            {}
