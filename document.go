package main

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// BlockKind represents the type of a markdown block.
type BlockKind int

const (
	BlockParagraph BlockKind = iota
	BlockHeading
	BlockFencedCode
	BlockList
	BlockBlockquote
	BlockThematicBreak
	BlockHTMLBlock
	BlockTable
	BlockUnknown
)

func (k BlockKind) String() string {
	switch k {
	case BlockParagraph:
		return "Paragraph"
	case BlockHeading:
		return "Heading"
	case BlockFencedCode:
		return "FencedCode"
	case BlockList:
		return "List"
	case BlockBlockquote:
		return "Blockquote"
	case BlockThematicBreak:
		return "ThematicBreak"
	case BlockHTMLBlock:
		return "HTMLBlock"
	case BlockTable:
		return "Table"
	default:
		return "Unknown"
	}
}

// Block represents a top-level markdown block element.
type Block struct {
	Kind      BlockKind
	StartLine int // inclusive, 0-indexed source line
	EndLine   int // inclusive, 0-indexed source line
	RawText   string
}

// Document holds the parsed block structure of a markdown document.
type Document struct {
	blocks []Block
	source []byte
}

// NewDocument parses source text into blocks.
func NewDocument(content string) *Document {
	doc := &Document{
		source: []byte(content),
	}
	doc.parse()
	return doc
}

// Blocks returns all blocks.
func (d *Document) Blocks() []Block {
	return d.blocks
}

// BlockAt returns the block containing the given source line, or nil.
func (d *Document) BlockAt(sourceLine int) *Block {
	for i := range d.blocks {
		if sourceLine >= d.blocks[i].StartLine && sourceLine <= d.blocks[i].EndLine {
			return &d.blocks[i]
		}
	}
	return nil
}

// BlockIndexAt returns the index of the block containing the given source line, or -1.
func (d *Document) BlockIndexAt(sourceLine int) int {
	for i := range d.blocks {
		if sourceLine >= d.blocks[i].StartLine && sourceLine <= d.blocks[i].EndLine {
			return i
		}
	}
	return -1
}

func (d *Document) parse() {
	d.blocks = nil
	if len(d.source) == 0 {
		return
	}

	md := goldmark.New()
	reader := text.NewReader(d.source)
	tree := md.Parser().Parse(reader)

	lines := strings.Split(string(d.source), "\n")

	// Walk top-level children of the document
	for child := tree.FirstChild(); child != nil; child = child.NextSibling() {
		block := d.nodeToBlock(child, lines)
		if block != nil {
			d.blocks = append(d.blocks, *block)
		}
	}

	// Fill gaps between blocks as implicit paragraphs
	d.fillGaps(lines)

	// Sort blocks by start line (should already be sorted)
	d.sortBlocks()
}

func (d *Document) nodeToBlock(node ast.Node, lines []string) *Block {
	kind := nodeKind(node)

	// Get line range from node segments
	startLine, endLine := d.nodeLineRange(node)
	if startLine < 0 || endLine < 0 {
		return nil
	}

	// Clamp endLine
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Build raw text
	rawLines := lines[startLine : endLine+1]
	rawText := strings.Join(rawLines, "\n")

	return &Block{
		Kind:      kind,
		StartLine: startLine,
		EndLine:   endLine,
		RawText:   rawText,
	}
}

func (d *Document) nodeLineRange(node ast.Node) (int, int) {
	startLine := -1
	endLine := -1

	// Helper to update line range from a node's segments (block nodes only)
	updateFromNode := func(n ast.Node) {
		// Only block nodes have Lines(); inline nodes panic.
		if n.Type() != ast.TypeBlock {
			return
		}
		segs := n.Lines()
		for i := 0; i < segs.Len(); i++ {
			seg := segs.At(i)
			sl := bytes.Count(d.source[:seg.Start], []byte("\n"))
			el := bytes.Count(d.source[:seg.Stop], []byte("\n"))
			if seg.Stop > 0 && d.source[seg.Stop-1] == '\n' {
				el--
			}
			if startLine < 0 || sl < startLine {
				startLine = sl
			}
			if el > endLine {
				endLine = el
			}
			if sl > endLine {
				endLine = sl
			}
		}
		// Also check inline children for text segments
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Type() == ast.TypeInline {
				if txt, ok := c.(*ast.Text); ok {
					seg := txt.Segment
					sl := bytes.Count(d.source[:seg.Start], []byte("\n"))
					el := bytes.Count(d.source[:seg.Stop], []byte("\n"))
					if seg.Stop > 0 && d.source[seg.Stop-1] == '\n' {
						el--
					}
					if startLine < 0 || sl < startLine {
						startLine = sl
					}
					if el > endLine {
						endLine = el
					}
				}
			}
		}
	}

	// Try to get line info from segments
	if node.HasChildren() {
		// Walk all block-level descendants to find the full range
		ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			// Skip inline nodes entirely to avoid panic
			if n.Type() == ast.TypeInline {
				return ast.WalkSkipChildren, nil
			}
			updateFromNode(n)
			return ast.WalkContinue, nil
		})
	} else {
		updateFromNode(node)
	}

	// For fenced code blocks, extend to include the fence markers
	if node.Kind() == ast.KindFencedCodeBlock {
		fcb := node.(*ast.FencedCodeBlock)
		// The node lines only include the content, not the fences.
		// We need to find the opening and closing fences.
		if startLine > 0 {
			startLine-- // Opening fence is line before content
		} else if startLine == 0 {
			// Content starts at line 0 means fence might be missing,
			// but usually goldmark gives us content after the fence
		}

		// Check if there's a closing fence after endLine
		totalLines := bytes.Count(d.source, []byte("\n"))
		if len(d.source) > 0 && d.source[len(d.source)-1] != '\n' {
			totalLines++
		}
		_ = fcb
		if endLine+1 < totalLines {
			closeLine := endLine + 1
			allLines := strings.Split(string(d.source), "\n")
			if closeLine < len(allLines) {
				trimmed := strings.TrimSpace(allLines[closeLine])
				if trimmed == "```" || trimmed == "~~~" || strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
					endLine = closeLine
				}
			}
		}
	}

	// For thematic breaks with no lines/segments, try raw content
	if startLine < 0 && node.Kind() == ast.KindThematicBreak {
		// Use the node's text position
		if node.Lines().Len() == 0 {
			// ThematicBreak may not have lines; use siblings to estimate
			prev := node.PreviousSibling()
			next := node.NextSibling()
			if prev != nil {
				_, prevEnd := d.nodeLineRange(prev)
				if prevEnd >= 0 {
					startLine = prevEnd + 1
					endLine = startLine
				}
			} else if next != nil {
				nextStart, _ := d.nodeLineRange(next)
				if nextStart > 0 {
					startLine = nextStart - 1
					endLine = startLine
				}
			}
		}
	}

	return startLine, endLine
}

func nodeKind(node ast.Node) BlockKind {
	switch node.Kind() {
	case ast.KindParagraph:
		return BlockParagraph
	case ast.KindHeading:
		return BlockHeading
	case ast.KindFencedCodeBlock, ast.KindCodeBlock:
		return BlockFencedCode
	case ast.KindList:
		return BlockList
	case ast.KindBlockquote:
		return BlockBlockquote
	case ast.KindThematicBreak:
		return BlockThematicBreak
	case ast.KindHTMLBlock:
		return BlockHTMLBlock
	default:
		return BlockUnknown
	}
}

func (d *Document) fillGaps(lines []string) {
	if len(d.blocks) == 0 {
		// Entire document is one implicit block
		if len(lines) > 0 {
			d.blocks = append(d.blocks, Block{
				Kind:      BlockParagraph,
				StartLine: 0,
				EndLine:   len(lines) - 1,
				RawText:   strings.Join(lines, "\n"),
			})
		}
		return
	}

	var filled []Block

	// Gap before first block
	if d.blocks[0].StartLine > 0 {
		gapEnd := d.blocks[0].StartLine - 1
		raw := strings.Join(lines[0:gapEnd+1], "\n")
		if strings.TrimSpace(raw) != "" {
			filled = append(filled, Block{
				Kind:      BlockParagraph,
				StartLine: 0,
				EndLine:   gapEnd,
				RawText:   raw,
			})
		}
	}

	for i, blk := range d.blocks {
		filled = append(filled, blk)

		// Gap between this block and next
		if i < len(d.blocks)-1 {
			gapStart := blk.EndLine + 1
			gapEnd := d.blocks[i+1].StartLine - 1
			if gapStart <= gapEnd {
				raw := strings.Join(lines[gapStart:gapEnd+1], "\n")
				if strings.TrimSpace(raw) != "" {
					filled = append(filled, Block{
						Kind:      BlockParagraph,
						StartLine: gapStart,
						EndLine:   gapEnd,
						RawText:   raw,
					})
				}
			}
		}
	}

	// Gap after last block
	lastEnd := d.blocks[len(d.blocks)-1].EndLine
	if lastEnd < len(lines)-1 {
		gapStart := lastEnd + 1
		raw := strings.Join(lines[gapStart:], "\n")
		if strings.TrimSpace(raw) != "" {
			filled = append(filled, Block{
				Kind:      BlockParagraph,
				StartLine: gapStart,
				EndLine:   len(lines) - 1,
				RawText:   raw,
			})
		}
	}

	d.blocks = filled
}

func (d *Document) sortBlocks() {
	// Simple insertion sort since blocks should be nearly sorted
	for i := 1; i < len(d.blocks); i++ {
		j := i
		for j > 0 && d.blocks[j].StartLine < d.blocks[j-1].StartLine {
			d.blocks[j], d.blocks[j-1] = d.blocks[j-1], d.blocks[j]
			j--
		}
	}
}
