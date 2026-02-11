package ui

const (
	DefaultColumnWidth = 80
	MinCenteringWidth  = 40
	MediumBreakpoint   = 80
	LargeBreakpoint    = 120
	MediumWidthPercent = 70
)

// Layout holds calculated layout values for the writing column.
type Layout struct {
	TerminalWidth  int
	TerminalHeight int
	ColumnWidth    int
	LeftMargin     int
}

// NewLayout calculates layout from terminal dimensions.
func NewLayout(terminalWidth, terminalHeight int) Layout {
	col := CalculateColumnWidth(terminalWidth)
	margin := CalculateMargin(terminalWidth, col)
	return Layout{
		TerminalWidth:  terminalWidth,
		TerminalHeight: terminalHeight,
		ColumnWidth:    col,
		LeftMargin:     margin,
	}
}

// CalculateColumnWidth returns the writing column width for the given terminal width.
// Breakpoints: 120+ → 80, 80-119 → min(70%, 80), <80 → full width.
func CalculateColumnWidth(terminalWidth int) int {
	if terminalWidth <= 0 {
		return 0
	}
	if terminalWidth >= LargeBreakpoint {
		return DefaultColumnWidth
	}
	if terminalWidth >= MediumBreakpoint {
		w := terminalWidth * MediumWidthPercent / 100
		if w > DefaultColumnWidth {
			return DefaultColumnWidth
		}
		return w
	}
	return terminalWidth
}

// CalculateMargin returns the left margin to center the column within the terminal.
func CalculateMargin(terminalWidth, columnWidth int) int {
	return (terminalWidth - columnWidth) / 2
}
