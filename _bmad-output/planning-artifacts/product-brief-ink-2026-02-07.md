---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments:
  - brainstorming-session-2026-02-07.md
date: 2026-02-07
author: Matheusmortatti
---

# Product Brief: ink

## Executive Summary

ink is a free, open-source, terminal-native markdown editor built for writing, not coding. It fills a gap that no existing tool addresses: a distraction-free, beautiful prose writing experience in the terminal with Obsidian-style inline live preview, vim motions, and a writing environment designed to disappear so the writer can focus entirely on their words.

Born from the Charm ecosystem (Bubbletea, Lip Gloss, Glamour), ink is the natural complement to glow — where glow lets you *read* markdown beautifully in the terminal, ink lets you *write* it beautifully. It ships zero-config, markdown-only, and writing-first: open it and start typing.

---

## Core Vision

### Problem Statement

Terminal users who write prose — journal entries, essays, documentation, even books — have no dedicated writing tool that respects both their environment and their craft. The terminal excels as a text-based system for everything else, yet writing remains an afterthought served by tools designed for code.

### Problem Impact

Writers in the terminal today face a forced choice: configure a code editor (Neovim) into a passable writing tool through plugins and settings, leave the terminal entirely for GUI applications like Obsidian that come loaded with knowledge-management baggage, or settle for tools like WordGrinder that lack markdown support and use proprietary formats. Each option introduces friction — setup friction, in-the-moment distraction, or format lock-in — that pulls the writer away from the act of writing itself.

### Why Existing Solutions Fall Short

- **Neovim + plugins:** Powerful but code-first by design. Distributions like LazyVim optimize for development, not prose. Configuring a comfortable writing environment requires significant effort, and the result still *feels* like a code editor.
- **Obsidian:** The closest to the desired writing experience, but GUI-only and burdened with identity baggage — graph views, plugin ecosystems, knowledge management paradigms. It's daunting to start when all you want to do is write.
- **Glow / Melt:** Beautiful markdown rendering in the terminal, but read-only. No editing capability.
- **WordGrinder:** Terminal word processor, but uses its own file format by default and lacks markdown rendering. Not part of the modern terminal ecosystem.
- **General text editors:** None are dedicated to markdown for editing, treating it as just another file type rather than a first-class writing medium.

### Proposed Solution

ink is a focused, minimal TUI markdown editor that treats writing as its sole purpose. It features Obsidian-style inline live preview (rendered markdown everywhere except the block you're actively editing), vim motions for efficient text navigation, and a writer-centric environment — centered text, fading UI chrome, configurable focus modes — all designed around one principle: the default is just you and your words.

Built with Go and the Charm ecosystem (Bubbletea, Lip Gloss, Glamour), ink is fast, zero-config out of the box, and opens to a blank canvas ready for writing. It handles `.md` files exclusively, auto-saves on pause, and keeps all configuration outside the writing space. It is free and open source.

### Key Differentiators

1. **Writing-first identity:** Not a code editor with markdown support — a writing tool that happens to live in the terminal. Every default optimizes for prose, not programming.
2. **Inline live preview in the terminal:** No other terminal editor offers Obsidian-style block-level rendering where the document looks beautiful while you edit it.
3. **The invisible tool philosophy:** The UI disappears in insert mode. No line numbers, no git status, no file paths — just words on screen with a fading status bar. The tool competes on friction *removed*, not features added.
4. **Zero-config, zero-baggage:** Opens instantly to a blank canvas. No setup, no decisions, no plugin ecosystem to navigate. Works perfectly out of the box.
5. **Charm ecosystem native:** Built on the same rendering stack as glow, inheriting its visual quality. Familiar to the existing Charm community and positioned as the natural write-side complement.

## Target Users

### Primary Users

**Persona: Alex — The Terminal Writer**

Alex is a software developer, technical writer, or creative who spends most of their working hours in the terminal. They're comfortable with command-line tools, likely use tmux and vim/neovim daily, and consider the terminal their home environment. But Alex also writes prose — blog posts, journal entries, documentation, essays, maybe even long-form fiction — and has never found a writing tool that fits naturally into their workflow.

**Environment:** Terminal-centric. May use a tiling window manager, tmux sessions, and vim-based tools. But the persona extends beyond power users — anyone comfortable enough with a terminal to want to write in one.

**Motivations:**
- Wants to write more but gets pulled away by tool friction and distractions
- Values simplicity, speed, and focus above feature richness
- Prefers tools that work immediately without configuration or setup rituals
- Believes the terminal is the best text-based environment for everything — writing should be no exception

**Current Pain:**
- Opens Obsidian and gets distracted by its GUI, past notes, and the weight of its feature set — seconds of friction that break the creative impulse
- Uses Neovim for writing but it *feels* like a code editor no matter how much configuration is applied
- Has tried other terminal tools (WordGrinder, general text editors) but none treat markdown as a first-class writing medium
- The gap between "I want to write" and "I am writing" is too wide

**Success Vision:**
Types `ink`, sees a blank canvas, and starts writing. The tool is invisible. The words flow. When they're done, they close it and the file is saved. No decisions, no distractions, no friction.

**Note on Accessibility:** While vim motions are the primary interface, unintrusive cheatsheets for common commands ensure that terminal-comfortable users who aren't vim experts can start writing immediately. The tool meets writers where they are rather than gatekeeping behind vim mastery.

### Secondary Users

N/A — ink serves a single, focused user archetype across different writing use cases (journaling, blogging, documentation, essays, long-form). The variation is in *what* they write, not in *who* they are.

### User Journey

1. **Discovery:** Alex encounters ink through the Charm ecosystem community, a terminal tools recommendation thread, or a blog post about distraction-free writing tools. The pitch — "glow lets you read markdown beautifully, ink lets you write it" — immediately clicks.
2. **Onboarding:** `go install` or package manager install. Alex types `ink` in their terminal. A blank canvas appears. They start writing. There is no step three. An unintrusive cheatsheet is available for those less familiar with vim motions.
3. **Core Usage:** ink becomes the default response to the urge to write. Blog post idea? `ink post.md`. Journal entry? `ink --journal`. Quick thought? Just `ink`. The tool opens instantly, auto-saves on pause, and gets out of the way.
4. **Success Moment:** The first time Alex finishes a full writing session and realizes they never once thought about the tool — they only thought about their words. The invisible tool philosophy delivered.
5. **Long-term:** ink lives in Alex's terminal alongside their other daily tools. It's muscle memory. The writing habit sticks because the friction that previously killed it is gone.

## Success Metrics

### User Success

- **Writing frequency increases:** Users write more often than they did before ink — the friction reduction translates into actual output.
- **ink becomes the default:** When the urge to write hits, ink is the first tool reached for. Not Neovim, not Obsidian, not a GUI app — `ink`.
- **Invisible tool achieved:** Users complete writing sessions without consciously thinking about the tool. The measure of success is the absence of friction, not the presence of features.
- **Zero-to-writing speed:** The time between "I want to write" and "I am writing" is measured in seconds — type `ink`, start typing.

### Business Objectives

ink is a free, open-source personal project. Success is measured by personal utility first and community resonance second. There are no revenue, profitability, or commercial growth targets.

- **Personal utility (primary):** The creator uses ink as their default writing tool and writes more because of it.
- **Community resonance (secondary):** The project resonates with other terminal writers who share the same frustrations. Adoption is organic, not pursued.
- **Ecosystem recognition (aspirational):** ink earns a place in the Charm ecosystem as the natural complement to glow — the write-side to glow's read-side.

### Key Performance Indicators

| KPI | Measurement | Target |
|-----|-------------|--------|
| Personal daily use | Creator reaches for ink first when writing | Consistent habit within 1 month of stable release |
| GitHub stars | Star count on repository | Organic growth; no specific target |
| Package downloads | Downloads via common package managers (Homebrew, AUR, etc.) | Available in at least 2 package managers within 6 months |
| Community signal | Issues, discussions, and contributions from users | Any non-trivial external contribution signals product-market fit |
| Writing output | Volume of prose written in ink vs. other tools | ink accounts for majority of personal prose writing |

**Strategic note:** Vanity metrics (stars, downloads) are tracked as signals of resonance, not as success criteria. The true KPI is simple: does ink make its creator — and people like them — write more?

## MVP Scope

### Core Features

**Writing Environment:**
- Centered text with responsive width (comfortable reading column within terminal width)
- Blank canvas start — no arguments opens an empty, unsaved buffer
- Writer chrome bottom bar: vim mode + word count + character count (no line numbers, no file path, no git status)
- UI fades in insert mode, visible in normal mode
- Zero animations — all state changes snap instantly

**Core Editing:**
- Inline live preview (Obsidian-style block-level rendering — all blocks rendered except the active editing block which shows raw markdown with syntax highlighting)
- Three vim modes: normal, insert, visual
- Undo/redo
- Auto-pairs and standard markdown indentation
- Markdown-only (`.md` files exclusively)

**Focus Modes:**
- Typewriter mode (fixed cursor position, document scrolls)
- Fade mode (dim non-adjacent lines)
- Both configurable, off by default

**File Management:**
- Single-file editing (one file at a time)
- Auto-save on typing pause
- Instant quit — saves and exits immediately; unsaved buffers with content prompt once for file location; empty unsaved buffers silently discarded
- Writing-first: no arguments = blank unsaved buffer; write first, decide where to save later

**System Integration:**
- Full clipboard integration (vim yank writes to system clipboard, paste via `p` and Ctrl+V)
- Mouse support

**Configuration:**
- Zero-config by default — works perfectly out of the box with sensible defaults
- All customization via config file or CLI args (no settings UI inside the editor)
- Configurable writing width
- Glow/glamour default color palette

**Technical Foundation:**
- Built with Go and Bubbletea/Lip Gloss/Glamour (Charm ecosystem)
- Near-instant startup (comparable to vim)
- Snappy rendering — block transitions must feel instant

### Out of Scope for MVP

- Tabs / multi-file editing
- Toggleable file explorer
- Journal mode (`--journal` flag)
- Multi-file CLI opening (`ink file1.md file2.md`)
- Search
- Code block syntax highlighting inside markdown
- Rendering engine as separable Go package
- Unintrusive vim cheatsheet
- All items on the permanent NOT-list: plugins, multiple cursors, split panes, replace, spell check, themes/color customization, command palette, outline view, session memory

### MVP Success Criteria

The MVP is successful when the creator can open ink, write a full blog post or journal entry comfortably, save it, and close the tool — and chooses to do so instead of reaching for Neovim or Obsidian. The invisible tool philosophy is validated when a complete writing session passes without conscious thought about the tool itself.

### Future Vision

ink's full vision is already defined and deliberately bounded. The brainstorming session established the complete feature ceiling — the MVP is a subset, and the future roadmap adds the remaining features without expanding beyond the original scope:

**Post-MVP additions (the full product):**
- Tabs for multi-file editing (the only multi-file paradigm)
- Search (without replace)
- Toggleable file explorer (off by default, shortcut activated)
- Journal mode (`--journal` flag for zero-friction daily writing)
- Multi-file CLI (`ink file1.md file2.md` opens in tabs)
- Code block syntax highlighting inside markdown
- Unintrusive vim cheatsheet for accessibility

**Permanent ceiling:** ink does not grow beyond the scope defined in the brainstorming session. The NOT-list is permanent. The tool's identity is restraint — it competes on friction removed, not features added. The full product is the brainstorming scope, nothing more.
