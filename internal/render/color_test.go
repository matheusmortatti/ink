package render

import (
	"math"
	"strings"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestDimColor_ZeroPercent_ReturnsFG(t *testing.T) {
	fg := colorful.Color{R: 1, G: 1, B: 1}
	bg := colorful.Color{R: 0, G: 0, B: 0}

	result := DimColor(fg, bg, 0.0)
	if result != fg {
		t.Errorf("DimColor at 0%% = %v, want %v", result, fg)
	}
}

func TestDimColor_FullPercent_ReturnsBG(t *testing.T) {
	fg := colorful.Color{R: 1, G: 1, B: 1}
	bg := colorful.Color{R: 0, G: 0, B: 0}

	result := DimColor(fg, bg, 1.0)
	if result != bg {
		t.Errorf("DimColor at 100%% = %v, want %v", result, bg)
	}
}

func TestDimColor_FiftyPercent_Midpoint(t *testing.T) {
	fg := colorful.Color{R: 1, G: 1, B: 1}
	bg := colorful.Color{R: 0, G: 0, B: 0}

	result := DimColor(fg, bg, 0.5)
	// In Lab space, midpoint of white and black should be a mid-gray
	// L* should be approximately 50
	l, _, _ := result.Lab()
	if math.Abs(l-0.5) > 0.05 {
		t.Errorf("DimColor at 50%% Lab L = %f, want ~0.5", l)
	}
}

func TestDimColor_SixtyPercent_CloserToBG(t *testing.T) {
	fg := colorful.Color{R: 1, G: 1, B: 1}
	bg := colorful.Color{R: 0, G: 0, B: 0}

	result := DimColor(fg, bg, 0.6)
	l, _, _ := result.Lab()
	// At 60% toward black, L should be around 0.4
	if l > 0.5 {
		t.Errorf("DimColor at 60%% should be closer to bg, L = %f", l)
	}
}

func TestDimColor_NegativePercent_ClampsFG(t *testing.T) {
	fg := colorful.Color{R: 1, G: 0, B: 0}
	bg := colorful.Color{R: 0, G: 0, B: 1}

	result := DimColor(fg, bg, -0.5)
	if result != fg {
		t.Errorf("DimColor at negative percent = %v, want fg %v", result, fg)
	}
}

func TestDimColor_OverOnePercent_ClampsBG(t *testing.T) {
	fg := colorful.Color{R: 1, G: 0, B: 0}
	bg := colorful.Color{R: 0, G: 0, B: 1}

	result := DimColor(fg, bg, 1.5)
	if result != bg {
		t.Errorf("DimColor at >100%% = %v, want bg %v", result, bg)
	}
}

func TestDimStyle_ReturnsStyle(t *testing.T) {
	style := DimStyle(0.6)
	// DimStyle should produce a style that can render text
	rendered := style.Render("test")
	// In non-TTY test environments, lipgloss may strip ANSI codes.
	// Just verify the text content survives rendering.
	if !strings.Contains(rendered, "test") {
		t.Errorf("DimStyle.Render lost text content, got %q", rendered)
	}
}

func TestDimStyle_DifferentPercents_ProduceStyles(t *testing.T) {
	// Verify DimStyle produces valid styles at different percentages
	// without panicking. In non-TTY environments, output may be identical
	// (no ANSI codes), so we just verify no errors occur.
	percents := []float64{0.0, 0.3, 0.6, 0.9, 1.0}
	for _, p := range percents {
		style := DimStyle(p)
		rendered := style.Render("x")
		if !strings.Contains(rendered, "x") {
			t.Errorf("DimStyle(%f).Render lost text content", p)
		}
	}
}

func TestColorToHex(t *testing.T) {
	white := colorful.Color{R: 1, G: 1, B: 1}
	hex := colorToHex(white)
	if hex != "#ffffff" {
		t.Errorf("colorToHex(white) = %q, want #ffffff", hex)
	}

	black := colorful.Color{R: 0, G: 0, B: 0}
	hex = colorToHex(black)
	if hex != "#000000" {
		t.Errorf("colorToHex(black) = %q, want #000000", hex)
	}
}
