# Story 1.6: Normal Mode Vim Navigation

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->
<!-- Epic: View and Navigate a Beautiful Document -->
<!-- Story Key: 1-6-normal-mode-vim-navigation -->
<!-- Date: 2026-02-11 -->

## Story

As a writer,
I want to navigate through my rendered document using vim motions,
so that I can quickly move to any part of my writing for review.

## Acceptance Criteria

1. **Given** the document is displayed in normal mode **When** the user presses `j` or `k` **Then** the cursor moves down or up by one line within the rendered content (FR15)

2. **Given** the document is displayed in normal mode **When** the user presses `h` or `l` **Then** the cursor moves left or right by one character within the rendered content (FR15)

3. **Given** the document is displayed in normal mode **When** the user presses `w` or `b` **Then** the cursor moves forward or backward by one word (FR15)

4. **Given** the document is displayed in normal mode **When** the user presses `G` **Then** the cursor jumps to the end of the document (FR15)

5. **Given** the document is displayed in normal mode **When** the user presses `gg` **Then** the cursor jumps to the beginning of the document (FR15)

6. **Given** the document is displayed in normal mode **When** the user presses `Ctrl+d` or `Ctrl+u` **Then** the viewport scrolls half a page down or up (FR15)

7. **Given** the cursor moves beyond the visible area **When** any navigation motion is used **Then** the viewport scrolls to keep the cursor visible (FR12) **And** scrolling feels smooth with no rendering stutter (NFR6)

## Tasks / Subtasks

- [x] Task 1: Implement vim mode infrastructure in `internal/vim` (AC: all)
  - [x] 1.1 Define `Mode` type (Normal, Insert, Visual, Command) and `ModeHandler` interface in `mode.go`
  - [x] 1.2 Define `Action` types (MoveCursor, ScrollViewport, ChangeMode, Quit, NoOp) in `action.go`
  - [x] 1.3 Implement `NormalHandler` in `normal.go` with operator-pending state for multi-key sequences (gg)
  - [x] 1.4 Implement shared motion functions (word, line, document motions) in `motion.go`
  - [x] 1.5 Write table-driven tests for NormalHandler key dispatch in `normal_test.go`

- [x] Task 2: Add document cursor state to `internal/editor` (AC: #1-#5, #7)
  - [x] 2.1 Add cursor state to `EditorModel`: `cursorLine int`, `cursorCol int` (document-level coordinates within rendered content lines)
  - [x] 2.2 Implement `ensureCursorVisible()` — adjusts viewport scroll offset to keep cursor on screen
  - [x] 2.3 Implement cursor rendering in `View()` — overlay cursor position on rendered content

- [x] Task 3: Refactor `EditorModel.Update()` to use vim mode delegation (AC: all)
  - [x] 3.1 Add `modeHandler` field (NormalHandler) to EditorModel
  - [x] 3.2 Replace hardcoded j/k handling with `modeHandler.HandleKey(msg) -> Action`
  - [x] 3.3 Implement `applyAction(action)` to translate vim Actions into state mutations (cursor moves, viewport scrolls)

- [x] Task 4: Implement all normal mode motions with viewport follow (AC: #1-#7)
  - [x] 4.1 Line motions: j/k move cursor line up/down with viewport follow
  - [x] 4.2 Character motions: h/l move cursor left/right within rendered line with wrapping at line boundaries
  - [x] 4.3 Word motions: w/b skip forward/backward by word boundaries
  - [x] 4.4 Document motions: G (end of document), gg (beginning of document)
  - [x] 4.5 Half-page motions: Ctrl+d/Ctrl+u scroll half-page and move cursor accordingly

- [x] Task 5: Write comprehensive tests (AC: all)
  - [x] 5.1 `internal/vim/normal_test.go` — key dispatch tests for all motions, multi-key gg, count prefixes if added
  - [x] 5.2 `internal/vim/motion_test.go` — unit tests for word boundary detection, line clamping
  - [x] 5.3 `internal/editor/editor_test.go` — update existing tests, add cursor movement tests, viewport follow tests
  - [x] 5.4 Manual verification: run `go run ./cmd/ink testfile.md` and test all navigation motions

## Dev Notes

### Technical Requirements

**Go version:** Go 1.25+ (current go.mod)

**This story introduces the vim mode system** — the architectural foundation for ALL future input handling. Every subsequent story (insert mode, visual mode, command mode) builds on what is created here. Getting the interfaces and patterns right is critical.

**Cursor model for normal mode:**

Normal mode navigation operates on the **rendered content lines** (the pre-composed `[]string` lines stored in the viewport). The cursor position is `(cursorLine, cursorCol)` where:
- `cursorLine` is an index into the total rendered content lines (0-indexed, document-level)
- `cursorCol` is a character offset within that line (0-indexed)

This is NOT the `(blockIndex, line, col)` representation from architecture — that triple is for mapping between rendered and raw when entering insert mode (Story 2.x). For normal mode navigation, the cursor moves over the flat rendered content. The block-level mapping will be added later.

**Cursor rendering approach:**

Bubbletea v2's `tea.View` has a `Cursor` field for declarative cursor placement:
```go
v := tea.NewView(content)
v.AltScreen = true
v.Cursor = tea.Cursor{
    Position: tea.CursorPosition{
        X: cursorScreenCol,
        Y: cursorScreenRow,
    },
}
```

Where `cursorScreenRow = cursorLine - scrollOffset` (the cursor's position relative to the visible viewport). `cursorScreenCol` must account for the left margin (centering padding).

**IMPORTANT: Cursor screen position calculation:**
- The viewport's `lines []string` contain centering padding (spaces prepended). The left margin is calculated by `ui.CalculateMargin(terminalWidth, columnWidth)`.
- `cursorScreenCol = leftMargin + cursorCol` — the cursor column in screen coordinates includes the centering offset.
- `cursorScreenRow = cursorLine - viewport.ScrollOffset()` — the cursor row relative to the visible area.

**Multi-key sequence handling (gg):**

The `NormalHandler` must support operator-pending state for multi-key sequences. `gg` is the only multi-key motion in this story. Implementation:
- `NormalHandler` has a `pending` field (string or rune) tracking the first key of a potential sequence
- When `g` is pressed: set `pending = 'g'`, return `NoOp` action
- When a second key arrives and `pending == 'g'`:
  - If second key is `g`: return `MoveCursor{Line: 0, Col: 0}` (go to top)
  - Otherwise: clear pending, process second key as standalone
- Any non-`g` key clears pending state

**Viewport follow ("cursor-follows-viewport"):**

After every cursor movement, `ensureCursorVisible()` adjusts the scroll offset:
```
if cursorLine < scrollOffset:
    scrollOffset = cursorLine  (cursor went above viewport)
if cursorLine >= scrollOffset + viewportHeight:
    scrollOffset = cursorLine - viewportHeight + 1  (cursor went below viewport)
```

For Ctrl+d/Ctrl+u (half-page), BOTH the cursor and scroll offset move by `viewportHeight / 2`. The cursor follows the scroll, not the other way around.

**Word motion definition (w/b):**

For `w` (word forward): advance cursor to the start of the next word. A "word" boundary is defined by transitions between character classes: alphanumeric+underscore vs. punctuation vs. whitespace. This matches vim's default `w` behavior.

For `b` (word backward): move cursor to the start of the previous word.

Word motions operate on the visible rendered content lines. They may cross line boundaries (moving from end of one line to start of the next).

**ANSI-aware character movement:**

The rendered content lines contain ANSI escape sequences from Glamour rendering. When calculating cursor positions for h/l/w/b motions, the cursor must move over **visible characters**, not raw bytes. Use a function that strips ANSI sequences to calculate visible character positions, but keep the cursor column in terms of visible character positions.

The viewport's `lines []string` contain ANSI-styled text with centering spaces prepended. When processing motions:
1. Strip ANSI codes to get the visible text for calculating word boundaries and character positions
2. Store `cursorCol` as a visible-character offset within the content area (excluding left margin)
3. Clamp `cursorCol` to `[0, visibleLineLength-1]` — vim-style, cursor sits ON a character, not after it

### Architecture Compliance

**Package: `internal/vim`** — Mode handler interface, NormalMode handler, action types, shared motions. Currently has only an empty `mode.go` (package declaration). This story creates the full vim infrastructure.

**Package: `internal/editor`** — EditorModel (the single Bubbletea model). Currently has working `editor.go` with hardcoded j/k scrolling. This story refactors it to use the vim mode delegation pattern.

**Dependency direction (MUST follow):**
```
cmd/ink → internal/editor (unchanged)
internal/editor → internal/vim (NEW — editor creates NormalHandler, calls HandleKey)
internal/editor → internal/block, internal/render, internal/ui (unchanged)
internal/vim → (no internal dependencies for this story — leaf-like for now)
```

**CRITICAL: `internal/vim` MUST NOT import `internal/editor`** — the dependency flows from editor to vim, never the reverse. The vim package defines actions as return values; the editor interprets and applies them. This is the action delegation pattern from the architecture.

**CRITICAL: `internal/vim` MUST NOT import `internal/ui`** — vim handlers don't know about the viewport. They return abstract actions (move cursor to line X, col Y). The editor translates those into viewport operations.

**FORBIDDEN imports for `internal/vim`:**
- `internal/editor` — never (would create a cycle)
- `internal/ui` — never (vim doesn't know about display)
- `internal/render` — never (vim doesn't know about rendering)
- `internal/file` — never (vim doesn't know about I/O)

**Allowed imports for `internal/vim`:**
- `internal/block` — in future stories for gap buffer operations in insert mode. NOT needed for this story.
- Standard library only for this story (`unicode`, `strings`, etc.)

**Naming conventions (enforce strictly):**
- Package name: `vim` (singular, lowercase)
- No stutter: `vim.NormalHandler` is fine, `vim.VimNormalHandler` is NOT
- Mode type: `vim.Mode` (exported int type with constants: `Normal`, `Insert`, `Visual`, `Command`)
- Handler interface: `vim.ModeHandler` — the one architecture-defined exception where provider defines interface
- Action types: `vim.Action` (interface or struct), specific actions as types: `vim.MoveCursorAction`, `vim.ScrollAction`, `vim.QuitAction`, `vim.NoOpAction`
- Receiver names: single letter — `func (n *NormalHandler)`, `func (a MoveCursorAction)`
- Error messages: lowercase, no trailing punctuation

**EditorModel delegation pattern (from architecture):**
```go
func (e *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        action := e.modeHandler.HandleKey(msg)
        return e.applyAction(action)
    case tea.WindowSizeMsg:
        // ... existing resize handling
    }
    return e, nil
}
```

Components return actions, EditorModel applies them. The `applyAction` method is the single coordination point for all state mutations (cursor position, viewport scroll, mode changes).

**Anti-patterns to AVOID:**
- Do NOT have NormalHandler directly call viewport.ScrollDown/Up — it returns scroll actions, editor applies them
- Do NOT import `tea` (bubbletea) in the vim package if avoidable — pass key info as simple types (string, rune) to keep vim package framework-independent. However, passing `tea.KeyPressMsg` directly is acceptable if it simplifies the interface.
- Do NOT add insert/visual/command mode handlers yet — define the interface and Mode constants, implement only NormalHandler
- Do NOT add count prefix support (e.g., `5j`) unless trivial — defer to a future enhancement
- Do NOT add operator support (d, y, c) — deferred to Epic 5
- Do NOT create `internal/vim/handler.go` or `internal/vim/manager.go` — the mode.go file holds the interface, each mode gets its own file

### Library & Framework Requirements

| Library | Import Path | Version in go.mod | Usage in This Story |
|---|---|---|---|
| bubbletea v2 | `charm.land/bubbletea/v2` | v2.0.0-rc.2 | `tea.KeyPressMsg`, `tea.View` with `Cursor` field, `tea.CursorPosition` for declarative cursor placement |
| vim (internal) | `github.com/matheusmortatti/ink/internal/vim` | internal | NEW — `vim.NormalHandler`, `vim.ModeHandler` interface, `vim.Action` types, `vim.Mode` constants |
| editor (internal) | `github.com/matheusmortatti/ink/internal/editor` | internal | MODIFY — Add `modeHandler`, cursor state, `applyAction()`, refactor `Update()` |
| ui (internal) | `github.com/matheusmortatti/ink/internal/ui` | internal | READ ONLY — `viewport.ScrollOffset()`, `viewport.ViewportHeight()`, `viewport.ContentHeight()`, `viewport.ScrollDown/Up()`, `viewport.ScrollToTop/Bottom()`, `ui.CalculateMargin()` |
| unicode | `unicode` | stdlib | Character classification for word motions (IsLetter, IsDigit, IsSpace, IsPunct) |
| strings | `strings` | stdlib | ANSI stripping, string manipulation for motion calculations |
| regexp | `regexp` | stdlib | ANSI escape sequence stripping pattern: `\x1b\[[0-9;]*[a-zA-Z]` |

**No new external dependencies required for this story.**

**Bubbletea v2 cursor API — critical for this story:**

In Bubbletea v2, cursor position is set declaratively in the `tea.View` struct. The exact API:
```go
import tea "charm.land/bubbletea/v2"

func (e *EditorModel) View() tea.View {
    v := tea.NewView(e.viewport.View())
    v.AltScreen = true

    if e.ready {
        screenRow := e.cursorLine - e.viewport.ScrollOffset()
        screenCol := e.leftMargin() + e.cursorCol
        v.Cursor = tea.Cursor{
            Position: tea.CursorPosition{
                X: screenCol,
                Y: screenRow,
            },
        }
    }

    return v
}
```

**WARNING — Common LLM mistakes with vim navigation stories:**

1. **Moving cursor in raw byte positions** — Rendered content contains ANSI escape sequences. Moving by bytes instead of visible characters will place the cursor inside escape sequences, causing visual corruption. ALWAYS strip ANSI before calculating positions.

2. **Forgetting centering margin in cursor X position** — The viewport lines have centering spaces prepended. If you set `Cursor.X = cursorCol` without adding the left margin, the cursor will appear in the margin area instead of on the content.

3. **Off-by-one on viewport boundary scroll** — When checking if cursor is below viewport, use `cursorLine >= scrollOffset + viewportHeight` (not `>`). The viewport shows lines `[scrollOffset, scrollOffset + viewportHeight)`.

4. **Word motions not crossing line boundaries** — vim's `w` and `b` cross line boundaries. If the cursor is at the end of a line and `w` is pressed, it should move to the first word of the next line. Don't stop at line ends.

5. **Not clamping cursor column on j/k** — When moving vertically, the target line may be shorter. Vim clamps the cursor column to `min(cursorCol, lineLength-1)`. Track a "desired column" (sticky column) that is restored when moving to a longer line.

6. **Mutating state in the vim handler** — The NormalHandler should NOT hold references to editor state. It receives key input and returns an action describing what should happen. The editor applies the action. This keeps the vim package testable in isolation.

7. **Using `tea.KeyMsg` instead of `tea.KeyPressMsg`** — In Bubbletea v2, the concrete key event type is `tea.KeyPressMsg` (based on the `Key` struct). `tea.KeyMsg` is an interface. The `Update()` switch should use `tea.KeyPressMsg` as established in Story 1.5.

### File Structure Requirements

**Files to create in this story:**

```
internal/vim/
├── mode.go           # MODIFY — Replace empty package declaration with Mode type, ModeHandler interface
├── action.go         # NEW — Action interface and concrete action types
├── normal.go         # NEW — NormalHandler implementing ModeHandler for normal mode
├── normal_test.go    # NEW — Table-driven tests for NormalHandler key dispatch
├── motion.go         # NEW — Shared motion helper functions (word boundaries, line content extraction)
└── motion_test.go    # NEW — Unit tests for motion helpers (word boundary detection, ANSI stripping)
```

**Files to modify in this story:**

```
internal/editor/
├── editor.go         # MODIFY — Add modeHandler field, cursor state, refactor Update(), add applyAction(), update View() with cursor
└── editor_test.go    # MODIFY — Update existing tests for new Update() behavior, add cursor movement tests
```

**Files NOT to create:**
- `internal/vim/insert.go` — Deferred to Story 2.2
- `internal/vim/visual.go` — Deferred to Story 5.4
- `internal/vim/command.go` — Deferred to Story 3.3
- `internal/editor/actions.go` — Per architecture this file exists for `applyAction`, but for this story the logic is simple enough to live in `editor.go`. Create `actions.go` when action handling grows complex (likely Story 2.x with block transitions).
- `internal/editor/startup.go` — Deferred per Story 1.5 notes
- `internal/ui/cursor.go` — No separate cursor component. Cursor position is a field on EditorModel, rendered via `tea.View.Cursor`.

**Total files: 8** (2 modify, 5 new, 1 modify from empty placeholder)

**`internal/vim/mode.go` content scope:**
```go
package vim

// Mode represents the current vim editing mode.
type Mode int

const (
    Normal Mode = iota
    Insert
    Visual
    Command
)

// String returns the display name of the mode.
func (m Mode) String() string {
    switch m {
    case Normal:
        return "NORMAL"
    case Insert:
        return "INSERT"
    case Visual:
        return "VISUAL"
    case Command:
        return "COMMAND"
    default:
        return "UNKNOWN"
    }
}

// ModeHandler processes key input and returns an Action.
// Each mode (Normal, Insert, Visual, Command) implements this interface.
type ModeHandler interface {
    // HandleKey processes a key string and returns the resulting action.
    HandleKey(key string) Action

    // Mode returns the current mode this handler represents.
    Mode() Mode
}
```

**`internal/vim/action.go` content scope:**
```go
package vim

// Action represents a command returned by a ModeHandler.
// The editor interprets and applies actions — handlers never mutate state directly.
type Action interface {
    actionTag() // unexported method prevents external implementations
}

// NoOpAction indicates no state change needed.
type NoOpAction struct{}

// MoveCursorAction moves the cursor to an absolute or relative position.
type MoveCursorAction struct {
    Line     int  // Target line (absolute if Relative=false, delta if Relative=true)
    Col      int  // Target column (absolute if Relative=false, delta if Relative=true)
    Relative bool // If true, Line/Col are deltas from current position
}

// ScrollAction scrolls the viewport by a number of lines.
type ScrollAction struct {
    Lines    int  // Positive = down, negative = up
    MoveCursor bool // If true, also move cursor by same amount
}

// DocumentPositionAction jumps to a document-level position.
type DocumentPositionAction struct {
    Position string // "top" or "bottom"
}

// QuitAction signals the editor to exit.
type QuitAction struct{}

func (NoOpAction) actionTag()             {}
func (MoveCursorAction) actionTag()       {}
func (ScrollAction) actionTag()           {}
func (DocumentPositionAction) actionTag() {}
func (QuitAction) actionTag()             {}
```

**`internal/vim/normal.go` content scope:**
```go
package vim

// NormalHandler processes key input in normal mode.
type NormalHandler struct {
    pending rune // For multi-key sequences (e.g., 'g' waiting for second key)
}

func NewNormalHandler() *NormalHandler { ... }
func (n *NormalHandler) HandleKey(key string) Action { ... }
func (n *NormalHandler) Mode() Mode { return Normal }
```

**`internal/vim/motion.go` content scope:**
```go
package vim

// StripANSI removes ANSI escape sequences from a string.
func StripANSI(s string) string { ... }

// VisibleLength returns the visible character count (excluding ANSI codes).
func VisibleLength(s string) int { ... }

// NextWordStart finds the next word boundary position from a given position in text.
func NextWordStart(text string, pos int) int { ... }

// PrevWordStart finds the previous word boundary position from a given position in text.
func PrevWordStart(text string, pos int) int { ... }
```

**`internal/editor/editor.go` modifications:**

Add to `EditorModel` struct:
```go
type EditorModel struct {
    // ... existing fields ...
    modeHandler vim.ModeHandler  // NEW — current mode's key handler
    mode        vim.Mode         // NEW — current mode (Normal for this story)
    cursorLine  int              // NEW — cursor line in document content (0-indexed)
    cursorCol   int              // NEW — cursor column in visible chars (0-indexed)
    desiredCol  int              // NEW — sticky column for j/k vertical movement
}
```

Add new methods:
```go
func (e *EditorModel) applyAction(action vim.Action) (tea.Model, tea.Cmd) { ... }
func (e *EditorModel) ensureCursorVisible() { ... }
func (e *EditorModel) leftMargin() int { ... }
func (e *EditorModel) visibleLineLength(line int) int { ... }  // ANSI-stripped length of content at line
func (e *EditorModel) clampCursorCol() { ... }  // Clamp col to [0, visibleLineLength-1]
```

### Testing Requirements

**Test location:** Co-located with source (Go convention)
- `internal/vim/normal_test.go`
- `internal/vim/motion_test.go`
- `internal/editor/editor_test.go` (modify existing)

**Test naming:** `TestFunctionName_Scenario_ExpectedBehavior`
- Example: `TestNormalHandler_JKey_ReturnsMoveDown`
- Example: `TestNormalHandler_GG_ReturnsDocumentTop`
- Example: `TestStripANSI_WithEscapeCodes_ReturnsCleanText`

**Test pattern:** Table-driven tests with `t.Run` subtests

**Test categories (all required):**

| Category | What to Test | Minimum Cases |
|---|---|---|
| NormalHandler — single key motions | j, k, h, l, w, b, G each return correct action type | 7 |
| NormalHandler — multi-key (gg) | g then g → DocumentPositionAction{top}, g then other → clears pending + processes key | 3 |
| NormalHandler — scroll motions | ctrl+d, ctrl+u return ScrollAction with correct direction | 2 |
| NormalHandler — quit | ctrl+c returns QuitAction | 1 |
| NormalHandler — unknown key | unrecognized key returns NoOpAction | 2 |
| NormalHandler — pending state cleanup | g followed by non-g key clears pending and processes key normally | 2 |
| StripANSI — various | plain text unchanged, single escape removed, multiple escapes removed, empty string | 4 |
| VisibleLength — various | plain text, ANSI-styled text, empty string | 3 |
| NextWordStart — word boundaries | middle of word, between words, punctuation boundary, end of text, cross-line | 5 |
| PrevWordStart — word boundaries | middle of word, start of word, beginning of text, cross-line | 4 |
| Editor — cursor movement j/k | j moves cursorLine +1, k moves cursorLine -1, clamped at bounds | 4 |
| Editor — cursor movement h/l | l moves cursorCol +1, h moves cursorCol -1, clamped at line bounds | 4 |
| Editor — cursor movement w/b | w advances to next word, b retreats to previous word | 2 |
| Editor — G and gg | G sets cursorLine to last line, gg sets cursorLine to 0 | 2 |
| Editor — Ctrl+d/u | Ctrl+d scrolls half-page down + moves cursor, Ctrl+u scrolls half-page up + moves cursor | 2 |
| Editor — ensureCursorVisible | cursor above viewport scrolls up, cursor below viewport scrolls down, cursor in view no change | 3 |
| Editor — sticky column (desiredCol) | j/k preserve desiredCol on shorter lines, h/l reset desiredCol | 3 |
| Editor — View cursor position | View() returns tea.View with Cursor.Position reflecting cursorLine/cursorCol + margin offset | 2 |
| Editor — existing tests updated | Previous j/k scroll tests updated to reflect new cursor-based behavior, ctrl+c still works | 3 |

**Testing tools:** Go's built-in `testing` package ONLY. No external test framework.

**Testing approach for NormalHandler (isolated):**

NormalHandler tests should be purely unit tests — create a handler, call HandleKey, assert the returned Action type and values:
```go
func TestNormalHandler_JKey_ReturnsMoveDown(t *testing.T) {
    h := vim.NewNormalHandler()
    action := h.HandleKey("j")

    move, ok := action.(vim.MoveCursorAction)
    if !ok {
        t.Fatalf("expected MoveCursorAction, got %T", action)
    }
    if move.Line != 1 || !move.Relative {
        t.Errorf("expected relative line +1, got line=%d relative=%v", move.Line, move.Relative)
    }
}
```

**Testing approach for EditorModel (integration):**

EditorModel tests create an editor with known blocks, simulate WindowSizeMsg to initialize, then send KeyPressMsg and verify cursor state:
```go
func TestEditorModel_JKey_MovesCursorDown(t *testing.T) {
    blocks := block.Parse([]byte("# Title\n\nParagraph one\n\nParagraph two"))
    e := editor.NewEditor("test.md", blocks)

    // Initialize viewport
    e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

    // Press j
    e.Update(tea.KeyPressMsg{Key: tea.Key{Code: 'j'}})

    if e.CursorLine() != 1 {
        t.Errorf("expected cursorLine=1, got %d", e.CursorLine())
    }
}
```

**Note on EditorModel field access for tests:** Story 1.5 established that tests access unexported fields directly (same package) or use exported getters. Add minimal exported getters for test-relevant cursor state: `CursorLine() int`, `CursorCol() int`, `CurrentMode() vim.Mode`.

**Regression: existing editor tests must still pass.** The refactored Update() changes key handling from direct scroll to vim delegation. Existing tests for j/k behavior and ctrl+c must be updated to match the new cursor-based behavior (j/k now move cursor, not just scroll). Ensure `go test ./internal/...` passes all packages.

### Previous Story Intelligence

**From Story 1.5 (Open and Display Existing Markdown File) — DONE:**

**Current `EditorModel` state that this story modifies:**
```go
type EditorModel struct {
    blocks   []block.Block
    viewport *ui.Viewport
    renderer *render.Renderer
    cache    *render.RenderCache
    filePath string
    ready    bool
    width    int
    height   int
}
```

The `Update()` method currently has hardcoded key handling:
```go
case tea.KeyPressMsg:
    switch msg.String() {
    case "ctrl+c":
        return e, tea.Quit
    case "j":
        if e.ready && e.viewport != nil {
            e.viewport.ScrollDown(1)
        }
    case "k":
        if e.ready && e.viewport != nil {
            e.viewport.ScrollUp(1)
        }
    }
```

This must be replaced with the vim mode delegation pattern. The j/k handling currently only scrolls the viewport — it has no concept of a cursor position. This story replaces this with cursor-based navigation where the viewport follows the cursor.

**Viewport API established in Stories 1.3-1.4 (used, not modified):**
- `viewport.ScrollDown(lines)` / `viewport.ScrollUp(lines)` — with bounds clamping
- `viewport.ScrollToTop()` / `viewport.ScrollToBottom()` — jump navigation
- `viewport.ScrollOffset()` — current scroll position (int)
- `viewport.ViewportHeight()` — visible rows (int)
- `viewport.ContentHeight()` — total content lines (int)
- `viewport.View()` — returns visible content as string
- `viewport.Resize(width, height)` — returns error

**Key insight: viewport stores `lines []string`** — the pre-composed content lines with centering padding baked in. The cursor navigates these lines. Access to the line content is needed for:
1. Calculating visible character length (ANSI stripping) for column clamping
2. Word boundary detection for w/b motions
3. Determining if a line is empty/blank for navigation behavior

**The viewport currently does NOT expose its `lines` field.** You will need to either:
1. Add a `Lines() []string` or `Line(n int) string` getter to `ui.Viewport` (preferred — minimal surface)
2. Or maintain a parallel line reference in EditorModel

Option 1 is recommended — add `Line(n int) string` to viewport that returns a specific content line by index (bounds-checked, returns empty string for out-of-range).

**Bubbletea v2 API learnings from Story 1.5:**
- `tea.KeyPressMsg` is the concrete type (NOT `tea.KeyMsg` which is an interface)
- `msg.String()` returns key as string: `"j"`, `"k"`, `"ctrl+c"`, `"ctrl+d"`, `"ctrl+u"`, `"G"`, `"g"`
- `tea.View` struct with `AltScreen`, `Cursor` fields — set declaratively in `View()`
- `Init()` returns only `tea.Cmd` (not `(tea.Model, tea.Cmd)`)
- `WindowSizeMsg` auto-delivered on startup

**Code review fixes from Story 1.5 to be aware of:**
- `Resize()` returns error — must be handled (currently `_ =` assignment, acceptable)
- `PreRenderAll()` and `SetContent()` errors acknowledged with `_ =`
- Editor tests access unexported fields directly (same package)

**Established development patterns (carry forward):**
- Tests co-located with source (`*_test.go`)
- Table-driven tests with descriptive `TestFunctionName_Scenario_Expected` naming
- `go test ./internal/...` for all tests
- `go vet ./internal/...` for static analysis
- `go mod tidy` after adding new imports
- Single-letter receiver names (e, v, r, c, b, n)
- Helper functions unexported with descriptive names
- Imports grouped: stdlib, then external, then internal

### Git Intelligence

**Recent commit history (newest first):**
```
bbd9172 open and display existing markdown file
dbeea60 document viewport with centered writing column
c1d62f3 block rendering
dc73dfb block parser
78e9544 initial folder structure and example main.go file
```

**Commit pattern:** Short, lowercase, descriptive — no prefixes, no ticket numbers. This story's commit should follow the same pattern, e.g., `normal mode vim navigation`.

**Last commit (bbd9172) — Story 1.5 files:**
```
.gitignore                                          |   1 +
1-5-open-and-display-existing-markdown-file.md      | 611 +++
sprint-status.yaml                                  |   2 +-
cmd/ink/main.go                                     |  46 +-
internal/editor/editor.go                           |  98 +
internal/editor/editor_test.go                      | 218 +
internal/file/file.go                               |  36 +
internal/file/file_test.go                          |  89 +
```

**File creation pattern across previous commits:**

| Commit | Pattern |
|---|---|
| `bbd9172` (1.5) | Fill empty placeholders (editor.go, file.go) + add test files + update sprint-status + add story.md |
| `dbeea60` (1.4) | Fill empty placeholder (viewport.go) + add layout.go, test files + update sprint-status + add story.md |
| `c1d62f3` (1.3) | Fill empty placeholder (renderer.go) + add cache.go, test files + update sprint-status + add story.md |
| `dc73dfb` (1.2) | Fill empty placeholder (block.go) + add parser.go, document.go, test files + update sprint-status + add story.md |

**Expected pattern for this story:**
- `internal/vim/mode.go` — MODIFY (fill empty placeholder with Mode type, ModeHandler interface)
- `internal/vim/action.go` — NEW
- `internal/vim/normal.go` — NEW
- `internal/vim/normal_test.go` — NEW
- `internal/vim/motion.go` — NEW
- `internal/vim/motion_test.go` — NEW
- `internal/editor/editor.go` — MODIFY (add vim delegation, cursor state)
- `internal/editor/editor_test.go` — MODIFY (update tests for new behavior)
- `internal/ui/viewport.go` — MODIFY (add Line() getter if needed)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — MODIFY (update story status)
- `_bmad-output/implementation-artifacts/1-6-normal-mode-vim-navigation.md` — ADD (this story file)

**Code conventions observed in recent commits:**
- Imports grouped: stdlib, then external (`charm.land/bubbletea/v2`), then internal (`github.com/matheusmortatti/ink/internal/...`)
- Exported functions have doc comments
- Receiver names: single letter (`e` for EditorModel, `v` for Viewport, `r` for Renderer)
- Helper functions unexported with descriptive camelCase names
- Error variables at package level with `Err` prefix
- No log statements in production code

### Latest Tech Information (Feb 2026)

**Bubbletea v2 RC2 — Cursor API (verified from source):**

Module path: `charm.land/bubbletea/v2` (v2.0.0-rc.2 in go.mod)

The cursor is set declaratively in the `View` struct. Here are the exact types from source:

```go
// cursor.go
type Position struct{ X, Y int }

type CursorShape int
const (
    CursorBlock CursorShape = iota
    CursorUnderline
    CursorBar
)

// tea.go
type Cursor struct {
    Position              // Embedded — has X, Y int fields
    Color    color.Color  // Optional cursor color
    Shape    CursorShape  // Block, Underline, or Bar
    Blink    bool         // Whether cursor blinks
}

func NewCursor(x, y int) *Cursor  // Helper constructor

type View struct {
    Content    Layer      // Main content (string or lipgloss canvas)
    Cursor     *Cursor    // nil = cursor hidden, non-nil = cursor shown at position
    AltScreen  bool       // Full-screen mode
    MouseMode  MouseMode  // Mouse event configuration
    // ... other fields
}
```

**Correct cursor usage pattern for this story:**
```go
func (e *EditorModel) View() tea.View {
    v := tea.NewView(e.viewport.View())
    v.AltScreen = true

    if e.ready {
        screenRow := e.cursorLine - e.viewport.ScrollOffset()
        screenCol := ui.CalculateMargin(e.width, ui.CalculateColumnWidth(e.width)) + e.cursorCol

        // Use NewCursor helper — sets Position with X, Y
        v.Cursor = tea.NewCursor(screenCol, screenRow)
        v.Cursor.Shape = tea.CursorBlock  // Block cursor for normal mode (vim convention)
    }

    return v
}
```

**Key API facts verified from source:**
1. `Cursor` is a **pointer** on View (`*Cursor`) — `nil` means cursor hidden, non-nil means shown
2. `Position` is **embedded** in Cursor (not a named field) — access via `cursor.X`, `cursor.Y` directly
3. `NewCursor(x, y)` returns `*Cursor` with default settings (no color, block shape, no blink specified)
4. Position is relative to **top-left of frame** (screen coordinates, not document coordinates)
5. `CursorBlock` is the default shape (iota = 0) — correct for normal mode

**KeyPressMsg API (verified, unchanged from Story 1.5):**
```go
// msg.String() returns key as string:
// "j", "k", "h", "l", "w", "b", "G", "g"
// "ctrl+c", "ctrl+d", "ctrl+u"
// Key struct has Code (rune), Mod (modifier flags)
```

**ANSI escape stripping — recommended approach:**

Use a compiled regexp for performance (called frequently during motions):
```go
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func StripANSI(s string) string {
    return ansiRegexp.ReplaceAllString(s, "")
}
```

Alternative: use `github.com/muesli/reflow/ansi` which is already an indirect dependency via Glamour/Lip Gloss. However, using a simple regexp avoids adding an explicit dependency and is sufficient for this use case.

**No new external dependencies required.** All needed functionality uses stdlib (`unicode`, `regexp`, `strings`) or existing Bubbletea v2 APIs.

### Project Structure Notes

- `internal/vim/mode.go` exists as an empty package declaration — this story fills it with Mode type, ModeHandler interface
- `internal/vim/` is the first package to receive real implementation in this story — all other vim files are new
- `internal/editor/editor.go` has working code from Story 1.5 — this story refactors the Update() method (breaking change to key handling) and extends the struct
- `internal/ui/viewport.go` may need a minor addition (Line() getter) — this is a non-breaking additive change
- No new packages created — `internal/vim/` already exists from Story 1.1
- Aligns with Architecture directory structure and dependency diagram

### References

- [Source: architecture.md#Vim Mode Architecture] — In-house implementation with per-mode handler pattern, multi-key sequences via operator-pending state
- [Source: architecture.md#Component Communication] — Single Bubbletea model, components return actions, editor applies them
- [Source: architecture.md#Package Boundary Rules] — `internal/vim` depends on `internal/block` (future), no other internal imports
- [Source: architecture.md#Implementation Patterns] — ModeHandler interface defined by provider (exception to consumer-defines rule)
- [Source: architecture.md#Cursor Position Representation] — (blockIndex, line, col) document-level, (line, col) within block — normal mode uses flat content lines for now
- [Source: epics.md#Story 1.6] — Acceptance criteria, user story, vim motions h/j/k/l/w/b/G/gg/Ctrl+d/Ctrl+u
- [Source: prd.md#FR12] — Normal mode navigation through fully rendered document
- [Source: prd.md#FR15] — Vim motions in normal mode (h, j, k, l, w, b, G, gg, Ctrl+d, Ctrl+u)
- [Source: prd.md#NFR6] — Scrolling through a rendered document must feel smooth with no rendering stutter
- [Source: ux-design-specification.md#Cursor & Navigation Patterns] — j/k line movement, h/l character, w/b word, G/gg document, Ctrl+d/Ctrl+u half-page, viewport follows cursor
- [Source: ux-design-specification.md#Normal Mode Navigation] — Complete motion table with behaviors
- [Source: ux-design-specification.md#Mode Transition Patterns] — Normal mode is the resting state, Esc always returns to normal
- [Source: 1-5-open-and-display-existing-markdown-file.md] — EditorModel struct, Update() with hardcoded j/k, Bubbletea v2 API patterns, viewport API

### Critical Validation Points

**Before marking this story done, verify:**

1. **j/k moves cursor line by line** — Cursor visually moves up/down through rendered content, not just scrolling the viewport
2. **h/l moves cursor within a line** — Cursor moves left/right by visible character, clamped to line bounds
3. **w/b moves by word** — Cursor jumps to next/previous word start, crosses line boundaries
4. **G jumps to end** — Cursor goes to last line of document
5. **gg jumps to beginning** — Two-key sequence, cursor goes to first line
6. **Ctrl+d/Ctrl+u scrolls half-page** — Both viewport and cursor move by viewportHeight/2
7. **Viewport follows cursor** — When cursor moves beyond visible area, viewport scrolls to keep it visible
8. **Cursor visible on screen** — Block cursor rendered at correct screen position (accounting for centering margin)
9. **Sticky column on j/k** — Moving to a shorter line clamps col, moving back to a longer line restores original col
10. **No ANSI corruption** — Cursor never lands inside an ANSI escape sequence; all character positions are visible-char based
11. **ctrl+c still exits** — Regression: existing quit behavior preserved
12. **Resize still works** — Regression: terminal resize recalculates layout, cursor position remains valid
13. **Empty document handled** — No crash when navigating with zero content
14. **All tests pass** — `go test ./internal/vim/...`, `go test ./internal/editor/...`, and full suite `go test ./internal/...`

**Acceptance criteria checklist:**
- [x] AC#1: j/k moves cursor down/up by one line within rendered content
- [x] AC#2: h/l moves cursor left/right by one character within rendered content
- [x] AC#3: w/b moves cursor forward/backward by one word
- [x] AC#4: G jumps to end of document
- [x] AC#5: gg jumps to beginning of document
- [x] AC#6: Ctrl+d/Ctrl+u scrolls half a page down/up
- [x] AC#7: Viewport scrolls to keep cursor visible, smooth with no stutter

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Fixed `shift+G` key string handling — bubbletea v2 returns `"shift+G"` not `"G"` when Mod: tea.ModShift is set. Added both variants to NormalHandler.
- ctrl+c now works even before viewport initialization (user should always be able to quit).

### Completion Notes List

- Implemented full vim mode infrastructure in `internal/vim` with Mode type, ModeHandler interface, Action types, NormalHandler, and motion helpers
- NormalHandler supports: j/k (line), h/l (char), w/b (word), G (doc bottom), gg (doc top, multi-key), ctrl+d/ctrl+u (half-page scroll), ctrl+c (quit)
- Added cursor state (cursorLine, cursorCol, desiredCol) to EditorModel with sticky column behavior for j/k
- Refactored EditorModel.Update() from hardcoded key handling to vim mode delegation pattern via applyAction()
- Implemented ensureCursorVisible() for viewport-follows-cursor scrolling
- View() now renders block cursor at correct screen position using tea.NewCursor with left margin offset
- Word motions (w/b) operate across line boundaries using flat-text conversion with ANSI stripping
- Added Line() and SetScrollOffset() getters to ui.Viewport for cursor/editor access
- All ANSI escape sequences properly stripped for character position calculations
- 33 editor tests + 17 vim tests all passing with zero regressions
- Build compiles cleanly, go vet passes

### File List

- `internal/vim/mode.go` — MODIFIED: Added Mode type, ModeHandler interface (was empty package declaration)
- `internal/vim/action.go` — NEW: Action interface and concrete types (NoOp, MoveCursor, WordMotion, Scroll, DocumentPosition, Quit)
- `internal/vim/normal.go` — NEW: NormalHandler with operator-pending state for gg sequence
- `internal/vim/normal_test.go` — NEW: 17 table-driven tests for key dispatch, multi-key, scroll, quit, unknown keys
- `internal/vim/motion.go` — NEW: StripANSI, VisibleLength, NextWordStart, PrevWordStart helpers
- `internal/vim/motion_test.go` — NEW: 16 tests for ANSI stripping, visible length, word boundary detection
- `internal/editor/editor.go` — MODIFIED: Added vim delegation, cursor state, applyAction, ensureCursorVisible, View cursor rendering
- `internal/editor/editor_test.go` — MODIFIED: Updated existing tests for cursor-based behavior, added 20 new tests for all motions
- `internal/ui/viewport.go` — MODIFIED: Added Line() and SetScrollOffset() methods
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — MODIFIED: Updated story status

### Change Log

- 2026-02-11: Implemented normal mode vim navigation — vim mode infrastructure, cursor state, all motions (j/k/h/l/w/b/G/gg/ctrl+d/ctrl+u), viewport follow, comprehensive tests
- 2026-02-11: Code review fixes — Fixed NextWordStart empty-text bug (returned -1), added real assertions to sticky column test, added cross-line word motion tests, removed duplicate ansiRegexp from editor (uses vim.StripANSI), optimized applyWordMotion with bounded window, renamed ScrollAction.Lines→Direction for clarity, added resize-with-cursor clamp test
