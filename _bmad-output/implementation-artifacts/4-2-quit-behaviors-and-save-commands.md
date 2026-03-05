# Story 4.2: Quit Behaviors and Save Commands

Status: done

## Story

As a writer,
I want predictable quit and save behaviors that match vim conventions,
so that ending a writing session is instant and my work is always preserved.

## Acceptance Criteria

1. **Given** the user is editing a named file with auto-save active **When** the user types `:q` or `ZZ` **Then** ink exits instantly — the file is already saved via auto-save (FR26)

2. **Given** the user has an unsaved buffer with content (no file path) **When** the user types `:q` or `ZZ` **Then** the save-as prompt appears in the status bar (FR26)

3. **Given** the user has an empty unsaved buffer (no content, no file path) **When** the user types `:q` or `ZZ` **Then** ink exits silently without any prompt (FR26)

4. **Given** the user is in any mode **When** the user types `:w <path>` where path ends in `.md` **Then** the document is saved to the specified path via atomic write, and the editor's working path updates to that path (FR27)

5. **Given** the user types `:w <path>` where path does not end in `.md` **When** the command is executed **Then** an error is displayed: `E: Only .md files supported` (FR29)

6. **Given** the user types `:wq` **When** the command is executed **Then** the document is saved (to existing path or prompting for path if unsaved) and ink exits (FR28)

7. **Given** the user types `:w` without a path on an unsaved buffer **When** the command is executed **Then** the save-as prompt appears for the user to specify a path

8. **Given** the user types `:w` without a path on a named file **When** the command is executed **Then** the file is saved to its existing path

## Tasks / Subtasks

- [x] Task 1: Add `:w <path>` handler in `executeCommand()` in `internal/editor/editor.go` (AC: #4, #5)
  - [x] 1.1 Add `case strings.HasPrefix(cmd, "w "):` branch after `case cmd == "wq":` (or in the logical order of cases)
  - [x] 1.2 Extract path: `path := strings.TrimSpace(cmd[2:])` (strip the `"w "` prefix)
  - [x] 1.3 Call `file.ValidatePath(path)` — if error: `return e, e.setErrorWithTimer("E: Only .md files supported")`
  - [x] 1.4 Call `file.WriteFile(path, e.serializeDocument())` — if error: `return e, e.setErrorWithTimer("E: " + err.Error())`
  - [x] 1.5 On success: set `e.filePath = path`, call `e.refreshStatusBar()`, return `e, nil`

- [x] Task 2: Add tests in `internal/editor/editor_test.go` (AC: #4, #5)
  - [x] 2.1 `TestEditor_WriteToPath_ValidMD` — type `:w <tempdir>/out.md`, verify file written with document content
  - [x] 2.2 `TestEditor_WriteToPath_NonMD_ShowsError` — type `:w foo.txt`, verify status bar shows `E: Only .md files supported`
  - [x] 2.3 `TestEditor_WriteToPath_UpdatesFilePath` — type `:w new.md`, verify `e.filePath == "new.md"` (or full temp path)

## Dev Notes

### Critical Insight: Most of Story 4.2 Is Already Implemented

**Story 4.1 (done) and Story 3.4 (done) together implemented all the quit/save behaviors except `:w <path>`:**

| AC | Status | Location |
|----|--------|----------|
| `:q` on named file → instant quit | ✅ Done | `editor.go executeCommand()` case `"q"` |
| `:q` on unsaved buffer with content → save-as | ✅ Done | `editor.go executeCommand()` case `"q"` |
| `:q` on empty buffer → silent quit | ✅ Done | `editor.go executeCommand()` case `"q"` |
| `ZZ` on named file → instant quit | ✅ Done | `editor.go applyAction()` case `QuitAction` |
| `ZZ` on unsaved buffer with content → save-as | ✅ Done | `editor.go applyAction()` case `QuitAction` |
| `ZZ` on empty buffer → silent quit | ✅ Done | `editor.go applyAction()` case `QuitAction` |
| `:wq` → write + quit | ✅ Done | `editor.go executeCommand()` case `"wq"` |
| `:w` on named file → write to existing path | ✅ Done | `editor.go executeCommand()` case `"w"` |
| `:w` on unsaved buffer → save-as prompt | ✅ Done | `editor.go executeCommand()` case `"w"` |
| **`:w <path>` → write to specific .md path** | ❌ **MISSING** | Needs Task 1 above |

**The only missing piece is the `:w <path>` command with path argument parsing.**

### Implementation Design

The `executeCommand()` function is in `internal/editor/editor.go` around line 507. Current switch cases (in order):
- `case cmd == "q":`
- `case cmd == "wq":`
- `case cmd == "w":`
- `case cmd == "":`
- `default:` (unknown command error)

Add the new case **before** `case cmd == "w":` to avoid ambiguity (or rely on Go's `strings.HasPrefix` check):

```go
case strings.HasPrefix(cmd, "w "):
    // :w <path> — write to specific path
    path := strings.TrimSpace(cmd[2:])
    if err := file.ValidatePath(path); err != nil {
        return e, e.setErrorWithTimer("E: Only .md files supported")
    }
    if err := file.WriteFile(path, e.serializeDocument()); err != nil {
        return e, e.setErrorWithTimer("E: " + err.Error())
    }
    e.filePath = path
    e.refreshStatusBar()
    return e, nil
```

**Important**: Use `"E: Only .md files supported"` specifically (not the `file.ValidatePath` error string which says `"not a markdown file: <path>"`). This matches the AC verbatim and the UX spec.

**`strings` import is already present** in `editor.go` — used elsewhere for string operations.

### Existing Pattern References

- `file.ValidatePath(path)` — already used in `attemptSaveAs()` at `editor.go:632` (validates save-as prompt path)
- `file.WriteFile(path, content)` — already used in `:w`, `:wq`, auto-save, and `attemptSaveAs()`
- `e.setErrorWithTimer(msg)` — already used throughout for status bar errors with 3s auto-dismiss
- `e.filePath = path` — already done in `attemptSaveAs()` at `editor.go:641` when save-as succeeds
- `e.refreshStatusBar()` — already done at end of `:w` handler

**The error string to use:** `"E: Only .md files supported"` — NOT `err.Error()` from `ValidatePath`. The AC specifies this exact message and the UX is clearer than the technical error.

### Already-Existing Tests (No Duplication Needed)

Tests that already cover the implemented behaviors (in `editor_test.go`):
- `TestEditor_CommandMode_TypeAndEnter_Quit` (line ~2308) — `:q` with named file
- `TestEditor_Quit_UnsavedContent_ShowsSavePrompt` (line ~2482) — `:q` with unsaved content
- `TestEditor_Quit_EmptyBuffer_QuitsDirectly` (line ~2508) — `:q` with empty buffer
- `TestEditor_Quit_NamedFile_QuitsDirectly` (line ~2533) — `:q` named file exits
- `TestEditor_ZZ_NamedFile_QuitsDirectly` (line ~2737) — ZZ named file
- `TestEditor_ZZ_UnsavedContent_ShowsSavePrompt` (line ~2757) — ZZ unsaved buffer
- `TestEditor_WriteQuit_NamedFile` (line ~2698) — `:wq` named file

Only write the 3 new tests for `:w <path>` (Task 2 above).

### Test Pattern

Use the established test pattern from this codebase:

```go
func TestEditor_WriteToPath_ValidMD(t *testing.T) {
    lipgloss.SetColorProfile(termenv.TrueColor)
    dir := t.TempDir()
    targetPath := filepath.Join(dir, "out.md")

    blocks := block.Parse([]byte("hello world"))
    e := NewEditor("", blocks)
    m, _ := e.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
    e = m.(*EditorModel)

    // Enter command mode and type :w <path>
    m, _ = e.Update(tea.KeyPressMsg{Code: ':', String: ":"})
    e = m.(*EditorModel)
    for _, ch := range targetPath {
        m, _ = e.Update(tea.KeyPressMsg{Code: ch, String: string(ch)})
        e = m.(*EditorModel)
    }
    e.Update(tea.KeyPressMsg{Code: tea.KeyEnter, String: "enter"})

    content, err := os.ReadFile(targetPath)
    if err != nil {
        t.Fatalf("file not written: %v", err)
    }
    if !strings.Contains(string(content), "hello world") {
        t.Errorf("saved content wrong, got: %q", content)
    }
}
```

**Note on command mode entry**: Looking at the existing test `TestEditor_CommandMode_TypeAndEnter_Quit`, commands are entered by passing `":"` first (as `InsertCharAction` or mode switch), then individual characters. Check the existing test at line ~2308 for the exact key sequence used.

**Alternative simpler approach** (used in `TestEditor_WriteQuit_NamedFile` at line ~2698): Call `executeCommand()` directly if it's accessible (package-level test `package editor`). Use the direct call approach:

```go
// Direct approach (package editor — same package as editor_test.go)
e.commandBuf = "w " + targetPath
m, _ := e.executeCommand()
```

Verify against the actual test at line ~2698 to confirm `executeCommand()` is called directly in tests.

### Files to Modify

- `internal/editor/editor.go` — add `case strings.HasPrefix(cmd, "w "):` in `executeCommand()` (~line 519, after the `case cmd == "wq":` block)
- `internal/editor/editor_test.go` — add 3 new tests

### Files NOT to Touch

- `internal/vim/command.go` — command parsing stays in editor; CommandHandler just passes `ExecuteCommandAction`
- `internal/file/file.go` — `ValidatePath` and `WriteFile` already handle everything needed
- All other packages

### Scope Confirmation

This story scope is **tiny**: one new switch case (~6 lines) + 3 test functions. The heavy lifting (`:q`, `ZZ`, `:wq`, `:w`, save-as prompt, atomic writes, `.md` validation) was all done in Stories 3.3, 3.4, and 4.1.

### Project Structure Notes

- **Files to modify:**
  - `internal/editor/editor.go` — `executeCommand()` function, add one case
  - `internal/editor/editor_test.go` — add 3 tests for `:w <path>`
- **Alignment with architecture:** `internal/vim/command.go` lists `:q, :w, :wq, :w <path> parsing` — the `:w <path>` was planned but not yet executed
- **No new packages, no new imports** (strings already imported in editor.go)

### References

- [Source: epics.md#Story 4.2] — User story and all 8 acceptance criteria
- [Source: prd.md#FR26] — Quit behaviors (instant, save-as, silent discard)
- [Source: prd.md#FR27] — Save to specific path with `:w <path>`
- [Source: prd.md#FR28] — Save and quit with `:wq`
- [Source: prd.md#FR29] — Only .md files supported
- [Source: architecture.md#Package Structure] — `file.go`: path validation (.md only); `command.go`: `:w <path>` parsing
- [Source: internal/editor/editor.go:507] — `executeCommand()` function (add new case here)
- [Source: internal/editor/editor.go:624] — `attemptSaveAs()` (template for save logic pattern)
- [Source: internal/file/file.go:16] — `ValidatePath()` (use for path validation, override error message)
- [Source: internal/file/file.go:26] — `WriteFile()` (atomic write, already works)
- [Source: 4-1-auto-save-on-typing-pause.md] — establishes test patterns, `package editor` for same-package access
- [Source: 3-4-error-display-and-save-as-prompt.md] — save-as prompt implementation (all the quit/save infrastructure)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

(none)

### Completion Notes List

- Added `case strings.HasPrefix(cmd, "w "):` in `executeCommand()` between the `"wq"` and `"w"` cases. Extracts path via `cmd[2:]`, validates with `file.ValidatePath`, writes with `file.WriteFile`, updates `e.filePath` and refreshes status bar on success. Uses hardcoded `"E: Only .md files supported"` per AC #5.
- Added 3 tests: `TestEditor_WriteToPath_ValidMD` (file written with correct content), `TestEditor_WriteToPath_NonMD_ShowsError` (status bar shows exact error), `TestEditor_WriteToPath_UpdatesFilePath` (filePath updated after write). All pass.
- All 6 pre-existing ACs (`:q`, `ZZ`, `:wq`, `:w`) were already implemented; only the `:w <path>` case was missing.

### File List

- `internal/editor/editor.go`
- `internal/editor/editor_test.go`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`

## Change Log

- 2026-03-04: Implemented `:w <path>` command — added `case strings.HasPrefix(cmd, "w "):` in `executeCommand()` with path extraction, `.md` validation, atomic write, filePath update. Added 3 tests covering valid write, non-.md error, and filePath update. All ACs satisfied.
- 2026-03-04: Code review fixes — added empty path guard with "E: No path specified" error, fixed discarded return value in TestEditor_WriteToPath_ValidMD, added 2 new tests (empty path error, write failure error). Updated File List to include sprint-status.yaml.
