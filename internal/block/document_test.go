package block

import "testing"

func TestDocument_Serialize_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"two paragraphs", "Hello\n\nWorld"},
		{"heading and paragraph", "# Title\n\nBody text"},
		{"multiple block types", "# Title\n\nParagraph text\n\n- item 1\n- item 2"},
		{"code fence", "Before\n\n```go\nfunc main() {}\n```\n\nAfter"},
		{"blockquote", "Text\n\n> quoted\n> text\n\nMore text"},
		{"horizontal rule", "Above\n\n---\n\nBelow"},
		{"table", "| A | B |\n| --- | --- |\n| 1 | 2 |"},
		{"heading, paragraph, code, list", "# Title\n\nIntro\n\n```\ncode\n```\n\n- item"},
		{"trailing newline", "Hello\n\nWorld\n"},
		{"trailing double newline", "Hello\n\nWorld\n\n"},
		{"single block trailing newline", "Hello\n"},
		{"trailing spaces in paragraph", "Hello  \n\nWorld"},
		{"empty code fence", "Before\n\n```\n```\n\nAfter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := NewDocument([]byte(tt.input))
			output := doc.Serialize()
			if output != tt.input {
				t.Errorf("round-trip failed:\ninput:  %q\noutput: %q", tt.input, output)
			}
		})
	}
}

func TestDocument_Serialize_Empty(t *testing.T) {
	doc := NewDocument([]byte(""))
	if s := doc.Serialize(); s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestDocument_Serialize_NilInput(t *testing.T) {
	doc := NewDocument(nil)
	if s := doc.Serialize(); s != "" {
		t.Errorf("expected empty string, got %q", s)
	}
}

func TestDocument_Blocks(t *testing.T) {
	doc := NewDocument([]byte("# Title\n\nBody\n\n- item"))
	if len(doc.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(doc.Blocks))
	}
	if doc.Blocks[0].Type != Heading {
		t.Errorf("block 0: got %s, want Heading", doc.Blocks[0].Type)
	}
	if doc.Blocks[1].Type != Paragraph {
		t.Errorf("block 1: got %s, want Paragraph", doc.Blocks[1].Type)
	}
	if doc.Blocks[2].Type != List {
		t.Errorf("block 2: got %s, want List", doc.Blocks[2].Type)
	}
}

func TestDocument_Serialize_SingleBlock(t *testing.T) {
	doc := NewDocument([]byte("Just one paragraph"))
	output := doc.Serialize()
	if output != "Just one paragraph" {
		t.Errorf("got %q, want %q", output, "Just one paragraph")
	}
}
