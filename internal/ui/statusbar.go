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
// Also clears any command mode or error state.
func (s *StatusBar) Set(modeLabel string, words, chars int) {
	s.commandMode = false
	s.commandBuf = ""
	s.errorMsg = ""
	s.modeLabel = modeLabel
	s.wordCount = words
	s.charCount = chars
}

// SetCommand activates command mode display with the given buffer string.
func (s *StatusBar) SetCommand(buf string) {
	s.commandMode = true
	s.commandBuf = buf
	s.errorMsg = ""
}

// SetError deactivates command mode and displays an error message.
func (s *StatusBar) SetError(msg string) {
	s.commandMode = false
	s.commandBuf = ""
	s.errorMsg = msg
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
// Command mode and error state are never dimmed.
func (s *StatusBar) View(dimmed bool) string {
	var plain string
	switch {
	case s.commandMode:
		plain = ":" + s.commandBuf
		dimmed = false
	case s.errorMsg != "":
		plain = s.errorMsg
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
