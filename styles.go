package main

import (
	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	glamourstyles "github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	colorNormal    = lipgloss.Color("4")   // blue
	colorInsert    = lipgloss.Color("2")   // green
	colorVisual    = lipgloss.Color("5")   // magenta
	colorCommand   = lipgloss.Color("3")   // yellow
	colorStatusBg  = lipgloss.Color("236") // dark gray
	colorStatusFg  = lipgloss.Color("252") // light gray
	colorCursorBg  = lipgloss.Color("7")   // white
	colorCursorFg  = lipgloss.Color("0")   // black
	colorLineNr    = lipgloss.Color("240") // dim gray
	colorHeading   = lipgloss.Color("6")   // cyan
	colorBold      = lipgloss.Color("15")  // bright white
	colorCode      = lipgloss.Color("3")   // yellow
	colorCodeBg    = lipgloss.Color("235") // very dark gray
	colorVisualBg  = lipgloss.Color("17")  // dark blue
	colorSearchBg  = lipgloss.Color("3")   // yellow
	colorSearchFg  = lipgloss.Color("0")   // black
	colorBlockquote = lipgloss.Color("243") // gray
	colorLink      = lipgloss.Color("4")   // blue
)

// Status bar styles
var (
	statusModeStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Background(colorStatusBg).
			Foreground(colorStatusFg)

	statusFileStyle = lipgloss.NewStyle().
			Background(colorStatusBg).
			Foreground(colorStatusFg).
			Padding(0, 1)

	statusPosStyle = lipgloss.NewStyle().
			Background(colorStatusBg).
			Foreground(colorStatusFg).
			Padding(0, 1).
			Align(lipgloss.Right)
)

// Active block syntax highlighting styles
var (
	headingStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorHeading)

	boldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBold)

	codeStyle = lipgloss.NewStyle().
			Foreground(colorCode).
			Background(colorCodeBg)

	blockquoteStyle = lipgloss.NewStyle().
			Foreground(colorBlockquote)

	linkStyle = lipgloss.NewStyle().
			Foreground(colorLink).
			Underline(true)
)

// Cursor styles
var (
	cursorStyle = lipgloss.NewStyle().
			Background(colorCursorBg).
			Foreground(colorCursorFg)

	visualHighlightStyle = lipgloss.NewStyle().
				Background(colorVisualBg)

	searchHighlightStyle = lipgloss.NewStyle().
				Background(colorSearchBg).
				Foreground(colorSearchFg)
)

// ModeColor returns the color associated with a vim mode.
func ModeColor(mode Mode) lipgloss.Color {
	switch mode {
	case ModeInsert:
		return colorInsert
	case ModeVisual, ModeVisualLine:
		return colorVisual
	case ModeCommand:
		return colorCommand
	default:
		return colorNormal
	}
}

// customGlamourStyle returns a glamour style with zeroed vertical margins for tight rendering.
func customGlamourStyle() glamouransi.StyleConfig {
	// Start from the dark style and zero out margins
	style := glamourstyles.DarkStyleConfig
	zero := uint(0)

	style.Document.Margin = &zero
	style.Document.BlockPrefix = ""
	style.Document.BlockSuffix = ""

	style.Heading.BlockPrefix = ""
	style.Heading.BlockSuffix = ""

	style.Paragraph.BlockPrefix = ""
	style.Paragraph.BlockSuffix = ""

	style.List.LevelIndent = 2

	if style.CodeBlock.Margin == nil {
		style.CodeBlock.Margin = &zero
	} else {
		*style.CodeBlock.Margin = 0
	}

	return style
}

// NewGlamourRenderer creates a glamour renderer with our custom style.
func NewGlamourRenderer(width int) (*glamour.TermRenderer, error) {
	style := customGlamourStyle()
	return glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)
}
