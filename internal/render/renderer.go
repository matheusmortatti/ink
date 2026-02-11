package render

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/matheusmortatti/ink/internal/block"
)

// Renderer wraps Glamour for block-level markdown rendering.
type Renderer struct {
	glamour *glamour.TermRenderer
	width   int
}

// NewRenderer creates a renderer with the given terminal width.
func NewRenderer(width int) (*Renderer, error) {
	if width <= 0 {
		return nil, fmt.Errorf("width must be positive, got %d", width)
	}
	tr, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	return &Renderer{glamour: tr, width: width}, nil
}

// Render converts a markdown block to ANSI-styled output via Glamour.
func (r *Renderer) Render(b block.Block) (string, error) {
	if r.glamour == nil {
		return "", errors.New("renderer not initialized")
	}
	raw := b.Raw
	out, err := r.glamour.Render(raw)
	if err != nil {
		return "", err
	}
	// Glamour adds leading/trailing newlines for document spacing.
	// Trim outer newlines so blocks compose cleanly in the viewport.
	out = strings.Trim(out, "\n")
	return out, nil
}

// RenderCached checks the cache first, renders on miss, and stores the result.
func (r *Renderer) RenderCached(b block.Block, cache *RenderCache) (string, error) {
	if cached, ok := cache.Get(b, r.width); ok {
		return cached, nil
	}
	rendered, err := r.Render(b)
	if err != nil {
		return "", err
	}
	cache.Put(b, r.width, rendered)
	return rendered, nil
}

// SetWidth updates the renderer's word wrap width.
// Glamour renderers are immutable regarding width, so this creates
// a new internal Glamour instance.
func (r *Renderer) SetWidth(width int) error {
	if width <= 0 {
		return fmt.Errorf("width must be positive, got %d", width)
	}
	tr, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return err
	}
	r.glamour = tr
	r.width = width
	return nil
}

// PreRenderAll renders all blocks and populates the cache.
// Continues on error so partial success is possible.
func (r *Renderer) PreRenderAll(blocks []block.Block, cache *RenderCache) error {
	var errs []error
	for _, b := range blocks {
		out, err := r.Render(b)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		cache.Put(b, r.width, out)
	}
	return errors.Join(errs...)
}
