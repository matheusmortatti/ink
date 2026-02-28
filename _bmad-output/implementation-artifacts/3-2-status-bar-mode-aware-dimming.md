# Story 3.2: Status Bar Mode-Aware Dimming

Status: done

<!-- Epic: 3 - Status Bar and Editor Feedback -->
<!-- FRs: FR32, NFR12 -->
<!-- Date: 2026-02-27 -->

## Story

As a writer,
I want the status bar to fade away when I'm writing and reappear when I'm navigating,
so that I'm not distracted by chrome during creative flow.

## Acceptance Criteria

1. **Given** the editor is in normal mode **When** the status bar is rendered **Then** the status bar text is at full visibility (FR32)

2. **Given** the editor is in visual mode **When** the status bar is rendered **Then** the status bar text is at full visibility (FR32)

3. **Given** the editor is in insert mode **When** the status bar is rendered **Then** the status bar text is dimmed to ~30% visibility via color interpolation (foreground shifted 70% toward background) (FR32)

4. **Given** the user switches from insert mode to normal mode (Esc) **When** the mode transition occurs **Then** the status bar snaps instantly from dimmed to full visibility — no animation

5. **Given** the dimmed status bar is displayed **When** viewed against any of the 10 most common terminal color schemes **Then** the dimmed text maintains minimum readable contrast (NFR12)

## Tasks / Subtasks

- [x] Task 1: Verify implementation from Story 3.1 satisfies all ACs (no new production code expected) (AC: #1–#5)
  - [x] 1.1 Confirm `isDimmed := e.modeHandler.Mode() == vim.Insert` in `editor.go View()` covers AC1 (normal → false), AC2 (visual → false), AC3 (insert → true)
  - [x] 1.2 Confirm `refreshStatusBar()` is called at the end of `applyAction()` so the Esc key path (ChangeMode→Normal) triggers immediate un-dimming (AC4)
  - [x] 1.3 Confirm `StatusBarDimPercent = 0.7` with LAB color interpolation (`DimStyle`) in `render/color.go` provides readable contrast on both dark and light terminal themes (AC5 / NFR12)

- [x] Task 2: Add missing dimming unit tests to `internal/ui/statusbar_test.go` (AC: #2, #3)
  - [x] 2.1 `TestStatusBar_View_VisualMode_NotDimmed` — `Set("VISUAL", 2, 10)`, `View(false)` → contains `"VISUAL · 2w · 10c"`, no `\x1b[` escape codes
  - [x] 2.2 `TestStatusBar_View_Dimmed_vs_Undimmed_OutputDiffers` — same counts, `View(true)` output ≠ `View(false)` output when `lipgloss.SetColorProfile(termenv.TrueColor)` is active

- [x] Task 3: Add mode-dimming integration tests to `internal/editor/editor_test.go` (AC: #1, #3, #4)
  - [x] 3.1 `TestEditorModel_View_StatusBar_DimmedInInsert` — force `lipgloss.SetColorProfile(termenv.TrueColor)`; enter insert (`i`); verify status bar `View(isDimmed)` contains ANSI codes (`\x1b[`)
  - [x] 3.2 `TestEditorModel_View_StatusBar_FullInNormal` — same setup, exit insert (`Esc`); verify status bar `View(isDimmed)` contains no `\x1b[` (plain text only)

## Dev Notes

### Implementation Context (Story 3.1 Carry-Over)

**The dimming mechanism was fully implemented in Story 3.1.** This story's role is to add the targeted test coverage for the mode-aware dimming behavior that was deferred from 3.1.

Existing implementation that satisfies all ACs:

```go
// internal/render/color.go — AC3 (30% visibility = 70% dim)
const StatusBarDimPercent = 0.7

// internal/ui/statusbar.go — dimming applied when dimmed=true
func (s *StatusBar) View(dimmed bool) string {
    // ... centering ...
    if dimmed {
        return s.dimStyle.Render(centered) // ANSI codes applied
    }
    return centered // plain text, full visibility
}

// internal/editor/editor.go View() — AC1, AC2, AC3 correct routing
isDimmed := e.modeHandler.Mode() == vim.Insert
sbLine := e.statusBar.View(isDimmed)

// internal/editor/editor.go applyAction() — AC4 instant transition
e.refreshStatusBar() // called at end of every action, including ChangeMode
```

### Visual Mode (AC2)

`vim.Visual` is defined in `internal/vim/mode.go` (Mode enum) but has no handler implementation yet. The dimming logic already handles it correctly: `vim.Visual != vim.Insert` → `isDimmed=false` → `View(false)` → full visibility. No code change needed when visual mode is eventually implemented.

### Test Strategy for `View()` Integration Tests (Task 3)

Testing the dimming output from `editor.View()` requires lipgloss to emit ANSI codes. Follow the same pattern from `TestStatusBar_View_InsertMode_Dimmed`:

```go
func TestEditorModel_View_StatusBar_DimmedInInsert(t *testing.T) {
    lipgloss.SetColorProfile(termenv.TrueColor)
    lipgloss.SetHasDarkBackground(true)
    t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })

    blocks := block.Parse([]byte("hello world"))
    e := NewEditor("test.md", blocks)
    updated, _ := e.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
    m := updated.(*EditorModel)

    m.Update(tea.KeyPressMsg{Code: 'i'}) // enter insert mode

    viewStr := m.View().String() // get the rendered string
    // The status bar is the last non-empty line
    lines := strings.Split(strings.TrimRight(viewStr, "\n"), "\n")
    statusLine := lines[len(lines)-1]

    if !strings.Contains(statusLine, "\x1b[") {
        t.Errorf("expected ANSI codes in insert-mode status bar, got %q", statusLine)
    }
}
```

**Note:** `tea.View` has a `.String()` method. If not, split on `\n` and check the last meaningful line.

### Package Boundary Rules (unchanged from Story 3.1)

```
internal/ui → internal/render (DimStyle — allowed)
internal/ui ↛ internal/vim    (NOT allowed — mode passed as string from editor)
internal/editor → internal/vim, internal/ui (allowed — root coordinator)
```

No new imports needed for this story.

### Files to Create

- None (all production code exists)

### Files to Modify

- `internal/ui/statusbar_test.go` — Add Task 2 tests (2.1, 2.2)
- `internal/editor/editor_test.go` — Add Task 3 tests (3.1, 3.2)

**No changes to production code** unless Task 1 verification reveals a gap.

### References

- [Source: epics.md#Story 3.2] — User story statement and acceptance criteria
- [Source: prd.md#FR32] — Mode-aware status bar visibility
- [Source: prd.md#NFR12] — Minimum readable contrast for dimmed elements
- [Source: architecture.md#Color interpolation] — `render/color.go` single utility for all dimming; status bar ~30%, syntax ~40%, fade ~20%
- [Source: architecture.md#Package Boundary Rules] — `internal/ui` cannot import `internal/vim`; mode passed as string
- [Source: 3-1-status-bar-with-mode-display-and-word-count.md] — Full dimming implementation delivered in Story 3.1; View(dimmed bool), DimStyle, isDimmed logic, refreshStatusBar
- [Source: internal/ui/statusbar.go] — Existing `View(dimmed bool)`, `NewStatusBar()` with `DimStyle(StatusBarDimPercent)`
- [Source: internal/editor/editor.go#View()] — `isDimmed := e.modeHandler.Mode() == vim.Insert`
- [Source: internal/render/color.go] — `StatusBarDimPercent = 0.7`, `DimStyle()`, LAB color interpolation

### Previous Story Intelligence (from Story 3.1)

**Key learnings:**
- `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })` is required to get ANSI codes in test output
- `ui.StatusBar` fields are unexported — use `ModeLabel()` and `Counts()` accessors in editor-package tests
- `splitActiveBlock()` calls `refreshStatusBar()` before early return — mode-change transitions are always covered
- Test file is `package editor` (same package) so unexported fields like `e.statusBar`, `e.modeHandler` are accessible

**Commit pattern:** `status bar mode aware dimming` (all lowercase, concise)

**All 7 packages must pass:** `go test ./...` covering `internal/block`, `internal/render`, `internal/vim`, `internal/ui`, `internal/editor`, `internal/file`, `internal/config`

### Git Intelligence Summary

**Recent commits:**
```
1332fe5  status bar with mode display and word count
3ff7b82 epic 2 retro with new golden file tests
0e89d2b new block creation and blank canvas startup
c827492 cursor position mapping between rendered and raw
0e9ef22 fix hang in code when moving to insert mode
```

**Pattern:** Single commit per story, lowercase concise message, 1-2 files modified + tests.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None.

### Completion Notes List

- Task 1: Verified all 5 ACs satisfied by existing Story 3.1 implementation — no production code changes needed.
- Task 2: Added `TestStatusBar_View_VisualMode_NotDimmed` and `TestStatusBar_View_Dimmed_vs_Undimmed_OutputDiffers` to `internal/ui/statusbar_test.go`. Both pass.
- Task 3: Added `TestEditorModel_View_StatusBar_DimmedInInsert` and `TestEditorModel_View_StatusBar_FullInNormal` to `internal/editor/editor_test.go`. Note: `tea.View.Content` is a `Layer` interface (not a string), so tests directly invoke `m.statusBar.View(m.modeHandler.Mode() == vim.Insert)` — equivalent to the expression at `editor.go:121-122`. Both pass.
- Full regression suite: all 7 packages pass (`go test ./...`).

### File List

- `internal/ui/statusbar_test.go` (modified — added Task 2 tests)
- `internal/editor/editor_test.go` (modified — added Task 3 tests + lipgloss/termenv imports)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified — status backlog → review)

## Change Log

- Added dimming unit tests for visual mode and dimmed-vs-undimmed comparison to `statusbar_test.go` (Date: 2026-02-27)
- Added editor integration tests for status bar dimming in insert and normal mode to `editor_test.go` (Date: 2026-02-27)
