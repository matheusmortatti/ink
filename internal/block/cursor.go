package block

import (
	"strings"
)

// MapRenderedToRaw maps a cursor position in the Glamour-rendered output
// to the corresponding position in the raw markdown source.
// renderedLine and renderedCol are block-relative (0-indexed).
// The rendered position should already be ANSI-stripped (visual characters only).
func MapRenderedToRaw(b Block, renderedLine, renderedCol int) (rawLine, rawCol int) {
	rawLines := strings.Split(b.Raw, "\n")

	switch b.Type {
	case Heading:
		rawLine = renderedLine
		prefix := b.Level + 1 // "## " = level chars + 1 space
		if rawLine < len(rawLines) {
			content := rawLines[rawLine]
			runes := []rune(content)
			if b.Level == 1 {
				// H1: Glamour replaces the "# " prefix with background styling only,
				// so the rendered content starts at visual column 0.
				if prefix < len(runes) {
					rawCol = prefix + rawColFromVisualCol(string(runes[prefix:]), renderedCol)
				} else {
					rawCol = prefix + renderedCol
				}
			} else {
				// H2+: Glamour keeps "## " etc. as visible text in the rendered output.
				// renderedCol is measured from the start of the full rendered line, so
				// cols within the prefix zone map 1:1 to raw; cols past the prefix need
				// inline-marker adjustment.
				if renderedCol < prefix {
					rawCol = renderedCol
				} else if prefix < len(runes) {
					rawCol = prefix + rawColFromVisualCol(string(runes[prefix:]), renderedCol-prefix)
				} else {
					rawCol = renderedCol
				}
			}
		} else {
			rawCol = b.Level + 1 + renderedCol
		}

	case CodeFence:
		// Glamour renders code fences with an empty padding line at the top
		// (from ANSI background styling), which visually replaces the raw ```
		// delimiter line. So rendered and raw line indices are 1:1.
		rawLine = renderedLine
		rawCol = renderedCol

	case BlockQuote:
		// Glamour adds one empty padding line before blockquote content.
		// rendered 0 = empty, rendered 1 = first content line, etc.
		rawLine = renderedLine - 1
		if rawLine < 0 {
			rawLine = 0
		}
		if rawLine < len(rawLines) {
			prefix := blockQuotePrefixLen(rawLines[rawLine])
			content := rawLines[rawLine]
			runes := []rune(content)
			// Glamour replaces ">" with "│" but keeps the same visual width.
			// renderedCol is measured from the start of the full rendered line
			// (including the │ prefix), so cols in the prefix zone map 1:1 to raw;
			// cols past the prefix need inline-marker adjustment.
			if renderedCol < prefix {
				rawCol = renderedCol
			} else if prefix < len(runes) {
				rawCol = prefix + rawColFromVisualCol(string(runes[prefix:]), renderedCol-prefix)
			} else {
				rawCol = renderedCol
			}
		} else {
			rawCol = renderedCol
		}

	case List:
		// Glamour adds one empty padding line before list content.
		// rendered 0 = empty, rendered 1 = first item, rendered 2 = second item, etc.
		rawLine = renderedLine - 1
		if rawLine < 0 {
			rawLine = 0
		}
		if rawLine < len(rawLines) {
			prefix := listPrefixLen(rawLines[rawLine])
			content := rawLines[rawLine]
			runes := []rune(content)
			// Glamour replaces "-" / "*" with "•" but keeps the same visual width.
			// renderedCol is measured from the start of the full rendered line,
			// so cols in the prefix zone map 1:1 to raw.
			if renderedCol < prefix {
				rawCol = renderedCol
			} else if prefix < len(runes) {
				rawCol = prefix + rawColFromVisualCol(string(runes[prefix:]), renderedCol-prefix)
			} else {
				rawCol = renderedCol
			}
		} else {
			rawCol = renderedCol
		}

	case Table:
		rawLine, rawCol = mapRenderedToRawTable(b, renderedLine, renderedCol)

	default:
		rawLine = renderedLine
		if rawLine < len(rawLines) {
			rawCol = rawColFromVisualCol(rawLines[rawLine], renderedCol)
		} else {
			rawCol = renderedCol
		}
	}

	// Clamp to valid positions
	if rawLine < 0 {
		rawLine = 0
	}
	if rawLine >= len(rawLines) {
		rawLine = len(rawLines) - 1
	}
	if rawCol < 0 {
		rawCol = 0
	}
	lineRunes := []rune(rawLines[rawLine])
	if rawCol > len(lineRunes) {
		rawCol = len(lineRunes)
	}

	return rawLine, rawCol
}

// MapRawToRendered maps a cursor position in the raw markdown source
// to the corresponding position in the Glamour-rendered output.
// rawLine and rawCol are 0-indexed within the block's raw content.
func MapRawToRendered(b Block, rawLine, rawCol int) (renderedLine, renderedCol int) {
	rawLines := strings.Split(b.Raw, "\n")

	switch b.Type {
	case Heading:
		renderedLine = rawLine
		prefix := b.Level + 1
		if rawLine < len(rawLines) {
			runes := []rune(rawLines[rawLine])
			if b.Level == 1 {
				// H1: Glamour strips the prefix; raw prefix cols all map to rendered 0.
				if rawCol <= prefix {
					renderedCol = 0
				} else if prefix < len(runes) {
					renderedCol = visualColFromRawCol(string(runes[prefix:]), rawCol-prefix)
				}
			} else {
				// H2+: Glamour keeps the prefix as visible text; 1:1 in prefix zone.
				if rawCol < prefix {
					renderedCol = rawCol
				} else if prefix < len(runes) {
					renderedCol = prefix + visualColFromRawCol(string(runes[prefix:]), rawCol-prefix)
				} else {
					renderedCol = rawCol
				}
			}
		}

	case CodeFence:
		// Glamour's empty padding line at the top matches the raw ```
		// delimiter position, so content lines are 1:1.
		// Delimiter lines (opening/closing ```) map to rendered line 0.
		if rawLine == 0 || rawLine == len(rawLines)-1 {
			renderedLine = 0
			renderedCol = 0
		} else {
			renderedLine = rawLine
			renderedCol = rawCol
		}

	case BlockQuote:
		// Glamour adds one empty padding line before blockquote content.
		renderedLine = rawLine + 1
		if rawLine < len(rawLines) {
			prefix := blockQuotePrefixLen(rawLines[rawLine])
			runes := []rune(rawLines[rawLine])
			// Glamour keeps the same visual width as "> "; 1:1 in prefix zone.
			if rawCol < prefix {
				renderedCol = rawCol
			} else if prefix < len(runes) {
				renderedCol = prefix + visualColFromRawCol(string(runes[prefix:]), rawCol-prefix)
			} else {
				renderedCol = rawCol
			}
		}

	case List:
		// Glamour adds one empty padding line before list content.
		renderedLine = rawLine + 1
		if rawLine < len(rawLines) {
			prefix := listPrefixLen(rawLines[rawLine])
			runes := []rune(rawLines[rawLine])
			// Glamour keeps the same visual width as "- "; 1:1 in prefix zone.
			if rawCol < prefix {
				renderedCol = rawCol
			} else if prefix < len(runes) {
				renderedCol = prefix + visualColFromRawCol(string(runes[prefix:]), rawCol-prefix)
			} else {
				renderedCol = rawCol
			}
		}

	case Table:
		renderedLine, renderedCol = mapRawToRenderedTable(b, rawLine, rawCol)

	default:
		renderedLine = rawLine
		if rawLine < len(rawLines) {
			renderedCol = visualColFromRawCol(rawLines[rawLine], rawCol)
		} else {
			renderedCol = rawCol
		}
	}

	if renderedLine < 0 {
		renderedLine = 0
	}
	if renderedCol < 0 {
		renderedCol = 0
	}

	return renderedLine, renderedCol
}

// classifyRunes returns a boolean slice where true means the rune at that
// index is visible (content) and false means it is hidden (marker syntax).
// This handles nested markers (e.g., bold containing italic).
func classifyRunes(rawLine string) []bool {
	runes := []rune(rawLine)
	visible := make([]bool, len(runes))
	for i := range visible {
		visible[i] = true
	}
	classifyRange(runes, visible, 0, len(runes))
	return visible
}

// classifyRange marks inline marker runes as hidden (false) within runes[start:end].
func classifyRange(runes []rune, visible []bool, start, end int) {
	i := start
	for i < end {
		// Escaped character: backslash hidden, next char visible
		if runes[i] == '\\' && i+1 < end {
			visible[i] = false
			i += 2
			continue
		}

		// Image: ![alt](url)
		if i+1 < end && runes[i] == '!' && runes[i+1] == '[' {
			parenEnd, altEnd, urlStart := findLinkMarker(runes, i+1)
			if parenEnd >= 0 && parenEnd < end {
				visible[i] = false     // '!'
				visible[i+1] = false   // '['
				visible[altEnd] = false // ']'
				if altEnd+1 < end {
					visible[altEnd+1] = false // '('
				}
				for j := urlStart; j <= parenEnd; j++ {
					visible[j] = false
				}
				// Recursively classify the alt text
				classifyRange(runes, visible, i+2, altEnd)
				i = parenEnd + 1
				continue
			}
		}

		// Link: [text](url)
		if runes[i] == '[' {
			parenEnd, altEnd, urlStart := findLinkMarker(runes, i)
			if parenEnd >= 0 && parenEnd < end {
				visible[i] = false     // '['
				visible[altEnd] = false // ']'
				if altEnd+1 < end {
					visible[altEnd+1] = false // '('
				}
				for j := urlStart; j <= parenEnd; j++ {
					visible[j] = false
				}
				// Recursively classify the link text
				classifyRange(runes, visible, i+1, altEnd)
				i = parenEnd + 1
				continue
			}
		}

		// Code span: `code`
		// Glamour renders backtick markers as spaces (still a visible position),
		// so we leave them marked as visible and skip nested-marker parsing inside.
		if runes[i] == '`' {
			closePos := findClosingDelim(runes, i+1, '`', 1)
			if closePos >= 0 && closePos < end {
				// visible[i] and visible[closePos] remain true — backticks appear
				// as spaces in the rendered output and occupy a visual column.
				i = closePos + 1
				continue
			}
		}

		// Bold+Italic: ***text*** or ___text___
		if i+2 < end {
			if (runes[i] == '*' && runes[i+1] == '*' && runes[i+2] == '*') ||
				(runes[i] == '_' && runes[i+1] == '_' && runes[i+2] == '_') {
				closePos := findClosingDelim(runes, i+3, runes[i], 3)
				if closePos >= 0 && closePos+2 < end {
					visible[i] = false
					visible[i+1] = false
					visible[i+2] = false
					visible[closePos] = false
					visible[closePos+1] = false
					visible[closePos+2] = false
					classifyRange(runes, visible, i+3, closePos)
					i = closePos + 3
					continue
				}
			}
		}

		// Bold: **text** or __text__
		if i+1 < end {
			if (runes[i] == '*' && runes[i+1] == '*') || (runes[i] == '_' && runes[i+1] == '_') {
				closePos := findClosingDelim(runes, i+2, runes[i], 2)
				if closePos >= 0 && closePos+1 < end {
					visible[i] = false
					visible[i+1] = false
					visible[closePos] = false
					visible[closePos+1] = false
					// Recursively classify content for nested markers
					classifyRange(runes, visible, i+2, closePos)
					i = closePos + 2
					continue
				}
			}
		}

		// Italic: *text* or _text_ (single, not double)
		if runes[i] == '*' && (i+1 >= end || runes[i+1] != '*') {
			closePos := findClosingDelim(runes, i+1, '*', 1)
			if closePos >= 0 && closePos < end {
				visible[i] = false
				visible[closePos] = false
				classifyRange(runes, visible, i+1, closePos)
				i = closePos + 1
				continue
			}
		}
		if runes[i] == '_' && (i+1 >= end || runes[i+1] != '_') {
			closePos := findClosingDelim(runes, i+1, '_', 1)
			if closePos >= 0 && closePos < end {
				visible[i] = false
				visible[closePos] = false
				classifyRange(runes, visible, i+1, closePos)
				i = closePos + 1
				continue
			}
		}

		i++
	}
}

// rawColFromVisualCol converts a visual (rendered) column position to the
// corresponding raw column position, accounting for inline markers.
func rawColFromVisualCol(rawLine string, visualCol int) int {
	vis := classifyRunes(rawLine)
	visual := 0
	for raw := 0; raw < len(vis); raw++ {
		if vis[raw] {
			if visual == visualCol {
				return raw
			}
			visual++
		}
	}
	return len(vis)
}

// visualColFromRawCol converts a raw column position to the corresponding
// visual (rendered) column position, accounting for inline markers.
func visualColFromRawCol(rawLine string, rawCol int) int {
	vis := classifyRunes(rawLine)
	visual := 0
	for raw := 0; raw < len(vis) && raw < rawCol; raw++ {
		if vis[raw] {
			visual++
		}
	}
	return visual
}

// findClosingDelim finds a closing delimiter of count consecutive chars
// starting from position start. Returns the position or -1 if not found.
func findClosingDelim(runes []rune, start int, char rune, count int) int {
	for i := start; i <= len(runes)-count; i++ {
		if i > 0 && runes[i-1] == '\\' {
			// Count consecutive backslashes to handle escaped backslashes
			backslashes := 0
			for j := i - 1; j >= 0 && runes[j] == '\\'; j-- {
				backslashes++
			}
			if backslashes%2 == 1 {
				continue // odd count means the delimiter char is truly escaped
			}
		}
		match := true
		for j := 0; j < count; j++ {
			if runes[i+j] != char {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// findLinkMarker finds a markdown link starting with '[' at pos.
// Returns (endParenIdx, closeBracketIdx, urlStartIdx) or (-1, -1, -1).
func findLinkMarker(runes []rune, pos int) (int, int, int) {
	if pos >= len(runes) || runes[pos] != '[' {
		return -1, -1, -1
	}
	bracketEnd := -1
	for i := pos + 1; i < len(runes); i++ {
		if runes[i] == ']' {
			bracketEnd = i
			break
		}
	}
	if bracketEnd < 0 {
		return -1, -1, -1
	}
	if bracketEnd+1 >= len(runes) || runes[bracketEnd+1] != '(' {
		return -1, -1, -1
	}
	parenEnd := -1
	for i := bracketEnd + 2; i < len(runes); i++ {
		if runes[i] == ')' {
			parenEnd = i
			break
		}
	}
	if parenEnd < 0 {
		return -1, -1, -1
	}
	return parenEnd, bracketEnd, bracketEnd + 2
}

// blockQuotePrefixLen returns the length of the block quote prefix,
// handling nested quotes (e.g., "> > " returns 4).
func blockQuotePrefixLen(line string) int {
	runes := []rune(line)
	i := 0
	for i < len(runes) {
		if runes[i] == '>' {
			i++
			if i < len(runes) && runes[i] == ' ' {
				i++
			}
		} else {
			break
		}
	}
	return i
}

// listPrefixLen returns the length of the list marker prefix.
func listPrefixLen(line string) int {
	runes := []rune(line)
	if len(runes) == 0 {
		return 0
	}

	indent := 0
	for indent < len(runes) && (runes[indent] == ' ' || runes[indent] == '\t') {
		indent++
	}
	if indent >= len(runes) {
		return 0
	}

	// Unordered: -, *, +
	if (runes[indent] == '-' || runes[indent] == '*' || runes[indent] == '+') &&
		indent+1 < len(runes) && runes[indent+1] == ' ' {
		return indent + 2
	}

	// Ordered: N.
	i := indent
	for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
		i++
	}
	if i > indent && i < len(runes) && runes[i] == '.' && i+1 < len(runes) && runes[i+1] == ' ' {
		return i + 2
	}

	return 0
}

// mapRenderedToRawTable maps rendered table position to raw.
//
// Glamour's table rendering layout (ANSI-stripped, after stripRenderPadding):
//
//	rendered line 0 : empty padding line
//	rendered line 1 : header row
//	rendered line 2 : graphical border (─────┼─────)
//	rendered line 3+: data rows (one per raw non-separator data line)
//
// The raw "| --- | --- |" separator does not appear as its own rendered line;
// instead Glamour draws the border at rendered line 2.
func mapRenderedToRawTable(b Block, renderedLine, renderedCol int) (int, int) {
	rawLines := strings.Split(b.Raw, "\n")
	// Build list of non-separator raw line indices (header, then data rows).
	var contentLines []int
	for i, line := range rawLines {
		if !isTableSeparator(line) {
			contentLines = append(contentLines, i)
		}
	}
	// Subtract 1 to account for the empty leading padding line.
	// rendered 0 → adjusted -1 → clamp to 0 (header)
	// rendered 1 → adjusted  0 → contentLines[0] (header)
	// rendered 2 → adjusted  1 → contentLines[1] (first data row — border maps here)
	// rendered 3 → adjusted  2 → fallback rawLine=2 (first data row)
	adjusted := renderedLine - 1
	if adjusted < 0 {
		adjusted = 0
	}
	rawLine := renderedLine - 1 // fallback
	if rawLine < 0 {
		rawLine = 0
	}
	if adjusted < len(contentLines) {
		rawLine = contentLines[adjusted]
	}
	if rawLine < len(rawLines) {
		return rawLine, mapTableCol(rawLines[rawLine], renderedCol)
	}
	return rawLine, renderedCol
}

// mapRawToRenderedTable maps raw table position to rendered.
//
// Inverse of mapRenderedToRawTable. Layout offsets:
//
//	raw header (non-sep idx 0)  → rendered line 1
//	raw separator               → rendered line 2 (Glamour border)
//	raw data row (non-sep idx N, N≥1) → rendered line N+2
func mapRawToRenderedTable(b Block, rawLine, rawCol int) (int, int) {
	rawLines := strings.Split(b.Raw, "\n")
	// Separator maps to the Glamour border line.
	if rawLine < len(rawLines) && isTableSeparator(rawLines[rawLine]) {
		return 2, 0
	}
	// Count non-separator lines before rawLine to get this line's content index.
	nonSepIdx := 0
	for i := 0; i < rawLine && i < len(rawLines); i++ {
		if !isTableSeparator(rawLines[i]) {
			nonSepIdx++
		}
	}
	// header (nonSepIdx=0) → rendered 1
	// data rows (nonSepIdx≥1) → rendered nonSepIdx+2
	var renderedLine int
	if nonSepIdx == 0 {
		renderedLine = 1
	} else {
		renderedLine = nonSepIdx + 2
	}
	if rawLine < len(rawLines) {
		return renderedLine, unmapTableCol(rawLines[rawLine], rawCol)
	}
	return renderedLine, rawCol
}

// isTableSeparator returns true if the line is a markdown table separator row
// (e.g., "| --- | --- |" or "|:---:|---:|").
func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	hasDash := false
	for _, r := range trimmed {
		switch r {
		case '|', '-', ':', ' ':
			if r == '-' {
				hasDash = true
			}
		default:
			return false
		}
	}
	return hasDash
}

// mapTableCol maps a visual column in a rendered table to the raw column.
func mapTableCol(rawLine string, renderedCol int) int {
	runes := []rune(rawLine)
	pipeIdx := -1
	for i, r := range runes {
		if r == '|' {
			pipeIdx = i
			break
		}
	}
	if pipeIdx < 0 {
		return renderedCol
	}
	start := pipeIdx + 1
	if start < len(runes) && runes[start] == ' ' {
		start++
	}
	return start + renderedCol
}

// unmapTableCol maps a raw column in a table row to the visual column.
func unmapTableCol(rawLine string, rawCol int) int {
	runes := []rune(rawLine)
	pipeIdx := -1
	for i, r := range runes {
		if r == '|' {
			pipeIdx = i
			break
		}
	}
	if pipeIdx < 0 {
		return rawCol
	}
	start := pipeIdx + 1
	if start < len(runes) && runes[start] == ' ' {
		start++
	}
	if rawCol <= start {
		return 0
	}
	return rawCol - start
}
