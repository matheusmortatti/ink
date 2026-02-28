# Story 3.3: Command Input via Status Bar

Status: done

<!-- Epic: 3 - Status Bar and Editor Feedback -->
<!-- FRs: FR33 -->
<!-- Date: 2026-02-28 -->

## Story

As a writer,
I want to type commands like `:q` and `:w` in the status bar,
so that I can control the editor using familiar vim command patterns.

## Acceptance Criteria

1. **Given** the editor is in normal mode **When** the user presses `:` **Then** the status bar content is replaced with `:` followed by a cursor, and command mode is activated (FR33)

2. **Given** the editor is in command mode **When** the user types characters **Then** the characters are appended to the command string after the `:` prefix, updating the status bar display live

3. **Given** the editor is in command mode **When** the user presses Backspace **Then** the last character of the command string is deleted and the status bar updates

4. **Given** the editor is in command mode **When** the user presses Enter **Then** the command is executed and the status bar returns to normal status display

5. **Given** the editor is in command mode **When** the user presses Esc **Then** the command is cancelled, input is discarded, and the editor returns to normal mode with the standard status bar display

6. **Given** the user enters `:q` **When** Enter is pressed **Then** the editor quits

7. **Given** the user enters an unrecognized command (e.g., `:xyz`) **When** Enter is pressed **Then** the status bar displays `E: Not a command: xyz` and the editor returns to normal mode

## Tasks / Subtasks

- [x] Task 1: Add `ExecuteCommandAction` to `internal/vim/action.go` (AC: #4, #6, #7)
  - [x] 1.1 Add `ExecuteCommandAction{Command string}` struct with `actionTag()` method

- [x] Task 2: Create `internal/vim/command.go` — CommandHandler (AC: #2, #3, #4, #5)
  - [x] 2.1 Define `CommandHandler` struct with `buf string` field
  - [x] 2.2 Implement `NewCommandHandler() *CommandHandler`
  - [x] 2.3 Implement `HandleKey(key string) Action`:
    - `"esc"` → `ChangeModeAction{Mode: Normal}`
    - `"backspace"` → `BackspaceAction{}`
    - `"enter"` → `ExecuteCommandAction{Command: h.buf}`
    - single printable rune (`runes[0] >= 32`) → `InsertCharAction{Char: rune}`
    - all other keys → `NoOpAction{}`
  - [x] 2.4 Implement `Mode() Mode` returning `Command`

- [x] Task 3: Update `internal/vim/normal.go` to enter command mode (AC: #1)
  - [x] 3.1 Add `case ":"` in `handleSingleKey` → `return ChangeModeAction{Mode: Command}`

- [x] Task 4: Update `internal/ui/statusbar.go` to support command mode display (AC: #1, #2, #3, #5, #7)
  - [x] 4.1 Add `commandMode bool`, `commandBuf string`, `errorMsg string` fields to `StatusBar`
  - [x] 4.2 Add `SetCommand(buf string)` method — sets `commandMode=true`, `commandBuf=buf`, `errorMsg=""`
  - [x] 4.3 Add `SetError(msg string)` method — sets `commandMode=false`, `errorMsg=msg`
  - [x] 4.4 Update `Set(...)` to also clear `commandMode`, `commandBuf`, `errorMsg`
  - [x] 4.5 Update `View(dimmed bool)` to render differently per state:
    - `commandMode` → render `:commandBuf` centered, **no dimming** (command mode is never dimmed)
    - `errorMsg != ""` → render `errorMsg` centered, no dimming
    - otherwise → existing `modeLabel · Xw · Yc` format

- [x] Task 5: Update `internal/editor/editor.go` to wire command mode (AC: #1–#7)
  - [x] 5.1 Add `commandBuf string` field to `EditorModel`
  - [x] 5.2 In `applyAction` `ChangeModeAction` branch:
    - Add case: `a.Mode == vim.Command && e.modeHandler.Mode() == vim.Normal` → `e.enterCommandMode()`
    - Add case: `a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Command` → `e.exitCommandMode()`
  - [x] 5.3 In `applyAction` `InsertCharAction` branch: add command-mode context — when in command mode, append `a.Char` to `e.commandBuf` instead of calling gap buffer
  - [x] 5.4 In `applyAction` `BackspaceAction` branch: add command-mode context — when in command mode, remove last rune from `e.commandBuf`
  - [x] 5.5 Add `applyAction` case for `vim.ExecuteCommandAction`:
    - Call `e.executeCommand(a.Command)`
  - [x] 5.6 Implement `enterCommandMode()`:
    - Set `e.commandBuf = ""`
    - Set `e.modeHandler = vim.NewCommandHandler()`
  - [x] 5.7 Implement `exitCommandMode()`:
    - Set `e.commandBuf = ""`
    - Set `e.modeHandler = vim.NewNormalHandler()`
  - [x] 5.8 Implement `executeCommand(cmd string)`:
    - `"q"` → return `e, tea.Quit` (handle via returning from applyAction)
    - `"wq"` → (deferred to Story 3.4/Epic 4) for now treat like `:q`
    - `"w"`, `"w <path>"` → (deferred to Epic 4) show error "E: File save not yet implemented"
    - anything else → `e.statusBar.SetError("E: Not a command: " + cmd)`, `e.exitCommandMode()`
  - [x] 5.9 Update `refreshStatusBar()`: if `e.modeHandler.Mode() == vim.Command`, call `e.statusBar.SetCommand(e.commandBuf)`; else existing logic
  - [x] 5.10 Update `View()` cursor positioning for command mode: cursor appears at end of `:commandBuf` text in the status bar row

- [x] Task 6: Create `internal/vim/command_test.go` (AC: #2–#5)
  - [x] 6.1 `TestCommandHandler_PrintableChar` — verify printable char → `InsertCharAction{Char: rune}`
  - [x] 6.2 `TestCommandHandler_Backspace` — verify backspace → `BackspaceAction{}`
  - [x] 6.3 `TestCommandHandler_Enter` — type `"quit"` sequence then enter → `ExecuteCommandAction{Command: ""}` (handler doesn't accumulate — editor does; OR test if handler accumulates)
  - [x] 6.4 `TestCommandHandler_Esc` — verify esc → `ChangeModeAction{Mode: Normal}`
  - [x] 6.5 `TestCommandHandler_Mode` — verify `Mode()` returns `vim.Command`
  - [x] 6.6 `TestCommandHandler_NonPrintableIgnored` — arrow keys, ctrl+x etc → `NoOpAction{}`

- [x] Task 7: Add StatusBar command mode tests to `internal/ui/statusbar_test.go` (AC: #1, #2, #7)
  - [x] 7.1 `TestStatusBar_View_CommandMode_ShowsColonBuf` — `SetCommand("q")`, `View(false)` → contains `:q`
  - [x] 7.2 `TestStatusBar_View_CommandMode_EmptyBuf` — `SetCommand("")`, `View(false)` → contains `:` only
  - [x] 7.3 `TestStatusBar_View_ErrorMsg` — `SetError("E: Not a command: xyz")`, `View(false)` → contains `E: Not a command: xyz`
  - [x] 7.4 `TestStatusBar_View_CommandMode_NotDimmed` — `SetCommand("q")`, `View(true)` → same output as `View(false)` (no dimming in command mode)
  - [x] 7.5 `TestStatusBar_Set_ClearsCommandMode` — call `SetCommand("q")` then `Set("NORMAL", 0, 0)` → `View(false)` shows normal format, not `:q`

- [x] Task 8: Add editor integration tests to `internal/editor/editor_test.go` (AC: #1, #4, #5, #6, #7)
  - [x] 8.1 `TestEditor_Colon_EntersCommandMode` — press `:`, verify `CurrentMode() == vim.Command`
  - [x] 8.2 `TestEditor_CommandMode_EscReturnsNormal` — press `:`, press `Esc`, verify `CurrentMode() == vim.Normal`, status bar shows normal format
  - [x] 8.3 `TestEditor_CommandMode_TypeAndEnter_Quit` — press `:`, type `q`, press `Enter`, verify `tea.Quit` cmd returned
  - [x] 8.4 `TestEditor_CommandMode_UnknownCommand_ShowsError` — press `:`, type `xyz`, press `Enter`, verify status bar contains `E: Not a command: xyz` and mode returns to Normal
  - [x] 8.5 `TestEditor_CommandMode_StatusBarShowsInput` — press `:`, type `q`, verify status bar view contains `:q`

## Dev Notes

### Architecture Overview

Command mode is the **4th vim mode** in the existing `vim.Mode` enum (`Normal=0, Insert=1, Visual=2, Command=3`). The `Command` constant and its `String()` method (`"COMMAND"`) are already defined in `internal/vim/mode.go:6-11`. No changes needed to `mode.go`.

The architecture follows the **existing per-mode handler pattern** used by `NormalHandler` and `InsertHandler`. The editor owns mutable state (cursor, blocks, command buffer), while handlers are stateless routers that return `Action` values.

### CommandHandler Design

`CommandHandler` is stateless — it does **not** accumulate the command buffer. The editor owns `commandBuf string`. This mirrors how `InsertHandler` doesn't own the gap buffer — the editor does. Handler just routes keys to actions:

```go
// internal/vim/command.go
package vim

type CommandHandler struct{}

func NewCommandHandler() *CommandHandler { return &CommandHandler{} }

func (h *CommandHandler) HandleKey(key string) Action {
    switch key {
    case "esc":
        return ChangeModeAction{Mode: Normal}
    case "backspace":
        return BackspaceAction{}
    case "enter":
        return ExecuteCommandAction{}  // editor reads e.commandBuf
    default:
        runes := []rune(key)
        if len(runes) == 1 && runes[0] >= 32 {
            return InsertCharAction{Char: runes[0]}
        }
        return NoOpAction{}
    }
}

func (h *CommandHandler) Mode() Mode { return Command }
```

Note: `ExecuteCommandAction` does not need a `Command string` field if the editor reads `e.commandBuf` directly. Alternatively, include the command in the action for clarity. Either works — choose whichever is cleaner.

### Action.go Additions

Add to `internal/vim/action.go`:
```go
// ExecuteCommandAction signals the editor to execute the accumulated command buffer.
type ExecuteCommandAction struct{}

func (ExecuteCommandAction) actionTag() {}
```

Or with the command embedded:
```go
type ExecuteCommandAction struct {
    Command string
}
```

The stateless approach (no Command field; editor reads `e.commandBuf`) is consistent with how `BackspaceAction` doesn't carry the deleted character.

### StatusBar Changes

StatusBar needs to render three distinct states:
1. **Normal/Insert/Visual**: `NORMAL · 5w · 20c` (existing)
2. **Command input**: `:q` (command mode active, showing typed text)
3. **Error**: `E: Not a command: xyz` (brief error display after bad command)

```go
// StatusBar additions
type StatusBar struct {
    // existing fields...
    commandMode bool
    commandBuf  string
    errorMsg    string
}

func (s *StatusBar) SetCommand(buf string) {
    s.commandMode = true
    s.commandBuf  = buf
    s.errorMsg    = ""
}

func (s *StatusBar) SetError(msg string) {
    s.commandMode = false
    s.commandBuf  = ""
    s.errorMsg    = msg
}

// Set() clears command/error state
func (s *StatusBar) Set(modeLabel string, words, chars int) {
    s.commandMode = false
    s.commandBuf  = ""
    s.errorMsg    = ""
    s.modeLabel   = modeLabel
    s.wordCount   = words
    s.charCount   = chars
}
```

`View(dimmed bool)` updated:
```go
func (s *StatusBar) View(dimmed bool) string {
    var plain string
    switch {
    case s.commandMode:
        plain = ":" + s.commandBuf
        dimmed = false  // command mode is never dimmed
    case s.errorMsg != "":
        plain = s.errorMsg
        dimmed = false
    default:
        plain = fmt.Sprintf("%s · %dw · %dc", s.modeLabel, s.wordCount, s.charCount)
    }
    // centering unchanged...
}
```

### Editor Changes: applyAction

The `InsertCharAction` and `BackspaceAction` branches in `applyAction` must be mode-aware:

```go
case vim.InsertCharAction:
    if e.modeHandler.Mode() == vim.Command {
        e.commandBuf += string(a.Char)
    } else if e.activeBuffer != nil {
        e.activeBuffer.Insert(a.Char)
        e.updateActiveBlockDisplay()
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
    }

case vim.ExecuteCommandAction:
    return e.executeCommand()
```

### Editor Changes: executeCommand()

```go
func (e *EditorModel) executeCommand() (tea.Model, tea.Cmd) {
    cmd := e.commandBuf
    e.exitCommandMode()
    switch cmd {
    case "q":
        return e, tea.Quit
    case "wq":
        // File save deferred to Epic 4 — quit without save for now
        return e, tea.Quit
    case "w":
        // File save deferred to Epic 4
        if e.statusBar != nil {
            e.statusBar.SetError("E: File save not yet implemented")
        }
    default:
        if e.statusBar != nil {
            e.statusBar.SetError("E: Not a command: " + cmd)
        }
    }
    e.refreshStatusBar()
    return e, nil
}
```

**Important**: `exitCommandMode()` must be called before `SetError()` so the mode transitions back to Normal before the error is displayed. Then `refreshStatusBar()` is NOT called (it would overwrite the error) — or handle this carefully.

Actually, simpler: call `exitCommandMode()` first (switches handler + clears commandBuf), then set the error:
```go
func (e *EditorModel) exitCommandMode() {
    e.commandBuf = ""
    e.modeHandler = vim.NewNormalHandler()
}
```

Then after `exitCommandMode()`, manually call `statusBar.SetError(...)` without going through `refreshStatusBar()`.

### Editor Changes: refreshStatusBar()

```go
func (e *EditorModel) refreshStatusBar() {
    if e.statusBar == nil {
        return
    }
    if e.modeHandler.Mode() == vim.Command {
        e.statusBar.SetCommand(e.commandBuf)
        return
    }
    words, chars := e.computeDocumentCounts()
    e.statusBar.Set(e.modeHandler.Mode().String(), words, chars)
}
```

### Editor Changes: View() cursor for command mode

In `View()`, add a third cursor case for command mode:

```go
case e.modeHandler.Mode() == vim.Command:
    // Cursor appears right after ":commandBuf" in the centered status bar
    text := ":" + e.commandBuf
    padLeft := (e.width - len([]rune(text))) / 2
    if padLeft < 0 { padLeft = 0 }
    cursorCol := padLeft + len([]rune(text))
    sbRow := e.height - 1  // status bar is always the last row
    v.Cursor = tea.NewCursor(cursorCol, sbRow)
    v.Cursor.Shape = tea.CursorBar
```

Note: `e.height - 1` is the last terminal row (0-indexed). Verify against the View() content assembly to ensure alignment. When `sbRows == 2`, the status bar content is at `height - 1` and the blank separator is at `height - 2`.

### ChangeModeAction in applyAction

Update the existing `ChangeModeAction` branch to handle all mode transitions:

```go
case vim.ChangeModeAction:
    switch {
    case a.Mode == vim.Insert && e.modeHandler.Mode() == vim.Normal:
        e.enterInsertMode(a.Variant)
    case a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Insert:
        e.exitInsertMode()
    case a.Mode == vim.Command && e.modeHandler.Mode() == vim.Normal:
        e.enterCommandMode()
    case a.Mode == vim.Normal && e.modeHandler.Mode() == vim.Command:
        e.exitCommandMode()
    }
```

### Package Boundary Rules (unchanged)

```
internal/ui → internal/render  (allowed — dimming)
internal/ui ↛ internal/vim     (NOT allowed — mode passed as string from editor)
internal/editor → internal/vim, internal/ui  (allowed — root coordinator)
internal/vim ↛ internal/ui     (NOT allowed — no horizontal imports)
```

CommandHandler lives in `internal/vim` — consistent with Normal/Insert/Visual handlers.

### Files to Create

- `internal/vim/command.go`
- `internal/vim/command_test.go`

### Files to Modify

- `internal/vim/action.go` — add `ExecuteCommandAction`
- `internal/vim/normal.go` — add `:` key → `ChangeModeAction{Mode: Command}`
- `internal/ui/statusbar.go` — add command/error display state
- `internal/editor/editor.go` — command mode wiring

### Testing Approach

Follow the existing test patterns:
- `internal/vim/command_test.go` in `package vim` — unit tests for handler key routing
- `internal/ui/statusbar_test.go` additions in `package ui`
- `internal/editor/editor_test.go` additions in `package editor`

For editor integration tests, follow the pattern from Story 3.2:
```go
blocks := block.Parse([]byte("hello world"))
e := NewEditor("test.md", blocks)
updated, _ := e.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
m := updated.(*EditorModel)
m.Update(tea.KeyPressMsg{Code: ':'})  // enter command mode
```

**Note on tea.KeyPressMsg**: The key string for `:` is `":"`. Verify via `msg.String()` — it should be `":"` as a single-character string.

### References

- [Source: epics.md#Story 3.3] — User story statement and acceptance criteria
- [Source: prd.md#FR33] — Command mode via status bar
- [Source: prd.md#FR26] — `:q` quit command
- [Source: architecture.md#Vim Mode Architecture] — Per-mode handler pattern, ModeHandler interface
- [Source: architecture.md#Package Boundary Rules] — import constraints
- [Source: architecture.md#Project Structure] — `vim/command.go`, `vim/command_test.go` expected file paths
- [Source: internal/vim/mode.go] — `Command` mode already defined (value 3)
- [Source: internal/vim/action.go] — Existing action types; add `ExecuteCommandAction`
- [Source: internal/vim/insert.go] — Reference implementation for new `CommandHandler`
- [Source: internal/vim/normal.go] — Add `:` → `ChangeModeAction{Mode: Command}`
- [Source: internal/ui/statusbar.go] — Existing `StatusBar`, extend with `SetCommand`/`SetError`
- [Source: internal/editor/editor.go] — `applyAction`, `refreshStatusBar`, `View`, mode transition helpers
- [Source: 3-2-status-bar-mode-aware-dimming.md] — Test patterns: lipgloss setup, `package editor`, accessor patterns
- [Source: 3-1-status-bar-with-mode-display-and-word-count.md] — StatusBar design and `refreshStatusBar()` origin

### Previous Story Intelligence (from Story 3.2)

**Key learnings:**
- `lipgloss.SetColorProfile(termenv.TrueColor)` + `t.Cleanup(func() { lipgloss.SetColorProfile(termenv.Ascii) })` for ANSI tests
- `ui.StatusBar` fields unexported — use accessor methods (`ModeLabel()`, `Counts()`) in editor-package tests; or add `CommandBuf()` / `InCommandMode()` accessors if needed
- Test file is `package editor` (same package) — unexported fields like `e.commandBuf`, `e.modeHandler` directly accessible
- `refreshStatusBar()` called at end of every `applyAction()` — must NOT be called after `SetError()` as it would overwrite the error
- **All 7 packages must pass:** `go test ./...` covering `internal/block`, `internal/render`, `internal/vim`, `internal/ui`, `internal/editor`, `internal/file`, `internal/config`

**Commit pattern:** `command input via status bar` (all lowercase, concise)

### Git Intelligence Summary

**Recent commits:**
```
56343b9 status bar mode aware dimming
1332fe5  status bar with mode display and word count
3ff7b82 epic 2 retro with new golden file tests
0e89d2b new block creation and blank canvas startup
c827492 cursor position mapping between rendered and raw
```

**Pattern:** Single commit per story, lowercase concise message.

**Key insight from recent commits:** Stories 3.1 and 3.2 touched `internal/ui/statusbar.go` and `internal/editor/editor.go`. This story also touches those same files plus adds `internal/vim/command.go`. No block or render changes expected.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None. Implementation was straightforward following the Dev Notes architecture exactly.

### Completion Notes List

- Implemented `ExecuteCommandAction` as a zero-field struct (stateless — editor reads `e.commandBuf` directly, consistent with `BackspaceAction` pattern)
- `CommandHandler` is stateless, mirrors `InsertHandler` design exactly
- `refreshStatusBar()` gates on `vim.Command` mode before calling `SetCommand()`, preventing `Set()` from overwriting error messages set by `executeCommand()`
- `executeCommand()` calls `exitCommandMode()` first (switches handler + clears buf), then `SetError()` — so `refreshStatusBar()` is NOT called afterward for error cases (returned early from `applyAction` via direct return)
- Cursor for command mode is placed after `":commandBuf"` in the last terminal row using the same centering math as `StatusBar.View()`
- All 7 packages pass: `go test ./...` clean

### File List

- `internal/vim/action.go` — added `ExecuteCommandAction` struct and `actionTag()`
- `internal/vim/command.go` — new file: `CommandHandler` implementation
- `internal/vim/command_test.go` — new file: unit tests for `CommandHandler`
- `internal/vim/normal.go` — added `case ":"` → `ChangeModeAction{Mode: Command}`
- `internal/ui/statusbar.go` — added `commandMode`, `commandBuf`, `errorMsg` fields; `SetCommand()`, `SetError()` methods; updated `Set()` and `View()`
- `internal/ui/statusbar_test.go` — added 5 command mode tests
- `internal/editor/editor.go` — added `commandBuf` field; updated `applyAction` with command mode branches; added `enterCommandMode()`, `exitCommandMode()`, `executeCommand()`; updated `refreshStatusBar()` and `View()`
- `internal/editor/editor_test.go` — added 7 integration tests (5 original + 2 review fixes)

## Change Log

- 2026-02-28: Implemented command input via status bar (Story 3.3) — added CommandHandler, ExecuteCommandAction, command mode wiring in editor, StatusBar command/error display, and comprehensive test coverage across vim, ui, and editor packages.
- 2026-02-28: Code review fixes (AI) — H1: empty command no longer shows broken error message; M2: command mode cursor guarded for hidden status bar (height<5); M3: added code comment documenting early-return control flow in executeCommand(); added 2 new tests (empty command, backspace on empty buffer).
