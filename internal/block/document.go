package block

import "strings"

// Document represents a parsed markdown document as a sequence of blocks.
type Document struct {
	Blocks []Block
	suffix string // trailing content after last block (e.g., trailing newline)
}

// NewDocument parses source markdown into a Document.
func NewDocument(source []byte) *Document {
	blocks := Parse(source)
	doc := &Document{Blocks: blocks}

	// Capture trailing content after the last block for round-trip fidelity.
	if len(blocks) > 0 {
		lastEnd := blocks[len(blocks)-1].EndByte
		if lastEnd < len(source) {
			doc.suffix = string(source[lastEnd:])
		}
	}

	return doc
}

// Serialize reconstructs the markdown source from blocks.
// Joins block raw content with "\n\n" separators and appends
// any trailing content from the original source.
func (d *Document) Serialize() string {
	if len(d.Blocks) == 0 {
		return ""
	}
	parts := make([]string, len(d.Blocks))
	for i, b := range d.Blocks {
		parts[i] = b.Raw
	}
	return strings.Join(parts, "\n\n") + d.suffix
}
