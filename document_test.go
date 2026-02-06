package main

import (
	"testing"
)

func TestDocumentParseHeadingAndParagraph(t *testing.T) {
	content := "# Hello\n\nThis is a paragraph."
	doc := NewDocument(content)
	blocks := doc.Blocks()

	if len(blocks) == 0 {
		t.Fatal("expected blocks, got none")
	}

	// Should have at least a heading and a paragraph
	foundHeading := false
	foundParagraph := false
	for _, b := range blocks {
		if b.Kind == BlockHeading {
			foundHeading = true
		}
		if b.Kind == BlockParagraph {
			foundParagraph = true
		}
	}
	if !foundHeading {
		t.Error("expected a heading block")
	}
	if !foundParagraph {
		t.Error("expected a paragraph block")
	}
}

func TestDocumentParseFencedCode(t *testing.T) {
	content := "# Title\n\n```go\nfmt.Println(\"hello\")\n```\n\nAfter code."
	doc := NewDocument(content)
	blocks := doc.Blocks()

	foundCode := false
	for _, b := range blocks {
		if b.Kind == BlockFencedCode {
			foundCode = true
		}
	}
	if !foundCode {
		t.Error("expected a fenced code block")
	}
}

func TestDocumentBlockAt(t *testing.T) {
	content := "# Hello\n\nParagraph text."
	doc := NewDocument(content)

	block := doc.BlockAt(0)
	if block == nil {
		t.Fatal("expected block at line 0")
	}
	if block.Kind != BlockHeading {
		t.Errorf("expected heading at line 0, got %s", block.Kind)
	}

	block = doc.BlockAt(2)
	if block == nil {
		t.Fatal("expected block at line 2")
	}
	if block.Kind != BlockParagraph {
		t.Errorf("expected paragraph at line 2, got %s", block.Kind)
	}
}

func TestDocumentParseList(t *testing.T) {
	content := "- item 1\n- item 2\n- item 3"
	doc := NewDocument(content)
	blocks := doc.Blocks()

	foundList := false
	for _, b := range blocks {
		if b.Kind == BlockList {
			foundList = true
		}
	}
	if !foundList {
		t.Error("expected a list block")
	}
}

func TestDocumentBlockquote(t *testing.T) {
	content := "> This is a quote\n> continued"
	doc := NewDocument(content)
	blocks := doc.Blocks()

	foundQuote := false
	for _, b := range blocks {
		if b.Kind == BlockBlockquote {
			foundQuote = true
		}
	}
	if !foundQuote {
		t.Error("expected a blockquote block")
	}
}

func TestDocumentEmptyContent(t *testing.T) {
	doc := NewDocument("")
	blocks := doc.Blocks()
	if len(blocks) != 0 {
		t.Errorf("expected 0 blocks for empty content, got %d", len(blocks))
	}
}
