package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
)

type model struct {
	textarea textarea.Model
	err      error
}

func initialModel() model {
	ti := textarea.New()
	ti.SetVirtualCursor(true)
	ti.Focus()

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.RequestBackgroundColor)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case "ctrl+c":
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() tea.View {
	const (
		footer = "\nPress Ctrl+C to quit.\n"
	)

	var c *tea.Cursor

	if !m.textarea.VirtualCursor() {
		c = m.textarea.Cursor()
	}

	f := strings.Join([]string{
		m.textarea.View(),
		footer,
	}, "\n")

	v := tea.NewView(f)
	v.Cursor = c
	return v
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
