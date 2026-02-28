package vim

// CommandHandler processes key input in command mode.
// It is stateless — the editor owns the command buffer (commandBuf).
type CommandHandler struct{}

// NewCommandHandler creates a CommandHandler for command mode key processing.
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{}
}

// HandleKey processes a key string and returns the resulting action.
func (h *CommandHandler) HandleKey(key string) Action {
	switch key {
	case "esc":
		return ChangeModeAction{Mode: Normal}
	case "backspace":
		return BackspaceAction{}
	case "enter":
		return ExecuteCommandAction{}
	default:
		runes := []rune(key)
		if len(runes) == 1 && runes[0] >= 32 {
			return InsertCharAction{Char: runes[0]}
		}
		return NoOpAction{}
	}
}

// Mode returns Command.
func (h *CommandHandler) Mode() Mode {
	return Command
}
