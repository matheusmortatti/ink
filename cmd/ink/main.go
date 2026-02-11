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
