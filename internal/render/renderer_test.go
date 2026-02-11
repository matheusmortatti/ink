package render

import (
	"strings"
	"testing"

	"github.com/matheusmortatti/ink/internal/block"
)

func TestNewRenderer_ValidWidth(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer(80) returned error: %v", err)
	}
	if r == nil {
		t.Fatal("NewRenderer(80) returned nil")
	}
	if r.width != 80 {
		t.Errorf("width = %d, want 80", r.width)
	}
}

func TestNewRenderer_SmallWidth(t *testing.T) {
	r, err := NewRenderer(20)
	if err != nil {
		t.Fatalf("NewRenderer(20) returned error: %v", err)
	}
	if r == nil {
		t.Fatal("NewRenderer(20) returned nil")
	}
}

func TestNewRenderer_InvalidWidth(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRenderer(tt.width)
			if err == nil {
				t.Errorf("NewRenderer(%d) should return error", tt.width)
			}
		})
	}
}

func TestRenderer_AllBlockTypes_RenderCorrectly(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	tests := []struct {
		name     string
		block    block.Block
		contains string
	}{
		{"H1", block.Block{Type: block.Heading, Raw: "# Title", Level: 1}, "Title"},
		{"H2", block.Block{Type: block.Heading, Raw: "## Subtitle", Level: 2}, "Subtitle"},
		{"H3", block.Block{Type: block.Heading, Raw: "### Section", Level: 3}, "Section"},
		{"H4", block.Block{Type: block.Heading, Raw: "#### Subsection", Level: 4}, "Subsection"},
		{"H5", block.Block{Type: block.Heading, Raw: "##### Minor", Level: 5}, "Minor"},
		{"H6", block.Block{Type: block.Heading, Raw: "###### Tiny", Level: 6}, "Tiny"},
		{"paragraph", block.Block{Type: block.Paragraph, Raw: "Hello world"}, "Hello world"},
		{"bold", block.Block{Type: block.Paragraph, Raw: "This is **bold** text"}, "bold"},
		{"italic", block.Block{Type: block.Paragraph, Raw: "This is *italic* text"}, "italic"},
		{"link", block.Block{Type: block.Paragraph, Raw: "Visit [example](http://example.com)"}, "example"},
		{"code_span", block.Block{Type: block.Paragraph, Raw: "Use `fmt.Println` here"}, "fmt.Println"},
		{"unordered_list", block.Block{Type: block.List, Raw: "- item one\n- item two\n- item three"}, "item one"},
		{"ordered_list", block.Block{Type: block.List, Raw: "1. first\n2. second\n3. third"}, "first"},
		{"code_fence", block.Block{Type: block.CodeFence, Raw: "```go\nfmt.Println(\"hello\")\n```"}, "hello"},
		{"code_block", block.Block{Type: block.CodeBlock, Raw: "    indented code"}, "indented code"},
		{"blockquote", block.Block{Type: block.BlockQuote, Raw: "> This is a quote"}, "This is a quote"},
		{"table", block.Block{Type: block.Table, Raw: "| A | B |\n|---|---|\n| 1 | 2 |"}, "1"},
		{"horizontal_rule", block.Block{Type: block.HorizontalRule, Raw: "---"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := r.Render(tt.block)
			if err != nil {
				t.Fatalf("Render(%s) error: %v", tt.name, err)
			}
			if rendered == "" {
				t.Errorf("Render(%s) returned empty string", tt.name)
			}
			if tt.contains != "" && !strings.Contains(rendered, tt.contains) {
				t.Errorf("Render(%s) = %q, want to contain %q", tt.name, rendered, tt.contains)
			}
			// Verify rendering transforms the input (not just passthrough)
			if tt.name != "horizontal_rule" && rendered == tt.block.Raw {
				t.Errorf("Render(%s) returned raw input unchanged — expected Glamour transformation", tt.name)
			}
		})
	}
}

func TestRenderer_DifferentWidths_DifferentOutput(t *testing.T) {
	longText := "This is a very long paragraph that should wrap differently at different terminal widths because it contains enough text to exceed the narrow width."
	b := block.Block{Type: block.Paragraph, Raw: longText}

	r80, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer(80): %v", err)
	}
	r40, err := NewRenderer(40)
	if err != nil {
		t.Fatalf("NewRenderer(40): %v", err)
	}

	out80, err := r80.Render(b)
	if err != nil {
		t.Fatalf("Render at 80: %v", err)
	}
	out40, err := r40.Render(b)
	if err != nil {
		t.Fatalf("Render at 40: %v", err)
	}

	if out80 == out40 {
		t.Error("expected different output at different widths, but got identical output")
	}
}

func TestRenderer_SetWidth(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	err = r.SetWidth(120)
	if err != nil {
		t.Fatalf("SetWidth(120): %v", err)
	}
	if r.width != 120 {
		t.Errorf("width = %d, want 120", r.width)
	}
}

func TestRenderer_SetWidth_ChangesOutput(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	longText := "This is a very long paragraph that should wrap differently at different terminal widths because it contains enough text to exceed the narrow width."
	b := block.Block{Type: block.Paragraph, Raw: longText}

	out80, err := r.Render(b)
	if err != nil {
		t.Fatalf("Render at 80: %v", err)
	}

	err = r.SetWidth(40)
	if err != nil {
		t.Fatalf("SetWidth(40): %v", err)
	}

	out40, err := r.Render(b)
	if err != nil {
		t.Fatalf("Render at 40: %v", err)
	}

	if out80 == out40 {
		t.Error("expected different output after SetWidth, but got identical output")
	}
}

func TestSetWidth_InvalidWidth(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	err = r.SetWidth(0)
	if err == nil {
		t.Fatal("SetWidth(0) should return error")
	}
	err = r.SetWidth(-1)
	if err == nil {
		t.Fatal("SetWidth(-1) should return error")
	}
	// Original width should be preserved after failed SetWidth
	if r.width != 80 {
		t.Errorf("width changed to %d after failed SetWidth, want 80", r.width)
	}
}

func TestRenderer_NilGlamour_ReturnsError(t *testing.T) {
	r := &Renderer{width: 80} // no glamour instance
	b := block.Block{Type: block.Paragraph, Raw: "test"}
	_, err := r.Render(b)
	if err == nil {
		t.Fatal("expected error from nil glamour renderer")
	}
}

func TestRenderer_EmptyRaw(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	b := block.Block{Type: block.Paragraph, Raw: ""}
	// Glamour handles empty input gracefully
	_, err = r.Render(b)
	if err != nil {
		t.Fatalf("Render with empty raw should not error: %v", err)
	}
}

func TestPreRenderAll_PopulatesCache(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := NewRenderCache()
	blocks := []block.Block{
		{Type: block.Heading, Raw: "# Hello", Level: 1},
		{Type: block.Paragraph, Raw: "Some text here"},
		{Type: block.List, Raw: "- a\n- b\n- c"},
	}

	err = r.PreRenderAll(blocks, cache)
	if err != nil {
		t.Fatalf("PreRenderAll: %v", err)
	}

	for _, b := range blocks {
		if _, ok := cache.Get(b, 80); !ok {
			t.Errorf("block %q not found in cache after PreRenderAll", b.Raw)
		}
	}
}

func TestPreRenderAll_ContinuesOnError(t *testing.T) {
	// Use a renderer with nil glamour to force render errors.
	// This verifies PreRenderAll processes all blocks and collects
	// all errors via errors.Join rather than stopping at the first.
	r := &Renderer{width: 80}
	cache := NewRenderCache()
	blocks := []block.Block{
		{Type: block.Heading, Raw: "# Block 1", Level: 1},
		{Type: block.Paragraph, Raw: "Block 2"},
		{Type: block.List, Raw: "- Block 3"},
	}

	err := r.PreRenderAll(blocks, cache)
	if err == nil {
		t.Fatal("expected error from uninitialized renderer")
	}
	// Verify all errors were collected (continues past first failure)
	errStr := err.Error()
	errCount := strings.Count(errStr, "renderer not initialized")
	if errCount != len(blocks) {
		t.Errorf("expected %d errors collected, got %d (error: %v)", len(blocks), errCount, err)
	}
	// Cache should be empty since all renders failed
	for _, b := range blocks {
		if _, ok := cache.Get(b, 80); ok {
			t.Errorf("block %q should not be in cache after failed render", b.Raw)
		}
	}
}

func TestPreRenderAll_Empty(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := NewRenderCache()

	err = r.PreRenderAll(nil, cache)
	if err != nil {
		t.Fatalf("PreRenderAll(nil): %v", err)
	}
}

func TestRenderCached_Hit(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := NewRenderCache()
	b := block.Block{Type: block.Paragraph, Raw: "cached content"}

	// First call should render and populate cache
	out1, err := r.RenderCached(b, cache)
	if err != nil {
		t.Fatalf("RenderCached: %v", err)
	}
	if _, ok := cache.Get(b, 80); !ok {
		t.Error("expected block in cache after RenderCached")
	}

	// Second call should hit cache and return identical result
	out2, err := r.RenderCached(b, cache)
	if err != nil {
		t.Fatalf("RenderCached (cached): %v", err)
	}
	if out1 != out2 {
		t.Error("cached result differs from original render")
	}
}

func TestRenderCached_Miss(t *testing.T) {
	r, err := NewRenderer(80)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	cache := NewRenderCache()
	b1 := block.Block{Type: block.Paragraph, Raw: "first block"}
	b2 := block.Block{Type: block.Paragraph, Raw: "second block"}

	// Cache b1
	_, err = r.RenderCached(b1, cache)
	if err != nil {
		t.Fatalf("RenderCached b1: %v", err)
	}

	// b2 should miss cache but still return valid result
	out, err := r.RenderCached(b2, cache)
	if err != nil {
		t.Fatalf("RenderCached b2: %v", err)
	}
	if out == "" {
		t.Error("RenderCached should return non-empty output on cache miss")
	}
	// b2 should now be cached
	if _, ok := cache.Get(b2, 80); !ok {
		t.Error("b2 should be in cache after RenderCached miss")
	}
}
