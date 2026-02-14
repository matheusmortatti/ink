# Story 2.2: Insert Mode and Text Input

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->
<!-- Epic: Write with Live Preview -->
<!-- Story Key: 2-2-insert-mode-and-text-input -->
<!-- Date: 2026-02-12 -->

## Story

As a writer,
I want to enter insert mode and type prose into a block,
so that I can write and edit my markdown content.

## Acceptance Criteria

1. **Given** the editor is in normal mode with cursor on a rendered block **When** the user presses `i` **Then** insert mode is activated, the block displays raw markdown, and the cursor is positioned at the current location within the raw text (FR6, FR13)

2. **Given** the editor is in normal mode with cursor on a rendered block **When** the user presses `a` **Then** insert mode is activated with the cursor positioned after the current character in the raw text (FR6)

3. **Given** the editor is in normal mode with cursor on a rendered block **When** the user presses `o` **Then** insert mode is activated with a new line created below the cursor line within the block (FR6)

4. **Given** the editor is in normal mode with cursor on a rendered block **When** the user presses `O` **Then** insert mode is activated with a new line created above the cursor line within the block (FR6)

5. **Given** the editor is in insert mode within a block **When** the user types printable characters **Then** characters are inserted at the cursor position via the gap buffer (FR17) **And** keystroke-to-screen latency is imperceptible (NFR3)

6. **Given** the editor is in insert mode **When** the user presses backspace or delete **Then** the character before or after the cursor is removed (FR18)

7. **Given** the editor is in insert mode **When** the user presses Tab **Then** a tab character is inserted at the cursor position (FR21)

8. **Given** the editor is in insert mode **When** the user presses `Esc` **Then** the active block is rendered via Glamour, the document returns to fully rendered state, and normal mode is activated (FR7, FR16)

## Tasks / Subtasks

- [x] Task 1: Add new Action types for insert mode operations (AC: #1-#8)
  - [x] 1.1 Add `ChangeModeAction` to `internal/vim/action.go` — carries target mode and context (cursor position, insert variant: i/a/o/O)
  - [x] 1.2 Add `InsertCharAction` — carries the rune to insert
  - [x] 1.3 Add `BackspaceAction` and `DeleteCharAction` — signal deletion operations
  - [x] 1.4 Add `InsertNewlineAction` — signal Enter key press within a block
  - [x] 1.5 Add `InsertTabAction` — signal Tab key press for markdown indentation

- [x] Task 2: Implement InsertHandler in `internal/vim/insert.go` (AC: #5, #6, #7, #8)
  - [x] 2.1 Create `InsertHandler` struct implementing `ModeHandler` interface
  - [x] 2.2 Map printable characters to `InsertCharAction`
  - [x] 2.3 Map `Esc` to `ChangeModeAction{Mode: Normal}`
  - [x] 2.4 Map `backspace` to `BackspaceAction`
  - [x] 2.5 Map `delete` to `DeleteCharAction`
  - [x] 2.6 Map `enter` to `InsertNewlineAction`
  - [x] 2.7 Map `tab` to `InsertTabAction`
  - [x] 2.8 Map arrow keys to cursor movement within the gap buffer (`left`, `right`, `up`, `down`)

- [x] Task 3: Add insert-mode triggers to NormalHandler (AC: #1, #2, #3, #4)
  - [x] 3.1 Map `i` to `ChangeModeAction{Mode: Insert, Variant: "i"}` — insert at cursor
  - [x] 3.2 Map `a` to `ChangeModeAction{Mode: Insert, Variant: "a"}` — insert after cursor
  - [x] 3.3 Map `o` to `ChangeModeAction{Mode: Insert, Variant: "o"}` — new line below
  - [x] 3.4 Map `O` (shift+o) to `ChangeModeAction{Mode: Insert, Variant: "O"}` — new line above

- [x] Task 4: Add active block and gap buffer state to EditorModel (AC: #1-#8)
  - [x] 4.1 Add `activeBlockIdx int` field — index of the block being edited (-1 when none)
  - [x] 4.2 Add `activeBuffer *block.GapBuffer` field — gap buffer for the active editing block
  - [x] 4.3 Add `blockCursorLine int` and `blockCursorCol int` fields — cursor within the raw block
  - [x] 4.4 Add helper `blockIndexForLine(docLine int) int` — maps document cursor line to block index
  - [x] 4.5 Add helper `docLineToBlockLine(docLine int, blockIdx int) (int, int)` — maps document line to block-relative line and col

- [x] Task 5: Implement enter-insert-mode logic in editor (AC: #1, #2, #3, #4)
  - [x] 5.1 Handle `ChangeModeAction` with target Insert in `applyAction()`
  - [x] 5.2 Determine which block the cursor is on using `blockIndexForLine()`
  - [x] 5.3 Create `GapBuffer` from `block.Raw` for the target block
  - [x] 5.4 Position gap buffer cursor based on variant: `i` → at mapped position, `a` → after mapped position, `o` → new line below current, `O` → new line above current
  - [x] 5.5 Switch `modeHandler` to `InsertHandler`
  - [x] 5.6 Recompose viewport to show raw markdown for the active block

- [x] Task 6: Implement exit-insert-mode logic in editor (AC: #8)
  - [x] 6.1 Handle `ChangeModeAction` with target Normal in `applyAction()` — when coming from Insert
  - [x] 6.2 Commit gap buffer content back to `block.Raw` (`blocks[activeBlockIdx].Raw = activeBuffer.Content()`)
  - [x] 6.3 Invalidate render cache for the modified block
  - [x] 6.4 Re-render the block via Glamour and update cache
  - [x] 6.5 Clear active block state (`activeBlockIdx = -1`, `activeBuffer = nil`)
  - [x] 6.6 Switch `modeHandler` back to `NormalHandler`
  - [x] 6.7 Recompose viewport to fully rendered state
  - [x] 6.8 Map cursor position from raw block back to document-level position

- [x] Task 7: Implement text insertion and deletion in editor (AC: #5, #6, #7)
  - [x] 7.1 Handle `InsertCharAction` — call `activeBuffer.Insert(r)`, update block cursor
  - [x] 7.2 Handle `BackspaceAction` — call `activeBuffer.Backspace()`, update block cursor
  - [x] 7.3 Handle `DeleteCharAction` — call `activeBuffer.Delete()`, update block cursor
  - [x] 7.4 Handle `InsertNewlineAction` — call `activeBuffer.Insert('\n')`, update block cursor
  - [x] 7.5 Handle `InsertTabAction` — call `activeBuffer.Insert('\t')`, update block cursor
  - [x] 7.6 After each operation, recompose the active block display in viewport

- [x] Task 8: Update viewport to support mixed rendered/raw display (AC: #1, #8)
  - [x] 8.1 Add `SetActiveBlock(blockIdx int, rawContent string)` method to Viewport — replaces one block's rendered output with raw markdown text
  - [x] 8.2 Add `ClearActiveBlock()` method — restores full rendered display
  - [x] 8.3 Add `UpdateActiveBlockContent(rawContent string)` method — live-updates the raw block content as user types
  - [x] 8.4 Ensure raw block display uses the same centering/margin as rendered blocks

- [x] Task 9: Update View() for insert mode cursor positioning (AC: #5)
  - [x] 9.1 When in insert mode, compute cursor screen position from active block position + block cursor line/col
  - [x] 9.2 Switch cursor shape to `CursorBar` (line) in insert mode, `CursorBlock` in normal mode
  - [x] 9.3 Ensure cursor stays visible within viewport when editing a block

- [x] Task 10: Write tests (AC: all)
  - [x] 10.1 InsertHandler tests: key mappings for printable chars, Esc, backspace, delete, enter, tab, arrows, space
  - [x] 10.2 NormalHandler insert triggers: i, a, o, O, shift+O → ChangeModeAction with correct variants
  - [x] 10.3 Editor integration tests: enter insert mode, type text, exit — verify block content updated
  - [x] 10.4 Viewport mixed mode tests: SetActiveBlock displays raw, ClearActiveBlock restores rendered

## Dev Notes

### Context & Purpose

This is **Story 2.2 of Epic 2** (Write with Live Preview) — the story that activates ink's defining interaction. Where Story 2.1 built the gap buffer data structure in isolation, this story wires it into the editor to enable actual text input. After this story, a user can press `i` on a rendered block, type prose, press `Esc`, and see the block render. This is the first time ink becomes an *editor*, not just a *viewer*.

**What this story delivers:**
- Insert mode activation via `i`, `a`, `o`, `O` from normal mode
- Text input via the gap buffer (printable characters, backspace, delete, enter, tab)
- Exit to normal mode with `Esc`, rendering the modified block
- The viewport shows one raw block among rendered blocks (the Mode-Unified Block Reveal)

**What this story does NOT deliver (deferred):**
- Syntax dimming of markdown characters in the editing block (Story 2.3)
- Block transitions with no layout shift guarantee (Story 2.4)
- Cursor position mapping between rendered and raw markdown (Story 2.5) — this story uses a simplified approximate mapping
- New block creation via double-Enter (Story 2.6)
- Undo/redo (Story 5.1)
- Auto-pairs (Story 5.2)

**Scope boundary for cursor mapping:** Story 2.5 is dedicated to accurate cursor position mapping between rendered and raw markdown. For THIS story, use a basic approach: when pressing `i`, place the gap buffer cursor at the beginning of the block (line 0, col 0) or at a rough line-based mapping. The raw-to-rendered back-mapping on `Esc` can similarly place the document cursor at the start of the block. This keeps the story focused on the core insert mode mechanics without getting blocked on the complex mapping problem.

### Technical Requirements

**New files to create:**

```
internal/vim/insert.go          # InsertHandler implementing ModeHandler
internal/vim/insert_test.go     # InsertHandler tests
```

**Files to modify:**

```
internal/vim/action.go          # New action types
internal/vim/normal.go          # Add i/a/o/O mappings
internal/vim/normal_test.go     # Test new mappings
internal/editor/editor.go       # Active block state, mode switching, text operations
internal/ui/viewport.go         # Mixed rendered/raw block display
```

**InsertHandler design:**

```go
package vim

// InsertHandler processes key input in insert mode.
type InsertHandler struct{}

func NewInsertHandler() *InsertHandler {
    return &InsertHandler{}
}

func (h *InsertHandler) HandleKey(key string) Action {
    switch key {
    case "esc":
        return ChangeModeAction{Mode: Normal}
    case "backspace":
        return BackspaceAction{}
    case "delete":
        return DeleteCharAction{}
    case "enter":
        return InsertNewlineAction{}
    case "tab":
        return InsertTabAction{}
    case "left":
        return MoveCursorAction{Col: -1, Relative: true}
    case "right":
        return MoveCursorAction{Col: 1, Relative: true}
    case "up":
        return MoveCursorAction{Line: -1, Relative: true}
    case "down":
        return MoveCursorAction{Line: 1, Relative: true}
    default:
        // Single printable character
        runes := []rune(key)
        if len(runes) == 1 && runes[0] >= 32 {
            return InsertCharAction{Char: runes[0]}
        }
        return NoOpAction{}
    }
}

func (h *InsertHandler) Mode() Mode {
    return Insert
}
```

**New action types to add to `internal/vim/action.go`:**

```go
// ChangeModeAction switches the editor to a different vim mode.
type ChangeModeAction struct {
    Mode    Mode   // Target mode
    Variant string // Entry variant: "i", "a", "o", "O" (for insert mode entry)
}

// InsertCharAction inserts a character at the cursor position.
type InsertCharAction struct {
    Char rune
}

// BackspaceAction deletes the character before the cursor.
type BackspaceAction struct{}

// DeleteCharAction deletes the character after the cursor.
type DeleteCharAction struct{}

// InsertNewlineAction inserts a newline at the cursor position.
type InsertNewlineAction struct{}

// InsertTabAction inserts a tab/indentation at the cursor position.
type InsertTabAction struct{}

func (ChangeModeAction) actionTag()    {}
func (InsertCharAction) actionTag()    {}
func (BackspaceAction) actionTag()     {}
func (DeleteCharAction) actionTag()    {}
func (InsertNewlineAction) actionTag() {}
func (InsertTabAction) actionTag()     {}
```

**EditorModel new state fields:**

```go
type EditorModel struct {
    // ... existing fields ...
    activeBlockIdx int              // Index of block being edited, -1 when none
    activeBuffer   *block.GapBuffer // Gap buffer for active editing block
}
```

**Key state flow:**

```
Normal mode: i pressed
→ NormalHandler.HandleKey("i") → ChangeModeAction{Mode: Insert, Variant: "i"}
→ editor.applyAction()
  → blockIdx = blockIndexForLine(cursorLine)
  → activeBlockIdx = blockIdx
  → activeBuffer = block.NewGapBuffer(blocks[blockIdx].Raw)
  → position cursor in gap buffer (start of block for now)
  → modeHandler = vim.NewInsertHandler()
  → viewport.SetActiveBlock(blockIdx, blocks[blockIdx].Raw)
  → update cursor shape to CursorBar

Insert mode: user types 'x'
→ InsertHandler.HandleKey("x") → InsertCharAction{Char: 'x'}
→ editor.applyAction()
  → activeBuffer.Insert('x')
  → viewport.UpdateActiveBlockContent(activeBuffer.Content())
  → update blockCursorLine/Col from gap buffer

Insert mode: Esc pressed
→ InsertHandler.HandleKey("esc") → ChangeModeAction{Mode: Normal}
→ editor.applyAction()
  → blocks[activeBlockIdx].Raw = activeBuffer.Content()
  → cache.Invalidate(blocks[activeBlockIdx], width)
  → renderer.RenderCached(blocks[activeBlockIdx], cache)
  → viewport.ClearActiveBlock()
  → viewport.SetContent(blocks, renderer, cache) (recompose)
  → modeHandler = vim.NewNormalHandler()
  → activeBlockIdx = -1, activeBuffer = nil
  → update cursor shape to CursorBlock
```

### Architecture Compliance

**Package: `internal/vim`** — InsertHandler lives here alongside NormalHandler, following the per-mode handler pattern. New action types in `action.go`.

**Package: `internal/editor`** — All state mutations happen in `applyAction()`. The editor coordinates between vim handlers, gap buffer, viewport, and render cache.

**Package: `internal/ui`** — Viewport gains methods to display mixed rendered/raw content.

**Dependency direction (MUST follow):**
```
internal/editor → internal/block, internal/render, internal/vim, internal/ui
internal/vim → (no internal dependencies — returns Action types only)
internal/ui → internal/block, internal/render (existing)
internal/block → (leaf package)
```

**CRITICAL: InsertHandler must NOT import any other internal package.** It returns Action types. The editor interprets and applies them. This preserves the unidirectional dependency flow.

**Naming conventions (enforce strictly):**
- `InsertHandler` (not `InsertMode` or `InsertModeHandler`)
- Receiver: `h` for handler — `func (h *InsertHandler) HandleKey(key string) Action`
- Action types: noun-based — `InsertCharAction`, `BackspaceAction`, `ChangeModeAction`
- Editor methods: verb-first — `applyAction`, `blockIndexForLine`, `enterInsertMode`, `exitInsertMode`

### Library & Framework Requirements

| Library | Import Path | Usage in This Story |
|---|---|---|
| Bubbletea v2 | `charm.land/bubbletea/v2` | `tea.KeyPressMsg`, `tea.CursorBar`, `tea.CursorBlock`, `tea.View` |
| Go stdlib | `unicode`, `strings` | Rune detection for printable characters |

**No new external dependencies required for this story.**

**Bubbletea v2 key handling notes:**
- `tea.KeyPressMsg.String()` returns key names: `"esc"`, `"backspace"`, `"delete"`, `"enter"`, `"tab"`, `"left"`, `"right"`, `"up"`, `"down"`
- Printable characters return the character itself: `"a"`, `"A"`, `"1"`, `" "` (space), etc.
- `"shift+o"` is how Bubbletea v2 represents uppercase O — check if `msg.String()` returns `"O"` or `"shift+o"` for capital letters and handle accordingly
- Arrow keys in insert mode should be mapped to gap buffer movement, not document-level cursor movement

### File Structure Requirements

**Files to create (2 new):**

```
internal/vim/insert.go          # NEW — InsertHandler struct and methods
internal/vim/insert_test.go     # NEW — InsertHandler test suite
```

**Files to modify (5 existing):**

```
internal/vim/action.go          # ADD 6 new action types
internal/vim/normal.go          # ADD i/a/o/O key mappings
internal/vim/normal_test.go     # ADD tests for new mappings
internal/editor/editor.go       # ADD active block state, mode switching, text ops
internal/ui/viewport.go         # ADD mixed rendered/raw display methods
```

**Total: 7 files (2 new, 5 modified)**

### Testing Requirements

**Test location:** Co-located per Go convention

**Test naming:** `TestType_Scenario_ExpectedBehavior`

**Required test categories:**

| Category | File | What to Test | Min Cases |
|---|---|---|---|
| InsertHandler keys | `insert_test.go` | Printable char → InsertCharAction, Esc → ChangeModeAction, backspace → BackspaceAction, etc. | 8 |
| NormalHandler insert | `normal_test.go` | i/a/o/O → ChangeModeAction with correct Mode and Variant | 4 |
| Viewport mixed mode | `viewport_test.go` | SetActiveBlock shows raw, ClearActiveBlock restores rendered, UpdateActiveBlockContent | 3 |

**Run all tests:** `go test ./internal/...` (zero regressions)

### Project Structure Notes

- `internal/vim/insert.go` follows the same pattern as `internal/vim/normal.go` — struct implementing `ModeHandler`
- New action types follow the existing sealed interface pattern (unexported `actionTag()` method)
- The `EditorModel` gains active block tracking state but no new packages or directories
- Viewport's mixed display is additive — existing `composeBlocks()` still works for fully-rendered state

### Previous Story Intelligence

**From Story 2.1 (Gap Buffer) — immediate predecessor:**

- GapBuffer is fully implemented at `internal/block/gapbuffer.go` with 16 public methods
- Constructor: `NewGapBuffer(content string) *GapBuffer` — cursor starts at end of content
- Key methods this story uses: `Insert(r rune)`, `InsertString(s string)`, `Backspace() bool`, `Delete() bool`, `Content() string`, `MoveLeft()`, `MoveRight()`, `MoveUp()`, `MoveDown()`, `CursorLineCol() (int, int)`, `SetCursorLineCol(line, col int)`, `MoveToLineEnd()`
- Line/col are 0-indexed throughout
- The gap buffer operates on `[]rune` — Unicode-safe
- `Backspace()` and `Delete()` return `bool` (false if at boundary) — use return value, don't need separate boundary checks
- No ANSI awareness needed — gap buffer works on raw markdown text only

**Code conventions established across Epic 1:**
- Import grouping: stdlib, external (`charm.land/bubbletea/v2`), internal
- Receiver names: single letter (`e` for EditorModel, `v` for Viewport, `h` for handler)
- Exported functions have doc comments
- Table-driven tests with `t.Run` subtests

**From Story 1.6 (Normal Mode Vim Navigation):**
- `StripANSI()` and `VisibleLength()` in `internal/vim/motion.go` handle ANSI-aware text processing for rendered content
- Cursor in normal mode is document-level: `(cursorLine, cursorCol)` over rendered/ANSI content
- `desiredCol` tracks sticky column for j/k navigation — must be preserved when returning from insert mode
- `leftMargin()` calculates centering offset — raw block in viewport needs same margin

### Git Intelligence

**Recent commits (newest first):**
```
de5bd1b gap buffer
8edff57 basic vim motion navigation
bbd9172 open and display existing markdown file
dbeea60 document viewport with centered writing column
c1d62f3 block rendering
dc73dfb block parser
78e9544 initial folder structure and example main.go file
```

**Commit convention:** Short, lowercase, descriptive. No prefixes, no ticket numbers.
**Expected commit for this story:** `insert mode and text input`

**Patterns from commit history:**
- Each story = one commit
- New handler files follow `internal/vim/{mode}.go` pattern
- Test files co-located: `*_test.go`

### References

- [Source: architecture.md#Vim Mode Architecture] — Per-mode handler pattern, ModeHandler interface
- [Source: architecture.md#Component Communication] — Single EditorModel with component structs, action delegation
- [Source: architecture.md#Package Boundary Rules] — vim package has no internal dependencies
- [Source: architecture.md#EditorModel Delegation Pattern] — Handlers return actions, editor applies
- [Source: architecture.md#Block & Document Conventions] — Modified blocks serialize from gap buffer content
- [Source: architecture.md#Cursor Position Representation] — (line, col) within block, (blockIndex, line, col) within document
- [Source: epics.md#Story 2.2] — Full acceptance criteria and user story
- [Source: prd.md#FR6] — Enter block via vim commands
- [Source: prd.md#FR7] — Exit block with Esc, renders and returns to normal
- [Source: prd.md#FR13] — Insert mode editing within active block
- [Source: prd.md#FR16] — Esc always returns to normal mode
- [Source: prd.md#FR17] — Text insertion at cursor
- [Source: prd.md#FR18] — Delete with backspace and delete keys
- [Source: prd.md#FR21] — Tab key for markdown indentation
- [Source: prd.md#NFR3] — Keystroke-to-screen latency imperceptible
- [Source: ux-design-specification.md#Mode-Unified Block Reveal] — The novel pattern this story implements
- [Source: ux-design-specification.md#Experience Mechanics] — Entering/editing/leaving blocks
- [Source: 2-1-gap-buffer-for-block-text-editing.md] — GapBuffer API, conventions, existing test patterns

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Discovered bubbletea v2 key representation: `{Code: ' '}.String()` returns `"space"` (not `" "`), and `{Code: 'O', Mod: tea.ModShift}.String()` returns `"shift+O"` (not `"O"`). Added explicit `"space"` case in InsertHandler and `"shift+O"` case in NormalHandler.

### Completion Notes List

- All 10 tasks and subtasks implemented and tested
- 6 new action types added to vim sealed Action interface
- InsertHandler implements ModeHandler with full key mapping (printable chars, space, esc, backspace, delete, enter, tab, arrows)
- NormalHandler extended with i/a/o/O/shift+O triggers for insert mode entry
- EditorModel gains activeBlockIdx and activeBuffer fields for insert mode state
- enterInsertMode/exitInsertMode methods handle full lifecycle: gap buffer creation, cursor positioning by variant, mode switching, viewport composition
- Viewport gains SetActiveBlock/ClearActiveBlock/UpdateActiveBlockContent for mixed rendered/raw display
- Viewport tracks block line ranges via blockRanges for block-to-line mapping
- Cursor shape switches between CursorBar (insert) and CursorBlock (normal)
- Simplified cursor mapping per Dev Notes scope: cursor placed at block start on 'i', block start on exit
- All 8 acceptance criteria satisfied
- Zero test regressions across all packages

### Change Log

- 2026-02-12: Implemented insert mode and text input (Story 2.2)
- 2026-02-13: Code review fixes — fixed `o` variant for multi-line blocks (MoveToEnd→MoveToLineEnd), removed double recomposition in exitInsertMode, fixed `a` variant safety for empty blocks, optimized per-keystroke viewport recomposition, added error propagation to viewport SetActiveBlock/ClearActiveBlock, added multi-line `o` test and delete key integration test

### File List

- internal/vim/action.go (modified) — 6 new action types: ChangeModeAction, InsertCharAction, BackspaceAction, DeleteCharAction, InsertNewlineAction, InsertTabAction
- internal/vim/insert.go (new) — InsertHandler implementing ModeHandler
- internal/vim/insert_test.go (new) — InsertHandler test suite (10 test functions)
- internal/vim/normal.go (modified) — Added i/a/o/O/shift+O key mappings
- internal/vim/normal_test.go (modified) — Added insert mode trigger tests
- internal/editor/editor.go (modified) — Active block state, mode switching, text operations, cursor positioning; review fixes: o/O/a variant corrections, error handling, removed double recomposition
- internal/editor/editor_test.go (modified) — Added 11 insert mode integration tests (including multi-line o and delete key)
- internal/ui/viewport.go (modified) — Mixed rendered/raw display with block range tracking; review fixes: SetActiveBlock/ClearActiveBlock return errors, optimized UpdateActiveBlockContent
- internal/ui/viewport_test.go (modified) — Added 3 viewport mixed mode tests; updated for error-returning API
