package backend

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"spotumn/internal/config"

	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/daemon"
)

type fileStateStore struct {
	path             string
	mu               sync.Mutex
	credentialsSaved chan struct{}
}

func newFileStateStore(path string) *fileStateStore {
	return &fileStateStore{
		path:             path,
		credentialsSaved: make(chan struct{}, 1),
	}
}

func (s *fileStateStore) Load() (*librespot.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := &librespot.AppState{}
	data, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(data, state)
	}

	if len(state.DeviceId) != 40 {
		state.DeviceId = generateDeviceID()
		_ = s.saveLocked(state)
	}

	return state, nil
}

func (s *fileStateStore) Save(state *librespot.AppState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.saveLocked(state)
	if err == nil && len(state.Credentials.Data) > 0 {
		select {
		case s.credentialsSaved <- struct{}{}:
		default:
		}
	}
	return err
}

func (s *fileStateStore) saveLocked(state *librespot.AppState) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(s.path), 0700)
	return os.WriteFile(s.path, data, 0600)
}

func generateDeviceID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		h := sha1.Sum([]byte(fmt.Sprintf("spotumn-%d", time.Now().UnixNano())))
		return hex.EncodeToString(h[:])
	}
	return hex.EncodeToString(b)
}

// eventBridge connects go-librespot events to the TUI
type eventBridge struct {
	events    chan *daemon.ApiEvent
	requests  chan daemon.ApiRequest
	authCodes chan *daemon.ApiDeviceAuth
	mu        sync.Mutex
	closed    bool
}

func newEventBridge() *eventBridge {
	return &eventBridge{
		events:    make(chan *daemon.ApiEvent, 32),
		requests:  make(chan daemon.ApiRequest, 16),
		authCodes: make(chan *daemon.ApiDeviceAuth, 2),
	}
}

func (b *eventBridge) Emit(ev *daemon.ApiEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	select {
	case b.events <- ev:
	default:
		// drop oldest to prevent blocking
		select {
		case <-b.events:
		default:
		}
		b.events <- ev
	}
}

func (b *eventBridge) Receive() <-chan daemon.ApiRequest { return b.requests }

func (b *eventBridge) SetAuthCode(auth *daemon.ApiDeviceAuth) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	if auth != nil {
		select {
		case b.authCodes <- auth:
		default:
		}
		if auth.Url != "" {
			OpenURL(auth.Url)
		}
	}
}

func (b *eventBridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.closed {
		b.closed = true
		close(b.events)
		close(b.requests)
	}
	return nil
}

func resolveAudioBackend(preferred string) string {
	preferred = strings.ToLower(strings.TrimSpace(preferred))
	if preferred == "alsa" || preferred == "pulseaudio" || preferred == "pipe" {
		return preferred
	}

	pulseSocket := os.Getenv("PULSE_SERVER")
	if pulseSocket == "" {
		if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
			pulseSocket = filepath.Join(runtimeDir, "pulse", "native")
		} else {
			pulseSocket = fmt.Sprintf("/run/user/%d/pulse/native", os.Getuid())
		}
	}

	if _, err := os.Stat(pulseSocket); err == nil {
		return "pulseaudio"
	}

	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if _, err := os.Stat(filepath.Join(runtimeDir, "pipewire-0")); err == nil {
			return "pulseaudio"
		}
	}

	return "alsa"
}

type Daemon struct {
	app        *daemon.App
	cancel     context.CancelFunc
	bridge     *eventBridge
	stateStore *fileStateStore
	deviceId   string
	running    bool
	mu         sync.Mutex
}

func NewDaemon() *Daemon {
	cacheDir := filepath.Join(config.GetCacheDir(), "librespot")
	_ = os.MkdirAll(cacheDir, 0700)

	store := newFileStateStore(filepath.Join(cacheDir, "state.json"))
	state, _ := store.Load()

	if len(state.Credentials.Data) == 0 {
		if home, err := os.UserHomeDir(); err == nil {
			spCredPath := filepath.Join(home, ".cache", "spotify-player", "credentials.json")
			if data, err := os.ReadFile(spCredPath); err == nil {
				var spCred struct {
					Username string `json:"username"`
					AuthData string `json:"auth_data"`
				}
				if json.Unmarshal(data, &spCred) == nil && spCred.Username != "" && spCred.AuthData != "" {
					if raw, err := base64.StdEncoding.DecodeString(spCred.AuthData); err == nil && len(raw) > 0 {
						state.Credentials.Username = spCred.Username
						state.Credentials.Data = raw
						_ = store.Save(state)
					}
				}
			}
		}
	}

	return &Daemon{
		stateStore: store,
		deviceId:   state.DeviceId,
	}
}

func (d *Daemon) DeviceId() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.deviceId
}

func (d *Daemon) HasStoredCredentials() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stateStore == nil {
		return false
	}
	state, err := d.stateStore.Load()
	return err == nil && len(state.Credentials.Data) > 0
}

func (d *Daemon) AuthCodes() <-chan *daemon.ApiDeviceAuth {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.bridge != nil {
		return d.bridge.authCodes
	}
	return nil
}

func (d *Daemon) WaitUntilReady(ctx context.Context) error {
	if d.HasStoredCredentials() {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-d.stateStore.credentialsSaved:
		return nil
	}
}

func (d *Daemon) Events() <-chan *daemon.ApiEvent {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.bridge != nil {
		return d.bridge.events
	}
	return nil
}

func (d *Daemon) Start(username string, token ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return nil
	}

	cfg := config.Get()
	audioBackend := resolveAudioBackend(cfg.AudioBackend)
	cacheDir := filepath.Join(config.GetCacheDir(), "librespot")
	_ = os.MkdirAll(cacheDir, 0700)

	credsCfg := daemon.CredentialsConfig{
		Type: "device_auth",
	}

	dCfg := &daemon.Config{
		DeviceId:              d.deviceId,
		DeviceName:            "spotumn",
		DeviceType:            "computer",
		AudioBackend:          audioBackend,
		VolumeSteps:           100,
		InitialVolume:         50,
		Bitrate:               cfg.Bitrate,
		CrossfadeDuration:     cfg.CrossfadeSec * 1000,
		NormalisationDisabled: !cfg.Normalisation,
		SkipDebounce:          200 * time.Millisecond,
		ZeroconfEnabled:       true,
		Cache: daemon.CacheConfig{
			Enabled: true,
			Dir:     filepath.Join(cacheDir, "audio"),
		},
		Metadata: daemon.MetadataConfig{
			Enabled: true,
		},
		Credentials: credsCfg,
	}

	bridge := newEventBridge()

	app, err := daemon.New(&daemon.Options{
		Logger:     &librespot.NullLogger{},
		Config:     dCfg,
		StateStore: d.stateStore,
		APIServer:  bridge,
	})
	if err != nil {
		_ = bridge.Close()
		return fmt.Errorf("embedded player init failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.app = app
	d.cancel = cancel
	d.bridge = bridge
	d.running = true

	go func() {
		if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "embedded player error: %v\n", err)
			d.mu.Lock()
			d.running = false
			d.mu.Unlock()
		}
	}()

	return nil
}

func (d *Daemon) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return
	}

	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
	if d.app != nil {
		_ = d.app.Close()
		d.app = nil
	}
	if d.bridge != nil {
		_ = d.bridge.Close()
		d.bridge = nil
	}
	d.running = false
}

func (d *Daemon) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

func (d *Daemon) Restart(username string, token ...string) error {
	d.Stop()
	time.Sleep(100 * time.Millisecond)
	return d.Start(username)
}

func (d *Daemon) SendCommand(reqType daemon.ApiRequestType, data any) error {
	d.mu.Lock()
	bridge := d.bridge
	running := d.running
	d.mu.Unlock()

	if !running || bridge == nil {
		return errors.New("player not running")
	}

	req, wait := daemon.NewApiRequest(reqType, data)
	select {
	case bridge.requests <- req:
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := wait(ctx)
		return err
	case <-time.After(500 * time.Millisecond):
		return errors.New("player busy")
	}
}

func (d *Daemon) PlayPause() error {
	return d.SendCommand(daemon.ApiRequestTypePlayPause, nil)
}

func (d *Daemon) Resume() error {
	return d.SendCommand(daemon.ApiRequestTypeResume, nil)
}

func (d *Daemon) Pause() error {
	return d.SendCommand(daemon.ApiRequestTypePause, nil)
}

func (d *Daemon) Next() error {
	return d.SendCommand(daemon.ApiRequestTypeNext, daemon.ApiNext{})
}

func (d *Daemon) Previous() error {
	return d.SendCommand(daemon.ApiRequestTypePrev, nil)
}

func (d *Daemon) Seek(posMs int) error {
	if posMs < 0 {
		posMs = 0
	}
	return d.SendCommand(daemon.ApiRequestTypeSeek, daemon.ApiSeek{Position: int64(posMs)})
}

func (d *Daemon) SetVolume(percent int) error {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}
	return d.SendCommand(daemon.ApiRequestTypeSetVolume, daemon.ApiSetVolume{Volume: int32(percent)})
}

func (d *Daemon) SetShuffle(shuffle bool) error {
	return d.SendCommand(daemon.ApiRequestTypeSetShufflingContext, shuffle)
}

func (d *Daemon) SetRepeat(mode string) error {
	switch mode {
	case "track":
		_ = d.SendCommand(daemon.ApiRequestTypeSetRepeatingContext, false)
		return d.SendCommand(daemon.ApiRequestTypeSetRepeatingTrack, true)
	case "context":
		_ = d.SendCommand(daemon.ApiRequestTypeSetRepeatingTrack, false)
		return d.SendCommand(daemon.ApiRequestTypeSetRepeatingContext, true)
	default:
		_ = d.SendCommand(daemon.ApiRequestTypeSetRepeatingTrack, false)
		return d.SendCommand(daemon.ApiRequestTypeSetRepeatingContext, false)
	}
}

func (d *Daemon) PlayURI(uri, skipToUri string, posMs int) error {
	return d.SendCommand(daemon.ApiRequestTypePlay, daemon.ApiPlay{
		Uri:       uri,
		SkipToUri: skipToUri,
		Position:  int64(posMs),
	})
}

func (d *Daemon) AddToQueue(uri string) error {
	return d.SendCommand(daemon.ApiRequestTypeAddToQueue, uri)
}

func (d *Daemon) StopPlayback() error {
	return d.SendCommand(daemon.ApiRequestTypeStop, nil)
}

func OpenURL(targetURL string) {
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
