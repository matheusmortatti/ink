# Story 4.3: Panic Recovery and Emergency Save

Status: done

## Story

As a writer,
I want my document recovered even if ink crashes unexpectedly,
so that I never lose my writing regardless of what goes wrong.

## Acceptance Criteria

1. **Given** ink is running with document content **When** an unrecoverable panic occurs **Then** the `defer recover()` in main catches the panic and attempts to write the current document content to `~/.local/state/ink/recovery-{timestamp}.md` (NFR10)

2. **Given** an emergency save is attempted after a panic **When** the save succeeds **Then** the recovery file path is printed to stderr after the TUI is cleaned up

3. **Given** an emergency save is attempted after a panic **When** the save fails (e.g., the recovery directory cannot be created) **Then** ink exits with the panic information printed to stderr — best effort, no secondary panic

4. **Given** the `~/.local/state/ink/` directory does not exist **When** an emergency save is triggered **Then** the directory is created before writing the recovery file

5. **Given** invalid markdown input is provided to the parser or renderer **When** processing occurs **Then** ink handles the input gracefully without panicking (NFR11)

## Tasks / Subtasks

- [x] Task 1: Add exported `Serialize()` method to `internal/editor/editor.go` (AC: #1)
  - [x] 1.1 Add `func (e *EditorModel) Serialize() []byte { return e.serializeDocument() }` after the existing `CurrentMode()` method (~line 98)
  - [x] 1.2 No new tests needed — `serializeDocument()` is already covered; this is a thin export wrapper

- [x] Task 2: Add emergency save support to `internal/file/file.go` (AC: #1–#4)
  - [x] 2.1 Add unexported `emergencySaveTo(dir string, content []byte) (string, error)` — testable implementation
    - `os.MkdirAll(dir, 0755)` — create dir if not exist (AC #4)
    - Generate filename: `"recovery-" + time.Now().Format("20060102-150405") + ".md"`
    - Write via `os.WriteFile(path, content, 0644)` — use direct write, NOT atomic (emergency context, simpler)
    - Return `(path, nil)` on success, `("", err)` on failure — best effort, no secondary panic
  - [x] 2.2 Add exported `EmergencySave(content []byte) (string, error)` — entry point for main
    - Resolve state dir via `userStateDir()` (uses `$XDG_STATE_HOME` or `~/.local/state`; `os.UserStateDir` not in this Go build)
    - Call `emergencySaveTo(filepath.Join(stateDir, "ink"), content)`
  - [x] 2.3 Add import `"time"` to `internal/file/file.go`
  - [x] 2.4 Add tests in `internal/file/file_test.go`:
    - `TestEmergencySaveTo_WritesContent` — `emergencySaveTo(t.TempDir(), content)`, verify file exists with correct bytes
    - `TestEmergencySaveTo_CreatesDirectory` — pass `filepath.Join(t.TempDir(), "new/nested/dir")`, verify it's created
    - `TestEmergencySaveTo_ReturnsRecoveryPath` — verify returned path matches pattern `recovery-YYYYMMDD-HHMMSS.md`
    - `TestEmergencySaveTo_FailsOnUnwritableDir` — chmod 0555, verify error returned (skip if root)

- [x] Task 3: Add panic recovery to `cmd/ink/main.go` (AC: #1–#3)
  - [x] 3.1 After `e := editor.NewEditor(filePath, blocks)`, add defer before `p := tea.NewProgram(e)`:
    ```go
    defer func() {
        r := recover()
        if r == nil {
            return
        }
        // TUI is already stopped (panic unwinds bubbletea's cleanup)
        fmt.Fprintf(os.Stderr, "panic: %v\n", r)
        path, err := file.EmergencySave(e.Serialize())
        if err != nil {
            fmt.Fprintf(os.Stderr, "emergency save failed: %v\n", err)
        } else {
            fmt.Fprintf(os.Stderr, "recovery file: %s\n", path)
        }
        os.Exit(2)
    }()
    ```
  - [x] 3.2 No new imports required — `fmt`, `os`, and `file` package are already imported

- [x] Task 4: Document NFR11 (invalid markdown resilience) with a parser test (AC: #5)
  - [x] 4.1 Add `TestParser_InvalidInput_NoPanic` in `internal/block/parser_test.go`:
    - Table-driven: empty bytes, random binary garbage, malformed UTF-8, deeply nested `**`, unclosed fences
    - Call `block.Parse(input)` inside each case — verify no panic (test itself proves it)
    - This is documentation/regression: goldmark is already resilient; test confirms it

## Dev Notes

### Key Insight: `e *EditorModel` Pointer Is Stable Through `p.Run()`

**CRITICAL**: The `e *EditorModel` pointer created in `main()` before `tea.NewProgram(e)` is the SAME pointer throughout execution. Every `EditorModel.Update()` call mutates `e` in place and returns `e` (the same pointer). Bubbletea stores this pointer internally as the active model. Therefore, `e.Serialize()` inside the `defer recover()` in main always reads the current document state at time of panic.

No shared atomic/channel mechanism is needed — the pointer itself is the shared reference.

### Why Direct `os.WriteFile` Instead of Atomic Write for Emergency Save

Normal `WriteFile` uses temp-file + rename (atomic). For emergency saves:
- The process is already in a panic — complexity is a liability
- The risk of a half-written recovery file is acceptable (better than none)
- `os.WriteFile` is one call, `CreateTemp + Write + Close + Chmod + Rename` is five
- Use `os.WriteFile(path, content, 0644)` directly in `emergencySaveTo`

### Panic Recovery Scope: Main Goroutine Only

The `defer recover()` in `main()` catches panics that propagate to the main goroutine stack (e.g., panics inside `p.Run()` if Bubbletea's run loop is synchronous, or panics in Init/Update that escape Bubbletea's own recover). Panics in Bubbletea's background goroutines that escape their goroutine will crash the process without triggering main's defer — this is unavoidable without per-goroutine wrappers and is out of scope per the architecture spec.

### `os.UserStateDir()` Platform Behavior

| Platform | Returns          |
|----------|------------------|
| Linux    | `$XDG_STATE_HOME` or `~/.local/state` |
| macOS    | `$HOME/Library/Application Support` |
| Windows  | `%AppData%` |

Available since Go 1.13. If it returns an error (uncommon), fall back to `os.TempDir()` to avoid panicking inside the panic handler.

### `defer` Placement in `main()`

The defer must be registered AFTER `e := editor.NewEditor(...)` so that `e` is in scope within the closure:

```go
func main() {
    // ... arg parsing, file loading ...
    e := editor.NewEditor(filePath, blocks)

    defer func() {          // ← after e is set
        r := recover()
        // uses e.Serialize()
    }()

    p := tea.NewProgram(e)
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
```

### NFR11: goldmark Handles Invalid Input Without Panicking

goldmark (the Markdown parser underlying Glamour and `internal/block/parser.go`) is designed to be resilient — it parses any byte sequence as a valid (possibly empty/degenerate) AST without panicking. This is a goldmark property, not something we need to implement. Task 4 adds a regression test to document and confirm this.

### File Timestamp Format

Use `time.Now().Format("20060102-150405")` — produces `20260304-143022` style. This is:
- Filesystem-safe (no colons, slashes, or spaces)
- Lexicographically sortable
- Human-readable at a glance
- Avoids collision when multiple panics happen in the same session (seconds granularity is sufficient)

### Project Structure Notes

- **Files to modify:**
  - `internal/editor/editor.go` — add `Serialize()` exported method (~line 100)
  - `internal/file/file.go` — add `emergencySaveTo()` and `EmergencySave()` functions; add `"time"` import
  - `internal/file/file_test.go` — add 4 tests for emergency save
  - `cmd/ink/main.go` — add defer recover block
  - `internal/block/parser_test.go` — add NFR11 regression test
- **No new packages or files created**
- **No existing behavior changed** — all additions are new code paths (panic recovery) that don't execute in normal operation

### References

- [Source: epics.md#Story 4.3] — User story and all 5 acceptance criteria
- [Source: architecture.md#Error Handling & Recovery] — "Standard Go error handling + top-level panic recovery; single defer recover() in main.go"
- [Source: architecture.md#Panic Recovery] — "On panic: attempt emergency save to ~/.local/state/ink/recovery-{timestamp}.md; print path to stderr after TUI cleanup"
- [Source: architecture.md#Requirements to Structure Mapping] — NFR10 → cmd/ink/main.go; NFR11 → block/parser.go
- [Source: architecture.md#File I/O patterns] — `internal/file/file.go`: atomic writes (WriteFile); emergency save is non-atomic by design
- [Source: cmd/ink/main.go:14] — `main()` function where defer goes; `e` pointer, `file` package already imported
- [Source: internal/editor/editor.go:88] — `CurrentMode()` exported method (pattern for adding `Serialize()`)
- [Source: internal/editor/editor.go:~380] — `serializeDocument()` (wrapped by new Serialize())
- [Source: internal/file/file.go:28] — `WriteFile()` (atomic write pattern; emergency save does NOT use this)
- [Source: 4-1-auto-save-on-typing-pause.md] — established auto-save patterns; recovery is separate concern
- [Source: 4-2-quit-behaviors-and-save-commands.md] — established file write patterns; file.ValidatePath/WriteFile usage

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

`os.UserStateDir` not available in this Go build (custom 1.26.0-X); implemented `userStateDir()` manually using `$XDG_STATE_HOME` → `~/.local/state` fallback via `os.UserHomeDir()`.

### Completion Notes List

- Task 1: Added `Serialize()` exported method to `EditorModel` — thin wrapper over `serializeDocument()`, safe for use in defer recover.
- Task 2: Added `emergencySaveTo()` + `EmergencySave()` to `internal/file/file.go`; added `"time"` import; 4 unit tests all pass.
- Task 3: Added `defer recover()` block in `main()` after `e` is set, before `p := tea.NewProgram(e)`. Uses `e.Serialize()` + `file.EmergencySave()`. Exits with code 2 on panic.
- Task 4: Added `TestParser_InvalidInput_NoPanic` (8 table-driven cases) confirming goldmark handles all invalid input gracefully.
- All 12 new tests pass; no regressions.

### Senior Developer Review (AI)

**Review Date:** 2026-03-05
**Reviewer Model:** claude-opus-4-6
**Review Outcome:** Approve (with fixes applied)

#### Action Items

- [x] [HIGH] Wrap `e.Serialize()` in nested recover to prevent secondary panic — fixed in `cmd/ink/main.go`
- [x] [MEDIUM] Add tests for `EmergencySave()` and `userStateDir()` — 3 tests added to `internal/file/file_test.go`
- [x] [MEDIUM] Add microsecond suffix to recovery filename to prevent collision — fixed in `internal/file/file.go`
- [ ] [LOW] `userStateDir()` Linux-only XDG assumption — deferred (cross-platform not a priority)
- [ ] [LOW] NFR11 test only covers parser, not renderer — deferred (goldmark resilience is sufficient)

### File List

- `internal/editor/editor.go` — added `Serialize()` exported method
- `internal/file/file.go` — added `emergencySaveTo()`, `userStateDir()`, `EmergencySave()`; added `"time"` import; microsecond suffix for filename uniqueness
- `internal/file/file_test.go` — added 7 tests (4 emergency save + 3 review-fix tests)
- `cmd/ink/main.go` — added `defer recover()` panic handler with nested recover protection
- `internal/block/parser_test.go` — added `TestParser_InvalidInput_NoPanic` (8 cases)

## Change Log

- 2026-03-04: Implemented panic recovery and emergency save (story 4.3) — all 4 tasks complete, 12 new tests passing, no regressions
- 2026-03-05: Code review fixes — nested recover for secondary panic safety, added EmergencySave/userStateDir tests, filename microsecond suffix for uniqueness
