package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/matheusmortatti/ink/internal/render"
)

// StatusBar displays the editor mode, word count, and character count
// centered at the bottom of the terminal. It is a plain component struct,
// not a Bubbletea model — the editor owns and calls it directly.
type StatusBar struct {
	width       int
	modeLabel   string
	wordCount   int
	charCount   int
	dimStyle    lipgloss.Style
	commandMode bool
	commandBuf  string
	errorMsg    string
	errorID     uint64

	// Save-as prompt state
	savePromptMode      bool
	savePromptBuf       string
	savePromptQuitAfter bool
}

// NewStatusBar creates a StatusBar initialized with NORMAL mode and zero counts.
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		width:     width,
		modeLabel: "NORMAL",
		wordCount: 0,
		charCount: 0,
		dimStyle:  render.DimStyle(render.StatusBarDimPercent),
	}
}

// Set updates the mode label, word count, and character count simultaneously.
// Clears command mode but preserves any active error or save prompt — those
// are cleared only by their own mechanisms (timer and user action, respectively).
func (s *StatusBar) Set(modeLabel string, words, chars int) {
	s.commandMode = false
	s.commandBuf = ""
	s.modeLabel = modeLabel
	s.wordCount = words
	s.charCount = chars
}

// SetCommand activates command mode display with the given buffer string.
// Does not clear any active error — error display takes priority in View().
func (s *StatusBar) SetCommand(buf string) {
	s.commandMode = true
	s.commandBuf = buf
}

// SetError deactivates command mode and displays an error message.
// Each call increments errorID so stale dismiss timers can be ignored.
func (s *StatusBar) SetError(msg string) {
	s.commandMode = false
	s.commandBuf = ""
	s.errorMsg = msg
	s.errorID++
}

// ErrorID returns the current error generation ID.
func (s *StatusBar) ErrorID() uint64 {
	return s.errorID
}

// HasError reports whether an error message is currently displayed.
func (s *StatusBar) HasError() bool {
	return s.errorMsg != ""
}

// ClearError clears the error message only when id matches the current errorID.
// This prevents stale timers from dismissing a newer error.
func (s *StatusBar) ClearError(id uint64) {
	if id == s.errorID {
		s.errorMsg = ""
	}
}

// SetSavePrompt activates the save-as prompt mode.
// When quitAfter is true, a successful save exits the editor.
func (s *StatusBar) SetSavePrompt(quitAfter bool) {
	s.savePromptMode = true
	s.savePromptBuf = ""
	s.savePromptQuitAfter = quitAfter
}

// SavePromptBuf returns the current save-as prompt input buffer.
func (s *StatusBar) SavePromptBuf() string {
	return s.savePromptBuf
}

// InSavePrompt reports whether the save-as prompt is active.
func (s *StatusBar) InSavePrompt() bool {
	return s.savePromptMode
}

// SavePromptQuitAfter reports whether a successful save should exit the editor.
func (s *StatusBar) SavePromptQuitAfter() bool {
	return s.savePromptQuitAfter
}

// AppendSavePrompt appends a character to the save-as prompt buffer.
func (s *StatusBar) AppendSavePrompt(ch rune) {
	s.savePromptBuf += string(ch)
}

// BackspaceSavePrompt removes the last rune from the save-as prompt buffer.
func (s *StatusBar) BackspaceSavePrompt() {
	runes := []rune(s.savePromptBuf)
	if len(runes) > 0 {
		s.savePromptBuf = string(runes[:len(runes)-1])
	}
}

// ClearSavePrompt deactivates the save-as prompt and clears its buffer.
func (s *StatusBar) ClearSavePrompt() {
	s.savePromptMode = false
	s.savePromptBuf = ""
	s.savePromptQuitAfter = false
}

// Resize updates the terminal width used for centering.
func (s *StatusBar) Resize(width int) {
	s.width = width
}

// ModeLabel returns the current mode label string.
func (s *StatusBar) ModeLabel() string {
	return s.modeLabel
}

// Counts returns the current word and character counts.
func (s *StatusBar) Counts() (words, chars int) {
	return s.wordCount, s.charCount
}

// View returns the formatted, centered status bar string.
// When dimmed is true the entire line is rendered with the dim style.
// Priority (highest to lowest): error > save prompt > command mode > normal status.
// Error, save prompt, and command mode are never dimmed.
func (s *StatusBar) View(dimmed bool) string {
	var plain string
	switch {
	case s.errorMsg != "":
		plain = s.errorMsg
		dimmed = false
	case s.savePromptMode:
		plain = "Save as: " + s.savePromptBuf
		dimmed = false
	case s.commandMode:
		plain = ":" + s.commandBuf
		dimmed = false
	default:
		plain = fmt.Sprintf("%s · %dw · %dc", s.modeLabel, s.wordCount, s.charCount)
	}
	padLeft := (s.width - len([]rune(plain))) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	centered := strings.Repeat(" ", padLeft) + plain
	if dimmed {
		return s.dimStyle.Render(centered)
	}
	return centered
}
