package ui

import (
	"context"
	"strings"
	"time"

	"radio/internal/client"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type searchMsg []client.Station
type errMsg error

func searchStations(ctx context.Context, query string) tea.Cmd {
	return func() tea.Msg {
		c := client.NewClient("", 10*time.Second)
		stations, err := c.SearchStations(ctx, map[string]string{
			"name": query,
		})
		if err != nil {
			return errMsg(err)
		}
		return searchMsg(stations)
	}
}

func (m *UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancel()
			return m, tea.Quit

		case "tab":
			m.searchVisible = !m.searchVisible
			if m.searchVisible {
				m.textinput.Focus()
			} else {
				m.textinput.Blur()
			}

		case "enter":
			if m.searchVisible && m.textinput.Focused() {
				query := strings.TrimSpace(m.textinput.Value())
				if query != "" {
					m.loading = true
					m.err = nil
					m.lastQuery = query
					cmds = append(cmds, searchStations(m.ctx, query))
					m.list.SetItems([]list.Item{})
				}
			} else if !m.searchVisible && len(m.list.Items()) > 0 {
				i, ok := m.list.SelectedItem().(StationItem)
				if ok {
					m.PlayStation(i, true)
				}
			}

		case "a":
			if item, ok := m.list.SelectedItem().(StationItem); ok {
				station := item.Station
				if m.storage.IsFavorite(station.URL) {
					_ = m.storage.RemoveFavorite(station.URL)
				} else {
					_ = m.storage.AddFavorite(station)
				}
				if m.favoritesMode {
					m.showFavorites()
				} else {
					m.filterStations(m.textinput.Value())
				}
			}
		case "[":
			if m.autoSwitchDelay > 1*time.Minute {
				m.autoSwitchDelay -= 1 * time.Minute
			}
		case "]":
			m.autoSwitchDelay += 1 * time.Minute

		case "m":
			cmds = append(cmds, m.toggleAutoSwitch())

		case "z":
			if m.favoritesMode {
				m.favoritesMode = false
				m.filterStations(m.textinput.Value())
			} else {
				m.favoritesMode = true
				m.showFavorites()
			}

		case "s":
			if m.playing != nil {
				_ = m.player.Stop()
				m.playing = nil
				m.filterStations(m.textinput.Value())
			}
		case "1":
			m.currentSort = SortByName
			m.filterStations(m.textinput.Value())

		case "2":
			m.currentSort = SortByBitrate
			m.filterStations(m.textinput.Value())

		case "3":
			m.currentSort = SortByCountry
			m.filterStations(m.textinput.Value())
		}

		if m.searchVisible && m.textinput.Focused() {
			m.filterStations(m.textinput.Value())
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width

		reservedHeight := 1 + 1 + 1 + 3 + 1 + 2
		availableHeight := msg.Height - reservedHeight
		if availableHeight < 5 {
			availableHeight = 5
		}

		m.list.SetSize(msg.Width-32, availableHeight)

	case autoSwitchMsg:
		if m.autoSwitching {
			m.randomStation()
			cmds = append(cmds, m.startAutoSwitchCmd())
		}

	case searchMsg:
		m.loading = false
		m.err = nil
		m.allStations = msg
		m.filterStations(m.textinput.Value())

	case errMsg:
		m.loading = false
		m.err = msg
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
