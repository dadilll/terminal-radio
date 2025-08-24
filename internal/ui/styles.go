package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Padding(0, 2).
			Background(lipgloss.Color("#1B1B1B"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF3333")).
			Bold(true)

	placeholder = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	positionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Italic(true)

	sortLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#999999")).
			Italic(true).
			Padding(0, 1)

	tagColors = []lipgloss.Color{
		"#FF6B6B", "#6BCB77", "#4D96FF", "#FFD93D", "#C77DFF",
	}
)
