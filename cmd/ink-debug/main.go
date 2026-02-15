package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/editor"
	"github.com/matheusmortatti/ink/internal/file"
	tea "charm.land/bubbletea/v2"
)

// Instrumented model that logs timing
type TimedModel struct {
	*editor.EditorModel
	logFile *os.File
}

func (m TimedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	start := time.Now()

	// Log the message type
	switch msg.(type) {
	case tea.KeyPressMsg:
		keyMsg := msg.(tea.KeyPressMsg)
		log.Printf("[UPDATE START] KeyPress: %c (code=%d)", keyMsg.Code, keyMsg.Code)
	case tea.WindowSizeMsg:
		log.Printf("[UPDATE START] WindowSizeMsg")
	}

	// Call the actual Update
	model, cmd := m.EditorModel.Update(msg)

	duration := time.Since(start)
	log.Printf("[UPDATE END] took %v", duration)

	if duration > 50*time.Millisecond {
		log.Printf("[WARNING] Update took %v - SLOW!", duration)
	}

	return TimedModel{model.(*editor.EditorModel), m.logFile}, cmd
}

func (m TimedModel) View() tea.View {
	start := time.Now()
	view := m.EditorModel.View()
	duration := time.Since(start)

	log.Printf("[VIEW] took %v", duration)

	if duration > 10*time.Millisecond {
		log.Printf("[WARNING] View took %v - SLOW!", duration)
	}

	return view
}

func main() {
	// Set up logging
	logFile, err := os.Create("/tmp/ink-debug.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	log.Println("=== INK DEBUG MODE STARTED ===")

	// Load file
	var blocks []block.Block
	var filePath string

	if len(os.Args) > 1 {
		filePath = os.Args[1]
		content, err := file.ReadFile(filePath)
		if err == nil {
			blocks = block.Parse(content)
			log.Printf("Loaded file: %s (%d blocks)", filePath, len(blocks))
		}
	}

	if blocks == nil {
		filePath = ""
		blocks = []block.Block{{Type: block.Paragraph, Raw: ""}}
		log.Println("Starting with empty document")
	}

	model := editor.NewEditor(filePath, blocks)
	timedModel := TimedModel{model, logFile}

	log.Println("Starting Bubbletea program...")
	p := tea.NewProgram(timedModel)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	log.Println("=== INK DEBUG MODE ENDED ===")
}
