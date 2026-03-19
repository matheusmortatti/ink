package vim

// pairRegistry maps a single opening rune to its closing string for simple pairs.
var pairRegistry = map[rune]string{
	'`': "`",
	'[': "]",
	'(': ")",
}

// AutoPairState tracks multi-keystroke state for double-char pairs (** and __).
type AutoPairState struct {
	pendingStar       bool
	pendingUnderscore bool
}

// NewAutoPairState creates a zero-value AutoPairState.
func NewAutoPairState() *AutoPairState {
	return &AutoPairState{}
}

// HandleChar processes a typed character and returns the resulting Action.
// nextChar is the character immediately after the cursor (0 if at end of buffer).
// nextNextChar is the character two positions after the cursor (0 if unavailable).
// consumed=true means the caller should use the returned Action; false means process normally.
// When consumed=false and action is non-nil, the action (a flush) must be applied first.
func (s *AutoPairState) HandleChar(char rune, nextChar rune, nextNextChar rune) (action Action, consumed bool) {
	// Handle double-char pair triggers (* and _).
	switch char {
	case '*':
		if s.pendingStar {
			s.pendingStar = false
			// Skip-over: if next two chars are **, move past them.
			if nextChar == '*' && nextNextChar == '*' {
				return SkipClosingAction{Count: 2}, true
			}
			// Otherwise emit the ** pair.
			return InsertPairAction{Opening: "**", Closing: "**"}, true
		}
		// First '*' — wait for next char.
		s.pendingStar = true
		return nil, true

	case '_':
		if s.pendingUnderscore {
			s.pendingUnderscore = false
			// Skip-over: if next two chars are __, move past them.
			if nextChar == '_' && nextNextChar == '_' {
				return SkipClosingAction{Count: 2}, true
			}
			// Otherwise emit the __ pair.
			return InsertPairAction{Opening: "__", Closing: "__"}, true
		}
		// First '_' — wait for next char.
		s.pendingUnderscore = true
		return nil, true
	}

	// Non-* / non-_ character: flush any pending buffered char first.
	// Return the flush action with consumed=false so the current char is processed again.
	if s.pendingStar {
		s.pendingStar = false
		return InsertCharAction{Char: '*'}, false
	}
	if s.pendingUnderscore {
		s.pendingUnderscore = false
		return InsertCharAction{Char: '_'}, false
	}

	// Skip-over closing char: if the typed char is a closing char and the next char
	// in the buffer already equals it, move the cursor past the existing char.
	// Must be checked BEFORE pair insertion to handle the ` case correctly.
	switch char {
	case '`', ']', ')':
		if nextChar == char {
			return SkipClosingAction{Count: 1}, true
		}
	}

	// Single-char pair insertion.
	if closing, ok := pairRegistry[char]; ok {
		return InsertPairAction{Opening: string(char), Closing: closing}, true
	}

	return nil, false
}

// Flush emits any buffered pending character as an InsertCharAction and clears state.
// Returns nil if no pending state exists.
// Must be called on Esc, Enter, or any event that terminates character input.
func (s *AutoPairState) Flush() Action {
	if s.pendingStar {
		s.pendingStar = false
		return InsertCharAction{Char: '*'}
	}
	if s.pendingUnderscore {
		s.pendingUnderscore = false
		return InsertCharAction{Char: '_'}
	}
	return nil
}
