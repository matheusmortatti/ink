# Story 2.5: Cursor Position Mapping Between Rendered and Raw

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a writer,
I want my cursor position to map accurately between rendered and raw markdown,
so that when I enter a block to edit, my cursor lands exactly where I expect.

## Acceptance Criteria

1. **Given** the cursor is on a word in a rendered paragraph **When** the user enters insert mode **Then** the cursor is positioned at the same word in the raw markdown text (FR10)

2. **Given** the cursor is on rendered heading text (e.g., "My Title") **When** the user enters insert mode **Then** the cursor is positioned at the corresponding text after the `#` characters in the raw markdown (FR10)

3. **Given** the cursor is on rendered bold text **When** the user enters insert mode **Then** the cursor is positioned within the `**...**` markers at the corresponding character (FR10)

4. **Given** the user exits a block with `Esc` from a position in raw markdown **When** the block renders and normal mode activates **Then** the cursor maps back to the corresponding position in the rendered output (FR10)

5. **Given** any markdown element type (paragraph, heading, list, code fence, block quote, table) **When** cursor mapping is performed in either direction **Then** the mapping is consistent and accurate for that element type

## Tasks / Subtasks

- [x] Task 1: Create cursor mapping types and core functions (AC: #1, #2, #3, #4, #5)
  - [x] 1.1 Create `internal/block/cursor.go` with `MapRenderedToRaw(block Block, renderedLine, renderedCol int) (rawLine, rawCol int)` and `MapRawToRendered(block Block, rawLine, rawCol int) (renderedLine, renderedCol int)`
  - [x] 1.2 Implement paragraph mapping (1:1 — no syntax markers to offset)
  - [x] 1.3 Implement heading mapping (offset by `#` count + space prefix)
  - [x] 1.4 Implement inline marker mapping for bold `**`, italic `*`/`_`, inline code `` ` ``, links `[text](url)`
  - [x] 1.5 Implement list item mapping (Glamour adds bullet/number styling, raw has `- ` or `1. ` prefixes)
  - [x] 1.6 Implement code fence mapping (raw includes ``` ``` ``` delimiters; rendered omits them)
  - [x] 1.7 Implement block quote mapping (raw has `> ` prefix per line)
  - [x] 1.8 Implement table mapping (raw has `|` delimiters; rendered has Glamour table styling)
- [x] Task 2: Comprehensive table-driven tests (AC: #1, #2, #3, #4, #5)
  - [x] 2.1 Create `internal/block/cursor_test.go` with tests for every block type
  - [x] 2.2 Test bidirectional consistency: `MapRenderedToRaw` then `MapRawToRendered` returns original position
  - [x] 2.3 Test edge cases: cursor at end of line, empty blocks, single-char blocks, cursor beyond content length
  - [x] 2.4 Test inline markers: nested bold+italic, adjacent markers, markers at line start/end
- [x] Task 3: Integrate mapping into editor mode transitions (AC: #1, #2, #3, #4)
  - [x] 3.1 Modify `enterInsertMode()` in `internal/editor/editor.go` to call `MapRenderedToRaw` before `SetCursorLineCol` — replace the current simplified `(0, 0)` placeholder
  - [x] 3.2 Modify `exitInsertMode()` to call `MapRawToRendered` and set `cursorLine`/`cursorCol` to the mapped position instead of block start
  - [x] 3.3 Handle all insert variants (`i`, `a`, `o`, `O`) correctly with mapped positions
- [x] Task 4: Integration tests for full transition cycle (AC: #1, #4)
  - [x] 4.1 Add editor-level tests: enter block at rendered position → verify gap buffer cursor at correct raw position
  - [x] 4.2 Add editor-level tests: exit block from raw position → verify document cursor at correct rendered position
  - [x] 4.3 Test round-trip: normal → insert → type nothing → Esc → cursor returns to same rendered position

## Dev Notes

### Architecture and Implementation Guidance

**New files to create:**
- `internal/block/cursor.go` — Cursor position types, `MapRenderedToRaw`, `MapRawToRendered`
- `internal/block/cursor_test.go` — Table-driven tests across all markdown element types

**Files to modify:**
- `internal/editor/editor.go` — `enterInsertMode()` (lines ~270-309) and `exitInsertMode()` (lines ~311-341)

**Package boundary:** `internal/block` is a leaf package with no internal dependencies. The mapping functions receive a `Block` struct and position coordinates — they do NOT depend on editor, viewport, or renderer.

### Current Simplified Mapping (TO BE REPLACED)

The current `enterInsertMode()` always sets cursor to `(0, 0)` regardless of where the cursor was in normal mode. The current `exitInsertMode()` always returns cursor to the block's start line with `cursorCol = 0`. This was intentionally deferred from Story 2.2.

**Current code in `editor.go` `enterInsertMode()`:**
```go
// REPLACE THIS: always places at (0, 0) — Story 2.5 maps accurately
e.activeBuffer.SetCursorLineCol(0, 0)
```

**Current code in `editor.go` `exitInsertMode()`:**
```go
// REPLACE THIS: always returns to block start — Story 2.5 maps back accurately
blockStart := e.viewport.BlockStartLine(e.activeBlockIdx)
e.cursorLine = blockStart
e.cursorCol = 0
```

### Coordinate Systems

- **Document cursor (normal mode):** `(cursorLine, cursorCol)` — line index in composed viewport `lines` array, col is visual offset after left margin
- **Block cursor (insert mode):** `(line, col)` in gap buffer — zero-indexed relative to block's raw content
- **Mapping input:** The rendered position must be computed relative to the block's start in the viewport, not as absolute screen coordinates. Use `viewport.BlockStartLine(blockIdx)` to get the block's start line, then subtract to get block-relative rendered position.

### Rendering Transform Reference

Each block type has different transforms between raw and rendered:

| Block Type | Raw Example | Rendered (Glamour) | Mapping Strategy |
|-----------|-------------|-------------------|-----------------|
| Paragraph | `hello world` | `hello world` | 1:1 (no offset) |
| Heading | `## My Title` | `My Title` (styled) | Skip `## ` prefix (level+1 chars) |
| Bold | `**bold text**` | `bold text` (styled) | Track `**` marker offsets inline |
| Italic | `*italic*` | `italic` (styled) | Track `*` marker offsets inline |
| Link | `[text](url)` | `text` (styled) | Map within `[text]` portion, skip `](url)` |
| List | `- item` | `  * item` (Glamour) | Map prefix differences per line |
| Code fence | ` ```lang\ncode\n``` ` | Highlighted code block | Skip delimiter lines, map content lines |
| Block quote | `> text` | Styled quote | Skip `> ` prefix per line |
| Table | `\| a \| b \|` | Styled table | Map column positions accounting for Glamour table styling |

### Glamour Rendering Considerations

- Glamour output is ANSI-styled — positions in rendered output must account for ANSI escape sequences being invisible characters
- The viewport's `wrapLine()` function is ANSI-aware and handles wrapping correctly
- Glamour may add trailing newlines or modify whitespace — the mapper should work with content characters, not raw byte positions
- For the rendered→raw direction: strip ANSI from rendered content to get visual character positions, then map to raw

### Performance

- Cursor mapping runs during mode transitions (enter/exit insert mode)
- Current transition time is <0.12ms (measured in Story 2.4 benchmarks) — ample budget
- NFR2 requires <50ms for block transitions — mapping must stay within this budget
- Mapping should be O(line_length) per line, not O(document)

### Previous Story Intelligence

**From Story 2.4 (Block Transitions):**
- Cache invalidation is conditional: only invalidates when content actually changed (`newContent != blocks[activeBlockIdx].Raw`)
- `blockRanges` tracks where each block lives in the composed `lines` array — use this for block-relative position calculation
- `viewport.BufferToScreenPos(bufLine, bufCol)` converts gap buffer position to screen position — this exists but handles display, not mapping
- All insert mode variants (`i`, `a`, `o`, `O`) must be tested
- Use `viewport.ColumnWidth()` getter for dynamic width, not hardcoded values

**From Story 2.3 (Syntax Dimming):**
- `render.DimSyntax()` already parses inline markdown syntax markers — consider reusing its marker detection logic for cursor offset calculation
- ANSI-aware `wrapLine()` handles styled content correctly

### Testing Strategy

- **Table-driven tests** for each block type with multiple position scenarios
- **Bidirectional consistency tests**: `MapRawToRendered(MapRenderedToRaw(pos)) == pos`
- **Edge cases**: cursor at (0,0), cursor at end of content, cursor on blank line, cursor beyond content
- **Integration tests** in `editor_test.go` for full enter→edit→exit cycle
- **Naming convention**: `TestMapRenderedToRaw_Heading_CursorOnTitle`, `TestMapRawToRendered_Bold_CursorInsideMarkers`

### Project Structure Notes

- Alignment: New files in `internal/block/` follow existing package conventions (`block.go`, `parser.go`, `gapbuffer.go`)
- No new packages needed — mapper belongs in the leaf `block` package
- No new dependencies — uses goldmark AST indirectly through existing Block struct data

### References

- [Source: architecture.md#Cursor Position Representation] — lines 341-344
- [Source: architecture.md#File Structure] — lines 419-420 (`cursor.go`, `cursor_test.go`)
- [Source: epics.md#Story 2.5] — Acceptance criteria and user story
- [Source: ux-design-specification.md#Cursor Position Mapping] — lines 977-982
- [Source: prd.md#FR10] — Functional requirement for cursor mapping
- [Source: prd.md#Critical Success Factor] — "Cursor position mapping accuracy — Build as a core primitive with exhaustive tests"
- [Source: 2-4-block-transitions-the-mode-unified-block-reveal.md] — Previous story dev notes and learnings

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Fixed rawColFromVisualCol logic: initial implementation returned raw position 0 for visual position 0 even when position 0 was a hidden marker. Fixed to skip hidden markers before matching visual positions.

### Completion Notes List

- Created `internal/block/cursor.go` with bidirectional cursor mapping between rendered (Glamour) and raw markdown positions
- Core approach: `classifyRunes()` identifies inline markers (bold `**`, italic `*`/`_`, code `` ` ``, links `[text](url)`, images `![alt](url)`, escapes `\`) as hidden characters, then `rawColFromVisualCol` and `visualColFromRawCol` convert between visual and raw positions
- Handles nested markers (e.g., bold containing italic) via recursive `classifyRange()`
- Block-type-specific handling: heading prefix (`## ` = level+1 chars), code fence delimiter lines, block quote `> ` prefix, list marker prefix (`- `, `1. `, etc.), table pipe delimiters
- Modified `enterInsertMode()` to compute block-relative rendered position, call `MapRenderedToRaw`, and position gap buffer cursor at the mapped raw position for all variants (`i`, `a`, `o`, `O`)
- Modified `exitInsertMode()` to capture raw cursor position, call `MapRawToRendered`, and set document cursor at blockStart + mapped rendered position
- All 7 packages pass tests (48+ cursor-specific tests covering every block type, bidirectional consistency, edge cases, and integration tests)

### Change Log

- 2026-02-14: Implemented cursor position mapping between rendered and raw markdown (Story 2.5)
- 2026-02-17: Code review fixes — 6 issues resolved (2 HIGH, 4 MEDIUM)

### Senior Developer Review (AI)

**Reviewer:** Claude Opus 4.6 (adversarial code review)
**Date:** 2026-02-17
**Outcome:** Approved after fixes

**Issues Found & Fixed (6):**
- **[H1] Fixed:** `classifyRunes` mishandled `***bold italic***` triple-asterisk syntax — added triple-delimiter detection before double-delimiter check
- **[H2] Fixed:** Table mapping didn't skip separator rows (`| --- | --- |`) — rewrote `mapRenderedToRawTable`/`mapRawToRenderedTable` with `isTableSeparator` helper
- **[M1] Fixed:** `blockQuotePrefixLen` only handled single-level quotes — rewritten to loop through nested `> > ` prefixes
- **[M2] Fixed:** Missing `TestMapRawToRendered_Table` bidirectional test — added with header, separator, and data row cases
- **[M3] Fixed:** Image marker `![alt](url)` classification had no test coverage — added `TestClassifyRunes_Image`
- **[M4] Fixed:** `findClosingDelim` didn't handle double-escaped backslashes (`\\*`) — added consecutive backslash counting

**Additional tests added:** `TestClassifyRunes_TripleAsterisk`, `TestClassifyRunes_UnmatchedBold`, `TestClassifyRunes_EscapedBackslash`, nested blockquote prefix cases

**All 7 packages pass (0 failures).**

### File List

- `internal/block/cursor.go` (new, review-modified) — MapRenderedToRaw, MapRawToRendered, classifyRunes, inline marker handling, table separator detection
- `internal/block/cursor_test.go` (new, review-modified) — 50+ tests covering all block types, bidirectional consistency, edge cases
- `internal/editor/editor.go` (modified) — enterInsertMode() and exitInsertMode() now use cursor mapping
- `internal/editor/editor_test.go` (modified) — 5 new integration tests for cursor mapping transitions
