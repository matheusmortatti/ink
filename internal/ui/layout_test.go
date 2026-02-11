package ui

import "testing"

func TestCalculateColumnWidth_LargeTerminal_Returns80(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		want          int
	}{
		{"exactly 120", 120, 80},
		{"wide terminal 200", 200, 80},
		{"very wide 500", 500, 80},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateColumnWidth(tt.terminalWidth)
			if got != tt.want {
				t.Errorf("CalculateColumnWidth(%d) = %d, want %d", tt.terminalWidth, got, tt.want)
			}
		})
	}
}

func TestCalculateColumnWidth_MediumTerminal_Returns70Percent(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		want          int
	}{
		{"exactly 80", 80, 56},   // 80 * 70 / 100 = 56
		{"100 wide", 100, 70},    // 100 * 70 / 100 = 70
		{"119 wide", 119, 80},    // 119 * 70 / 100 = 83, capped at 80
		{"90 wide", 90, 63},      // 90 * 70 / 100 = 63
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateColumnWidth(tt.terminalWidth)
			if got != tt.want {
				t.Errorf("CalculateColumnWidth(%d) = %d, want %d", tt.terminalWidth, got, tt.want)
			}
		})
	}
}

func TestCalculateColumnWidth_MediumTerminal_CapsAt80(t *testing.T) {
	// Terminal widths 116-119 would produce >80 at 70%, but must cap at 80
	tests := []struct {
		name          string
		terminalWidth int
		want          int
	}{
		{"116 wide", 116, 80},  // 116 * 70 / 100 = 81, capped at 80
		{"117 wide", 117, 80},  // 117 * 70 / 100 = 81, capped at 80
		{"118 wide", 118, 80},  // 118 * 70 / 100 = 82, capped at 80
		{"119 wide", 119, 80},  // 119 * 70 / 100 = 83, capped at 80
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateColumnWidth(tt.terminalWidth)
			if got != tt.want {
				t.Errorf("CalculateColumnWidth(%d) = %d, want %d", tt.terminalWidth, got, tt.want)
			}
		})
	}
}

func TestCalculateColumnWidth_SmallTerminal_ReturnsFullWidth(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		want          int
	}{
		{"exactly 40", 40, 40},
		{"60 wide", 60, 60},
		{"79 wide", 79, 79},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateColumnWidth(tt.terminalWidth)
			if got != tt.want {
				t.Errorf("CalculateColumnWidth(%d) = %d, want %d", tt.terminalWidth, got, tt.want)
			}
		})
	}
}

func TestCalculateColumnWidth_VerySmallTerminal_ReturnsFullWidth(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		want          int
	}{
		{"20 wide", 20, 20},
		{"39 wide", 39, 39},
		{"1 wide", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateColumnWidth(tt.terminalWidth)
			if got != tt.want {
				t.Errorf("CalculateColumnWidth(%d) = %d, want %d", tt.terminalWidth, got, tt.want)
			}
		})
	}
}

func TestCalculateColumnWidth_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		want          int
	}{
		{"zero width", 0, 0},
		{"negative width", -1, 0},
		{"negative large", -100, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateColumnWidth(tt.terminalWidth)
			if got != tt.want {
				t.Errorf("CalculateColumnWidth(%d) = %d, want %d", tt.terminalWidth, got, tt.want)
			}
		})
	}
}

func TestCalculateMargin(t *testing.T) {
	tests := []struct {
		name          string
		terminalWidth int
		columnWidth   int
		want          int
	}{
		{"120 terminal 80 column", 120, 80, 20},
		{"200 terminal 80 column", 200, 80, 60},
		{"100 terminal 70 column", 100, 70, 15},
		{"80 terminal 80 column (no margin)", 80, 80, 0},
		{"40 terminal 40 column (no margin)", 40, 40, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateMargin(tt.terminalWidth, tt.columnWidth)
			if got != tt.want {
				t.Errorf("CalculateMargin(%d, %d) = %d, want %d", tt.terminalWidth, tt.columnWidth, got, tt.want)
			}
		})
	}
}

func TestNewLayout(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		wantCol    int
		wantMargin int
	}{
		{"large terminal", 120, 40, 80, 20},
		{"medium terminal", 100, 30, 70, 15},
		{"small terminal", 60, 20, 60, 0},
		{"very small terminal", 20, 10, 20, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLayout(tt.width, tt.height)
			if l.ColumnWidth != tt.wantCol {
				t.Errorf("NewLayout(%d, %d).ColumnWidth = %d, want %d", tt.width, tt.height, l.ColumnWidth, tt.wantCol)
			}
			if l.LeftMargin != tt.wantMargin {
				t.Errorf("NewLayout(%d, %d).LeftMargin = %d, want %d", tt.width, tt.height, l.LeftMargin, tt.wantMargin)
			}
			if l.TerminalWidth != tt.width {
				t.Errorf("NewLayout(%d, %d).TerminalWidth = %d, want %d", tt.width, tt.height, l.TerminalWidth, tt.width)
			}
			if l.TerminalHeight != tt.height {
				t.Errorf("NewLayout(%d, %d).TerminalHeight = %d, want %d", tt.width, tt.height, l.TerminalHeight, tt.height)
			}
		})
	}
}
