# Story 5.3: Clipboard Integration

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a writer,
I want to yank and paste text using vim commands and the system clipboard,
so that I can move text freely in and out of ink.

## Acceptance Criteria

1. **Given** the user has selected text in visual mode **When** the user presses `y` **Then** the selected text is copied to the system clipboard (FR22)

2. **Given** text has been yanked to the clipboard **When** the user presses `p` in normal mode **Then** the clipboard content is pasted after the cursor position (FR22)

3. **Given** text has been yanked to the clipboard **When** the user presses `P` in normal mode **Then** the clipboard content is pasted before the cursor position

4. **Given** text has been copied from an external application to the system clipboard **When** the user presses `p` in normal mode within ink **Then** the external clipboard content is pasted into the document

5. **Given** text is yanked within ink **When** the user switches to another application and pastes **Then** the yanked text is available in the system clipboard

6. **Given** the user presses `dd` in normal mode **When** the current line is deleted **Then** the deleted line content is placed in the system clipboard

## Tasks / Subtasks

- [ ] Task 1: Add clipboard abstraction (`internal/editor/clipboard.go`) (AC: #1-#5)
  - [ ] 1.1 Create `Clipboard` interface with `Read() (string, error)` and `Write(text string) error` methods
  - [ ] 1.2 Implement `execClipboard` struct using `os/exec` to call platform-specific commands: `pbcopy`/`pbpaste` (macOS), `xclip -selection clipboard` / `xclip -selection clipboard -o` (Linux X11), `wl-copy`/`wl-paste` (Linux Wayland), `clip.exe`/`powershell.exe Get-Clipboard` (Windows)
  - [ ] 1.3 Implement `detectPlatform() string` to identify the clipboard backend: check `WAYLAND_DISPLAY` env for Wayland, `DISPLAY` env for X11, `runtime.GOOS` for macOS/Windows
  - [ ] 1.4 Implement `NewClipboard() Clipboard` factory that returns the correct implementation (or a no-op clipboard that returns `ErrClipboardUnavailable` if no tools found)
  - [ ] 1.5 Add `clipboard Clipboard` field to `EditorModel` and initialize in `NewEditor()`

- [ ] Task 2: Add new action types (`internal/vim/action.go`) (AC: #2, #3, #6)
  - [ ] 2.1 Add `PasteAction struct { Before bool }` with `actionTag()` -- `Before=false` for `p`, `Before=true` for `P`
  - [ ] 2.2 Add `YankAction struct {}` with `actionTag()` -- used by visual mode (5.4) to signal clipboard write
  - [ ] 2.3 Add `DeleteLineAction struct {}` with `actionTag()` -- for `dd` in normal mode

- [ ] Task 3: Add `dd` and `p`/`P` key bindings in normal mode (`internal/vim/normal.go`) (AC: #2, #3, #6)
  - [ ] 3.1 Add `d` as a pending operator: first `d` sets `n.pending = 'd'`, second `d` emits `DeleteLineAction{}`
  - [ ] 3.2 Add `p` key: emit `PasteAction{Before: false}`
  - [ ] 3.3 Add `P` / `shift+P` key: emit `PasteAction{Before: true}`
  - [ ] 3.4 Add `y` key handling for future visual mode use (currently NoOp in normal mode without selection)

- [ ] Task 4: Handle new actions in editor (`internal/editor/editor.go`) (AC: #2-#6)
  - [ ] 4.1 Handle `PasteAction` in `applyAction()`:
    - If `blockPendingCommit`: re-enter insert mode in the block, paste at cursor, return to normal
    - If no active block: enter insert mode on current block, paste clipboard text at cursor position, exit back to normal
    - Use `e.clipboard.Read()` to get text; on error show `E: Clipboard unavailable` via status bar
    - `Before=false` (p): paste after cursor; `Before=true` (P): paste before cursor
    - Record undo before pasting, trigger auto-save after
  - [ ] 4.2 Handle `DeleteLineAction` in `applyAction()`:
    - Must have active block (blockPendingCommit or insert mode)
    - If `blockPendingCommit`: use existing activeBuffer
    - Extract current line content from gap buffer
    - Delete the line (from line start to next `\n` or end of content)
    - Write deleted text to clipboard via `e.clipboard.Write()`; on error show `E: Clipboard unavailable`
    - Record undo before deletion, trigger auto-save after
    - If block becomes empty after `dd`, commit/remove the block
  - [ ] 4.3 Handle `YankAction` in `applyAction()`: placeholder for visual mode -- write selected text to clipboard

- [ ] Task 5: Add gap buffer helper for line operations (`internal/block/gapbuffer.go`) (AC: #6)
  - [ ] 5.1 Add `CurrentLineContent() string` -- returns the text of the current cursor line (without trailing `\n`)
  - [ ] 5.2 Add `DeleteCurrentLine()` -- deletes from line start to next `\n` (inclusive) or end of content; positions cursor at start of next line (or end of previous line if last line deleted)

- [ ] Task 6: Write unit tests for clipboard (`internal/editor/clipboard_test.go`) (AC: #1-#5)
  - [ ] 6.1 `TestClipboard_DetectPlatform` -- verify platform detection based on env/OS
  - [ ] 6.2 Create `mockClipboard` struct implementing `Clipboard` interface for testing
  - [ ] 6.3 `TestClipboard_NoopFallback` -- verify graceful degradation when no clipboard tools available

- [ ] Task 7: Write unit tests for normal mode key bindings (`internal/vim/normal_test.go`) (AC: #2, #3, #6)
  - [ ] 7.1 `TestNormal_p_ReturnsPasteAction`
  - [ ] 7.2 `TestNormal_P_ReturnsPasteActionBefore`
  - [ ] 7.3 `TestNormal_dd_ReturnsDeleteLineAction`
  - [ ] 7.4 `TestNormal_d_ThenOtherKey_CancelsPending`

- [ ] Task 8: Write unit tests for gap buffer line operations (`internal/block/gapbuffer_test.go`) (AC: #6)
  - [ ] 8.1 `TestGapBuffer_CurrentLineContent_SingleLine`
  - [ ] 8.2 `TestGapBuffer_CurrentLineContent_MultiLine`
  - [ ] 8.3 `TestGapBuffer_DeleteCurrentLine_Middle`
  - [ ] 8.4 `TestGapBuffer_DeleteCurrentLine_LastLine`
  - [ ] 8.5 `TestGapBuffer_DeleteCurrentLine_OnlyLine`

- [ ] Task 9: Write integration tests (`internal/editor/editor_test.go`) (AC: #2-#6)
  - [ ] 9.1 `TestEditor_Paste_InsertsClipboardText`
  - [ ] 9.2 `TestEditor_PasteBefore_InsertsBeforeCursor`
  - [ ] 9.3 `TestEditor_DeleteLine_CopiesAndDeletes`
  - [ ] 9.4 `TestEditor_Paste_ClipboardUnavailable_ShowsError`
  - [ ] 9.5 `TestEditor_DeleteLine_EmptyBlock_CommitsBlock`

## Dev Notes

### Clipboard Library Decision: Shell Exec (No External Dependency)

The architecture doc identified three options for clipboard access. After analysis:

- **`golang.design/x/clipboard`**: Requires CGo on Linux (X11 headers), breaks cross-compilation, no native Wayland support. Too heavy.
- **`atotto/clipboard`**: Shell exec to platform tools, no CGo. But inactive/unmaintained since 2023.
- **Bubbletea v2 OSC52**: Built-in `tea.SetClipboard()` / `tea.ReadClipboard()`. Works over SSH but `ReadClipboard` has limited terminal support (many terminals don't implement OSC52 read).
- **Custom shell exec**: Same approach as `atotto/clipboard` but owned by us. Simple, ~80 lines, no external dependency.

**Decision: Build a minimal clipboard abstraction in `internal/editor/clipboard.go`** using `os/exec`. This is the same pattern used by neovim, micro, and other TUI editors. Platform detection:
- macOS: `pbcopy` / `pbpaste` (always available)
- Linux Wayland: `wl-copy` / `wl-paste` (check `WAYLAND_DISPLAY` env)
- Linux X11: `xclip -selection clipboard` / `xclip -selection clipboard -o` (check `DISPLAY` env)
- Windows: `clip.exe` (write) / `powershell.exe -command Get-Clipboard` (read)

Clipboard lives in `internal/editor/` (not a new package) because:
1. Only `editor` needs clipboard access (per architecture: "handled in `editor/actions.go`")
2. Avoids creating a new package for ~80 lines
3. Clipboard is an editor concern, not a block or vim concern

### Paste Flow (p/P in Normal Mode)

Pasting requires an active editing block because text is inserted via the gap buffer. The paste flow:

1. Normal mode handler returns `PasteAction{Before: bool}`
2. `applyAction()` receives the action
3. If `blockPendingCommit` is true: reuse existing `activeBuffer` (block is already revealed)
4. If no active block: temporarily enter insert mode on the block under cursor, paste, then exit back to normal
5. Read text from clipboard: `e.clipboard.Read()`
6. For `p` (after cursor): position is already correct in the gap buffer
7. For `P` (before cursor): move cursor left by one position before inserting
8. Insert each rune via `activeBuffer.Insert(r)`
9. Record undo before pasting (single undo step for entire paste)
10. Call `updateActiveBlockDisplay()` and `startAutoSaveTimer()`
11. If block was not already pending: commit the block back to rendered

**Handling multi-line paste**: If pasted text contains `\n\n` (block separator), the paste should still insert into the current block as-is. The user can re-parse blocks manually. Block splitting on paste is complex and out of scope for this story.

### Delete Line Flow (dd in Normal Mode)

The `dd` command requires a two-key sequence, similar to `gg` and `ZZ`:

1. First `d`: sets `pending = 'd'` in NormalHandler (like `g` and `Z` already work)
2. Second `d`: emits `DeleteLineAction{}`
3. Any other key after `d`: cancels pending, processes the key normally

In `applyAction()`:
1. Ensure active block exists (enter insert mode if needed via `blockPendingCommit` or `enterInsertMode`)
2. Record undo state
3. Get current line content via `activeBuffer.CurrentLineContent()` (new helper)
4. Write line content to clipboard via `e.clipboard.Write()`
5. Delete the line via `activeBuffer.DeleteCurrentLine()` (new helper)
6. Update display, trigger auto-save
7. If block is now empty: remove it from `[]Block` and commit

### Visual Mode Yank (AC #1) -- Partial Implementation

Visual mode (Story 5.4) doesn't exist yet. For this story:
- Define `YankAction` in action.go for future use
- The `y` key in normal mode without selection is a no-op (standard vim behavior)
- When 5.4 implements visual mode, the `YankAction` handler in `applyAction()` will read the selection and write to clipboard

### Gap Buffer Additions

Two new methods on `GapBuffer`:

```go
// CurrentLineContent returns the text of the line at the cursor position (no trailing \n).
func (g *GapBuffer) CurrentLineContent() string

// DeleteCurrentLine removes the entire line at the cursor position including the trailing \n.
// Cursor lands at the start of the next line, or end of the previous line if the last line was deleted.
func (g *GapBuffer) DeleteCurrentLine()
```

Implementation approach for `DeleteCurrentLine()`:
1. `MoveToLineStart()` to position at line beginning
2. Delete chars forward until `\n` or end of content
3. If a `\n` was found, delete it too (removes the line separator)
4. If cursor was on the last line and there's a preceding `\n`, backspace to remove the separator instead

### Error Handling

Clipboard operations can fail (tool not installed, permission denied, empty clipboard). Handle gracefully:
- `Read()` failure: show `E: Clipboard unavailable` in status bar, do not crash
- `Write()` failure: show `E: Clipboard unavailable` in status bar, still complete the delete/yank operation (data only lost from clipboard, not from document)
- `ErrClipboardUnavailable` sentinel error for the no-op fallback clipboard

### Previous Story Learnings (from 5-2-auto-pairs-for-markdown)

- Two-stage Esc flow (`blockPendingCommit`) is still active. Paste and dd must work correctly whether block is pending commit or not.
- `MultiAction` exists if we need to chain actions (unlikely needed here, but available).
- All gap buffer mutations must go through `applyAction()` with undo recording.
- `updateActiveBlockDisplay()` must be called after any buffer mutation.
- `startAutoSaveTimer()` must be called after content changes.
- New action types need `actionTag()` method and registration in the `switch` in `applyAction()`.

### Git Context (Recent Commits)

Latest commit `794f398` implemented auto-pairs. The codebase has:
- `internal/vim/action.go` -- all action types (InsertCharAction, BackspaceAction, InsertPairAction, etc.)
- `internal/vim/normal.go` -- NormalHandler with pending key support for `g` and `Z`
- `internal/editor/editor.go` -- applyAction() with all mutation hooks, blockPendingCommit flow
- `internal/block/gapbuffer.go` -- full gap buffer with line-aware operations

### No New Package Dependencies

This feature requires no new Go modules. Clipboard access uses `os/exec` from the standard library. Platform detection uses `runtime.GOOS` and `os.Getenv()`.

### Project Structure Notes

- **New file:** `internal/editor/clipboard.go` -- Clipboard interface, platform detection, exec-based implementation
- **New file:** `internal/editor/clipboard_test.go` -- Clipboard unit tests with mock
- **Modified:** `internal/vim/action.go` -- add PasteAction, YankAction, DeleteLineAction
- **Modified:** `internal/vim/normal.go` -- add `d` pending, `p`, `P` key bindings
- **Modified:** `internal/vim/normal_test.go` -- tests for new key bindings
- **Modified:** `internal/block/gapbuffer.go` -- add CurrentLineContent(), DeleteCurrentLine()
- **Modified:** `internal/block/gapbuffer_test.go` -- tests for new methods
- **Modified:** `internal/editor/editor.go` -- clipboard field, handle PasteAction/DeleteLineAction/YankAction in applyAction()
- **Modified:** `internal/editor/editor_test.go` -- integration tests with mock clipboard
- No new packages, respects dependency direction (clipboard lives in editor, not a separate package)

### References

- [Source: epics.md#Story 5.3] -- User story and all 6 acceptance criteria with BDD format
- [Source: architecture.md#Gap Analysis Results, Gap 3] -- "System clipboard integration unspecified... Options: golang.design/x/clipboard or shell exec to pbcopy/xclip/wl-copy"
- [Source: architecture.md#Component Communication] -- "Components return actions, editor applies them"
- [Source: architecture.md#Vim Mode Architecture] -- "Multi-key sequences tracked via operator-pending state"
- [Source: architecture.md#Package Boundary Rules] -- clipboard handled in editor (editor imports everything)
- [Source: architecture.md#Text Buffer] -- "Gap buffer: O(1) inserts/deletes at cursor position"
- [Source: prd.md#FR22] -- "User can yank (copy) text to the system clipboard and paste from it using vim commands (y, p) and Ctrl+V"
- [Source: internal/vim/normal.go] -- NormalHandler with pending key pattern for g, Z
- [Source: internal/vim/action.go] -- Current action types
- [Source: internal/block/gapbuffer.go] -- Gap buffer with line-aware operations
- [Source: internal/editor/editor.go:applyAction()] -- Mutation hooks with undo recording, blockPendingCommit flow
- [Source: 5-2-auto-pairs-for-markdown.md] -- Previous story learnings: MultiAction, undo integration, auto-save timer pattern

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
