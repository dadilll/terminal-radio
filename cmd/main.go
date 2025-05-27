package main

import (
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"radio/internal/api"
	"radio/internal/player"
	"radio/internal/ui"
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

	// Обработка SIGINT (Ctrl+C)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	client := api.NewClient("", 10*time.Second)
	pl := player.New()
	m := ui.NewUIModel(client, pl)
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
