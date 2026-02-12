package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
)

func main() {
	content := []byte("# Test file\n\nThis is a test markdown file.\n\n```go\nfunc TestFile(m string) {\n  // do some stuff\n}\n```\n\n|| Column1 | Column2 | Column3 |\n| ------------- | -------------- | -------------- |\n| Item1 | Item1 | Item1 |\n||\n\nLorem ipsum dolor sit amet.")
	blocks := block.Parse(content)

	r, err := render.NewRenderer(80)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	ansiRe := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	for i, b := range blocks {
		rendered, err := r.Render(b)
		if err != nil {
			fmt.Printf("block %d error: %v\n", i, err)
			continue
		}
		fmt.Printf("=== Block %d (type=%s) ===\n", i, b.Type)
		fmt.Printf("Raw: %q\n", b.Raw)
		fmt.Printf("Rendered (raw bytes): %q\n", rendered)

		clean := ansiRe.ReplaceAllString(rendered, "")
		lines := strings.Split(clean, "\n")
		for j, line := range lines {
			fmt.Printf("  Line %d: %q (len=%d)\n", j, line, len([]rune(line)))
		}
		fmt.Println()
	}
}
