---
stepsCompleted:
  - step-01-validate-prerequisites
  - step-02-design-epics
  - step-03-create-stories
  - step-04-final-validation
inputDocuments:
  - prd.md
  - architecture.md
  - ux-design-specification.md
---

# ink - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for ink, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

FR1: User can view a markdown document with all non-active blocks rendered via Glamour
FR2: User can see rendered blocks update instantly when the active editing block is exited
FR3: User can see the active editing block display raw markdown with syntax characters dimmed and content text at full brightness
FR4: User can view all standard markdown elements rendered inline (headings, bold, italic, links, code spans, block quotes, lists, tables, horizontal rules, code fences)
FR5: User can see the document rendered within a centered writing column that adapts to terminal width
FR6: User can enter a block for editing by pressing insert-initiating vim commands (i, a, o, O) or clicking with the mouse
FR7: User can exit a block by pressing Esc, which renders the block and returns to normal mode
FR8: User can create a new block by pressing Enter twice at the end of the current block, which renders the previous block and creates a new empty editing block
FR9: User can edit any paragraph-level markdown element as a single block (paragraphs, headings, lists, code fences, block quotes, tables, horizontal rules)
FR10: User can have their cursor position accurately mapped between rendered and raw markdown when entering and exiting a block
FR11: User can see surrounding blocks remain stable (no layout shift) when a block transitions between raw and rendered states
FR12: User can operate in normal mode to navigate through a fully rendered document
FR13: User can operate in insert mode to edit raw markdown within the active block
FR14: User can operate in visual mode to select text across one or more blocks, which reveal raw markdown during selection
FR15: User can navigate in normal mode using vim motions (h, j, k, l, w, b, G, gg, Ctrl+d, Ctrl+u)
FR16: User can always return to normal mode by pressing Esc from any other mode
FR17: User can type text and have it inserted at the cursor position in insert mode
FR18: User can delete text using backspace and delete keys
FR19: User can undo and redo edits
FR20: User can have matching characters auto-inserted for markdown pairs (**, __, `, [], ())
FR21: User can use standard markdown indentation (Tab key)
FR22: User can yank (copy) text to the system clipboard and paste from it using vim commands (y, p) and Ctrl+V
FR23: User can open ink with no arguments to get a blank canvas in insert mode
FR24: User can open ink with a file path argument to view the file rendered in normal mode
FR25: User can have their work auto-saved silently on a typing pause
FR26: User can quit instantly with :q or ZZ — named files are already saved via auto-save, unsaved buffers with content prompt once for a file path, empty unsaved buffers are silently discarded
FR27: User can save to a specific path using :w <path>
FR28: User can save and quit using :wq
FR29: User can only open and save .md files
FR30: User can see the current vim mode displayed in a centered status bar (NORMAL, INSERT, VISUAL)
FR31: User can see a live word count and character count in the status bar
FR32: User can see the status bar at full visibility in normal and visual modes, and dimmed (~30%) in insert mode
FR33: User can enter commands via the status bar when pressing : in normal mode
FR34: User can see error messages displayed in the status bar with E: prefix, auto-dismissing after 3 seconds
FR35: User can see the save-as prompt in the status bar when quitting with unsaved buffer content
FR36: User can have the writing column automatically recalculate and content reflow when the terminal is resized
FR37: User can have rendered blocks re-render at the new width on terminal resize
FR38: User can use ink at any terminal width, with centering disabled below 40 characters and full-width used instead
FR39: User can use ink at any terminal height, with the status bar hidden below 5 rows to reclaim space for writing
FR40: User can click on a rendered block to enter insert mode at the click position
FR41: User can scroll the document using the mouse wheel
FR42: User can click to position the cursor in normal mode
FR43: User can use ink with zero configuration — all settings have sensible defaults
FR44: User can optionally customize settings via a YAML config file at ~/.config/ink/config.yml
FR45: User can override config file values with CLI arguments
FR46: User can configure writing column width
FR47: User can have invalid config values silently fall back to defaults

### NonFunctional Requirements

NFR1: Startup time from command execution to usable writing state must be under 100ms
NFR2: Block transitions (raw to rendered and rendered to raw) must complete in under 50ms
NFR3: Keystroke-to-screen latency in insert mode must be imperceptible
NFR4: Word count and character count updates must not cause perceptible delay during typing
NFR5: Terminal resize must recalculate layout and re-render all visible blocks without perceptible delay
NFR6: Scrolling through a rendered document must feel smooth with no rendering stutter
NFR7: Documents of at least 10,000 words must remain performant across all operations
NFR8: Auto-save must never fail silently — save failures must be communicated via status bar error
NFR9: Auto-save must never corrupt file content — writes must be atomic (write to temp, rename)
NFR10: ink must never crash in a way that loses unsaved content — panic recovery should attempt emergency save
NFR11: Invalid markdown input must never cause a crash
NFR12: All dimmed UI elements must maintain minimum readable contrast against the 10 most common terminal color schemes
NFR13: No information may be conveyed by color alone — vim mode via text label, errors via E: prefix, block state via content appearance
NFR14: All functionality must be achievable via keyboard alone — mouse is optional
NFR15: Rendered markdown blocks must output clean text suitable for screen reader parsing
NFR16: The status bar must be positioned consistently for predictable screen reader access
NFR17: ink must function correctly in major terminal emulators: Kitty, Alacritty, WezTerm, iTerm2, GNOME Terminal, Windows Terminal
NFR18: ink must function correctly within tmux and screen sessions
NFR19: ink must support true color (24-bit), 256 color, and 16 color terminals with graceful visual degradation
NFR20: ink must handle SSH sessions with varying terminal capabilities without crashing
NFR21: ink must function on Linux, macOS, and Windows (via Windows Terminal)

### Additional Requirements

**From Architecture:**

- Starter template: Bubbletea v2 (RC) + Standard Go Project Layout with specific initialization commands (go mod init, go get dependencies). Project initialization should be Epic 1, Story 1.
- Markdown parsing via goldmark + custom Block struct for CommonMark-compliant AST parsing
- Document data structure: []Block slice for document model (simple to iterate, serialize)
- Text buffer: Gap buffer for within-block editing (O(1) inserts/deletes at cursor)
- Rendering pipeline: Pre-render all blocks on load + cache per block, keyed by (content hash, terminal width)
- Vim mode architecture: In-house implementation with per-mode handler pattern (NormalMode, InsertMode, VisualMode, CommandMode)
- Component communication: Single top-level Bubbletea model (EditorModel) with component structs, action delegation pattern
- Error handling: Standard Go error handling + top-level panic recovery with emergency save to ~/.local/state/ink/recovery-{timestamp}.md
- CI/CD: GitHub Actions + GoReleaser for cross-platform release builds (set up late, doesn't block development)
- Project structure: cmd/ink/ entry point, internal/ packages (block, render, vim, ui, file, config, editor)
- Package boundary rules: strict dependency direction (no cycles), editor imports everything, block and config are leaf packages
- Block serialization: round-trip (parse then serialize) must produce identical output for unmodified blocks
- Render cache lifecycle: populate on load, hit on transition, invalidate per-block on change, invalidate globally on resize
- Undo/redo: operates within active editing block only (gap buffer level), Esc clears undo stack
- Cursor position: (line, col) within block, (blockIndex, line, col) within document, MapRenderedToRaw function in internal/block
- Testing: co-located test files, table-driven tests, exhaustive block parser tests, no TUI integration tests in MVP

**From UX Design:**

- Mode-Unified Block Reveal pattern: vim mode transition tied to block reveal/render transition (novel UX pattern)
- Centered writing column: 80 chars or 70% of terminal width (whichever smaller), configurable
- Status bar design: centered, middle-dot separators, format "MODE . Xw . Xc"
- No active block boundary/border — raw-vs-rendered distinction is sufficient
- Dimmed syntax characters in editing block: ~60% toward background color
- Status bar dimming in insert mode: ~30% visibility via color interpolation
- Context-aware startup: blank/new document opens in insert mode, existing document opens in normal mode
- Terminal width breakpoints: 120+ (80 char column), 80-119 (70%), 40-79 (full width), below 40 (full width, best effort)
- Terminal height breakpoints: 10+ (full layout), 5-9 (no blank separator), below 5 (status bar hidden)
- Color system: terminal-adaptive, Glamour adaptive theme, no custom background
- Block definition: paragraph-level markdown elements separated by blank lines
- Double-Enter block split: renders previous block, creates new empty editing block
- Editing block boundary behavior: cursor at first/last line + up/down exits block back to normal mode
- Save prompt is the only dialog ink ever shows
- Error messages in status bar: E: prefix, 3-second auto-dismiss, never block the user
- Feedback through state changes only — no notifications, toasts, or success messages
- Accessibility: test contrast against 10 common terminal themes, test keyboard-only operation, minimize decorative Unicode

### FR Coverage Map

FR1: Epic 1 - View rendered markdown document
FR2: Epic 2 - Rendered blocks update on block exit
FR3: Epic 2 - Active editing block with syntax dimming
FR4: Epic 1 - All standard markdown elements rendered
FR5: Epic 1 - Centered writing column adapts to terminal
FR6: Epic 2 - Enter block via vim commands or mouse click
FR7: Epic 2 - Exit block with Esc, renders and returns to normal
FR8: Epic 2 - Double-Enter creates new block
FR9: Epic 2 - Edit any paragraph-level markdown element as block
FR10: Epic 2 - Cursor position mapped between rendered and raw
FR11: Epic 2 - No layout shift on block transitions
FR12: Epic 1 - Normal mode navigation through rendered document
FR13: Epic 2 - Insert mode editing within active block
FR14: Epic 5 - Visual mode selection across blocks
FR15: Epic 1 - Vim motions in normal mode
FR16: Epic 2 - Esc always returns to normal mode
FR17: Epic 2 - Text insertion at cursor in insert mode
FR18: Epic 2 - Delete text with backspace and delete keys
FR19: Epic 5 - Undo and redo edits
FR20: Epic 5 - Auto-pairs for markdown characters
FR21: Epic 2 - Tab key for markdown indentation
FR22: Epic 5 - Clipboard yank and paste
FR23: Epic 2 - Blank canvas in insert mode (no arguments)
FR24: Epic 1 - Open file rendered in normal mode
FR25: Epic 4 - Auto-save on typing pause
FR26: Epic 4 - Quit behaviors (instant, save-as, silent discard)
FR27: Epic 4 - Save to specific path with :w
FR28: Epic 4 - Save and quit with :wq
FR29: Epic 4 - .md files only enforcement
FR30: Epic 3 - Vim mode displayed in centered status bar
FR31: Epic 3 - Live word and character count
FR32: Epic 3 - Status bar dimming in insert mode
FR33: Epic 3 - Command entry via : in normal mode
FR34: Epic 3 - Error messages with E: prefix, auto-dismiss
FR35: Epic 3 - Save-as prompt in status bar
FR36: Epic 6 - Writing column recalculates on resize
FR37: Epic 6 - Rendered blocks re-render on resize
FR38: Epic 6 - Width adaptation with centering breakpoints
FR39: Epic 6 - Height adaptation with status bar hiding
FR40: Epic 6 - Mouse click to enter insert mode
FR41: Epic 6 - Mouse wheel scrolling
FR42: Epic 6 - Mouse click to position cursor in normal mode
FR43: Epic 7 - Zero configuration with sensible defaults
FR44: Epic 7 - YAML config file at ~/.config/ink/config.yml
FR45: Epic 7 - CLI arguments override config file
FR46: Epic 7 - Configurable writing column width
FR47: Epic 7 - Invalid config values fall back to defaults

## Epic List

### Epic 1: View and Navigate a Beautiful Document
The user can open a markdown file and read it as beautifully rendered prose in a centered terminal column, navigating with vim motions — the "glow with a cursor" experience.
**FRs covered:** FR1, FR4, FR5, FR12, FR15, FR24

### Epic 2: Write with Live Preview
The user can enter blocks to edit raw markdown, exit to see instant rendering — the Mode-Unified Block Reveal. Write from a blank canvas or edit existing files. The defining ink experience.
**FRs covered:** FR2, FR3, FR6, FR7, FR8, FR9, FR10, FR11, FR13, FR16, FR17, FR18, FR21, FR23

### Epic 3: Status Bar and Editor Feedback
The user always knows the editor state — mode, word count, character count — through a non-intrusive centered status bar that dims during writing and supports command input.
**FRs covered:** FR30, FR31, FR32, FR33, FR34, FR35

### Epic 4: File Management and Auto-Save
The user never loses work. Auto-save runs silently, quit is instant for named files, save-as prompts once for new buffers. Complete writing sessions from open to close.
**FRs covered:** FR25, FR26, FR27, FR28, FR29

### Epic 5: Advanced Editing Capabilities
Full editing power — visual mode for multi-block selection, undo/redo for safe experimentation, auto-pairs for markdown, clipboard for moving text in and out.
**FRs covered:** FR14, FR19, FR20, FR22

### Epic 6: Terminal Adaptation and Mouse Support
Comfortable writing in any terminal — responsive column on resize, graceful degradation at small sizes, mouse click-to-edit and scroll.
**FRs covered:** FR36, FR37, FR38, FR39, FR40, FR41, FR42

### Epic 7: Configuration and Distribution
Optionally customize the writing experience via config file or CLI flags. Project ready for distribution via package managers.
**FRs covered:** FR43, FR44, FR45, FR46, FR47

## Epic 1: View and Navigate a Beautiful Document

The user can open a markdown file and read it as beautifully rendered prose in a centered terminal column, navigating with vim motions — the "glow with a cursor" experience.

### Story 1.1: Project Initialization and Structure

As a developer,
I want a properly initialized Go project with Bubbletea v2 dependencies and the Architecture-specified directory structure,
So that I have a solid foundation to build all ink components on.

**Acceptance Criteria:**

**Given** a fresh project directory
**When** the initialization commands are run (`go mod init github.com/matheusmortatti/ink`, `go get` for all dependencies)
**Then** `go.mod` contains Bubbletea v2, Lip Gloss v2, Bubbles v2, Glamour, goldmark, and yaml.v3

**Given** the project is initialized
**When** the directory structure is created
**Then** all `internal/` packages exist (`block`, `render`, `vim`, `ui`, `file`, `config`, `editor`) and `cmd/ink/main.go` exists

**Given** `cmd/ink/main.go` exists with a minimal Bubbletea program
**When** `go run ./cmd/ink` is executed
**Then** a Bubbletea program starts and can be exited with `Ctrl+C`
**And** `go build ./cmd/ink` produces a single binary without errors

### Story 1.2: Markdown Block Parser

As a writer,
I want my markdown document parsed into distinct blocks (paragraphs, headings, lists, code fences, block quotes, tables, horizontal rules),
So that each block can be independently rendered and edited.

**Acceptance Criteria:**

**Given** a markdown string containing paragraphs separated by blank lines
**When** the parser processes the string
**Then** each paragraph is returned as a separate `Block` with type `Paragraph` and the correct raw content

**Given** a markdown string containing headings (H1-H6), lists, code fences, block quotes, tables, and horizontal rules
**When** the parser processes the string
**Then** each element is returned as a `Block` with the correct type and raw content

**Given** a code fence containing blank lines
**When** the parser processes the string
**Then** the entire code fence (opening fence through closing fence) is a single `Block`

**Given** a parsed `[]Block` document
**When** the blocks are serialized back to markdown (joined with `\n\n`)
**Then** the output is identical to the original input for unmodified blocks

**Given** invalid or unusual markdown input
**When** the parser processes it
**Then** it returns blocks without crashing (NFR11)

**Given** a document of 10,000+ words
**When** parsed into blocks
**Then** parsing completes without perceptible delay (NFR7)

### Story 1.3: Glamour Block Rendering and Cache

As a writer,
I want each markdown block rendered beautifully via Glamour with results cached for instant retrieval,
So that the document looks polished and rendering is fast.

**Acceptance Criteria:**

**Given** a `Block` containing any supported markdown element (heading, paragraph, bold, italic, links, code spans, block quotes, lists, tables, horizontal rules, code fences)
**When** the block is rendered via Glamour
**Then** the output matches glow-quality rendering with the adaptive dark/light theme (FR4)

**Given** a rendered block result
**When** it is stored in the render cache keyed by (content hash, terminal width)
**Then** subsequent requests for the same block at the same width return the cached result without re-rendering

**Given** a block whose content has changed
**When** a render is requested
**Then** the cache misses and the block is re-rendered with Glamour

**Given** a terminal width change
**When** the cache is invalidated globally
**Then** all blocks are marked for re-rendering at the new width

**Given** a `[]Block` document loaded from a file
**When** all blocks are pre-rendered on load
**Then** the entire cache is populated and ready for display

### Story 1.4: Document Viewport with Centered Writing Column

As a writer,
I want my rendered document displayed in a centered writing column that feels calm and focused,
So that I have a comfortable, distraction-free reading experience.

**Acceptance Criteria:**

**Given** a terminal width of 120+ characters
**When** the viewport renders the document
**Then** the writing column is 80 characters wide, horizontally centered with equal margins on both sides (FR5)

**Given** a terminal width between 80 and 119 characters
**When** the viewport renders the document
**Then** the writing column is 70% of the terminal width, horizontally centered

**Given** a terminal width below 40 characters
**When** the viewport renders the document
**Then** the writing column uses full terminal width with no centering (FR38)

**Given** a document with multiple rendered blocks
**When** the viewport composites them
**Then** blocks are displayed sequentially with standard spacing (1 blank line between blocks)

**Given** a document taller than the terminal height
**When** the viewport is displayed
**Then** only visible blocks are shown and the viewport is scrollable

### Story 1.5: Open and Display Existing Markdown File

As a writer,
I want to type `ink file.md` and immediately see my document rendered beautifully,
So that I can read and review my markdown without leaving the terminal.

**Acceptance Criteria:**

**Given** the user runs `ink myfile.md` with a valid `.md` file
**When** the application starts
**Then** the file is read, parsed into blocks, all blocks are pre-rendered, and the document is displayed in the centered viewport in normal mode (FR24)
**And** startup completes in under 100ms (NFR1)

**Given** the user runs `ink myfile.md` but the file does not exist
**When** the application starts
**Then** the application opens a blank canvas (graceful handling)

**Given** the user runs `ink myfile.txt` (non-.md file)
**When** the application evaluates the argument
**Then** the application rejects the file with an appropriate message (FR29)

**Given** a document with 10,000+ words
**When** opened with `ink largefile.md`
**Then** the document loads and displays without perceptible delay (NFR7)

### Story 1.6: Normal Mode Vim Navigation

As a writer,
I want to navigate through my rendered document using vim motions,
So that I can quickly move to any part of my writing for review.

**Acceptance Criteria:**

**Given** the document is displayed in normal mode
**When** the user presses `j` or `k`
**Then** the cursor moves down or up by one line within the rendered content (FR15)

**Given** the document is displayed in normal mode
**When** the user presses `h` or `l`
**Then** the cursor moves left or right by one character within the rendered content (FR15)

**Given** the document is displayed in normal mode
**When** the user presses `w` or `b`
**Then** the cursor moves forward or backward by one word (FR15)

**Given** the document is displayed in normal mode
**When** the user presses `G`
**Then** the cursor jumps to the end of the document (FR15)

**Given** the document is displayed in normal mode
**When** the user presses `gg`
**Then** the cursor jumps to the beginning of the document (FR15)

**Given** the document is displayed in normal mode
**When** the user presses `Ctrl+d` or `Ctrl+u`
**Then** the viewport scrolls half a page down or up (FR15)

**Given** the cursor moves beyond the visible area
**When** any navigation motion is used
**Then** the viewport scrolls to keep the cursor visible (FR12)
**And** scrolling feels smooth with no rendering stutter (NFR6)

## Epic 2: Write with Live Preview

The user can enter blocks to edit raw markdown, exit to see instant rendering — the Mode-Unified Block Reveal. Write from a blank canvas or edit existing files. The defining ink experience.

### Story 2.1: Gap Buffer for Block Text Editing

As a writer,
I want efficient text editing within a block using a gap buffer,
So that my keystrokes are inserted and deleted instantly at the cursor position.

**Acceptance Criteria:**

**Given** an empty gap buffer
**When** characters are inserted at the cursor position
**Then** the content reflects all inserted characters in order

**Given** a gap buffer with content
**When** the cursor is moved left, right, to start, or to end
**Then** the cursor position updates correctly and subsequent inserts happen at the new position

**Given** a gap buffer with content and cursor positioned mid-text
**When** backspace is pressed
**Then** the character before the cursor is deleted

**Given** a gap buffer with content and cursor positioned mid-text
**When** delete is pressed
**Then** the character after the cursor is deleted

**Given** a gap buffer with content
**When** the full content is extracted
**Then** the returned string matches all inserted text with deletions applied

**Given** a large block of text (e.g., a long code fence or list)
**When** insert and delete operations are performed at the cursor
**Then** operations complete in O(1) time

### Story 2.2: Insert Mode and Text Input

As a writer,
I want to enter insert mode and type prose into a block,
So that I can write and edit my markdown content.

**Acceptance Criteria:**

**Given** the editor is in normal mode with cursor on a rendered block
**When** the user presses `i`
**Then** insert mode is activated, the block displays raw markdown, and the cursor is positioned at the current location within the raw text (FR6, FR13)

**Given** the editor is in normal mode with cursor on a rendered block
**When** the user presses `a`
**Then** insert mode is activated with the cursor positioned after the current character in the raw text (FR6)

**Given** the editor is in normal mode with cursor on a rendered block
**When** the user presses `o`
**Then** insert mode is activated with a new line created below the cursor line within the block (FR6)

**Given** the editor is in normal mode with cursor on a rendered block
**When** the user presses `O`
**Then** insert mode is activated with a new line created above the cursor line within the block (FR6)

**Given** the editor is in insert mode within a block
**When** the user types printable characters
**Then** characters are inserted at the cursor position via the gap buffer (FR17)
**And** keystroke-to-screen latency is imperceptible (NFR3)

**Given** the editor is in insert mode
**When** the user presses backspace or delete
**Then** the character before or after the cursor is removed (FR18)

**Given** the editor is in insert mode
**When** the user presses Tab
**Then** standard markdown indentation is inserted at the cursor position (FR21)

### Story 2.3: Syntax Dimming for Active Editing Block

As a writer,
I want markdown syntax characters dimmed while my content text stays bright in the editing block,
So that I can focus on my words rather than the formatting markup.

**Acceptance Criteria:**

**Given** a block is in raw editing mode containing heading syntax (`#`, `##`, etc.)
**When** the block is rendered for display
**Then** the `#` characters are displayed at ~60% dimmed toward the background color while the heading text is at full brightness (FR3)

**Given** a block in raw editing mode containing bold (`**`), italic (`_`), or other inline syntax
**When** the block is rendered for display
**Then** the syntax characters (`**`, `_`, `` ` ``, `[]`, `()`) are dimmed and the content text is at full brightness (FR3)

**Given** a block in raw editing mode containing a code fence (` ``` `)
**When** the block is rendered for display
**Then** the fence delimiters are dimmed while the code content is at full brightness

**Given** any dimmed syntax characters
**When** displayed against the terminal background
**Then** the dimmed characters maintain minimum readable contrast (NFR12)
**And** block state is distinguishable by content appearance, not color alone (NFR13)

### Story 2.4: Block Transitions — The Mode-Unified Block Reveal

As a writer,
I want blocks to instantly snap between raw markdown and rendered form as I switch modes,
So that the editing experience feels seamless and the document is always beautiful except where I'm actively editing.

**Acceptance Criteria:**

**Given** the editor is in insert mode with an active editing block
**When** the user presses `Esc`
**Then** the active block is rendered via Glamour (from cache if unchanged, re-rendered if modified), the document returns to fully rendered state, and normal mode is activated (FR2, FR7, FR16)
**And** the transition completes in under 50ms (NFR2)

**Given** the editor is in normal mode
**When** the user presses `i`, `a`, `o`, or `O` on a rendered block
**Then** the block reveals raw markdown with syntax dimming and the rest of the document remains rendered (FR6)
**And** the transition completes in under 50ms (NFR2)

**Given** a block transitions between raw and rendered states
**When** the transition occurs
**Then** surrounding blocks remain visually stable with no layout shift (FR11)

**Given** a block is modified in insert mode and then exited with `Esc`
**When** the block re-renders
**Then** the render cache is updated with the new content

**Given** a block is entered and exited without modification
**When** the block re-renders
**Then** the cached rendered output is used (no Glamour call needed)

### Story 2.5: Cursor Position Mapping Between Rendered and Raw

As a writer,
I want my cursor position to map accurately between rendered and raw markdown,
So that when I enter a block to edit, my cursor lands exactly where I expect.

**Acceptance Criteria:**

**Given** the cursor is on a word in a rendered paragraph
**When** the user enters insert mode
**Then** the cursor is positioned at the same word in the raw markdown text (FR10)

**Given** the cursor is on rendered heading text (e.g., "My Title")
**When** the user enters insert mode
**Then** the cursor is positioned at the corresponding text after the `#` characters in the raw markdown (FR10)

**Given** the cursor is on rendered bold text
**When** the user enters insert mode
**Then** the cursor is positioned within the `**...**` markers at the corresponding character (FR10)

**Given** the user exits a block with `Esc` from a position in raw markdown
**When** the block renders and normal mode activates
**Then** the cursor maps back to the corresponding position in the rendered output (FR10)

**Given** any markdown element type (paragraph, heading, list, code fence, block quote, table)
**When** cursor mapping is performed in either direction
**Then** the mapping is consistent and accurate for that element type

### Story 2.6: New Block Creation and Blank Canvas Startup

As a writer,
I want to create new blocks while writing and start with a blank canvas when opening ink without a file,
So that I can begin writing immediately and grow my document naturally.

**Acceptance Criteria:**

**Given** the editor is in insert mode at the end of a block
**When** the user presses Enter twice (creating a blank line)
**Then** the current block is rendered, a new empty editing block is created below, and the cursor is in the new block in insert mode (FR8)

**Given** the user runs `ink` with no arguments
**When** the application starts
**Then** a blank canvas is displayed with the cursor at the top of the centered writing column in insert mode (FR23)

**Given** the user runs `ink` with no arguments
**When** the blank canvas is displayed
**Then** no content is visible except the cursor position — no welcome screen, no tips, no file browser

**Given** the user runs `ink existingfile.md` where the file has content
**When** the application starts
**Then** the document opens fully rendered in normal mode (FR24, context-aware startup)

**Given** the user runs `ink emptyfile.md` where the file exists but is empty
**When** the application starts
**Then** the application opens in insert mode (blank/empty content triggers insert mode)

**Given** a new block is created via double-Enter
**When** the previous block renders
**Then** the block split is instant and the user's typing flow is uninterrupted

## Epic 3: Status Bar and Editor Feedback

The user always knows the editor state — mode, word count, character count — through a non-intrusive centered status bar that dims during writing and supports command input.

### Story 3.1: Status Bar with Mode Display and Word Count

As a writer,
I want a centered status bar showing my current mode, word count, and character count,
So that I always know my editor state and writing progress at a glance.

**Acceptance Criteria:**

**Given** the editor is in normal mode
**When** the status bar is displayed
**Then** it shows `NORMAL · {words}w · {chars}c` centered within the terminal width (FR30, FR31)

**Given** the editor is in insert mode
**When** the status bar is displayed
**Then** it shows `INSERT · {words}w · {chars}c` centered within the terminal width (FR30, FR31)

**Given** the editor is in visual mode
**When** the status bar is displayed
**Then** it shows `VISUAL · {words}w · {chars}c` centered within the terminal width (FR30)

**Given** the user types or deletes text in insert mode
**When** the document content changes
**Then** the word count and character count update in real-time without perceptible delay (FR31, NFR4)

**Given** any mode is active
**When** the status bar is displayed
**Then** the mode is communicated via text label, not color alone (NFR13)
**And** the status bar is positioned consistently at the bottom row of the terminal (NFR16)

**Given** a terminal height of 10+ rows
**When** the layout is calculated
**Then** the status bar occupies the last row with 1 blank line separating it from content

### Story 3.2: Status Bar Mode-Aware Dimming

As a writer,
I want the status bar to fade away when I'm writing and reappear when I'm navigating,
So that I'm not distracted by chrome during creative flow.

**Acceptance Criteria:**

**Given** the editor is in normal mode
**When** the status bar is rendered
**Then** the status bar text is at full visibility (FR32)

**Given** the editor is in visual mode
**When** the status bar is rendered
**Then** the status bar text is at full visibility (FR32)

**Given** the editor is in insert mode
**When** the status bar is rendered
**Then** the status bar text is dimmed to ~30% visibility via color interpolation (foreground shifted 70% toward background) (FR32)

**Given** the user switches from insert mode to normal mode (Esc)
**When** the mode transition occurs
**Then** the status bar snaps instantly from dimmed to full visibility — no animation

**Given** the dimmed status bar is displayed
**When** viewed against any of the 10 most common terminal color schemes
**Then** the dimmed text maintains minimum readable contrast (NFR12)

### Story 3.3: Command Input via Status Bar

As a writer,
I want to type commands like `:q` and `:w` in the status bar,
So that I can control the editor using familiar vim command patterns.

**Acceptance Criteria:**

**Given** the editor is in normal mode
**When** the user presses `:`
**Then** the status bar content is replaced with `:` followed by a cursor, and command mode is activated (FR33)

**Given** the editor is in command mode
**When** the user types characters
**Then** the characters are appended to the command string after the `:` prefix

**Given** the editor is in command mode
**When** the user presses backspace
**Then** the last character of the command string is deleted

**Given** the editor is in command mode
**When** the user presses `Enter`
**Then** the command is executed and the status bar returns to normal status display

**Given** the editor is in command mode
**When** the user presses `Esc`
**Then** the command is cancelled, input is discarded, and the editor returns to normal mode with the standard status bar display

**Given** the user enters an unrecognized command (e.g., `:xyz`)
**When** `Enter` is pressed
**Then** an error is displayed: `E: Not a command: xyz`

### Story 3.4: Error Display and Save-As Prompt

As a writer,
I want errors shown briefly in the status bar and a save prompt when quitting with unsaved work,
So that I'm informed of problems without being interrupted and can name my file when needed.

**Acceptance Criteria:**

**Given** an error occurs (file not found, permission denied, unknown command, etc.)
**When** the error is triggered
**Then** the status bar displays the error with `E:` prefix (e.g., `E: File not found: path.md`) (FR34)

**Given** an error message is displayed in the status bar
**When** 3 seconds have elapsed
**Then** the error auto-dismisses and the status bar returns to normal status display (FR34)

**Given** an error is displayed
**When** the user performs any action before the 3-second timeout
**Then** the error remains visible until the timeout completes — user actions do not dismiss errors early

**Given** the user quits (`:q` or `ZZ`) with an unsaved buffer that has content
**When** the quit is initiated
**Then** the status bar displays `Save as: ` with a text input cursor (FR35)

**Given** the save-as prompt is active
**When** the user types a file path and presses `Enter`
**Then** the file is saved to the specified path and ink exits

**Given** the save-as prompt is active
**When** the user presses `Esc`
**Then** the prompt is dismissed, the editor returns to normal mode, and no data is lost

**Given** the save-as prompt is active and the user enters an invalid path
**When** `Enter` is pressed
**Then** an error is displayed (e.g., `E: Invalid path: ...`) and the save prompt remains active for retry

## Epic 4: File Management and Auto-Save

The user never loses work. Auto-save runs silently, quit is instant for named files, save-as prompts once for new buffers. Complete writing sessions from open to close.

### Story 4.1: Auto-Save on Typing Pause

As a writer,
I want my work saved automatically and silently when I pause typing,
So that I never lose work and never think about saving.

**Acceptance Criteria:**

**Given** the user is editing a named file (opened with `ink file.md`)
**When** the user pauses typing for the auto-save delay (default 1000ms)
**Then** the document is saved silently to the file path without any visible indication (FR25)

**Given** the auto-save triggers
**When** the file is written
**Then** the write is atomic — content is written to a temporary file first, then renamed to the target path (NFR9)

**Given** the auto-save triggers but the write fails (e.g., permission denied, disk full)
**When** the failure occurs
**Then** an error is displayed in the status bar with `E:` prefix (e.g., `E: Cannot save: permission denied`) (NFR8)
**And** the document remains open and editable — no data is lost

**Given** the user is typing continuously
**When** each keystroke occurs
**Then** the auto-save timer resets (debounce), preventing saves during active typing

**Given** the user is editing an unsaved buffer (no file path)
**When** the auto-save timer fires
**Then** no save is attempted — auto-save only operates on named files

**Given** the user edits a 10,000+ word document
**When** auto-save triggers
**Then** the save completes without perceptible interruption to typing (NFR7)

### Story 4.2: Quit Behaviors and Save Commands

As a writer,
I want predictable quit and save behaviors that match vim conventions,
So that ending a writing session is instant and my work is always preserved.

**Acceptance Criteria:**

**Given** the user is editing a named file with auto-save active
**When** the user types `:q` or `ZZ`
**Then** ink exits instantly — the file is already saved via auto-save (FR26)

**Given** the user has an unsaved buffer with content (no file path)
**When** the user types `:q` or `ZZ`
**Then** the save-as prompt appears in the status bar (FR26)

**Given** the user has an empty unsaved buffer (no content, no file path)
**When** the user types `:q` or `ZZ`
**Then** ink exits silently without any prompt (FR26)

**Given** the user is in any mode
**When** the user types `:w <path>` where path ends in `.md`
**Then** the document is saved to the specified path via atomic write (FR27)

**Given** the user types `:w <path>` where path does not end in `.md`
**When** the command is executed
**Then** an error is displayed: `E: Only .md files supported` (FR29)

**Given** the user types `:wq`
**When** the command is executed
**Then** the document is saved (to existing path or prompting for path if unsaved) and ink exits (FR28)

**Given** the user types `:w` without a path on an unsaved buffer
**When** the command is executed
**Then** the save-as prompt appears for the user to specify a path

**Given** the user types `:w` without a path on a named file
**When** the command is executed
**Then** the file is saved to its existing path

### Story 4.3: Panic Recovery and Emergency Save

As a writer,
I want my document recovered even if ink crashes unexpectedly,
So that I never lose my writing regardless of what goes wrong.

**Acceptance Criteria:**

**Given** ink is running with document content
**When** an unrecoverable panic occurs
**Then** the `defer recover()` in main catches the panic and attempts to write the current document content to `~/.local/state/ink/recovery-{timestamp}.md` (NFR10)

**Given** an emergency save is attempted after a panic
**When** the save succeeds
**Then** the recovery file path is printed to stderr after the TUI is cleaned up

**Given** an emergency save is attempted after a panic
**When** the save fails (e.g., the recovery directory cannot be created)
**Then** ink exits with the panic information printed to stderr — best effort, no secondary panic

**Given** the `~/.local/state/ink/` directory does not exist
**When** an emergency save is triggered
**Then** the directory is created before writing the recovery file

**Given** invalid markdown input is provided to the parser or renderer
**When** processing occurs
**Then** ink handles the input gracefully without panicking (NFR11)

## Epic 5: Advanced Editing Capabilities

Full editing power — visual mode for multi-block selection, undo/redo for safe experimentation, auto-pairs for markdown, clipboard for moving text in and out.

### Story 5.1: Undo and Redo

As a writer,
I want to undo and redo my edits within a block,
So that I can experiment freely knowing I can reverse any change.

**Acceptance Criteria:**

**Given** the user has made edits within an active editing block
**When** the user presses `u` in normal mode
**Then** the most recent edit is undone and the block content reverts to the previous state (FR19)

**Given** the user has undone one or more edits
**When** the user presses `Ctrl+r` in normal mode
**Then** the most recently undone edit is reapplied (FR19)

**Given** the user has made multiple edits within a block
**When** `u` is pressed repeatedly
**Then** edits are undone in reverse chronological order, one at a time

**Given** there are no more edits to undo
**When** the user presses `u`
**Then** nothing happens — no error, no crash

**Given** the user exits a block with `Esc` (rendering it)
**When** the block transitions to rendered state
**Then** the undo/redo stack for that block is cleared — the rendered block is the committed state

**Given** the user re-enters the same block
**When** insert mode is activated
**Then** a fresh undo/redo stack begins — previous session edits are not available for undo

### Story 5.2: Auto-Pairs for Markdown

As a writer,
I want matching characters auto-inserted when I type markdown formatting pairs,
So that I can format my text efficiently without manually closing each pair.

**Acceptance Criteria:**

**Given** the user is in insert mode
**When** the user types `**`
**Then** a matching `**` is auto-inserted after the cursor, with the cursor positioned between the pairs (FR20)

**Given** the user is in insert mode
**When** the user types `__`
**Then** a matching `__` is auto-inserted after the cursor, with the cursor positioned between the pairs (FR20)

**Given** the user is in insert mode
**When** the user types a single backtick `` ` ``
**Then** a matching backtick is auto-inserted after the cursor, with the cursor positioned between them (FR20)

**Given** the user is in insert mode
**When** the user types `[`
**Then** a matching `]` is auto-inserted after the cursor, with the cursor positioned between them (FR20)

**Given** the user is in insert mode
**When** the user types `(`
**Then** a matching `)` is auto-inserted after the cursor, with the cursor positioned between them (FR20)

**Given** auto-paired characters have been inserted
**When** the user types the closing character manually
**Then** the cursor moves past the existing closing character rather than inserting a duplicate

### Story 5.3: Clipboard Integration

As a writer,
I want to yank and paste text using vim commands and the system clipboard,
So that I can move text freely in and out of ink.

**Acceptance Criteria:**

**Given** the user has selected text in visual mode
**When** the user presses `y`
**Then** the selected text is copied to the system clipboard (FR22)

**Given** text has been yanked to the clipboard
**When** the user presses `p` in normal mode
**Then** the clipboard content is pasted after the cursor position (FR22)

**Given** text has been yanked to the clipboard
**When** the user presses `P` in normal mode
**Then** the clipboard content is pasted before the cursor position

**Given** text has been copied from an external application to the system clipboard
**When** the user presses `p` in normal mode within ink
**Then** the external clipboard content is pasted into the document

**Given** text is yanked within ink
**When** the user switches to another application and pastes
**Then** the yanked text is available in the system clipboard

**Given** the user presses `dd` in normal mode
**When** the current line is deleted
**Then** the deleted line content is placed in the system clipboard

### Story 5.4: Visual Mode Selection

As a writer,
I want to select text visually across one or more blocks,
So that I can yank, delete, or replace specific portions of my writing.

**Acceptance Criteria:**

**Given** the editor is in normal mode
**When** the user presses `v`
**Then** character-wise visual mode is activated, the current block reveals raw markdown, and the cursor position marks the selection start (FR14)

**Given** the editor is in normal mode
**When** the user presses `V`
**Then** line-wise visual mode is activated, the current block reveals raw markdown, and the current line is selected (FR14)

**Given** visual mode is active within a single block
**When** the user moves the cursor with vim motions
**Then** the selection extends or contracts between the anchor and cursor position, with selected text highlighted

**Given** visual mode is active and the cursor moves into an adjacent block
**When** the selection crosses a block boundary
**Then** the adjacent block also reveals raw markdown for precise selection (FR14)

**Given** text is selected in visual mode
**When** the user presses `y`
**Then** the selected text is yanked to the clipboard and visual mode exits — all revealed blocks snap back to rendered

**Given** text is selected in visual mode
**When** the user presses `d`
**Then** the selected text is deleted, visual mode exits, and all blocks snap back to rendered

**Given** the editor is in visual mode
**When** the user presses `Esc`
**Then** the selection is cleared, all revealed blocks render, and normal mode is restored (FR16)

## Epic 6: Terminal Adaptation and Mouse Support

Comfortable writing in any terminal — responsive column on resize, graceful degradation at small sizes, mouse click-to-edit and scroll.

### Story 6.1: Terminal Resize Handling

As a writer,
I want ink to adapt seamlessly when I resize my terminal,
So that my writing environment stays comfortable regardless of window changes.

**Acceptance Criteria:**

**Given** the user resizes the terminal while ink is running
**When** the `WindowSizeMsg` event is received
**Then** the writing column width recalculates according to the breakpoints: 120+ chars → 80 char column, 80-119 → 70%, 40-79 → full width, below 40 → full width (FR36, FR38)

**Given** the writing column width changes due to resize
**When** the new width is applied
**Then** content reflows within the new column width and all visible rendered blocks re-render at the new width via Glamour (FR37)

**Given** the render cache contains entries at the old terminal width
**When** a resize occurs
**Then** the entire render cache is invalidated and visible blocks are re-rendered lazily (visible first)

**Given** the user is in insert mode editing a block during a resize
**When** the resize occurs
**Then** the active editing block adapts its line wrapping to the new width and the cursor remains in a valid position

**Given** any resize event
**When** the layout recalculates
**Then** the centering margins update to `(terminal_width - column_width) / 2` and the viewport adjusts without perceptible delay (NFR5)
**And** no flickering or rendering artifacts occur

### Story 6.2: Terminal Height Adaptation

As a writer,
I want ink to make the best use of available vertical space,
So that I can write comfortably even in small terminal windows.

**Acceptance Criteria:**

**Given** a terminal height of 10 or more rows
**When** the layout is calculated
**Then** the full layout is displayed: content area + 1 blank line separator + status bar on the last row (FR39)

**Given** a terminal height between 5 and 9 rows
**When** the layout is calculated
**Then** the layout omits the blank separator: content area + status bar on the last row, no gap between them (FR39)

**Given** a terminal height below 5 rows
**When** the layout is calculated
**Then** the status bar is hidden entirely and all rows are used for the writing content (FR39)

**Given** the terminal height changes during use (resize)
**When** the height crosses a breakpoint threshold (e.g., from 10 to 4 rows)
**Then** the layout adapts immediately to the appropriate configuration

**Given** the status bar is hidden due to small terminal height
**When** the terminal is resized to 5+ rows
**Then** the status bar reappears with the correct mode and word/char count

### Story 6.3: Mouse Click to Edit and Position Cursor

As a writer,
I want to click on my document to position the cursor or enter editing,
So that I can navigate and edit intuitively with the mouse as an alternative to keyboard.

**Acceptance Criteria:**

**Given** the editor is in normal mode and the user clicks on a rendered block
**When** the click event is processed
**Then** insert mode is activated at the click position — the block reveals raw markdown with the cursor mapped to the corresponding raw text position (FR40)

**Given** the editor is in normal mode and the user clicks on a position within the rendered content
**When** the click event is processed
**Then** the cursor moves to the clicked position within the rendered document (FR42)

**Given** the editor is in insert mode and the user clicks on a different rendered block
**When** the click event is processed
**Then** the current block renders (Esc behavior), the clicked block reveals raw markdown, and insert mode continues at the click position in the new block

**Given** the editor is in insert mode and the user clicks within the same active editing block
**When** the click event is processed
**Then** the cursor moves to the clicked position within the raw markdown

**Given** all mouse interactions
**When** the click is processed
**Then** the behavior is identical to the equivalent keyboard-initiated action — mouse is an alternative input, not a different interaction model (NFR14 — keyboard remains sufficient)

### Story 6.4: Mouse Wheel Scrolling

As a writer,
I want to scroll through my document using the mouse wheel,
So that I can quickly browse through long documents without keyboard commands.

**Acceptance Criteria:**

**Given** the document is longer than the visible viewport
**When** the user scrolls down with the mouse wheel
**Then** the viewport scrolls down, revealing content below (FR41)

**Given** the document is longer than the visible viewport
**When** the user scrolls up with the mouse wheel
**Then** the viewport scrolls up, revealing content above (FR41)

**Given** the user scrolls to the end of the document
**When** further scroll-down events occur
**Then** nothing happens — no bounce, no error, no scroll past the end

**Given** the user scrolls to the beginning of the document
**When** further scroll-up events occur
**Then** nothing happens — the viewport stays at the top

**Given** the user scrolls while in insert mode
**When** the active editing block scrolls out of view
**Then** the viewport allows the scroll — the editing block remains active and scrolling back reveals it in its editing state

## Epic 7: Configuration and Distribution

Optionally customize the writing experience via config file or CLI flags. Project ready for distribution via package managers.

### Story 7.1: Configuration Loading with Sensible Defaults

As a writer,
I want ink to work perfectly out of the box and optionally let me tweak settings via a config file,
So that I never need to configure anything but can customize if I want to.

**Acceptance Criteria:**

**Given** no config file exists at `~/.config/ink/config.yml`
**When** ink starts
**Then** all settings use sensible defaults: width 80, typewriter false, fade false, autosave_delay 1000ms (FR43)

**Given** a valid config file exists at `~/.config/ink/config.yml`
**When** ink starts
**Then** the config values are loaded and applied (FR44)

**Given** the config file contains a `width` setting
**When** ink starts
**Then** the writing column width uses the configured value (FR46)

**Given** the config file contains invalid values (e.g., `width: -5` or `width: "banana"`)
**When** ink starts
**Then** the invalid values silently fall back to their defaults — no error, no crash (FR47)

**Given** the config file is malformed YAML
**When** ink starts
**Then** the entire config file is ignored and all defaults are used — no error displayed, no crash

**Given** the config file contains unknown keys (e.g., `theme: dark`)
**When** ink starts
**Then** the unknown keys are silently ignored

### Story 7.2: CLI Argument Overrides

As a writer,
I want to override settings with command-line flags,
So that I can adjust ink's behavior for a specific session without changing my config file.

**Acceptance Criteria:**

**Given** the user runs `ink --width 60`
**When** ink starts
**Then** the writing column width is set to 60 characters, overriding both the default and any config file value (FR45)

**Given** the user runs `ink --width 60 file.md`
**When** ink starts
**Then** the file opens with the column width set to 60 — CLI args and file path coexist

**Given** the config file sets `width: 100` and the user runs `ink --width 60`
**When** ink starts
**Then** the column width is 60 — CLI arguments take precedence over config file (FR45)

**Given** the user runs `ink --version`
**When** the flag is processed
**Then** the version string is printed to stdout and ink exits immediately

**Given** the user runs `ink --help`
**When** the flag is processed
**Then** usage information is printed to stdout (command structure, available flags) and ink exits immediately

**Given** the user provides an unrecognized flag (e.g., `ink --unknown`)
**When** the flag is processed
**Then** an error message is printed to stderr and ink exits with a non-zero exit code

### Story 7.3: CI/CD and Distribution Setup

As a developer,
I want automated testing, linting, and cross-platform release builds,
So that ink is reliable, consistent, and easy to install for users.

**Acceptance Criteria:**

**Given** a push or pull request to the repository
**When** GitHub Actions CI runs
**Then** `go test ./internal/...` passes and `golangci-lint run ./...` passes

**Given** a git tag is pushed (e.g., `v0.1.0`)
**When** GoReleaser is triggered via GitHub Actions
**Then** cross-platform binaries are built for Linux (amd64, arm64), macOS (amd64, arm64), and Windows (amd64)

**Given** a release is built by GoReleaser
**When** the release is published
**Then** GitHub Releases contains the binaries with checksums

**Given** a release is published
**When** the Homebrew tap is configured
**Then** users can install ink via `brew install matheusmortatti/tap/ink`

**Given** the CI pipeline
**When** `.golangci.yml` is configured
**Then** the linter configuration matches the project's coding standards from the Architecture document
