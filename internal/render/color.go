package render

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// SyntaxDimPercent is the default dimming percentage for syntax characters.
const (
	SyntaxDimPercent    = 0.6 // syntax chars in active block: ~40% visible
	StatusBarDimPercent = 0.7 // status bar in insert mode: ~30% visible
)

// DimColor interpolates a foreground color toward a background color by the given
// percentage (0.0 = full fg, 1.0 = full bg). Used for syntax dimming (~0.6),
// status bar dimming (~0.7), and fade mode (~0.8).
func DimColor(fg, bg colorful.Color, percent float64) colorful.Color {
	if percent <= 0 {
		return fg
	}
	if percent >= 1 {
		return bg
	}
	return fg.BlendLab(bg, percent)
}

// DimStyle returns a Lip Gloss style with the foreground dimmed by the given
// percentage toward the terminal background.
func DimStyle(percent float64) lipgloss.Style {
	darkFg := colorful.Color{R: 1, G: 1, B: 1} // white text on dark bg
	darkBg := colorful.Color{R: 0, G: 0, B: 0}
	lightFg := colorful.Color{R: 0, G: 0, B: 0} // black text on light bg
	lightBg := colorful.Color{R: 1, G: 1, B: 1}

	dimmedDark := DimColor(darkFg, darkBg, percent)
	dimmedLight := DimColor(lightFg, lightBg, percent)

	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Dark:  colorToHex(dimmedDark),
		Light: colorToHex(dimmedLight),
	})
}

// colorToHex converts a go-colorful Color to a hex string like "#rrggbb".
func colorToHex(c colorful.Color) string {
	return c.Hex()
}

