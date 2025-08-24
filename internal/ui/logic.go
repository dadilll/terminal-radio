package ui

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"radio/internal/client"

	"github.com/charmbracelet/bubbles/list"
)

func (m *UIModel) filterStations(query string) {
	var stations []client.Station
	if m.favoritesMode {
		stations = m.storage.ListFavorites()
	} else {
		stations = m.allStations
	}

	var filtered []client.Station
	query = strings.ToLower(query) // для нечувствительного к регистру поиска
	for _, s := range stations {
		if query == "" || strings.Contains(strings.ToLower(s.Name), query) {
			filtered = append(filtered, s)
		}
	}

	// сортировка
	switch m.currentSort {
	case SortByName:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	case SortByBitrate:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Bitrate > filtered[j].Bitrate
		})
	case SortByCountry:
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Country < filtered[j].Country
		})
	}

	items := make([]list.Item, 0, len(filtered))
	for _, s := range filtered {
		items = append(items, StationItem{Station: s})
	}

	m.list.SetItems(items)
}

func (m *UIModel) showFavorites() {
	stations := m.storage.ListFavorites()

	items := make([]list.Item, 0, len(stations))
	for _, s := range stations {
		items = append(items, StationItem{Station: s})
	}

	m.list.SetItems(items)
}
func (m *UIModel) PlayStation(item StationItem, stopFirst bool) {
	if stopFirst {
		_ = m.player.Stop()
	}
	err := m.player.Play(m.ctx, item.Station.URL)
	if err != nil {
		m.err = fmt.Errorf("failed to play station: %w", err)
		return
	}
	m.playing = &item.Station
	m.filterStations(m.textinput.Value())
}

func (m *UIModel) startAutoSwitchCmd() tea.Cmd {
	return tea.Tick(m.autoSwitchDelay, func(t time.Time) tea.Msg {
		m.autoSwitchRemaining = m.autoSwitchDelay
		return autoSwitchMsg{}
	})
}

func (m *UIModel) randomStation() {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(items))
	m.list.Select(randomIndex)

	if item, ok := items[randomIndex].(StationItem); ok {
		_ = m.player.Stop()
		_ = m.player.Play(m.ctx, item.Station.URL)
		m.playing = &item.Station
	}
}

func (m *UIModel) toggleAutoSwitch() tea.Cmd {
	if m.autoSwitching {
		m.autoSwitching = false
		return nil
	} else {
		m.autoSwitching = true
		return m.startAutoSwitchCmd()
	}
}
