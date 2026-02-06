package main

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderCache caches glamour-rendered blocks by content hash.
type RenderCache struct {
	entries map[int]renderCacheEntry
}

type renderCacheEntry struct {
	contentHash string
	rendered    []string
}

// NewRenderCache creates a new render cache.
func NewRenderCache() *RenderCache {
	return &RenderCache{
		entries: make(map[int]renderCacheEntry),
	}
}

// Get returns cached rendered lines for a block if the content matches.
func (rc *RenderCache) Get(blockIndex int, content string) ([]string, bool) {
	entry, ok := rc.entries[blockIndex]
	if !ok {
		return nil, false
	}
	hash := contentHash(content)
	if entry.contentHash != hash {
		return nil, false
	}
	return entry.rendered, true
}

// Set stores rendered lines for a block.
func (rc *RenderCache) Set(blockIndex int, content string, lines []string) {
	rc.entries[blockIndex] = renderCacheEntry{
		contentHash: contentHash(content),
		rendered:    lines,
	}
}

// Invalidate removes all entries.
func (rc *RenderCache) Invalidate() {
	rc.entries = make(map[int]renderCacheEntry)
}

func contentHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:8])
}

// HybridRenderer produces the hybrid display output.
type HybridRenderer struct {
	cache    *RenderCache
	lineMap  *LineMap
	width    int
}

// NewHybridRenderer creates a new hybrid renderer.
func NewHybridRenderer(width int) *HybridRenderer {
	return &HybridRenderer{
		cache:   NewRenderCache(),
		lineMap: NewLineMap(),
		width:   width,
	}
}

// Render produces display lines and a line map from the document and buffer state.
func (hr *HybridRenderer) Render(doc *Document, buf *Buffer, vim *VimState, activeBlockIdx int) *LineMap {
	hr.lineMap.Clear()

	if doc == nil || len(doc.Blocks()) == 0 {
		// No blocks: render raw buffer lines
		for i := 0; i < buf.LineCount(); i++ {
			line := hr.padLine(hr.renderRawLine(buf.Line(i), i, buf, vim))
			hr.lineMap.AddEntry(-1, true, i, i, []string{line})
		}
		return hr.lineMap
	}

	blocks := doc.Blocks()
	lastEndLine := -1

	for i, block := range blocks {
		// Add blank lines for gaps between blocks
		if block.StartLine > lastEndLine+1 {
			gapStart := lastEndLine + 1
			for g := gapStart; g < block.StartLine; g++ {
				sourceLine := buf.Line(g)
				if strings.TrimSpace(sourceLine) == "" {
					hr.lineMap.AddBlankLine()
				}
			}
		}

		isActive := i == activeBlockIdx

		if isActive {
			hr.renderActiveBlock(block, buf, vim)
		} else {
			hr.renderInactiveBlock(i, block)
		}

		lastEndLine = block.EndLine
	}

	// Handle trailing lines after last block
	if len(blocks) > 0 && blocks[len(blocks)-1].EndLine < buf.LineCount()-1 {
		for g := blocks[len(blocks)-1].EndLine + 1; g < buf.LineCount(); g++ {
			hr.lineMap.AddBlankLine()
		}
	}

	return hr.lineMap
}

// renderActiveBlock renders a block as raw text with syntax highlighting and cursor.
func (hr *HybridRenderer) renderActiveBlock(block Block, buf *Buffer, vim *VimState) {
	var displayLines []string
	for line := block.StartLine; line <= block.EndLine; line++ {
		rendered := hr.renderRawLine(buf.Line(line), line, buf, vim)
		displayLines = append(displayLines, hr.padLine(rendered))
	}
	hr.lineMap.AddEntry(-1, true, block.StartLine, block.EndLine, displayLines)
}

// renderRawLine renders a single source line with syntax highlighting, cursor, and visual selection.
func (hr *HybridRenderer) renderRawLine(text string, sourceLine int, buf *Buffer, vim *VimState) string {
	isCursorLine := sourceLine == buf.cursor.Row
	cursorCol := buf.cursor.Col

	// Apply lightweight syntax highlighting to the text
	highlighted := hr.syntaxHighlight(text)

	// If this is the cursor line, we need to show the cursor
	if isCursorLine && (vim.Mode == ModeNormal || vim.Mode == ModeInsert || vim.Mode == ModeReplace) {
		return hr.renderLineWithCursor(highlighted, text, cursorCol, vim.Mode)
	}

	// Visual mode highlighting
	if vim.Mode == ModeVisual || vim.Mode == ModeVisualLine {
		start, end := vim.visualRange(buf)
		if sourceLine >= start.Row && sourceLine <= end.Row {
			return hr.renderLineWithVisual(text, sourceLine, start, end, vim.Mode, isCursorLine, cursorCol)
		}
	}

	return highlighted
}

// syntaxHighlight applies lightweight markdown syntax highlighting to raw text.
func (hr *HybridRenderer) syntaxHighlight(text string) string {
	if len(text) == 0 {
		return ""
	}

	trimmed := strings.TrimLeft(text, " \t")

	// Headings: lines starting with #
	if strings.HasPrefix(trimmed, "#") {
		return headingStyle.Render(text)
	}

	// Blockquote: lines starting with >
	if strings.HasPrefix(trimmed, ">") {
		return blockquoteStyle.Render(text)
	}

	// Fenced code block markers
	if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
		return codeStyle.Render(text)
	}

	// Thematic breaks
	if len(trimmed) >= 3 {
		allDash := true
		allStar := true
		allUnder := true
		for _, ch := range trimmed {
			if ch != '-' && ch != ' ' {
				allDash = false
			}
			if ch != '*' && ch != ' ' {
				allStar = false
			}
			if ch != '_' && ch != ' ' {
				allUnder = false
			}
		}
		if allDash || allStar || allUnder {
			// Could be a thematic break
		}
	}

	// Inline highlighting: bold, code, etc.
	result := hr.highlightInline(text)
	return result
}

// highlightInline handles **bold**, `code`, and other inline markdown.
func (hr *HybridRenderer) highlightInline(text string) string {
	var result strings.Builder
	i := 0
	for i < len(text) {
		// Inline code: `...`
		if text[i] == '`' {
			end := strings.Index(text[i+1:], "`")
			if end >= 0 {
				codeText := text[i : i+end+2]
				result.WriteString(codeStyle.Render(codeText))
				i += end + 2
				continue
			}
		}

		// Bold: **...**
		if i+1 < len(text) && text[i] == '*' && text[i+1] == '*' {
			end := strings.Index(text[i+2:], "**")
			if end >= 0 {
				boldText := text[i : i+end+4]
				result.WriteString(boldStyle.Render(boldText))
				i += end + 4
				continue
			}
		}

		result.WriteByte(text[i])
		i++
	}
	return result.String()
}

// renderLineWithCursor inserts the cursor block character into the line.
func (hr *HybridRenderer) renderLineWithCursor(highlighted, raw string, cursorCol int, mode Mode) string {
	// For cursor rendering, we work with the raw text to position correctly,
	// then build the display with the cursor highlight.
	if mode == ModeInsert {
		// Insert mode: thin cursor (line between chars)
		// We render by inserting a visible cursor marker
		if cursorCol >= len(raw) {
			return highlighted + cursorStyle.Render(" ")
		}
		// Split raw text and rebuild with cursor
		before := raw[:cursorCol]
		ch := string(raw[cursorCol])
		after := raw[cursorCol+1:]
		return hr.syntaxHighlight(before) + cursorStyle.Render(ch) + hr.syntaxHighlight(after)
	}

	// Normal mode: block cursor over character
	if len(raw) == 0 {
		return cursorStyle.Render(" ")
	}
	if cursorCol >= len(raw) {
		cursorCol = len(raw) - 1
	}
	if cursorCol < 0 {
		cursorCol = 0
	}
	before := raw[:cursorCol]
	ch := string(raw[cursorCol])
	after := ""
	if cursorCol+1 < len(raw) {
		after = raw[cursorCol+1:]
	}
	return hr.syntaxHighlight(before) + cursorStyle.Render(ch) + hr.syntaxHighlight(after)
}

// renderLineWithVisual renders a line with visual selection highlighting.
func (hr *HybridRenderer) renderLineWithVisual(text string, sourceLine int, start, end Position, mode Mode, isCursorLine bool, cursorCol int) string {
	if mode == ModeVisualLine {
		result := visualHighlightStyle.Render(text)
		if isCursorLine && cursorCol < len(text) {
			// Also show cursor within visual line
			before := text[:cursorCol]
			ch := string(text[cursorCol])
			after := ""
			if cursorCol+1 < len(text) {
				after = text[cursorCol+1:]
			}
			return visualHighlightStyle.Render(before) + cursorStyle.Render(ch) + visualHighlightStyle.Render(after)
		}
		if len(text) == 0 {
			result = visualHighlightStyle.Render(" ")
		}
		return result
	}

	// Character-wise visual
	selStart := 0
	selEnd := len(text)

	if sourceLine == start.Row {
		selStart = start.Col
	}
	if sourceLine == end.Row {
		selEnd = end.Col + 1
	}

	if selStart > len(text) {
		selStart = len(text)
	}
	if selEnd > len(text) {
		selEnd = len(text)
	}

	before := text[:selStart]
	selected := text[selStart:selEnd]
	after := text[selEnd:]

	result := before + visualHighlightStyle.Render(selected) + after

	// Overlay cursor
	if isCursorLine && cursorCol >= 0 && cursorCol < len(text) {
		// Rebuild with cursor
		b := text[:cursorCol]
		ch := string(text[cursorCol])
		a := ""
		if cursorCol+1 < len(text) {
			a = text[cursorCol+1:]
		}
		_ = result
		// Determine if cursor is in selection
		if cursorCol >= selStart && cursorCol < selEnd {
			beforeSel := text[:selStart]
			selBeforeCur := text[selStart:cursorCol]
			selAfterCur := ""
			if cursorCol+1 < selEnd {
				selAfterCur = text[cursorCol+1 : selEnd]
			}
			afterSel := text[selEnd:]
			return beforeSel +
				visualHighlightStyle.Render(selBeforeCur) +
				cursorStyle.Render(ch) +
				visualHighlightStyle.Render(selAfterCur) +
				afterSel
		}
		return b + cursorStyle.Render(ch) + a
	}

	return result
}

// renderInactiveBlock renders a block through glamour (cached), or as raw text for paragraphs.
func (hr *HybridRenderer) renderInactiveBlock(blockIndex int, block Block) {
	if block.Kind == BlockParagraph {
		hr.renderInactiveRawBlock(blockIndex, block)
		return
	}

	// Check cache
	if cached, ok := hr.cache.Get(blockIndex, block.RawText); ok {
		hr.lineMap.AddEntry(blockIndex, false, block.StartLine, block.EndLine, cached)
		return
	}

	// Render with glamour
	rendered := hr.glamourRender(block)
	hr.cache.Set(blockIndex, block.RawText, rendered)
	hr.lineMap.AddEntry(blockIndex, false, block.StartLine, block.EndLine, rendered)
}

// renderInactiveRawBlock renders a block as raw text with syntax highlighting (no cursor/visual overlays).
// Used for paragraph blocks so they look the same whether active or inactive.
func (hr *HybridRenderer) renderInactiveRawBlock(blockIndex int, block Block) {
	lines := strings.Split(block.RawText, "\n")
	var displayLines []string
	for _, line := range lines {
		highlighted := hr.syntaxHighlight(line)
		displayLines = append(displayLines, hr.padLine(highlighted))
	}
	hr.lineMap.AddEntry(blockIndex, false, block.StartLine, block.EndLine, displayLines)
}

// glamourRender renders markdown text through glamour and returns display lines.
func (hr *HybridRenderer) glamourRender(block Block) []string {
	renderer, err := NewGlamourRenderer(hr.width)
	if err != nil {
		// Fallback: return raw text
		return strings.Split(block.RawText, "\n")
	}

	// Render the block as a standalone markdown fragment
	rendered, err := renderer.Render(block.RawText)
	if err != nil {
		return strings.Split(block.RawText, "\n")
	}

	// Clean up: trim trailing whitespace and empty lines
	rendered = strings.TrimRight(rendered, "\n \t")
	if rendered == "" {
		return []string{""}
	}

	lines := strings.Split(rendered, "\n")

	// Trim trailing empty lines
	for len(lines) > 1 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	// Trim leading empty lines
	for len(lines) > 1 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	return lines
}

// padLine pads a line to the viewport width.
func (hr *HybridRenderer) padLine(line string) string {
	visWidth := lipgloss.Width(line)
	if visWidth < hr.width {
		return line + strings.Repeat(" ", hr.width-visWidth)
	}
	return line
}
