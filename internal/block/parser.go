package block

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// md is a shared goldmark instance configured with GFM for table support.
var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

// Parse takes raw markdown source and returns a slice of Blocks.
// Uses goldmark for CommonMark-compliant AST parsing, then extracts
// raw source text for each top-level block element.
func Parse(source []byte) []Block {
	if len(source) == 0 {
		return nil
	}

	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	// Collect top-level block info from goldmark AST.
	type blockInfo struct {
		kind  ast.NodeKind
		start int // -1 means unresolved
		end   int // -1 means unresolved
		level int
	}

	var infos []blockInfo
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		start, end := nodeByteRange(child, source)
		level := 0
		if h, ok := child.(*ast.Heading); ok {
			level = h.Level
		}
		infos = append(infos, blockInfo{
			kind:  child.Kind(),
			start: start,
			end:   end,
			level: level,
		})
	}

	if len(infos) == 0 {
		return nil
	}

	// Resolve unresolved ranges (e.g., thematic breaks) using gaps.
	for i := range infos {
		if infos[i].start >= 0 && infos[i].end >= 0 {
			continue
		}
		// Determine the gap this block must occupy.
		gapStart := 0
		if i > 0 && infos[i-1].end >= 0 {
			gapStart = infos[i-1].end
		}
		gapEnd := len(source)
		if i+1 < len(infos) && infos[i+1].start >= 0 {
			gapEnd = infos[i+1].start
		}
		infos[i].start = gapStart
		infos[i].end = gapEnd
	}

	// Build Block slice, extracting raw text from source.
	blocks := make([]Block, 0, len(infos))
	for _, info := range infos {
		rawStart := info.start
		rawEnd := info.end
		if rawStart < 0 {
			rawStart = 0
		}
		if rawEnd < 0 || rawEnd > len(source) {
			rawEnd = len(source)
		}
		if rawStart > rawEnd {
			rawStart = rawEnd
		}

		raw, trimOffset := trimBlockSeparators(source[rawStart:rawEnd])
		contentStart := rawStart + trimOffset

		blocks = append(blocks, Block{
			Type:      kindToBlockType(info.kind),
			Raw:       string(raw),
			Level:     info.level,
			StartByte: contentStart,
			EndByte:   contentStart + len(raw),
		})
	}

	return blocks
}

// nodeByteRange computes the byte range [start, end) in source for a given
// AST node. Returns (-1, -1) if the range cannot be determined (e.g.,
// thematic breaks). The range includes syntax markers like # for headings,
// > for blockquotes, etc.
func nodeByteRange(n ast.Node, source []byte) (int, int) {
	// For fenced code blocks, include the fence lines.
	if fc, ok := n.(*ast.FencedCodeBlock); ok {
		return fencedCodeRange(fc, source)
	}

	// Try Lines() first — works for Paragraph, Heading, CodeBlock.
	if lines := n.Lines(); lines != nil && lines.Len() > 0 {
		first := lines.At(0)
		last := lines.At(lines.Len() - 1)
		// Expand to line start to include syntax markers (e.g., # for headings).
		start := lineStart(source, first.Start)
		end := last.Stop

		// goldmark strips trailing whitespace from line segments. Extend
		// end to recover trailing spaces/tabs (meaningful in markdown,
		// e.g. two trailing spaces = hard line break).
		for end < len(source) && (source[end] == ' ' || source[end] == '\t') {
			end++
		}

		// For setext headings, the underline (=== or ---) is on the next line
		// but NOT included in Lines(). Detect and include it.
		if h, ok := n.(*ast.Heading); ok {
			raw := source[start:end]
			// Setext headings don't start with #
			if len(raw) > 0 && raw[0] != '#' {
				end = setextEnd(source, end, h.Level)
			}
		}

		return start, end
	}

	// For container blocks (List, BlockQuote, Table), walk descendants
	// but only check block-type nodes for Lines() to avoid panics on inlines.
	minStart := len(source)
	maxEnd := 0
	_ = ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		// Only call Lines() on block-type nodes; inline nodes panic.
		if child.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}
		lines := child.Lines()
		if lines == nil {
			return ast.WalkContinue, nil
		}
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			if seg.Start < minStart {
				minStart = seg.Start
			}
			if seg.Stop > maxEnd {
				maxEnd = seg.Stop
			}
		}
		return ast.WalkContinue, nil
	})

	if minStart < maxEnd {
		// Expand to line start to include container syntax (>, -, *, etc.)
		minStart = lineStart(source, minStart)
		// Expand to line end to include trailing delimiters (e.g., | for tables).
		maxEnd = lineEnd(source, maxEnd)
		return minStart, maxEnd
	}

	// Cannot determine range (e.g., ThematicBreak has no lines).
	return -1, -1
}

// fencedCodeRange returns the byte range for a fenced code block,
// including the opening and closing fence lines.
func fencedCodeRange(fc *ast.FencedCodeBlock, source []byte) (int, int) {
	lines := fc.Lines()

	var contentStart, contentEnd int
	hasContent := lines != nil && lines.Len() > 0
	if hasContent {
		contentStart = lines.At(0).Start
		contentEnd = lines.At(lines.Len() - 1).Stop
	}

	// Find opening fence line by scanning backwards from content start.
	fenceStart := 0
	if hasContent && contentStart > 0 {
		pos := contentStart - 1
		// Skip newline(s) between fence and content.
		for pos > 0 && source[pos] == '\n' {
			pos--
		}
		// Scan to start of the opening fence line.
		for pos > 0 && source[pos-1] != '\n' {
			pos--
		}
		fenceStart = pos
	} else if !hasContent {
		// Empty code fence — no content lines. We can't easily find it
		// without sibling context, so return -1,-1 for gap resolution.
		return -1, -1
	}

	// Find closing fence line after content.
	fenceEnd := contentEnd
	if contentEnd < len(source) {
		pos := contentEnd
		// Scan to end of closing fence line.
		for pos < len(source) && source[pos] != '\n' {
			pos++
		}
		if pos < len(source) && source[pos] == '\n' {
			pos++ // include newline
		}
		fenceEnd = pos
	}

	return fenceStart, fenceEnd
}

// setextEnd extends the end position past the setext underline for headings.
// level 1 uses '=' and level 2 uses '-'.
func setextEnd(source []byte, pos int, level int) int {
	// Skip the newline after the heading text.
	if pos < len(source) && source[pos] == '\n' {
		pos++
	}
	// Determine expected underline character.
	var ch byte = '='
	if level == 2 {
		ch = '-'
	}
	// Scan past the underline characters.
	start := pos
	for pos < len(source) && source[pos] == ch {
		pos++
	}
	if pos > start {
		// Found underline — include up to end of line.
		for pos < len(source) && source[pos] != '\n' {
			pos++
		}
		return pos
	}
	// No underline found, return original position.
	return start
}

// lineStart returns the position of the start of the line containing pos.
func lineStart(source []byte, pos int) int {
	for pos > 0 && source[pos-1] != '\n' {
		pos--
	}
	return pos
}

// lineEnd returns the position just past the end of the line containing pos.
// If pos is already at a newline or past the last character, it finds the
// end of the current line content (excluding the newline itself).
func lineEnd(source []byte, pos int) int {
	// If pos is at start or after a newline, we may already be past the line.
	// Back up if we're sitting right on a newline to get the right line.
	if pos > 0 && pos <= len(source) && (pos == len(source) || source[pos] == '\n') {
		// We're at the end of a line already or at EOF. Check if the
		// previous character is also a newline (blank line). If pos
		// points into the content, scan forward.
		return pos
	}
	for pos < len(source) && source[pos] != '\n' {
		pos++
	}
	return pos
}

// trimBlockSeparators trims leading blank lines and trailing line endings
// from a raw block slice, producing the raw content of the block itself.
// Returns the trimmed content and the byte offset where content starts
// (relative to the input slice), needed for accurate StartByte calculation.
func trimBlockSeparators(b []byte) ([]byte, int) {
	start := 0
	end := len(b)

	// Trim leading blank lines.
	for start < end {
		// Find end of current line.
		lineEnd := start
		for lineEnd < end && b[lineEnd] != '\n' {
			lineEnd++
		}
		// Check if line is blank (only whitespace).
		blank := true
		for j := start; j < lineEnd; j++ {
			if b[j] != ' ' && b[j] != '\t' && b[j] != '\r' {
				blank = false
				break
			}
		}
		if !blank {
			break
		}
		// Skip past the newline.
		if lineEnd < end {
			lineEnd++
		}
		start = lineEnd
	}

	// Trim trailing line endings only (not spaces/tabs — trailing spaces
	// are meaningful in markdown, e.g. two trailing spaces = hard line break).
	for end > start && (b[end-1] == '\n' || b[end-1] == '\r') {
		end--
	}

	return b[start:end], start
}

// kindToBlockType maps a goldmark AST node kind to our BlockType.
func kindToBlockType(kind ast.NodeKind) BlockType {
	switch kind {
	case ast.KindParagraph:
		return Paragraph
	case ast.KindHeading:
		return Heading
	case ast.KindFencedCodeBlock:
		return CodeFence
	case ast.KindCodeBlock:
		return CodeBlock
	case ast.KindBlockquote:
		return BlockQuote
	case ast.KindList:
		return List
	case ast.KindThematicBreak:
		return HorizontalRule
	case east.KindTable:
		return Table
	default:
		return Paragraph
	}
}
