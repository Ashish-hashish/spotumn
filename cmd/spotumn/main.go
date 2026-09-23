package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"spotumn/internal/auth"
	"spotumn/internal/backend"
	"spotumn/internal/config"
	"spotumn/internal/ui"
)

func main() {
	debug.SetMemoryLimit(40 * 1024 * 1024)
	debug.SetGCPercent(20)

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

	// Launch embedded player daemon
	daemon := backend.NewDaemon()
	client.SetLocalDeviceID(daemon.DeviceId())

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

	if !daemon.HasStoredCredentials() {
		fmt.Println()
		fmt.Println("==> Spotumn Embedded Player Setup (one-time)")
		fmt.Println("    Authenticating Spotify Connect playback engine...")
		if err := daemon.Start(""); err != nil {
			fmt.Fprintf(os.Stderr, "spotumn: failed to start embedded player: %v\n", err)
		} else {
			defer daemon.Stop()

			select {
			case auth := <-daemon.AuthCodes():
				if auth != nil {
					fmt.Println()
					fmt.Printf("    Pairing Code: %s\n", auth.Code)
					fmt.Printf("    Pairing URL:  %s\n", auth.Url)
					fmt.Println()
					fmt.Println("    If your browser did not open automatically, visit the URL above.")
					fmt.Println("    Please click 'Link account' / 'Pair' to connect the player.")
					fmt.Println("    Waiting for authorization...")
				}
			case <-time.After(10 * time.Second):
			}

			waitCtx, waitCancel := context.WithTimeout(ctx, 2*time.Minute)
			defer waitCancel()
			if err := daemon.WaitUntilReady(waitCtx); err == nil {
				fmt.Println("    ✔ Player authenticated and ready!")
				time.Sleep(500 * time.Millisecond)
			}
		}
	} else {
		if err := daemon.Start(""); err == nil {
			defer daemon.Stop()
		}
	}

	// Run Bubble Tea TUI
	app := ui.NewAppModel(client, daemon, cfg)
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
