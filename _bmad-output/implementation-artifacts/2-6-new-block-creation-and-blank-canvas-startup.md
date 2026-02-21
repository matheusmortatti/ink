# Story 2.6: New Block Creation and Blank Canvas Startup

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->
<!-- Epic: 2 - Write with Live Preview -->
<!-- FRs: FR8, FR23 -->
<!-- Date: 2026-02-17 -->

## Story

As a writer,
I want to create new blocks while writing and start with a blank canvas when opening ink without a file,
so that I can begin writing immediately and grow my document naturally.

## Acceptance Criteria

1. **Given** the editor is in insert mode at the end of a block **When** the user presses Enter twice (creating a blank line) **Then** the current block is rendered, a new empty editing block is created below, and the cursor is in the new block in insert mode (FR8)

2. **Given** the user runs `ink` with no arguments **When** the application starts **Then** a blank canvas is displayed with the cursor at the top of the centered writing column in insert mode (FR23)

3. **Given** the user runs `ink` with no arguments **When** the blank canvas is displayed **Then** no content is visible except the cursor position — no welcome screen, no tips, no file browser

4. **Given** the user runs `ink existingfile.md` where the file has content **When** the application starts **Then** the document opens fully rendered in normal mode (FR24, context-aware startup)

5. **Given** the user runs `ink emptyfile.md` where the file exists but is empty **When** the application starts **Then** the application opens in insert mode (blank/empty content triggers insert mode)

6. **Given** a new block is created via double-Enter **When** the previous block renders **Then** the block split is instant and the user's typing flow is uninterrupted

## Tasks / Subtasks

- [x] Task 1: Implement double-Enter block splitting in editor (AC: #1, #6)
  - [x] 1.1 Detect consecutive newlines in `InsertNewlineAction` handler — when the gap buffer content ends with `\n` and another `\n` is inserted, trigger block split instead of inserting
  - [x] 1.2 Implement `splitActiveBlock()` in `internal/editor/editor.go` — extract content before the double newline as the completed block, render it, create a new empty block, and enter insert mode on the new block
  - [x] 1.3 Update `e.blocks` slice: modify the active block's Raw to the pre-split content, insert a new empty `block.Block{Type: Paragraph, Raw: ""}` at `activeBlockIdx + 1`
  - [x] 1.4 Invalidate render cache for the modified block, recompose viewport with the new block list
  - [x] 1.5 Set up gap buffer on the new empty block and position cursor at (0, 0) in insert mode
- [x] Task 2: Implement blank canvas startup (AC: #2, #3)
  - [x] 2.1 Modify `cmd/ink/main.go` to detect no-argument launch and pass a signal to the editor for blank canvas mode
  - [x] 2.2 Modify `NewEditor()` or add startup logic so when blocks are empty/nil, a single empty Paragraph block is created and insert mode is activated after viewport initialization
  - [x] 2.3 Ensure the blank canvas shows only the cursor at the top of the centered writing column — no placeholder text, no welcome screen
- [x] Task 3: Implement context-aware startup mode selection (AC: #4, #5)
  - [x] 3.1 In editor startup (after `initViewport`), check if blocks are empty or all blocks have empty Raw content — if so, auto-enter insert mode on the first block
  - [x] 3.2 Existing file with content: stay in normal mode (current behavior, already works)
  - [x] 3.3 Empty file (`ink emptyfile.md` where file exists but is empty): open in insert mode
- [x] Task 4: Comprehensive tests (AC: #1, #2, #3, #4, #5, #6)
  - [x] 4.1 Unit test `splitActiveBlock()`: verify block slice grows by one, original block Raw is trimmed, new block is empty Paragraph
  - [x] 4.2 Unit test double-Enter detection: verify single Enter inserts newline, double Enter at end triggers split
  - [x] 4.3 Unit test double-Enter mid-block: verify double-Enter only triggers split when at the end of block content (not mid-text)
  - [x] 4.4 Integration test blank canvas: verify NewEditor with nil/empty blocks + initViewport results in insert mode with a single empty block
  - [x] 4.5 Integration test context-aware startup: verify file with content opens in normal mode, empty file opens in insert mode
  - [x] 4.6 Integration test block split flow: enter insert mode, type text, press Enter twice, verify previous block rendered and new block active

## Dev Notes

### Architecture and Implementation Guidance

**This story has two distinct features:**
1. **Double-Enter block splitting** — while in insert mode, pressing Enter twice at the end of a block renders the current block and creates a new empty block below
2. **Blank canvas startup** — `ink` with no arguments opens in insert mode with a single empty block; context-aware startup selects mode based on content

**Files to modify:**
- `internal/editor/editor.go` — `applyAction()` for block split logic, new `splitActiveBlock()` method, startup mode selection in `initViewport()`
- `cmd/ink/main.go` — Pass blank canvas signal to editor (may need a `blankCanvas bool` field or detect from empty blocks)
- `internal/ui/viewport.go` — May need to handle single-block recompose after block insertion (existing `SetContent` and `composeBlocks` should work, but verify with empty blocks)

**Files to create:**
- None — all logic fits in existing files

**Package boundary:** All block split logic lives in `internal/editor` (the coordinator). Block creation uses `block.Block{Type: block.Paragraph, Raw: ""}` from the leaf `block` package. No new cross-package dependencies.

### Double-Enter Block Split — Detailed Design

**Detection logic in `applyAction()` `InsertNewlineAction` case:**

The current handler simply inserts `\n` into the gap buffer. The new logic must check:
1. Get the gap buffer content BEFORE inserting the newline
2. Check if the content ends with `\n` (meaning the cursor is at the end of a line that is itself at the end of the block, and the previous character was also a newline)
3. More precisely: check if the text before cursor ends with `\n` — this means the user just pressed Enter (which inserted `\n`), and now they're pressing Enter again

**Recommended approach:**
```go
case vim.InsertNewlineAction:
    if e.activeBuffer != nil {
        // Check for double-Enter at end of block
        content := e.activeBuffer.Content()
        cursorPos := e.activeBuffer.CursorPos()
        if cursorPos == len([]rune(content)) && strings.HasSuffix(content, "\n") {
            e.splitActiveBlock()
            return e, nil
        }
        e.activeBuffer.Insert('\n')
        e.updateActiveBlockDisplay()
    }
```

**Key constraint:** Double-Enter should ONLY trigger block split when the cursor is at the END of the block content. If the user presses Enter twice in the middle of text, it should just insert two newlines (normal editing behavior).

**`splitActiveBlock()` implementation:**
1. Get content from gap buffer, trim the trailing `\n` (the first Enter)
2. Update `e.blocks[e.activeBlockIdx].Raw` to the trimmed content
3. Invalidate render cache for the modified block: `e.cache.InvalidateBlock(e.blocks[e.activeBlockIdx])`
4. Create new block: `newBlock := block.Block{Type: block.Paragraph, Raw: ""}`
5. Insert into `e.blocks` slice at `e.activeBlockIdx + 1`
6. Clear the active block state: `e.viewport.ClearActiveBlock()`
7. Recompose viewport: `e.viewport.SetContent(e.blocks, e.renderer, e.cache)`
8. Enter insert mode on the new block: set `e.activeBlockIdx = oldIdx + 1`, create new gap buffer, call `e.viewport.SetActiveBlock()`
9. Set `e.modeHandler = vim.NewInsertHandler()` (already in insert mode, but ensure state is clean)

**Block slice insertion pattern (Go idiom):**
```go
idx := e.activeBlockIdx + 1
e.blocks = append(e.blocks, block.Block{}) // grow
copy(e.blocks[idx+1:], e.blocks[idx:])     // shift right
e.blocks[idx] = newBlock                    // insert
```

### Blank Canvas Startup — Detailed Design

**Current behavior (`cmd/ink/main.go`):**
- No arguments: `blocks` is nil, editor created with nil blocks
- File not found: `blocks` is nil (graceful fallback)
- File with content: `blocks` populated from parser

**What needs to change:**

The editor must detect empty/nil blocks after viewport initialization and auto-enter insert mode. The cleanest approach:

1. In `initViewport()`, after `e.viewport.SetContent(...)`, check if blocks are empty
2. If blocks are empty or nil, create a single empty Paragraph block, add it to `e.blocks`, recompose viewport, and call `enterInsertMode("i")`
3. This handles both `ink` (no args) and `ink emptyfile.md` (empty file parsed to nil/empty blocks)

**Important:** The startup mode selection must happen AFTER `initViewport()` because insert mode needs the viewport to be ready. Add the logic at the end of `initViewport()`:

```go
// Context-aware startup: blank/empty content → insert mode
if len(e.blocks) == 0 {
    e.blocks = []block.Block{{Type: block.Paragraph, Raw: ""}}
    _ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)
    e.enterInsertMode("i")
}
```

**For empty files:** `block.Parse("")` returns `nil` or empty slice (verify this). If it returns a single empty block, the check should also handle `len(e.blocks) == 1 && e.blocks[0].Raw == ""`.

### Current Code That Must Be Modified

**`editor.go` `applyAction()` — InsertNewlineAction case (line ~152-156):**
```go
// CURRENT: always inserts newline
case vim.InsertNewlineAction:
    if e.activeBuffer != nil {
        e.activeBuffer.Insert('\n')
        e.updateActiveBlockDisplay()
    }
```
Replace with double-Enter detection + split logic.

**`editor.go` `initViewport()` (line ~598-624):**
Add blank canvas / context-aware startup at the end, after `e.ready = true`.

### Render Cache Interaction

When splitting a block:
- The OLD block content changes → invalidate its cache entry via `e.cache.InvalidateBlock(oldBlock)`
- The NEW empty block has no cache entry yet — it will be rendered on demand when `ClearActiveBlock()` or `SetContent()` is called
- After split, recompose entire viewport via `SetContent()` to rebuild `blockRanges` for all blocks

### Viewport Recomposition After Split

After inserting a new block into `e.blocks`, the viewport's `blockRanges` is stale. You MUST call `e.viewport.SetContent(e.blocks, e.renderer, e.cache)` to recompose. Then immediately call `e.viewport.SetActiveBlock(newBlockIdx, "")` to set the new block as active for editing.

### Edge Cases to Handle

1. **Double-Enter on first block with no content** — Should create a second block. The first block stays empty (or is it removed?). Per UX spec: the current block is rendered (even if empty — renders to nothing) and a new block is created below.
2. **Double-Enter mid-text** — Should NOT trigger block split. Only split when cursor is at the absolute end of block content AND content ends with `\n`.
3. **Double-Enter on the only block** — Should work: the document grows from 1 block to 2 blocks.
4. **Rapid double-Enter** — Should create multiple blocks in sequence (each split creates a new empty block, another double-Enter on that empty block would... create another empty block? Need to handle: empty block content is `""`, which does NOT end with `\n`, so pressing Enter once inserts `\n`, pressing Enter again — content is `"\n"` which ends with `\n` and cursor is at end → split. This is correct behavior.)
5. **Blank canvas with `ink nonexistent.md`** — Current code already sets blocks to nil for missing files. The startup logic should create the empty block and enter insert mode. File path should be preserved so auto-save knows the target.

### Performance

- Block split should be imperceptible — no Glamour re-rendering needed for the OLD block if it hasn't changed since last render (cache hit). The NEW empty block is trivial to render.
- `SetContent()` recomposes all blocks, but this is fast for typical document sizes (50-200 blocks per architecture doc).
- NFR2 (<50ms block transitions) should be easily met.

### Project Structure Notes

- No new files or packages — all changes in existing `editor.go` and `main.go`
- Follows EditorModel delegation pattern: editor coordinates block manipulation, viewport handles display
- Block creation uses leaf `block` package types only
- Tests co-located: `editor_test.go` for integration tests

### References

- [Source: epics.md#Story 2.6] — Acceptance criteria and user story
- [Source: architecture.md#Document Data Structure] — `[]Block` slice, block insertion
- [Source: architecture.md#Block Serialization] — Blocks joined with `\n\n`, round-trip preservation
- [Source: architecture.md#Render Cache Lifecycle] — Cache invalidation per-block on content change
- [Source: architecture.md#Component Communication] — EditorModel delegation pattern
- [Source: ux-design-specification.md#Creating New Blocks] — Double-Enter splits block, renders previous, cursor in new block
- [Source: ux-design-specification.md#Blank Canvas Layout] — Cursor at top of writing column, no centering vertically
- [Source: ux-design-specification.md#Design Direction Decision] — Startup Mode Logic: blank/new → insert mode, existing → normal mode
- [Source: prd.md#FR8] — Enter twice creates new block
- [Source: prd.md#FR23] — Blank canvas in insert mode
- [Source: prd.md#FR24] — Open file in normal mode
- [Source: 2-5-cursor-position-mapping-between-rendered-and-raw.md] — Previous story dev notes, enterInsertMode/exitInsertMode patterns

### Previous Story Intelligence (from Story 2.5)

**Key patterns established:**
- `enterInsertMode(variant)` handles all insert variants (`i`, `a`, `o`, `O`) with cursor mapping via `MapRenderedToRaw`. Story 2.6 must NOT break this — the `splitActiveBlock()` method bypasses `enterInsertMode()` for the new block since we're already in insert mode and just need to set up a fresh gap buffer at (0,0).
- `exitInsertMode()` uses `MapRawToRendered` and conditional cache invalidation (`newContent != blocks[activeBlockIdx].Raw`). The split operation changes block content BEFORE calling `ClearActiveBlock()`, so cache invalidation must happen explicitly.
- `updateActiveBlockDisplay()` calls `viewport.UpdateActiveBlockContent()` which does optimized partial recompose. After a block split, a FULL recompose via `SetContent()` is needed instead because `blockRanges` must be rebuilt for the new block count.

**Code review fixes from 2.5 that inform 2.6:**
- Table separator row handling was missing — reminder to handle edge cases in block type when creating new blocks (always use `block.Paragraph` for new empty blocks, never infer type).
- Nested blockquote prefix handling was incomplete — not directly relevant but reinforces: test edge cases thoroughly.

**Test patterns from 2.5:**
- Table-driven tests with descriptive names: `TestSplitActiveBlock_AtEndOfParagraph`, `TestBlankCanvasStartup_NoArgs`
- Integration tests that simulate full enter→edit→action→verify cycle using `NewEditor` + `initViewport` via `WindowSizeMsg`
- All 7 packages must pass after changes

### Git Intelligence Summary

**Recent commits (last 10):**
```
c827492 cursor position mapping between rendered and raw
0e9ef22 fix hang in code when moving to insert mode
d184dcd block transitions the mode unified block reveal
9012550 syntax dimming for active editing block
6c2ce84 insert mode and text input
de5bd1b gap buffer
8edff57 basic vim motion navigation
bbd9172 open and display existing markdown file
dbeea60 document viewport with centered writing column
c1d62f3 block rendering
```

**Patterns observed:**
- Commit messages are lowercase, concise, describing the feature
- Each story is typically a single commit
- Most recent commit (`c827492`) added `internal/block/cursor.go`, `internal/block/cursor_test.go` and modified `internal/editor/editor.go` + `internal/editor/editor_test.go`
- The codebase has grown organically with each story building on the previous

**Files touched in recent work relevant to Story 2.6:**
- `internal/editor/editor.go` — The main file to modify. Last changed in Story 2.5 (cursor mapping integration). Lines ~152-156 (InsertNewlineAction) and ~598-624 (initViewport) are the primary modification targets.
- `internal/editor/editor_test.go` — Add new tests here. Last expanded in Story 2.5 with cursor mapping integration tests.
- `internal/ui/viewport.go` — Provides `SetContent()`, `SetActiveBlock()`, `ClearActiveBlock()`. Should NOT need modification for Story 2.6 — existing API is sufficient.
- `internal/block/block.go` — Block type definition. No changes needed — just create `block.Block{Type: block.Paragraph, Raw: ""}`.

### Testing Requirements

**Test file:** `internal/editor/editor_test.go` (co-located, extend existing)

**Test naming convention:** `TestFunctionName_Scenario_ExpectedBehavior`

**Required tests:**

| Test Name | Type | What it Verifies |
|-----------|------|-----------------|
| `TestSplitActiveBlock_AtEndOfParagraph` | Unit | Block slice grows by 1, original trimmed, new block empty Paragraph |
| `TestSplitActiveBlock_CursorInNewBlock` | Unit | After split, activeBlockIdx points to new block, gap buffer is empty |
| `TestSplitActiveBlock_CacheInvalidated` | Unit | Old block's cache entry invalidated after content change |
| `TestDoubleEnter_AtEndTriggersSplit` | Integration | Enter + Enter at block end → split happens |
| `TestDoubleEnter_MidTextNoSplit` | Integration | Enter + Enter mid-text → two newlines inserted, no split |
| `TestDoubleEnter_SingleEnterNoSplit` | Integration | Single Enter → newline inserted, no split |
| `TestBlankCanvas_NoArgs` | Integration | Empty blocks → insert mode, single empty block, cursor at (0,0) |
| `TestBlankCanvas_EmptyFile` | Integration | File exists but empty → insert mode |
| `TestStartup_FileWithContent_NormalMode` | Integration | File with content → normal mode (existing behavior) |
| `TestSplitActiveBlock_EmptyBlock` | Edge case | Split on empty block with just `\n` → creates second empty block |
| `TestSplitActiveBlock_MultipleRapidSplits` | Edge case | Three consecutive splits → 4 blocks total |

**Integration test pattern (from Story 2.5):**
```go
func TestBlankCanvas_NoArgs(t *testing.T) {
    e := editor.NewEditor("", nil)
    // Simulate WindowSizeMsg to trigger initViewport
    model, _ := e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
    ed := model.(*editor.EditorModel)
    // Verify insert mode and single empty block
    if ed.CurrentMode() != vim.Insert { t.Error("expected insert mode") }
}
```

### Latest Technical Context

- **Bubbletea v2 RC:** No breaking changes since Story 2.5. `tea.KeyPressMsg`, `tea.WindowSizeMsg`, `tea.NewView` APIs stable.
- **Glamour:** Rendering an empty string (`""`) returns an empty string — safe for the new empty block. No special handling needed.
- **Go 1.25:** Slice insertion idiom (`append` + `copy`) is the standard approach. No generics-based slice insert needed for this simple case.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

None — implementation proceeded cleanly.

### Completion Notes List

- Implemented `splitActiveBlock()` in `editor.go`: trims trailing `\n` from the active buffer, updates the current block's Raw, invalidates cache, inserts a new empty `block.Paragraph` at `activeBlockIdx+1` using the Go append+copy idiom, calls `SetContent()` to rebuild `blockRanges`, then enters insert mode on the new block.
- Modified `InsertNewlineAction` handler to detect double-Enter (cursor at end of content that ends with `\n`) and delegate to `splitActiveBlock()` instead of inserting.
- Added `isContentEmpty()` helper to check if all blocks have empty Raw content.
- Added blank canvas startup logic at end of `initViewport()`: if content is empty, creates a single empty Paragraph block, calls `SetContent()`, then enters insert mode directly (not via `enterInsertMode()` to avoid cursor mapping overhead on an empty block).
- `main.go` required no changes — it already passes nil blocks for no-args and missing-file cases; the editor handles the rest.
- All 7 packages pass. 11 new tests added covering: block split mechanics, cache invalidation, double-Enter detection (end vs mid-text vs single), blank canvas (no args, empty file), context-aware startup (file with content stays normal), edge cases (empty block split, rapid multiple splits), and post-split typing.

### File List

- `internal/editor/editor.go` (modified)
- `internal/editor/editor_test.go` (modified)

### Review Fixes Applied

- `editor.go:367` — `strings.TrimRight` → `strings.TrimSuffix` (precision: removes exactly one trailing `\n`)
- `editor.go:615–617` — Removed stale TODO/debug comment in `clampCursor()`
- `editor.go:678–688` — Added explanatory comment for blank canvas init bypassing `enterInsertMode()`, added `ensureInsertCursorVisible()` call
- `editor_test.go:650–670` — Fixed `TestEditorModel_EmptyDocument_NoNavigationCrash`: exits insert mode before testing Normal-mode navigation (Story 2.6 changed empty docs to start in insert mode)
- `editor_test.go:1843–1856` — Fixed `TestBlankCanvas_EmptyFile`: uses `block.Parse([]byte(""))` to match real `main.go` code path
- `editor_test.go` — Added `TestSplitActiveBlock_Performance` to validate AC6 NFR2 <50ms for block split

## Change Log

- 2026-02-18: Story 2.6 implemented — double-Enter block splitting, blank canvas startup, context-aware mode selection. 11 new tests added. All tests pass.
