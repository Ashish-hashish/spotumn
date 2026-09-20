package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"spotumn/internal/backend"
	"spotumn/internal/lyrics"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		ms       int
		expected string
	}{
		{0, "00:00"},
		{5000, "00:05"},
		{65000, "01:05"},
		{215000, "03:35"},
	}

	for _, c := range cases {
		res := FormatDuration(c.ms)
		if res != c.expected {
			t.Errorf("FormatDuration(%d) = %s, expected %s", c.ms, res, c.expected)
		}
	}
}

func TestTruncateAndPad(t *testing.T) {
	s := "Hello World"
	pad := PadToWidth(s, 20)
	if w := ansi.StringWidth(pad); w != 20 {
		t.Errorf("expected padded visible width 20, got %d", w)
	}

	trunc := TruncateString("Long song title that exceeds width", 10)
	if !strings.HasSuffix(trunc, "…") {
		t.Errorf("expected ellipsis suffix, got %s", trunc)
	}
}

func TestRenderFullUI(t *testing.T) {
	params := ViewParams{
		Width:        120,
		Height:       36,
		Focused:      PaneNav,
		CurrentTab:       TabTracks,
		ShowLeftSidebar:  true,
		ShowRightSidebar: true,
		NavIndex:     0,
		CenterIndex:  0,
		QueueIndex:   0,
		Playlists: []backend.Playlist{
			{ID: "1", Name: "Chill Mix", TrackCount: 42},
			{ID: "2", Name: "Rock Classics", TrackCount: 88},
		},
		Playback: &backend.PlaybackState{
			Playing:    true,
			ProgressMs: 75000,
			DurationMs: 180000,
			Volume:     65,
			CurrentTrack: &backend.Track{
				Name:   "Midnight City",
				Artist: "M83",
				Album:  "Hurry Up, We're Dreaming",
			},
		},
		Queue: []backend.Track{
			{Name: "Next Song", Artist: "Artist 1"},
			{Name: "Another Song", Artist: "Artist 2"},
		},
		LyricsLines: []lyrics.Line{
			{TimeMs: 70000, Text: "Active lyric line"},
			{TimeMs: 80000, Text: "Upcoming lyric line"},
		},
	}

	output := RenderFullUI(params)
	if output == "" {
		t.Fatal("expected non-empty rendered UI")
	}

	// Verify rounded border curves exist in panels
	if !strings.Contains(output, "╭") || !strings.Contains(output, "╰") {
		t.Error("rounded border characters missing")
	}

	// Verify Playlists panel is rendered
	if !strings.Contains(output, "Playlists") {
		t.Error("Playlists panel missing from rendered UI")
	}

	// Verify sticky player controls are present
	if !strings.Contains(output, "Vol:") {
		t.Error("volume indicator missing from player")
	}
	if !strings.Contains(output, "Midnight City") {
		t.Error("track title missing from UI")
	}
}

func TestRenderZenMode(t *testing.T) {
	params := ViewParams{
		Width:      100,
		Height:     30,
		ZenMode:    true,
		Playback: &backend.PlaybackState{
			Playing:    true,
			ProgressMs: 30000,
			DurationMs: 120000,
			Volume:     80,
			CurrentTrack: &backend.Track{
				Name:   "Zen Track",
				Artist: "Ambient Artist",
			},
		},
	}

	output := RenderFullUI(params)
	if !strings.Contains(output, "ZEN MODE") {
		t.Error("zen mode title missing")
	}
	if !strings.Contains(output, "Zen Track") {
		t.Error("track name missing from zen mode")
	}
}

func TestDynamicThemeSwitching(t *testing.T) {
	// Dark mode
	SetTheme(true)
	if CurrentTheme.Purple == nil || CurrentTheme.Mint == nil {
		t.Fatal("dark theme colors missing")
	}

	// Light mode
	SetTheme(false)
	if CurrentTheme.Purple == nil || CurrentTheme.Mint == nil {
		t.Fatal("light theme colors missing")
	}

	// Revert to dark for consistent testing
	SetTheme(true)
}

func TestRenderLyricsPointer(t *testing.T) {
	lines := []lyrics.Line{
		{TimeMs: 1000, Text: "First line"},
		{TimeMs: 5000, Text: "Second line"},
		{TimeMs: 10000, Text: "Third line"},
	}

	// At 6000ms, line index 1 ("Second line") is active
	rendered := renderLyrics(lines, 1, 6000, true, 40, 10)
	joined := strings.Join(rendered, "\n")
	if !strings.Contains(joined, "❯") {
		t.Errorf("expected lyric pointer ❯ in rendered lyrics, got: %s", joined)
	}
	if !strings.Contains(joined, "Second line") {
		t.Errorf("expected active lyric text in rendered lyrics")
	}
}

func TestBordersAndIcons(t *testing.T) {
	params := ViewParams{
		Width:            120,
		Height:           30,
		Username:         "TestUser",
		ShowLeftSidebar:  true,
		ShowRightSidebar: true,
	}
	output := RenderFullUI(params)

	// Top border checks: green dot before spotumn, person icon before username
	if !strings.Contains(output, "●") {
		t.Error("expected green dot ● before spotumn in top border")
	}
	if !strings.Contains(output, "👤") {
		t.Error("expected person icon 👤 before username in top border")
	}
	if !strings.Contains(output, "TestUser") {
		t.Error("expected username in top border")
	}

	// Bottom border check: keybinds embedded inside the bottom frame
	if !strings.Contains(output, "[Space] Play") {
		t.Error("expected [Space] Play keybind embedded in bottom border")
	}
}

func TestTrackRowBlockHighlight(t *testing.T) {
	track := backend.Track{
		Name:       "Test Track",
		Artist:     "Test Artist",
		DurationMs: 180000,
	}

	// Selected and focused
	rowFocused := renderTrackRow(0, track, true, false, true, 20, 15, 60)
	if !strings.Contains(rowFocused, "Test Track") || !strings.Contains(rowFocused, "Test Artist") {
		t.Error("expected track and artist in focused row")
	}

	// Unselected - distinct colors
	rowUnfocused := renderTrackRow(0, track, false, false, false, 20, 15, 60)
	if !strings.Contains(rowUnfocused, "Test Track") || !strings.Contains(rowUnfocused, "Test Artist") {
		t.Error("expected track and artist in unselected row")
	}
}

func TestRenderDevicesModal(t *testing.T) {
	modal := RenderDevicesModal(nil, 0, false, 80, 24)
	if !strings.Contains(modal, "[r] Rescan") {
		t.Error("expected [r] Rescan hint in devices modal")
	}

	modalScanning := RenderDevicesModal(nil, 0, true, 80, 24)
	if !strings.Contains(modalScanning, "Scanning...") {
		t.Error("expected Scanning... indicator when isScanning is true")
	}
}
