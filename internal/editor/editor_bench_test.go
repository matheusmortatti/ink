package editor

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/matheusmortatti/ink/internal/block"
)

// BenchmarkEnterInsertMode_SingleLineHeading benchmarks entering insert mode on a short heading.
func BenchmarkEnterInsertMode_SingleLineHeading(b *testing.B) {
	blocks := block.Parse([]byte("# Short Heading"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		e.exitInsertMode() // Reset for next iteration
	}
}

// BenchmarkExitInsertMode_SingleLineHeading benchmarks exiting insert mode on a short heading.
func BenchmarkExitInsertMode_SingleLineHeading(b *testing.B) {
	blocks := block.Parse([]byte("# Short Heading"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		b.StartTimer()
		e.exitInsertMode()
		b.StopTimer()
	}
}

// BenchmarkEnterInsertMode_MultiLineParagraph benchmarks entering insert mode on a multi-line paragraph.
func BenchmarkEnterInsertMode_MultiLineParagraph(b *testing.B) {
	content := "This is a longer paragraph with multiple lines of text. " +
		"It contains enough content to wrap across several lines when rendered. " +
		"This helps benchmark the transition performance for typical editing scenarios. " +
		"The goal is to ensure transitions complete in under 50ms, ideally under 10ms."

	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		e.exitInsertMode()
	}
}

// BenchmarkExitInsertMode_MultiLineParagraph benchmarks exiting insert mode on a multi-line paragraph.
func BenchmarkExitInsertMode_MultiLineParagraph(b *testing.B) {
	content := "This is a longer paragraph with multiple lines of text. " +
		"It contains enough content to wrap across several lines when rendered. " +
		"This helps benchmark the transition performance for typical editing scenarios. " +
		"The goal is to ensure transitions complete in under 50ms, ideally under 10ms."

	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		b.StartTimer()
		e.exitInsertMode()
		b.StopTimer()
	}
}

// BenchmarkEnterInsertMode_LargeCodeFence benchmarks entering insert mode on a large code block.
func BenchmarkEnterInsertMode_LargeCodeFence(b *testing.B) {
	// Create a code fence with 50+ lines
	var codeLines []string
	codeLines = append(codeLines, "```go")
	for i := 0; i < 60; i++ {
		codeLines = append(codeLines, "func example() {")
		codeLines = append(codeLines, "    // This is a comment line")
		codeLines = append(codeLines, "    return nil")
		codeLines = append(codeLines, "}")
	}
	codeLines = append(codeLines, "```")

	content := strings.Join(codeLines, "\n")
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		e.exitInsertMode()
	}
}

// BenchmarkExitInsertMode_LargeCodeFence benchmarks exiting insert mode on a large code block.
func BenchmarkExitInsertMode_LargeCodeFence(b *testing.B) {
	// Create a code fence with 50+ lines
	var codeLines []string
	codeLines = append(codeLines, "```go")
	for i := 0; i < 60; i++ {
		codeLines = append(codeLines, "func example() {")
		codeLines = append(codeLines, "    // This is a comment line")
		codeLines = append(codeLines, "    return nil")
		codeLines = append(codeLines, "}")
	}
	codeLines = append(codeLines, "```")

	content := strings.Join(codeLines, "\n")
	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		b.StartTimer()
		e.exitInsertMode()
		b.StopTimer()
	}
}

// BenchmarkExitInsertMode_CachedPath benchmarks exiting without content changes (cached path).
func BenchmarkExitInsertMode_CachedPath(b *testing.B) {
	blocks := block.Parse([]byte("# Title\n\nParagraph content"))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		// Don't modify content - exit immediately to test cached path
		b.StartTimer()
		e.exitInsertMode()
		b.StopTimer()
	}
}

// BenchmarkFullTransitionCycle benchmarks the complete enter/exit cycle.
func BenchmarkFullTransitionCycle(b *testing.B) {
	content := "# Document Title\n\n" +
		"This is a multi-paragraph document with several blocks. " +
		"It simulates a realistic editing scenario.\n\n" +
		"```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```\n\n" +
		"Final paragraph with more content."

	blocks := block.Parse([]byte(content))
	e := NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.enterInsertMode("i")
		e.exitInsertMode()
	}
}
