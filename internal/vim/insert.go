package vim

// InsertHandler processes key input in insert mode.
type InsertHandler struct{}

// NewInsertHandler creates an InsertHandler for insert mode key processing.
func NewInsertHandler() *InsertHandler {
	return &InsertHandler{}
}

// HandleKey processes a key string and returns the resulting action.
func (h *InsertHandler) HandleKey(key string) Action {
	switch key {
	case "esc":
		return ChangeModeAction{Mode: Normal}
	case "backspace":
		return BackspaceAction{}
	case "delete":
		return DeleteCharAction{}
	case "enter":
		return InsertNewlineAction{}
	case "tab":
		return InsertTabAction{}
	case "space":
		return InsertCharAction{Char: ' '}
	case "left":
		return MoveCursorAction{Col: -1, Relative: true}
	case "right":
		return MoveCursorAction{Col: 1, Relative: true}
	case "up":
		return MoveCursorAction{Line: -1, Relative: true}
	case "down":
		return MoveCursorAction{Line: 1, Relative: true}
	default:
		runes := []rune(key)
		if len(runes) == 1 && runes[0] >= 32 {
			return InsertCharAction{Char: runes[0]}
		}
		return NoOpAction{}
	}
}

// Mode returns Insert.
func (h *InsertHandler) Mode() Mode {
	return Insert
}
