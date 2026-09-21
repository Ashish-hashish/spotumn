package backend

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

	credDest := filepath.Join(cacheDir, "credentials.json")
	hasCreds := false
	if _, err := os.Stat(credDest); err == nil {
		hasCreds = true
	}

	// If no credentials found, perform automated 1-time OAuth sign-in for librespot
	if !hasCreds {
		cmdOAuth := exec.Command(bin,
			"--name", "spotumn",
			"--device-type", "computer",
			"--cache", cacheDir,
			"--enable-oauth",
		)
		stdout, err := cmdOAuth.StdoutPipe()
		if err == nil {
			cmdOAuth.Stderr = cmdOAuth.Stdout
			if err := cmdOAuth.Start(); err == nil {
				d.cmd = cmdOAuth
				d.running = true

				// Scan output for "Browse to: (https://...)" and auto-open browser
				go func() {
					scanner := bufio.NewScanner(stdout)
					for scanner.Scan() {
						line := scanner.Text()
						if strings.Contains(line, "Browse to: https://") {
							idx := strings.Index(line, "https://")
							if idx != -1 {
								url := strings.Fields(line[idx:])[0]
								openURL(url)
							}
						}
					}
					_ = cmdOAuth.Wait()
					d.mu.Lock()
					d.running = false
					d.mu.Unlock()
				}()

				// Wait up to 5 seconds for credentials to be created
				for i := 0; i < 25; i++ {
					time.Sleep(200 * time.Millisecond)
					if _, err := os.Stat(credDest); err == nil {
						hasCreds = true
						break
					}
				}

				return nil
			}
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

func openURL(targetURL string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", targetURL)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
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
