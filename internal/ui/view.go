package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m *UIModel) View() string {
	if m.searchVisible {
		modalStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFA500")).
			Padding(1, 2).
			Width(60).
			Align(lipgloss.Center)

		searchBox := titleStyle.Render("🔍 Search stations") + "\n" + m.textinput.View()
		modal := modalStyle.Render(searchBox)

		return lipgloss.Place(m.Width, 10, lipgloss.Center, lipgloss.Center, modal)
	}

	var contentParts []string

	var header string
	if m.favoritesMode {
		header = "🌟 Favorites"
	} else {
		header = "📻 Radio Stations"
	}
	contentParts = append(contentParts, titleStyle.Render(header))

	switch {
	case m.loading:
		contentParts = append(contentParts, loadingStyle.Render(m.spinner.View()+" Loading stations..."))

	case m.err != nil:
		contentParts = append(contentParts, errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))

	case len(m.filteredItems) == 0:
		placeholderBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			Width(60)

		var msg string
		switch {
		case m.favoritesMode:
			msg = "No favorite stations yet. Press 'a' to add some."
		case m.textinput.Value() != "":
			msg = "No stations match your search."
		default:
			msg = "No stations found. Enter a search query and press Enter."
		}

		contentParts = append(contentParts, placeholderBox.Render(placeholder.Render(msg)))

	default:
		listBoxStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(0, 1).
			MarginTop(1).
			MarginBottom(1)

		listInBox := listBoxStyle.Render(m.list.View())
		contentParts = append(contentParts, listInBox)

		contentParts = append(contentParts,
			positionStyle.Render(fmt.Sprintf("Station %d of %d", m.list.Index()+1, len(m.filteredItems))),
		)
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

	footer := m.renderPlayer()

	help := helpStyle.Render("Tab: toggle search • Enter: play/search • s: stop • a: toggle favorite • " +
		"z: favorites • 1/2/3: sort • m: toggle auto • [/] adjust delay • Esc/Ctrl+C: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		mainContent,
		footer,
		help,
	)
}

func (m *UIModel) renderPlayer() string {
	if m.playing == nil {
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(1, 2).
			Width(50).
			Italic(true).
			Align(lipgloss.Center)
		if m.autoSwitching {
			style = style.Background(lipgloss.Color("#FFAA00"))
		}
		return style.Render("⏸️  No station playing")
	}

	const maxNameLen = 20
	const maxCountryLen = 15

	favIcon := "❌"
	if m.storage.IsFavorite(m.playing.URL) {
		favIcon = "⭐"
	}

	name := truncateText(m.playing.Name, maxNameLen)
	country := truncateText(m.playing.Country, maxCountryLen)

	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C0C0C0")).Padding(0, 1)

	label := "▶️"
	if m.autoSwitching {
		label += " [AUTO]"
	}

	nameCol := nameStyle.Render(fmt.Sprintf("%s %s", label, name))
	countryCol := infoStyle.Render(fmt.Sprintf("🌍 %s", country))
	bitrateCol := infoStyle.Render(fmt.Sprintf("🎵 %d kbps", m.playing.Bitrate))
	favCol := infoStyle.Render(fmt.Sprintf("❤️ %s", favIcon))

	delayCol := ""
	if m.autoSwitching {
		delayCol = infoStyle.Render(fmt.Sprintf("⏩ %v", m.autoSwitchDelay.Round(time.Minute)))
	}

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(" | ")

	cols := []string{nameCol, separator, countryCol, separator, bitrateCol, separator, favCol}
	if delayCol != "" {
		cols = append(cols, separator, delayCol)
	}

	playerContent := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

	playerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("#5FD3F3")).
		Background(lipgloss.AdaptiveColor{
			Light: "#1E90FF",
			Dark:  "#0D1B2A",
		}).
		Padding(1, 3).
		Width(lipgloss.Width(playerContent) + 6)

	if m.autoSwitching {
		playerStyle = playerStyle.Background(lipgloss.Color("#FFAA00"))
		playerStyle = playerStyle.Bold(true)
	}

	return playerStyle.Render(playerContent)
}
