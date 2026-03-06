package main

import (
	"errors"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/editor"
	"github.com/matheusmortatti/ink/internal/file"
)

func main() {
	var filePath string
	var blocks []block.Block

	if len(os.Args) > 1 {
		filePath = os.Args[1]

		if err := file.ValidatePath(filePath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		content, err := file.ReadFile(filePath)
		if err != nil {
			if errors.Is(err, file.ErrFileNotFound) {
				// Missing file opens blank canvas (AC#2)
				blocks = nil
			} else {
				fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
				os.Exit(1)
			}
		} else {
			blocks = block.Parse(content)
		}
	}

	e := editor.NewEditor(filePath, blocks)

	defer func() {
		r := recover()
		if r == nil {
			return
		}
		// TUI is already stopped (panic unwinds bubbletea's cleanup)
		fmt.Fprintf(os.Stderr, "panic: %v\n", r)

		// Nested recover: e.Serialize() could panic if the editor state
		// is corrupted (e.g., gap buffer out of bounds). Best effort — no
		// secondary panic per architecture spec.
		var content []byte
		func() {
			defer func() { recover() }()
			content = e.Serialize()
		}()
		if content == nil {
			fmt.Fprintf(os.Stderr, "emergency save failed: could not serialize document\n")
			os.Exit(2)
		}

		path, err := file.EmergencySave(content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "emergency save failed: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "recovery file: %s\n", path)
		}
		os.Exit(2)
	}()

	p := tea.NewProgram(e)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
