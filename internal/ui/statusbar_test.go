package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestStatusBar_View_NormalMode(t *testing.T) {
	sb := NewStatusBar(40)
	sb.Set("NORMAL", 5, 25)
	got := sb.View(false)

	if !strings.Contains(got, "NORMAL · 5w · 25c") {
		t.Errorf("View(false) = %q, want it to contain %q", got, "NORMAL · 5w · 25c")
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("View(false) should not contain ANSI codes, got %q", got)
	}
}

func TestStatusBar_View_InsertMode_Dimmed(t *testing.T) {
	// Force color output so lipgloss renders ANSI codes in test environment.
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(termenv.Ascii)
		lipgloss.SetHasDarkBackground(false)
	})

	sb := NewStatusBar(40)
	sb.Set("INSERT", 3, 12)
	got := sb.View(true)

	if !strings.Contains(got, "\x1b[") {
		t.Errorf("View(true) should contain ANSI escape codes for dimming, got %q", got)
	}
}

func TestStatusBar_View_Centering(t *testing.T) {
	// "NORMAL · 0w · 0c" has length 17 runes
	// For width=40: padLeft = (40-17)/2 = 11
	sb := NewStatusBar(40)
	sb.Set("NORMAL", 0, 0)
	got := sb.View(false)

	text := "NORMAL · 0w · 0c"
	textLen := len([]rune(text))
	padLeft := (40 - textLen) / 2

	if !strings.HasPrefix(got, strings.Repeat(" ", padLeft)) {
		t.Errorf("View(false) should start with %d spaces, got %q", padLeft, got)
	}
}

func TestStatusBar_View_VisualMode_NotDimmed(t *testing.T) {
	sb := NewStatusBar(40)
	sb.Set("VISUAL", 2, 10)
	got := sb.View(false)

	if !strings.Contains(got, "VISUAL · 2w · 10c") {
		t.Errorf("View(false) = %q, want it to contain %q", got, "VISUAL · 2w · 10c")
	}
	if strings.Contains(got, "\x1b[") {
		t.Errorf("View(false) should not contain ANSI codes, got %q", got)
	}
}

func TestStatusBar_View_Dimmed_vs_Undimmed_OutputDiffers(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(termenv.Ascii)
		lipgloss.SetHasDarkBackground(false)
	})

	sb := NewStatusBar(40)
	sb.Set("INSERT", 3, 12)

	dimmed := sb.View(true)
	undimmed := sb.View(false)

	if dimmed == undimmed {
		t.Errorf("View(true) and View(false) should produce different output\ndimmed:   %q\nundimmed: %q", dimmed, undimmed)
	}
}

func TestStatusBar_View_CommandMode_ShowsColonBuf(t *testing.T) {
	sb := NewStatusBar(40)
	sb.SetCommand("q")
	got := sb.View(false)

	if !strings.Contains(got, ":q") {
		t.Errorf("View(false) = %q, want it to contain %q", got, ":q")
	}
}

func TestStatusBar_View_CommandMode_EmptyBuf(t *testing.T) {
	sb := NewStatusBar(40)
	sb.SetCommand("")
	got := sb.View(false)

	if !strings.Contains(got, ":") {
		t.Errorf("View(false) = %q, want it to contain %q", got, ":")
	}
	if strings.Contains(got, "NORMAL") {
		t.Errorf("View(false) = %q, should not contain NORMAL in command mode", got)
	}
}

func TestStatusBar_View_ErrorMsg(t *testing.T) {
	sb := NewStatusBar(40)
	sb.SetError("E: Not a command: xyz")
	got := sb.View(false)

	if !strings.Contains(got, "E: Not a command: xyz") {
		t.Errorf("View(false) = %q, want it to contain %q", got, "E: Not a command: xyz")
	}
}

func TestStatusBar_View_CommandMode_NotDimmed(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(termenv.Ascii)
		lipgloss.SetHasDarkBackground(false)
	})

	sb := NewStatusBar(40)
	sb.SetCommand("q")

	dimmed := sb.View(true)
	undimmed := sb.View(false)

	if dimmed != undimmed {
		t.Errorf("Command mode View(true) and View(false) should produce identical output (no dimming)\ndimmed:   %q\nundimmed: %q", dimmed, undimmed)
	}
}

func TestStatusBar_Set_ClearsCommandMode(t *testing.T) {
	sb := NewStatusBar(40)
	sb.SetCommand("q")
	sb.Set("NORMAL", 0, 0)
	got := sb.View(false)

	if strings.Contains(got, ":q") {
		t.Errorf("View(false) after Set() = %q, should not contain :q", got)
	}
	if !strings.Contains(got, "NORMAL") {
		t.Errorf("View(false) after Set() = %q, should contain NORMAL", got)
	}
}

func TestStatusBar_Resize(t *testing.T) {
	sb := NewStatusBar(40)
	sb.Set("NORMAL", 0, 0)

	// With width=40
	got40 := sb.View(false)
	text := "NORMAL · 0w · 0c"
	textLen := len([]rune(text))
	padLeft40 := (40 - textLen) / 2

	if !strings.HasPrefix(got40, strings.Repeat(" ", padLeft40)) {
		t.Errorf("before Resize: want %d leading spaces, got %q", padLeft40, got40)
	}

	// After Resize to width=80
	sb.Resize(80)
	got80 := sb.View(false)
	padLeft80 := (80 - textLen) / 2

	if !strings.HasPrefix(got80, strings.Repeat(" ", padLeft80)) {
		t.Errorf("after Resize(80): want %d leading spaces, got %q", padLeft80, got80)
	}

	// Padding should be larger for wider terminal
	if padLeft80 <= padLeft40 {
		t.Errorf("expected padding to increase after Resize: padLeft40=%d padLeft80=%d", padLeft40, padLeft80)
	}
}
