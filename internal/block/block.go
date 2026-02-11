package block

// BlockType represents the type of a markdown block element.
type BlockType int

const (
	Paragraph BlockType = iota
	Heading
	List
	CodeFence
	CodeBlock
	BlockQuote
	Table
	HorizontalRule
)

// String returns the string representation of a BlockType.
func (bt BlockType) String() string {
	switch bt {
	case Paragraph:
		return "Paragraph"
	case Heading:
		return "Heading"
	case List:
		return "List"
	case CodeFence:
		return "CodeFence"
	case CodeBlock:
		return "CodeBlock"
	case BlockQuote:
		return "BlockQuote"
	case Table:
		return "Table"
	case HorizontalRule:
		return "HorizontalRule"
	default:
		return "Unknown"
	}
}

// Block represents a single parsed markdown block element.
type Block struct {
	Type      BlockType
	Raw       string // Original markdown source text
	Level     int    // Heading level (1-6), 0 for non-headings
	StartByte int    // Start position in original source
	EndByte   int    // End position in original source (exclusive)
}
