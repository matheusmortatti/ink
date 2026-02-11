# Story 1.1: Project Initialization and Structure

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a developer,
I want a properly initialized Go project with Bubbletea v2 dependencies and the Architecture-specified directory structure,
so that I have a solid foundation to build all ink components on.

## Acceptance Criteria

1. **Given** a fresh project directory **When** the initialization commands are run (`go mod init github.com/matheusmortatti/ink`, `go get` for all dependencies) **Then** `go.mod` contains Bubbletea v2, Lip Gloss v2, Bubbles v2, Glamour, goldmark, and yaml.v3

2. **Given** the project is initialized **When** the directory structure is created **Then** all `internal/` packages exist (`block`, `render`, `vim`, `ui`, `file`, `config`, `editor`) and `cmd/ink/main.go` exists

3. **Given** `cmd/ink/main.go` exists with a minimal Bubbletea program **When** `go run ./cmd/ink` is executed **Then** a Bubbletea program starts and can be exited with `Ctrl+C` **And** `go build ./cmd/ink` produces a single binary without errors

## Tasks / Subtasks

- [x] Task 1: Initialize Go module (AC: #1)
  - [x] 1.1 Run `go mod init github.com/matheusmortatti/ink`
  - [x] 1.2 Run `go get charm.land/bubbletea/v2@latest`
  - [x] 1.3 Run `go get github.com/charmbracelet/lipgloss/v2@latest`
  - [x] 1.4 Run `go get github.com/charmbracelet/bubbles/v2@latest` (or `charm.land/bubbles/v2@latest`)
  - [x] 1.5 Run `go get github.com/charmbracelet/glamour@latest`
  - [x] 1.6 Run `go get github.com/yuin/goldmark@latest`
  - [x] 1.7 Run `go get gopkg.in/yaml.v3`
  - [x] 1.8 Verify `go.mod` lists all dependencies correctly
- [x] Task 2: Create project directory structure (AC: #2)
  - [x] 2.1 Create `cmd/ink/` directory
  - [x] 2.2 Create `internal/block/` directory
  - [x] 2.3 Create `internal/render/` directory
  - [x] 2.4 Create `internal/vim/` directory
  - [x] 2.5 Create `internal/ui/` directory
  - [x] 2.6 Create `internal/file/` directory
  - [x] 2.7 Create `internal/config/` directory
  - [x] 2.8 Create `internal/editor/` directory
- [x] Task 3: Create minimal Bubbletea application (AC: #3)
  - [x] 3.1 Create `cmd/ink/main.go` with minimal Bubbletea v2 program (empty model, Init/Update/View, Ctrl+C exit)
  - [x] 3.2 Create placeholder `.go` files in each `internal/` package (package declaration only, to satisfy Go tooling)
  - [x] 3.3 Verify `go run ./cmd/ink` starts and exits cleanly with Ctrl+C
  - [x] 3.4 Verify `go build ./cmd/ink` produces a binary without errors
  - [x] 3.5 Run `go vet ./...` to confirm no issues

## Dev Notes

### Technical Requirements

- **Go version:** Go 1.26 (latest stable as of Feb 2026)
- **Module path:** `github.com/matheusmortatti/ink` — this is the canonical import path for all packages
- **Bubbletea v2 module path:** `charm.land/bubbletea/v2` — NOT the old `github.com/charmbracelet/bubbletea` path. Bubbletea v2 moved to the `charm.land` domain.
- **Bubbletea v2 status:** RC-2 (v2.0.0-rc.2). API is stable. Pin exact version in go.mod.
- **Lip Gloss v2 module path:** `github.com/charmbracelet/lipgloss/v2`
- **Bubbles v2 module path:** `charm.land/bubbles/v2` — also moved to `charm.land` domain
- **Glamour module path:** `github.com/charmbracelet/glamour` (v0.10.0, no v2 yet)
- **goldmark module path:** `github.com/yuin/goldmark` (v1.7.12)
- **yaml.v3 module path:** `gopkg.in/yaml.v3` (v3.0.1)
- **Bubbletea v2 API changes from v1:**
  - The Cursed Renderer is now the default and only renderer
  - Mouse messages are split: `tea.MouseClickMsg`, `tea.MouseReleaseMsg`, `tea.MouseWheelMsg`, `tea.MouseMotionMsg`
  - Several message types changed from type aliases to structs
  - `WithKeyboardEnhancements` split into `WithKeyReleases`, `WithUniformKeyLayout`, `RequestKeyDisambiguation`
- **CRITICAL:** Do NOT use the old `github.com/charmbracelet/bubbletea` import path. The v2 module lives at `charm.land/bubbletea/v2`.
- **CRITICAL:** Do NOT add any business logic in this story. The minimal Bubbletea program should be the absolute minimum: an empty model struct, `Init()` returning `nil`, `Update()` handling only `tea.KeyMsg` for Ctrl+C quit, and `View()` returning a simple placeholder string.
- **CRITICAL:** Placeholder files in `internal/` packages must contain ONLY the `package` declaration (e.g., `package block`). Do NOT add types, functions, or imports that aren't needed yet — future stories will build these properly.

### Architecture Compliance

**Dependency Direction (MUST follow — no exceptions):**

```
cmd/ink → internal/editor
internal/editor → internal/block, internal/render, internal/vim, internal/ui, internal/file, internal/config
internal/render → internal/block
internal/vim → internal/block
internal/ui → internal/block
internal/file → internal/block
internal/config → (no internal dependencies — leaf package)
internal/block → (no internal dependencies — leaf package)
```

- `block` and `config` are leaf packages — they MUST NOT import any other `internal/` package
- `editor` is the root — it imports everything, nothing imports it
- No horizontal imports between sibling packages (e.g., `vim` cannot import `render`)
- This story establishes the skeleton. Future stories MUST respect these boundaries.

**Package Naming Rules (enforce from day one):**
- Packages use singular, lowercase, one-word names: `block`, `render`, `vim`, `file`, `config`, `editor`, `ui`
- No `utils`, `helpers`, `common`, or `misc` packages ever
- Package names should not stutter: `block.Block` is fine, `block.BlockParser` is not — use `block.Parser`

**Placeholder File Naming:**
- Each `internal/` package gets exactly ONE placeholder file named after the package's primary responsibility:
  - `internal/block/block.go` → `package block`
  - `internal/render/renderer.go` → `package render`
  - `internal/vim/mode.go` → `package vim`
  - `internal/ui/viewport.go` → `package ui`
  - `internal/file/file.go` → `package file`
  - `internal/config/config.go` → `package config`
  - `internal/editor/editor.go` → `package editor`

**main.go Pattern (Architecture-mandated):**
- `cmd/ink/main.go` is the entry point — thin, delegates to `internal/editor`
- In this story, `main.go` creates a minimal Bubbletea program directly (editor package not wired yet)
- Future stories will move the model to `internal/editor/editor.go` and have `main.go` call it
- Include a `defer recover()` placeholder comment for future panic recovery (Story 4.3)

**Anti-Patterns to Avoid in This Story:**
- Do NOT create a `README.md` (not requested)
- Do NOT create `.github/workflows/` or `.goreleaser.yml` (Story 7.3)
- Do NOT create `.golangci.yml` (Story 7.3)
- Do NOT create `LICENSE` (not in scope)
- Do NOT add any type definitions, interfaces, or functions beyond the minimal Bubbletea model
- Do NOT create test files (nothing to test yet)

### Library & Framework Requirements

**Complete dependency table (exact versions and import paths):**

| Library | Module Path | Version | Purpose |
|---------|-------------|---------|---------|
| Bubbletea v2 | `charm.land/bubbletea/v2` | v2.0.0-rc.2 | TUI framework (Model-Update-View) |
| Lip Gloss v2 | `github.com/charmbracelet/lipgloss/v2` | v2.0.0-beta.2 | Component styling, color interpolation |
| Bubbles v2 | `charm.land/bubbles/v2` | v2.0.0-beta.1+ | TUI components (viewport, textinput) |
| Glamour | `github.com/charmbracelet/glamour` | v0.10.0 | Markdown rendering (glow engine) |
| goldmark | `github.com/yuin/goldmark` | v1.7.12 | CommonMark-compliant markdown AST parsing |
| yaml.v3 | `gopkg.in/yaml.v3` | v3.0.1 | YAML config file loading |

**Go get commands (run in this exact order):**
```bash
go mod init github.com/matheusmortatti/ink
go get charm.land/bubbletea/v2@latest
go get github.com/charmbracelet/lipgloss/v2@latest
go get charm.land/bubbles/v2@latest
go get github.com/charmbracelet/glamour@latest
go get github.com/yuin/goldmark@latest
go get gopkg.in/yaml.v3
```

**Important notes on Bubbletea v2 minimal program:**
- Import as: `tea "charm.land/bubbletea/v2"`
- Model must implement: `Init() tea.Cmd`, `Update(msg tea.Msg) (tea.Model, tea.Cmd)`, `View() tea.View`
- Start with: `p := tea.NewProgram(model)` then `p.Run()`
- Ctrl+C handling: check for `tea.KeyMsg` where `msg.String() == "ctrl+c"` and return `tea.Quit`
- The Cursed Renderer is automatic in v2 — no renderer configuration needed

**WARNING — Common LLM mistakes with Charm v2:**
- Using `github.com/charmbracelet/bubbletea` instead of `charm.land/bubbletea/v2`
- Using `github.com/charmbracelet/bubbles` instead of `charm.land/bubbles/v2`
- Using v1 API patterns (e.g., `tea.MouseMsg` instead of the split `tea.MouseClickMsg`/`tea.MouseWheelMsg`)
- Trying to configure the renderer (v2 only has the Cursed Renderer, no selection needed)
- Importing libraries not yet needed (only `charm.land/bubbletea/v2` is used in `main.go` for this story)

### File Structure Requirements

**Exact files to create in this story (nothing more, nothing less):**

```
ink/
├── cmd/
│   └── ink/
│       └── main.go              # Minimal Bubbletea v2 program
├── internal/
│   ├── block/
│   │   └── block.go             # package block
│   ├── render/
│   │   └── renderer.go          # package render
│   ├── vim/
│   │   └── mode.go              # package vim
│   ├── ui/
│   │   └── viewport.go          # package ui
│   ├── file/
│   │   └── file.go              # package file
│   ├── config/
│   │   └── config.go            # package config
│   └── editor/
│       └── editor.go            # package editor
├── go.mod                        # Generated by go mod init + go get
└── go.sum                        # Generated by go get
```

**Total files created by developer: 8** (1 main.go + 7 placeholder .go files)
**Total files auto-generated: 2** (go.mod + go.sum)

**Files NOT to create:**
- No `README.md`, `LICENSE`, `.gitignore`
- No `.github/` directory or CI files
- No `.goreleaser.yml` or `.golangci.yml`
- No test files (`*_test.go`)
- No additional `.go` files beyond the 8 listed above

**Placeholder file content pattern (identical for all 7 internal packages):**
```go
package <packagename>
```
That's it. One line. No imports, no comments, no types, no functions.

**main.go content scope:**
```go
package main

import (
    "fmt"
    "os"

    tea "charm.land/bubbletea/v2"
)

// model is a minimal Bubbletea model for project initialization verification.
type model struct{}

func (m model) Init() tea.Cmd {
    return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "ctrl+c" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() tea.View {
    return tea.NewView("ink - press ctrl+c to exit\n")
}

func main() {
    // TODO: panic recovery (Story 4.3)
    p := tea.NewProgram(model{})
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

**NOTE:** The above is the reference implementation. The dev agent should write code matching this pattern exactly. Do NOT add altscreen mode, mouse support, or any other program options — those come in later stories.

### Testing Requirements

**This story has NO unit tests.** There is no logic to test — only project scaffolding.

**Verification is done via build tooling instead:**

| Verification | Command | Expected Result |
|---|---|---|
| Module compiles | `go build ./cmd/ink` | Produces binary, zero errors |
| Program runs | `go run ./cmd/ink` | Starts Bubbletea program, shows message, exits on Ctrl+C |
| Code is clean | `go vet ./...` | Zero warnings |
| All packages valid | `go build ./...` | All packages compile (even placeholder ones) |
| Dependencies resolve | `go mod tidy` | **Known:** tidy removes 5 unused indirect deps (lipgloss, bubbles, glamour, goldmark, yaml.v3). Do NOT run tidy — deps were added via `go get` for AC#1. Future stories will import them, making them permanent. |

**Why no tests:**
- Placeholder packages have no code to test
- `main.go` is a minimal Bubbletea bootstrap — testing TUI programs requires Bubbletea test helpers which are out of scope
- Architecture mandates: "No TUI integration tests in MVP — test components via their Go interfaces" [Source: architecture.md#Testing Patterns]
- Test files begin in Story 1.2 (`internal/block/parser_test.go`) when there's actual logic to test

**Future testing patterns (for awareness, NOT for implementation now):**
- Co-located test files: `internal/block/parser_test.go` alongside `internal/block/parser.go`
- Table-driven tests for functions with multiple input/output cases
- Naming: `TestFunctionName_Scenario_ExpectedBehavior`
- Go's built-in `testing` package only — no external test frameworks

### Latest Tech Information (Feb 2026)

**Go Language:**
- Go 1.26 released Feb 4, 2026 (latest stable)
- Go 1.24.13 is latest patch in 1.24 series
- New features in 1.26 include `crypto/hpke`, experimental `simd/archsimd`
- None of these new features are relevant to ink — standard Go modules work fine

**Bubbletea v2 (charm.land/bubbletea/v2):**
- Still at RC-2 (v2.0.0-rc.2) as of Feb 2026 — stable release imminent but not yet shipped
- API is considered stable at RC stage — breaking changes unlikely
- The Cursed Renderer is the default and only renderer — provides better altscreen control and cursor visibility
- Mode 2026 (Synchronized Output) support added in RC-2
- **Risk:** Low. Pin to `v2.0.0-rc.2` in go.mod. If v2.0.0 stable releases during development, upgrade is expected to be zero-change.

**Lip Gloss v2 (github.com/charmbracelet/lipgloss/v2):**
- Still in beta (v2.0.0-beta.2, June 2025)
- Requires Go 1.23.0+
- v2 breaking changes: styles are deterministic, no hex/int color format, manual color downsampling
- **Risk:** Medium-low. Beta but actively used with Bubbletea v2. API should be stable enough.

**Bubbles v2 (charm.land/bubbles/v2):**
- Still in beta
- v0.21.0 (latest v1) released Feb 3, 2026 with new viewport horizontal scrolling
- **Risk:** Low for this story (not used yet). Monitor for stable v2 release.

**Glamour (github.com/charmbracelet/glamour):**
- v0.10.0 (April 2025) — latest stable
- New in v0.10.0: links/images rendered at table footer with reference numbers
- `WithInlineTableLinks` option available for old behavior
- **Risk:** None for this story. Relevant for Story 1.3 when Glamour rendering begins.

**goldmark (github.com/yuin/goldmark):**
- v1.7.12 (May 2025) — latest stable, actively maintained
- CommonMark compliant, extensible
- **Risk:** None. Mature and stable library.

### Project Structure Notes

- This is story 1 of 7 in Epic 1 ("View and Navigate a Beautiful Document")
- The project is fully greenfield — no existing Go code, no go.mod, no directories
- The git repo contains only BMAD planning artifacts (`_bmad/`, `_bmad-output/`)
- All code will live under the project root `/home/matheusmortatti/git/ink/`
- The `_bmad-output/` directory is for planning artifacts only and should NOT be mixed with source code
- No conflicts or variances detected — this is the very first code story

### References

- [Source: architecture.md#Starter Template Evaluation] — Initialization commands, project structure, dependency list
- [Source: architecture.md#Project Structure & Boundaries] — Complete directory structure with file purposes
- [Source: architecture.md#Package Boundary Rules] — Dependency direction diagram, what goes where
- [Source: architecture.md#Go Naming Patterns] — Package naming, exports, receivers, constants
- [Source: architecture.md#Implementation Patterns & Consistency Rules] — Anti-patterns to avoid
- [Source: architecture.md#Testing Patterns] — No tests in this story, co-located pattern for future
- [Source: epics.md#Story 1.1] — Acceptance criteria and user story statement
- [Source: prd.md#Technical Architecture] — Bubbletea framework, Lip Gloss, Glamour, distribution requirements
- [Source: architecture.md#Architecture Validation Results] — "First Implementation Priority" confirms init as step 1

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- Bubbletea v2 API change discovered: `View()` returns `tea.View` instead of `string`. Fixed by using `tea.NewView()` wrapper. The story Dev Notes reference implementation was based on v1 API pattern.
- `go mod tidy` removes unused indirect dependencies. After tidy, only `charm.land/bubbletea/v2` remained as direct dep. Re-ran `go get` for all 6 deps to satisfy AC #1 requirement that all deps are in go.mod. Note: running `go mod tidy` again will remove unused deps until future stories import them.

### Completion Notes List

- Ultimate context engine analysis completed — comprehensive developer guide created
- Story scope is deliberately minimal: scaffolding only, no logic, no tests
- Next story (1.2: Markdown Block Parser) depends on this structure being in place
- All 6 dependencies must be in go.mod before Story 1.2 can begin
- All 3 tasks and all subtasks completed successfully
- `go build ./cmd/ink` produces binary without errors
- `go vet ./...` returns zero warnings
- `go build ./...` compiles all packages (including placeholders)
- Bubbletea v2 API required `View() tea.View` instead of `View() string` — adapted using `tea.NewView()`
- Actual dependency versions resolved: bubbletea v2.0.0-rc.2, lipgloss v2.0.0-beta1, bubbles v2.0.0-rc.1, glamour v0.10.0, goldmark v1.7.16, yaml.v3 v3.0.1

### Change Log

- 2026-02-10: Story created by create-story workflow with exhaustive artifact analysis
- 2026-02-10: Implementation complete — Go module initialized, directory structure created, minimal Bubbletea v2 program built and verified
- 2026-02-10: Code review (AI) — 7 issues found (2H, 3M, 2L). Fixed: .gitignore corrected (stale project name, missing binary exclusion), build artifact removed, reference implementation updated to v2 API (View() tea.View), verification table updated to document known go mod tidy behavior

### File List

**Files created:**
- `cmd/ink/main.go`
- `internal/block/block.go`
- `internal/render/renderer.go`
- `internal/vim/mode.go`
- `internal/ui/viewport.go`
- `internal/file/file.go`
- `internal/config/config.go`
- `internal/editor/editor.go`

**Files modified:**
- `.gitignore` (updated: replaced stale `markdown-term` with `/ink`, added build binary exclusion)

**Files auto-generated:**
- `go.mod`
- `go.sum`
