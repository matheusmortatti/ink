package main

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// QuitMsg signals the program should exit.
type QuitMsg struct{}

// SaveMsg signals the file was saved.
type SaveMsg struct{ Err error }

// OpenFileMsg signals a file should be opened.
type OpenFileMsg struct{ Path string }

// StatusMsgMsg sets a temporary status message.
type StatusMsgMsg struct{ Text string }

func executeExCommand(cmd string, buf *Buffer, vim *VimState) tea.Cmd {
	cmd = strings.TrimSpace(cmd)

	// Line number: :<number>
	if n, err := strconv.Atoi(cmd); err == nil && n > 0 {
		buf.SetCursor(n-1, buf.FirstNonBlank(clampInt(n-1, 0, buf.LineCount()-1)))
		return nil
	}

	switch cmd {
	case "w":
		return func() tea.Msg {
			err := buf.SaveFile()
			if err != nil {
				return StatusMsgMsg{Text: "Error: " + err.Error()}
			}
			return SaveMsg{Err: nil}
		}
	case "q":
		if buf.dirty {
			vim.StatusMsg = "No write since last change (add ! to override)"
			return nil
		}
		return tea.Quit
	case "q!":
		return tea.Quit
	case "wq", "x":
		return func() tea.Msg {
			err := buf.SaveFile()
			if err != nil {
				return StatusMsgMsg{Text: "Error: " + err.Error()}
			}
			return QuitMsg{}
		}
	default:
		// :e <path>
		if strings.HasPrefix(cmd, "e ") {
			path := strings.TrimSpace(cmd[2:])
			return func() tea.Msg {
				return OpenFileMsg{Path: path}
			}
		}

		// /<search> from command mode (when : was typed first)
		if strings.HasPrefix(cmd, "/") {
			vim.SearchQuery = cmd[1:]
			vim.SearchDir = 1
			vim.searchNext(buf, 1)
			return nil
		}

		vim.StatusMsg = "Unknown command: " + cmd
	}
	return nil
}
