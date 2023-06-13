package config

import "github.com/charmbracelet/lipgloss"

var (
	SkippedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#b8b8b8"))
	PassedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00b300"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	FailedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6600"))
	StdoutStyle  = lipgloss.NewStyle().Faint(true)
)
