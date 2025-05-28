package ui

import (
	"context"
	"fmt"
	"radio/internal/api"
	"radio/internal/player"
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
	Station api.Station
	Playing bool
}

func (i StationItem) Title() string {
	title := i.Station.Name
	if i.Playing {
		title = "🎵 " + title
	}
	return title
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

type UIModel struct {
	list           list.Model
	textinput      textinput.Model
	spinner        spinner.Model
	allStations    []api.Station
	filteredItems  []list.Item
	loading        bool
	err            error
	playing        *api.Station
	ctx            context.Context
	cancel         context.CancelFunc
	client         *api.Client
	player         *player.Player
	lastInputTime  time.Time
	searchVisible  bool
	lastQuery      string
	currentSort    SortMode
	sortLabelStyle lipgloss.Style
}

var (
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFA500"))
	placeholder     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	loadingStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	nowPlayingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true)
	positionStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Italic(true)
	tagColors       = []lipgloss.Color{"#FF6B6B", "#6BCB77", "#4D96FF", "#FFD93D", "#C77DFF"}
)

func NewUIModel(client *api.Client, player *player.Player) *UIModel {
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
		sortLabelStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")).Italic(true),
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
		items = append(items, StationItem{Station: s, Playing: playing})
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
			} else if len(m.list.Items()) > 0 {
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
		m.list.SetSize(msg.Width, msg.Height-10)

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

func (m *UIModel) View() string {
	var b strings.Builder

	if m.searchVisible {
		b.WriteString(titleStyle.Render("🔍 Search stations") + "\n")
		b.WriteString(m.textinput.View() + "\n\n")
	}

	if m.loading {
		b.WriteString(loadingStyle.Render(m.spinner.View()+" Loading stations...") + "\n\n")
	} else if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n")
	} else if len(m.filteredItems) == 0 {
		b.WriteString(placeholder.Render("No stations found. Enter search and press Enter.") + "\n\n")
	} else {
		sortLabel := fmt.Sprintf("Sorted by: %s", map[SortMode]string{
			SortByName:    "Name (1)",
			SortByBitrate: "Bitrate (2)",
			SortByCountry: "Country (3)",
		}[m.currentSort])
		b.WriteString(m.sortLabelStyle.Render(sortLabel) + "\n")
		b.WriteString(m.list.View() + "\n")
		b.WriteString(positionStyle.Render(fmt.Sprintf("Station %d of %d", m.list.Index()+1, len(m.filteredItems))) + "\n\n")
	}

	if m.playing != nil {
		b.WriteString(nowPlayingStyle.Render(fmt.Sprintf("▶️ Now playing: %s (%s)", m.playing.Name, m.playing.URL)) + "\n\n")
	}

	b.WriteString(helpStyle.Render("Tab: toggle search • Enter: play/search • s: stop • 1/2/3: sort • Esc/Ctrl+C: quit") + "\n")
	return b.String()
}

// ===== Вспомогательная функция для цветных тегов =====

func colorTags(tags []string) string {
	var b strings.Builder
	for i, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		style := lipgloss.NewStyle().Foreground(tagColors[i%len(tagColors)])
		b.WriteString(style.Render("#" + t))
		if i < len(tags)-1 {
			b.WriteString(" ")
		}
	}
	return b.String()
}
