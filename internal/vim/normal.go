package vim

// NormalHandler processes key input in normal mode.
type NormalHandler struct {
	pending rune // For multi-key sequences (e.g., 'g' waiting for second key)
}

// NewNormalHandler creates a NormalHandler for normal mode key processing.
func NewNormalHandler() *NormalHandler {
	return &NormalHandler{}
}

// HandleKey processes a key string and returns the resulting action.
func (n *NormalHandler) HandleKey(key string) Action {
	// Handle second key of multi-key sequence
	if n.pending == 'g' {
		n.pending = 0
		if key == "g" {
			return DocumentPositionAction{Position: "top"}
		}
		// Not a valid sequence — process the second key as standalone
		return n.handleSingleKey(key)
	}

	return n.handleSingleKey(key)
}

// Mode returns Normal.
func (n *NormalHandler) Mode() Mode {
	return Normal
}

func (n *NormalHandler) handleSingleKey(key string) Action {
	switch key {
	case "j":
		return MoveCursorAction{Line: 1, Relative: true}
	case "k":
		return MoveCursorAction{Line: -1, Relative: true}
	case "h":
		return MoveCursorAction{Col: -1, Relative: true}
	case "l":
		return MoveCursorAction{Col: 1, Relative: true}
	case "w":
		return WordMotionAction{Forward: true}
	case "b":
		return WordMotionAction{Forward: false}
	case "G", "shift+G":
		return DocumentPositionAction{Position: "bottom"}
	case "g":
		n.pending = 'g'
		return NoOpAction{}
	case "ctrl+d":
		return ScrollAction{Direction: 1, MoveCursor: true}
	case "ctrl+u":
		return ScrollAction{Direction: -1, MoveCursor: true}
	case "ctrl+c":
		return QuitAction{}
	default:
		return NoOpAction{}
	}
}
