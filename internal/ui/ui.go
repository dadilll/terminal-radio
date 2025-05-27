package ui

import (
	"context"
	"fmt"
	"radio/internal/player"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"radio/internal/api"
)

type StationItem struct {
	Station api.Station
}

func (i StationItem) Title() string { return i.Station.Name }
func (i StationItem) Description() string {
	return fmt.Sprintf("%s • %dkbps • %s", i.Station.Country, i.Station.Bitrate, i.Station.Tags)
}
func (i StationItem) FilterValue() string { return i.Station.Name }

type tickMsg struct{}

type UIModel struct {
	list            list.Model
	textinput       textinput.Model
	stations        []api.Station
	loading         bool
	err             error
	playing         *api.Station
	searchTerm      string
	ctx             context.Context
	cancel          context.CancelFunc
	progress        progress.Model
	progressPercent float64
	client          *api.Client
	player          *player.Player
}

func NewUIModel(client *api.Client, player *player.Player) UIModel {
	ti := textinput.New()
	ti.Placeholder = "Search stations"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 30

	const defaultWidth = 50
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, 15)
	l.Title = "Radio Stations"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ctx, cancel := context.WithCancel(context.Background())

	prog := progress.New(progress.WithScaledGradient("#00f", "#0ff"))

	return UIModel{
		list:            l,
		textinput:       ti,
		ctx:             ctx,
		cancel:          cancel,
		loading:         false,
		client:          client,
		player:          player,
		progress:        prog,
		progressPercent: 0.0,
	}
}

func (m UIModel) Init() tea.Cmd {
	return nil
}

type searchMsg []api.Station
type errMsg error

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textinput.Focused() {
				m.loading = true
				m.err = nil
				m.playing = nil
				m.searchTerm = m.textinput.Value()
				m.progressPercent = 0.0
				m.progress.SetPercent(0.0)
				cmds = append(cmds, searchStations(m.ctx, m.searchTerm), tick())
			} else if len(m.list.Items()) > 0 {
				i, ok := m.list.SelectedItem().(StationItem)
				if ok {
					_ = m.player.Stop()
					err := m.player.Play(m.ctx, i.Station.URL)
					if err != nil {
						m.err = fmt.Errorf("failed to play station: %w", err)
					}
					m.playing = &i.Station
				}
			}
		case tea.KeyTab:
			if m.textinput.Focused() {
				m.textinput.Blur()
			} else {
				m.textinput.Focus()
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancel()
			return m, tea.Quit
		}

	case searchMsg:
		m.loading = false
		m.err = nil
		m.stations = []api.Station(msg)
		items := make([]list.Item, len(m.stations))
		for i, s := range m.stations {
			items[i] = StationItem{Station: s}
		}
		m.list.SetItems(items)
		m.list.Select(0)
		m.textinput.Blur()
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg
		return m, nil

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)

	case tickMsg:
		if m.loading {
			m.progressPercent += 0.03
			if m.progressPercent > 1.0 {
				m.progressPercent = 0.0
			}
			m.progress.SetPercent(m.progressPercent)
			cmds = append(cmds, tick())
		}
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m UIModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.textinput.View())
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("🔄 Searching...\n")
		b.WriteString(m.progress.View() + "\n")
	} else if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	} else if len(m.stations) == 0 {
		b.WriteString("No stations found. Enter search and press Enter.\n")
	} else {
		b.WriteString(m.list.View())
	}

	b.WriteString("\n\n")

	if m.playing != nil {
		b.WriteString(fmt.Sprintf("▶️ Now playing: %s (%s)\n", m.playing.Name, m.playing.URL))
	}

	b.WriteString("\nPress Ctrl+C or Esc to quit.\n")

	return b.String()
}

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
