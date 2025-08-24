package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"radio/internal/client"
	"radio/internal/player"
	"radio/internal/storage"
	"radio/internal/ui"
	"runtime"
	"time"

	"radio/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

func clearTerminal() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

func main() {
	clearTerminal()

	logger.Init()
	logger.Log.Info().Msg("Logger initialized")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	client := client.NewClient("", 10*time.Second)
	pl := player.New()

	storagePath := "./jsonfile/favorites.json"
	storageDir := filepath.Dir(storagePath)
	if err := os.MkdirAll(storageDir, os.ModePerm); err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to create storage directory")
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		logger.Log.Warn().Msgf("Favorites file does not exist: %s, will be created on save", storagePath)
	} else if err != nil {
		logger.Log.Error().Err(err).Msgf("Error checking favorites file: %s", storagePath)
	} else {
		logger.Log.Info().Msgf("Favorites file found: %s", storagePath)
	}

	stor, err := storage.NewStorage(storagePath)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to initialize storage")
	}

	// создаём UIModel
	m := ui.NewUIModel(client, pl, stor)

	p := tea.NewProgram(m)

	go func() {
		<-sigs
		logger.Log.Info().Msg("Interrupt signal received, quitting...")
		p.Quit()
	}()

	if err := p.Start(); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to start UI program")
		os.Exit(1)
	}
}
