package ui

import (
	"context"
	"fmt"
	"radio/internal/api"
	"radio/internal/player"
	"radio/internal/storage"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StationItem struct {
	Station  api.Station
	Playing  bool
	Favorite bool
}

func (i StationItem) Title() string {
	title := i.Station.Name
	if i.Playing {
		title = "🎵 " + title
	}
	if i.Favorite {
		title = "★ " + title
	}
	return truncate(title, 30)
}

func (i StationItem) Description() string {
	tags := colorTags(strings.Split(i.Station.Tags, ","))
	return fmt.Sprintf("%s • %dkbps • %s", i.Station.Country, i.Station.Bitrate, tags)
}

func (i StationItem) FilterValue() string {
	return i.Station.Name
}

type SortMode int

const (
	SortByName SortMode = iota
	SortByBitrate
	SortByCountry
)

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

type UIModel struct {
	isPlaying        bool
	playbackProgress float64
	playerTicker     *time.Ticker
	favoriteStations []api.Station
	storage          *storage.Storage
	favoritesMode    bool
	list             list.Model
	textinput        textinput.Model
	spinner          spinner.Model
	allStations      []api.Station
	filteredItems    []list.Item
	loading          bool
	err              error
	playing          *api.Station
	ctx              context.Context
	cancel           context.CancelFunc
	client           *api.Client
	player           *player.Player
	lastInputTime    time.Time
	searchVisible    bool
	lastQuery        string
	currentSort      SortMode
	sortLabelStyle   lipgloss.Style
	Width            int
}

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

	nowPlayingStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Padding(0, 1).
		Background(lipgloss.Color("#0B3D0B"))

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

	tagColors = []lipgloss.Color{"#FF6B6B", "#6BCB77", "#4D96FF", "#FFD93D", "#C77DFF"}
)

func NewUIModel(client *api.Client, player *player.Player, storage *storage.Storage) *UIModel {
	ti := textinput.New()
	ti.Placeholder = "Search stations"
	ti.CharLimit = 100
	ti.Width = 40

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 60, 15)
	l.Title = "Radio Stations"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	sp := spinner.New()
	sp.Style = loadingStyle

	ctx, cancel := context.WithCancel(context.Background())

	return &UIModel{
		storage:        storage,
		favoritesMode:  false,
		list:           l,
		textinput:      ti,
		spinner:        sp,
		ctx:            ctx,
		cancel:         cancel,
		client:         client,
		player:         player,
		lastInputTime:  time.Now(),
		searchVisible:  false,
		currentSort:    SortByBitrate,
		sortLabelStyle: sortLabelStyle,
	}
}

func (m *UIModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

type searchMsg []api.Station
type errMsg error

func searchStations(ctx context.Context, query string) tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient("", 10*time.Second)
		stations, err := client.SearchStations(ctx, query)
		if err != nil {
			return errMsg(err)
		}
		return searchMsg(stations)
	}
}

func (m *UIModel) filterStations(query string) {
	query = strings.ToLower(query)
	var filtered []api.Station
	for _, s := range m.allStations {
		if strings.Contains(strings.ToLower(s.Name), query) {
			filtered = append(filtered, s)
		}
	}

	m.sortStations(&filtered)
	var items []list.Item
	for _, s := range filtered {
		playing := m.playing != nil && m.playing.URL == s.URL
		items = append(items, StationItem{
			Station:  s,
			Playing:  playing,
			Favorite: m.storage.IsFavorite(s.URL),
		})
	}
	m.filteredItems = items
	m.list.SetItems(items)
}

func (m *UIModel) sortStations(stations *[]api.Station) {
	switch m.currentSort {
	case SortByBitrate:
		sort.Slice(*stations, func(i, j int) bool {
			return (*stations)[i].Bitrate > (*stations)[j].Bitrate
		})
	case SortByName:
		sort.Slice(*stations, func(i, j int) bool {
			return (*stations)[i].Name < (*stations)[j].Name
		})
	case SortByCountry:
		sort.Slice(*stations, func(i, j int) bool {
			return (*stations)[i].Country < (*stations)[j].Country
		})
	}
}

func (m *UIModel) showFavorites() {
	var favStations []api.Station
	for _, s := range m.allStations {
		if m.storage.IsFavorite(s.URL) {
			favStations = append(favStations, s)
		}
	}
	m.favoriteStations = favStations
	m.sortStations(&favStations)

	var items []list.Item
	for _, s := range favStations {
		playing := m.playing != nil && m.playing.URL == s.URL
		items = append(items, StationItem{
			Station:  s,
			Playing:  playing,
			Favorite: true,
		})
	}
	m.filteredItems = items
	m.list.SetItems(items)
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
					_ = m.player.Stop()
					err := m.player.Play(m.ctx, i.Station.URL)
					if err != nil {
						m.err = fmt.Errorf("failed to play station: %w", err)
					}
					m.playing = &i.Station
					m.filterStations(m.textinput.Value())
				}
			}

		case "a":
			if item, ok := m.list.SelectedItem().(StationItem); ok {
				id := item.Station.URL
				if m.storage.IsFavorite(id) {
					_ = m.storage.RemoveFavorite(id)
				} else {
					_ = m.storage.AddFavorite(id)
				}
				if m.favoritesMode {
					m.showFavorites()
				} else {
					m.filterStations(m.textinput.Value())
				}
			}

		case "z":
			m.favoritesMode = !m.favoritesMode
			if m.favoritesMode {
				m.showFavorites()
			} else {
				m.filterStations(m.textinput.Value())
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

	case searchMsg:
		m.loading = false
		m.err = nil
		m.allStations = []api.Station(msg)
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

func (m *UIModel) renderPlayer() string {
	if m.playing == nil {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(lipgloss.Color("#666666")).
			Padding(1, 2).
			Width(50).
			Italic(true).
			Align(lipgloss.Center).
			Render("⏸️  No station playing")
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

	nameCol := nameStyle.Render(fmt.Sprintf("▶️ %s", name))
	countryCol := infoStyle.Render(fmt.Sprintf("🌍 %s", country))
	bitrateCol := infoStyle.Render(fmt.Sprintf("🎵 %d kbps", m.playing.Bitrate))
	favCol := infoStyle.Render(fmt.Sprintf("❤️ %s", favIcon))

	separator := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render(" | ")

	playerContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		nameCol,
		separator,
		countryCol,
		separator,
		bitrateCol,
		separator,
		favCol,
	)

	playerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(lipgloss.Color("#5FD3F3")).
		Background(lipgloss.AdaptiveColor{
			Light: "#1E90FF",
			Dark:  "#0D1B2A",
		}).
		Padding(1, 3).
		Width(lipgloss.Width(playerContent) + 6) // ширина по содержимому + отступы

	return playerStyle.Render(playerContent)
}

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

	if m.loading {
		contentParts = append(contentParts, loadingStyle.Render(m.spinner.View()+" Loading stations..."))
	} else if m.err != nil {
		contentParts = append(contentParts, errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	} else if len(m.filteredItems) == 0 {
		contentParts = append(contentParts, placeholder.Render("No stations found. Enter search and press Enter."))
	} else {
		// Сортировка больше не отображается
		contentParts = append(contentParts, lipgloss.NewStyle().PaddingTop(1).Render(m.list.View()))
		contentParts = append(contentParts, positionStyle.Render(fmt.Sprintf("Station %d of %d", m.list.Index()+1, len(m.filteredItems))))
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Left, contentParts...)

	contentRow := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Padding(1, 2).
		Render(mainContent)

	footer := m.renderPlayer()

	help := helpStyle.Render("Tab: toggle search • Enter: play/search • s: stop • a: toggle favorite • z: show favorites • 1/2/3: sort • Esc/Ctrl+C: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		contentRow,
		footer,
		help,
	)
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}
	return s[:maxLen]
}

func colorTags(tags []string) string {
	var b strings.Builder
	for i, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(tagColors[i%len(tagColors)]).
			Padding(0, 1).
			MarginRight(1)
		b.WriteString(style.Render(t))
	}
	return b.String()
}
