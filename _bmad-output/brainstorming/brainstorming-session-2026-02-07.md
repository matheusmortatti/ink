---
stepsCompleted: [1, 2, 3, 4]
inputDocuments: []
session_topic: 'Focused, minimal TUI markdown editor with Obsidian-style inline preview, vim motions, built with Go/Bubbletea'
session_goals: 'Scope definition and competitive differentiation'
selected_approach: 'ai-recommended'
techniques_used: ['First Principles Thinking', 'SCAMPER Method', 'Reverse Brainstorming']
ideas_generated: 29
context_file: ''
session_active: false
workflow_completed: true
priority_theme: 'Writer-Centric UI'
---

# Brainstorming Session Results

**Facilitator:** Matheusmortatti
**Date:** 2026-02-07

## Session Overview

**Topic:** A focused, minimal TUI markdown editor — clean writing experience in the terminal with Obsidian-style inline preview, vim motions, built with Go/Bubbletea

**Goals:**
1. Scope definition — What features belong in this tool and what does NOT belong
2. Competitive differentiation — What makes this distinct from Neovim+plugins, Obsidian, glow, melt, and other terminal markdown tools

### Session Setup

The user identified a gap between Neovim (powerful but code-centric, complex setup, not visually appealing for prose) and Obsidian (beautiful for writing but GUI-only, feature-heavy). The desired tool sits at the intersection: terminal-native, distraction-free, beautiful for prose writing, with inline markdown preview and vim motions.

## Technique Selection

**Approach:** AI-Recommended Techniques
**Analysis Context:** TUI markdown editor with focus on scope definition and competitive differentiation

**Recommended Techniques:**

1. **First Principles Thinking:** Strip assumptions about what a markdown editor "must be" — rebuild from fundamental truths about writing prose in a terminal. Establishes the irreducible foundation for all scope decisions.
2. **SCAMPER Method:** Systematically dissect existing tools (Obsidian, Neovim, glow, melt) through 7 lenses to map what to borrow, eliminate, invert, and combine for a unique identity.
3. **Reverse Brainstorming:** "How would we make this terrible?" — reveals bloat to avoid, anti-patterns, and explicit exclusions that sharpen scope through a clear NOT-list.

**AI Rationale:** This sequence moves from foundational truths -> competitive analysis -> scope sharpening, ensuring the product identity is defined by what it IS, what makes it DIFFERENT, and what it deliberately IS NOT.

## Technique Execution Results

### First Principles Thinking

**Interactive Focus:** Stripped away assumptions about markdown editors to identify irreducible truths about writing prose in a terminal.

**Key Breakthroughs:**

- **The Invisible Tool (#1):** The tool's job is to get out of the way. Every design decision should be measured by "does this bring the experience closer to or further from the direct act of writing?" Most editors compete on features added. This tool competes on friction removed.
- **Markdown as Minimal Formatting (#2):** Markdown isn't chosen because it's powerful — it's chosen because it's the simplest structured formatting for digital text. Rendering is subtraction (hiding syntax noise), not addition (showing pretty formatting).
- **Interface is Preference, Not Feature (#3):** Vim motions are the initial interface choice but not the product's identity. The tool shouldn't evangelize its interface method — the product's soul is "a writing space," not "a vim editor."
- **Digital Equals Paper, Different Medium (#4):** Digital writing isn't inherently better than pen and paper. The tool doesn't justify its existence by adding capabilities — it justifies itself by being a clean, honest digital surface for writing.

**Foundational Insight:** Restraint is the defining design value. Reading and writing patterns are universal within a language — optimal visual layout for reading prose isn't a preference, it's closer to a constraint. The natural writing flow (thinking -> typing -> reviewing -> revising) must not be interrupted by the tool.

### SCAMPER Method

**Building on First Principles:** Used the four foundational principles as a scalpel to systematically evaluate existing tools across seven lenses.

#### S — Substitute
- **Inline Live Preview (#5):** Replace split-pane editor/preview with Obsidian-style inline rendering. All blocks rendered except the active editing block, which shows raw markdown with syntax highlighting. Proven GUI pattern transplanted to TUI.
- **Writer Chrome (#6):** Bottom bar shows vim mode + command input + word count + character count. No line numbers, file path, or git status. Fades in insert mode, visible in normal mode. The UI metadata declares "this is a writing tool."

#### C — Combine
- **Charm Ecosystem Member (#7):** Position as the missing Charm ecosystem piece — glow-quality rendering with inline editing. Use lipgloss/glamour for rendering, bubbletea for the framework.
- **Focus Modes (#8):** Configurable typewriter mode (fixed cursor, scrolling document) and fade mode (dim non-adjacent lines). Layered on the block-level editing model. Both configurable.
- **Toggleable File Explorer (#9):** Optional side panel file browser via shortcut (e.g., leader+e). Only relevant when opening a folder. Off by default.

#### A — Adapt
- **Adjustable Writing Width (#10):** Configurable column width via config file or CLI parameters. The terminal may be wide, but prose renders in a comfortable reading width.
- **Blank Canvas Start (#11):** New or empty files open as a blank canvas — just a cursor, nothing else.
- **Configuration Outside Writing (#12):** All tool customization lives in config files or CLI args. Zero settings UI inside the editor. The writing space is sacred.

#### M — Modify
- **Three Vim Modes (#13):** Support normal, insert, and visual modes only. Additional modes added only if a concrete gap is discovered. Vim-inspired, not vim-complete.

#### P — Put to Other Uses
- **Journal Mode (#14):** A `--journal` flag opens a new file named with the current date in a configured journal directory. No prompts, no decisions — zero-friction daily writing.
- **Rendering Engine as Library (#15):** Build the markdown rendering engine as a separable Go package. Inline live-preview rendering for bubbletea applications becomes a reusable community component.

#### E — Eliminate (The NOT-List)
- **Deliberate Exclusions (#16):**
  - Plugin/extension system
  - Multiple cursors
  - Split panes
  - Replace (search stays, replace goes)
  - Spell check (squiggly lines break visual flow)
  - Themes/color scheme customization (revisit later)
  - Command palette
  - Outline/structure view
  - Session memory (cursor position, scroll state)

- **Confirmed Inclusions (#17):**
  - Tabs (only multi-file paradigm)
  - Search (without replace)
  - Undo/redo
  - Code block syntax highlighting inside markdown
  - Auto-save
  - Mouse support

#### R — Reverse
- **Writing-First (#18):** Opening the tool with no arguments drops into a blank, unsaved buffer. Write first, decide where to save later. Most editors require a file decision before typing — this reverses that.

### Reverse Brainstorming

**Building on SCAMPER:** Stress-tested every scope decision by asking "how would we make this terrible?"

**Key Ideas from Stress Testing:**

- **Instant Quit with Auto-Save (#19):** Quitting saves and exits immediately — no prompt. Unsaved buffers with content prompt once for file location. Empty unsaved buffers silently discarded. Auto-save means no data loss risk.
- **Zero Animations (#20):** No transitions, no animations, no visual flourishes anywhere. Block rendering snaps. Mode changes update bottom bar text. Every frame of animation is friction.
- **Assume Vim Literacy (#21):** No onboarding, no help overlay, no tutorial. Documentation lives outside the tool (README, man page). Mouse support softens the curve for partial vim users.
- **Minimal Auto-Behavior (#22):** Auto-pairs and standard markdown indentation — yes. Auto-formatting, paragraph reflow, heading correction — no. The tool types what you type with only the lightest assists.
- **Save on Pause (#23):** Auto-save triggers on a brief typing pause. Tunable with user feedback. No manual save action needed.
- **Glow Standard Palette (#24):** Ship with glow/glamour default colors. One theme, done right. No decision paralysis.
- **Full Clipboard Integration (#25):** Vim yank writes to system clipboard. Paste via `p` (normal mode) and Ctrl+V. Text moves freely in and out.
- **Multi-File via CLI and Explorer (#26):** `ink file1.md file2.md` opens both in tabs. File explorer opens files in new tabs.
- **Markdown Only (#27):** Opens `.md` files exclusively. Not a general text editor.
- **Centered Writing, Responsive Width (#28):** Text centered in terminal by default. Width responsive to terminal size, configurable as percentage with comfortable cap.
- **Zero-Config by Default (#29):** Works perfectly out of the box with no config file. Every setting has a sensible default. Config is purely optional tuning.

**Performance Principles Established:** Startup must be near-instant (comparable to vim). Rendering must be snappy — the snap from raw-to-rendered when leaving a block must feel instant. These are non-negotiable for a terminal tool.

## Idea Organization and Prioritization

### Thematic Organization

**Theme 1: Design Philosophy (The Soul of the Tool)**
The principles that filter every decision:
- #1 The Invisible Tool
- #2 Markdown as Minimal Formatting
- #3 Interface is Preference, Not Feature
- #4 Digital Equals Paper, Different Medium

**Theme 2: Core Editing Model**
The fundamental interaction pattern:
- #5 Inline Live Preview
- #13 Three Vim Modes
- #20 Zero Animations
- #22 Minimal Auto-Behavior

**Theme 3: Writer-Centric UI (TOP PRIORITY)**
What the user actually feels — the heart of competitive differentiation:
- #6 Writer Chrome (fading bottom bar)
- #8 Focus Modes (typewriter + fade)
- #10 Adjustable Writing Width
- #28 Centered Writing, Responsive Width
- #11 Blank Canvas Start
- #21 Assume Vim Literacy

**Theme 4: File & Session Management**
How files enter and leave the tool:
- #18 Writing-First (no args = blank buffer)
- #26 Multi-File via CLI and Explorer
- #9 Toggleable File Explorer
- #17 Tabs
- #19 Instant Quit with Auto-Save
- #23 Save on Pause
- #27 Markdown Only

**Theme 5: Ecosystem & Identity**
Where the tool fits in the world:
- #7 Charm Ecosystem Member
- #24 Glow Standard Palette
- #25 Full Clipboard Integration
- #15 Rendering Engine as Library

**Theme 6: Bonus Modes**
Smart defaults for specific use cases:
- #14 Journal Mode

**Theme 7: The NOT-List**
Explicit exclusions:
- #16 No plugins, multiple cursors, split panes, replace, spell check, themes (for now), command palette, outline view, session memory

**Cross-Cutting Principles:**
- #12 Configuration Outside Writing
- #29 Zero-Config by Default

### Prioritization Results

- **Top Priority Theme:** Writer-Centric UI (Theme 3) — This is where the tool becomes unmistakably different from every terminal editor. No terminal editor today offers centered responsive-width prose + fading UI chrome in insert mode + typewriter/fade focus modes + inline live preview.
- **Foundation Required First:** Core Editing Model (Theme 2) — The inline live preview block editing is the technical prerequisite for everything in Theme 3.
- **Competitive Identity:** Ecosystem & Identity (Theme 5) — Positioning as the missing Charm ecosystem editor gives immediate community context and adoption path.

### Competitive Differentiation Summary

**What this tool is:** A terminal-native, distraction-free markdown writing tool for people who live in the terminal. The missing piece of the Charm ecosystem — glow lets you read markdown beautifully, this tool lets you write it beautifully.

**What this tool is NOT:** A code editor with markdown support. A Neovim replacement. A feature-rich note-taking system. An Obsidian clone.

**The unique combination no other tool offers:** Obsidian-style inline live preview + Charm ecosystem rendering quality + vim motions + writer-centric UI (fading chrome, focus modes, centered text) + zero-config simplicity — all in the terminal.

## Session Summary and Insights

**Key Achievements:**
- 29 specific scope decisions covering features, exclusions, and design principles
- A clear NOT-list that defines the product as much as the feature list
- A design philosophy ("the invisible tool") that serves as a decision filter for every future question
- Competitive positioning identified: the missing Charm ecosystem writing tool

**Session Reflections:**
The most powerful insight from this session was that the user's every instinct pointed toward restraint. When given the choice to add or remove, the answer was consistently "remove." This isn't a tool being built by subtraction from existing editors — it's being built from first principles upward, and the first principles demand simplicity. The Writer-Centric UI theme emerged as the priority because it's the tangible expression of the invisible tool philosophy — it's what makes the restraint feel intentional rather than incomplete.

## Recommended Next Steps

1. **Create Product Brief** — `/bmad-bmm-create-product-brief` in a fresh context window. Formalize this brainstorming session into the foundational product document.
2. **Create PRD** — `/bmad-bmm-create-prd` — Turn the brief into full product requirements.
3. **Create UX Design** — `/bmad-bmm-create-ux-design` — Strongly recommended given Theme 3's priority. Define the writer-centric UI in detail.
