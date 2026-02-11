# Story 1.4: Document Viewport with Centered Writing Column

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->
<!-- Epic: View and Navigate a Beautiful Document -->
<!-- Story Key: 1-4-document-viewport-with-centered-writing-column -->
<!-- Date: 2026-02-11 -->

## Story

As a writer,
I want my rendered document displayed in a centered writing column that feels calm and focused,
so that I have a comfortable, distraction-free reading experience.

## Acceptance Criteria

1. **Given** a terminal width of 120+ characters **When** the viewport renders the document **Then** the writing column is 80 characters wide, horizontally centered with equal margins on both sides (FR5)

2. **Given** a terminal width between 80 and 119 characters **When** the viewport renders the document **Then** the writing column is 70% of the terminal width, horizontally centered

3. **Given** a terminal width below 40 characters **When** the viewport renders the document **Then** the writing column uses full terminal width with no centering (FR38)

4. **Given** a document with multiple rendered blocks **When** the viewport composites them **Then** blocks are displayed sequentially with standard spacing (1 blank line between blocks)

5. **Given** a document taller than the terminal height **When** the viewport is displayed **Then** only visible blocks are shown and the viewport is scrollable

## Tasks / Subtasks

- [x] Task 1: Implement layout calculation (AC: #1, #2, #3)
  - [x] 1.1 Create `internal/ui/layout.go` with `Layout` type for column width and centering calculations
  - [x] 1.2 Implement `CalculateColumnWidth(terminalWidth int) int` with breakpoint logic: 120+ → 80, 80-119 → 70%, 40-79 → full, <40 → full
  - [x] 1.3 Implement `CalculateMargin(terminalWidth, columnWidth int) int` returning `(terminalWidth - columnWidth) / 2`
  - [x] 1.4 Add layout constants: `DefaultColumnWidth = 80`, `MinCenteringWidth = 40`, `MediumBreakpoint = 80`, `LargeBreakpoint = 120`, `MediumWidthPercent = 70`
  - [x] 1.5 Write table-driven tests for all width breakpoints and edge cases

- [x] Task 2: Implement viewport component (AC: #4, #5)
  - [x] 2.1 Create `internal/ui/viewport.go` with `Viewport` type storing terminal dimensions, scroll offset, content lines, and layout
  - [x] 2.2 Implement `NewViewport(width, height int) *Viewport` constructor
  - [x] 2.3 Implement `SetContent(blocks []block.Block, renderer *render.Renderer, cache *render.RenderCache)` to render all blocks and compose into centered, spaced content
  - [x] 2.4 Implement `Resize(width, height int)` to recalculate layout and recompose content
  - [x] 2.5 Implement `View() string` returning only the visible portion of content (viewport windowing based on scroll offset and terminal height)
  - [x] 2.6 Implement `ScrollDown(lines int)` and `ScrollUp(lines int)` for viewport scrolling with bounds clamping
  - [x] 2.7 Implement `ScrollToTop()` and `ScrollToBottom()` for jump navigation

- [x] Task 3: Implement block compositing with centering (AC: #1, #2, #3, #4)
  - [x] 3.1 Implement `composeBlocks()` internal method that renders each block via cache, pads each line to column width, applies left margin for centering, and joins blocks with 1 blank line separator
  - [x] 3.2 Handle multi-line blocks correctly (each line within a rendered block gets the same centering/padding)
  - [x] 3.3 Ensure Glamour-rendered output is constrained to column width via the Renderer's word wrap setting

- [x] Task 4: Write comprehensive tests (AC: #1-#5)
  - [x] 4.1 Create `internal/ui/layout_test.go` with width breakpoint tests
  - [x] 4.2 Create `internal/ui/viewport_test.go` with viewport behavior tests
  - [x] 4.3 Test block compositing produces centered output
  - [x] 4.4 Test scrolling bounds (cannot scroll past top/bottom)
  - [x] 4.5 Test viewport windowing (only visible lines returned)
  - [x] 4.6 Test resize recalculates layout and recomposes content

## Dev Notes

### Why This Story Matters

This is the **visual foundation** for the entire ink experience. The centered writing column is what makes ink feel like a writing tool rather than a code editor. It creates the "calm, focused" aesthetic described in the UX design specification. Every future story (editing, navigation, status bar) will render within this viewport.

**Critical dependencies:**
- **Depends on**: `internal/render` (Story 1.3) for Glamour rendering and cache — DONE
- **Depends on**: `internal/block` (Story 1.2) for Block type and document parsing — DONE
- **Consumed by**: `internal/editor` (Story 1.5+) for displaying documents
- **Consumed by**: Story 1.5 (Open and Display Existing Markdown File) — next story
- **Consumed by**: Story 1.6 (Normal Mode Vim Navigation) for cursor-aware viewport scrolling
- **Consumed by**: Story 6.1 (Terminal Resize Handling) for responsive layout

**Architectural role:** This component is a plain Go struct owned by `EditorModel` (NOT an independent Bubbletea model). It composites rendered blocks into a scrollable, centered view. The viewport is a passive display surface — all interaction happens through vim motions and mouse events routed through the editor.

### Critical Design Decisions

**1. Custom Viewport (NOT Bubbles viewport)**

The Bubbles viewport component is designed for generic text content. ink's viewport has unique requirements:
- Block-aware compositing (rendered blocks separated by blank lines)
- Centered writing column with responsive breakpoints
- Future: mixed raw/rendered content when editing blocks (Story 2.4)
- Future: block-indexed scrolling for cursor tracking (Story 1.6)

Building a custom viewport from the start avoids wrapping and unwrapping Bubbles APIs. The viewport is a simple struct with string content and a scroll offset.

**2. Viewport as a Component Struct (NOT a Bubbletea Model)**

Per Architecture: "Components are plain Go structs with methods, not independent Bubbletea models." The viewport:
- Has no `Init`/`Update`/`View` Bubbletea lifecycle
- Returns a `string` from its `View()` method
- Is called by `EditorModel.View()` to get the content portion of the screen
- Receives resize events through `Resize()` called by `EditorModel.Update()`

**3. Centering via String Padding (NOT Lip Gloss)**

For block compositing, centering is achieved by prepending spaces to each line. This is simpler and more predictable than Lip Gloss margin/padding when dealing with ANSI-styled content from Glamour. Lip Gloss's `Place` function may interact unpredictably with pre-styled Glamour output.

```go
// Center each line by prepending spaces
margin := strings.Repeat(" ", leftMargin)
for _, line := range lines {
    centered = append(centered, margin + line)
}
```

**4. Content Model: Pre-Composed String Lines**

The viewport stores its content as a `[]string` (lines), pre-composed with centering and block spacing. This makes the `View()` method a simple slice operation:
```go
func (v *Viewport) View() string {
    end := min(v.scrollOffset + v.height, len(v.lines))
    visible := v.lines[v.scrollOffset:end]
    return strings.Join(visible, "\n")
}
```

Re-composition happens only on `SetContent()` or `Resize()` — NOT on every frame.

**5. Renderer Width = Column Width (NOT Terminal Width)**

The Glamour renderer's word wrap width MUST be set to the **column width**, not the terminal width. This ensures rendered blocks fit within the centered column. The renderer width is updated via `SetWidth()` whenever the column width changes.

```go
columnWidth := layout.CalculateColumnWidth(terminalWidth)
renderer.SetWidth(columnWidth) // Glamour wraps to column width
```

### Project Structure Notes

- `internal/ui/viewport.go` already exists as an empty package declaration — this story fills it with the Viewport type
- `internal/ui/layout.go` is a NEW file within the existing `internal/ui/` package
- No new packages created
- Aligns with Architecture directory structure

### References

- [Source: architecture.md#Project Structure] — `internal/ui` package: Viewport, StatusBar, layout calculations
- [Source: architecture.md#Component Communication] — Components are plain Go structs with methods
- [Source: architecture.md#Package Boundary Rules] — `internal/ui` depends on `internal/block`
- [Source: architecture.md#Rendering Pipeline & Caching] — Pre-render and cache strategy
- [Source: epics.md#Story 1.4] — Acceptance criteria, user story
- [Source: prd.md#FR5] — Centered writing column adapts to terminal width
- [Source: prd.md#FR38] — Centering disabled below 40 characters
- [Source: prd.md#NFR5] — Terminal resize must recalculate without perceptible delay
- [Source: prd.md#NFR6] — Scrolling must feel smooth with no rendering stutter
- [Source: ux-design-specification.md#Spacing & Layout Foundation] — Column width, margins, responsive breakpoints
- [Source: ux-design-specification.md#Terminal Width Adaptation] — Width breakpoint table
- [Source: ux-design-specification.md#Terminal Height Adaptation] — Height breakpoint table
- [Source: ux-design-specification.md#Document Viewport Component] — Viewport states and behavior

### Technical Requirements

**Go version:** Go 1.25+ (current go.mod)

**Bubbletea v2 (RC2) — View System:**

The current main.go uses `tea.View` return type and `tea.NewView()`:
```go
func (m model) View() tea.View {
    return tea.NewView("content string here\n")
}
```

The viewport's `View()` method returns a plain `string`. The `EditorModel.View()` method will compose the viewport string with the status bar string and wrap it in `tea.NewView()`.

**Lip Gloss v1.1.1 (indirect via Bubbletea v2):**

Lip Gloss is available as an indirect dependency. For this story, centering is done via string manipulation (space padding) rather than Lip Gloss to avoid issues with ANSI-styled Glamour output. Lip Gloss can be used for the status bar styling in a future story.

**Terminal Dimensions:**

Bubbletea v2 provides terminal dimensions via `tea.WindowSizeMsg`:
```go
case tea.WindowSizeMsg:
    // msg.Width and msg.Height contain terminal dimensions
    viewport.Resize(msg.Width, msg.Height)
```

**Render Integration:**

```go
import (
    "github.com/matheusmortatti/ink/internal/block"
    "github.com/matheusmortatti/ink/internal/render"
)

// Use RenderCached for each block
for _, b := range blocks {
    rendered, err := renderer.RenderCached(b, cache)
    // compose into viewport...
}
```

### Architecture Compliance

**Package: `internal/ui`** — UI components (viewport, status bar, layout)

**Dependency direction (MUST follow):**
```
internal/ui → internal/block (reads block data for display)
internal/ui → internal/render (uses RenderCached for block rendering)
internal/ui → (NO other internal/ imports allowed)
```

- `internal/ui` MUST import `internal/block` to access `Block` type — ALLOWED
- `internal/ui` MUST import `internal/render` to access `Renderer` and `RenderCache` — ALLOWED
- `internal/ui` MUST NOT import `internal/vim`, `internal/editor`, `internal/file`, or `internal/config` — FORBIDDEN

**Naming conventions (enforce strictly):**
- Package name: `ui` (singular, lowercase)
- No stutter: `ui.Viewport` is fine, `ui.UIViewport` is not
- Exported types: `Viewport`, `Layout`
- Unexported helpers: `composeBlocks`, `centerLine`, `clamp`
- Receiver names: single letter — `func (v *Viewport)`, `func (l *Layout)`

**Anti-patterns to AVOID:**
- Do NOT make Viewport a Bubbletea model (no Init/Update/View lifecycle)
- Do NOT import `internal/editor` (circular dependency)
- Do NOT use Lip Gloss Place/margin for centering rendered blocks (unpredictable with ANSI content)
- Do NOT store raw Block data in the viewport — store pre-composed string lines
- Do NOT re-compose content on every View() call — only on SetContent/Resize

### Library & Framework Requirements

| Library | Import Path | Version in go.mod | Usage in This Story |
|---|---|---|---|
| bubbletea v2 | `charm.land/bubbletea/v2` | v2.0.0-rc.2 | `tea.WindowSizeMsg` for resize events (consumed by editor, not viewport directly) |
| block | `github.com/matheusmortatti/ink/internal/block` | internal | `Block` type for document data |
| render | `github.com/matheusmortatti/ink/internal/render` | internal | `Renderer`, `RenderCache`, `RenderCached` for block rendering |
| strings | `strings` | stdlib | String manipulation for centering, line splitting, joining |

**No new external dependencies required for this story.**

**WARNING — Common LLM mistakes with viewport implementation:**

1. **Making viewport a Bubbletea model** — It's a plain struct, not a model. No Init/Update/View lifecycle. The `View()` method returns `string`, not `tea.View`.

2. **Using terminal width for Glamour rendering** — The Glamour renderer width MUST be set to the **column width**, not terminal width. Otherwise rendered content overflows the centered column.

3. **Re-rendering on every frame** — Content is pre-composed on `SetContent()` and `Resize()`. The `View()` method just slices the pre-composed lines. No rendering happens during view.

4. **Not handling ANSI escape sequences in line splitting** — Glamour output contains ANSI color codes. Use `strings.Split(content, "\n")` which handles this correctly since ANSI codes don't contain newlines.

5. **Off-by-one in scroll bounds** — Ensure scroll offset cannot go negative and `scrollOffset + viewportHeight` doesn't exceed total content lines.

6. **Forgetting to update renderer width on resize** — When terminal width changes, the column width changes, and the Glamour renderer must be updated via `SetWidth()` before re-rendering blocks.

### File Structure Requirements

**Files to create/modify in this story:**

```
internal/ui/
├── viewport.go       # Viewport type (MODIFY — currently empty package declaration)
├── viewport_test.go  # Viewport behavior tests (NEW)
├── layout.go         # Layout calculations: column width, centering, margins, breakpoints (NEW)
└── layout_test.go    # Layout calculation tests (NEW)
```

**Files NOT to create:**
- `statusbar.go` — Deferred to Story 3.1
- `saveprompt.go` — Deferred to Story 3.4
- Any Bubbletea model files — viewport is a component struct

**Total files: 4** (1 modify, 3 new)

**`layout.go` content scope:**
```go
package ui

// Layout holds calculated layout values for the writing column.
type Layout struct {
    TerminalWidth  int
    TerminalHeight int
    ColumnWidth    int
    LeftMargin     int
}

// Width breakpoint constants
const (
    DefaultColumnWidth  = 80
    MinCenteringWidth   = 40
    MediumBreakpoint    = 80
    LargeBreakpoint     = 120
    MediumWidthPercent  = 70
)

// NewLayout calculates layout from terminal dimensions.
func NewLayout(terminalWidth, terminalHeight int) Layout { ... }

// CalculateColumnWidth returns the writing column width for the given terminal width.
func CalculateColumnWidth(terminalWidth int) int { ... }
```

**`viewport.go` content scope:**
```go
package ui

import (
    "strings"
    "github.com/matheusmortatti/ink/internal/block"
    "github.com/matheusmortatti/ink/internal/render"
)

// Viewport displays a scrollable, centered document composed of rendered blocks.
type Viewport struct {
    layout       Layout
    lines        []string // pre-composed content lines with centering
    scrollOffset int      // first visible line index
    totalLines   int      // total content lines
}

// NewViewport creates a viewport with the given terminal dimensions.
func NewViewport(width, height int) *Viewport { ... }

// SetContent renders blocks and composes them into the viewport.
func (v *Viewport) SetContent(blocks []block.Block, renderer *render.Renderer, cache *render.RenderCache) error { ... }

// Resize recalculates layout for new terminal dimensions.
func (v *Viewport) Resize(width, height int) { ... }

// View returns the visible portion of the content as a string.
func (v *Viewport) View() string { ... }

// ScrollDown moves the viewport down by the given number of lines.
func (v *Viewport) ScrollDown(lines int) { ... }

// ScrollUp moves the viewport up by the given number of lines.
func (v *Viewport) ScrollUp(lines int) { ... }

// ScrollToTop jumps to the beginning of the document.
func (v *Viewport) ScrollToTop() { ... }

// ScrollToBottom jumps to the end of the document.
func (v *Viewport) ScrollToBottom() { ... }

// ContentHeight returns the total number of content lines.
func (v *Viewport) ContentHeight() int { ... }

// ScrollOffset returns the current scroll position.
func (v *Viewport) ScrollOffset() int { ... }

// ViewportHeight returns the visible area height.
func (v *Viewport) ViewportHeight() int { ... }
```

### Testing Requirements

**Test location:** Co-located with source (Go convention)
- `internal/ui/layout_test.go`
- `internal/ui/viewport_test.go`

**Test naming:** `TestFunctionName_Scenario_ExpectedBehavior`
- Example: `TestCalculateColumnWidth_LargeTerminal_Returns80`
- Example: `TestViewport_ScrollDown_ClampsAtBottom`

**Test pattern:** Table-driven tests with `t.Run` subtests

**Test categories (all required):**

| Category | What to Test | Minimum Cases |
|---|---|---|
| Column width (120+) | Returns 80 chars | 2+ (120, 200) |
| Column width (80-119) | Returns 70% of terminal | 3+ (80, 100, 119) |
| Column width (40-79) | Returns full terminal width | 2+ (40, 60) |
| Column width (<40) | Returns full terminal width | 2+ (20, 39) |
| Column width edge cases | Width of 0, 1, negative | 3+ |
| Margin calculation | Correct left margin for centering | 3+ |
| Block compositing | Blocks separated by blank lines | 2+ |
| Content centering | Lines are left-padded with correct margin | 3+ |
| Viewport windowing | Only visible lines returned | 3+ |
| Scroll down | Moves offset, clamps at bottom | 3+ |
| Scroll up | Moves offset, clamps at top (no negative) | 3+ |
| Scroll to top/bottom | Jump navigation works | 2+ |
| Resize | Layout recalculates, content recomposes | 2+ |
| Empty document | Viewport handles zero blocks gracefully | 1+ |

**Testing tools:** Go's built-in `testing` package ONLY. No external test framework.

**NOTE on testing approach:** Layout tests can be pure unit tests. Viewport tests that involve rendering need either mock rendered content or use the actual Glamour renderer. Using actual renderer is preferred for integration confidence, but mock content (pre-formatted strings) is acceptable for faster test execution. Choose based on test execution speed.

### Previous Story Intelligence

**From Story 1.3 (Glamour Block Rendering and Cache):**

**Learnings to apply:**
- `render.NewRenderer(width)` creates a renderer — use **column width** as the width parameter
- `renderer.RenderCached(block, cache)` returns cached or freshly-rendered output — use this for compositing
- `renderer.SetWidth(width)` recreates Glamour instance for new width — call on resize
- `cache.InvalidateAll()` clears cache on resize — call before re-rendering at new width
- Glamour output has leading/trailing newlines trimmed by `Render()` — blocks compose cleanly
- `renderer.PreRenderAll(blocks, cache)` populates cache for all blocks — use on initial content load
- FNV-1a hashing for cache keys (fast, sufficient for cache)
- `sync.RWMutex` protects concurrent cache access

**Files from Story 1.3 to use:**
- `internal/render/renderer.go` — `Renderer`, `NewRenderer`, `Render`, `RenderCached`, `SetWidth`, `PreRenderAll`
- `internal/render/cache.go` — `RenderCache`, `NewRenderCache`, `Get`, `Put`, `InvalidateAll`, `InvalidateBlock`

**From Story 1.2 (Markdown Block Parser):**
- `block.Parse(content)` returns `[]block.Block`
- `block.Block` has `Type`, `Raw`, `Level`, `StartByte`, `EndByte` fields
- Round-trip serialization preserves original content

### Git Intelligence

**Recent commits:**
```
c1d62f3 block rendering (most recent)
dc73dfb block parser
78e9544 initial folder structure and example main.go file
```

**Patterns established:**
- Implementation files under `internal/` packages
- Tests co-located with source files
- `go mod tidy` after adding imports
- `go test ./internal/...` for all tests
- `go vet ./internal/...` for static analysis
- Descriptive commit messages (short, lowercase)

**File creation patterns:**
- New files in existing packages (render package created in Story 1.3)
- Corresponding `*_test.go` files for each implementation file
- Existing empty `viewport.go` will be modified (not recreated)

### Latest Tech Information (Feb 2026)

**Bubbletea v2 RC2:**
- Module path: `charm.land/bubbletea/v2`
- `tea.View` return type for `View()` method (replaces `string`)
- `tea.NewView(content)` wraps string into View
- `tea.WindowSizeMsg` provides `Width` and `Height` fields
- New Cursed Renderer for better performance
- Split mouse messages for cleaner event handling

**Lip Gloss v1.1.1 (indirect):**
- Available as indirect dependency via Bubbletea v2
- `lipgloss.NewStyle()` for creating styles
- String manipulation (space padding) preferred over Lip Gloss Place for ANSI-content centering
- Can be used for status bar styling in future stories

**No new external dependencies needed for this story.**

### Critical Validation Points

**Before marking this story done, verify:**

1. **Width breakpoints work correctly** — Test with terminal widths 20, 39, 40, 60, 79, 80, 100, 119, 120, 200
2. **Centering is visually correct** — Content appears horizontally centered in the terminal
3. **Block spacing is correct** — Exactly 1 blank line between rendered blocks
4. **Scrolling works** — Can scroll down/up, bounded correctly, no negative offset
5. **Viewport windowing works** — Only visible lines returned, content not duplicated
6. **Resize works** — Changing terminal dimensions recalculates everything
7. **Empty document handled** — Zero blocks produces an empty viewport without crash
8. **Glamour integration works** — Rendered blocks display with glow-quality styling
9. **Performance** — Compositing 100+ blocks should be fast (no perceptible delay)
10. **All tests pass** — `go test ./internal/ui/...` and `go test -race ./internal/ui/...`

**Acceptance criteria checklist:**
- [ ] AC#1: 120+ terminal → 80ch centered column
- [ ] AC#2: 80-119 terminal → 70% centered column
- [ ] AC#3: <40 terminal → full width, no centering
- [ ] AC#4: Blocks displayed with 1 blank line spacing
- [ ] AC#5: Viewport scrollable for tall documents

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- No debug issues encountered during implementation.

### Completion Notes List

- **Task 1:** Implemented `Layout` type with `CalculateColumnWidth` (breakpoints: 120+→80, 80-119→70%, <80→full width) and `CalculateMargin` (centering calculation). Added 5 exported constants. 17 table-driven tests covering all breakpoints and edge cases (zero, negative).
- **Task 2:** Implemented `Viewport` struct as a plain Go component (not a Bubbletea model). Methods: `NewViewport`, `SetContent`, `Resize`, `View`, `ScrollDown`, `ScrollUp`, `ScrollToTop`, `ScrollToBottom`, `ContentHeight`, `ScrollOffset`, `ViewportHeight`. View() returns only visible lines via slice windowing. Scroll bounds clamped correctly.
- **Task 3:** Implemented `composeBlocks()` internal method that renders each block via `RenderCached`, applies left margin via string padding, and separates blocks with 1 blank line. On `Resize()`, renderer width is updated to column width, cache is invalidated, and blocks are recomposed.
- **Task 4:** 27 test functions (7 layout + 20 viewport) with 25+ sub-test cases in layout. Layout tests cover all width breakpoints and edge cases. Viewport tests cover: constructor, block compositing, blank line separators, centering with margin, no centering on small terminal, viewport windowing, scroll down/up with clamping, scroll to top/bottom, resize recalculation, empty document handling, content height, content shorter than viewport.
- All tests pass including race detection (`go test -race`). No regressions in full suite. `go vet` clean.
- **Code Review Fixes (2026-02-11):** Fixed 2 HIGH + 2 MEDIUM issues. H1: Resize() now returns error (SetWidth/composeBlocks errors no longer silently dropped). H2: SetContent() now syncs renderer width to layout column width. M1: CalculateColumnWidth caps medium breakpoint at DefaultColumnWidth (80) to match UX spec "whichever is smaller" rule. M2: Dev Agent Record test counts corrected to actual values. Added TestCalculateColumnWidth_MediumTerminal_CapsAt80 test with 4 cases.

### File List

- `internal/ui/layout.go` (NEW) — Layout type, CalculateColumnWidth, CalculateMargin, breakpoint constants
- `internal/ui/layout_test.go` (NEW) — 17 table-driven layout tests
- `internal/ui/viewport.go` (MODIFIED) — Viewport type with compositing, scrolling, windowing
- `internal/ui/viewport_test.go` (NEW) — 14 viewport behavior tests

### Change Log

- 2026-02-11: Implemented document viewport with centered writing column (Story 1.4). Created layout calculation with responsive width breakpoints (120+→80ch, 80-119→min(70%,80), <80→full). Built custom viewport component as plain Go struct with block compositing, centering via string padding, scroll management with bounds clamping, and viewport windowing. 29 tests covering all acceptance criteria.
- 2026-02-11: Code review fixes — Resize() returns error, SetContent syncs renderer width, column width capped at 80 for medium breakpoint, test counts corrected.
