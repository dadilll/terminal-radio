package player

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sync"

	"radio/pkg/logger"
)

type Player struct {
	cmd      *exec.Cmd
	mu       sync.Mutex
	running  bool
	stopping bool
	wg       sync.WaitGroup
}

func New() *Player {
	return &Player{}
}

func (p *Player) Play(ctx context.Context, streamURL string) error {
	p.mu.Lock()

	if streamURL == "" {
		p.mu.Unlock()
		return errors.New("empty stream URL")
	}

	if p.running {
		p.stopping = true
		oldCmd := p.cmd
		p.mu.Unlock()

		_ = oldCmd.Process.Kill()
		p.wg.Wait()

		p.mu.Lock()
		p.stopping = false
		p.running = false
		p.cmd = nil
	}

	cmd := exec.CommandContext(ctx, "mpv",
		"--no-video", "--really-quiet", "--no-terminal", "--force-window=no", "--idle=no",
		streamURL,
	)

	if err := cmd.Start(); err != nil {
		p.mu.Unlock()
		return fmt.Errorf("starting player: %w", err)
	}

	p.cmd = cmd
	p.running = true
	p.wg.Add(1)
	p.mu.Unlock()

	go func() {
		err := cmd.Wait()
		p.mu.Lock()
		defer p.mu.Unlock()
		defer p.wg.Done()

		if p.stopping {
			logger.Log.Debug().Msg("player stopped for restart")
		} else if err != nil {
			logger.Log.Error().Err(err).Msg("player crashed or exited with error")
		} else {
			logger.Log.Info().Msg("player finished normally")
		}

		p.running = false
		p.cmd = nil
		p.stopping = false
	}()

	logger.Log.Info().Str("url", streamURL).Msg("started playing")
	return nil
}

func (p *Player) Stop() error {
	p.mu.Lock()

	if !p.running || p.cmd == nil {
		p.mu.Unlock()
		return errors.New("no running player")
	}

	p.stopping = true
	cmd := p.cmd
	p.mu.Unlock()

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to stop player: %w", err)
	}

	p.wg.Wait()

	p.mu.Lock()
	p.running = false
	p.cmd = nil
	p.stopping = false
	p.mu.Unlock()

	logger.Log.Info().Msg("player stopped")
	return nil
}

func (p *Player) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
