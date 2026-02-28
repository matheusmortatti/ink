# Story 3.1: Status Bar with Mode Display and Word Count

Status: done

<!-- Epic: 3 - Status Bar and Editor Feedback -->
<!-- FRs: FR30, FR31 -->
<!-- Date: 2026-02-26 -->

## Story

As a writer,
I want a centered status bar showing my current mode, word count, and character count,
so that I always know my editor state and writing progress at a glance.

## Acceptance Criteria

1. **Given** the editor is in normal mode **When** the status bar is displayed **Then** it shows `NORMAL · {words}w · {chars}c` centered within the terminal width (FR30, FR31)

2. **Given** the editor is in insert mode **When** the status bar is displayed **Then** it shows `INSERT · {words}w · {chars}c` centered within the terminal width (FR30, FR31)

3. **Given** the editor is in visual mode **When** the status bar is displayed **Then** it shows `VISUAL · {words}w · {chars}c` centered within the terminal width (FR30)

4. **Given** the user types or deletes text in insert mode **When** the document content changes **Then** the word count and character count update in real-time without perceptible delay (FR31, NFR4)

5. **Given** any mode is active **When** the status bar is displayed **Then** the mode is communicated via text label, not color alone (NFR13) **And** the status bar is positioned consistently at the bottom row of the terminal (NFR16)

6. **Given** a terminal height of 10+ rows **When** the layout is calculated **Then** the status bar occupies the last row with 1 blank line separating it from content

7. **Given** a terminal height of 5–9 rows **When** the layout is calculated **Then** the status bar occupies the last row with no blank separator (content area shrinks by 1)

8. **Given** a terminal height below 5 rows **When** the layout is calculated **Then** the status bar is hidden and the full height is used for content

## Tasks / Subtasks

- [x] Task 1: Add `StatusBarDimPercent` constant to `internal/render/color.go` (AC: #2, #7)
  - [x] 1.1 Add `const StatusBarDimPercent = 0.7` alongside the existing `SyntaxDimPercent = 0.6`

- [x] Task 2: Create `internal/ui/statusbar.go` with `StatusBar` component (AC: #1–#6)
  - [x] 2.1 Define `StatusBar` struct with fields: `width int`, `modeLabel string`, `wordCount int`, `charCount int`, `dimStyle lipgloss.Style`
  - [x] 2.2 Implement `NewStatusBar(width int) *StatusBar` — initializes with "NORMAL" label, zero counts, `render.DimStyle(render.StatusBarDimPercent)` dim style
  - [x] 2.3 Implement `(s *StatusBar) Set(modeLabel string, words, chars int)` — updates all three fields at once
  - [x] 2.4 Implement `(s *StatusBar) Resize(width int)` — updates the terminal width
  - [x] 2.5 Implement `(s *StatusBar) View(dimmed bool) string` — formats `"{MODE} · {words}w · {chars}c"`, centers it within `s.width` using manual left-pad (same approach as viewport margins: `strings.Repeat(" ", padLeft) + text`), applies `s.dimStyle.Render()` to the entire centered string when `dimmed == true`

- [x] Task 3: Add `statusBarRows()` helper and `computeDocumentCounts()` to `internal/editor/editor.go` (AC: #6, #7, #8)
  - [x] 3.1 Add package-level (unexported) function `statusBarRows(terminalHeight int) int` — returns 2 if height >= 10, 1 if 5–9, 0 if < 5
  - [x] 3.2 Add `(e *EditorModel) computeDocumentCounts() (words, chars int)` — iterates `e.blocks`, using `e.activeBuffer.Content()` for the active block (if any) instead of `block.Raw` to reflect live edits; counts chars as `len([]rune(raw))`, words as `len(strings.Fields(raw))`
  - [x] 3.3 Add `(e *EditorModel) refreshStatusBar()` — calls `e.computeDocumentCounts()` and calls `e.statusBar.Set(e.modeHandler.Mode().String(), words, chars)`; no-ops if `e.statusBar == nil`

- [x] Task 4: Integrate `StatusBar` into `EditorModel` struct and initialization (AC: #1–#8)
  - [x] 4.1 Add `statusBar *ui.StatusBar` field to `EditorModel` struct
  - [x] 4.2 In `initViewport()`: compute `sbRows := statusBarRows(e.height)` BEFORE creating the viewport; pass `e.height - sbRows` instead of `e.height` to `ui.NewViewport()`; create `e.statusBar = ui.NewStatusBar(e.width)`; call `e.refreshStatusBar()` at the end of `initViewport()` (after `e.ready = true` and any blank-canvas startup)
  - [x] 4.3 In `Update()` `WindowSizeMsg` branch (the `else` branch for resize): compute `sbRows := statusBarRows(e.height)` and call `e.viewport.Resize(e.width, e.height-sbRows)` instead of `e.viewport.Resize(e.width, e.height)`; call `e.statusBar.Resize(e.width)` if non-nil

- [x] Task 5: Update `View()` to compose viewport content + status bar (AC: #1–#8)
  - [x] 5.1 After building viewport content with `e.viewport.View()`, compute `sbRows := statusBarRows(e.height)`
  - [x] 5.2 If `sbRows > 0` and `e.statusBar != nil`: append `"\n\n"` + status bar (when sbRows == 2) or `"\n"` + status bar (when sbRows == 1) to the viewport string before passing to `tea.NewView()`
  - [x] 5.3 If `sbRows == 0` or `e.statusBar == nil`: use viewport content only (no change from current behavior)
  - [x] 5.4 Cursor positioning logic in `View()` is unchanged — it operates within the viewport area only

- [x] Task 6: Refresh status bar after every state change (AC: #4)
  - [x] 6.1 Add `e.refreshStatusBar()` call at the end of `applyAction()` (just before `return e, nil`), covering: mode changes, text edits, block splits, and any navigation that might change mode
  - [x] 6.2 Verify `refreshStatusBar()` is also called from the blank-canvas startup path in `initViewport()` so the status bar shows "INSERT · 0w · 0c" on startup

- [x] Task 7: Create `internal/ui/statusbar_test.go` (AC: #1–#8)
  - [x] 7.1 `TestStatusBar_View_NormalMode` — verifies format "NORMAL · 42w · 186c", no ANSI codes
  - [x] 7.2 `TestStatusBar_View_InsertMode_Dimmed` — verifies ANSI escape codes present when dimmed=true
  - [x] 7.3 `TestStatusBar_View_Centering` — verifies left padding = `(width - len("NORMAL · 0w · 0c")) / 2` for a given width
  - [x] 7.4 `TestStatusBar_Resize` — verifies centering adjusts after Resize()
  - [x] 7.5 `TestComputeDocumentCounts_NormalMode` — multiple blocks, counts chars/words from Raw
  - [x] 7.6 `TestComputeDocumentCounts_InsertMode_UsesLiveBuffer` — active buffer content overrides block.Raw in count
  - [x] 7.7 `TestStatusBarRows_HeightThresholds` — verifies return values at heights 4, 5, 9, 10, 40
  - [x] 7.8 `TestEditorModel_ViewportHeight_ReducedByStatusBar` — after WindowSizeMsg(width=120, height=40), viewport.ViewportHeight() == 38 (40-2)
  - [x] 7.9 `TestEditorModel_ViewportHeight_SmallTerminal_5Rows` — height=7, viewport height == 6 (7-1)
  - [x] 7.10 `TestEditorModel_ViewportHeight_TooSmall_HidesStatusBar` — height=4, viewport height == 4 (no reduction)
  - [x] 7.11 `TestEditorModel_StatusBar_UpdatesOnInsert` — type a char in insert mode, verify char count increases in status bar
  - [x] 7.12 `TestEditorModel_StatusBar_ModeLabel` — verify status bar shows "NORMAL" after Esc, "INSERT" after entering insert

## Dev Notes

### Architecture and Implementation Guidance

**Overview:** This story adds a centered status bar to the bottom of the terminal. It introduces one new file (`internal/ui/statusbar.go`), modifies three existing files (`internal/render/color.go`, `internal/editor/editor.go`), and adds a constant + test file.

**Zero breaking changes to existing behavior:** All existing functionality (viewport rendering, cursor movement, insert mode, block transitions) remains unchanged. The only change to existing files is reducing the viewport height by `statusBarRows()` and appending the status bar line in `View()`.

---

### Critical Implementation Details

#### 1. Viewport Height Reduction

The viewport currently receives the FULL terminal height:
```go
// CURRENT in initViewport():
e.viewport = ui.NewViewport(e.width, e.height)
```

Change to:
```go
// NEW in initViewport():
sbRows := statusBarRows(e.height)
e.viewport = ui.NewViewport(e.width, e.height-sbRows)
```

And in the `WindowSizeMsg` resize branch:
```go
// CURRENT:
_ = e.viewport.Resize(e.width, e.height)
// NEW:
sbRows := statusBarRows(e.height)
_ = e.viewport.Resize(e.width, e.height-sbRows)
if e.statusBar != nil {
    e.statusBar.Resize(e.width)
}
```

**Why this works:** `Viewport.ViewportHeight()` returns `layout.TerminalHeight`, which is used in `ensureCursorVisible()`, `ensureInsertCursorVisible()`, `ScrollAction`, and `clampScroll()`. Reducing this value ensures all scroll and cursor visibility calculations stay within the writing area, never overlapping the status bar rows.

#### 2. `View()` Composition

```go
func (e *EditorModel) View() tea.View {
    if !e.ready {
        return tea.NewView("loading...")
    }

    viewContent := e.viewport.View()
    sbRows := statusBarRows(e.height)
    var content string
    if sbRows > 0 && e.statusBar != nil {
        isDimmed := e.modeHandler.Mode() == vim.Insert
        sbLine := e.statusBar.View(isDimmed)
        if sbRows == 2 {
            content = viewContent + "\n\n" + sbLine
        } else { // sbRows == 1
            content = viewContent + "\n" + sbLine
        }
    } else {
        content = viewContent
    }

    v := tea.NewView(content)
    v.AltScreen = true

    // Cursor positioning — unchanged from current implementation
    if e.activeBlockIdx >= 0 && e.activeBuffer != nil {
        bufLine, bufCol := e.activeBuffer.CursorLineCol()
        screenRow, screenCol := e.viewport.BufferToScreenPos(bufLine, bufCol)
        screenRow -= e.viewport.ScrollOffset()
        v.Cursor = tea.NewCursor(screenCol, screenRow)
        v.Cursor.Shape = tea.CursorBar
    } else {
        screenRow := e.cursorLine - e.viewport.ScrollOffset()
        screenCol := e.leftMargin() + e.cursorCol
        v.Cursor = tea.NewCursor(screenCol, screenRow)
        v.Cursor.Shape = tea.CursorBlock
    }

    return v
}
```

**String layout math** (for height=40, sbRows=2):
- `viewport.View()` = 38 lines joined by 37 `\n` chars
- Appending `"\n\n"` creates 2 more newlines → lines 38 (blank) and 39 (status bar start)
- `sbLine` = the centered status bar text (no newline)
- Total: 40 rows on screen ✓

#### 3. `StatusBar.View()` — Centering Implementation

Use the same manual-pad approach as the viewport's margin centering:
```go
func (s *StatusBar) View(dimmed bool) string {
    plain := fmt.Sprintf("%s · %dw · %dc", s.modeLabel, s.wordCount, s.charCount)
    // Center the plain text string
    padLeft := (s.width - len([]rune(plain))) / 2
    if padLeft < 0 {
        padLeft = 0
    }
    centered := strings.Repeat(" ", padLeft) + plain
    if dimmed {
        return s.dimStyle.Render(centered)
    }
    return centered
}
```

**Important:** Apply dimming to the FULL centered string (including padding spaces). This ensures consistent behavior when Lip Gloss applies background color. The middle-dot separator is U+00B7 (`·`), typed as a literal in the format string.

#### 4. `computeDocumentCounts()` — Live Buffer Override

```go
func (e *EditorModel) computeDocumentCounts() (words, chars int) {
    for i, b := range e.blocks {
        raw := b.Raw
        // Use live gap buffer content for active block (reflects uncommitted edits)
        if i == e.activeBlockIdx && e.activeBuffer != nil {
            raw = e.activeBuffer.Content()
        }
        chars += len([]rune(raw))
        words += len(strings.Fields(raw))
    }
    return
}
```

**Why:** When the user is typing, `e.blocks[activeBlockIdx].Raw` still holds the content from when insert mode was entered. Only the gap buffer has the current typed text. Using the live buffer content makes the count update in real-time with each keystroke (FR31).

#### 5. `refreshStatusBar()` Placement

Add at the END of `applyAction()`, just before `return e, nil`:
```go
func (e *EditorModel) applyAction(action vim.Action) (tea.Model, tea.Cmd) {
    switch a := action.(type) {
    // ... all existing cases unchanged ...
    }
    e.refreshStatusBar() // ← add this line
    return e, nil
}
```

**Exception:** `QuitAction` returns early — that's fine, no need to update status bar when quitting.

Also add `e.refreshStatusBar()` at the end of `initViewport()` (after the blank-canvas startup block), so the initial display shows correct counts.

#### 6. Package Boundary: `StatusBar` cannot import `internal/vim`

`internal/ui` cannot import `internal/vim` per the dependency rules:
```
internal/ui → internal/block  (allowed)
internal/ui → internal/render (already used by viewport.go)
internal/ui ↛ internal/vim    (NOT allowed — siblings under editor)
```

Solution: `StatusBar` takes the mode as a `string` parameter (the editor calls `.Mode().String()` before passing). This is consistent with the interface pattern where components receive only what they need.

#### 7. `render/color.go` — New Constant

Add alongside `SyntaxDimPercent`:
```go
const (
    SyntaxDimPercent    = 0.6 // syntax chars in active block: ~40% visible
    StatusBarDimPercent = 0.7 // status bar in insert mode: ~30% visible
)
```

This keeps all dim percentages in one place per the architecture's color interpolation utility mandate.

---

### Files to Create

- `internal/ui/statusbar.go` — `StatusBar` struct with `NewStatusBar()`, `Set()`, `Resize()`, `View()`
- `internal/ui/statusbar_test.go` — Unit tests for StatusBar, statusBarRows, computeDocumentCounts

### Files to Modify

- `internal/render/color.go` — Add `StatusBarDimPercent = 0.7` constant
- `internal/editor/editor.go`:
  - Add `statusBar *ui.StatusBar` field to `EditorModel`
  - Add `statusBarRows()` function
  - Add `computeDocumentCounts()` method
  - Add `refreshStatusBar()` method
  - Modify `initViewport()`: reduce viewport height by sbRows, create StatusBar, call refreshStatusBar
  - Modify `Update()` WindowSizeMsg branch: reduce viewport height by sbRows, resize StatusBar
  - Modify `View()`: compose viewport + status bar
  - Modify `applyAction()`: call refreshStatusBar at end

**No changes to:** `internal/ui/viewport.go`, `internal/ui/layout.go`, `internal/vim/`, `internal/block/`, `cmd/ink/main.go`

---

### Required Imports for `statusbar.go`

```go
package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/matheusmortatti/ink/internal/render"
)
```

Note: `internal/block` is NOT needed in `statusbar.go` because the editor computes counts itself and calls `Set()` — the StatusBar only stores and renders pre-computed values.

---

### `statusBarRows()` Function

```go
// statusBarRows returns how many terminal rows the status bar occupies,
// based on terminal height. Returns 2 (content + blank separator + status bar)
// for 10+ rows, 1 (status bar only, no separator) for 5-9 rows, 0 (hidden) for <5 rows.
func statusBarRows(terminalHeight int) int {
    if terminalHeight >= 10 {
        return 2
    }
    if terminalHeight >= 5 {
        return 1
    }
    return 0
}
```

This matches the UX spec's terminal height adaptation table exactly.

---

### Edge Cases

1. **Zero-content document**: `computeDocumentCounts()` on empty blocks → `0w · 0c`. The blank canvas starts in insert mode showing `INSERT · 0w · 0c` dimmed. ✓

2. **Very long status bar text**: If the terminal is very narrow, `padLeft` calculation may go negative. The `if padLeft < 0 { padLeft = 0 }` guard prevents negative repeat. ✓

3. **Terminal resize from sbRows=0 to sbRows=2**: When terminal grows from height 4 → height 40, the `WindowSizeMsg` handler recomputes sbRows=2, calls `Resize(width, 38)` on the viewport. The status bar gets initialized. ✓ But wait — `e.statusBar` might be nil if the terminal started small (height < 5) and never had a status bar created. Handle: create the status bar in `initViewport()` unconditionally (even for small terminals), just don't display it when sbRows==0. The View() already guards with `if sbRows > 0 && e.statusBar != nil`.

   **Actually**: Always create `e.statusBar = ui.NewStatusBar(e.width)` in `initViewport()`, regardless of terminal height. The `View()` and resize logic handle display/hide via sbRows. This avoids nil-check complexity.

4. **Unicode in document**: `len([]rune(raw))` counts Unicode code points, not bytes. A string with emojis counts each emoji as 1 character. `strings.Fields()` correctly splits on any Unicode whitespace. ✓

5. **Active block mid-edit**: When the user is typing and calls `e.activeBuffer.Content()`, the gap buffer returns the correct current string including all inserts. `strings.Fields()` handles empty strings gracefully (returns empty slice → 0 words). ✓

---

### Testing Requirements

**Test file:** `internal/ui/statusbar_test.go` (new file, package `ui`)
**Test file:** `internal/editor/editor_test.go` (existing file, extend with new tests)

**Test naming convention:** `TestFunctionName_Scenario_ExpectedBehavior`

**New tests for `statusbar_test.go`:**

```go
// TestStatusBar_View_NormalMode
// Create StatusBar with width=40, Set("NORMAL", 5, 25)
// View(false) should contain "NORMAL · 5w · 25c", no ANSI escapes
//
// TestStatusBar_View_InsertMode_Dimmed
// View(true) should contain ANSI escape sequences (\x1b[)
//
// TestStatusBar_View_Centering
// Status bar text "NORMAL · 0w · 0c" has len 17
// For width=40: padLeft = (40-17)/2 = 11
// View(false) should start with 11 spaces
//
// TestStatusBar_Resize
// Start with width=40, Resize(80), verify re-centering at new width
//
// TestStatusBarRows_HeightThresholds (table-driven)
// heights: 4→0, 5→1, 9→1, 10→2, 40→2
```

**New tests for `editor_test.go`:**

```go
// TestStatusBarRows_Values (unit, package editor)
// tests statusBarRows() at thresholds 4, 5, 9, 10, 40

// TestEditorModel_ViewportHeight_ReducedForStatusBar
// NewEditor + WindowSizeMsg(120, 40)
// viewport.ViewportHeight() should == 38 (40 - 2 for sbRows=2)

// TestEditorModel_ViewportHeight_SmallTerminal
// WindowSizeMsg(80, 7) → viewport height == 6 (7 - 1 for sbRows=1)

// TestEditorModel_ViewportHeight_TinyTerminal
// WindowSizeMsg(40, 4) → viewport height == 4 (4 - 0, no status bar)

// TestEditorModel_StatusBar_InitialMode_NormalMode
// Editor with content, after WindowSizeMsg, statusBar.modeLabel should be "NORMAL"

// TestEditorModel_StatusBar_InitialMode_BlankCanvas
// Editor with nil blocks, after WindowSizeMsg, statusBar.modeLabel should be "INSERT"

// TestEditorModel_ComputeDocumentCounts_MultipleBlocks
// blocks = ["hello world", "foo"]
// computeDocumentCounts() → words=3, chars=14

// TestEditorModel_ComputeDocumentCounts_ActiveBufferOverrides
// block.Raw = "old", activeBuffer content = "new text"
// computeDocumentCounts() should use "new text" not "old"
```

---

### Performance Notes

- `computeDocumentCounts()` iterates all blocks on every keypress. For a 10,000-word document (~50-200 blocks), this is a loop over 50-200 short strings — sub-microsecond. Well within NFR4 (imperceptible keystroke latency). No optimization needed.
- `statusBar.View()` is called once per frame render. Pure string formatting — negligible cost.
- No Glamour re-rendering, no cache operations, no goroutines.

### Project Structure Notes

- All changes confined to existing packages — no new packages introduced
- `internal/ui/statusbar.go` follows the same package structure as `viewport.go` and `layout.go`
- `StatusBar` is a plain Go struct, NOT a Bubbletea model (consistent with component delegation pattern)
- The editor remains the sole Bubbletea model; StatusBar is a component it owns and calls directly
- Dependency direction respected: `statusbar.go` imports only `render` and `lipgloss` (external) — no `vim` import

### References

- [Source: epics.md#Story 3.1] — Acceptance criteria and user story statement
- [Source: architecture.md#Component Communication] — EditorModel delegation pattern; StatusBar is a component struct, not a Bubbletea model
- [Source: architecture.md#Project Structure] — `internal/ui/statusbar.go` file location specified
- [Source: architecture.md#Package Boundary Rules] — `internal/ui` cannot import `internal/vim`; mode passed as string
- [Source: architecture.md#Requirements to Structure Mapping] — FR30-35: `ui/statusbar` + `vim/mode`
- [Source: architecture.md#Error Handling] — `EditorModel.showError()` pattern (future, not this story)
- [Source: render/color.go] — `DimStyle(percent)` already exists; add `StatusBarDimPercent = 0.7`
- [Source: ui/viewport.go#composeBlocks] — margin centering uses `strings.Repeat(" ", margin)` — StatusBar uses same pattern
- [Source: ui/layout.go] — `NewViewport(width, height)` signature; `TerminalHeight` used in scroll calculations
- [Source: ux-design-specification.md#Status Bar Component] — format `{MODE} · {words}w · {chars}c`, middle-dot `·` (U+00B7), centered
- [Source: ux-design-specification.md#Color System] — insert mode: foreground shifted 70% toward background → `StatusBarDimPercent = 0.7`
- [Source: ux-design-specification.md#Terminal Height Adaptation] — 10+ rows: +blank+bar; 5-9: +bar; <5: hidden
- [Source: ux-design-specification.md#Mode Transition Behavior] — status bar snaps instantly, no animation
- [Source: prd.md#FR30] — Mode display in status bar
- [Source: prd.md#FR31] — Real-time word/char count updates
- [Source: 2-6-new-block-creation-and-blank-canvas-startup.md] — Pattern: `refreshStatusBar()` mirrors `updateActiveBlockDisplay()` update-after-action pattern

### Previous Story Intelligence (from Story 2.6)

**Key patterns established:**
- `enterInsertMode()` and `exitInsertMode()` are the canonical mode-switch paths — `refreshStatusBar()` placed at end of `applyAction()` will naturally be called after both, since they are invoked from `ChangeModeAction` case. No special handling needed.
- `splitActiveBlock()` leaves the editor in insert mode with a new empty block — `refreshStatusBar()` called after will show correct "INSERT · Xw · Xc" with the new block counted but empty.
- The blank canvas startup in `initViewport()` sets `e.modeHandler = vim.NewInsertHandler()` before returning — `refreshStatusBar()` called at the end of `initViewport()` will correctly read "INSERT" mode.
- `isContentEmpty()` check in `initViewport()` runs after `e.ready = true` — status bar refresh should also happen after this block (at the very end of `initViewport()`).

**Test patterns from 2.6:**
- Table-driven tests with descriptive names
- Integration tests simulate `NewEditor` + `WindowSizeMsg` → verify state
- All 7 packages must pass: run `go test ./...` — `internal/block`, `internal/render`, `internal/vim`, `internal/ui`, `internal/editor`, `internal/file`, `internal/config`

**Editor field access in tests:** The test file is `package editor` (same package), so it can access unexported fields like `e.statusBar`, `e.viewport`, `e.modeHandler`. Use this for test assertions.

### Git Intelligence Summary

**Recent commits:**
```
3ff7b82 epic 2 retro with new golden file tests
0e89d2b new block creation and blank canvas startup
c827492 cursor position mapping between rendered and raw
0e9ef22 fix hang in code when moving to insert mode
d184dcd block transitions the mode unified block reveal
```

**Patterns observed:**
- Single commit per story, lowercase concise message
- Story 3.1 commit should be: `status bar with mode display and word count`
- Typical change footprint: 2-3 files modified + 1-2 files created

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

No blocking issues. Two minor fixes during implementation:
1. `ui.StatusBar` fields are unexported — added `ModeLabel()` and `Counts()` accessor methods to support editor package tests.
2. Lipgloss suppresses ANSI codes in non-TTY test environments — fixed `TestStatusBar_View_InsertMode_Dimmed` to call `lipgloss.SetColorProfile(termenv.TrueColor)` before assertion.

### Completion Notes List

- Implemented `StatusBarDimPercent = 0.7` constant in `internal/render/color.go` alongside `SyntaxDimPercent`
- Created `internal/ui/statusbar.go`: `StatusBar` struct with `NewStatusBar()`, `Set()`, `Resize()`, `View()`, `ModeLabel()`, `Counts()` — plain component struct, not a Bubbletea model
- Added `statusBarRows()`, `computeDocumentCounts()`, `refreshStatusBar()` to `internal/editor/editor.go`
- Modified `initViewport()` to reduce viewport height by `statusBarRows(e.height)` and always create `e.statusBar`
- Modified `Update()` WindowSizeMsg resize branch to account for status bar rows
- Modified `View()` to compose viewport content + status bar separator + status bar line
- Added `e.refreshStatusBar()` at end of `applyAction()` for real-time updates
- All 12 story subtests pass across 7 packages (zero regressions)

### File List

- `internal/render/color.go` (modified — added `StatusBarDimPercent` constant)
- `internal/ui/statusbar.go` (created — StatusBar component)
- `internal/ui/statusbar_test.go` (created — unit tests for StatusBar)
- `internal/editor/editor.go` (modified — statusBar field, statusBarRows, computeDocumentCounts, refreshStatusBar, View, initViewport, Update, applyAction)
- `internal/editor/editor_test.go` (modified — added 8 new status bar tests)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified — story status tracking)

## Change Log

- 2026-02-27: Implemented story 3-1. Added centered status bar showing mode/word/char count at bottom of terminal. StatusBar component created in internal/ui; integrated into EditorModel with height-aware layout (sbRows=2 for 10+ rows, 1 for 5–9, 0 for <5). Real-time count updates on every keypress via refreshStatusBar() in applyAction().
- 2026-02-27: Code review (claude-opus-4-6). Fixed 1 HIGH + 3 MEDIUM issues: (H1) splitActiveBlock() now calls refreshStatusBar() before early return — stale counts after block split; (M1) TestEditorModel_StatusBar_UpdatesOnInsert now asserts word count increase, not just chars; (M2) Added sprint-status.yaml to File List; (M3) Renamed TestEditorModel_ViewportHeight_SmallTerminal → _5to9Rows for spec consistency. 2 LOW issues noted but not fixed (L1: story spec test location mismatch, L2: len([]rune) vs runewidth — consistent with codebase). M4 (empty viewport status bar position) noted as broader viewport concern.
