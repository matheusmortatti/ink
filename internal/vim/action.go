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

func (NoOpAction) actionTag()             {}
func (MoveCursorAction) actionTag()       {}
func (ScrollAction) actionTag()           {}
func (DocumentPositionAction) actionTag() {}
func (WordMotionAction) actionTag()       {}
func (QuitAction) actionTag()             {}
