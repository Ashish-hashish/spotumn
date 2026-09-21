package backend

import (
	"os"
	"path/filepath"
	"testing"

	"spotumn/internal/config"
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
