package vim

// InsertHandler processes key input in insert mode.
type InsertHandler struct {
	autoPair     *AutoPairState
	nextChar     rune // char immediately after cursor, set by editor before each HandleKey call
	nextNextChar rune // char two positions after cursor, for double-pair skip-over detection
}

// NewInsertHandler creates an InsertHandler for insert mode key processing.
func NewInsertHandler() *InsertHandler {
	return &InsertHandler{
		autoPair: NewAutoPairState(),
	}
}

// SetNextChars sets the two characters after the cursor so that HandleChar
// can perform skip-over detection for both single and double pairs.
// Must be called by the editor before HandleKey.
func (h *InsertHandler) SetNextChars(next, nextNext rune) {
	h.nextChar = next
	h.nextNextChar = nextNext
}

// HandleKey processes a key string and returns the resulting action.
func (h *InsertHandler) HandleKey(key string) Action {
	switch key {
	case "esc":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, ChangeModeAction{Mode: Normal}}}
		}
		return ChangeModeAction{Mode: Normal}

	case "backspace":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, BackspaceAction{}}}
		}
		return BackspaceAction{}

	case "delete":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, DeleteCharAction{}}}
		}
		return DeleteCharAction{}

	case "enter":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, InsertNewlineAction{}}}
		}
		return InsertNewlineAction{}

	case "tab":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, InsertTabAction{}}}
		}
		return InsertTabAction{}

	case "space":
		// Space is a printable char — route through auto-pair (no pair for space, so it passes through).
		apAction, consumed := h.autoPair.HandleChar(' ', h.nextChar, h.nextNextChar)
		if consumed {
			if apAction != nil {
				return apAction
			}
			return InsertCharAction{Char: ' '}
		}
		// consumed=false with possible flush action
		if apAction != nil {
			return MultiAction{Actions: []Action{apAction, InsertCharAction{Char: ' '}}}
		}
		return InsertCharAction{Char: ' '}

	case "left":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, MoveCursorAction{Col: -1, Relative: true}}}
		}
		return MoveCursorAction{Col: -1, Relative: true}
	case "right":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, MoveCursorAction{Col: 1, Relative: true}}}
		}
		return MoveCursorAction{Col: 1, Relative: true}
	case "up":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, MoveCursorAction{Line: -1, Relative: true}}}
		}
		return MoveCursorAction{Line: -1, Relative: true}
	case "down":
		if flush := h.autoPair.Flush(); flush != nil {
			return MultiAction{Actions: []Action{flush, MoveCursorAction{Line: 1, Relative: true}}}
		}
		return MoveCursorAction{Line: 1, Relative: true}

	default:
		runes := []rune(key)
		if len(runes) == 1 && runes[0] >= 32 {
			char := runes[0]
			apAction, consumed := h.autoPair.HandleChar(char, h.nextChar, h.nextNextChar)
			if consumed {
				if apAction != nil {
					return apAction
				}
				// consumed but no action (pending state set, e.g. first `*`) — no-op for now.
				return NoOpAction{}
			}
			// not consumed: apAction may be a flush, current char goes through normally.
			if apAction != nil {
				return MultiAction{Actions: []Action{apAction, InsertCharAction{Char: char}}}
			}
			return InsertCharAction{Char: char}
		}
		return NoOpAction{}
	}
}

// Mode returns Insert.
func (h *InsertHandler) Mode() Mode {
	return Insert
}
