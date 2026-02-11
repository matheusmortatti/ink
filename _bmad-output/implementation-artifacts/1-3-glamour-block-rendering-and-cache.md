# Story 1.3: Glamour Block Rendering and Cache

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->
<!-- Epic: View and Navigate a Beautiful Document -->
<!-- Story Key: 1-3-glamour-block-rendering-and-cache -->
<!-- Date: 2026-02-10 -->

## Story

As a writer,
I want each markdown block rendered beautifully via Glamour with results cached for instant retrieval,
so that the document looks polished and rendering is fast.

## Acceptance Criteria

1. **Given** a `Block` containing any supported markdown element (heading, paragraph, bold, italic, links, code spans, block quotes, lists, tables, horizontal rules, code fences) **When** the block is rendered via Glamour **Then** the output matches glow-quality rendering with the adaptive dark/light theme (FR4)

2. **Given** a rendered block result **When** it is stored in the render cache keyed by (content hash, terminal width) **Then** subsequent requests for the same block at the same width return the cached result without re-rendering

3. **Given** a block whose content has changed **When** a render is requested **Then** the cache misses and the block is re-rendered with Glamour

4. **Given** a terminal width change **When** the cache is invalidated globally **Then** all blocks are marked for re-rendering at the new width

5. **Given** a `[]Block` document loaded from a file **When** all blocks are pre-rendered on load **Then** the entire cache is populated and ready for display

## Tasks / Subtasks

- [x] Task 1: Implement Glamour renderer wrapper (AC: #1)
  - [x] 1.1 Create `internal/render/renderer.go` with `Renderer` type
  - [x] 1.2 Initialize single `glamour.TermRenderer` instance with `WithAutoStyle()` and `WithWordWrap(width)`
  - [x] 1.3 Implement `Render(block Block, width int) (string, error)` that converts Block to markdown and renders via Glamour
  - [x] 1.4 Handle block-level rendering workaround (Glamour renders full documents, need to wrap single blocks appropriately)
  - [x] 1.5 Add `SetWidth(width int)` method to update word wrap width

- [x] Task 2: Implement render cache (AC: #2, #3, #4)
  - [x] 2.1 Create `internal/render/cache.go` with `RenderCache` type
  - [x] 2.2 Define cache key structure: `(contentHash string, width int)`
  - [x] 2.3 Implement `Get(block Block, width int) (string, bool)` for cache lookups
  - [x] 2.4 Implement `Put(block Block, width int, rendered string)` for cache storage
  - [x] 2.5 Implement `InvalidateAll()` to clear all cache entries (for terminal resize)
  - [x] 2.6 Implement `InvalidateBlock(block Block)` to clear specific block cache (for content changes)
  - [x] 2.7 Add content hashing function using SHA256 or faster alternative

- [x] Task 3: Implement pre-render functionality (AC: #5)
  - [x] 3.1 Create `PreRenderAll(blocks []Block, width int) error` in `renderer.go`
  - [x] 3.2 Iterate through all blocks, render each via Glamour, and populate cache
  - [x] 3.3 Return aggregate error if any blocks fail to render (but continue processing others)

- [x] Task 4: Handle Glamour block-level rendering limitations
  - [x] 4.1 Research and test rendering isolated block markdown (wrap in minimal document if needed)
  - [x] 4.2 Verify heading, paragraph, list, code fence, blockquote, table, horizontal rule all render correctly in isolation
  - [x] 4.3 Add fallback/wrapper logic if Glamour requires document context for proper rendering

- [x] Task 5: Write comprehensive tests (AC: #1-#5)
  - [x] 5.1 Create `internal/render/renderer_test.go` with Glamour rendering tests
  - [x] 5.2 Test all markdown block types render correctly (headings H1-H6, paragraphs, bold, italic, links, code spans, lists, blockquotes, tables, code fences, horizontal rules)
  - [x] 5.3 Test rendering at different widths produces different output (reflow test)
  - [x] 5.4 Create `internal/render/cache_test.go` with cache behavior tests
  - [x] 5.5 Test cache hit/miss for same/different content
  - [x] 5.6 Test cache invalidation (global and per-block)
  - [x] 5.7 Test pre-render population
  - [x] 5.8 Test width-specific caching

## Dev Notes

### Why This Story Matters

This is the **make-or-break technical milestone** for the entire ink project. The <50ms block transition performance target (NFR2) depends entirely on this cache strategy working correctly. If Glamour rendering is too slow or the cache strategy is flawed, the core "Mode-Unified Block Reveal" experience fails.

**Critical dependencies:**
- **Consumed by**: `internal/ui/viewport` (Story 1.4) for displaying rendered blocks
- **Consumed by**: `internal/editor` (Story 2.4) for block transitions (insert ↔ normal mode)
- **Consumed by**: `internal/ui` (Story 6.1) for terminal resize handling (cache invalidation + re-render)

**Architectural role:** This package sits between the `block` parser (Story 1.2) and the TUI display layer. It transforms raw markdown blocks into beautiful ANSI-styled terminal output, with caching to ensure the transformation is near-instant on repeated access.

### Critical Design Decisions

**1. Single TermRenderer Instance (NOT per-block)**

From Glamour research: Creating a new `glamour.TermRenderer` for each render call is wasteful—each instance parses stylesheets and initializes Goldmark. The renderer MUST be created once and reused.

```go
// CORRECT: Create once, reuse many times
type Renderer struct {
    glamour *glamour.TermRenderer
    width   int
}

// WRONG: Creating new renderer per render call
func Render(block Block) string {
    r, _ := glamour.NewTermRenderer(...) // DON'T DO THIS
    return r.Render(block.Raw)
}
```

**2. Block-Level Rendering Workaround**

**Critical limitation discovered in research:** Glamour has NO public API for rendering isolated blocks. It's designed as a full-document renderer. This means we need a workaround.

**Recommended approach:**
- Wrap each block's raw markdown in minimal document context if needed
- Test extensively to ensure headings, lists, blockquotes, tables, etc. render correctly in isolation
- If wrapping breaks rendering quality, consider alternative: render full document and cache individual block outputs by extracting from rendered output

**Example test needed:**
```go
// Test that a heading block renders correctly in isolation
block := Block{Type: Heading, Raw: "# My Heading", Level: 1}
rendered := renderer.Render(block, 80)
// Should produce styled heading output, not fail or render incorrectly
```

**3. Cache Key Design: (contentHash, width)**

Cache entries MUST be keyed by BOTH content hash AND terminal width because:
- Same markdown renders differently at different widths (word wrap, table columns)
- Terminal resize (Story 6.1) invalidates ALL cached entries at the old width

**Content hashing:**
- Use `hash/fnv` (fast, built-in) or `crypto/sha256` (slower, collision-resistant)
- Hash the `Block.Raw` string content
- Store as hex string for readability in debugging

**Cache structure:**
```go
type RenderCache struct {
    entries map[string]string // key: "hash_width" → value: rendered output
    mu      sync.RWMutex      // concurrent access protection
}
```

**4. Pre-Rendering Strategy**

From Architecture: "Pre-render all blocks on load + cache per block" ensures the first display is instant. The `PreRenderAll` function should:
- Iterate through all blocks in the document
- Render each via Glamour
- Populate the cache
- **Continue on error**: If one block fails to render, log/collect error but continue rendering others (partial success better than total failure)

### Technical Requirements

**Go version:** Go 1.25+ (current go.mod)

**Glamour version:** v0.10.0 (latest, April 2025)
- Confirmed API: `glamour.NewTermRenderer(options...)`
- Options: `WithAutoStyle()`, `WithWordWrap(width)`, `WithStandardStyle(name)`
- Key improvement in v0.10.0: Better table rendering with link footers

**Glamour API patterns:**

```go
import "github.com/charmbracelet/glamour"

// Initialize renderer (ONCE per application)
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),      // Auto-detect dark/light terminal
    glamour.WithWordWrap(80),      // Set width
)

// Render markdown string to ANSI output
output, err := renderer.Render(markdownString)

// Update width (for terminal resize)
renderer, _ = glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(newWidth), // Updated width
)
```

**IMPORTANT:** Glamour does NOT support changing width on an existing renderer. Terminal resize requires creating a new renderer instance with the new width. This is acceptable because resize is infrequent.

**Hashing strategy:**

```go
import (
    "crypto/sha256"
    "encoding/hex"
)

func hashContent(raw string) string {
    h := sha256.New()
    h.Write([]byte(raw))
    return hex.EncodeToString(h.Sum(nil))
}

func cacheKey(contentHash string, width int) string {
    return fmt.Sprintf("%s_%d", contentHash, width)
}
```

**Alternative faster hash (if performance is critical):**

```go
import "hash/fnv"

func hashContent(raw string) string {
    h := fnv.New64a()
    h.Write([]byte(raw))
    return strconv.FormatUint(h.Sum64(), 16)
}
```

### Architecture Compliance

**Package: `internal/render`** — This package depends on `internal/block` (imports `Block` type)

**Dependency direction (MUST follow):**
```
internal/render → internal/block
internal/render → (NO other internal/ imports allowed)
```

- `internal/render` MUST import `internal/block` to access `Block` type
- `internal/render` MUST NOT import `internal/ui`, `internal/vim`, `internal/editor`, `internal/file`, or `internal/config`
- `internal/render` CAN import `github.com/charmbracelet/glamour` and standard library

**From Architecture (render package responsibilities):**
> internal/render — Glamour rendering, render cache, block compositing, color interpolation utilities. Depends on block for block data.

**Package boundary validation:**
- ✓ Importing `internal/block` — ALLOWED (leaf package)
- ✗ Importing `internal/ui` — FORBIDDEN (sibling package)
- ✓ Importing `glamour` — ALLOWED (external dependency)

**Naming conventions (enforce strictly):**
- Package name: `render` (singular, lowercase)
- No stutter: `render.Render` is fine, `render.RenderMarkdown` is not
- Exported types: `Renderer`, `RenderCache`
- Unexported helpers: `hashContent`, `cacheKey`, `wrapBlock` (if needed)
- Receiver names: single letter — `func (r *Renderer)`, `func (c *RenderCache)`

**Anti-patterns to AVOID:**
- Do NOT create `utils` or `helpers` package
- Do NOT import `internal/ui` or any sibling packages
- Do NOT add TUI logic — this is pure rendering, no Bubbletea
- Do NOT add file I/O — rendering operates on in-memory blocks only

### Library & Framework Requirements

| Library | Import Path | Version in go.mod | Usage in This Story |
|---|---|---|---|
| glamour | `github.com/charmbracelet/glamour` | v0.10.0 (latest) | Markdown → ANSI rendering with glow-quality output |
| crypto/sha256 | `crypto/sha256` | stdlib | Content hashing for cache keys |
| hash/fnv | `hash/fnv` | stdlib | (Alternative) Faster content hashing |
| encoding/hex | `encoding/hex` | stdlib | Convert hash bytes to hex string |
| sync | `sync` | stdlib | `sync.RWMutex` for concurrent cache access |

**IMPORTANT:** glamour is already a direct dependency from Story 1.1 initialization. It will NOT appear as `// indirect` in `go.mod`.

**WARNING — Common LLM mistakes with Glamour:**

1. **Creating new renderer per call** — Glamour renderers should be created ONCE and reused. Each creation parses stylesheets and initializes Goldmark (wasteful).

2. **Assuming block-level API exists** — Glamour has NO public API for rendering individual blocks. You must wrap blocks in minimal markdown or test extensively that isolated block rendering works.

3. **Forgetting width-specific caching** — Same markdown renders differently at different widths. Cache key MUST include width.

4. **Not handling rendering errors** — `renderer.Render()` returns `(string, error)`. Block rendering can fail on malformed input. Handle errors gracefully.

5. **Mutating renderer width** — Glamour renderers are immutable regarding width. Terminal resize requires creating a NEW renderer with the new width.

### File Structure Requirements

**Files to create in this story:**

```
internal/render/
├── renderer.go       # Renderer type, Render method, Glamour wrapper, SetWidth
├── renderer_test.go  # Rendering tests for all block types, width handling
├── cache.go          # RenderCache type, Get/Put/Invalidate methods, hashing
├── cache_test.go     # Cache hit/miss, invalidation, pre-render tests
└── color.go          # (Future) Color interpolation for dimming (Story 2.3, defer for now)
```

**Files NOT to create:**
- `color.go` — Deferred to Story 2.3 (syntax dimming)
- Any Bubbletea integration files — this package has no TUI dependencies

**Total files: 4** (all new)

**`renderer.go` content scope:**
```go
package render

import (
    "github.com/charmbracelet/glamour"
    "github.com/matheusmortatti/ink/internal/block"
)

// Renderer wraps Glamour for block-level markdown rendering.
type Renderer struct {
    glamour *glamour.TermRenderer
    width   int
}

// NewRenderer creates a renderer with the given terminal width.
func NewRenderer(width int) (*Renderer, error) { ... }

// Render converts a markdown block to ANSI-styled output via Glamour.
func (r *Renderer) Render(b block.Block) (string, error) { ... }

// SetWidth updates the renderer's word wrap width (requires recreating Glamour instance).
func (r *Renderer) SetWidth(width int) error { ... }

// PreRenderAll renders all blocks and populates the cache.
func PreRenderAll(blocks []block.Block, cache *RenderCache, width int) error { ... }
```

**`cache.go` content scope:**
```go
package render

import (
    "sync"
    "github.com/matheusmortatti/ink/internal/block"
)

// RenderCache stores rendered block output keyed by content hash and width.
type RenderCache struct {
    entries map[string]string
    mu      sync.RWMutex
}

// NewRenderCache creates an empty render cache.
func NewRenderCache() *RenderCache { ... }

// Get retrieves cached rendering if available.
func (c *RenderCache) Get(b block.Block, width int) (string, bool) { ... }

// Put stores a rendered block in the cache.
func (c *RenderCache) Put(b block.Block, width int, rendered string) { ... }

// InvalidateAll clears the entire cache (for terminal resize).
func (c *RenderCache) InvalidateAll() { ... }

// InvalidateBlock clears all cache entries for a specific block (for content changes).
func (c *RenderCache) InvalidateBlock(b block.Block) { ... }
```

### Testing Requirements

**Test location:** Co-located with source (Go convention)
- `internal/render/renderer_test.go`
- `internal/render/cache_test.go`

**Test naming:** `TestFunctionName_Scenario_ExpectedBehavior`
- Example: `TestRenderer_Heading_RendersStyled`
- Example: `TestRenderCache_SameContentDifferentWidth_CacheMiss`

**Test pattern:** Table-driven tests with `t.Run` subtests

**Test categories (all required):**

| Category | What to Test | Minimum Cases |
|---|---|---|
| Renderer initialization | `NewRenderer` with valid/invalid widths | 2+ |
| Block type rendering | H1-H6, paragraph, list, code fence, blockquote, table, horizontal rule | 10+ |
| Inline element rendering | Bold, italic, links, code spans within paragraphs | 4+ |
| Width handling | Same block rendered at different widths produces different output | 3+ |
| SetWidth behavior | Changing width recreates renderer correctly | 2+ |
| Pre-render | `PreRenderAll` populates cache for all blocks | 3+ |
| Error handling | Rendering malformed blocks returns error gracefully | 3+ |
| Cache Get/Put | Cache hit/miss behavior | 5+ |
| Cache invalidation | `InvalidateAll`, `InvalidateBlock` work correctly | 3+ |
| Content hashing | Same content produces same hash, different content produces different hash | 3+ |
| Concurrent access | Cache handles concurrent Get/Put without race conditions (use `go test -race`) | 2+ |

**CRITICAL TEST:** Block-level rendering correctness

Since Glamour is designed for full documents, we MUST verify that rendering isolated blocks produces correct output. Test EVERY block type listed in the architecture:

```go
func TestRenderer_AllBlockTypes_RenderCorrectly(t *testing.T) {
    tests := []struct {
        name     string
        block    block.Block
        contains string // substring that should appear in rendered output
    }{
        {"H1", block.Block{Type: block.Heading, Raw: "# Title", Level: 1}, "Title"},
        {"paragraph", block.Block{Type: block.Paragraph, Raw: "Hello world"}, "Hello world"},
        {"bold", block.Block{Type: block.Paragraph, Raw: "This is **bold** text"}, "bold"},
        // ... all block types
    }
    // ...
}
```

**Testing tools:** Go's built-in `testing` package ONLY. No external test framework.

**Race condition testing:** Run `go test -race ./internal/render/...` to verify concurrent cache access is safe.

### Previous Story Intelligence

**From Story 1.2 (Markdown Block Parser):**

**Learnings to apply:**
- The `Block` type has fields: `Type`, `Raw`, `Level`, `StartByte`, `EndByte`
- Use `Block.Raw` as the source markdown string for rendering
- Use `Block.Type` to determine if special handling is needed (e.g., headings have `Level`)
- Block parser guarantees `Block.Raw` is valid markdown for that block type
- Round-trip is guaranteed: blocks serialize back to original markdown
- All blocks pass through goldmark parsing — Glamour also uses goldmark, so compatibility is high

**Files created in Story 1.2:**
- `internal/block/block.go` — `Block` type, `BlockType` enum
- `internal/block/parser.go` — `Parse` function
- `internal/block/document.go` — `Document` type, `Serialize` method
- Tests: `parser_test.go`, `document_test.go` (55+ test cases)

**Dependency to import:**
```go
import "github.com/matheusmortatti/ink/internal/block"
```

**Block types to handle in rendering:**
```go
block.Paragraph
block.Heading   // Check Block.Level for H1-H6
block.List
block.CodeFence
block.CodeBlock // Indented code blocks
block.BlockQuote
block.Table
block.HorizontalRule
```

### Git Intelligence

**Recent commits:**
```
dc73dfb block parser (most recent)
78e9544 initial folder structure and example main.go file
```

**Patterns established:**
- Story files created in `_bmad-output/implementation-artifacts/`
- Implementation files under `internal/` packages
- Run `go mod tidy` after adding new imports
- All tests must pass: `go test ./internal/...`
- Code must vet cleanly: `go vet ./internal/...`

**File creation pattern from Story 1.2:**
- Created `internal/block/*.go` files
- Created corresponding `*_test.go` files
- Updated `go.mod` when new dependencies become direct (goldmark)
- Committed with descriptive message

### Latest Tech Information (Feb 2026)

**Glamour v0.10.0 (April 2025 - Latest Stable):**

**Key Features:**
- Automatic dark/light theme detection via `WithAutoStyle()`
- Word wrap control via `WithWordWrap(width)`
- Table rendering with link footers (v0.10.0 improvement)
- Adaptive color profiles (TrueColor, ANSI256, ANSI)
- Built on Goldmark (same parser as our block parser — compatibility!)

**Performance Notes:**
- No built-in caching — we implement our own (this story)
- Renderer reuse is critical — create once, use many times
- Main cost is Goldmark AST parsing (we minimize by caching rendered output)
- Used in production by GitHub CLI, GitLab CLI, Glow — proven performance

**Known Limitations:**
- NO public API for rendering isolated blocks (workaround required)
- Renderer width is immutable (need new instance for resize)
- TTY detection can strip ANSI codes in non-TTY contexts (not an issue for us — TUI always has TTY)

**Best Practices (from research):**
- Use `WithAutoStyle()` for automatic light/dark terminal detection
- Implement content-hash caching keyed by `(hash, width)`
- Create one `TermRenderer` per width, reuse across renders
- Handle rendering errors gracefully (malformed markdown can fail)
- For TUI: render in background goroutines to avoid freezing event loop (NOT needed for this story — Story 1.4 viewport handles async)

**API Example:**
```go
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(80),
)
if err != nil {
    return err
}

output, err := renderer.Render("# Heading\n\nParagraph text")
// output is ANSI-styled string ready for terminal display
```

### Project Structure Notes

- `internal/render/` is a NEW package created in this story
- Depends on `internal/block` (allowed — block is leaf package)
- Will be imported by `internal/ui` (Story 1.4) and `internal/editor` (Story 2.4)
- No conflicts with existing code — package doesn't exist yet

### Critical Validation Points

**Before marking this story done, verify:**

1. **All block types render correctly** — Create test document with every markdown element, render each block, visually inspect output looks like glow
2. **Cache hit rate** — Render same block twice, second render should hit cache (measure with logs or test assertion)
3. **Width-specific caching works** — Render same block at 80ch and 120ch, outputs should differ, cache should store both
4. **Pre-render populates cache** — After `PreRenderAll`, all blocks should be in cache (verify with cache size or Get calls)
5. **Invalidation works** — After `InvalidateAll`, cache should be empty
6. **No races** — `go test -race ./internal/render/...` should pass
7. **Performance** — Rendering 100 blocks should be fast (< 100ms total, ideally < 50ms)

**Acceptance criteria checklist:**
- [x] AC#1: All markdown elements render with glow quality
- [x] AC#2: Cache stores and retrieves by (hash, width) key
- [x] AC#3: Content change causes cache miss and re-render
- [x] AC#4: Width change invalidates cache globally
- [x] AC#5: Pre-render populates full cache on document load

## Change Log

- 2026-02-10: Implemented Glamour block rendering and cache — `internal/render` package with Renderer, RenderCache, PreRenderAll. All 5 ACs satisfied. 35+ tests pass including race detection. Glamour v0.10.0 added as direct dependency.
- 2026-02-10: Code review fixes — Fixed leading newline bug (Glamour output not fully trimmed), added width validation, switched SHA256→FNV-1a for cache performance, added RenderCached convenience method, added nil-safety to Render, added CodeBlock test coverage, rewrote broken error-continuation test, added error handling tests. 31 tests pass with race detection.

### References

- [Source: architecture.md#Rendering Pipeline & Caching] — Pre-render + cache strategy, cache key design
- [Source: architecture.md#Markdown Parsing & Block Model] — goldmark integration, Block type usage
- [Source: architecture.md#Package Boundary Rules] — `internal/render` dependencies
- [Source: architecture.md#Go Naming Patterns] — Naming conventions for this package
- [Source: architecture.md#Testing Patterns] — Co-located tests, table-driven
- [Source: epics.md#Story 1.3] — Acceptance criteria, user story
- [Source: prd.md#FR1, FR4] — Glamour rendering quality, all markdown elements
- [Source: prd.md#NFR2] — <50ms block transition target (depends on this cache!)
- [Source: ux-design-specification.md#Component Strategy#Rendered Block] — Rendering via Glamour
- [Glamour GitHub](https://github.com/charmbracelet/glamour) — Latest API, releases
- [Glamour Docs](https://pkg.go.dev/github.com/charmbracelet/glamour) — API reference

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Resolved dependency version conflict: `cellbuf@v0.0.13` incompatible with `x/ansi@v0.11.1` from bubbletea/v2. Upgraded to `cellbuf@v0.0.15`, `x/ansi@v0.11.5`, `colorprofile@v0.4.1`.
- Glamour handles isolated block rendering without needing document wrapper — all block types (H1-H6, paragraph, list, code fence, blockquote, table, horizontal rule) render correctly when passed individually.

### Completion Notes List

- Story depends on: Story 1.2 (Markdown Block Parser) — DONE
- Story blocks: Story 1.4 (Document Viewport), Story 2.4 (Block Transitions), Story 6.1 (Terminal Resize)
- **CRITICAL MILESTONE**: This is the make-or-break story for <50ms block transitions (NFR2)
- Glamour v0.10.0 confirmed and added as direct dependency
- **Key finding**: Glamour DOES handle isolated block rendering correctly — no wrapping workaround needed. Each block's Raw markdown renders as expected when passed directly to `renderer.Render()`.
- Cache strategy: `(FNV-1a(content), width)` tuple keys for render memoization (switched from SHA256 to FNV-1a for speed)
- Single `TermRenderer` instance reuse is implemented — created once in `NewRenderer`, reused across `Render` calls
- Width changes create new renderer instance via `SetWidth` (Glamour limitation — immutable width)
- `PreRenderAll` continues on error, collecting failures via `errors.Join`
- All 31+ tests pass including race detection (`go test -race`)
- Concurrent cache access is race-free via `sync.RWMutex`
- Leading and trailing newlines trimmed from Glamour output for clean block compositing in viewport
- `RenderCached` convenience method added for cache-aware rendering (check cache, render on miss, store)
- Width validation added to `NewRenderer` and `SetWidth` (rejects zero/negative)
- Nil-safety check added to `Render` method

### File List

**Files created:**
- `internal/render/renderer.go` — Renderer type, NewRenderer, Render, RenderCached, SetWidth, PreRenderAll
- `internal/render/renderer_test.go` — 31 tests: all block types (incl. CodeBlock), width handling, SetWidth, PreRenderAll, error handling, RenderCached
- `internal/render/cache.go` — RenderCache type, Get/Put/InvalidateAll/InvalidateBlock, hashContent (FNV-1a)
- `internal/render/cache_test.go` — 13 tests: hit/miss, invalidation, concurrency, hashing, width-specific

**Files modified:**
- `go.mod` — Added glamour v0.10.0 as direct dependency, upgraded cellbuf/ansi/colorprofile
- `go.sum` — Updated checksums
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — Story status set to review

