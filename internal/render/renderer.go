package render

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/matheusmortatti/ink/internal/block"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

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
	out = stripRenderPadding(out)
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

// stripRenderPadding removes Glamour's base indent and trailing whitespace
// padding from rendered output. Glamour adds a fixed leading indent (typically
// 2 spaces) and pads lines with trailing spaces to fill the column width.
// This function is ANSI-aware: it strips ANSI codes to detect padding, then
// manipulates the original ANSI-containing string to preserve styling.
func stripRenderPadding(rendered string) string {
	if rendered == "" {
		return rendered
	}
	lines := strings.Split(rendered, "\n")

	// Strip ANSI to get clean text for indent detection.
	cleanLines := make([]string, len(lines))
	for i, line := range lines {
		cleanLines[i] = ansiRe.ReplaceAllString(line, "")
	}

	// Find minimum leading whitespace across non-blank clean lines.
	minIndent := math.MaxInt
	for _, clean := range cleanLines {
		if strings.TrimSpace(clean) == "" {
			continue
		}
		indent := len(clean) - len(strings.TrimLeft(clean, " "))
		if indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent == math.MaxInt {
		minIndent = 0
	}

	// Strip padding from each line, preserving ANSI codes.
	for i, line := range lines {
		if strings.TrimSpace(cleanLines[i]) == "" {
			lines[i] = ""
			continue
		}
		lines[i] = stripLeadingVisibleSpaces(line, minIndent)
		lines[i] = trimTrailingVisibleSpaces(lines[i])
	}

	return strings.Join(lines, "\n")
}

// stripLeadingVisibleSpaces removes the first n visible space characters from s,
// skipping over (and dropping) any ANSI escape sequences that precede those spaces.
func stripLeadingVisibleSpaces(s string, n int) string {
	if n <= 0 {
		return s
	}
	stripped := 0
	i := 0
	for i < len(s) && stripped < n {
		if i+1 < len(s) && s[i] == '\x1b' && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && (s[j] >= '0' && s[j] <= '9' || s[j] == ';') {
				j++
			}
			if j < len(s) {
				j++ // skip terminator
			}
			i = j
			continue
		}
		if s[i] == ' ' {
			stripped++
			i++
			continue
		}
		break
	}
	return s[i:]
}

// trimTrailingVisibleSpaces removes trailing visible whitespace from s,
// walking backwards past ANSI escape sequences and space characters.
func trimTrailingVisibleSpaces(s string) string {
	end := len(s)
	hadAnsi := false
	for end > 0 {
		// Check for trailing ANSI sequence: ESC [ digits/semicolons letter
		if end >= 3 {
			j := end - 1
			if (s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= 'a' && s[j] <= 'z') {
				k := j - 1
				for k >= 0 && (s[k] >= '0' && s[k] <= '9' || s[k] == ';') {
					k--
				}
				if k >= 1 && s[k] == '[' && s[k-1] == '\x1b' {
					end = k - 1
					hadAnsi = true
					continue
				}
			}
		}
		if s[end-1] == ' ' {
			end--
			continue
		}
		break
	}
	if end == len(s) {
		return s // nothing trimmed
	}
	if end <= 0 {
		return ""
	}
	result := s[:end]
	if hadAnsi {
		result += "\x1b[0m"
	}
	return result
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
