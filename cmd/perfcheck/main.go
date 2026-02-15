package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/editor"
	tea "charm.land/bubbletea/v2"
)

func main() {
	content, err := ioutil.ReadFile("testfile.md")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	blocks := block.Parse(content)
	e := editor.NewEditor("test.md", blocks)
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	fmt.Printf("Total blocks: %d\n", len(blocks))
	for i, b := range blocks {
		lineCount := len([]rune(b.Raw))
		fmt.Printf("Block %d: Type=%v, Size=%d chars\n", i, b.Type, lineCount)
	}

	// Test entering insert mode on first block (heading)
	fmt.Println("\n=== Testing insert mode on FIRST heading (block 0) ===")
	start := time.Now()
	e.Update(tea.KeyPressMsg{Code: 'i'})
	duration := time.Since(start)
	fmt.Printf("Time to enter insert mode: %v\n", duration)

	if duration > 100*time.Millisecond {
		fmt.Printf("⚠️  SLOW! Took %v (>100ms)\n", duration)
	} else {
		fmt.Printf("✓ Fast enough\n")
	}

	// Exit and test another block
	e.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	// Move to a later block
	for i := 0; i < 10; i++ {
		e.Update(tea.KeyPressMsg{Code: 'j'})
	}

	fmt.Println("\n=== Testing insert mode on LATER block ===")
	start = time.Now()
	e.Update(tea.KeyPressMsg{Code: 'i'})
	duration = time.Since(start)
	fmt.Printf("Time to enter insert mode: %v\n", duration)

	if duration > 100*time.Millisecond {
		fmt.Printf("⚠️  SLOW! Took %v (>100ms)\n", duration)
	} else {
		fmt.Printf("✓ Fast enough\n")
	}
}
