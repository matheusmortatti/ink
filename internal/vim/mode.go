package vim

// Mode represents the current vim editing mode.
type Mode int

const (
	Normal Mode = iota
	Insert
	Visual
	Command
)

// String returns the display name of the mode.
func (m Mode) String() string {
	switch m {
	case Normal:
		return "NORMAL"
	case Insert:
		return "INSERT"
	case Visual:
		return "VISUAL"
	case Command:
		return "COMMAND"
	default:
		return "UNKNOWN"
	}
}

// ModeHandler processes key input and returns an Action.
// Each mode (Normal, Insert, Visual, Command) implements this interface.
type ModeHandler interface {
	// HandleKey processes a key string and returns the resulting action.
	HandleKey(key string) Action

	// Mode returns the current mode this handler represents.
	Mode() Mode
}
