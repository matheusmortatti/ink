---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
lastStep: 8
status: 'complete'
completedAt: '2026-02-09'
inputDocuments:
  - prd.md
  - prd-validation-report.md
  - product-brief-ink-2026-02-07.md
  - ux-design-specification.md
workflowType: 'architecture'
project_name: 'ink'
user_name: 'Matheusmortatti'
date: '2026-02-09'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements:**
47 FRs across 9 categories. The requirements describe a full-screen TUI markdown editor with Obsidian-style inline live preview, vim modal editing, and a writer-centric environment. Architecturally, the FRs break into three tiers:
- **Core rendering pipeline** (FR1-11): Block parsing, Glamour rendering, block transitions, cursor position mapping — the foundation everything else sits on
- **Input/interaction layer** (FR12-22, FR40-42): Vim mode system, text editing, auto-pairs, clipboard, mouse support — the user-facing interaction model
- **Application shell** (FR23-39, FR43-47): File I/O, auto-save, status bar, terminal adaptation, configuration — the operational wrapper

**Non-Functional Requirements:**
21 NFRs across 4 categories driving architectural decisions:
- **Performance** (NFR1-7): Startup <100ms, block transitions <50ms, imperceptible keystroke latency, smooth scrolling, 10,000+ word document support. These constrain rendering pipeline design, caching strategy, and Glamour integration approach.
- **Reliability** (NFR8-11): Atomic file writes, panic recovery with emergency save, never-crash-on-invalid-input. These drive file I/O architecture and error handling strategy.
- **Accessibility** (NFR12-16): Minimum contrast for dimmed elements, no color-only signaling, keyboard-complete operation, screen reader consideration. These constrain the color system and UI component design.
- **Compatibility** (NFR17-21): 6+ terminal emulators, tmux/screen, true color/256/16 color support, SSH sessions, Linux/macOS/Windows. These constrain rendering choices and require graceful degradation paths.

**Scale & Complexity:**

- Primary domain: TUI application (Go/Bubbletea)
- Complexity level: Low-to-Medium
- Estimated architectural components: 8-10 (block parser, Glamour renderer/cache, document viewport, editing block, rendered block, vim mode manager, status bar, file manager, config loader, optional focus mode overlay)

### Technical Constraints & Dependencies

- **Language & framework:** Go with Bubbletea (Model-Update-View, event-driven), Lip Gloss (styling), Glamour (markdown rendering)
- **Distribution:** Single binary, no runtime dependencies. Cross-compiled for Linux, macOS, Windows.
- **Rendering engine:** Glamour (designed for full-document rendering) must be adapted for block-level compositing — this is an unproven usage pattern that carries technical risk.
- **Terminal constraints:** Monospace character grid, ANSI escape sequences, variable color depth support, terminal-native cursor.
- **File format:** Markdown (.md) exclusively. No export, no conversion.
- **Configuration:** YAML at `~/.config/ink/config.yml` (XDG), optional, zero-config defaults.

### Cross-Cutting Concerns Identified

1. **Block awareness** — The block parser's output is consumed by the viewport (composition), editing block (scope), rendered block (content), cursor mapper (position translation), and auto-save (content reconstruction). Changes to block structure ripple through all components.
2. **Vim mode state** — The current mode (normal/insert/visual/command) affects rendering behavior, input handling, status bar display, block transition triggers, and UI visibility. Mode state is the central coordination mechanism.
3. **Terminal adaptation** — Resize events trigger column width recalculation, Glamour re-rendering of all visible blocks, layout reflow, and scroll position adjustment. This cuts across viewport, rendering, and layout logic.
4. **Color interpolation** — Dimming logic (status bar ~30%, fade mode ~20%, syntax chars ~40%) uses a shared calculation pattern (interpolate foreground toward background). Should be a single utility, not reimplemented per component.
5. **Performance budget** — The <50ms block transition target constrains how rendering, layout, and display are orchestrated. Caching, pre-rendering, and lazy evaluation strategies must be considered architecturally, not bolted on.

## Starter Template Evaluation

### Primary Technology Domain

Go TUI application (Bubbletea/Charm ecosystem) — established by PRD technical requirements.

### Starter Options Considered

**Option 1: Bubbletea v2 (RC) + Manual Go Project Layout**
The Charm ecosystem is transitioning to v2 across all libraries. Bubbletea v2.0.0-rc.2 is near final release with a new Cursed Renderer, split mouse messages, and the `charm.land` module path. Starting on v2 for a greenfield project avoids a future migration and gains performance improvements from the new renderer.

**Option 2: Bubbletea v1 (Stable) + Manual Go Project Layout**
The stable, battle-tested path. However, v1 will become legacy once v2 releases (imminent based on RC status). Starting on v1 means a migration cost within months.

**No formal scaffold tools exist** for Bubbletea projects — both options use `go mod init` + standard Go project structure.

### Selected Approach: Bubbletea v2 + Standard Go Project Layout

**Rationale:**
- Greenfield project — no legacy code to migrate
- v2 RC stage indicates API stability (breaking changes unlikely)
- New Cursed Renderer directly benefits ink's <50ms block transition performance target
- Split mouse messages provide cleaner architecture for ink's mouse support
- Lip Gloss v2's deterministic styles and precise I/O control benefit ink's dimming system
- Avoids certain migration cost when v2 reaches stable (likely within weeks/months)

**Initialization Commands:**

```bash
mkdir ink && cd ink
go mod init github.com/matheusmortatti/ink
go get charm.land/bubbletea/v2@latest
go get github.com/charmbracelet/lipgloss/v2@latest
go get github.com/charmbracelet/bubbles/v2@latest
go get github.com/charmbracelet/glamour@latest
go get gopkg.in/yaml.v3
```

**Architectural Decisions Provided by This Approach:**

**Language & Runtime:**
Go 1.25.x with modules. Single binary compilation. Cross-platform via `GOOS`/`GOARCH`.

**Project Structure:**

```
ink/
├── cmd/ink/
│   └── main.go              # Entry point, CLI arg parsing
├── internal/
│   ├── editor/              # Core editor model (top-level Bubbletea model)
│   ├── block/               # Block parser and block types
│   ├── render/              # Glamour rendering, caching, block compositing
│   ├── vim/                 # Vim mode system (normal, insert, visual, command)
│   ├── ui/                  # UI components (status bar, viewport, save prompt)
│   ├── file/                # File I/O, auto-save, atomic writes
│   └── config/              # YAML config loading, defaults, CLI overrides
├── go.mod
├── go.sum
└── LICENSE
```

**Build Tooling:**
`go build ./cmd/ink` for development. `goreleaser` for cross-platform release builds and distribution (Homebrew, AUR, GitHub Releases).

**Testing Framework:**
Go's built-in `testing` package. No external test framework needed for a project of this scale.

**Code Organization:**
`internal/` enforces package privacy at the compiler level — all ink-specific code is non-importable. `cmd/ink/` keeps the entry point thin. Each `internal/` package maps to an architectural component identified in the context analysis.

**Development Experience:**
`go run ./cmd/ink` for rapid iteration. Go's fast compilation (sub-second for projects this size) serves as the "hot reload" equivalent.

**Risk: Bubbletea v2 RC Stability**
v2 is at RC-2 stage with a stable API. Risk is low but non-zero — if a breaking change lands before final release, it would require code adjustments. Mitigation: pin exact versions in `go.mod`, upgrade deliberately.

**Note:** Project initialization using these commands should be the first implementation story.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1. Markdown parsing via goldmark + custom Block struct
2. Document model as `[]Block` slice
3. Gap buffer for within-block text editing
4. Pre-render + cache rendering pipeline
5. In-house vim mode system with per-mode handlers

**Important Decisions (Shape Architecture):**
6. Single Bubbletea model with component structs
7. Standard Go error handling + panic recovery

**Deferred Decisions (Post-MVP):**
8. CI/CD & Release (GitHub Actions + GoReleaser — set up late, doesn't block development)

### Markdown Parsing & Block Model

- **Decision:** goldmark for AST parsing + ink's own `Block` struct
- **Rationale:** goldmark provides correct CommonMark-compliant parsing including edge cases (code fences with internal blank lines, nested lists, tables). ink wraps goldmark's AST nodes in its own `Block` struct containing raw source text, block type, source byte range, and cached Glamour output.
- **Version:** goldmark latest (actively maintained, updated Jan 2026)
- **Affects:** block parser, document model, rendering pipeline, cursor mapping

### Document Data Structure

- **Decision:** `[]Block` slice
- **Rationale:** At ink's target scale (10,000 words ≈ 50-200 blocks), slice operations are negligible cost. Simple to iterate for rendering, simple to serialize back to markdown for saving. No need for linked list or tree complexity.
- **Affects:** document model, viewport rendering, file I/O, auto-save

### Text Buffer (Within-Block Editing)

- **Decision:** Gap buffer
- **Rationale:** Classic text editor data structure with O(1) inserts/deletes at cursor position. Proven pattern, handles larger blocks (long code fences, big lists) well. Slightly more complex than `[]rune` but a better foundation for the editing experience.
- **Affects:** editing block, undo/redo, text manipulation

### Rendering Pipeline & Caching

- **Decision:** Pre-render all blocks on load + cache per block
- **Cache key:** `(block raw content, terminal width)`
- **Rationale:** Pre-rendering on file open means all blocks are immediately available from cache during editing. Block transitions (the <50ms target) become cache lookups — the single edited block is re-rendered via Glamour on `Esc`, all others serve from cache. Terminal resize invalidates all caches but re-renders lazily (visible blocks first).
- **Affects:** rendering pipeline, block transitions, terminal resize handling, memory usage

### Vim Mode Architecture

- **Decision:** In-house implementation with per-mode handler pattern
- **Rationale:** ink's vim modes are tightly coupled with the block reveal/render system — the defining UX pattern. Each mode (normal, insert, visual, command) is a handler struct implementing a common interface. Multi-key sequences (gg, dd) tracked via operator-pending state within the handler. VimTea used as reference for motion/command implementations, not as a dependency.
- **Affects:** all input handling, block transitions, status bar, rendering state

### Component Communication

- **Decision:** Single top-level Bubbletea model with component structs
- **Rationale:** ink's components are too tightly coupled for independent sub-models — a single `Esc` press simultaneously changes mode, renders a block, updates the status bar, and adjusts the viewport. The `EditorModel` owns the Bubbletea lifecycle and calls component methods directly. Components (`Viewport`, `StatusBar`, `EditBlock`, `ModeHandler`) are plain Go structs with methods, not independent Bubbletea models.
- **Affects:** all components, application architecture, testability

### Error Handling & Recovery

- **Decision:** Standard Go error handling + top-level panic recovery
- **Rationale:** Idiomatic `error` returns throughout. A single `defer recover()` in the main program catches panics and attempts emergency save (write document to temp file). User-facing errors displayed in status bar with `E:` prefix and 3-second auto-dismiss per UX spec. ink's error surface is small (file I/O, Glamour rendering, terminal operations) — no need for a custom error framework.
- **Affects:** file I/O, rendering, terminal operations, status bar

### CI/CD & Release

- **Decision:** GitHub Actions + GoReleaser
- **Rationale:** Ecosystem standard (glow, mods, soft-serve use the same pattern). GitHub Actions for CI (test + golangci-lint on push). GoReleaser for cross-platform release builds, Homebrew tap, AUR package, and GitHub Releases with checksums. Set up as a late implementation story.
- **Affects:** build pipeline, distribution, release process

### Decision Impact Analysis

**Implementation Sequence:**
1. goldmark integration + Block struct + `[]Block` document model (foundation)
2. Glamour rendering + cache layer (enables display)
3. Gap buffer + editing block (enables editing)
4. Vim mode handlers (enables interaction)
5. Component structs + EditorModel composition (ties it together)
6. Error handling + panic recovery (reliability)
7. GitHub Actions + GoReleaser (distribution)

**Cross-Component Dependencies:**
- Block parser output feeds into: rendering cache, viewport, editing block, cursor mapper, file serialization
- Vim mode state drives: block transitions, status bar updates, input routing, rendering state
- Rendering cache is consumed by: viewport (all blocks), block transitions (single block re-render), resize handler (cache invalidation)
- Gap buffer is scoped to: editing block only (one active block at a time)

## Implementation Patterns & Consistency Rules

### Purpose

These patterns prevent conflicts when multiple AI agents implement different parts of ink. Each rule addresses a specific point where agents could reasonably make different choices.

### Go Naming Patterns

**Package Naming:**
- Packages use singular, lowercase, one-word names: `block`, `render`, `vim`, `file`, `config`, `editor`, `ui`
- No `utils`, `helpers`, `common`, or `misc` packages — find a meaningful name or put the code where it belongs
- Package names should not stutter with their contents: `block.Block` is fine, `block.BlockParser` is not — use `block.Parser`

**Exported vs Unexported:**
- Types, functions, and methods that cross package boundaries are exported (uppercase)
- Everything else is unexported (lowercase)
- Rule of thumb: start unexported, export only when another package needs it

**Interfaces:**
- Defined by the consumer, not the provider (Go convention)
- Named with `-er` suffix when describing a single method: `Renderer`, `Parser`
- Kept small — prefer one or two methods per interface
- The `ModeHandler` interface (used by vim mode system) is the one exception where the provider package defines it, since all modes implement it

**Error Variables:**
- Sentinel errors use `Err` prefix: `ErrFileNotFound`, `ErrPermissionDenied`
- Defined as package-level `var` using `errors.New()`
- Error messages are lowercase, no trailing punctuation (Go convention): `"file not found: %s"`

**Receiver Names:**
- Single letter, consistent within a type: `func (e *EditorModel)`, `func (b *Block)`, `func (v *Viewport)`
- Never `self` or `this`

**Constants:**
- Unexported constants use `camelCase`: `defaultWidth`, `maxBlockSize`
- Exported constants use `PascalCase`: `DefaultColumnWidth`
- Group related constants with `const ()` blocks

### Package Boundary Rules

**Dependency Direction (strict — no cycles):**

```
cmd/ink → internal/editor
internal/editor → internal/block, internal/render, internal/vim, internal/ui, internal/file, internal/config
internal/render → internal/block (reads block data to render)
internal/vim → internal/block (reads/modifies block content via gap buffer)
internal/ui → internal/block (reads block data for display)
internal/file → internal/block (serializes blocks to markdown)
internal/config → (no internal dependencies — leaf package)
internal/block → (no internal dependencies — leaf package)
```

**What goes where:**
- `internal/block` — Block type definition, block parser (goldmark), gap buffer, document model (`[]Block`). This is the leaf data package — it has no dependencies on other internal packages.
- `internal/render` — Glamour rendering, render cache, block compositing, color interpolation utilities. Depends on `block` for block data.
- `internal/vim` — Mode handler interface, NormalMode/InsertMode/VisualMode/CommandMode handler structs, key dispatch, motion implementations, operator-pending state. Depends on `block` for gap buffer operations.
- `internal/ui` — Viewport struct, StatusBar struct, SavePrompt struct. Depends on `block` for block data to display.
- `internal/file` — File reading, block-to-markdown serialization, atomic writes (temp file + rename), auto-save timer. Depends on `block` for serialization.
- `internal/config` — YAML config loading, default values, CLI override merging. No internal dependencies.
- `internal/editor` — `EditorModel` (the single Bubbletea model), owns all component structs, implements `Init`/`Update`/`View`, coordinates state changes. Depends on everything.

**Shared types:**
- Types shared across packages live in the package that owns the concept: `Block` lives in `internal/block`, `Mode` enum lives in `internal/vim`, `RenderCache` lives in `internal/render`
- No separate `types` package

### Bubbletea Component Patterns

**EditorModel Delegation Pattern:**
The `EditorModel` is the only Bubbletea model. Component structs are fields on this model. The `Update` method follows this pattern:

```go
func (e *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        action := e.modeHandler.HandleKey(msg)
        e.applyAction(action)
    case tea.WindowSizeMsg:
        e.viewport.Resize(msg.Width, msg.Height)
        e.renderCache.InvalidateAll()
    // ... other messages
    }
    return e, nil
}
```

**Component Method Signatures:**
- Components receive what they need as method parameters, not by reaching into EditorModel
- Components return actions/results, they don't mutate EditorModel directly
- Example: `modeHandler.HandleKey(key) → Action` — the editor applies the action, not the mode handler

**State Access Pattern:**
- Components read shared state through method parameters or through a read-only reference
- Only `EditorModel.applyAction()` mutates top-level state
- This creates a clear flow: input → mode handler → action → editor applies → components update

### Block & Document Conventions

**Block Serialization:**
- Blocks serialize back to markdown by joining raw content with `\n\n` (double newline) separators
- The block parser preserves original whitespace — round-tripping (parse → serialize) must produce identical output for unmodified blocks
- Modified blocks serialize from the gap buffer content

**Render Cache Lifecycle:**
- Cache populated on file load (all blocks rendered)
- Cache hit on block transition (Esc → render from cache if content unchanged, re-render if modified)
- Cache invalidated per-block on content change
- Cache invalidated globally on terminal resize (width changes column width)
- Cache entries are `string` (Glamour output) keyed by block index + content hash + width

**Undo/Redo:**
- Undo/redo operates within the active editing block only (gap buffer level)
- Undo stack stores gap buffer snapshots or operation deltas
- Exiting a block (Esc) clears the undo stack for that block — the rendered block is the "committed" state
- No document-level undo (undoing block transitions, block deletion, etc.) in MVP

**Cursor Position Representation:**
- Within a block (editing): `(line, col)` relative to the block's raw content, zero-indexed
- Within the document (normal mode): `(blockIndex, line, col)` — block index in the `[]Block` slice, line/col within the rendered content
- Cursor mapping function: `MapRenderedToRaw(blockIndex, renderedLine, renderedCol) → (rawLine, rawCol)` lives in `internal/block`

### Error & Logging Patterns

**Error Handling Rules:**
- Functions that can fail return `error` as the last return value
- Errors are wrapped with context using `fmt.Errorf("operation: %w", err)`
- Caller decides whether to handle, propagate, or display — not the callee
- User-facing errors go through `EditorModel.showError(msg string)` which sets the status bar `E:` message with 3-second auto-dismiss
- Internal errors that don't affect the user are silently ignored (e.g., render cache miss — just re-render)

**Logging:**
- No logging framework in MVP — ink is a focused TUI app, not a server
- Debug output (if needed during development) via `log.Printf` to a file, never to stdout/stderr (would corrupt the TUI)
- No log statements in production code

**Panic Recovery:**
- Single `defer recover()` in `main.go` wrapping the Bubbletea program
- On panic: attempt emergency save of current document content to `~/.local/state/ink/recovery-{timestamp}.md`
- Print recovery file path to stderr after TUI cleanup

### Testing Patterns

**Test Location:**
- Tests co-located with source: `internal/block/parser_test.go` alongside `internal/block/parser.go` (standard Go convention)
- No separate `test/` directory

**Test Naming:**
- `TestFunctionName_Scenario_ExpectedBehavior`: `TestParser_CodeFenceWithBlankLines_SingleBlock`
- Table-driven tests for functions with multiple input/output cases

**What to Test:**
- Block parser: exhaustive — this is the foundation. Test every markdown element type, edge cases, round-trip serialization.
- Gap buffer: unit tests for insert, delete, cursor movement, content extraction
- Render cache: cache hit/miss behavior, invalidation
- Vim mode handlers: key sequence → action mapping for each mode
- File I/O: atomic write behavior, auto-save triggering
- No TUI integration tests in MVP — test components via their Go interfaces, not through Bubbletea message passing

### Enforcement Guidelines

**All AI Agents MUST:**
1. Check the dependency direction diagram before importing between internal packages — no cycles allowed
2. Put new types in the package that owns the concept — no `types` or `common` packages
3. Follow the EditorModel delegation pattern — components return actions, editor applies them
4. Write tests co-located with source files using table-driven patterns
5. Use the error handling chain: return errors → wrap with context → caller decides

**Anti-Patterns to Avoid:**
- Creating a `utils` or `helpers` package
- Having components directly mutate EditorModel state
- Importing `internal/editor` from any other internal package (dependency flows outward from editor, never inward)
- Writing error messages with uppercase or trailing punctuation
- Adding log statements that write to stdout/stderr
- Creating interfaces before a second implementation exists (except `ModeHandler`)

## Project Structure & Boundaries

### Complete Project Directory Structure

```
ink/
├── cmd/
│   └── ink/
│       └── main.go                      # Entry point: CLI arg parsing, Bubbletea program setup, panic recovery
│
├── internal/
│   ├── block/
│   │   ├── block.go                     # Block type definition (Type enum, raw content, source range, cache slot)
│   │   ├── parser.go                    # goldmark AST → []Block conversion, block boundary extraction
│   │   ├── parser_test.go              # Exhaustive block parsing tests (all markdown element types, edge cases)
│   │   ├── document.go                  # Document type ([]Block), insert/delete/reorder operations, serialize to markdown
│   │   ├── document_test.go            # Round-trip serialization tests, block manipulation tests
│   │   ├── gapbuffer.go                 # Gap buffer implementation (insert, delete, cursor movement, content extraction)
│   │   ├── gapbuffer_test.go           # Gap buffer unit tests
│   │   ├── cursor.go                    # Cursor position types, MapRenderedToRaw, MapRawToRendered
│   │   └── cursor_test.go             # Cursor mapping tests across all markdown element types
│   │
│   ├── render/
│   │   ├── renderer.go                  # Glamour rendering wrapper, render single block to styled string
│   │   ├── renderer_test.go            # Render output tests
│   │   ├── cache.go                     # RenderCache: keyed by (content hash, width), populate/invalidate/lookup
│   │   ├── cache_test.go              # Cache hit/miss/invalidation tests
│   │   └── color.go                     # Color interpolation utility: dim(foreground, background, percentage) → color
│   │
│   ├── vim/
│   │   ├── mode.go                      # Mode enum (Normal, Insert, Visual, Command), ModeHandler interface
│   │   ├── action.go                    # Action type definitions (InsertChar, DeleteChar, MoveCursor, ChangeMode, etc.)
│   │   ├── normal.go                    # NormalMode handler: navigation, mode transitions, operator-pending state
│   │   ├── normal_test.go             # Normal mode key sequence → action tests
│   │   ├── insert.go                    # InsertMode handler: text input, auto-pairs, block splitting
│   │   ├── insert_test.go             # Insert mode tests
│   │   ├── visual.go                    # VisualMode handler: selection extension, yank/delete operations
│   │   ├── visual_test.go             # Visual mode tests
│   │   ├── command.go                   # CommandMode handler: :q, :w, :wq, :w <path> parsing
│   │   ├── command_test.go            # Command parsing tests
│   │   └── motion.go                    # Shared motion implementations (word, line, document motions)
│   │
│   ├── ui/
│   │   ├── viewport.go                  # Viewport: block composition, scroll management, visible block calculation
│   │   ├── statusbar.go                 # StatusBar: mode label, word/char count, dimming, error display, auto-dismiss timer
│   │   ├── saveprompt.go               # SavePrompt: text input for file path, confirm/cancel
│   │   └── layout.go                    # Layout calculations: column width, centering, margins, responsive breakpoints
│   │
│   ├── file/
│   │   ├── file.go                      # File read, file write (atomic: temp + rename), path validation (.md only)
│   │   ├── file_test.go               # Atomic write tests, permission handling tests
│   │   └── autosave.go                 # Auto-save: debounce timer, trigger on typing pause, silent operation
│   │
│   ├── config/
│   │   ├── config.go                    # Config struct, YAML loading, default values, CLI override merging
│   │   └── config_test.go             # Config loading tests, default fallback tests, invalid value handling
│   │
│   └── editor/
│       ├── editor.go                    # EditorModel: Init/Update/View, owns all component structs, state coordination
│       ├── actions.go                   # applyAction: translates vim Actions into state mutations across components
│       └── startup.go                   # Startup logic: blank canvas vs file open, mode selection, initial render
│
├── .github/
│   └── workflows/
│       └── ci.yml                       # GitHub Actions: test + golangci-lint on push/PR
│
├── .goreleaser.yml                      # GoReleaser config: cross-compilation, Homebrew tap, AUR, checksums
├── .golangci.yml                        # Linter config
├── go.mod                               # Module definition + dependencies
├── go.sum                               # Dependency checksums
├── LICENSE                              # MIT (or chosen license)
└── README.md                            # Project README
```

### Architectural Boundaries

**Component Boundaries (enforced by Go's `internal/` + import rules):**

```
                    ┌──────────────────┐
                    │    cmd/ink/main   │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │  internal/editor  │  ← single Bubbletea model, owns everything
                    └──┬──┬──┬──┬──┬───┘
                       │  │  │  │  │
          ┌────────────┘  │  │  │  └────────────┐
          ▼               ▼  │  ▼               ▼
   ┌──────────┐  ┌────────┐  │  ┌──────┐  ┌────────┐
   │  render   │  │  vim   │  │  │  ui  │  │  file  │
   └─────┬────┘  └───┬────┘  │  └──┬───┘  └───┬────┘
         │           │       │     │           │
         └─────┬─────┘       │     └─────┬─────┘
               ▼             │           ▼
          ┌────────┐         │      ┌────────┐
          │ block  │◄────────┘      │ config │
          └────────┘                └────────┘
              ▲                         ▲
              │    leaf packages         │
              │    (no internal deps)    │
              └─────────────────────────┘
```

**Boundary rules:**
- Arrows show allowed import direction
- `block` and `config` are leaf packages — they import nothing from `internal/`
- `editor` is the root — it imports everything, nothing imports it
- No horizontal imports between sibling packages (e.g., `vim` cannot import `render`)
- Cross-component coordination happens in `editor` via the action pattern

### Requirements to Structure Mapping

**FR Category → Package Mapping:**

| FR Category | FRs | Primary Package | Supporting Packages |
|---|---|---|---|
| Document Rendering | FR1-5 | `render` | `block`, `ui/viewport` |
| Block Editing | FR6-11 | `block` (gap buffer, cursor) | `editor` (coordination), `render` (cache) |
| Vim Mode System | FR12-16 | `vim` | `editor` (action dispatch) |
| Text Editing | FR17-22 | `vim` (insert mode), `block` (gap buffer) | `editor` (clipboard) |
| File Management | FR23-29 | `file` | `editor` (startup), `block` (serialization) |
| Status Bar & Feedback | FR30-35 | `ui/statusbar` | `vim` (mode state) |
| Terminal Adaptation | FR36-39 | `ui/layout`, `ui/viewport` | `render` (cache invalidation) |
| Mouse Support | FR40-42 | `editor` (event routing) | `vim` (mode transitions) |
| Configuration | FR43-47 | `config` | `cmd/ink` (CLI overrides) |

**NFR → Implementation Location:**

| NFR Category | NFRs | Implementation Location |
|---|---|---|
| Startup performance | NFR1 | `cmd/ink/main.go`, `editor/startup.go` |
| Block transition performance | NFR2 | `render/cache.go`, `editor/actions.go` |
| Keystroke latency | NFR3-4 | `editor/editor.go` (Update loop) |
| Scroll & resize performance | NFR5-6 | `ui/viewport.go`, `render/cache.go` |
| Large document performance | NFR7 | `block/parser.go`, `render/cache.go` |
| Atomic writes | NFR8-9 | `file/file.go` |
| Panic recovery | NFR10 | `cmd/ink/main.go` |
| Invalid input resilience | NFR11 | `block/parser.go` |
| Accessibility | NFR12-16 | `render/color.go`, `ui/statusbar.go` |
| Terminal compatibility | NFR17-21 | `render/renderer.go`, `ui/layout.go` |

### Data Flow

**Write flow (user types → file saved):**
```
KeyMsg → editor.Update → vim.HandleKey → Action(InsertChar)
→ editor.applyAction → block.GapBuffer.Insert → ui.StatusBar.UpdateCount
→ [typing pause] → file.AutoSave → block.Document.Serialize → file.AtomicWrite
```

**Block transition flow (Esc pressed):**
```
KeyMsg(Esc) → editor.Update → vim.HandleKey → Action(ChangeMode: Normal)
→ editor.applyAction → block.GapBuffer.Content → render.Cache.RenderOrGet
→ ui.Viewport.UpdateBlock → ui.StatusBar.SetMode(Normal) → editor.View
```

**File open flow:**
```
cmd/ink/main → config.Load → file.Read → block.Parse(content)
→ render.Cache.PopulateAll(blocks) → editor.Init(blocks, config)
→ editor.View (all blocks from cache)
```

### Development Workflow

**Build & Run:**
- `go run ./cmd/ink` — development iteration
- `go run ./cmd/ink file.md` — test with a file
- `go build -o ink ./cmd/ink` — local binary

**Test:**
- `go test ./internal/...` — run all tests
- `go test ./internal/block/...` — run block package tests
- `go test -v -run TestParser ./internal/block/...` — run specific test

**Lint:**
- `golangci-lint run ./...` — run linters (configured in `.golangci.yml`)

**Release:**
- `goreleaser release --snapshot --clean` — test release build locally
- Tag + push triggers GitHub Actions → GoReleaser → GitHub Release + Homebrew + AUR

## Architecture Validation Results

### Coherence Validation

**Decision Compatibility:** PASS
- goldmark (CommonMark parser) + Glamour (CommonMark renderer): same specification, compatible output
- Bubbletea v2 + Lip Gloss v2 + Bubbles v2: designed as a coordinated release, API-compatible
- Glamour (v0.x) + Bubbletea v2: confirmed compatible via Bubbletea v2 examples that include Glamour
- Gap buffer inside Block + []Block document model: clean composition, no conflicts
- Per-mode handlers returning Actions + single EditorModel applying them: consistent unidirectional flow
- Pre-render cache + block transitions: cache serves all rendered blocks, transition re-renders only the edited block
- No contradictory decisions found

**Pattern Consistency:** PASS
- Package naming (singular, lowercase) consistent across all packages
- Dependency direction (editor → components → block/config) enforced by Go compiler via `internal/`
- Error handling chain (return → wrap → caller decides) applied uniformly
- Component delegation pattern (return actions, editor applies) consistent across vim, ui, and file components

**Structure Alignment:** PASS
- Every architectural decision maps to a specific package and file
- Dependency diagram matches the import rules defined in patterns
- No structural element exists without a corresponding decision, and no decision lacks a structural home

### Requirements Coverage Validation

**Functional Requirements:** 47/47 COVERED

| FR Range | Category | Covered By |
|---|---|---|
| FR1-5 | Document Rendering | `render/`, `block/`, `ui/viewport` |
| FR6-11 | Block Editing | `block/gapbuffer`, `block/cursor`, `editor/actions`, `render/cache` |
| FR12-16 | Vim Mode System | `vim/` (all mode handlers) |
| FR17-22 | Text Editing | `vim/insert`, `block/gapbuffer`, `editor/actions` (clipboard) |
| FR23-29 | File Management | `file/`, `editor/startup`, `block/document` (serialization) |
| FR30-35 | Status Bar & Feedback | `ui/statusbar`, `vim/mode` |
| FR36-39 | Terminal Adaptation | `ui/layout`, `ui/viewport`, `render/cache` (invalidation) |
| FR40-42 | Mouse Support | `editor/editor` (event routing), `vim/` (mode transitions) |
| FR43-47 | Configuration | `config/`, `cmd/ink/main` |

**Non-Functional Requirements:** 21/21 COVERED

| NFR Range | Category | Architectural Support |
|---|---|---|
| NFR1 | Startup <100ms | Go binary, minimal init path in `editor/startup` |
| NFR2 | Block transition <50ms | Render cache — transitions are cache lookups + single Glamour call |
| NFR3-7 | Perceived performance | Bubbletea v2 Cursed Renderer, direct method calls (no message indirection), lazy resize |
| NFR8-9 | Atomic writes | `file/file.go` — temp file + rename pattern |
| NFR10 | Panic recovery | `cmd/ink/main.go` — `defer recover()` + emergency save |
| NFR11 | Invalid input resilience | goldmark handles any markdown input without panicking |
| NFR12-16 | Accessibility | `render/color.go` (contrast), `ui/statusbar` (text labels, no color-only signaling) |
| NFR17-21 | Terminal compatibility | Lip Gloss/Glamour adaptive themes, color depth detection, graceful degradation |

### Implementation Readiness Validation

**Decision Completeness:** PASS
- 8 core decisions documented with rationale, version info, and affected components
- Implementation sequence defined (7 ordered steps)
- Cross-component dependencies mapped

**Structure Completeness:** PASS
- 30+ files defined with specific purposes
- Every file has a descriptive comment explaining its responsibility
- Boundary diagram and import rules specified

**Pattern Completeness:** PASS
- Naming patterns cover packages, exports, interfaces, errors, receivers, constants
- Structural patterns cover dependency direction, package contents, shared types
- Process patterns cover error handling, logging, panic recovery, testing

### Gap Analysis Results

**No critical gaps found.** All requirements are architecturally supported and all decisions are coherent.

**Important gaps (non-blocking, address during implementation):**

1. **Glamour block-level rendering is unproven** — Glamour is designed for full-document rendering. Calling it per-block with isolated markdown snippets should work for most elements, but may produce inconsistent results for some (e.g., numbered list continuation, cross-reference links). This was identified as a risk in the context analysis and is the highest-risk area to validate during implementation step 3 (build order: rendered blocks + document viewport).
   - **Mitigation:** Build and validate this early (build order step 2-3). If Glamour produces inconsistent per-block output for specific elements, wrap the block in minimal markdown context before rendering.

2. **Visual mode multi-block selection is complex** — FR14 specifies visual mode selecting across blocks, with selected blocks revealing raw markdown. The vim/visual.go handler must track selection across block boundaries, trigger multiple blocks to reveal simultaneously, and map selection coordinates across the rendered/raw boundary. This is architecturally covered but implementation-complex.
   - **Mitigation:** Implement single-block visual mode first, extend to multi-block as a second pass.

3. **System clipboard integration unspecified** — FR22 requires clipboard support. The architecture maps it to `editor/actions.go` but doesn't specify the Go mechanism. Options: `golang.design/x/clipboard` package, or shell exec to `pbcopy`/`xclip`/`wl-copy` per platform.
   - **Mitigation:** Decide during implementation. Either approach works; the architectural boundary (handled in `editor/actions.go`) is correct regardless.

4. **Auto-save timer mechanism** — The architecture specifies "debounce timer on typing pause" but doesn't specify the Bubbletea mechanism. Should use `tea.Tick` commands (Bubbletea's built-in timer) rather than standalone goroutine timers, to stay within the framework's event model.
   - **Mitigation:** Document in `file/autosave.go` — use `tea.Tick` for timer, reset on each keystroke, trigger save on tick completion.

**Nice-to-have gaps (defer to post-MVP):**
- No architecture for focus modes (typewriter, fade) — these are post-MVP per PRD
- No architecture for tabs/multi-file — post-MVP
- No architecture for search — post-MVP

### Architecture Completeness Checklist

**Requirements Analysis**
- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed (low-to-medium)
- [x] Technical constraints identified (Charm ecosystem, terminal, .md only)
- [x] Cross-cutting concerns mapped (5 concerns)

**Architectural Decisions**
- [x] Critical decisions documented with versions (8 decisions)
- [x] Technology stack fully specified (Go 1.25, Bubbletea v2 RC, Lip Gloss v2, Glamour, goldmark)
- [x] Integration patterns defined (action pattern, render cache, component delegation)
- [x] Performance considerations addressed (cache strategy, lazy re-render, Cursed Renderer)

**Implementation Patterns**
- [x] Naming conventions established (packages, exports, errors, receivers, constants)
- [x] Structure patterns defined (dependency direction, package contents, shared types)
- [x] Communication patterns specified (action delegation, state access, EditorModel coordination)
- [x] Process patterns documented (error handling, logging, testing, panic recovery)

**Project Structure**
- [x] Complete directory structure defined (30+ files)
- [x] Component boundaries established (dependency diagram)
- [x] Integration points mapped (data flow diagrams)
- [x] Requirements to structure mapping complete (all 47 FRs + 21 NFRs)

### Architecture Readiness Assessment

**Overall Status:** READY FOR IMPLEMENTATION

**Confidence Level:** High

**Key Strengths:**
- Clean unidirectional dependency graph with no cycles
- Every requirement maps to a specific file and package
- The action delegation pattern (mode handler → action → editor applies) creates a single coordination point that prevents state inconsistencies
- Render cache strategy directly addresses the make-or-break performance target (<50ms block transitions)
- Risk-first build order validates the highest-risk component (Glamour block-level rendering) before building the rest

**Areas for Future Enhancement:**
- Focus modes (typewriter, fade) — post-MVP, layer on top of existing render/color infrastructure
- Multi-file support (tabs) — post-MVP, extends editor model with tab state
- Search — post-MVP, operates on the []Block document model
- Glamour per-block rendering may need adaptation — validate early in build order

### Implementation Handoff

**AI Agent Guidelines:**
- Follow all architectural decisions exactly as documented
- Use implementation patterns consistently across all components
- Respect package boundaries — check the dependency diagram before importing
- Follow the component delegation pattern — components return actions, editor applies them
- Refer to this document for all architectural questions

**First Implementation Priority:**
1. Run initialization commands (go mod init, go get dependencies)
2. Create project directory structure
3. Implement `internal/block` package (parser, Block type, gap buffer) — the foundation everything depends on
4. Implement `internal/render` package (Glamour wrapper, cache) — validates the make-or-break rendering approach
5. If block-level Glamour rendering works well at step 4, the architecture is validated and remaining implementation proceeds with high confidence
