# Story 4.1: Auto-Save on Typing Pause

Status: done

## Story

As a writer,
I want my work saved automatically and silently when I pause typing,
so that I never lose work and never think about saving.

## Acceptance Criteria

1. **Given** the user is editing a named file (opened with `ink file.md`) **When** the user pauses typing for 1000ms **Then** the document is saved silently to the file path without any visible indication (FR25)

2. **Given** the auto-save triggers **When** the file is written **Then** the write is atomic — content is written to a temporary file first, then renamed to the target path (NFR9)

3. **Given** the auto-save triggers but the write fails (e.g., permission denied, disk full) **When** the failure occurs **Then** an error is displayed in the status bar with `E:` prefix (e.g., `E: Cannot save: permission denied`) (NFR8) **And** the document remains open and editable — no data is lost

4. **Given** the user is typing continuously **When** each keystroke occurs **Then** the auto-save timer resets (debounce), preventing saves during active typing

5. **Given** the user is editing an unsaved buffer (no file path) **When** the auto-save timer fires **Then** no save is attempted — auto-save only operates on named files

6. **Given** the user edits a 10,000+ word document **When** auto-save triggers **Then** the save completes without perceptible interruption to typing (NFR7)

## Tasks / Subtasks

- [x] Task 1: Add `AutoSaveMsg` and `autoSaveID` to `internal/editor/editor.go` (AC: #1, #4)
  - [x] 1.1 Define `AutoSaveMsg struct { ID uint64 }` at package level (analogous to `ErrorDismissMsg`)
  - [x] 1.2 Add `autoSaveID uint64` field to `EditorModel` struct
  - [x] 1.3 Implement `startAutoSaveTimer() tea.Cmd` — increments `autoSaveID`, returns `tea.Tick(1*time.Second, ...)` sending `AutoSaveMsg{ID: id}`

- [x] Task 2: Handle `AutoSaveMsg` in `Update()` (AC: #1, #2, #3, #5)
  - [x] 2.1 Add `case AutoSaveMsg:` to the `Update()` switch
  - [x] 2.2 Reject stale timers: if `msg.ID != e.autoSaveID`, return immediately (no-op)
  - [x] 2.3 Skip save if `e.filePath == ""` (unnamed buffer — AC #5)
  - [x] 2.4 Call `file.WriteFile(e.filePath, e.serializeDocument())` for named files
  - [x] 2.5 On write failure: `return e, e.setErrorWithTimer("E: Cannot save: " + err.Error())`
  - [x] 2.6 On success: return `e, nil` — silent (no status bar update)

- [x] Task 3: Reset timer on each insert-mode keystroke in `applyAction()` (AC: #4)
  - [x] 3.1 Add `var autoSaveCmd tea.Cmd` local variable at top of `applyAction()`
  - [x] 3.2 In `case vim.InsertCharAction`: set `autoSaveCmd = e.startAutoSaveTimer()` when `e.activeBuffer != nil`
  - [x] 3.3 In `case vim.BackspaceAction`: set `autoSaveCmd = e.startAutoSaveTimer()` when `e.activeBuffer != nil`
  - [x] 3.4 In `case vim.DeleteCharAction`: set `autoSaveCmd = e.startAutoSaveTimer()` when `e.activeBuffer != nil`
  - [x] 3.5 In `case vim.InsertNewlineAction` (block-split early-return path): `return e, e.startAutoSaveTimer()`
  - [x] 3.6 In `case vim.InsertNewlineAction` (normal newline path): set `autoSaveCmd = e.startAutoSaveTimer()`
  - [x] 3.7 In `case vim.InsertTabAction`: set `autoSaveCmd = e.startAutoSaveTimer()` when `e.activeBuffer != nil`
  - [x] 3.8 Change final `return e, nil` to `return e, autoSaveCmd` (nil when no insert-mode edit occurred)

- [x] Task 4: Write tests in `internal/editor/editor_test.go` (AC: #1, #2, #3, #4, #5)
  - [x] 4.1 `TestEditor_AutoSave_NamedFile_Fires` — editor with `filePath`, enter insert mode, type char, send `AutoSaveMsg{ID: e.autoSaveID}`, verify file written with updated content
  - [x] 4.2 `TestEditor_AutoSave_UnnamedBuffer_NoSave` — editor without `filePath`, type char, send `AutoSaveMsg{ID: e.autoSaveID}`, verify no file was written
  - [x] 4.3 `TestEditor_AutoSave_StaleTimer_Ignored` — type two chars (ID increments to 2), send `AutoSaveMsg{ID: 1}`, verify no file write
  - [x] 4.4 `TestEditor_AutoSave_WriteError_ShowsError` — named file on read-only dir, send `AutoSaveMsg`, verify `E: Cannot save:` visible in status bar
  - [x] 4.5 `TestEditor_AutoSave_IDIncrementsOnEachKeystroke` — type 3 chars in insert mode, verify `e.autoSaveID == 3`
  - [x] 4.6 `TestEditor_ApplyAction_InsertChar_ReturnsAutoSaveCmd` — call `applyAction(InsertCharAction)` with active buffer, verify returned cmd is non-nil

## Dev Notes

### Architecture Overview

Story 4.1 adds a debounce auto-save timer using the `ErrorDismissMsg` pattern established in Story 3.4 as the direct template:

- `AutoSaveMsg struct { ID uint64 }` — custom Bubbletea message (mirrors `ErrorDismissMsg`)
- `autoSaveID uint64` on `EditorModel` — incremented on each insert-mode edit, invalidating stale timers
- `startAutoSaveTimer()` — analogous to `setErrorWithTimer()`, returns a `tea.Tick` cmd
- `case AutoSaveMsg:` in `Update()` — saves if ID matches and `filePath != ""`

**Why timer logic lives in `internal/editor/editor.go` (not `internal/file/autosave.go`):**

The architecture spec lists `auto-save timer` under `internal/file`, but the architecture gap note explicitly states: _"use tea.Tick (Bubbletea's built-in timer) rather than standalone goroutine timers, to stay within the framework's event model."_ Bubbletea's `tea.Tick` and message types are editor-coordination concerns. Putting them in `internal/file` would require a Bubbletea dependency in the file package — violating the leaf-package design. The actual file write uses `file.WriteFile()` (already implemented, already tested). This follows the `ErrorDismissMsg` precedent established in Story 3.4: timer lifecycle in editor, I/O in file package.

### Key Implementation Design

```go
// AutoSaveMsg is sent by a tea.Tick timer to trigger auto-save.
// The ID prevents stale timer triggers when the timer was reset by a later keystroke.
type AutoSaveMsg struct {
    ID uint64
}

// On EditorModel:
//   autoSaveID uint64  // incremented on each insert-mode keystroke

// startAutoSaveTimer resets the debounce timer and returns the timer tea.Cmd.
// Each call increments autoSaveID, invalidating any pending stale timers.
func (e *EditorModel) startAutoSaveTimer() tea.Cmd {
    e.autoSaveID++
    id := e.autoSaveID
    return tea.Tick(1*time.Second, func(_ time.Time) tea.Msg {
        return AutoSaveMsg{ID: id}
    })
}
```

**Update() handler (add to switch alongside ErrorDismissMsg):**
```go
case AutoSaveMsg:
    if msg.ID != e.autoSaveID {
        return e, nil // stale timer — a later keystroke already reset it
    }
    if e.filePath == "" {
        return e, nil // unnamed buffer — auto-save only works on named files
    }
    if err := file.WriteFile(e.filePath, e.serializeDocument()); err != nil {
        return e, e.setErrorWithTimer("E: Cannot save: " + err.Error())
    }
    return e, nil
```

**applyAction() changes (minimal — only insert-type actions):**
```go
func (e *EditorModel) applyAction(action vim.Action) (tea.Model, tea.Cmd) {
    var autoSaveCmd tea.Cmd  // set only when an insert-mode edit modifies the buffer

    switch a := action.(type) {
    // ... all non-insert cases UNCHANGED ...

    case vim.InsertCharAction:
        if e.modeHandler.Mode() == vim.Command {
            e.commandBuf += string(a.Char)
        } else if e.activeBuffer != nil {
            e.activeBuffer.Insert(a.Char)
            e.updateActiveBlockDisplay()
            autoSaveCmd = e.startAutoSaveTimer()  // NEW
        }

    case vim.BackspaceAction:
        if e.modeHandler.Mode() == vim.Command {
            runes := []rune(e.commandBuf)
            if len(runes) > 0 {
                e.commandBuf = string(runes[:len(runes)-1])
            }
        } else if e.activeBuffer != nil {
            e.activeBuffer.Backspace()
            e.updateActiveBlockDisplay()
            autoSaveCmd = e.startAutoSaveTimer()  // NEW
        }

    case vim.DeleteCharAction:
        if e.activeBuffer != nil {
            e.activeBuffer.Delete()
            e.updateActiveBlockDisplay()
            autoSaveCmd = e.startAutoSaveTimer()  // NEW
        }

    case vim.InsertNewlineAction:
        if e.activeBuffer != nil {
            content := e.activeBuffer.Content()
            cursorPos := e.activeBuffer.CursorPos()
            if cursorPos == len([]rune(content)) && strings.HasSuffix(content, "\n") {
                e.splitActiveBlock()
                e.refreshStatusBar()
                return e, e.startAutoSaveTimer()  // NEW (was: return e, nil)
            }
            e.activeBuffer.Insert('\n')
            e.updateActiveBlockDisplay()
            autoSaveCmd = e.startAutoSaveTimer()  // NEW
        }

    case vim.InsertTabAction:
        if e.activeBuffer != nil {
            e.activeBuffer.Insert('\t')
            e.updateActiveBlockDisplay()
            autoSaveCmd = e.startAutoSaveTimer()  // NEW
        }
    }

    e.refreshStatusBar()
    return e, autoSaveCmd  // CHANGED from nil
}
```

### Auto-Save Debounce Semantics

The debounce is "trailing-edge": saves 1000ms after the LAST keystroke, not the first.

```
keystroke → autoSaveID++ → tea.Tick(1s, AutoSaveMsg{ID: N})
keystroke → autoSaveID++ → tea.Tick(1s, AutoSaveMsg{ID: N+1})   ← previous timer now stale
... (1 second of silence) ...
AutoSaveMsg{ID: N}   → ignored (ID != autoSaveID which is N+1)
AutoSaveMsg{ID: N+1} → save triggered
```

No goroutine management, no channels — pure Bubbletea event model.

### Scope: Insert-Mode Actions Only

Auto-save triggers only on edits:
- `InsertCharAction` — typing a character
- `BackspaceAction` — when buffer active (not command-mode backspace)
- `DeleteCharAction`
- `InsertNewlineAction` — both normal newline and block-split paths
- `InsertTabAction`

Normal-mode actions (cursor movement, `:w`, `:wq`, mode transitions) do NOT trigger auto-save. The `:w`/`:wq` explicit saves from Story 3.4 remain unchanged. Auto-save is additive, not a replacement.

### Already Implemented (No Changes Needed)

- `file.WriteFile(path, content)` — atomic temp+rename+chmod 0644 (Story 3.4, `internal/file/file.go`)
- `serializeDocument()` — joins blocks with `\n\n` + trailing `\n` (Story 3.4, `editor.go:609`)
- `e.filePath` field — set by `NewEditor()` or `attemptSaveAs()` (Story 3.4)
- `setErrorWithTimer()` — error display with 3s auto-dismiss (Story 3.4, `editor.go:527`)
- `ErrorDismissMsg` — the exact pattern to copy for `AutoSaveMsg` (Story 3.4, `editor.go:17`)

### Testing Pattern

Tests use `package editor` (same package) for direct access to unexported fields like `autoSaveID`:

```go
func TestEditor_AutoSave_NamedFile_Fires(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "test.md")
    _ = os.WriteFile(path, []byte("hello\n"), 0644)

    blocks := []block.Block{{Type: block.Paragraph, Raw: "hello"}}
    e := NewEditor(path, blocks)
    e.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

    // Switch to insert mode and type
    e.Update(tea.KeyPressMsg{Code: 'i', String: "i"})
    e.Update(tea.KeyPressMsg{Code: 'x', Rune: 'x', String: "x"})
    savedID := e.autoSaveID  // direct field access (same package)

    // Simulate timer fire with matching ID
    e.Update(AutoSaveMsg{ID: savedID})

    content, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("file not written: %v", err)
    }
    if !strings.Contains(string(content), "x") {
        t.Errorf("auto-save did not write typed content, got: %q", string(content))
    }
}
```

For `TestEditor_AutoSave_WriteError_ShowsError`: create a temp dir, make it read-only (`os.Chmod(dir, 0444)`), set `e.filePath` to a path within it, fire `AutoSaveMsg` — verify status bar shows `E: Cannot save:`. Restore permissions in `t.Cleanup()`.

### Previous Story Intelligence (from Story 3.4)

**Directly reusable patterns:**
- `ErrorDismissMsg` → `AutoSaveMsg` (same struct shape, same ID invalidation mechanism)
- `setErrorWithTimer()` returns `tea.Cmd` — same pattern for write failure in `AutoSaveMsg` handler
- `tea.Tick(3*time.Second, ...)` → `tea.Tick(1*time.Second, ...)` for auto-save
- Tests are `package editor`, use `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup()`

**Key dev learnings from Story 3.4 completion notes:**
- `refreshStatusBar()` skips updates when `savePromptActive` — auto-save success does NOT call `refreshStatusBar()` (returns immediately)
- Error auto-dismiss uses generation IDs to prevent stale dismissals — same approach here
- All 7 packages must pass: `go test ./...`
- Commit style: `auto-save on typing pause` (all lowercase)

### Git Intelligence Summary

Recent commits (last 5):
```
63121d2 error display and save as prompt      ← Story 3.4 (direct predecessor)
78be6d7 command input via status bar          ← Story 3.3
56343b9 status bar mode aware dimming         ← Story 3.2
1332fe5 status bar with mode display and word count ← Story 3.1
3ff7b82 epic 2 retro with new golden file tests
```

**Pattern:** Every Epic 3 story touched `internal/editor/editor.go` + `internal/editor/editor_test.go`. Story 4.1 continues this pattern — same two files, no new packages.

### Project Structure Notes

**Files to modify:**
- `internal/editor/editor.go` — add `AutoSaveMsg`, `autoSaveID`, `startAutoSaveTimer()`, `case AutoSaveMsg:` in `Update()`, `var autoSaveCmd` + 5 timer assignments in `applyAction()`
- `internal/editor/editor_test.go` — add 6 tests

**Files NOT touched:**
- `internal/file/file.go` — no changes, `WriteFile()` already exists
- `internal/file/autosave.go` — NOT created (timer logic belongs in editor per Bubbletea constraint)
- All other packages — no changes

**Architecture boundary check:**
- `internal/editor` → `internal/file` — ALLOWED (editor is the root, imports everything)
- No reverse imports introduced

### References

- [Source: epics.md#Story 4.1] — User story and all acceptance criteria
- [Source: prd.md#FR25] — Auto-save on typing pause, default 1000ms
- [Source: prd.md#NFR7] — Large document performance (10,000+ words, no perceptible delay)
- [Source: prd.md#NFR8] — Error display on write failure (E: prefix)
- [Source: prd.md#NFR9] — Atomic file writes
- [Source: architecture.md#Error Handling & Recovery] — tea.Tick for timers, setErrorWithTimer pattern
- [Source: architecture.md#Package Boundary Rules] — editor → file allowed; file is a leaf (no Bubbletea)
- [Source: architecture.md#Data Flow] — Write flow: block serialize → atomic write
- [Source: architecture.md#Gap Analysis #4] — "Use tea.Tick for timer, reset on each keystroke, trigger save on tick completion"
- [Source: internal/editor/editor.go:17] — ErrorDismissMsg (template for AutoSaveMsg)
- [Source: internal/editor/editor.go:527] — setErrorWithTimer() (reused for write failures)
- [Source: internal/editor/editor.go:609] — serializeDocument() (already implemented)
- [Source: internal/file/file.go:26] — WriteFile() (atomic write, already implemented)
- [Source: 3-4-error-display-and-save-as-prompt.md] — ErrorDismissMsg precedent, tea.Tick usage, test patterns

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Implemented `AutoSaveMsg` + `autoSaveID` debounce pattern mirroring `ErrorDismissMsg` (Story 3.4 precedent)
- `startAutoSaveTimer()` increments `autoSaveID` on each call — invalidates prior stale timers
- `case AutoSaveMsg:` in `Update()` guards: stale ID check, unnamed-buffer check, then `file.WriteFile()` with error display on failure
- `applyAction()`: added `var autoSaveCmd tea.Cmd` at top; set on all 5 insert-mode actions (InsertChar, Backspace, DeleteChar, InsertNewline both paths, InsertTab); final return changed to `return e, autoSaveCmd`
- 6 tests added to `internal/editor/editor_test.go` — all pass; full `./...` suite green (7 packages)
- `tea.KeyPressMsg` struct uses `Code rune` (no `Rune`/`String` fields) — used `Code: 'x'` pattern consistent with existing test suite

### Code Review Fixes (2026-03-04)

- **CRITICAL BUG FIX**: `serializeDocument()` now reads from `activeBuffer.Content()` for the active block during insert mode — previously saved stale `blocks[i].Raw` content, dropping uncommitted edits
- Extracted `autoSaveDelay` constant (was hardcoded `1*time.Second`)
- `startAutoSaveTimer()` now uses `autoSaveDelay` constant
- Fixed `TestEditor_AutoSave_WriteError_ShowsError` to enter insert mode and type before sending `AutoSaveMsg` (was exploiting ID=0 initial state)
- Fixed `TestEditor_AutoSave_NamedFile_Fires` assertion to verify typed character appears in saved content (was only checking non-empty)
- Removed `_ = m` dead code from 3 test functions
- Added `TestEditor_ApplyAction_AllInsertActions_ReturnAutoSaveCmd` — 4 subtests covering Backspace, Delete, Newline, Tab
- Added `BenchmarkAutoSave_LargeDocument` — confirms ~83μs/op on 10K+ word doc (AC#6)

### File List

- `internal/editor/editor.go`
- `internal/editor/editor_test.go`

## Change Log

- Added `AutoSaveMsg` debounce auto-save: 1s trailing-edge timer resets on each insert-mode keystroke, writes atomically via `file.WriteFile`, displays `E: Cannot save:` on failure (Date: 2026-03-04)
- Code review: fixed critical `serializeDocument()` bug (stale content during insert mode), added constant, improved tests, added benchmark (Date: 2026-03-04)
