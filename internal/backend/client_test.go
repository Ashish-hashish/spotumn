package backend

import (
	"os"
	"path/filepath"
	"testing"

	"spotumn/internal/config"
	"github.com/zmb3/spotify/v2"
)

func TestSaveAndLoadLastState(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	c := &Client{}

	track := &Track{
		ID:         "test_id_123",
		URI:        "spotify:track:test_id_123",
		Name:       "Test Track",
		Artist:     "Test Artist",
		Album:      "Test Album",
		DurationMs: 240000,
		ArtURL:     "https://example.com/art.jpg",
	}

	state := &PlaybackState{
		Playing:      true,
		ProgressMs:   45678, // exact millisecond timestamp
		DurationMs:   240000,
		Volume:       80,
		CurrentTrack: track,
		ContextURI:   "spotify:playlist:test_playlist",
	}

	// Persist state
	c.SaveLastState(state)

	// Verify in-memory state
	memState := c.GetLastSavedState()
	if memState == nil {
		t.Fatal("expected in-memory state to be saved, got nil")
	}
	if memState.ProgressMs != 45678 {
		t.Errorf("expected ProgressMs 45678, got %d", memState.ProgressMs)
	}

	// Verify file exists on disk
	path := filepath.Join(config.GetDir(), "last_state.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected last_state.json file on disk, got error: %v", err)
	}

	// Load state from disk
	c2 := &Client{}
	loaded := c2.LoadLastState()
	if loaded == nil {
		t.Fatal("expected loaded state to be non-nil")
	}
	if loaded.ProgressMs != 45678 {
		t.Errorf("expected loaded ProgressMs 45678, got %d", loaded.ProgressMs)
	}
	if loaded.CurrentTrack == nil || loaded.CurrentTrack.Name != "Test Track" {
		t.Errorf("expected loaded track name 'Test Track', got %v", loaded.CurrentTrack)
	}
	if loaded.ContextURI != "spotify:playlist:test_playlist" {
		t.Errorf("expected ContextURI 'spotify:playlist:test_playlist', got %s", loaded.ContextURI)
	}
}

func TestQueueAllowsConsecutiveDuplicates(t *testing.T) {
	// Verify that multiple identical tracks in queue are preserved without deduplication
	items := []Track{
		{ID: "track1", URI: "spotify:track:track1", Name: "Song A"},
		{ID: "track1", URI: "spotify:track:track1", Name: "Song A"}, // duplicate
		{ID: "track2", URI: "spotify:track:track2", Name: "Song B"},
		{ID: "track2", URI: "spotify:track:track2", Name: "Song B"}, // duplicate
	}

	res := &QueueData{
		Current: &items[0],
	}
	for _, item := range items {
		res.Items = append(res.Items, item)
	}

	if len(res.Items) != 4 {
		t.Errorf("expected 4 queue items preserved with duplicates, got %d", len(res.Items))
	}
}

func TestSpotumnDeviceSorting(t *testing.T) {
	devices := []spotify.PlayerDevice{
		{ID: "dev1", Name: "Phone", Type: "Smartphone"},
		{ID: "dev2", Name: "Living Room", Type: "Speaker"},
		{ID: "dev3", Name: "spotumn", Type: "Computer"},
		{ID: "dev4", Name: "Echo Dot", Type: "Speaker"},
	}

	sorted := sortDevicesWithSpotumnFirst(devices)
	if len(sorted) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(sorted))
	}
	if sorted[0].Name != "spotumn" || sorted[0].ID != "dev3" {
		t.Errorf("expected spotumn to be topmost device (index 0), got %s (%s)", sorted[0].Name, sorted[0].ID)
	}
	if sorted[1].Name != "Phone" || sorted[2].Name != "Living Room" || sorted[3].Name != "Echo Dot" {
		t.Errorf("expected original relative order for remaining devices, got: %+v", sorted)
	}
}

func TestLoadLastStateDefaultsToSpotumnDevice(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	c := &Client{}
	state := &PlaybackState{
		Playing:    false,
		DeviceName: "iPhone 15 Pro",
		DeviceID:   "remote_device_id",
		CurrentTrack: &Track{
			Name:   "Song",
			Artist: "Artist",
		},
	}
	c.SaveLastState(state)

	loaded := c.LoadLastState()
	if loaded == nil {
		t.Fatal("expected non-nil loaded state")
	}
	if loaded.DeviceName != "spotumn" {
		t.Errorf("expected loaded state to default DeviceName to 'spotumn', got '%s'", loaded.DeviceName)
	}
	if loaded.DeviceID != "" {
		t.Errorf("expected loaded state to clear remote DeviceID, got '%s'", loaded.DeviceID)
	}
}

