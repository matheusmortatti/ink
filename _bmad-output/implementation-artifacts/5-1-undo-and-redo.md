# Story 5.1: Undo and Redo

Status: done

## Story

As a writer,
I want to undo and redo my edits within a block,
so that I can experiment freely knowing I can reverse any change.

## Acceptance Criteria

1. **Given** the user has made edits within an active editing block **When** the user presses `u` in normal mode **Then** the most recent edit is undone and the block content reverts to the previous state (FR19)

2. **Given** the user has undone one or more edits **When** the user presses `Ctrl+r` in normal mode **Then** the most recently undone edit is reapplied (FR19)

3. **Given** the user has made multiple edits within a block **When** `u` is pressed repeatedly **Then** edits are undone in reverse chronological order, one at a time

4. **Given** there are no more edits to undo **When** the user presses `u` **Then** nothing happens -- no error, no crash

5. **Given** the user exits a block with `Esc` (rendering it) **When** the block transitions to rendered state **Then** the undo/redo stack for that block is cleared -- the rendered block is the committed state

6. **Given** the user re-enters the same block **When** insert mode is activated **Then** a fresh undo/redo stack begins -- previous session edits are not available for undo

## Tasks / Subtasks

- [x] Task 1: Create undo/redo history manager (`internal/editor/history.go`) (AC: #1-#4)
  - [x] 1.1 Define `UndoEntry` struct containing: gap buffer content snapshot (`string`), cursor position (`int`), and gap buffer cursor line/col (`int, int`)
  - [x] 1.2 Define `UndoManager` struct with: `undoStack []UndoEntry`, `redoStack []UndoEntry`, `maxSize int` (default 100)
  - [x] 1.3 Implement `NewUndoManager(maxSize int) *UndoManager`
  - [x] 1.4 Implement `Record(content string, cursorPos int, cursorLine int, cursorCol int, actionType string)` -- pushes current state to undo stack, clears redo stack, enforces maxSize cap; added `actionType` param for grouping logic per Dev Notes
  - [x] 1.5 Implement `Undo(currentContent string, currentCursorPos int, currentCursorLine int, currentCursorCol int) (entry UndoEntry, ok bool)` -- saves current state to redo stack, pops and returns from undo stack
  - [x] 1.6 Implement `Redo(currentContent string, currentCursorPos int, currentCursorLine int, currentCursorCol int) (entry UndoEntry, ok bool)` -- saves current state to undo stack, pops and returns from redo stack
  - [x] 1.7 Implement `Clear()` -- resets both stacks (called on block exit per AC #5)
  - [x] 1.8 Implement `CanUndo() bool` and `CanRedo() bool` helpers

- [x] Task 2: Add `UndoAction` and `RedoAction` to vim action system (`internal/vim/action.go`) (AC: #1-#2)
  - [x] 2.1 Add `UndoAction struct{}` with `actionTag()` method
  - [x] 2.2 Add `RedoAction struct{}` with `actionTag()` method

- [x] Task 3: Handle `u` and `Ctrl+r` in normal mode handler (`internal/vim/normal.go`) (AC: #1-#2)
  - [x] 3.1 Add case for `"u"` key in `handleSingleKey()` returning `UndoAction{}`
  - [x] 3.2 Add case for `"ctrl+r"` key in `handleSingleKey()` returning `RedoAction{}`

- [x] Task 4: Integrate undo/redo and pending-commit state into editor (`internal/editor/editor.go`) (AC: #1-#6)
  - [x] 4.1 Add fields to `EditorModel`: `undoManager *UndoManager`, `blockPendingCommit bool`
  - [x] 4.2 Initialize `undoManager` in `NewEditor()` with `NewUndoManager(100)`
  - [x] 4.3 Record state BEFORE mutations in `applyAction()` for all 5 mutation actions with debounce/grouping by actionType
  - [x] 4.4 Handle `UndoAction` in `applyAction()`: check `e.activeBuffer != nil && e.blockPendingCommit`, apply undo and update display
  - [x] 4.5 Handle `RedoAction` in `applyAction()`: same pattern as undo but using `e.undoManager.Redo(...)`
  - [x] 4.6 In `enterInsertMode()` and `initViewport()` empty-canvas path: record the initial state as the first undo entry
  - [x] 4.7 Add `commitPendingBlock()` method: calls `exitInsertMode()`, clears undo history, sets `blockPendingCommit = false`
  - [x] 4.8 Modify `ChangeModeAction` handler for Normal: set `blockPendingCommit = true` and switch mode handler without calling `exitInsertMode()`
  - [x] 4.9 Modify `ChangeModeAction` handler for Insert: if `blockPendingCommit`, reuse existing `activeBuffer`, set `blockPendingCommit = false`
  - [x] 4.10 Modify movement action handlers (`MoveCursorAction`, `DocumentPositionAction`, `ScrollAction`, `WordMotionAction`): call `commitPendingBlock()` if pending
  - [x] 4.11 Modify `QuitAction` handler: call `commitPendingBlock()` if pending. Also added same guard for `ChangeModeAction` to Command mode.

- [x] Task 5: Write unit tests for UndoManager (`internal/editor/history_test.go`) (AC: #1-#6)
  - [x] 5.1 `TestUndoManager_Record_PushesToUndoStack`
  - [x] 5.2 `TestUndoManager_Undo_RestoresPreviousState`
  - [x] 5.3 `TestUndoManager_Redo_ReappliesUndoneState`
  - [x] 5.4 `TestUndoManager_Undo_EmptyStack_ReturnsFalse`
  - [x] 5.5 `TestUndoManager_Redo_EmptyStack_ReturnsFalse`
  - [x] 5.6 `TestUndoManager_Record_ClearsRedoStack`
  - [x] 5.7 `TestUndoManager_Clear_ResetsBothStacks`
  - [x] 5.8 `TestUndoManager_MaxSize_EnforcesLimit`
  - [x] 5.9 `TestUndoManager_MultipleUndos_ReverseChronological`

- [x] Task 6: Write integration-level tests for undo/redo key handling (`internal/vim/normal_test.go`) (AC: #1-#2)
  - [x] 6.1 `TestNormalHandler_UndoKey_ReturnsUndoAction`
  - [x] 6.2 `TestNormalHandler_RedoKey_ReturnsRedoAction`

## Dev Notes

### Architecture: Block-Scoped Undo/Redo

Per the architecture document, undo/redo operates **within the active editing block only** at the gap buffer level. There is NO document-level undo (undoing block transitions, block deletion, etc.) in MVP. This significantly simplifies implementation:

- The `UndoManager` lives on the `EditorModel`, NOT on individual blocks
- It is populated when entering insert mode and cleared when exiting
- Each undo entry is a snapshot of the gap buffer content + cursor position
- Undo/redo only works while actively editing a block (insert mode)

[Source: architecture.md#Undo/Redo] -- "Undo/redo operates within the active editing block only (gap buffer level). Undo stack stores gap buffer snapshots or operation deltas. Exiting a block (Esc) clears the undo stack for that block."

### Key Implementation Decision: Snapshot vs. Delta

**Recommended: Content snapshots** (not operation deltas).

Rationale:
- Blocks are typically small (a paragraph, a heading, a code fence) -- snapshot cost is negligible
- Snapshot approach is simpler, more reliable, and easier to debug
- No risk of delta replay bugs (e.g., applying operations in wrong order)
- 100 snapshots of a typical paragraph (~500 bytes) = ~50KB -- trivial memory

Each `UndoEntry` stores:
```go
type UndoEntry struct {
    Content    string  // Full gap buffer content at that point
    CursorPos  int     // Absolute cursor position in content
    CursorLine int     // For restoring line/col positioning
    CursorCol  int
}
```

### Key Implementation Decision: Keystroke Grouping

Recording every single keystroke as a separate undo entry would create an unusable undo experience (user presses `u` and only one character is removed). Instead, use **time-based grouping with action-type boundaries**:

1. **Time threshold**: Only create a new undo entry if >500ms has elapsed since the last recorded entry
2. **Action-type boundary**: Always create a new entry when the action type changes (e.g., switching from typing characters to pressing backspace, or vice versa)
3. **Newline boundary**: Always create a new entry on newline insertion (Enter key)

This gives natural "word-at-a-time" undo granularity similar to vim's behavior, where `u` undoes the last "chunk" of typing rather than individual characters.

Implementation: Track `lastRecordTime time.Time` and `lastActionType` on the `UndoManager`. In the `Record` method, compare against these to decide whether to create a new entry or skip (letting the previous entry remain as the restore point).

### Mutation Points in Editor (Where to Hook)

All gap buffer mutations happen in `applyAction()` in `internal/editor/editor.go`:

| Action | Line | Gap Buffer Call | Record Before? |
|--------|------|-----------------|----------------|
| `InsertCharAction` | ~265 | `activeBuffer.Insert(a.Char)` | Yes (with grouping) |
| `BackspaceAction` | ~277 | `activeBuffer.Backspace()` | Yes (with grouping) |
| `DeleteCharAction` | ~284 | `activeBuffer.Delete()` | Yes (with grouping) |
| `InsertNewlineAction` | ~299 | `activeBuffer.Insert('\n')` | Yes (always -- newline boundary) |
| `InsertTabAction` | ~306 | `activeBuffer.Insert('\t')` | Yes (with grouping) |

Cursor movement actions (`MoveCursorAction`, etc.) do NOT need undo recording -- they don't change content.

### Block Pending-Commit State: Two-Stage Esc for Undo/Redo

**Problem**: AC #1-#2 require `u`/`Ctrl+r` in normal mode, but undo history is built during insert mode. Currently, `Esc` in insert mode immediately calls `exitInsertMode()` which commits the block content and clears `activeBuffer`. There is no intermediate state where the block is still active but the editor is in normal mode.

**DEFINITIVE DESIGN**: Introduce a `blockPendingCommit bool` flag on `EditorModel`. This creates a two-stage Esc flow:

1. **First `Esc`** (insert mode): Switch mode to normal, set `blockPendingCommit = true`, but do NOT call `exitInsertMode()`. The block remains active (`activeBuffer` and `activeBlockIdx` stay alive), the viewport continues showing raw content.
2. **`u` / `Ctrl+r`** (normal mode, `blockPendingCommit == true`): Apply undo/redo on the active buffer. Update display via `updateActiveBlockDisplay()`.
3. **`i` / `a`** (normal mode, `blockPendingCommit == true`): Re-enter insert mode in the same block. No new gap buffer created -- reuse `activeBuffer`.
4. **Second `Esc`** (normal mode, `blockPendingCommit == true`): NOW call `exitInsertMode()` -- commit block content, clear `activeBuffer`, clear undo history, set `blockPendingCommit = false`.
5. **Navigation away** (`j`/`k`/`G`/`gg` etc., `blockPendingCommit == true`): Same as second Esc -- commit and clear before navigating.

**Implementation in `editor.go`**:
- Add `blockPendingCommit bool` field to `EditorModel`
- In the `ChangeModeAction` handler for `Normal` mode: if `activeBuffer != nil`, set `blockPendingCommit = true` and return (skip `exitInsertMode()`)
- In the `ChangeModeAction` handler for `Insert` mode: if `blockPendingCommit`, just switch mode back to insert (reuse active buffer), set `blockPendingCommit = false`
- In `UndoAction`/`RedoAction` handlers: check `blockPendingCommit && activeBuffer != nil`, apply undo/redo, else no-op
- In movement action handlers (`MoveCursorAction`, `DocumentPositionAction`, `ScrollAction`): if `blockPendingCommit`, call `commitPendingBlock()` first (which runs `exitInsertMode()` + clears undo + sets flag false), then apply movement
- Add `commitPendingBlock()` method that encapsulates the commit-and-clear logic

**State transitions**:
```
INSERT MODE (activeBuffer != nil, blockPendingCommit = false)
    │ Esc
    ▼
NORMAL MODE + PENDING COMMIT (activeBuffer != nil, blockPendingCommit = true)
    │ u/Ctrl+r → undo/redo on activeBuffer
    │ i/a → back to INSERT MODE (reuse activeBuffer)
    │ Esc / j/k/navigation → commitPendingBlock() → NORMAL MODE (no active block)
    ▼
NORMAL MODE (activeBuffer = nil, blockPendingCommit = false)
```

This satisfies ALL 6 acceptance criteria:
- AC #1-#2: `u`/`Ctrl+r` work in normal mode while block is pending commit
- AC #3-#4: Multiple undos work; empty stack = no-op
- AC #5: Committing (second Esc / navigation) clears undo history
- AC #6: Re-entering a committed block starts fresh

### Gap Buffer Replacement on Undo/Redo

When undoing/redoing, the active gap buffer must be replaced with the snapshot content:
```go
e.activeBuffer = block.NewGapBuffer(entry.Content)
e.activeBuffer.SetCursorPos(entry.CursorPos)
```
Then call `e.updateActiveBlockDisplay()` to refresh the viewport.

### No New Package Dependencies

This feature requires no new Go packages. Everything is built with standard library types (`string`, `int`, `time.Time`, slices).

### Testing Standards

Per architecture:
- Tests co-located with source files
- Table-driven tests for multiple scenarios
- `TestFunctionName_Scenario_ExpectedBehavior` naming
- No external test framework

### Project Structure Notes

- **New file:** `internal/editor/history.go` -- UndoManager and UndoEntry types
- **New file:** `internal/editor/history_test.go` -- UndoManager unit tests
- **Modified:** `internal/vim/action.go` -- add UndoAction, RedoAction
- **Modified:** `internal/vim/normal.go` -- handle "u" and "ctrl+r" keys
- **Modified:** `internal/editor/editor.go` -- integrate UndoManager into EditorModel, hook mutations, handle undo/redo actions, modify block exit flow for pending-commit state
- **Modified:** `internal/vim/normal_test.go` -- add tests for undo/redo key mappings
- No new packages, no dependency changes

### References

- [Source: epics.md#Story 5.1] -- User story and all 6 acceptance criteria with BDD format
- [Source: architecture.md#Undo/Redo] -- "Undo/redo operates within the active editing block only (gap buffer level). Undo stack stores gap buffer snapshots or operation deltas. Exiting a block (Esc) clears the undo stack."
- [Source: architecture.md#Text Buffer] -- "Gap buffer: Classic text editor data structure with O(1) inserts/deletes at cursor position"
- [Source: architecture.md#Component Communication] -- "Single top-level Bubbletea model with component structs; EditorModel owns lifecycle"
- [Source: prd.md#FR19] -- "User can undo and redo edits"
- [Source: internal/block/gapbuffer.go] -- Gap buffer implementation (273 lines): Insert, Backspace, Delete, Content(), CursorPos(), SetCursorPos(), CursorLineCol()
- [Source: internal/editor/editor.go:~265-306] -- All gap buffer mutation points in applyAction()
- [Source: internal/editor/editor.go:~418-459] -- enterInsertMode(): creates new GapBuffer from block raw content
- [Source: internal/editor/editor.go:~461-497] -- exitInsertMode(): commits gap buffer content to block, clears activeBuffer
- [Source: internal/vim/action.go] -- Current action types (76 lines); no undo/redo actions yet
- [Source: internal/vim/normal.go:handleSingleKey()] -- Where "u" and "ctrl+r" cases need to be added
- [Source: 4-3-panic-recovery-and-emergency-save.md] -- Most recent story; confirms e pointer stability through p.Run()
- [Source: internal/editor/editor.go:~699-736] -- splitActiveBlock() is another mutation point (double-Enter creates new block)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None — implementation proceeded cleanly without requiring debug logging.

### Completion Notes List

- Implemented block-scoped undo/redo per architecture spec: `UndoManager` on `EditorModel`, cleared on block exit
- Added `actionType string` param to `Record()` beyond story spec (Dev Notes required it for grouping logic)
- Two-stage Esc flow implemented: first Esc → pending-commit (block stays alive), second Esc or navigation → commit. All 18+ existing tests updated to reflect this new exit flow.
- Added `"esc"` case to `NormalHandler` to support second-Esc commit via `ChangeModeAction{Mode: Normal}` from normal mode
- Also added `commitPendingBlock()` guard before entering command mode (`:`) to prevent stale pending state
- `forceRecord()` internal method on `UndoManager` added for deterministic test setup (bypasses debounce)
- All 9 UndoManager unit tests pass; all 2 vim integration tests pass; full regression suite passes (all packages green)

### File List

- `internal/editor/history.go` (new) — UndoEntry, UndoManager types and all methods
- `internal/editor/history_test.go` (new) — 9 UndoManager unit tests
- `internal/vim/action.go` (modified) — added UndoAction, RedoAction types
- `internal/vim/normal.go` (modified) — added "u" → UndoAction, "ctrl+r" → RedoAction, "esc" → ChangeModeAction{Normal} cases
- `internal/vim/normal_test.go` (modified) — added TestNormalHandler_UndoKey and TestNormalHandler_RedoKey
- `internal/editor/editor.go` (modified) — integrated UndoManager, blockPendingCommit flag, pending-commit state machine, commitPendingBlock(), UndoAction/RedoAction handlers, mutation hooks with debounce, View() cursor shape for pending state, isDimmed fix
- `internal/editor/editor_test.go` (modified) — updated 18+ existing tests for two-stage Esc flow; added 10 undo/redo integration tests

## Change Log

- 2026-03-05: Implemented story 5.1 — undo/redo within active editing block. Created history.go with UndoManager (snapshot-based, 100-entry cap, time/type debounce). Added UndoAction/RedoAction to vim action system. Implemented two-stage Esc pending-commit flow. Updated existing test suite (18+ tests) for new two-stage exit behavior. All tests pass.
- 2026-03-05: Code review fixes — (H1) Fixed undo history leak across splitActiveBlock by clearing undo stack and recording initial state for new block. (H2) Added 10 editor-level integration tests for undo/redo flow. (H3) Fixed variant cursor adjustment (a/o/O) when re-entering insert from pending-commit state. (M1) Added editor_test.go to File List.
