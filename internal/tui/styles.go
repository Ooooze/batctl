package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6600")).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	labelStyle = lipgloss.NewStyle().
			Width(16).
			Foreground(lipgloss.Color("#AAAAAA"))

	valueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6600"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CC66"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF3333"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	gaugeFullStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CC66"))

	gaugeEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6600"))

	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6600"))

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1, 2)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1)
)
