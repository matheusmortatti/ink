# Story 1.2: Markdown Block Parser

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->
<!-- Epic: View and Navigate a Beautiful Document -->
<!-- Story Key: 1-2-markdown-block-parser -->
<!-- Date: 2026-02-10 -->

## Story

As a writer,
I want my markdown document parsed into distinct blocks (paragraphs, headings, lists, code fences, block quotes, tables, horizontal rules),
so that each block can be independently rendered and edited.

## Acceptance Criteria

1. **Given** a markdown string containing paragraphs separated by blank lines **When** the parser processes the string **Then** each paragraph is returned as a separate `Block` with type `Paragraph` and the correct raw content

2. **Given** a markdown string containing headings (H1-H6), lists, code fences, block quotes, tables, and horizontal rules **When** the parser processes the string **Then** each element is returned as a `Block` with the correct type and raw content

3. **Given** a code fence containing blank lines **When** the parser processes the string **Then** the entire code fence (opening fence through closing fence) is a single `Block`

4. **Given** a parsed `[]Block` document **When** the blocks are serialized back to markdown (joined with `\n\n`) **Then** the output is identical to the original input for unmodified blocks

5. **Given** invalid or unusual markdown input **When** the parser processes it **Then** it returns blocks without crashing (NFR11)

6. **Given** a document of 10,000+ words **When** parsed into blocks **Then** parsing completes without perceptible delay (NFR7)

## Tasks / Subtasks

- [x] Task 1: Define Block type and BlockType enum (AC: #1, #2)
  - [x] 1.1 Create `Block` struct in `internal/block/block.go` with fields: `Type BlockType`, `Raw string`, `StartByte int`, `EndByte int`
  - [x] 1.2 Define `BlockType` enum: `Paragraph`, `Heading`, `List`, `CodeFence`, `BlockQuote`, `Table`, `HorizontalRule`, `CodeBlock`
  - [x] 1.3 Add heading level field to Block (or separate HeadingBlock type)
- [x] Task 2: Implement block parser using goldmark (AC: #1, #2, #3)
  - [x] 2.1 Create `Parse(source []byte) []Block` function in `internal/block/parser.go`
  - [x] 2.2 Configure goldmark with GFM extension (for tables) and parse source into AST
  - [x] 2.3 Walk goldmark AST, extracting top-level block nodes into `[]Block` with correct types
  - [x] 2.4 Extract raw source text for each block using goldmark's `text.Segment` byte ranges
  - [x] 2.5 Handle code fences with internal blank lines as single blocks (goldmark's `KindFencedCodeBlock`)
  - [x] 2.6 Handle block quotes as single blocks (goldmark's `KindBlockquote`)
  - [x] 2.7 Handle lists as single blocks (goldmark's `KindList`)
  - [x] 2.8 Handle tables as single blocks (GFM extension `KindTable`)
- [x] Task 3: Implement Document type with serialization (AC: #4)
  - [x] 3.1 Create `Document` type in `internal/block/document.go` wrapping `[]Block`
  - [x] 3.2 Implement `Serialize() string` that joins block raw content with `\n\n`
  - [x] 3.3 Ensure round-trip fidelity: `Parse(source).Serialize() == source` for well-formed markdown
- [x] Task 4: Handle edge cases and invalid input (AC: #5, #6)
  - [x] 4.1 Ensure parser never panics on any input (empty, binary, malformed markdown)
  - [x] 4.2 Handle empty input (returns empty `[]Block`)
  - [x] 4.3 Handle input with no block separators (single block)
  - [x] 4.4 Performance: verify 10,000+ word document parses without perceptible delay
- [x] Task 5: Write exhaustive tests (AC: #1-#6)
  - [x] 5.1 Create `internal/block/parser_test.go` with table-driven tests
  - [x] 5.2 Test every block type: paragraphs, H1-H6, ordered/unordered lists, fenced code blocks, indented code blocks, block quotes, tables, horizontal rules
  - [x] 5.3 Test code fence with internal blank lines
  - [x] 5.4 Test round-trip serialization for all block types
  - [x] 5.5 Test edge cases: empty input, single block, nested elements, unusual whitespace
  - [x] 5.6 Test invalid/malformed input never panics
  - [x] 5.7 Create `internal/block/document_test.go` with serialization tests

## Dev Notes

### Why This Story Matters

This is the **foundation story** for the entire ink editor. The block parser's output is consumed by:
- **Rendering pipeline** (`internal/render`) — each Block is independently rendered via Glamour
- **Viewport** (`internal/ui`) — composites rendered blocks for display
- **Editing block** (`internal/vim`) — scopes insert mode editing to a single block
- **Cursor mapper** (`internal/block/cursor.go`, Story 2.5) — maps positions between rendered and raw markdown
- **File I/O** (`internal/file`) — serializes blocks back to markdown for saving
- **Render cache** (`internal/render/cache.go`) — caches per-block Glamour output keyed by content hash

If the parser is wrong, **everything downstream is wrong**. This is why the architecture mandates "exhaustive block parser tests" — more tests here than anywhere else in the codebase.

### Critical Design Decisions

**1. Block Definition (from UX spec):**
A block is a paragraph-level markdown element separated by blank lines:
- A heading (single line, including `#` prefix) = one block
- A paragraph (may be multiple lines of text) = one block
- A list (all items until the next blank line) = one block
- A code fence (opening ` ``` ` through closing ` ``` `) = one block
- A blockquote (continuous `>` lines) = one block
- A table (header + separator + rows) = one block
- A horizontal rule (`---`, `***`, `___`) = one block

**2. Raw Content Preservation:**
Each `Block` stores the **original raw markdown source text** exactly as written. This is critical for:
- Round-trip fidelity (AC #4): `Parse(source).Serialize() == source`
- Editing: the gap buffer (Story 2.1) operates on this raw text
- Syntax dimming (Story 2.3): needs the raw markdown characters to dim

**3. goldmark for AST, NOT for raw text extraction:**
goldmark provides correct CommonMark-compliant block boundary detection. However, goldmark's AST nodes store `text.Segment` byte ranges into the source, NOT the raw text itself. The parser must:
- Use goldmark to determine WHERE blocks start and end in the source
- Extract raw text directly from the source `[]byte` using those byte ranges
- NOT rely on goldmark's text rendering, which strips markdown syntax

**4. Source Byte Range Strategy:**
goldmark nodes provide byte ranges via `node.Lines()` (for block nodes with text) and direct segment access. However, these ranges may not include the full raw source for container blocks (e.g., code fence delimiters, blockquote `>` markers). The parser must handle this by:
- For leaf blocks with `Lines()`: extracting from first line start to last line end
- For container blocks (blockquotes, lists): using child node ranges but expanding to include the container's own syntax markers
- For fenced code blocks: including the opening and closing fence lines
- **Fallback approach**: If goldmark's byte ranges are unreliable for raw extraction, use goldmark ONLY for block boundary detection and then slice the original source between consecutive block boundaries

### Technical Requirements

**Go version:** Go 1.25+ (current go.mod says 1.25.6)

**goldmark API patterns to use:**
```go
import (
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/ast"
    "github.com/yuin/goldmark/extension"
    "github.com/yuin/goldmark/text"
)

// Parse to AST (do NOT render to HTML)
md := goldmark.New(goldmark.WithExtensions(extension.GFM))
reader := text.NewReader(source)
doc := md.Parser().Parse(reader)

// Walk top-level children of Document node
ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
    if !entering || n.Kind() == ast.KindDocument {
        return ast.WalkContinue, nil
    }
    // Process top-level block node
    // Return ast.WalkSkipChildren to avoid descending into inline nodes
    return ast.WalkSkipChildren, nil
})
```

**goldmark block node kinds to handle:**

| goldmark Kind | ink BlockType | Notes |
|---|---|---|
| `ast.KindParagraph` | `Paragraph` | May span multiple lines |
| `ast.KindHeading` | `Heading` | Access `.Level` (1-6). Includes ATX (`#`) and setext headings |
| `ast.KindFencedCodeBlock` | `CodeFence` | Includes opening/closing fences and all content (even blank lines) |
| `ast.KindCodeBlock` | `CodeBlock` | Indented code blocks (4+ spaces) |
| `ast.KindBlockquote` | `BlockQuote` | Container block — children are nested blocks |
| `ast.KindList` | `List` | Container block — children are `KindListItem` |
| `ast.KindThematicBreak` | `HorizontalRule` | `---`, `***`, `___` |
| `extast.KindTable` (GFM) | `Table` | Requires `extension.GFM` or `extension.Table` |

**goldmark byte range extraction:**
```go
// Block nodes with lines (paragraphs, headings, code blocks)
lines := node.Lines()
if lines.Len() > 0 {
    firstLine := lines.At(0)
    lastLine := lines.At(lines.Len() - 1)
    startByte := firstLine.Start
    endByte := lastLine.Stop
    rawText := source[startByte:endByte]
}

// Container blocks (blockquotes, lists) - use HasChildren()
// Walk children to find min start and max end byte positions
```

**CRITICAL — goldmark `text.Segment` fields:**
- `Start` — start byte position in source (inclusive)
- `Stop` — end byte position in source (exclusive)
- `Padding` — internal padding (usually 0)
- `Value(source []byte) []byte` — extracts the segment text from source

### Architecture Compliance

**Package: `internal/block`** — This is a LEAF package with NO internal dependencies.

**Dependency direction (MUST follow):**
```
internal/block → (no internal dependencies)
```

- `internal/block` MUST NOT import any other `internal/` package
- `internal/block` CAN import standard library and external packages (goldmark, etc.)
- Other packages (`render`, `vim`, `ui`, `file`, `editor`) will import `internal/block`

**Naming conventions (enforce strictly):**
- Package name: `block` (singular, lowercase)
- No stutter: `block.Parse` not `block.BlockParse`, `block.Block` is fine
- Exported types/functions: `Block`, `BlockType`, `Document`, `Parse`
- Unexported helpers: `extractRaw`, `nodeToBlockType`, etc.
- Receiver names: single letter — `func (b *Block)`, `func (d *Document)`
- Error variables: `ErrEmpty` (if needed), lowercase messages, no trailing punctuation

**Anti-patterns to AVOID:**
- Do NOT import `internal/render`, `internal/vim`, or any other `internal/` package
- Do NOT create a `utils` or `helpers` package
- Do NOT add rendering logic — that belongs in `internal/render`
- Do NOT add gap buffer — that's a separate file `internal/block/gapbuffer.go` (Story 2.1)
- Do NOT add cursor mapping — that's `internal/block/cursor.go` (Story 2.5)
- Do NOT add any Bubbletea imports — this package has no TUI dependencies

### Library & Framework Requirements

| Library | Import Path | Version in go.mod | Usage in This Story |
|---|---|---|---|
| goldmark | `github.com/yuin/goldmark` | v1.7.16 (indirect) | AST parsing for block boundary detection |
| goldmark/ast | `github.com/yuin/goldmark/ast` | (same) | AST node types, Walk function, node kinds |
| goldmark/text | `github.com/yuin/goldmark/text` | (same) | `text.NewReader`, `text.Segment` for byte ranges |
| goldmark/extension | `github.com/yuin/goldmark/extension` | (same) | GFM extension for table support |
| goldmark/extension/ast | `github.com/yuin/goldmark/extension/ast` | (same) | `KindTable`, `KindTableHeader`, etc. |

**IMPORTANT:** goldmark is currently an `// indirect` dependency in `go.mod` (added in Story 1.1 but not yet imported). After this story, it becomes a direct dependency. Run `go mod tidy` after implementation to clean up.

**WARNING — Common LLM mistakes with goldmark:**
- Using `goldmark.Convert()` to render HTML instead of `md.Parser().Parse()` to get the AST — we need the AST, NOT HTML output
- Trying to extract raw text from goldmark's rendered output — we must extract from the original source bytes
- Forgetting to enable GFM extension for table support — tables are NOT in core CommonMark
- Assuming `node.Lines()` works for all block types — container blocks (blockquote, list) don't have direct lines
- Using `node.Text(source)` which only returns the first line's text — use `Lines()` iteration for multi-line blocks

### File Structure Requirements

**Files to create in this story:**

```
internal/block/
├── block.go          # Block type, BlockType enum (REPLACE existing placeholder)
├── parser.go         # Parse function: goldmark AST → []Block
├── parser_test.go    # Exhaustive parser tests (table-driven)
├── document.go       # Document type ([]Block wrapper), Serialize method
└── document_test.go  # Serialization/round-trip tests
```

**Files NOT to create:**
- `gapbuffer.go` — Story 2.1
- `gapbuffer_test.go` — Story 2.1
- `cursor.go` — Story 2.5
- `cursor_test.go` — Story 2.5

**Total files: 5** (1 replaced placeholder, 4 new)

**`block.go` content scope:**
```go
package block

// BlockType represents the type of a markdown block element.
type BlockType int

const (
    Paragraph BlockType = iota
    Heading
    List
    CodeFence
    CodeBlock
    BlockQuote
    Table
    HorizontalRule
)

// Block represents a single parsed markdown block element.
type Block struct {
    Type      BlockType
    Raw       string // Original markdown source text
    Level     int    // Heading level (1-6), 0 for non-headings
    StartByte int    // Start position in original source
    EndByte   int    // End position in original source (exclusive)
}
```

**`parser.go` content scope:**
```go
package block

// Parse takes raw markdown source and returns a slice of Blocks.
// Uses goldmark for CommonMark-compliant AST parsing, then extracts
// raw source text for each top-level block element.
func Parse(source []byte) []Block { ... }
```

**`document.go` content scope:**
```go
package block

// Document represents a parsed markdown document as a sequence of blocks.
type Document struct {
    Blocks []Block
    source []byte // Original source for reference
}

// NewDocument parses source markdown into a Document.
func NewDocument(source []byte) *Document { ... }

// Serialize reconstructs the markdown source from blocks.
// Joins block raw content with "\n\n" separators.
func (d *Document) Serialize() string { ... }
```

### Testing Requirements

**This story has the MOST tests of any story in the project.** The architecture mandates "exhaustive block parser tests" because every downstream component depends on correct parsing.

**Test location:** Co-located with source (Go convention)
- `internal/block/parser_test.go`
- `internal/block/document_test.go`

**Test naming:** `TestFunctionName_Scenario_ExpectedBehavior`
- Example: `TestParse_ParagraphsSeparatedByBlankLines_ReturnsSeparateBlocks`
- Example: `TestParse_CodeFenceWithBlankLines_SingleBlock`
- Example: `TestDocument_Serialize_RoundTrip`

**Test pattern:** Table-driven tests with `t.Run` subtests

```go
func TestParse_BlockTypes(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []BlockType
    }{
        {"single paragraph", "Hello world", []BlockType{Paragraph}},
        {"two paragraphs", "Hello\n\nWorld", []BlockType{Paragraph, Paragraph}},
        {"heading and paragraph", "# Title\n\nBody", []BlockType{Heading, Paragraph}},
        // ... more cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            blocks := Parse([]byte(tt.input))
            // assert block types match
        })
    }
}
```

**Test categories (all required):**

| Category | What to Test | Minimum Cases |
|---|---|---|
| Paragraph parsing | Single, multiple, multi-line paragraphs | 5+ |
| Heading parsing | H1-H6 ATX, setext H1/H2 | 8+ |
| List parsing | Ordered, unordered, nested, multi-item | 5+ |
| Code fence parsing | Basic, with language, with blank lines, with nested fences | 5+ |
| Indented code | Basic indented code block | 2+ |
| Block quote parsing | Single line, multi-line, nested | 3+ |
| Table parsing (GFM) | Basic table, alignment, multi-row | 3+ |
| Horizontal rule | `---`, `***`, `___` variants | 3+ |
| Mixed blocks | Document with multiple block types | 3+ |
| Edge cases | Empty input, single block, no separators, trailing newlines | 5+ |
| Invalid input | Binary data, malformed markdown, unclosed fences | 3+ |
| Round-trip | Parse then serialize produces identical output | 5+ |
| Performance | 10,000+ word document parses quickly | 1 |

**Testing tools:** Go's built-in `testing` package ONLY. No external test framework (testify, etc.) per architecture decision.

### Previous Story Intelligence

**From Story 1.1 (Project Initialization):**
- Bubbletea v2 `View()` returns `tea.View` not `string` — used `tea.NewView()` wrapper. Not relevant to this story but indicates v2 API differences from documentation.
- `go mod tidy` removes unused indirect dependencies. After this story imports goldmark, it will become a direct dependency and survive tidy.
- Actual dependency versions: goldmark v1.7.16 (newer than architecture's v1.7.12 reference)
- All placeholder files contain only `package <name>` — `internal/block/block.go` currently just has `package block`

**Learnings to apply:**
- Verify goldmark API against actual v1.7.16 behavior, not documentation assumptions
- The `internal/block/block.go` placeholder must be REPLACED entirely (not appended to)
- Run `go vet ./...` and `go build ./...` after implementation to verify

### Git Intelligence

**Recent commits:**
```
78e9544 initial folder structure and example main.go file
e8300bf epics and implementation readiness
```

**Patterns established:**
- Project root: `/home/matheusmortatti/git/ink/`
- Source code lives under `cmd/` and `internal/`
- Planning artifacts under `_bmad-output/`
- `.gitignore` updated to exclude `/ink` binary

### Latest Tech Information (Feb 2026)

**goldmark v1.7.16:**
- Latest stable (resolved in go.mod)
- CommonMark 0.31.2 compliant
- Stable AST API — no breaking changes expected
- GFM extension built-in for tables, strikethrough, linkify, task lists

**GFM Tables in goldmark:**
- Enable via `extension.GFM` or `extension.Table`
- Table AST nodes: `extast.KindTable`, `extast.KindTableHeader`, `extast.KindTableRow`, `extast.KindTableCell`
- Import: `github.com/yuin/goldmark/extension/ast` (aliased as `extast`)

**Go testing (Go 1.25+):**
- Loop variable scoping fixed since Go 1.22 — no need for `tt := tt` in table-driven tests
- `t.Run()` subtests for hierarchical organization
- `go test -v -run TestParse ./internal/block/...` for selective test execution

### Project Structure Notes

- `internal/block/` is a leaf package — no internal dependencies allowed
- This story adds the first real code to `internal/block/` (replacing the placeholder)
- No conflicts with existing code — only `block.go` placeholder exists
- File naming follows architecture: `block.go`, `parser.go`, `document.go` + test files

### References

- [Source: architecture.md#Markdown Parsing & Block Model] — goldmark + Block struct decision, rationale
- [Source: architecture.md#Document Data Structure] — `[]Block` slice decision
- [Source: architecture.md#Block & Document Conventions] — Serialization rules, round-trip requirement
- [Source: architecture.md#Package Boundary Rules] — `internal/block` is leaf, dependency direction
- [Source: architecture.md#Go Naming Patterns] — Naming conventions for this package
- [Source: architecture.md#Testing Patterns] — Co-located tests, table-driven, exhaustive parser tests
- [Source: architecture.md#Project Structure & Boundaries] — File list for `internal/block/`
- [Source: epics.md#Story 1.2] — Acceptance criteria, user story
- [Source: ux-design-specification.md#Experience Mechanics] — Block definition (paragraph-level elements)
- [Source: prd.md#FR1, FR4, FR9] — Document rendering, all markdown elements, block editing
- [Source: prd.md#NFR7] — 10,000+ word performance requirement
- [Source: prd.md#NFR11] — Invalid input must never crash

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Setext headings: goldmark Lines() only returns text content, not the underline (`===`/`---`). Fixed by adding `setextEnd()` to scan forward past the underline.
- Container blocks (blockquote, list): calling `Lines()` on inline descendant nodes panics. Fixed by checking `node.Type() == ast.TypeBlock` before calling Lines().
- Tables: goldmark's cell text segments don't include trailing `|` delimiters. Fixed by adding `lineEnd()` expansion for container blocks.
- Thematic breaks: have no Lines() and no text children. Fixed by using gap-based resolution between adjacent block ranges.

### Completion Notes List

- Ultimate context engine analysis completed — comprehensive developer guide created
- Story depends on: Story 1.1 (project initialization) — DONE
- Story blocks: Story 1.3 (Glamour rendering), Story 1.4 (viewport), Story 1.5 (file open), Story 2.1 (gap buffer), Story 2.5 (cursor mapping)
- This is the highest-test-density story in the project — exhaustive parser tests are architecturally mandated
- goldmark is used for AST parsing ONLY — raw text extraction comes from source byte slicing
- GFM extension required for table support (not in core CommonMark)
- Round-trip serialization fidelity is a hard requirement — `Parse(source).Serialize() == source`
- Implementation complete: Block type, BlockType enum, Parse function, Document type, Serialize method
- 55 test cases across 13 test functions covering all block types, edge cases, invalid input, round-trip, and performance
- Performance: ~0.6ms for 10,000+ word document (well within "no perceptible delay" requirement)
- goldmark promoted from indirect to direct dependency after `go mod tidy`
- All tests pass, no regressions, `go vet` clean

### File List

**Files modified:**
- `internal/block/block.go` (REPLACED placeholder — Block type, BlockType enum, String method)
- `go.mod` (goldmark promoted from indirect to direct dependency)
- `go.sum` (unchanged, already had goldmark)

**Files created:**
- `internal/block/parser.go` (Parse function, goldmark AST walking, byte range extraction)
- `internal/block/parser_test.go` (exhaustive parser tests — 55 test cases across 13 test functions)
- `internal/block/document.go` (Document type, NewDocument, Serialize method)
- `internal/block/document_test.go` (serialization/round-trip tests — 13 test cases)

## Change Log

- 2026-02-10: Implemented Story 1.2 — Markdown Block Parser. Created Block type with BlockType enum, Parse function using goldmark AST, Document type with round-trip Serialize method. 55+ test cases covering all markdown block types, edge cases, invalid input, and performance.
- 2026-02-10: Code review (AI) — 10 issues found (3 HIGH, 4 MEDIUM, 3 LOW). All HIGH and MEDIUM fixed:
  - H1: Fixed StartByte/EndByte inaccuracy for gap-resolved blocks (thematic breaks, empty code fences)
  - H2: Fixed round-trip failure for documents with trailing newlines (added suffix tracking to Document)
  - H3: Fixed trimBlockSeparators stripping meaningful trailing whitespace (spaces/tabs)
  - H3 (cont): Fixed goldmark trailing whitespace stripping by extending byte ranges in nodeByteRange
  - M1-M3: Added missing tests for empty code fences, byte position accuracy, trailing whitespace/newline round-trip
  - M4: Replaced unused Document.source field with Document.suffix for round-trip fidelity
  - L2: Fixed import order in parser.go
  - L3: Added BlockType.String() tests
