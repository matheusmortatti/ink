package main

// LineMapEntry maps a block's source lines to display lines.
type LineMapEntry struct {
	BlockIndex   int
	IsActive     bool
	SourceStart  int
	SourceEnd    int
	DisplayStart int
	DisplayEnd   int
	DisplayLines []string
}

// LineMap provides bidirectional mapping between source and display coordinates.
type LineMap struct {
	entries          []LineMapEntry
	totalDisplayLines int
}

// NewLineMap creates an empty line map.
func NewLineMap() *LineMap {
	return &LineMap{}
}

// Clear resets the line map.
func (lm *LineMap) Clear() {
	lm.entries = lm.entries[:0]
	lm.totalDisplayLines = 0
}

// AddEntry adds a block mapping entry. displayLines are the rendered lines for this block.
func (lm *LineMap) AddEntry(blockIndex int, isActive bool, sourceStart, sourceEnd int, displayLines []string) {
	entry := LineMapEntry{
		BlockIndex:   blockIndex,
		IsActive:     isActive,
		SourceStart:  sourceStart,
		SourceEnd:    sourceEnd,
		DisplayStart: lm.totalDisplayLines,
		DisplayEnd:   lm.totalDisplayLines + len(displayLines) - 1,
		DisplayLines: displayLines,
	}
	lm.entries = append(lm.entries, entry)
	lm.totalDisplayLines += len(displayLines)
}

// AddBlankLine adds a single blank display line (for spacing between blocks).
func (lm *LineMap) AddBlankLine() {
	lm.entries = append(lm.entries, LineMapEntry{
		BlockIndex:   -1,
		IsActive:     false,
		SourceStart:  -1,
		SourceEnd:    -1,
		DisplayStart: lm.totalDisplayLines,
		DisplayEnd:   lm.totalDisplayLines,
		DisplayLines: []string{""},
	})
	lm.totalDisplayLines++
}

// TotalDisplayLines returns the total number of display lines.
func (lm *LineMap) TotalDisplayLines() int {
	return lm.totalDisplayLines
}

// SourceToDisplay converts a source line to a display line.
// For active blocks, this is a 1:1 mapping offset by DisplayStart.
// For inactive blocks, returns the start of the block's display range.
func (lm *LineMap) SourceToDisplay(sourceRow int) int {
	for _, entry := range lm.entries {
		if entry.SourceStart < 0 {
			continue
		}
		if sourceRow >= entry.SourceStart && sourceRow <= entry.SourceEnd {
			if entry.IsActive {
				offset := sourceRow - entry.SourceStart
				return entry.DisplayStart + offset
			}
			return entry.DisplayStart
		}
	}
	// Fallback: return sourceRow if not found
	if sourceRow < lm.totalDisplayLines {
		return sourceRow
	}
	return lm.totalDisplayLines - 1
}

// AllDisplayLines returns all display lines concatenated.
func (lm *LineMap) AllDisplayLines() []string {
	result := make([]string, 0, lm.totalDisplayLines)
	for _, entry := range lm.entries {
		result = append(result, entry.DisplayLines...)
	}
	return result
}

// Entries returns the line map entries.
func (lm *LineMap) Entries() []LineMapEntry {
	return lm.entries
}
