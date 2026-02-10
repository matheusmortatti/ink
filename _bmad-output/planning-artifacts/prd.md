---
stepsCompleted:
  - step-01-init
  - step-02-discovery
  - step-03-success
  - step-04-journeys
  - step-05-domain
  - step-06-innovation
  - step-07-project-type
  - step-08-scoping
  - step-09-functional
  - step-10-nonfunctional
  - step-11-polish
inputDocuments:
  - product-brief-ink-2026-02-07.md
  - brainstorming-session-2026-02-07.md
  - ux-design-specification.md
documentCounts:
  briefs: 1
  research: 0
  brainstorming: 1
  projectDocs: 0
  uxDesign: 1
classification:
  projectType: cli_tool
  domain: general
  complexity: low
  projectContext: greenfield
workflowType: 'prd'
---

# Product Requirements Document - ink

**Author:** Matheusmortatti
**Date:** 2026-02-08

## Executive Summary

ink is a free, open-source, terminal-native markdown editor built for writing, not coding. It fills a gap no existing tool addresses: a distraction-free, beautiful prose writing experience in the terminal with Obsidian-style inline live preview, vim motions, and a writing environment designed to disappear so the writer can focus entirely on their words.

Built with Go and the Charm ecosystem (Bubbletea, Lip Gloss, Glamour), ink is the natural complement to glow — where glow lets you *read* markdown beautifully in the terminal, ink lets you *write* it beautifully. It ships zero-config, markdown-only, and writing-first: open it and start typing.

**Core differentiator:** The Mode-Unified Block Reveal — a novel pattern that ties vim modal editing to Obsidian's block-level preview. Entering insert mode reveals raw markdown for editing; returning to normal mode renders it beautifully. No other terminal editor offers this interaction. The tool competes on friction *removed*, not features added.

**Target user:** Terminal-native writers — developers, technical writers, and creatives who live in the terminal, are comfortable with vim motions, and want the gap between "I want to write" and "I am writing" measured in seconds.

**Project context:** Greenfield, open-source, solo developer (evenings/weekends). No commercial targets — success is personal utility first, community resonance second.

## Success Criteria

### User Success

- **Writing frequency increases:** The user writes more often than before ink — friction reduction translates into actual output.
- **ink becomes the default:** When the urge to write hits, `ink` is the first tool reached for. Not Neovim, not Obsidian, not a GUI app.
- **The invisible tool:** Users complete writing sessions without consciously thinking about the tool. This is a qualitative, felt experience — validated by the creator and confirmed by user sentiment over time.
- **Zero-to-writing speed:** The time between "I want to write" and "I am writing" is measured in seconds — type `ink`, start typing.

### Business Success

ink is a free, open-source personal project. No revenue or commercial targets.

- **Personal utility (primary):** The creator uses ink as their default writing tool. Consistent habit within 1 month of stable release.
- **Community resonance (secondary):** Organic adoption by terminal writers who share the same frustrations. Any non-trivial external contribution signals product-market fit.
- **Ecosystem recognition (aspirational):** ink earns a place in the Charm ecosystem as the write-side complement to glow.

### Technical Success

Performance targets are critical to ink's identity — the invisible tool philosophy demands that the user never perceives the tool's mechanics. See Non-Functional Requirements for full measurable specifications.

- **Startup:** Under 100ms to a usable writing state (NFR1)
- **Block transitions:** Under 50ms for raw ↔ rendered snap (NFR2)
- **No layout shift:** Surrounding blocks remain stable during transitions (FR11)
- **Rendering quality:** Visually identical to glow's output via Glamour
- **Auto-save reliability:** Zero data loss, silent and unfailing (NFR8-9)

### Measurable Outcomes

| Outcome | Measurement | Target |
|---------|-------------|--------|
| Personal daily use | Creator reaches for ink first when writing | Consistent habit within 1 month of stable release |
| Startup performance | Time from command to usable writing state | < 100ms |
| Block transition speed | Time for raw ↔ rendered snap | < 50ms (perceptually instant) |
| Writing output | Prose written in ink vs. other tools | Majority of personal prose writing |
| Community signal | Issues, discussions, contributions | Any non-trivial external contribution |
| Package availability | Presence in package managers | At least 2 (e.g., Homebrew, AUR) within 6 months |

## User Journeys

### Journey 1: Alex Writes from Scratch

**Opening Scene:** It's 10 PM. Alex just finished a long coding session and an idea for a blog post hits — something about a pattern they noticed in their codebase. The creative impulse is fragile. In the past, they'd open Obsidian, get distracted by the sidebar of old notes, lose the thread. Or they'd open Neovim, stare at line numbers and a statusline full of git info, and feel like they're still coding. Tonight, they type `ink`.

**Rising Action:** A blank canvas appears. Cursor at the top of a centered column. Nothing else — no file browser, no welcome screen, no tip of the day. Just a dimmed status bar whispering `INSERT · 0w · 0c`. Alex starts typing. The words come. A heading, a paragraph, another paragraph. They don't think about saving. They don't think about formatting. They type `**important**` out of habit and keep going. When they pause to think, the file silently saves itself. They don't notice.

**Climax:** Three paragraphs in, Alex presses `Esc` to re-read what they've written. The raw markdown snaps into beautifully rendered prose — headings styled, bold text weighted, the document looking *finished* even though they just started. They navigate down with `j`, reading their own words in glow-quality rendering. They spot a weak sentence, press `i`, the block reveals, they fix it, press `Esc` — it renders again. The edit cycle feels like reaching into a finished document and adjusting it.

**Resolution:** Forty minutes later, the post is done. Alex types `:q`. A single prompt: `Save as:`. They type `~/posts/codebase-patterns.md`, hit Enter. ink exits. The terminal prompt returns. The blog post exists. Alex never once thought about the tool — they thought about codebase patterns. The invisible tool delivered.

**Requirements revealed:** Blank canvas start, insert mode on launch for new documents, auto-save on pause, block-level rendering transitions, centered writing column, vim motions, save-as prompt for unsaved buffers, instant quit.

### Journey 2: Alex Edits an Existing File

**Opening Scene:** Next morning. Alex wants to revise last night's blog post before publishing. They type `ink ~/posts/codebase-patterns.md`. The document opens fully rendered — headings, bold, paragraphs all beautifully formatted. Normal mode. Status bar reads `NORMAL · 847w · 4,692c`. It feels like reading the post in glow, except there's a cursor.

**Rising Action:** Alex navigates with `j` and `k`, reading through the rendered document. They reach a paragraph that needs reworking. They press `i` — the block snaps to raw markdown, syntax characters dimmed, their words prominent. They rewrite the paragraph. Auto-save fires silently on a pause. They press `Esc` — the block renders. The revised paragraph sits beautifully among the others.

**Climax:** Alex moves to a section with a list. They press `i` on the list block — the entire list reveals as raw markdown. They add two items, reorder another. Press `Esc`. The list renders with proper bullets and indentation. They scroll to the top with `gg`, read the whole post in rendered form. It's ready.

**Resolution:** Alex types `ZZ`. ink saves and exits instantly — no prompt, because the file already has a path and auto-save has been running throughout. Terminal prompt returns. The revision took eight minutes. Alex never opened a browser, never saw a sidebar, never clicked a button.

**Requirements revealed:** Normal mode on launch for existing files, rendered document navigation, cursor-to-raw position mapping, block editing for different block types (paragraphs, lists), `ZZ` and `:q` quit commands, auto-save to existing file path, instant exit without prompt for named files.

### Journey 3: Alex Recovers from the Unexpected

**Opening Scene:** Alex is mid-session writing a long essay. Three sections done, deep in flow.

**Scenario A — Terminal resize:** Alex's tiling window manager shifts the terminal from half-screen to quarter-screen. ink instantly recalculates — the writing column narrows, content reflows, rendered blocks re-render at the new width. The cursor stays in view. Alex barely notices. They keep writing.

**Scenario B — Accidental quit:** Alex's muscle memory fires `:q` when they meant `:w`. But it doesn't matter — auto-save has been running the whole time. The file is already saved. ink exits. Alex types `ink essay.md` and is back exactly where the document was. The only thing lost is cursor position — the content is intact.

**Scenario C — Non-existent file:** Alex types `ink drafts/new-idea.md` but the `drafts` directory doesn't exist. The status bar briefly shows `E: File not found: drafts/new-idea.md` for three seconds, then ink opens a blank canvas. Alex can write and save to a valid path later.

**Scenario D — Permission denied:** Alex opens a file they can read but can't write to. When auto-save triggers, the status bar shows `E: Cannot save: permission denied` for three seconds. The document stays open, nothing is lost. Alex saves to a different path with `:w ~/my-copy.md`.

**Resolution:** In every scenario, the pattern is the same: errors communicated briefly in the status bar, the document never lost, the user never trapped. Recovery is automatic (auto-save) or one step away (re-open, save elsewhere).

**Requirements revealed:** Terminal resize handling, responsive column recalculation, re-rendering on resize, status bar error display with auto-dismiss, graceful file-not-found handling, permission error handling, `:w [path]` for save-to-alternate-path, auto-save as data loss prevention.

### Journey 4: Alex Discovers ink for the First Time

**Opening Scene:** Alex sees a post on a terminal tools forum: "glow lets you read markdown beautifully, ink lets you write it." They're intrigued. They've been using Neovim for writing blog posts but it never feels right — too much code-editor energy. They install ink with `go install` and type `ink`.

**Rising Action:** A blank screen. Centered cursor. A faintly dimmed bar at the bottom: `INSERT · 0w · 0c`. Alex starts typing — it's already in insert mode. They write a sentence. Then a heading with `# My First Test`. Then another paragraph. It feels like typing into a void, but a comfortable void. They notice the word count ticking up in the dimmed bar.

**Climax:** Curiosity hits. Alex presses `Esc`. The markdown transforms — the heading becomes bold and styled, the paragraph text sits clean and rendered, the syntax characters disappear. "Oh." Alex presses `i` on the heading — it reveals back to `# My First Test` with the `#` dimmed. They get it immediately: the document is always beautiful, and they reach into it to edit. The mental model clicks in ten seconds.

**Resolution:** Alex writes for twenty minutes, experimenting. They try `o` to create new blocks, `v` for visual selection, navigate with `j`/`k` through the rendered document. Everything behaves like vim but feels like a writing tool. When they quit, the save prompt appears, they save their test file. Alex opens their terminal config and aliases their writing command to ink. The next blog post will be written here.

**Requirements revealed:** Intuitive self-teaching through the block reveal model, vim conventions as expected behavior, visual mode selection across blocks, the "aha moment" of first Esc press, zero-config first experience.

### Journey Requirements Summary

| Capability | Revealed By |
|-----------|------------|
| Blank canvas / insert mode startup | Journey 1, 4 |
| Normal mode startup for existing files | Journey 2 |
| Block-level inline live preview (raw ↔ rendered) | Journey 1, 2, 4 |
| Cursor position mapping (rendered ↔ raw) | Journey 2 |
| Auto-save on typing pause | Journey 1, 2, 3 |
| Save-as prompt for unsaved buffers | Journey 1, 4 |
| Instant quit for named files | Journey 2 |
| Vim motions (normal, insert, visual) | Journey 1, 2, 4 |
| Centered responsive writing column | Journey 1, 3 |
| Terminal resize handling | Journey 3 |
| Status bar error display with auto-dismiss | Journey 3 |
| Graceful error recovery (file not found, permissions) | Journey 3 |
| `:w [path]` save-to-alternate-path | Journey 3 |
| Word/char count in status bar | Journey 1, 4 |
| Mode-aware UI (dimmed in insert, visible in normal) | Journey 1, 2, 4 |
| Multiple block type editing (paragraphs, lists, headings) | Journey 2 |
| Visual mode multi-block selection | Journey 4 |
| Self-teaching block model (no onboarding needed) | Journey 4 |

## Innovation & Novel Patterns

### Detected Innovation Areas

**1. Inline Live Preview in a TUI**
No existing terminal editor offers Obsidian-style block-level inline preview — where the entire document is rendered except the active editing block. Proven in GUI editors (Obsidian, Typora) but never implemented in a terminal environment. ink transplants this pattern into the TUI space using Glamour's rendering engine.

**2. Mode-Unified Block Reveal**
A novel UX pattern unifying vim modal editing with Obsidian's block-level preview. Entering insert mode simultaneously reveals raw markdown; returning to normal mode simultaneously renders the block. Single-gesture, dual-state-change interaction. Self-teaching — users discover it in the first 10 seconds without instruction.

**3. Writer-Centric Terminal UI**
Centered responsive-width prose, fading UI chrome in insert mode, and focus modes (typewriter, fade) create a distraction-free writing environment that rivals dedicated GUI writing apps — in the terminal. No terminal tool offers this level of writer-focused UI design.

### Market Context & Competitive Landscape

| Competitor | Strength | Limitation ink Addresses |
|-----------|----------|------------------------|
| Obsidian | Owns inline preview paradigm | GUI-only, knowledge-management baggage |
| Neovim + plugins | Powerful, extensible | Code-first identity, heavy configuration |
| glow/melt | Beautiful terminal markdown rendering | Read-only — no editing |
| WordGrinder | Terminal-native, writing-first, zero-config | Proprietary format, no markdown rendering |

**No direct competitor** combines: terminal-native + inline live preview + vim motions + writer-centric UI + zero-config.

### Validation Approach

- **Block transition feel:** Validated by achieving < 50ms transition with zero layout shift
- **Self-teaching pattern:** Validated when new users discover the block reveal model within 10 seconds without instruction
- **Writer preference:** Validated when the creator consistently chooses ink over Neovim and Obsidian within 1 month of stable release

## CLI/TUI Specific Requirements

### Project-Type Overview

ink is a full-screen, interactive TUI application — not a traditional CLI tool. No scriptable mode, no piping, no composability with other shell commands. The CLI surface is minimal; complexity lives entirely in the interactive TUI experience.

### Command Structure

- `ink` — Open blank canvas in insert mode
- `ink <file.md>` — Open existing file in normal mode
- `ink --version` — Print version and exit
- `ink --help` — Print usage and exit
- `--width <n>` — Override writing column width (CLI args take precedence over config file)

No subcommands. No flags beyond basic meta-flags and config overrides.

### Configuration Schema

**Format:** YAML (consistent with Charm ecosystem — glow uses `glow.yml`)
**Location:** `~/.config/ink/config.yml` (XDG conventions)

```yaml
# Writing column width (characters or percentage)
width: 80

# Focus modes
typewriter: false
fade: false

# Auto-save pause duration (milliseconds)
autosave_delay: 1000
```

**Principles:**
- Every value has a sensible default — config file is purely optional
- No config file created by default — zero-config out of the box
- Config loaded once at startup, not watched for changes
- Invalid values fall back to defaults silently

### File Handling

- `.md` files exclusively — no other formats
- No export, no conversion, no format transformation

### Technical Architecture

**Framework:** Bubbletea (Model-Update-View, event-driven, composable components)
**Rendering:** Lip Gloss (styling) + Glamour (markdown rendering, adaptive dark/light theme)
**Distribution:** Single binary, no runtime dependencies. Go cross-compilation for multi-platform builds. Targets: `go install`, Homebrew, AUR.

## Project Scoping & Phased Development

### MVP Strategy

**Approach:** Experience MVP — the minimum feature set that delivers the *feeling* of the invisible tool. Validation is qualitative: does a complete writing session pass without conscious thought about the tool?

**Resource reality:** Solo developer, evenings and weekends. Build order prioritized by risk. Each session produces a testable increment. The NOT-list protects the builder as much as the product.

### MVP Feature Set (Phase 1)

**Core journeys supported:** All four user journeys fully supported.

**Must-have capabilities:**

| Capability | Rationale |
|-----------|-----------|
| Block parser (markdown → block elements) | Foundation — every component depends on this |
| Glamour-rendered blocks | Document must display before it can be edited |
| Document viewport with centered column | Writing-focused layout |
| Editing block with syntax dimming | The defining experience |
| Block transitions (raw ↔ rendered) | Make-or-break interaction (< 50ms, zero layout shift) |
| Cursor position mapping (rendered ↔ raw) | Users must enter a block at the expected position |
| Three vim modes (normal, insert, visual) | Core interaction model |
| Status bar (mode + word/char count, dimming) | Mode awareness and writing feedback |
| Auto-save on typing pause | Zero data loss, zero friction |
| Context-aware startup (blank canvas or rendered file) | Zero-decision startup |
| Quit behaviors (instant, save-as, silent discard) | Complete session lifecycle |
| Command input (`:q`, `:w`, `:wq`, `ZZ`) | Essential vim commands |
| Undo/redo | Non-negotiable editing capability |
| Auto-pairs | Minimal auto-behavior for markdown |
| Full clipboard integration | Text moves freely in and out |
| Mouse support | Accessibility and exploration |
| Terminal resize handling | Responsive writing column |
| YAML config file (optional, XDG) | Zero-config defaults with optional tuning |

### Build Order (Risk-First)

1. **Block parser** — Pure logic, low risk, high dependency
2. **Rendered blocks + document viewport** — Validates rendering performance
3. **Editing block + block transitions** — **Make-or-break milestone. Validate < 50ms and zero layout shift before building anything else.**
4. **Vim mode system** — Mode-unified block reveal pattern
5. **Status bar** — Mode display, word/char count, dimming
6. **File I/O + auto-save** — Open, save, auto-save, quit behaviors
7. **Command input + save prompt** — `:q`, `:w`, `:wq`, `ZZ`, save-as
8. **Remaining MVP** — Undo/redo, auto-pairs, clipboard, mouse, config, error handling

**Critical milestone:** After step 3, the creator should be able to open a markdown file, see it rendered, enter a block, edit it, and see it render on exit. If this feels right, the product is viable.

### Post-MVP Features (Phase 2 — Growth)

- Focus modes (typewriter, fade)
- Tabs for multi-file editing
- Search (without replace)
- Toggleable file explorer
- Journal mode (`--journal` flag)
- Multi-file CLI (`ink file1.md file2.md`)
- Code block syntax highlighting inside markdown
- Unintrusive vim cheatsheet

**There is no Phase 3.** ink's vision is fully scoped. The Growth features ARE the complete product. The NOT-list is permanent:
- No plugins/extension system
- No multiple cursors
- No split panes
- No replace
- No spell check
- No themes/color customization
- No command palette
- No outline/structure view
- No session memory

### Risk Mitigation

**Technical risks:**
- **Block transition performance (highest risk)** — Mitigate by building and benchmarking first (build order step 3). Pre-render and cache Glamour output. If Glamour is too slow for inline use, investigate caching or partial rendering.
- **Cursor position mapping accuracy** — Build as a core primitive with exhaustive tests across all markdown element types. Inaccurate mapping breaks the mental model entirely.
- **Glamour limitations for inline editing** — Glamour is designed for full-document rendering, not block-level compositing. May need to adapt or extend. Investigate early.
- **Layout shift between raw and rendered states** — Fixed-width writing column constrains both states. Height differences expected; ensure surrounding blocks reposition without flicker.

**Resource risks:**
- **Solo developer, evenings/weekends** — Mitigate with tight scope (NOT-list), risk-first build order, and testable increments per session.
- **Burnout** — The project should be fun. If it stops being fun, the scope is too big.
- **Contingency:** MVP can be further reduced to: block parser + rendered blocks + editing block + basic vim modes + file I/O.

## Functional Requirements

### Document Rendering

- **FR1:** User can view a markdown document with all non-active blocks rendered via Glamour
- **FR2:** User can see rendered blocks update instantly when the active editing block is exited
- **FR3:** User can see the active editing block display raw markdown with syntax characters dimmed and content text at full brightness
- **FR4:** User can view all standard markdown elements rendered inline (headings, bold, italic, links, code spans, block quotes, lists, tables, horizontal rules, code fences)
- **FR5:** User can see the document rendered within a centered writing column that adapts to terminal width

### Block Editing

- **FR6:** User can enter a block for editing by pressing insert-initiating vim commands (`i`, `a`, `o`, `O`) or clicking with the mouse
- **FR7:** User can exit a block by pressing `Esc`, which renders the block and returns to normal mode
- **FR8:** User can create a new block by pressing Enter twice at the end of the current block, which renders the previous block and creates a new empty editing block
- **FR9:** User can edit any paragraph-level markdown element as a single block (paragraphs, headings, lists, code fences, block quotes, tables, horizontal rules)
- **FR10:** User can have their cursor position accurately mapped between rendered and raw markdown when entering and exiting a block
- **FR11:** User can see surrounding blocks remain stable (no layout shift) when a block transitions between raw and rendered states

### Vim Mode System

- **FR12:** User can operate in normal mode to navigate through a fully rendered document
- **FR13:** User can operate in insert mode to edit raw markdown within the active block
- **FR14:** User can operate in visual mode to select text across one or more blocks, which reveal raw markdown during selection
- **FR15:** User can navigate in normal mode using vim motions (`h`, `j`, `k`, `l`, `w`, `b`, `G`, `gg`, `Ctrl+d`, `Ctrl+u`)
- **FR16:** User can always return to normal mode by pressing `Esc` from any other mode

### Text Editing

- **FR17:** User can type text and have it inserted at the cursor position in insert mode
- **FR18:** User can delete text using backspace and delete keys
- **FR19:** User can undo and redo edits
- **FR20:** User can have matching characters auto-inserted for markdown pairs (`**`, `__`, `` ` ``, `[]`, `()`)
- **FR21:** User can use standard markdown indentation (Tab key)
- **FR22:** User can yank (copy) text to the system clipboard and paste from it using vim commands (`y`, `p`) and `Ctrl+V`

### File Management

- **FR23:** User can open ink with no arguments to get a blank canvas in insert mode
- **FR24:** User can open ink with a file path argument to view the file rendered in normal mode
- **FR25:** User can have their work auto-saved silently on a typing pause
- **FR26:** User can quit instantly with `:q` or `ZZ` — named files are already saved via auto-save, unsaved buffers with content prompt once for a file path, empty unsaved buffers are silently discarded
- **FR27:** User can save to a specific path using `:w <path>`
- **FR28:** User can save and quit using `:wq`
- **FR29:** User can only open and save `.md` files

### Status Bar & Feedback

- **FR30:** User can see the current vim mode displayed in a centered status bar (`NORMAL`, `INSERT`, `VISUAL`)
- **FR31:** User can see a live word count and character count in the status bar
- **FR32:** User can see the status bar at full visibility in normal and visual modes, and dimmed (~30%) in insert mode
- **FR33:** User can enter commands via the status bar when pressing `:` in normal mode
- **FR34:** User can see error messages displayed in the status bar with `E:` prefix, auto-dismissing after 3 seconds
- **FR35:** User can see the save-as prompt in the status bar when quitting with unsaved buffer content

### Terminal Adaptation

- **FR36:** User can have the writing column automatically recalculate and content reflow when the terminal is resized
- **FR37:** User can have rendered blocks re-render at the new width on terminal resize
- **FR38:** User can use ink at any terminal width, with centering disabled below 40 characters and full-width used instead
- **FR39:** User can use ink at any terminal height, with the status bar hidden below 5 rows to reclaim space for writing

### Mouse Support

- **FR40:** User can click on a rendered block to enter insert mode at the click position
- **FR41:** User can scroll the document using the mouse wheel
- **FR42:** User can click to position the cursor in normal mode

### Configuration

- **FR43:** User can use ink with zero configuration — all settings have sensible defaults
- **FR44:** User can optionally customize settings via a YAML config file at `~/.config/ink/config.yml`
- **FR45:** User can override config file values with CLI arguments
- **FR46:** User can configure writing column width
- **FR47:** User can have invalid config values silently fall back to defaults

## Non-Functional Requirements

### Performance

- **NFR1:** Startup time from command execution to usable writing state must be under 100ms
- **NFR2:** Block transitions (raw → rendered and rendered → raw) must complete in under 50ms
- **NFR3:** Keystroke-to-screen latency in insert mode must be imperceptible
- **NFR4:** Word count and character count updates must not cause perceptible delay during typing
- **NFR5:** Terminal resize must recalculate layout and re-render all visible blocks without perceptible delay
- **NFR6:** Scrolling through a rendered document must feel smooth with no rendering stutter
- **NFR7:** Documents of at least 10,000 words must remain performant across all operations

### Reliability

- **NFR8:** Auto-save must never fail silently — save failures must be communicated via status bar error
- **NFR9:** Auto-save must never corrupt file content — writes must be atomic (write to temp, rename)
- **NFR10:** ink must never crash in a way that loses unsaved content — panic recovery should attempt emergency save
- **NFR11:** Invalid markdown input must never cause a crash

### Accessibility

- **NFR12:** All dimmed UI elements must maintain minimum readable contrast against the 10 most common terminal color schemes
- **NFR13:** No information may be conveyed by color alone — vim mode via text label, errors via `E:` prefix, block state via content appearance
- **NFR14:** All functionality must be achievable via keyboard alone — mouse is optional
- **NFR15:** Rendered markdown blocks must output clean text suitable for screen reader parsing
- **NFR16:** The status bar must be positioned consistently for predictable screen reader access

### Compatibility

- **NFR17:** ink must function correctly in major terminal emulators: Kitty, Alacritty, WezTerm, iTerm2, GNOME Terminal, Windows Terminal
- **NFR18:** ink must function correctly within tmux and screen sessions
- **NFR19:** ink must support true color (24-bit), 256 color, and 16 color terminals with graceful visual degradation
- **NFR20:** ink must handle SSH sessions with varying terminal capabilities without crashing
- **NFR21:** ink must function on Linux, macOS, and Windows (via Windows Terminal)
