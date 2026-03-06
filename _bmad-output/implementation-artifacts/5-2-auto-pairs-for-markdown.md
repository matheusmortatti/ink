# Story 5.2: Auto-Pairs for Markdown

Status: ready-for-dev

## Story

As a writer,
I want matching characters auto-inserted when I type markdown formatting pairs,
so that I can format my text efficiently without manually closing each pair.

## Acceptance Criteria

1. **Given** the user is in insert mode **When** the user types `**` (two asterisks) **Then** a matching `**` is auto-inserted after the cursor, with the cursor positioned between the pairs — resulting in `**|**` (FR20)

2. **Given** the user is in insert mode **When** the user types `__` (two underscores) **Then** a matching `__` is auto-inserted after the cursor, with the cursor positioned between the pairs — resulting in `__|__` (FR20)

3. **Given** the user is in insert mode **When** the user types a single backtick `` ` `` **Then** a matching backtick is auto-inserted after the cursor, with the cursor positioned between them — resulting in `` `|` `` (FR20)

4. **Given** the user is in insert mode **When** the user types `[` **Then** a matching `]` is auto-inserted after the cursor, with the cursor positioned between them — resulting in `[|]` (FR20)

5. **Given** the user is in insert mode **When** the user types `(` **Then** a matching `)` is auto-inserted after the cursor, with the cursor positioned between them — resulting in `(|)` (FR20)

6. **Given** auto-paired characters have been inserted (e.g. `**|**`) **When** the user types the closing character(s) manually **Then** the cursor moves past the existing closing character rather than inserting a duplicate

## Tasks / Subtasks

- [ ] Task 1: Add `PeekRight() (rune, bool)` to GapBuffer (`internal/block/gapbuffer.go`) (AC: #6)
  - [ ] 1.1 Implement `PeekRight() (rune, bool)` — returns the rune immediately after the cursor (at `buf[gapEnd]`) and true; returns `0, false` if cursor is at end of content
  - [ ] 1.2 Add test `TestGapBuffer_PeekRight_ReturnsNextChar` and `TestGapBuffer_PeekRight_AtEnd_ReturnsFalse` to `internal/block/gapbuffer_test.go`

- [ ] Task 2: Implement auto-pair logic helper in editor (`internal/editor/editor.go`) (AC: #1–#6)
  - [ ] 2.1 Add package-level helper `autoPairClosing(char rune, buf *block.GapBuffer) (closing string, ok bool)`:
    - `` ` `` → `"`"` (single backtick)
    - `[` → `"]"`
    - `(` → `")"`
    - `*` when the char immediately before cursor is `*` → `"**"`
    - `_` when the char immediately before cursor is `_` → `"__"`
    - All other chars → `"", false`
    - "Char immediately before cursor" is determined by reading `Content()[:CursorPos()]` and checking the last rune
  - [ ] 2.2 Add package-level helper `isAutoClosingChar(char rune) bool` — returns true for: `*`, `_`, `` ` ``, `]`, `)`
    - Used for skip-over detection
  - [ ] 2.3 Modify `InsertCharAction` handler in `applyAction()`:
    - **Skip-over check** (BEFORE inserting): if `isAutoClosingChar(a.Char)` AND `PeekRight()` returns the same rune → call `e.activeBuffer.MoveRight()` and return early (no insert, no undo record)
    - **Insert** the character as before (with undo recording)
    - **Auto-pair check** (AFTER inserting): call `autoPairClosing(a.Char, e.activeBuffer)` — if it returns a closing string, insert each rune of closing string then move cursor left by `len([]rune(closing))` positions

- [ ] Task 3: Write editor integration tests (`internal/editor/editor_test.go`) (AC: #1–#6)
  - [ ] 3.1 `TestEditor_AutoPair_Backtick_InsertsClosingAndPositionsCursor`
  - [ ] 3.2 `TestEditor_AutoPair_OpenBracket_InsertsClosingAndPositionsCursor`
  - [ ] 3.3 `TestEditor_AutoPair_OpenParen_InsertsClosingAndPositionsCursor`
  - [ ] 3.4 `TestEditor_AutoPair_DoubleAsterisk_InsertsClosingAndPositionsCursor`
  - [ ] 3.5 `TestEditor_AutoPair_DoubleUnderscore_InsertsClosingAndPositionsCursor`
  - [ ] 3.6 `TestEditor_AutoPair_SingleAsterisk_NoAutoPair` (single `*` does NOT trigger auto-pair)
  - [ ] 3.7 `TestEditor_AutoPair_SingleUnderscore_NoAutoPair` (single `_` does NOT trigger auto-pair)
  - [ ] 3.8 `TestEditor_AutoPair_SkipOver_Backtick`
  - [ ] 3.9 `TestEditor_AutoPair_SkipOver_CloseBracket`
  - [ ] 3.10 `TestEditor_AutoPair_SkipOver_CloseParen`
  - [ ] 3.11 `TestEditor_AutoPair_SkipOver_AsteriskInPair` (typing `*` when `*` is next char)
  - [ ] 3.12 `TestEditor_AutoPair_NoSkipOver_WhenNextCharDiffers` (no skip when next char is different)

## Dev Notes

### Architecture: Where Auto-Pairs Live

Per the architecture document, `internal/editor/editor.go` owns all state mutation via `applyAction()`. The `InsertHandler` (`internal/vim/insert.go`) remains **stateless** — it continues to return `InsertCharAction{Char: r}` for all printable characters. All auto-pair intelligence lives in the editor's `InsertCharAction` handler.

This follows the architecture's component delegation pattern: components (mode handlers) return actions → editor applies them and makes all state decisions.

[Source: architecture.md#Component Communication] — "Only EditorModel.applyAction() mutates top-level state"
[Source: architecture.md#Vim Mode Architecture] — "Each mode is a handler struct implementing a common interface. VimTea used as reference for motion/command implementations, not as a dependency."

### Trigger Rules (Critical: Double-Char vs Single-Char)

| Typed | Trigger condition | Closing inserted | Cursor after |
|---|---|---|---|
| `` ` `` | Always | `` ` `` (1 char) | Between backticks |
| `[` | Always | `]` (1 char) | `[|]` |
| `(` | Always | `)` (1 char) | `(|)` |
| `*` | Prev char is `*` | `**` (2 chars) | `**|**` |
| `_` | Prev char is `_` | `__` (2 chars) | `__|__` |

The `**`/`__` double-char trigger avoids false positives on single `*` (italic, list bullets, horizontal rules) and single `_`. The "previous char" check happens AFTER inserting `a.Char`, so `Content()[:CursorPos()]` ends with `**` or `__`.

### Skip-Over Behavior (AC #6)

When the user types a closing character that is already immediately after the cursor, move right instead of inserting. This allows the cursor to "escape" auto-paired content naturally.

```
Before:  **|**  (cursor between pairs)
Type `*` → next char is `*` → MoveRight → ***|*
Type `*` again → next char is `*` → MoveRight → ****|
```

Skip-over chars: `*`, `_`, `` ` ``, `]`, `)`

**Implementation order in `InsertCharAction` handler:**
1. Check skip-over FIRST (before inserting). If triggered: `MoveRight()`, do NOT record undo, break.
2. Record undo (as before).
3. Insert `a.Char`.
4. Check auto-pair. If triggered: insert closing chars, move cursor left.
5. `updateActiveBlockDisplay()` + `startAutoSaveTimer()`.

The skip-over check must come before undo recording and insertion to avoid corrupt buffer state.

### GapBuffer: PeekRight Addition

The gap buffer needs a `PeekRight() (rune, bool)` method to cleanly check the character after the cursor without moving it. Internally, the char after cursor is at `buf[gapEnd]`.

```go
// PeekRight returns the rune immediately after the cursor without moving it.
// Returns 0, false if the cursor is at end of content.
func (g *GapBuffer) PeekRight() (rune, bool) {
    if g.gapEnd == len(g.buf) {
        return 0, false
    }
    return g.buf[g.gapEnd], true
}
```

This is a leaf-package addition (`internal/block`) — no import changes needed.

[Source: architecture.md#Package Boundary Rules] — "`internal/block` is a leaf package — no dependencies on other internal packages"

### Content-Before-Cursor Inspection (for `**`/`__` detection)

After inserting `a.Char`, use:
```go
content := e.activeBuffer.Content()
pos := e.activeBuffer.CursorPos()
runes := []rune(content)
if pos >= 2 && runes[pos-1] == '*' && runes[pos-2] == '*' {
    // double-asterisk trigger
}
```
No new gap buffer method needed — `Content()` and `CursorPos()` are already available.

### No New Package Dependencies

This feature requires zero new Go packages. Everything uses:
- `internal/block` (GapBuffer additions: `PeekRight`)
- `internal/editor` (auto-pair logic in `applyAction`)
- Standard library only (`strings` is already imported in editor.go)

### Previous Story Learnings (from 5-1-undo-and-redo)

- **Two-stage Esc flow**: The editor uses `blockPendingCommit` for pending block commit. The `InsertCharAction` handler only runs while `modeHandler.Mode() == vim.Insert` (activeBuffer is set). Auto-pair changes don't interact with the pending-commit state at all — they only fire during active insert mode.
- **Undo recording hook**: `e.undoManager.Record(...)` is called BEFORE each mutation. For auto-pair, record undo ONCE before the first char insert. The subsequent closing char insertions + cursor moves are part of the same "atomic" auto-pair operation and should NOT be recorded separately. This gives correct undo behavior: pressing `u` after auto-pair undoes the entire pair insertion at once.
- **`updateActiveBlockDisplay()` call**: Always call once at the end of the `InsertCharAction` handler after all buffer mutations — do not call multiple times.
- **Test infrastructure**: Editor tests use `setupEditorWithText(t, content)` or `NewEditor()` style helpers — check `internal/editor/editor_test.go` for existing test setup patterns before writing new tests.
- **No new files needed** in `internal/vim/` — `InsertHandler` is unchanged.

[Source: 5-1-undo-and-redo.md#Dev Notes] — Undo recording, two-stage Esc, mutation hook patterns

### Git Intelligence (Recent Commits)

| Commit | Relevance |
|---|---|
| `948654e undo and redo` | Direct predecessor; established UndoManager, blockPendingCommit, applyAction mutation pattern |
| `2388aa5 panic recovery and emergency save` | N/A |
| `a447e0e quit behaviors and save commands` | N/A |

The most relevant prior commit is `948654e`. Check `internal/editor/editor.go` InsertCharAction handler (around line 327–336) for the exact mutation pattern to follow.

### Project Structure Notes

**Modified files only — no new files:**
- `internal/block/gapbuffer.go` — add `PeekRight() (rune, bool)`
- `internal/block/gapbuffer_test.go` — add 2 PeekRight tests
- `internal/editor/editor.go` — add helpers `autoPairClosing` + `isAutoClosingChar`; modify `InsertCharAction` handler in `applyAction()`
- `internal/editor/editor_test.go` — add 12 auto-pair integration tests

**No changes to:**
- `internal/vim/insert.go` — InsertHandler stays stateless, no changes
- `internal/vim/action.go` — no new action types needed
- Any other file

### Alignment with Architecture FR Mapping

FR20 ("Auto-pairs for markdown characters") maps to:
- Primary: `vim/insert.go` per architecture table — however, since InsertHandler is stateless and doesn't have buffer access, the actual logic lives in `editor/actions.go` equivalent (editor.go's applyAction). This is correct per the delegation pattern.
- Supporting: `block/gapbuffer.go` (PeekRight)

[Source: architecture.md#Requirements to Structure Mapping — FR17-22 row] — "vim (insert mode), block (gap buffer), editor (clipboard)"

### References

- [Source: epics.md#Story 5.2] — Full story statement and 6 BDD acceptance criteria
- [Source: architecture.md#Vim Mode Architecture] — In-house implementation, per-mode handler pattern, EditorModel delegation
- [Source: architecture.md#Component Communication] — "Only EditorModel.applyAction() mutates top-level state"
- [Source: architecture.md#Package Boundary Rules] — block is leaf package, no cycles
- [Source: architecture.md#Testing Patterns] — Co-located tests, table-driven, TestFunctionName_Scenario_ExpectedBehavior
- [Source: internal/vim/insert.go] — InsertHandler (47 lines): stateless, returns InsertCharAction for all printable chars
- [Source: internal/vim/action.go] — Action types; InsertCharAction{Char rune}
- [Source: internal/block/gapbuffer.go] — GapBuffer implementation (274 lines): Insert, Backspace, Delete, MoveLeft, MoveRight, Content(), CursorPos(), gapEnd is the index of first char after cursor in buf
- [Source: internal/editor/editor.go:327–336] — InsertCharAction handler in applyAction: undo record → Insert → updateDisplay → autoSave
- [Source: internal/editor/editor.go:338–350] — BackspaceAction handler pattern to follow
- [Source: 5-1-undo-and-redo.md#Dev Notes] — Undo recording, blockPendingCommit, commitPendingBlock patterns from previous story

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

### Completion Notes List

### File List
