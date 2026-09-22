package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"spotumn/internal/auth"
	"spotumn/internal/backend"
	"spotumn/internal/config"
	"spotumn/internal/ui"
)

func main() {
	// Strict memory management (<35MB target)
	debug.SetMemoryLimit(35 * 1024 * 1024)
	debug.SetGCPercent(50)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "spotumn: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Authenticate via Spotify OAuth PKCE
	authService := auth.NewAuthService(cfg)
	if _, err := authService.Authorize(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "spotumn: authorization failed: %v\n", err)
		os.Exit(1)
	}

	tokenSource := authService.GetTokenSource(ctx)
	client := backend.NewClient(ctx, tokenSource)

	// Launch librespot daemon for direct audio playback
	daemon := backend.NewDaemon()
	if err := daemon.Start(); err == nil {
		defer daemon.Stop()
	}

	// Setup clean signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if last := client.GetLastSavedState(); last != nil {
			client.SaveLastState(last)
		}
		daemon.Stop()
		cancel()
		os.Exit(0)
	}()

	// Run Bubble Tea TUI
	app := ui.NewAppModel(client, cfg)
	prog := tea.NewProgram(app)

	finalModel, err := prog.Run()
	if appModel, ok := finalModel.(*ui.AppModel); ok {
		if pb := appModel.GetPlaybackState(); pb != nil && pb.CurrentTrack != nil {
			client.SaveLastState(pb)
		}
	}

	if err != nil {
		daemon.Stop()
		fmt.Fprintf(os.Stderr, "spotumn error: %v\n", err)
		os.Exit(1)
	}
}
