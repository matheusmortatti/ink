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
	p := tea.NewProgram(e)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
