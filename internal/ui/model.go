package ui

import (
	"context"
	"time"

	"radio/internal/client"
	"radio/internal/player"
	"radio/internal/storage"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type autoSwitchMsg struct{}

type UIModel struct {
	autoSwitchRemaining time.Duration
	autoSwitchDelay     time.Duration
	autoSwitching       bool
	storage             *storage.Storage
	favoritesMode       bool
	list                list.Model
	textinput           textinput.Model
	spinner             spinner.Model
	allStations         []client.Station
	filteredItems       []list.Item
	loading             bool
	err                 error
	playing             *client.Station
	ctx                 context.Context
	cancel              context.CancelFunc
	client              *client.Client
	player              *player.Player
	lastInputTime       time.Time
	searchVisible       bool
	lastQuery           string
	currentSort         SortMode
	sortLabelStyle      lipgloss.Style
	Width               int
}

func NewUIModel(client *client.Client, player *player.Player, storage *storage.Storage) *UIModel {
	ti := textinput.New()
	ti.Placeholder = "Search stations"
	ti.CharLimit = 100
	ti.Width = 40

	sp := spinner.New()
	sp.Style = loadingStyle

	ctx, cancel := context.WithCancel(context.Background())

	stations, err := client.SearchStations(ctx, map[string]string{
		"name": "rock",
	})
	var items []list.Item
	if err == nil {
		for _, st := range stations {
			items = append(items, st)
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 60, 15)
	l.Title = "Radio Stations"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	return &UIModel{
		autoSwitchDelay:     1 * time.Minute,
		autoSwitchRemaining: 1 * time.Minute,
		storage:             storage,
		favoritesMode:       false,
		list:                l,
		textinput:           ti,
		spinner:             sp,
		ctx:                 ctx,
		cancel:              cancel,
		client:              client,
		player:              player,
		allStations:         stations,
		filteredItems:       items,
		lastInputTime:       time.Now(),
		searchVisible:       false,
		currentSort:         SortByBitrate,
		sortLabelStyle:      sortLabelStyle,
	}
}

func (m *UIModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}
