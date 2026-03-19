# Story 5.2: Auto-Pairs for Markdown

Status: done

## Story

As a writer,
I want matching characters auto-inserted when I type markdown formatting pairs,
so that I can format my text efficiently without manually closing each pair.

## Acceptance Criteria

1. **Given** the user is in insert mode **When** the user types `**` **Then** a matching `**` is auto-inserted after the cursor, with the cursor positioned between the pairs (FR20)

2. **Given** the user is in insert mode **When** the user types `__` **Then** a matching `__` is auto-inserted after the cursor, with the cursor positioned between the pairs (FR20)

3. **Given** the user is in insert mode **When** the user types a single backtick `` ` `` **Then** a matching backtick is auto-inserted after the cursor, with the cursor positioned between them (FR20)

4. **Given** the user is in insert mode **When** the user types `[` **Then** a matching `]` is auto-inserted after the cursor, with the cursor positioned between them (FR20)

5. **Given** the user is in insert mode **When** the user types `(` **Then** a matching `)` is auto-inserted after the cursor, with the cursor positioned between them (FR20)

6. **Given** auto-paired characters have been inserted **When** the user types the closing character manually **Then** the cursor moves past the existing closing character rather than inserting a duplicate

## Tasks / Subtasks

- [x] Task 1: Define auto-pair data structures and pair registry (`internal/vim/autopair.go`) (AC: #1-#5)
  - [x] 1.1 Define `Pair` struct with `Opening rune`, `Closing rune`, `IsDouble bool` (for `**` and `__` which are two-char pairs)
  - [x] 1.2 Define `AutoPairState` struct tracking: `pendingStar bool`, `pendingUnderscore bool` (for detecting `**` and `__` double-char triggers)
  - [x] 1.3 Define `pairRegistry` -- a lookup map from opening char to `Pair` for single-char pairs (`` ` ``, `[`, `(`)
  - [x] 1.4 Implement `NewAutoPairState() *AutoPairState`
  - [x] 1.5 Implement `HandleChar(char rune, nextChar rune) (action Action, consumed bool)` -- main dispatch logic
  - [x] 1.6 Implement `Flush() Action` -- if pending star/underscore exists, emit `InsertCharAction` for the pending char and clear state

- [x] Task 2: Add new action types (`internal/vim/action.go`) (AC: #1-#6)
  - [x] 2.1 Add `InsertPairAction struct { Opening string; Closing string }` with `actionTag()` method
  - [x] 2.2 Add `SkipClosingAction struct { Count int }` with `actionTag()` method
  - [x] 2.3 Add `MultiAction struct { Actions []Action }` to support multi-step sequences (flush+mode change, flush+newline)

- [x] Task 3: Integrate AutoPairState into insert mode handler (`internal/vim/insert.go`) (AC: #1-#6)
  - [x] 3.1 Add `autoPair *AutoPairState` field to `InsertHandler` struct
  - [x] 3.2 Initialize `autoPair` in `NewInsertHandler()`
  - [x] 3.3 In `HandleKey()`, before returning `InsertCharAction` for printable chars: call `autoPair.HandleChar(char, nextCharFromEditor)`
  - [x] 3.4 On `Esc` key handling: call `autoPair.Flush()` and emit pending char action via MultiAction before ChangeModeAction
  - [x] 3.5 On `Enter` key handling: call `autoPair.Flush()` before emitting `InsertNewlineAction` via MultiAction

- [x] Task 4: Handle new actions in editor (`internal/editor/editor.go`) (AC: #1-#6)
  - [x] 4.1 Handle `InsertPairAction` in `applyAction()`: record undo, insert opening+closing, move cursor left by len(closing)
  - [x] 4.2 Handle `SkipClosingAction` in `applyAction()`: move cursor right by Count positions (no undo, no auto-save)
  - [x] 4.3 Handle `MultiAction` in `applyAction()`: apply each sub-action in sequence, batch returned commands

- [x] Task 5: Provide next-char context to insert handler (AC: #6)
  - [x] 5.1 Add `SetNextChar(r rune)` method to `InsertHandler`
  - [x] 5.2 In `editor.go`, before dispatching key in insert mode: read char after cursor, call `ih.SetNextChar()`
  - [x] 5.3 The insert handler uses this context in `AutoPairState.HandleChar()` to decide skip-over vs insert

- [x] Task 6: Write unit tests for AutoPairState (`internal/vim/autopair_test.go`) (AC: #1-#6)
  - [x] 6.1 `TestAutoPair_SingleBacktick_InsertsPair`
  - [x] 6.2 `TestAutoPair_OpenBracket_InsertsPair`
  - [x] 6.3 `TestAutoPair_OpenParen_InsertsPair`
  - [x] 6.4 `TestAutoPair_DoubleStar_InsertsPair`
  - [x] 6.5 `TestAutoPair_DoubleUnderscore_InsertsPair`
  - [x] 6.6 `TestAutoPair_SingleStar_NoPair`
  - [x] 6.7 `TestAutoPair_SkipClosing_Backtick`
  - [x] 6.8 `TestAutoPair_SkipClosing_Bracket`
  - [x] 6.9 `TestAutoPair_SkipClosing_Paren`
  - [x] 6.10 `TestAutoPair_Flush_PendingStar`
  - [x] 6.11 `TestAutoPair_Flush_NoPending`
  - [x] 6.12 `TestAutoPair_NormalChar_NotConsumed`

- [x] Task 7: Write integration-level tests (`internal/editor/editor_test.go`) (AC: #1-#6)
  - [x] 7.1 `TestEditor_AutoPair_Backtick_InsertsAndPositionsCursor`
  - [x] 7.2 `TestEditor_AutoPair_DoubleStar_InsertsAndPositionsCursor`
  - [x] 7.3 `TestEditor_AutoPair_SkipClosing`
  - [x] 7.4 `TestEditor_AutoPair_SingleStar_NoAutoPair`

## Dev Notes

### Architecture: Auto-Pairs in the Vim Action Pattern

Per the architecture, all input handling follows: `KeyMsg -> vim.HandleKey -> Action -> editor.applyAction`. Auto-pairs follow this same pattern. The `InsertHandler` in `internal/vim/insert.go` decides WHAT action to emit, and `editor.go` APPLIES the action to the gap buffer.

The `AutoPairState` struct lives on the `InsertHandler` because pair detection requires tracking multi-keystroke sequences (`*` then `*` for bold). This is similar to how `NormalHandler` tracks pending operators for multi-key vim commands.

### Key Design Decision: Double-Char Pairs (`**` and `__`)

Unlike simple single-char pairs (`` ` ``, `[`, `(`), bold and underscore emphasis require TWO characters to trigger. This means:
- First `*` keystroke: consumed by AutoPairState, sets `pendingStar = true`, returns consumed=true (no action emitted yet)
- Second `*` keystroke: detects pending, emits `InsertPairAction{Opening: "**", Closing: "**"}`, clears pending
- If user types `*` then ANY OTHER CHAR: flush the pending `*` as a regular `InsertCharAction`, then process the new char normally

This two-keystroke buffering is essential. The insert handler must NOT immediately insert `*` on the first keystroke -- it must wait to see if a second `*` follows.

**Edge case: Esc after single `*`**: The `Flush()` method handles this -- any pending char is emitted as a regular insert before mode transition.

### Key Design Decision: Skip-Over Behavior

When the user types a closing character and that same character already exists immediately after the cursor, the cursor should SKIP OVER the existing character rather than inserting a duplicate. This is the standard auto-pair UX.

Implementation approach:
- The editor provides "next char after cursor" context to the insert handler before each keystroke
- `AutoPairState.HandleChar()` checks: if the typed char is a closing char AND nextChar matches, return `SkipClosingAction`
- For `**` skip-over: when `*` is typed and pending `*` exists, AND the next two chars are `**`, emit `SkipClosingAction{Count: 2}` instead of inserting

Getting the next char context: Read from `activeBuffer.Content()` at `activeBuffer.CursorPos()` position. This is O(n) for Content() but blocks are small so it's negligible.

### Gap Buffer Operations for Auto-Pairs

The gap buffer (`internal/block/gapbuffer.go`) provides all needed operations:
- `Insert(rune)` -- insert opening/closing chars (O(1) at cursor)
- `MoveLeft() bool` -- reposition cursor between pair after insertion
- `MoveRight() bool` -- skip-over closing char
- `Content() string` -- read buffer to check next char for skip-over detection
- `CursorPos() int` -- get cursor position for next-char lookup

Pattern for pair insertion:
```go
// Insert "**" opening
buf.Insert('*')
buf.Insert('*')
// Insert "**" closing
buf.Insert('*')
buf.Insert('*')
// Move cursor back between pairs
buf.MoveLeft()
buf.MoveLeft()
// Result: "**|**" where | is cursor
```

### Undo Integration

Auto-pair insertions should record undo state BEFORE the pair is inserted, just like `InsertCharAction`. This way `u` undoes the entire pair insertion as a single undo step. Use the same `e.undoManager.Record()` call with actionType `"insert"` so the debounce grouping works naturally with surrounding text.

Skip-over actions do NOT need undo recording since they don't change content (cursor movement only).

### No New Package Dependencies

This feature requires no new Go packages. Everything is built with standard library types.

### Previous Story Learnings (from 5-1-undo-and-redo)

- Two-stage Esc flow: `blockPendingCommit` flag on EditorModel. First Esc keeps activeBuffer alive for undo/redo, second Esc or navigation commits. Auto-pairs work within insert mode, so this flow is unaffected -- but `Flush()` must be called when Esc is pressed to handle any pending `*`/`_`.
- All gap buffer mutations in `applyAction()` already have undo recording hooks. New auto-pair actions follow the same pattern.
- `updateActiveBlockDisplay()` is the standard call after any buffer mutation to refresh the viewport.
- `startAutoSaveTimer()` must be called after pair insertion (content changed) but NOT after skip-over (no content change).

### Git Context (Recent Commits)

Latest commit `948654e` implemented undo/redo. The codebase currently has:
- `internal/editor/history.go` -- UndoManager with debounce/grouping
- `internal/vim/action.go` -- action types including InsertCharAction, BackspaceAction
- `internal/vim/insert.go` -- InsertHandler with HandleKey dispatch
- `internal/editor/editor.go` -- applyAction with all mutation hooks

All these files will be modified or extended for auto-pairs.

### Project Structure Notes

- **New file:** `internal/vim/autopair.go` -- AutoPairState struct and pair registry
- **New file:** `internal/vim/autopair_test.go` -- AutoPairState unit tests
- **Modified:** `internal/vim/action.go` -- add InsertPairAction, SkipClosingAction
- **Modified:** `internal/vim/insert.go` -- integrate AutoPairState into InsertHandler
- **Modified:** `internal/editor/editor.go` -- handle InsertPairAction and SkipClosingAction in applyAction()
- **Modified:** `internal/editor/editor_test.go` -- integration tests for auto-pair behavior
- No new packages, no dependency changes
- All new code stays within `internal/vim` (auto-pair logic) and `internal/editor` (action application) -- respects dependency direction

### References

- [Source: epics.md#Story 5.2] -- User story and all 6 acceptance criteria with BDD format
- [Source: architecture.md#Vim Mode Architecture] -- "In-house implementation with per-mode handler pattern"
- [Source: architecture.md#Component Communication] -- "Components return actions, editor applies them"
- [Source: architecture.md#Text Buffer] -- "Gap buffer: O(1) inserts/deletes at cursor position"
- [Source: architecture.md#Package Boundary Rules] -- "internal/vim depends on internal/block"
- [Source: ux-design-specification.md#Insert Mode] -- "Auto-pairs active; closing character auto-inserted, cursor positioned between"
- [Source: prd.md#FR20] -- "User can have matching characters auto-inserted for markdown pairs (**, __, `, [], ())"
- [Source: internal/vim/insert.go] -- InsertHandler.HandleKey() dispatch for printable chars
- [Source: internal/vim/action.go] -- Current action types
- [Source: internal/block/gapbuffer.go] -- Gap buffer: Insert(), MoveLeft(), MoveRight(), Content(), CursorPos()
- [Source: internal/editor/editor.go:applyAction()] -- Mutation hooks with undo recording
- [Source: 5-1-undo-and-redo.md] -- Two-stage Esc flow, UndoManager integration, mutation points

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

- Implemented `AutoPairState` in `internal/vim/autopair.go` with `HandleChar` and `Flush` methods. Skip-over check placed before pair insertion in HandleChar so backtick skip-over takes priority over new pair creation.
- Added `InsertPairAction`, `SkipClosingAction`, and `MultiAction` to `internal/vim/action.go`. MultiAction was added (beyond story spec) to handle flush+mode-change and flush+newline sequences atomically from a single HandleKey call.
- Modified `InsertHandler` with `autoPair *AutoPairState` and `nextChar rune` fields. `SetNextChar` is called by the editor before each keystroke dispatch in insert mode.
- Editor reads next char from `activeBuffer.Content()[CursorPos()]` before dispatching to handler. InsertPairAction records undo, inserts both halves, moves cursor left. SkipClosingAction does bounds-checked MoveRight with no undo/auto-save.
- All 12 unit tests and 4 integration tests pass. Full regression suite clean.

### File List

- `internal/vim/autopair.go` (new)
- `internal/vim/autopair_test.go` (new)
- `internal/vim/action.go` (modified — added InsertPairAction, SkipClosingAction, MultiAction)
- `internal/vim/insert.go` (modified — integrated AutoPairState)
- `internal/editor/editor.go` (modified — next-char context, new action handlers)
- `internal/editor/editor_test.go` (modified — added 4 auto-pair integration tests)

## Change Log

- 2026-03-18: Implemented auto-pairs for markdown (**, __, `, [], ()) with skip-over behavior. Added AutoPairState, new action types (InsertPairAction, SkipClosingAction, MultiAction), integrated into InsertHandler and editor. 12 unit tests + 4 integration tests added.
- 2026-03-18: Code review fixes — (1) Added ** and __ skip-over logic with second lookahead char (was CRITICAL: skip-over only worked for single-char pairs). (2) Added pending state flush on arrow keys, backspace, delete, and tab (was HIGH: pending chars disappeared on cursor movement). (3) Added 5 new unit tests and 5 new integration tests covering skip-over for double pairs, underscore flush, bracket/paren pairs, and arrow-key flush. Total: 17 unit + 9 integration tests.
