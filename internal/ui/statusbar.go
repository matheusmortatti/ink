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
	width     int
	modeLabel string
	wordCount int
	charCount int
	dimStyle  lipgloss.Style
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
func (s *StatusBar) Set(modeLabel string, words, chars int) {
	s.modeLabel = modeLabel
	s.wordCount = words
	s.charCount = chars
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
func (s *StatusBar) View(dimmed bool) string {
	plain := fmt.Sprintf("%s · %dw · %dc", s.modeLabel, s.wordCount, s.charCount)
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
