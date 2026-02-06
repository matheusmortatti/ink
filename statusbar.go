package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar renders the status bar with mode indicator, filename, and position.
func RenderStatusBar(vim *VimState, buf *Buffer, width int) string {
	modeColor := ModeColor(vim.Mode)

	// Mode indicator
	modeStr := statusModeStyle.
		Background(modeColor).
		Foreground(lipgloss.Color("0")).
		Render(vim.Mode.String())

	// File info
	fileName := buf.filePath
	if fileName == "" {
		fileName = "[No Name]"
	}
	modified := ""
	if buf.dirty {
		modified = " [+]"
	}
	fileStr := statusFileStyle.Render(fileName + modified)

	// Cursor position
	posStr := statusPosStyle.Render(
		fmt.Sprintf("Ln %d, Col %d", buf.cursor.Row+1, buf.cursor.Col+1),
	)

	// Status message
	var midStr string
	if vim.StatusMsg != "" && vim.Mode != ModeCommand {
		midStr = statusFileStyle.Render(vim.StatusMsg)
	}

	// Calculate remaining width
	usedWidth := lipgloss.Width(modeStr) + lipgloss.Width(fileStr) + lipgloss.Width(posStr)
	if midStr != "" {
		usedWidth += lipgloss.Width(midStr)
	}

	remaining := width - usedWidth
	if remaining < 0 {
		remaining = 0
	}

	if midStr != "" {
		gap1 := statusBarStyle.Render(" ")
		gap2Len := remaining - 1
		if gap2Len < 0 {
			gap2Len = 0
		}
		gap2 := statusBarStyle.Render(strings.Repeat(" ", gap2Len))
		return modeStr + fileStr + gap1 + midStr + gap2 + posStr
	}

	gap := statusBarStyle.Render(strings.Repeat(" ", remaining))
	return modeStr + fileStr + gap + posStr
}

// RenderCommandLine renders the command/search input line.
func RenderCommandLine(vim *VimState, width int) string {
	if vim.Mode == ModeCommand {
		prefix := string(vim.CommandPrefix)
		line := prefix + vim.CommandLine
		cursor := "█"
		padLen := width - len(line) - 1
		if padLen < 0 {
			padLen = 0
		}
		return line + cursor + strings.Repeat(" ", padLen)
	}

	if vim.StatusMsg != "" {
		padLen := width - len(vim.StatusMsg)
		if padLen < 0 {
			padLen = 0
		}
		return vim.StatusMsg + strings.Repeat(" ", padLen)
	}

	return strings.Repeat(" ", width)
}
