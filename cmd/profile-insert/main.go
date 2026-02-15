package main

import (
	"fmt"
	"time"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/editor"
	tea "charm.land/bubbletea/v2"
)

func main() {
	content := []byte("# Simple Test")
	blocks := block.Parse(content)
	e := editor.NewEditor("test.md", blocks)

	// Initialize with window size
	t1 := time.Now()
	e.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	fmt.Printf("Init time: %v\n", time.Since(t1))

	// First View call (baseline)
	t2 := time.Now()
	view1 := e.View()
	fmt.Printf("View() in normal mode: %v\n", time.Since(t2))
	fmt.Printf("View content length: %d\n", len(fmt.Sprint(view1.Content)))

	// Enter insert mode
	fmt.Println("\n=== Entering insert mode ===")
	t3 := time.Now()
	e.Update(tea.KeyPressMsg{Code: 'i'})
	fmt.Printf("Update(i) time: %v\n", time.Since(t3))

	// View call after entering insert mode - THIS IS WHERE THE HANG MIGHT BE
	t4 := time.Now()
	view2 := e.View()
	fmt.Printf("View() in insert mode: %v\n", time.Since(t4))
	fmt.Printf("View content length: %d\n", len(fmt.Sprint(view2.Content)))

	if time.Since(t4) > 100*time.Millisecond {
		fmt.Printf("\n⚠️  FOUND THE HANG! View() took %v in insert mode\n", time.Since(t4))
	}

	// Try calling View() again
	t5 := time.Now()
	view3 := e.View()
	fmt.Printf("View() second call: %v\n", time.Since(t5))
	fmt.Printf("View content length: %d\n", len(fmt.Sprint(view3.Content)))
}
