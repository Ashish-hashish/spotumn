package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"spotumn/internal/config"
)

type Daemon struct {
	cmd     *exec.Cmd
	running bool
	mu      sync.Mutex
}

func NewDaemon() *Daemon {
	return &Daemon{}
}

// FindLibrespot searches known paths for librespot binary
func FindLibrespot() (string, error) {
	candidates := []string{
		"/usr/sbin/librespot",
		"/usr/bin/librespot",
		"/usr/local/bin/librespot",
	}

	for _, path := range candidates {
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			return path, nil
		}
	}

	return exec.LookPath("librespot")
}

// Start launches librespot safely for direct local audio playback
func (d *Daemon) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return nil
	}

	bin, err := FindLibrespot()
	if err != nil {
		return fmt.Errorf("librespot not found: %w", err)
	}

	cacheDir := filepath.Join(config.GetDir(), "cache")
	_ = os.MkdirAll(cacheDir, 0700)

	// Auto-copy existing librespot credentials if available
	credDest := filepath.Join(cacheDir, "credentials.json")
	if _, err := os.Stat(credDest); os.IsNotExist(err) {
		home, _ := os.UserHomeDir()
		altSource := filepath.Join(home, ".cache", "spotify-player", "credentials.json")
		if data, err := os.ReadFile(altSource); err == nil {
			_ = os.WriteFile(credDest, data, 0600)
		}
	}

	args := []string{
		"--name", "spotumn",
		"--device-type", "computer",
		"--cache", cacheDir,
		"--bitrate", "320",
	}

	cmd := exec.Command(bin, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start librespot: %w", err)
	}

	d.cmd = cmd
	d.running = true

	go func() {
		_ = cmd.Wait()
		d.mu.Lock()
		d.running = false
		d.mu.Unlock()
	}()

	return nil
}

// Stop cleanly terminates librespot
func (d *Daemon) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running || d.cmd == nil || d.cmd.Process == nil {
		return
	}

	// Try graceful SIGTERM first
	_ = d.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		state, err := d.cmd.Process.Wait()
		_ = state
		done <- err
	}()

	select {
	case <-done:
	case <-time.After(600 * time.Millisecond):
		_ = d.cmd.Process.Kill()
	}

	d.running = false
}

func (d *Daemon) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}
