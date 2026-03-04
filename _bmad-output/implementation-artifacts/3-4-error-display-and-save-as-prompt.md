# Story 3.4: Error Display and Save-As Prompt

Status: done

<!-- Epic: 3 - Status Bar and Editor Feedback -->
<!-- FRs: FR34, FR35 -->
<!-- Date: 2026-02-28 -->

## Story

As a writer,
I want errors shown briefly in the status bar and a save prompt when quitting with unsaved work,
so that I'm informed of problems without being interrupted and can name my file when needed.

## Acceptance Criteria

1. **Given** an error occurs (file not found, permission denied, unknown command, etc.) **When** the error is triggered **Then** the status bar displays the error with `E:` prefix (e.g., `E: File not found: path.md`) (FR34)

2. **Given** an error message is displayed in the status bar **When** 3 seconds have elapsed **Then** the error auto-dismisses and the status bar returns to normal status display (FR34)

3. **Given** an error is displayed **When** the user performs any action before the 3-second timeout **Then** the error remains visible until the timeout completes — user actions do not dismiss errors early

4. **Given** the user quits (`:q` or `ZZ`) with an unsaved buffer that has content **When** the quit is initiated **Then** the status bar displays `Save as: ` with a text input cursor (FR35)

5. **Given** the save-as prompt is active **When** the user types a file path and presses `Enter` **Then** the file is saved to the specified path and ink exits

6. **Given** the save-as prompt is active **When** the user presses `Esc` **Then** the prompt is dismissed, the editor returns to normal mode, and no data is lost

7. **Given** the save-as prompt is active and the user enters an invalid path **When** `Enter` is pressed **Then** an error is displayed (e.g., `E: Invalid path: ...`) and the save prompt remains active for retry

## Tasks / Subtasks

- [x] Task 1: Add error auto-dismiss timer to StatusBar (AC: #2, #3)
  - [x] 1.1 Define `ErrorDismissMsg` struct in `internal/editor/editor.go` (custom tea message)
  - [x] 1.2 Define `errorID uint64` field on `StatusBar` to track which error the timer belongs to
  - [x] 1.3 `setErrorWithTimer()` in editor.go calls `statusBar.SetError()` and returns `tea.Tick` cmd with ErrorDismissMsg after 3 seconds
  - [x] 1.4 Increment `errorID` on each `SetError` call; embed ID in `ErrorDismissMsg` to prevent stale dismissals
  - [x] 1.5 Add `ClearError(id uint64)` method — only clears if `id` matches current `errorID` (prevents dismissing newer errors)
  - [x] 1.6 Handle `ErrorDismissMsg` in `EditorModel.Update()` — call `statusBar.ClearError(msg.ID)` then `refreshStatusBar()`

- [x] Task 2: Add file write capability to `internal/file/file.go` (AC: #5)
  - [x] 2.1 Add `WriteFile(path string, content []byte) error` — atomic write (temp file + rename)
  - [x] 2.2 Use `os.CreateTemp` in same directory as target, write content, then `os.Rename`
  - [x] 2.3 Handle errors: permission denied, disk full, invalid path
  - [x] 2.4 No `ErrPermissionDenied` sentinel needed — errors are wrapped with context

- [x] Task 3: Add `WriteFile` tests to `internal/file/file_test.go` (AC: #5)
  - [x] 3.1 `TestWriteFile_Success` — write to temp dir, verify content matches
  - [x] 3.2 `TestWriteFile_AtomicWrite` — verify temp file + rename pattern (target appears atomically)
  - [x] 3.3 `TestWriteFile_PermissionDenied` — write to read-only directory, verify error
  - [x] 3.4 `TestWriteFile_InvalidDirectory` — write to non-existent directory, verify error

- [x] Task 4: Add save-as prompt state to `internal/ui/statusbar.go` (AC: #4, #5, #6, #7)
  - [x] 4.1 Add `savePromptMode bool`, `savePromptBuf string`, `savePromptQuitAfter bool` fields
  - [x] 4.2 Add `SetSavePrompt(quitAfter bool)` method — activates save prompt mode
  - [x] 4.3 Add `SavePromptBuf() string` accessor
  - [x] 4.4 Add `InSavePrompt() bool` accessor
  - [x] 4.5 Add `SavePromptQuitAfter() bool` accessor
  - [x] 4.6 Add `AppendSavePrompt(ch rune)` method — appends character to save prompt buffer
  - [x] 4.7 Add `BackspaceSavePrompt()` method — removes last rune from save prompt buffer
  - [x] 4.8 Add `ClearSavePrompt()` method — clears save prompt state
  - [x] 4.9 Update `View(dimmed bool)` to render `Save as: {buf}` when save prompt is active (no dimming); priority: error > savePrompt > command > normal

- [x] Task 5: Wire save-as prompt in `internal/editor/editor.go` (AC: #4, #5, #6, #7)
  - [x] 5.1 Add `savePromptActive bool` field to `EditorModel`
  - [x] 5.2 Update `executeCommand()`: when `:q` with unsaved content and no filePath → activate save prompt instead of quitting
  - [x] 5.3 Update `executeCommand()`: when `:wq` with no filePath → activate save prompt with quitAfter=true
  - [x] 5.4 Update `executeCommand()`: when `:w` with no filePath → activate save prompt with quitAfter=false
  - [x] 5.5 Update `executeCommand()`: when `:w` with filePath → call `file.WriteFile()`, show success or error
  - [x] 5.6 Update `executeCommand()`: when `:wq` with filePath → call `file.WriteFile()`, then quit
  - [x] 5.7 Add save-as prompt input handling in `Update()`: when save prompt active, route keys directly via `handleSavePromptKey()`
  - [x] 5.8 Implement `attemptSaveAs()`: validate path via `file.ValidatePath()`, serialize via `serializeDocument()`, write via `file.WriteFile()`, handle errors
  - [x] 5.9 On successful save: update `e.filePath`, clear save prompt, quit if `quitAfter`
  - [x] 5.10 On save error: display error in status bar (auto-dismiss), keep save prompt active for retry
  - [x] 5.11 Update `View()` cursor: when save prompt active, cursor at end of `Save as: {buf}` in status bar row

- [x] Task 6: Implement document content check (AC: #4)
  - [x] 6.1 Add `hasContent() bool` method to `EditorModel` — returns true if any block has non-empty raw content
  - [x] 6.2 Used by quit logic: empty buffer + no filePath → quit silently (no save prompt)

- [x] Task 7: Add error auto-dismiss tests (AC: #2, #3)
  - [x] 7.1 `TestStatusBar_ErrorAutoID` — verify `errorID` increments on each `SetError()`
  - [x] 7.2 `TestStatusBar_ClearError_MatchingID` — verify error clears when ID matches
  - [x] 7.3 `TestStatusBar_ClearError_StaleID` — verify error persists when ID doesn't match (newer error)
  - [x] 7.4 `TestEditor_ErrorDismissMsg_ClearsError` — simulate `ErrorDismissMsg`, verify status bar returns to normal
  - [x] 7.5 `TestEditor_ErrorDismissMsg_StaleIgnored` — simulate stale `ErrorDismissMsg`, verify current error persists

- [x] Task 8: Add save-as prompt tests (AC: #4, #5, #6, #7)
  - [x] 8.1 `TestStatusBar_View_SavePrompt` — `SetSavePrompt()`, verify View shows `Save as: `
  - [x] 8.2 `TestStatusBar_View_SavePrompt_WithInput` — type chars into save prompt, verify View shows `Save as: filename.md`
  - [x] 8.3 `TestStatusBar_View_SavePrompt_NotDimmed` — verify save prompt ignores dimming
  - [x] 8.4 `TestEditor_Quit_UnsavedContent_ShowsSavePrompt` — blank canvas with content, `:q`, verify save prompt appears
  - [x] 8.5 `TestEditor_Quit_EmptyBuffer_QuitsDirectly` — blank canvas with no content, `:q`, verify `tea.Quit`
  - [x] 8.6 `TestEditor_Quit_NamedFile_QuitsDirectly` — editor with filePath, `:q`, verify `tea.Quit` (auto-save handles persistence later)
  - [x] 8.7 `TestEditor_SavePrompt_EscCancels` — activate save prompt, press Esc, verify returns to normal mode
  - [x] 8.8 `TestEditor_SavePrompt_TypeAndSave` — activate save prompt, type path, press Enter, verify file written and editor quits
  - [x] 8.9 `TestEditor_SavePrompt_InvalidPath_ShowsError` — type non-.md path, verify error shown and prompt remains
  - [x] 8.10 `TestEditor_Write_NamedFile` — `:w` with existing filePath, verify `file.WriteFile` called
  - [x] 8.11 `TestEditor_WriteQuit_NamedFile` — `:wq` with existing filePath, verify file written and `tea.Quit`

## Dev Notes

### Architecture Overview

This story has two distinct features sharing the status bar:

1. **Error auto-dismiss** (FR34): Adds a 3-second timer to the existing `SetError()` mechanism from Story 3.3
2. **Save-as prompt** (FR35): A new inline text input in the status bar, triggered when quitting with unsaved content

Both features follow the existing component delegation pattern: StatusBar manages display state, editor coordinates behavior.

### Error Auto-Dismiss Design

**Timer mechanism:** Use Bubbletea's `tea.Tick` command system.

```go
// internal/editor/editor.go

// ErrorDismissMsg is sent after the error timer expires
type ErrorDismissMsg struct {
    ID uint64
}

// In applyAction or wherever SetError is called:
func (e *EditorModel) setErrorWithTimer(msg string) tea.Cmd {
    e.statusBar.SetError(msg)
    id := e.statusBar.ErrorID()
    return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
        return ErrorDismissMsg{ID: id}
    })
}

// In Update():
case ErrorDismissMsg:
    e.statusBar.ClearError(msg.ID)
    e.refreshStatusBar()
```

**Key design decisions:**
- Error ID prevents stale timer from dismissing a newer error
- `SetError()` now returns (or StatusBar exposes) an ID for timer correlation
- User actions do NOT dismiss errors early — only the timer does (per AC #3)
- The `refreshStatusBar()` call after `ClearError()` restores normal display

**Important:** All existing `SetError()` call sites in `executeCommand()` must be updated to use the timer-returning variant. Currently, errors from Story 3.3 (unknown commands) never auto-dismiss — this story fixes that.

### Save-As Prompt Design

The save-as prompt reuses the status bar area (same as command mode), but with different display text and behavior.

**State machine:**
```
Normal mode → :q with unsaved content → Save-As Prompt active
Save-As Prompt → Esc → Normal mode (no data lost)
Save-As Prompt → Enter with valid path → File saved → ink exits
Save-As Prompt → Enter with invalid path → Error shown → Save-As Prompt remains
```

**StatusBar changes:**
```go
// StatusBar View() priority (highest to lowest):
// 1. errorMsg (with auto-dismiss)
// 2. savePromptMode: "Save as: {buf}"
// 3. commandMode: ":{buf}"
// 4. Normal status: "MODE · Xw · Yc"
```

Actually, errors during save-as should appear ABOVE the save prompt — but since we only have one status bar line, the error replaces the save prompt briefly (3 seconds), then the save prompt returns. This means:
- `SetError()` during save prompt should NOT clear `savePromptMode`
- After error auto-dismisses (`ClearError`), if `savePromptMode` is still true, restore save prompt display
- `refreshStatusBar()` must check: error > savePrompt > command > normal

**Editor save-as flow:**
```go
func (e *EditorModel) activateSavePrompt(quitAfter bool) {
    e.savePromptActive = true
    e.statusBar.SetSavePrompt(quitAfter)
    // No mode handler change — save prompt has its own key routing in Update()
}
```

**Key routing when save prompt is active:**
The save prompt does NOT use a vim mode handler. Instead, `Update()` intercepts keys directly when `e.savePromptActive == true`:
- Printable chars → `e.statusBar.AppendSavePrompt(ch)`
- Backspace → `e.statusBar.BackspaceSavePrompt()`
- Enter → `e.attemptSaveAs()`
- Esc → `e.cancelSavePrompt()` → return to normal mode

This approach is simpler than creating a new vim mode — the save prompt is a transient overlay, not a full mode.

### File Write Implementation

**Atomic write pattern (architecture requirement):**
```go
// internal/file/file.go

func WriteFile(path string, content []byte) error {
    dir := filepath.Dir(path)
    tmp, err := os.CreateTemp(dir, ".ink-save-*")
    if err != nil {
        return fmt.Errorf("cannot save: %w", err)
    }
    tmpPath := tmp.Name()

    _, err = tmp.Write(content)
    if closeErr := tmp.Close(); err == nil {
        err = closeErr
    }
    if err != nil {
        os.Remove(tmpPath)
        return fmt.Errorf("cannot save: %w", err)
    }

    if err := os.Rename(tmpPath, path); err != nil {
        os.Remove(tmpPath)
        return fmt.Errorf("cannot save: %w", err)
    }
    return nil
}
```

### Document Serialization

The editor needs to serialize all blocks back to markdown for saving. Check existing `block.Document` methods or `document.go` for `Serialize()` or similar. If not available, reconstruct by joining block raw content with `\n\n` separators (per architecture: "Blocks serialize back to markdown by joining raw content with `\n\n` separators").

```go
func (e *EditorModel) serializeDocument() []byte {
    var parts []string
    for _, b := range e.blocks {
        parts = append(parts, b.RawContent())
    }
    return []byte(strings.Join(parts, "\n\n"))
}
```

### Quit Logic Decision Tree

```
:q pressed
├── Has filePath (named file)
│   └── Quit immediately (auto-save will handle persistence in Epic 4)
├── No filePath + has content (unsaved buffer)
│   └── Show save-as prompt (quitAfter=true)
└── No filePath + no content (empty buffer)
    └── Quit immediately (nothing to save)

:w pressed
├── Has filePath
│   └── Write to filePath via file.WriteFile()
└── No filePath
    └── Show save-as prompt (quitAfter=false)

:wq pressed
├── Has filePath
│   └── Write to filePath via file.WriteFile(), then quit
└── No filePath
    └── Show save-as prompt (quitAfter=true)
```

### Cmd Return Pattern

Currently, `executeCommand()` returns `(tea.Model, tea.Cmd)`. The timer cmd from `setErrorWithTimer()` must be returned through the Bubbletea pipeline. When using `tea.Tick`, the cmd is returned from `Update()` and Bubbletea dispatches the `ErrorDismissMsg` when the timer fires.

**Critical:** All `applyAction` paths that call `setErrorWithTimer()` must propagate the returned `tea.Cmd` back through `Update()`. Check how `executeCommand()` currently returns commands — it already returns `(tea.Model, tea.Cmd)`, so this should work naturally.

### View() Cursor for Save-As Prompt

```go
case e.savePromptActive:
    text := "Save as: " + e.statusBar.SavePromptBuf()
    padLeft := (e.width - len([]rune(text))) / 2
    if padLeft < 0 { padLeft = 0 }
    cursorCol := padLeft + len([]rune(text))
    sbRow := e.height - 1
    v.Cursor = tea.NewCursor(cursorCol, sbRow)
    v.Cursor.Shape = tea.CursorBar
```

### Previous Story Intelligence (from Story 3.3)

**Key learnings:**
- `StatusBar` fields are unexported — use accessor methods in cross-package tests
- `refreshStatusBar()` is called at end of most `applyAction` branches — must NOT overwrite error state
- `executeCommand()` uses early return pattern before `refreshStatusBar()` for error paths
- `lipgloss.SetColorProfile(termenv.TrueColor)` + cleanup for ANSI tests
- Test file uses `package editor` (same package) — direct access to unexported fields
- All 7 packages must pass: `go test ./...`

**Commit pattern:** `error display and save-as prompt` (all lowercase, concise)

### Git Intelligence Summary

**Recent commits:**
```
78be6d7 command input via status bar
56343b9 status bar mode aware dimming
1332fe5  status bar with mode display and word count
3ff7b82 epic 2 retro with new golden file tests
0e89d2b new block creation and blank canvas startup
```

**Key insight:** Stories 3.1-3.3 all touched `internal/ui/statusbar.go` and `internal/editor/editor.go`. This story continues that pattern and adds `internal/file/file.go` modifications.

### Project Structure Notes

- Alignment with architecture: `internal/file/file.go` for atomic writes, `internal/ui/statusbar.go` for display, `internal/editor/editor.go` for coordination
- No new packages needed — all work fits within existing package boundaries
- `internal/ui/saveprompt.go` from architecture is NOT created — save-as prompt is simple enough to live in `statusbar.go` as additional state (avoids over-engineering)

### References

- [Source: epics.md#Story 3.4] — User story statement and acceptance criteria
- [Source: prd.md#FR34] — Error messages in status bar with E: prefix, 3-second auto-dismiss
- [Source: prd.md#FR35] — Save-as prompt in status bar for unsaved buffer content
- [Source: ux-design-specification.md#Error Handling Patterns] — Error display rules, 3-second timer, E: prefix
- [Source: ux-design-specification.md#Save-as prompt component] — Save prompt UX behavior
- [Source: ux-design-specification.md#Flow Optimization Principles] — "One prompt, ever" — save-as is the only dialog
- [Source: architecture.md#Error Handling & Recovery] — E: prefix, 3-second auto-dismiss, showError pattern
- [Source: architecture.md#Rendering Pipeline & Caching] — tea.Tick for timers (architecture gap #4)
- [Source: architecture.md#Package Boundary Rules] — editor → vim, ui, file allowed
- [Source: architecture.md#Data Flow] — Write flow: block serialize → atomic write
- [Source: internal/ui/statusbar.go] — Current SetError, View(), command/error state
- [Source: internal/editor/editor.go] — executeCommand(), applyAction, refreshStatusBar, View cursor
- [Source: internal/file/file.go] — ReadFile, ValidatePath; WriteFile to be added
- [Source: internal/vim/action.go] — 13 existing action types
- [Source: 3-3-command-input-via-status-bar.md] — Previous story with command mode, error display foundation

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

No debug issues encountered. Implementation followed the Dev Notes design exactly.

### Completion Notes List

- Implemented `ErrorDismissMsg` in `editor.go` and `setErrorWithTimer()` helper that wraps `statusBar.SetError()` + returns a 3-second `tea.Tick` cmd. The timer ID prevents stale dismissals.
- `StatusBar.Set()` and `SetCommand()` no longer clear `errorMsg` — only the timer (via `ClearError(id)`) can dismiss errors, satisfying AC #3.
- `StatusBar.View()` priority reordered: error > savePrompt > command > normal, so errors always surface even during command input.
- `WriteFile()` uses atomic temp-file + rename pattern (same directory as target, `.ink-save-*` prefix).
- Save-as prompt is a transient overlay in the status bar — not a vim mode. `handleSavePromptKey()` intercepts keys directly in `Update()` when `savePromptActive`.
- `refreshStatusBar()` skips updates when save prompt is active to avoid overwriting prompt display.
- After successful save: `e.filePath` updated, save prompt cleared, quit if `quitAfter`.
- 16 new tests added across `internal/file`, `internal/ui`, and `internal/editor`. All 7 packages pass with no regressions.

### File List

- `internal/editor/editor.go`
- `internal/editor/editor_test.go`
- `internal/file/file.go`
- `internal/file/file_test.go`
- `internal/ui/statusbar.go`
- `internal/ui/statusbar_test.go`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/3-4-error-display-and-save-as-prompt.md`

### Change Log

- Added error auto-dismiss timer (3s `tea.Tick`) using generation IDs to prevent stale dismissals (AC #2, #3)
- Added atomic `WriteFile()` to `internal/file` package (AC #5)
- Added save-as prompt in status bar for unnamed buffers: `:q`/`:w`/`:wq` trigger prompt when no filePath (AC #4, #5, #6, #7)
- Updated `StatusBar.View()` priority: error > savePrompt > command > normal (Date: 2026-02-28)

**Senior Developer Review fixes (Date: 2026-03-04):**
- [H1] Implemented `ZZ` two-key quit sequence in `NormalHandler` — AC #4 was partial without it; also wired same save-prompt logic as `:q` for unnamed buffers
- [H2] Fixed `WriteFile()` to `os.Chmod(tmpPath, 0644)` before rename — `os.CreateTemp` uses 0600 which would lock files to owner-only read
- [M1] Added `StatusBar.HasError()` and guarded printable input in save prompt: typing is ignored during error overlay to prevent invisible buffer mutation
- [M2] Added `TestEditor_SavePrompt_ErrorThenRestore` — verifies save prompt reappears after `ErrorDismissMsg` fires following a failed save attempt
- [M3] `serializeDocument()` now appends trailing `\n` (POSIX convention, preserves round-trip fidelity for files that originally had a trailing newline)
- Added `TestNormalHandler_ZZ_Quit`, `TestNormalHandler_ZNonZ_FallsThrough`, `TestEditor_ZZ_NamedFile_QuitsDirectly`, `TestEditor_ZZ_UnsavedContent_ShowsSavePrompt`, `TestWriteFile_Permissions`
