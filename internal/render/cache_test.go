package render

import (
	"sync"
	"testing"

	"github.com/matheusmortatti/ink/internal/block"
)

func TestNewRenderCache(t *testing.T) {
	c := NewRenderCache()
	if c == nil {
		t.Fatal("NewRenderCache returned nil")
	}
}

func TestRenderCache_PutAndGet_Hit(t *testing.T) {
	c := NewRenderCache()
	b := block.Block{Type: block.Paragraph, Raw: "hello world"}

	c.Put(b, 80, "rendered output")
	got, ok := c.Get(b, 80)
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if got != "rendered output" {
		t.Errorf("got %q, want %q", got, "rendered output")
	}
}

func TestRenderCache_Get_Miss(t *testing.T) {
	c := NewRenderCache()
	b := block.Block{Type: block.Paragraph, Raw: "hello world"}

	_, ok := c.Get(b, 80)
	if ok {
		t.Fatal("expected cache miss, got hit")
	}
}

func TestRenderCache_SameContentSameWidth_Hit(t *testing.T) {
	c := NewRenderCache()
	b1 := block.Block{Type: block.Paragraph, Raw: "same content"}
	b2 := block.Block{Type: block.Paragraph, Raw: "same content"}

	c.Put(b1, 80, "rendered")
	got, ok := c.Get(b2, 80)
	if !ok {
		t.Fatal("same content should hit cache")
	}
	if got != "rendered" {
		t.Errorf("got %q, want %q", got, "rendered")
	}
}

func TestRenderCache_SameContentDifferentWidth_Miss(t *testing.T) {
	c := NewRenderCache()
	b := block.Block{Type: block.Paragraph, Raw: "hello world"}

	c.Put(b, 80, "rendered at 80")
	_, ok := c.Get(b, 120)
	if ok {
		t.Fatal("different width should miss cache")
	}
}

func TestRenderCache_DifferentContent_Miss(t *testing.T) {
	c := NewRenderCache()
	b1 := block.Block{Type: block.Paragraph, Raw: "content A"}
	b2 := block.Block{Type: block.Paragraph, Raw: "content B"}

	c.Put(b1, 80, "rendered A")
	_, ok := c.Get(b2, 80)
	if ok {
		t.Fatal("different content should miss cache")
	}
}

func TestRenderCache_InvalidateAll(t *testing.T) {
	c := NewRenderCache()
	b1 := block.Block{Type: block.Paragraph, Raw: "block 1"}
	b2 := block.Block{Type: block.Heading, Raw: "# heading", Level: 1}

	c.Put(b1, 80, "rendered 1")
	c.Put(b2, 80, "rendered 2")

	c.InvalidateAll()

	if _, ok := c.Get(b1, 80); ok {
		t.Error("expected miss after InvalidateAll for block 1")
	}
	if _, ok := c.Get(b2, 80); ok {
		t.Error("expected miss after InvalidateAll for block 2")
	}
}

func TestRenderCache_InvalidateBlock(t *testing.T) {
	c := NewRenderCache()
	b1 := block.Block{Type: block.Paragraph, Raw: "block 1"}
	b2 := block.Block{Type: block.Heading, Raw: "# heading", Level: 1}

	c.Put(b1, 80, "rendered 1")
	c.Put(b1, 120, "rendered 1 wide")
	c.Put(b2, 80, "rendered 2")

	c.InvalidateBlock(b1)

	if _, ok := c.Get(b1, 80); ok {
		t.Error("expected miss for invalidated block at width 80")
	}
	if _, ok := c.Get(b1, 120); ok {
		t.Error("expected miss for invalidated block at width 120")
	}
	if _, ok := c.Get(b2, 80); !ok {
		t.Error("expected hit for non-invalidated block")
	}
}

func TestRenderCache_ConcurrentAccess(t *testing.T) {
	c := NewRenderCache()
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b := block.Block{Type: block.Paragraph, Raw: "concurrent block"}
			c.Put(b, i, "rendered")
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b := block.Block{Type: block.Paragraph, Raw: "concurrent block"}
			c.Get(b, i)
		}(i)
	}

	wg.Wait()
}

func TestRenderCache_ConcurrentReadWrite(t *testing.T) {
	c := NewRenderCache()
	b := block.Block{Type: block.Paragraph, Raw: "shared block"}
	c.Put(b, 80, "initial")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Get(b, 80)
		}()
		go func() {
			defer wg.Done()
			c.Put(b, 80, "updated")
		}()
	}
	wg.Wait()
}

func TestHashContent(t *testing.T) {
	tests := []struct {
		name     string
		a, b     string
		wantSame bool
	}{
		{"same content", "hello", "hello", true},
		{"different content", "hello", "world", false},
		{"empty vs non-empty", "", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ha := hashContent(tt.a)
			hb := hashContent(tt.b)
			if (ha == hb) != tt.wantSame {
				t.Errorf("hashContent(%q) = %q, hashContent(%q) = %q, wantSame = %v",
					tt.a, ha, tt.b, hb, tt.wantSame)
			}
		})
	}
}

func TestHashContent_Deterministic(t *testing.T) {
	content := "# Hello World\n\nSome paragraph text."
	h1 := hashContent(content)
	h2 := hashContent(content)
	if h1 != h2 {
		t.Errorf("hash not deterministic: %q != %q", h1, h2)
	}
}

func TestRenderCache_WidthSpecificCaching(t *testing.T) {
	c := NewRenderCache()
	b := block.Block{Type: block.Paragraph, Raw: "test content"}

	c.Put(b, 80, "at 80")
	c.Put(b, 120, "at 120")
	c.Put(b, 40, "at 40")

	tests := []struct {
		width int
		want  string
	}{
		{80, "at 80"},
		{120, "at 120"},
		{40, "at 40"},
	}

	for _, tt := range tests {
		got, ok := c.Get(b, tt.width)
		if !ok {
			t.Errorf("cache miss at width %d", tt.width)
			continue
		}
		if got != tt.want {
			t.Errorf("at width %d: got %q, want %q", tt.width, got, tt.want)
		}
	}
}
