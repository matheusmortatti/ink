# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ink** is a terminal markdown editor with live preview and Vim motions, built in Go using the Bubble Tea TUI framework. Its key innovation is hybrid rendering: inactive markdown blocks are fully rendered (like a preview), while the active block (where the cursor is) shows raw markdown text for editing.

## Build & Test Commands

```bash
go build -o markdown-term .   # Build the binary
go test ./...                  # Run all tests
go test -v ./...               # Run tests with verbose output
go test -run TestName ./...    # Run a single test
go vet ./...                   # Static analysis
```

Run the editor: `./markdown-term [file.md]`

## Architecture

The app follows the **Bubble Tea Model-Update-View** pattern. All state flows through message passing.

### Core Data Flow

```
User Input → vim.go (modal key handling) → buffer.go (content mutation) → document.go (re-parse markdown) → render.go (hybrid display) → View()
```

### Key Files

- **`model.go`** — Top-level Bubble Tea model. Owns all state: Buffer, VimState, Viewport, HybridRenderer, Document. Handles the Update loop and coordinates between subsystems.

- **`vim.go`** (~1100 lines, largest file) — Vim state machine. Implements Normal, Insert, Visual, Visual-Line, Command, and Replace modes. Handles operator-pending state (e.g., `d` waits for motion), count prefixes, registers, dot-repeat, and search (`/`, `?`, `n`, `N`).

- **`buffer.go`** — Source-of-truth for document content. Stores lines as `[]string`. Manages cursor position, undo/redo (full snapshot-based), and all text mutations (insert, delete, word motions, find-char, join-line, etc.).

- **`document.go`** — Parses buffer content into markdown blocks using Goldmark. Each `Block` has a `Kind` (heading, paragraph, fenced code, list, blockquote, etc.), source line range, and raw text.

- **`render.go`** — The hybrid renderer. Active block → raw text with Chroma syntax highlighting. Inactive blocks → fully rendered markdown via Glamour. Uses SHA256-based render cache to avoid re-rendering unchanged blocks.

- **`linemap.go`** — Bidirectional mapping between source lines (editing coordinates) and display lines (rendered output coordinates). Essential because rendered markdown expands lines (e.g., a heading becomes multiple display lines).

- **`viewport.go`** — Scroll offset management. Keeps cursor visible with margin.

- **`commands.go`** — Ex command execution (`:w`, `:q`, `:e`, `:<line>`).

- **`styles.go`** — Lip Gloss color palette and mode-specific styling.

- **`statusbar.go`** — Mode indicator, filename, cursor position, and command-line input rendering.

### Key Design Patterns

- **Debounced re-parsing**: Markdown re-parsing during insert mode is debounced (150ms) to avoid excessive parsing overhead.
- **Undo/redo via full snapshots**: Each snapshot stores the entire `[]string` lines and cursor position.
- **Operator-pending mode**: In vim.go, operators like `d`, `c`, `y` set `PendingOp` and wait for the next motion/text-object to determine the range.
- **Content width**: Rendering is constrained to 100 chars max, centered in the terminal.

## Dependencies

- **Bubble Tea** (`charmbracelet/bubbletea`) — TUI framework
- **Glamour** (`charmbracelet/glamour`) — Markdown rendering for inactive blocks
- **Lip Gloss** (`charmbracelet/lipgloss`) — Terminal styling
- **Goldmark** (`yuin/goldmark`) — Markdown parsing into block AST
- **Chroma** (`alecthomas/chroma`) — Syntax highlighting for active blocks
