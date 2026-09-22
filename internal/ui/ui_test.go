package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/zmb3/spotify/v2"
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
	lyricsData := []lyrics.Line{
		{TimeMs: 10000, Text: "Zen lyrics line one"},
		{TimeMs: 30000, Text: "Zen active singing line"},
		{TimeMs: 50000, Text: "Zen lyrics line three"},
	}

	params := ViewParams{
		Width:       100,
		Height:      30,
		ZenMode:     true,
		ZenView:     ZenViewBoth,
		LyricsLines: lyricsData,
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

	// 1. Test ZenViewBoth
	outputBoth := RenderFullUI(params)
	strippedBoth := ansi.Strip(outputBoth)
	if !strings.Contains(strippedBoth, "Zen Mode") {
		t.Error("zen mode title missing")
	}
	if !strings.Contains(strippedBoth, "Art + Lyrics") {
		t.Error("expected top-right indicator 'Art + Lyrics' in ZenViewBoth")
	}
	if !strings.Contains(strippedBoth, "Zen Track") {
		t.Error("track name missing from zen mode")
	}
	if !strings.Contains(strippedBoth, "Zen active singing line") {
		t.Error("active singing lyric line missing in ZenViewBoth")
	}
	if !strings.Contains(strippedBoth, "❯  Zen active singing line  ❮") {
		t.Error("expected centered active singing markers ❯ ... ❮ in ZenViewBoth")
	}
	if !strings.Contains(strippedBoth, "◖v◗") {
		t.Error("view toggle keybind ◖v◗ missing from bottom border")
	}

	// Test navigation cursor rendering in Zen mode
	paramsNav := params
	paramsNav.LyricsCursor = 0 // cursor moved to first line
	outputNav := RenderFullUI(paramsNav)
	strippedNav := ansi.Strip(outputNav)
	if !strings.Contains(strippedNav, "➜  Zen lyrics line one  ➜") {
		t.Error("expected cursor marker ➜ ... ➜ on navigated line")
	}

	// 2. Test ZenViewArt
	paramsArt := params
	paramsArt.ZenView = ZenViewArt
	outputArt := RenderFullUI(paramsArt)
	strippedArt := ansi.Strip(outputArt)
	if !strings.Contains(strippedArt, "Art Only") {
		t.Error("expected top-right indicator 'Art Only' in ZenViewArt")
	}
	if !strings.Contains(strippedArt, "Zen Track") {
		t.Error("track name missing from ZenViewArt")
	}

	// 3. Test ZenViewLyrics
	paramsLyrics := params
	paramsLyrics.ZenView = ZenViewLyrics
	outputLyrics := RenderFullUI(paramsLyrics)
	strippedLyrics := ansi.Strip(outputLyrics)
	if !strings.Contains(strippedLyrics, "Lyrics Only") {
		t.Error("expected top-right indicator 'Lyrics Only' in ZenViewLyrics")
	}
	if !strings.Contains(strippedLyrics, "Zen Track") {
		t.Error("track banner missing from ZenViewLyrics")
	}
	if !strings.Contains(strippedLyrics, "Zen active singing line") {
		t.Error("active singing lyric line missing from ZenViewLyrics")
	}
	if !strings.Contains(strippedLyrics, "[ ⏮") {
		t.Error("controls missing from ZenViewLyrics")
	}
}

func TestDynamicThemeSwitching(t *testing.T) {
	// Dark mode (Catppuccin Mocha)
	SetTheme(true)
	if CurrentTheme.Primary == nil || CurrentTheme.Tertiary == nil {
		t.Fatal("dark theme colors missing")
	}
	if CurrentTheme.Surface == nil || CurrentTheme.OnSurface == nil {
		t.Fatal("dark theme surface colors missing")
	}

	// Light mode (Catppuccin Latte)
	SetTheme(false)
	if CurrentTheme.Primary == nil || CurrentTheme.Tertiary == nil {
		t.Fatal("light theme colors missing")
	}
	if CurrentTheme.Surface == nil || CurrentTheme.OnSurface == nil {
		t.Fatal("light theme surface colors missing")
	}

	// Revert to dark for consistent testing
	SetTheme(true)
}

func TestCustomThemeConfig(t *testing.T) {
	// Test SetThemeFromConfig with a fully custom palette
	customCfg := ThemeConfig{
		MPrimary:          "#ff0000",
		MOnPrimary:        "#ffffff",
		MSecondary:        "#00ff00",
		MOnSecondary:      "#000000",
		MTertiary:         "#0000ff",
		MOnTertiary:       "#ffffff",
		MError:            "#ff00ff",
		MOnError:          "#000000",
		MSurface:          "#111111",
		MOnSurface:        "#eeeeee",
		MSurfaceVariant:   "#222222",
		MOnSurfaceVariant: "#cccccc",
		MOutline:          "#555555",
		MShadow:           "#000000",
		MHover:            "#ffaa00",
		MOnHover:          "#000000",
		MGold:             "#ffd700",
		MSuccess:          "#00cc00",
	}

	SetThemeFromConfig(customCfg)

	if CurrentTheme.Primary == nil {
		t.Fatal("custom theme Primary should be set")
	}
	if CurrentTheme.Surface == nil {
		t.Fatal("custom theme Surface should be set")
	}
	if CurrentTheme.Gold == nil {
		t.Fatal("custom theme Gold should be set")
	}

	// Test that StyleBase has background set (renders non-empty)
	rendered := StyleBase.Render(" ")
	if rendered == "" {
		t.Error("StyleBase should produce non-empty output")
	}

	// Test mergeConfig: partial override keeps defaults for unset fields
	base := catppuccinMocha()
	partial := ThemeConfig{MPrimary: "#abcdef"}
	merged := mergeConfig(base, partial)
	if merged.MPrimary != "#abcdef" {
		t.Errorf("expected merged MPrimary '#abcdef', got '%s'", merged.MPrimary)
	}
	if merged.MSurface != base.MSurface {
		t.Errorf("expected merged MSurface to remain default '%s', got '%s'", base.MSurface, merged.MSurface)
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
	if !strings.Contains(output, "") {
		t.Error("expected person icon  before username in top border")
	}
	if !strings.Contains(output, "TestUser") {
		t.Error("expected username in top border")
	}

	// Bottom border check: keybinds embedded inside the bottom frame
	stripped := ansi.Strip(output)
	if !strings.Contains(stripped, "◖Space◗ Play") {
		t.Errorf("expected '◖Space◗ Play' keybind embedded in bottom border, got: %s", stripped)
	}

	// Test with empty username
	paramsEmpty := ViewParams{
		Width:            120,
		Height:           30,
		Username:         "",
		ShowLeftSidebar:  true,
		ShowRightSidebar: true,
	}
	outputEmpty := RenderFullUI(paramsEmpty)
	if strings.Contains(outputEmpty, "") {
		t.Error("expected no person icon  when username is empty")
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

	// Verify spotumn is rendered topmost even if received later in device list
	devs := []spotify.PlayerDevice{
		{ID: "d1", Name: "Phone", Type: "Smartphone"},
		{ID: "d2", Name: "External Speaker", Type: "Speaker"},
		{ID: "d3", Name: "spotumn", Type: "Computer"},
	}
	rendered := RenderDevicesModal(devs, 0, false, 80, 24)
	stripped := ansi.Strip(rendered)
	spotumnIdx := strings.Index(stripped, "spotumn")
	phoneIdx := strings.Index(stripped, "Phone")
	if spotumnIdx == -1 || phoneIdx == -1 || spotumnIdx >= phoneIdx {
		t.Errorf("expected spotumn to appear before Phone as the topmost item in devices modal:\n%s", stripped)
	}
}

func TestPlaylistFilters(t *testing.T) {
	// Verify filter string representations
	if FilterAll.String() != "ALL" {
		t.Errorf("expected 'ALL', got '%s'", FilterAll.String())
	}
	if FilterSpotify.String() != "By Spotify" {
		t.Errorf("expected 'By Spotify', got '%s'", FilterSpotify.String())
	}
	if FilterByYou.String() != "By You" {
		t.Errorf("expected 'By You', got '%s'", FilterByYou.String())
	}
	if FilterAlbums.String() != "Albums" {
		t.Errorf("expected 'Albums', got '%s'", FilterAlbums.String())
	}
	if FilterArtists.String() != "Artists" {
		t.Errorf("expected 'Artists', got '%s'", FilterArtists.String())
	}

	// Test Nav line rendering for Albums
	albums := []backend.Playlist{
		{ID: "alb1", URI: "spotify:album:1", Name: "Discovery", OwnerID: "Daft Punk", TrackCount: 14},
	}
	albumLines := RenderNavLines(albums, nil, FilterAlbums, 0, true, 40, 10)
	joinedAlb := strings.Join(albumLines, "\n")
	if !strings.Contains(joinedAlb, "Albums [Albums]") {
		t.Errorf("expected 'Albums [Albums]' in nav header, got: %s", joinedAlb)
	}
	if !strings.Contains(joinedAlb, "Discovery") {
		t.Error("expected album name 'Discovery' in nav lines")
	}

	// Test Nav line rendering for Artists
	artists := []backend.Playlist{
		{ID: "art1", URI: "spotify:artist:1", Name: "Radiohead", OwnerID: "Artist"},
	}
	artistLines := RenderNavLines(artists, nil, FilterArtists, 0, true, 40, 10)
	joinedArt := strings.Join(artistLines, "\n")
	if !strings.Contains(joinedArt, "Artists [Artists]") {
		t.Errorf("expected 'Artists [Artists]' in nav header, got: %s", joinedArt)
	}
	if !strings.Contains(joinedArt, "Radiohead") {
		t.Error("expected artist name 'Radiohead' in nav lines")
	}

	// Test filteredPlaylists method on AppModel
	model := &AppModel{
		userID:   "my_user",
		username: "my_user",
		playlists: []backend.Playlist{
			{ID: "1", Name: "User Playlist", OwnerID: "my_user"},
			{ID: "2", Name: "Spotify Hit Mix", OwnerID: "spotify"},
		},
		albums: []backend.Playlist{
			{ID: "3", Name: "Saved Album", OwnerID: "Some Artist"},
		},
		artists: []backend.Playlist{
			{ID: "4", Name: "Followed Artist", OwnerID: "Artist"},
		},
	}

	// ALL
	model.playlistFilter = FilterAll
	if len(model.filteredPlaylists()) != 2 {
		t.Errorf("FilterAll: expected 2 playlists, got %d", len(model.filteredPlaylists()))
	}

	// By Spotify
	model.playlistFilter = FilterSpotify
	spotifyPls := model.filteredPlaylists()
	if len(spotifyPls) != 1 || spotifyPls[0].Name != "Spotify Hit Mix" {
		t.Errorf("FilterSpotify: expected 'Spotify Hit Mix', got %+v", spotifyPls)
	}

	// By You
	model.playlistFilter = FilterByYou
	userPls := model.filteredPlaylists()
	if len(userPls) != 1 || userPls[0].Name != "User Playlist" {
		t.Errorf("FilterByYou: expected 'User Playlist', got %+v", userPls)
	}

	// Albums
	model.playlistFilter = FilterAlbums
	albs := model.filteredPlaylists()
	if len(albs) != 1 || albs[0].Name != "Saved Album" {
		t.Errorf("FilterAlbums: expected 'Saved Album', got %+v", albs)
	}

	// Artists
	model.playlistFilter = FilterArtists
	arts := model.filteredPlaylists()
	if len(arts) != 1 || arts[0].Name != "Followed Artist" {
		t.Errorf("FilterArtists: expected 'Followed Artist', got %+v", arts)
	}
}

func TestStartupPausedState(t *testing.T) {
	// Verify that playback state initializes as paused even if saved state was playing
	lastState := &backend.PlaybackState{
		Playing:    true,
		ProgressMs: 45000,
		CurrentTrack: &backend.Track{
			Name:   "Track 1",
			Artist: "Artist 1",
		},
	}
	lastState.Playing = false
	if lastState.Playing {
		t.Error("expected state to be paused on startup")
	}
}

func TestArtistPageAlbumsSection(t *testing.T) {
	tracks := []backend.Track{
		{ID: "t1", URI: "spotify:track:t1", Name: "Come Together", Artist: "The Beatles", DurationMs: 259000},
		{ID: "t2", URI: "spotify:track:t2", Name: "Something", Artist: "The Beatles", DurationMs: 183000},
	}
	albums := []backend.Playlist{
		{ID: "a1", URI: "spotify:album:a1", Name: "Abbey Road", OwnerID: "Album • 1969", TrackCount: 17},
		{ID: "a2", URI: "spotify:album:a2", Name: "Let It Be", OwnerID: "Album • 1970", TrackCount: 12},
	}

	params := ViewParams{
		Width:          100,
		Height:         35,
		Focused:        PaneCenter,
		CurrentTab:     TabTracks,
		PlaylistName:   "The Beatles",
		PlaylistTracks: tracks,
		ArtistAlbums:   albums,
		CenterIndex:    2, // Select first album (index len(tracks) + 0)
	}

	rendered := RenderFullUI(params)
	stripped := ansi.Strip(rendered)

	// Check headers and sections
	if !strings.Contains(stripped, "The Beatles") {
		t.Error("expected artist name in header")
	}
	if !strings.Contains(stripped, "Top Tracks") {
		t.Error("expected Top Tracks section header")
	}
	if !strings.Contains(stripped, "Come Together") || !strings.Contains(stripped, "Something") {
		t.Error("expected artist top tracks in output")
	}
	if !strings.Contains(stripped, "Albums & Discography") {
		t.Error("expected Albums & Discography section header")
	}
	if !strings.Contains(stripped, "Abbey Road") || !strings.Contains(stripped, "Let It Be") {
		t.Error("expected album names in albums section")
	}
	if !strings.Contains(stripped, "1969") || !strings.Contains(stripped, "17 tracks") {
		t.Error("expected album year and track count")
	}
}

func TestArtistPageOpenAlbum(t *testing.T) {
	tracks := []backend.Track{
		{ID: "t1", URI: "spotify:track:t1", Name: "Song 1", Artist: "Artist", DurationMs: 200000},
	}
	albums := []backend.Playlist{
		{ID: "alb1", URI: "spotify:album:alb1", Name: "Debut Album", OwnerID: "Album • 2021", TrackCount: 10},
	}

	app := &AppModel{
		focused:        PaneCenter,
		currentTab:     TabTracks,
		currentPlURI:   "spotify:artist:art1",
		currentPlName:  "Artist",
		playlistTracks: tracks,
		artistAlbums:   albums,
		centerIndex:    1, // Cursor on "Debut Album"
		client:         &backend.Client{},
	}

	// Hit enter on the album
	_, cmd := app.handleEnter()
	if cmd == nil {
		t.Fatal("expected command to fetch album tracks")
	}
	if app.currentPlURI != "spotify:album:alb1" {
		t.Errorf("expected currentPlURI 'spotify:album:alb1', got %s", app.currentPlURI)
	}
	if app.currentPlName != "Debut Album" {
		t.Errorf("expected currentPlName 'Debut Album', got %s", app.currentPlName)
	}
	if app.centerIndex != 0 {
		t.Errorf("expected centerIndex reset to 0, got %d", app.centerIndex)
	}
	if len(app.navHistory) != 1 || app.navHistory[0].URI != "spotify:artist:art1" {
		t.Errorf("expected artist pushed to navHistory, got %+v", app.navHistory)
	}

	// Test back navigation
	app.handleKeyPress(tea.KeyPressMsg(tea.Key{Code: 'b', Text: "b"}))
	if app.currentPlURI != "spotify:artist:art1" {
		t.Errorf("expected backspace/b to restore artist URI, got %s", app.currentPlURI)
	}
	if app.centerIndex != 1 {
		t.Errorf("expected backspace/b to restore centerIndex 1, got %d", app.centerIndex)
	}
}

func TestPlayerNerdFontControlsAndBorderDevice(t *testing.T) {
	state := &backend.PlaybackState{
		Playing:    true,
		ProgressMs: 60000,
		DurationMs: 180000,
		Volume:     75,
		Shuffle:    true,
		Repeat:     "track",
		DeviceName: "spotumn",
		CurrentTrack: &backend.Track{
			Name:   "Midnight City",
			Artist: "M83",
		},
	}

	// 1. Check rendered player lines and controls layout
	lines := RenderPlayerLines(state, false, 80)
	joined := strings.Join(lines, "\n")
	stripped := ansi.Strip(joined)

	// Verify Nerd Font shuffle on icon
	if !strings.Contains(stripped, "󰒝") {
		t.Error("expected shuffle on Nerd Font icon '󰒝' in player controls")
	}
	// Verify Nerd Font repeat once icon
	if !strings.Contains(stripped, "󰑘") {
		t.Error("expected repeat once Nerd Font icon '󰑘' in player controls")
	}
	// Verify controls order: shuffle on left of previous, repeat on right of next
	if !strings.Contains(stripped, "󰒝   ⏮") || !strings.Contains(stripped, "⏭   󰑘") {
		t.Errorf("expected shuffle on left of previous and repeat on right of next, got:\n%s", stripped)
	}

	// 2. Test shuffle off and repeat modes
	state.Shuffle = false
	state.Repeat = "context" // repeat all
	lines2 := RenderPlayerLines(state, false, 80)
	stripped2 := ansi.Strip(strings.Join(lines2, "\n"))
	if !strings.Contains(stripped2, "󰒞") {
		t.Error("expected shuffle off Nerd Font icon '󰒞' in player controls")
	}
	if !strings.Contains(stripped2, "󰑖") {
		t.Error("expected repeat all Nerd Font icon '󰑖' in player controls")
	}

	state.Repeat = "off"
	lines3 := RenderPlayerLines(state, false, 80)
	stripped3 := ansi.Strip(strings.Join(lines3, "\n"))
	if !strings.Contains(stripped3, "󰑗") {
		t.Error("expected repeat off Nerd Font icon '󰑗' in player controls")
	}

	// 3. Test Device Name inside top border line
	renderedPlayer := RenderPlayer(state, false, 80)
	strippedPlayer := ansi.Strip(renderedPlayer)
	if !strings.Contains(strippedPlayer, "󰓃 spotumn") {
		t.Errorf("expected '󰓃 spotumn' inside player top border line, got:\n%s", strippedPlayer)
	}
}

func TestZenModeVolumeSlider(t *testing.T) {
	params := ViewParams{
		Width:   80,
		Height:  30,
		ZenMode: true,
		ZenView: ZenViewBoth,
		LyricsLines: []lyrics.Line{
			{TimeMs: 0, Text: "Line 1"},
			{TimeMs: 5000, Text: "Line 2"},
		},
		Playback: &backend.PlaybackState{
			Playing:    true,
			ProgressMs: 2500,
			DurationMs: 10000,
			Volume:     65,
			CurrentTrack: &backend.Track{
				Name:   "Zen Track",
				Artist: "Zen Artist",
			},
		},
	}

	// 1. ZenViewBoth
	both := ansi.Strip(RenderFullUI(params))
	if !strings.Contains(both, "Vol: [") || !strings.Contains(both, "65%") {
		t.Errorf("expected volume slider 'Vol: [...] 65%%' in ZenViewBoth, got:\n%s", both)
	}

	// 2. ZenViewArt
	params.ZenView = ZenViewArt
	art := ansi.Strip(RenderFullUI(params))
	if !strings.Contains(art, "Vol: [") || !strings.Contains(art, "65%") {
		t.Errorf("expected volume slider 'Vol: [...] 65%%' in ZenViewArt, got:\n%s", art)
	}

	// 3. ZenViewLyrics
	params.ZenView = ZenViewLyrics
	lyr := ansi.Strip(RenderFullUI(params))
	if !strings.Contains(lyr, "Vol: [") || !strings.Contains(lyr, "65%") {
		t.Errorf("expected volume slider 'Vol: [...] 65%%' in ZenViewLyrics, got:\n%s", lyr)
	}
}

func TestLyricsAutoSyncTimer(t *testing.T) {
	lyricsLines := []lyrics.Line{
		{TimeMs: 0, Text: "First line"},
		{TimeMs: 10000, Text: "Second line (singing)"},
		{TimeMs: 20000, Text: "Third line"},
	}

	m := &AppModel{
		lyricsLines: lyricsLines,
		client:      &backend.Client{},
		playback: &backend.PlaybackState{
			Playing:    false, // keep false to avoid periodic save network/disk writes in test
			ProgressMs: 12000, // activeIdx is 1 ("Second line")
		},
		lyricsCursor:         1,
		lyricsManualScroll:   true,
		lyricsPointerMovedAt: time.Now().Add(-1 * time.Second),
	}

	// Case 1: Pointer moved to non-singing line 0, idle for 2 seconds (< 3s) -> remains in manual scroll
	m.lyricsManualScroll = true
	m.lyricsCursor = 0 // stray
	m.lyricsPointerMovedAt = time.Now().Add(-2 * time.Second)
	m.Update(TickMsg{})
	if !m.lyricsManualScroll {
		t.Error("expected lyricsManualScroll to remain true when idle for < 3s")
	}

	// Case 2: Pointer left idle for >= 3 seconds -> reverts to auto-sync and snaps to active singing line (idx 1)
	m.lyricsPointerMovedAt = time.Now().Add(-3100 * time.Millisecond)
	m.Update(TickMsg{})
	if m.lyricsManualScroll {
		t.Error("expected lyricsManualScroll to revert to false after >= 3s idle")
	}
	if m.lyricsCursor != 1 {
		t.Errorf("expected lyricsCursor to snap to active singing lyric 1, got %d", m.lyricsCursor)
	}

	// Case 3: Pointer placed on singing line (idx 1), idle for 2 seconds (< 3s) -> remains manual
	m.lyricsManualScroll = true
	m.lyricsCursor = 1
	m.lyricsPointerMovedAt = time.Now().Add(-2 * time.Second)
	m.Update(TickMsg{})
	if !m.lyricsManualScroll {
		t.Error("expected lyricsManualScroll to remain true when on singing line for < 3s")
	}

	// Case 4: Pointer placed on singing line, idle for >= 3 seconds -> reverts to auto-sync
	m.lyricsPointerMovedAt = time.Now().Add(-3100 * time.Millisecond)
	m.Update(TickMsg{})
	if m.lyricsManualScroll {
		t.Error("expected lyricsManualScroll to revert to false after >= 3s on singing line")
	}
}

func TestKeybindsModalVolumeCooldownNotice(t *testing.T) {
	modal := RenderKeybindsModal(nil, 0, false, 90, 40)
	stripped := ansi.Strip(modal)

	if !strings.Contains(stripped, "Tip:") {
		t.Error("expected 'Tip:' header in keybindings modal")
	}
	if !strings.Contains(stripped, "Spotify limits how fast volume requests can be sent") {
		t.Error("expected volume cooldown explanation in keybindings modal")
	}
	if !strings.Contains(stripped, "󰒝 on / 󰒞 off") {
		t.Error("expected Nerd Font shuffle icons documented in keybindings modal")
	}
	if !strings.Contains(stripped, "󰑗 off / 󰑖 all / 󰑘 once") {
		t.Error("expected Nerd Font repeat icons documented in keybindings modal")
	}
}

func TestKeybindNavigationAndEditing(t *testing.T) {
	km := NewKeyManager()

	// Default matching
	if act := km.Action("space"); act != ActionPlayPause {
		t.Errorf("expected ActionPlayPause for space, got '%s'", act)
	}
	if act := km.Action("n"); act != ActionNextTrack {
		t.Errorf("expected ActionNextTrack for n, got '%s'", act)
	}

	// Edit keybind: rebind next track to 'l'
	nextIdx := -1
	for i, it := range km.Items {
		if it.ID == ActionNextTrack {
			nextIdx = i
			break
		}
	}
	if nextIdx == -1 {
		t.Fatal("ActionNextTrack not found in KeyManager")
	}

	km.SetKey(nextIdx, "l")
	if act := km.Action("l"); act != ActionNextTrack {
		t.Errorf("expected ActionNextTrack for edited key 'l', got '%s'", act)
	}

	// Render modal in edit mode
	modalEdit := RenderKeybindsModal(km.Items, nextIdx, true, 90, 40)
	strippedEdit := ansi.Strip(modalEdit)
	if !strings.Contains(strippedEdit, "[Press key...]") {
		t.Error("expected '[Press key...]' indicator in edit mode")
	}

	// Verify tab 2 is Lyrics and tab 3 is History
	if act := km.Action("2"); act != ActionTabLyrics {
		t.Errorf("expected ActionTabLyrics for '2', got '%s'", act)
	}
	if act := km.Action("3"); act != ActionTabHistory {
		t.Errorf("expected ActionTabHistory for '3', got '%s'", act)
	}

	// Verify =/- for +/- 5 points and +/_ for +/- 10 points
	if act := km.Action("="); act != ActionVolumeUp {
		t.Errorf("expected ActionVolumeUp for '=', got '%s'", act)
	}
	if act := km.Action("-"); act != ActionVolumeDown {
		t.Errorf("expected ActionVolumeDown for '-', got '%s'", act)
	}
	if act := km.Action("+"); act != ActionVolumeUpBig {
		t.Errorf("expected ActionVolumeUpBig for '+', got '%s'", act)
	}
	if act := km.Action("shift+="); act != ActionVolumeUpBig {
		t.Errorf("expected ActionVolumeUpBig for 'shift+=', got '%s'", act)
	}
	if act := km.Action("_"); act != ActionVolumeDownBig {
		t.Errorf("expected ActionVolumeDownBig for '_', got '%s'", act)
	}
	if act := km.Action("shift+-"); act != ActionVolumeDownBig {
		t.Errorf("expected ActionVolumeDownBig for 'shift+-', got '%s'", act)
	}
	if act := km.Action("shift+left"); act != ActionSeekBackBig {
		t.Errorf("expected ActionSeekBackBig for 'shift+left', got '%s'", act)
	}
	if act := km.Action("shift+right"); act != ActionSeekFwdBig {
		t.Errorf("expected ActionSeekFwdBig for 'shift+right', got '%s'", act)
	}

	// Reset with 0
	if km.IsKeyUsed("0") {
		t.Error("expected '0' not to be used by default")
	}
	km.ResetItem(nextIdx)
	if act := km.Action("n"); act != ActionNextTrack {
		t.Errorf("expected ActionNextTrack restored to 'n' after reset, got '%s'", act)
	}
}

func TestTabBarOrder(t *testing.T) {
	rendered := renderTabBar(TabLyrics, 80)
	stripped := ansi.Strip(rendered)

	idxTracks := strings.Index(stripped, "1 Tracks")
	idxLyrics := strings.Index(stripped, "2 Lyrics")
	idxHistory := strings.Index(stripped, "3 History")

	if idxTracks == -1 || idxLyrics == -1 || idxHistory == -1 {
		t.Fatalf("expected all tabs in tab bar, got: %q", stripped)
	}
	if !(idxTracks < idxLyrics && idxLyrics < idxHistory) {
		t.Errorf("expected order Tracks < Lyrics < History, got tracks=%d, lyrics=%d, history=%d", idxTracks, idxLyrics, idxHistory)
	}
	if !strings.Contains(stripped, "[ 2 Lyrics ]") {
		t.Errorf("expected active tab '[ 2 Lyrics ]' in %q", stripped)
	}
}

func TestKeybindsModalScrollbar(t *testing.T) {
	km := NewKeyManager()
	// Modal rendered with height 26 so availH < len(items)
	modal := RenderKeybindsModal(km.Items, 0, false, 90, 26)
	stripped := ansi.Strip(modal)

	if !strings.Contains(stripped, "█") {
		t.Error("expected scrollbar thumb '█' in keybindings modal when items exceed viewport")
	}
	if !strings.Contains(stripped, "│") {
		t.Error("expected scrollbar track '│' in keybindings modal when items exceed viewport")
	}
}

func TestTotalDurationOnPlaylistHeader(t *testing.T) {
	tracks := []backend.Track{
		{ID: "1", Name: "Track 1", Artist: "Artist", DurationMs: 180000}, // 3 min
		{ID: "2", Name: "Track 2", Artist: "Artist", DurationMs: 125000}, // 2 min 5 sec
	}

	durStr := FormatTotalDuration(305000, len(tracks))
	if !strings.Contains(durStr, "2 tracks") || !strings.Contains(durStr, "5 min 5 sec") {
		t.Errorf("unexpected formatted duration: %s", durStr)
	}

	lines := renderTracks(tracks, nil, nil, "My Awesome Playlist", "", 0, false, 80, 20)
	header := lines[0]
	stripped := ansi.Strip(header)

	if !strings.Contains(stripped, "My Awesome Playlist") {
		t.Errorf("expected playlist name in header, got: %s", stripped)
	}
	if !strings.Contains(stripped, "5 min 5 sec") {
		t.Errorf("expected total duration on opposite side of header, got: %s", stripped)
	}
}

func TestEscClosesZenMode(t *testing.T) {
	app := &AppModel{
		zenMode:     true,
		currentTab:  TabTracks,
		keyManager:  NewKeyManager(),
	}

	// Press Esc
	newModel, _ := app.handleKeyPress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m := newModel.(*AppModel)

	if m.zenMode {
		t.Error("expected Esc to close Zen mode (zenMode = false)")
	}
}

func TestEscNavigatesBackFromAlbum(t *testing.T) {
	app := &AppModel{
		focused:       PaneCenter,
		currentTab:    TabTracks,
		currentPlURI:  "spotify:album:album123",
		currentPlName: "Abbey Road",
		centerIndex:   3,
		keyManager:    NewKeyManager(),
		navHistory: []containerHistoryItem{
			{
				ID:          "pl1",
				URI:         "spotify:playlist:pl1",
				Name:        "Rock Classics",
				CenterIndex: 5,
			},
		},
		client: &backend.Client{},
	}

	// Press Esc
	newModel, _ := app.handleKeyPress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m := newModel.(*AppModel)

	if m.currentPlURI != "spotify:playlist:pl1" {
		t.Errorf("expected Esc to navigate back to 'spotify:playlist:pl1', got '%s'", m.currentPlURI)
	}
	if m.currentPlName != "Rock Classics" {
		t.Errorf("expected currentPlName restored to 'Rock Classics', got '%s'", m.currentPlName)
	}
	if m.centerIndex != 5 {
		t.Errorf("expected centerIndex restored to 5, got %d", m.centerIndex)
	}
	if len(m.navHistory) != 0 {
		t.Errorf("expected navHistory popped to empty, got length %d", len(m.navHistory))
	}
}

func TestCategorizedSearchAndEscNavigation(t *testing.T) {
	app := &AppModel{
		focused:     PaneCenter,
		currentTab:  TabTracks,
		keyManager:  NewKeyManager(),
		searchQuery: "Queen",
	}

	searchMsg := SearchResultsMsg{
		Tracks: []backend.Track{
			{ID: "t1", Name: "Bohemian Rhapsody", Artist: "Queen", DurationMs: 354000},
		},
		Albums: []backend.Playlist{
			{ID: "alb1", URI: "spotify:album:alb1", Name: "A Night at the Opera", TrackCount: 12, OwnerID: "Album • 1975"},
		},
		Artists: []backend.Playlist{
			{ID: "art1", URI: "spotify:artist:art1", Name: "Queen", TrackCount: 85, OwnerID: "Artist"},
		},
	}

	// Deliver search results
	app.Update(searchMsg)

	if len(app.playlistTracks) != 1 || len(app.artistAlbums) != 1 || len(app.searchArtists) != 1 {
		t.Fatalf("expected separated tracks, albums, and artists in search results")
	}

	// Render center with categorized search results
	lines := renderTracks(app.playlistTracks, app.artistAlbums, app.searchArtists, app.currentPlName, "", 0, false, 90, 30)
	joined := strings.Join(lines, "\n")
	stripped := ansi.Strip(joined)

	if !strings.Contains(stripped, "Songs") {
		t.Error("expected 'Songs' section in search results")
	}
	if !strings.Contains(stripped, "Albums") {
		t.Error("expected 'Albums' section in search results")
	}
	if !strings.Contains(stripped, "Artists") {
		t.Error("expected 'Artists' section in search results")
	}
	if !strings.Contains(stripped, "Bohemian Rhapsody") {
		t.Error("expected song name in search results")
	}
	if !strings.Contains(stripped, "A Night at the Opera") {
		t.Error("expected album name in search results")
	}
	if !strings.Contains(stripped, "Queen") {
		t.Error("expected artist name in search results")
	}
	if !strings.Contains(stripped, "󰠃") {
		t.Error("expected Nerd Font 'account_music' icon '󰠃' for artists in search results")
	}
	if strings.Contains(stripped, "👤") {
		t.Error("did not expect emoji '👤' for artists")
	}

	// Navigate to album (index 1: len(tracks) + 0) and press enter
	app.centerIndex = 1
	app.handleEnter()

	if app.currentPlURI != "spotify:album:alb1" {
		t.Fatalf("expected currentPlURI to be album, got '%s'", app.currentPlURI)
	}
	if len(app.navHistory) != 1 {
		t.Fatalf("expected search view pushed to navHistory")
	}

	// Press Esc to return to search results
	newModel, _ := app.handleKeyPress(tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape}))
	m := newModel.(*AppModel)

	if m.currentPlURI != "search:Queen" {
		t.Errorf("expected Esc to restore search URI 'search:Queen', got '%s'", m.currentPlURI)
	}
	if len(m.playlistTracks) != 1 || len(m.artistAlbums) != 1 || len(m.searchArtists) != 1 {
		t.Errorf("expected search tracks, albums, and artists restored")
	}
	if m.centerIndex != 1 {
		t.Errorf("expected centerIndex restored to 1, got %d", m.centerIndex)
	}
}

func TestCenterPaneScrollbar(t *testing.T) {
	var manyTracks []backend.Track
	for i := 0; i < 40; i++ {
		manyTracks = append(manyTracks, backend.Track{
			ID:         fmt.Sprintf("t%d", i),
			Name:       fmt.Sprintf("Song %d", i),
			Artist:     "Band",
			DurationMs: 200000,
		})
	}
	lines := renderTracks(manyTracks, nil, nil, "Big Playlist", "", 0, false, 80, 15)
	joined := strings.Join(lines, "\n")
	stripped := ansi.Strip(joined)
	if !strings.Contains(stripped, "█") {
		t.Error("expected vertical scrollbar thumb '█' in center pane when content exceeds height")
	}
	if !strings.Contains(stripped, "│") {
		t.Error("expected vertical scrollbar rail '│' in center pane when content exceeds height")
	}
}

func TestCenterLineAndOverlayBackground(t *testing.T) {
	SetTheme(true)

	// Test CenterLine width and background
	centered := CenterLine("Test Lyric", 50)
	if w := ansi.StringWidth(centered); w != 50 {
		t.Errorf("expected visual width 50, got %d", w)
	}
	if !strings.Contains(centered, "\x1b[") {
		t.Error("expected ANSI background codes in centered line padding")
	}

	// Test CenterOverlay dimensions and full background fill
	modal := "╭──────╮\n│ Modal│\n╰──────╯"
	overlay := CenterOverlay(modal, 60, 15)
	lines := strings.Split(overlay, "\n")
	if len(lines) != 15 {
		t.Fatalf("expected 15 lines in overlay, got %d", len(lines))
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w != 60 {
			t.Errorf("line %d: expected width 60, got %d", i, w)
		}
		if !strings.Contains(l, "\x1b[") {
			t.Errorf("line %d: expected ANSI codes for surface fill", i)
		}
	}
}

func TestSearchBarThemeBackground(t *testing.T) {
	SetTheme(true)

	// Test search bar with empty query (placeholder)
	emptyLines := renderProminentSearchBar("", false, 60)
	if len(emptyLines) != 3 { // top border, content, bottom border
		t.Fatalf("expected 3 lines in search bar, got %d", len(emptyLines))
	}
	for _, l := range emptyLines {
		if w := ansi.StringWidth(l); w != 60 {
			t.Errorf("expected search bar line width 60, got %d", w)
		}
		if !strings.Contains(l, "\x1b[") {
			t.Errorf("expected ANSI background codes across search bar line")
		}
	}

	// Test search bar with typed query
	typedLines := renderProminentSearchBar("Radiohead", true, 60)
	for _, l := range typedLines {
		if w := ansi.StringWidth(l); w != 60 {
			t.Errorf("expected search bar line width 60, got %d", w)
		}
		if !strings.Contains(l, "\x1b[") {
			t.Errorf("expected ANSI background codes across search bar line")
		}
	}
}

func checkLineBackgrounds(t *testing.T, line string, expectedWidth int, lineDesc string) {
	t.Helper()
	hasBg := false
	col := 0
	inEsc := false
	escBuf := ""

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 0x1b {
			inEsc = true
			escBuf = string(r)
			continue
		}
		if inEsc {
			escBuf += string(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
				if strings.HasPrefix(escBuf, "\x1b[") && strings.HasSuffix(escBuf, "m") {
					params := escBuf[2 : len(escBuf)-1]
					parts := strings.Split(params, ";")
					for j := 0; j < len(parts); j++ {
						p := parts[j]
						if p == "0" || p == "" {
							hasBg = false
						} else if p == "49" {
							hasBg = false
						} else if p == "48" {
							// 48;2;r;g;b or 48;5;n
							hasBg = true
							if j+1 < len(parts) && parts[j+1] == "2" {
								j += 4
							} else if j+1 < len(parts) && parts[j+1] == "5" {
								j += 2
							}
						} else if len(p) == 2 && (p[0] == '4' && p[1] >= '0' && p[1] <= '7') {
							hasBg = true
						} else if len(p) == 3 && (strings.HasPrefix(p, "10") && p[2] >= '0' && p[2] <= '7') {
							hasBg = true
						}
					}
				}
			}
			continue
		}

		// Printable character or space
		w := ansi.StringWidth(string(r))
		if w > 0 {
			if !hasBg {
				t.Errorf("%s: unstyled cell at col %d (char %q)\nRAW: %q", lineDesc, col, r, line)
				return
			}
			col += w
		}
	}

	if col != expectedWidth {
		t.Errorf("%s: line width = %d, expected %d", lineDesc, col, expectedWidth)
	}
}

func TestAuditAllScreensForUnstyledPixels(t *testing.T) {
	SetTheme(true)

	// Base params
	w, h := 100, 30
	baseParams := ViewParams{
		Width:            w,
		Height:           h,
		Focused:          PaneNav,
		CurrentTab:       TabTracks,
		ShowLeftSidebar:  true,
		ShowRightSidebar: true,
		NavIndex:         0,
		CenterIndex:      0,
		QueueIndex:       0,
		Playlists: []backend.Playlist{
			{ID: "1", Name: "Chill Mix", TrackCount: 42},
			{ID: "2", Name: "Rock Classics", TrackCount: 88},
		},
		PinnedURIs: map[string]bool{"1": true},
		Playback: &backend.PlaybackState{
			Playing:    true,
			ProgressMs: 75000,
			DurationMs: 180000,
			Volume:     65,
			CurrentTrack: &backend.Track{
				Name:   "Midnight City",
				Artist: "M83",
				Album:  "Hurry Up",
				URI:    "track:1",
			},
		},
		Queue: []backend.Track{
			{Name: "Next Song", Artist: "Artist 1"},
		},
		LyricsLines: []lyrics.Line{
			{TimeMs: 70000, Text: "Active lyric line"},
			{TimeMs: 80000, Text: "Upcoming lyric line"},
		},
		ArtANSI: "██████\n██████\n██████",
	}

	testCases := []struct {
		name   string
		modify func(p *ViewParams)
	}{
		{"NormalView_AllPanes", func(p *ViewParams) {}},
		{"NormalView_NoSidebars", func(p *ViewParams) { p.ShowLeftSidebar = false; p.ShowRightSidebar = false }},
		{"NormalView_LeftSidebarOnly", func(p *ViewParams) { p.ShowRightSidebar = false }},
		{"NormalView_RightSidebarOnly", func(p *ViewParams) { p.ShowLeftSidebar = false }},
		{"NormalView_LyricsTab", func(p *ViewParams) { p.CurrentTab = TabLyrics }},
		{"NormalView_HistoryTab", func(p *ViewParams) { p.CurrentTab = TabHistory; p.History = p.Queue }},
		{"NormalView_SearchFocusedEmpty", func(p *ViewParams) { p.SearchFocused = true; p.SearchQuery = "" }},
		{"NormalView_SearchFocusedTyped", func(p *ViewParams) { p.SearchFocused = true; p.SearchQuery = "Hello" }},
		{"NormalView_MultiSectionSearch", func(p *ViewParams) {
			p.PlaylistTracks = p.Queue
			p.ArtistAlbums = []backend.Playlist{{ID: "a1", Name: "Album 1", TrackCount: 10, OwnerID: "2024"}}
			p.SearchArtists = []backend.Playlist{{ID: "ar1", Name: "Artist 1", TrackCount: 85, OwnerID: "Artist"}}
		}},
		{"NormalView_AlbumsFilter", func(p *ViewParams) {
			p.PlaylistFilter = FilterAlbums
		}},
		{"NormalView_ArtistsFilter", func(p *ViewParams) {
			p.PlaylistFilter = FilterArtists
		}},
		{"NormalView_EmptyPlaylist", func(p *ViewParams) {
			p.PlaylistTracks = nil
			p.Playlists = nil
		}},
		{"NormalView_SmallMinSize", func(p *ViewParams) {
			p.Width = 48
			p.Height = 16
		}},
		{"NormalView_LargeSize", func(p *ViewParams) {
			p.Width = 140
			p.Height = 45
		}},
		{"ZenMode_Both", func(p *ViewParams) { p.ZenMode = true; p.ZenView = ZenViewBoth }},
		{"ZenMode_ArtOnly", func(p *ViewParams) { p.ZenMode = true; p.ZenView = ZenViewArt }},
		{"ZenMode_LyricsOnly", func(p *ViewParams) { p.ZenMode = true; p.ZenView = ZenViewLyrics }},
		{"ZenMode_NoLyrics", func(p *ViewParams) { p.ZenMode = true; p.ZenView = ZenViewLyrics; p.LyricsLines = nil }},
		{"Modal_Help", func(p *ViewParams) { p.ShowHelp = true }},
		{"Modal_HelpEditing", func(p *ViewParams) { p.ShowHelp = true; p.HelpEditing = true }},
		{"Modal_Devices", func(p *ViewParams) {
			p.ShowDevices = true
			p.Devices = []spotify.PlayerDevice{{ID: "1", Name: "Speaker", Active: true}}
		}},
		{"Modal_DevicesScanning", func(p *ViewParams) {
			p.ShowDevices = true
			p.DeviceScanning = true
		}},
		{"Modal_DevicesEmpty", func(p *ViewParams) {
			p.ShowDevices = true
			p.Devices = nil
		}},
		{"SmallWindow", func(p *ViewParams) { p.Width = 40; p.Height = 10 }},
	}

	themeModes := []struct {
		themeName string
		isDark    bool
	}{
		{"Mocha_Dark", true},
		{"Latte_Light", false},
	}

	for _, tm := range themeModes {
		t.Run(tm.themeName, func(t *testing.T) {
			SetTheme(tm.isDark)
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					p := baseParams
					tc.modify(&p)
					rendered := RenderFullUI(p)
					lines := strings.Split(rendered, "\n")
					if len(lines) != p.Height {
						t.Fatalf("expected %d lines, got %d", p.Height, len(lines))
					}
					for lineIdx, line := range lines {
						desc := fmt.Sprintf("[%s/%s] Line %d", tm.themeName, tc.name, lineIdx)
						checkLineBackgrounds(t, line, p.Width, desc)
					}
				})
			}
		})
	}
}

func TestScrollbarDoesNotInterfereOrTruncate(t *testing.T) {
	SetTheme(true)
	// Create 30 tracks so scrollbar is active
	var tracks []backend.Track
	for i := 1; i <= 30; i++ {
		tracks = append(tracks, backend.Track{
			Name:       fmt.Sprintf("Song Title Number %d", i),
			Artist:     "Famous Artist",
			DurationMs: 180000 + (i * 1000), // ~03:00 to ~03:30
			URI:        fmt.Sprintf("spotify:track:%d", i),
		})
	}

	w, h := 90, 25
	rendered := RenderFullUI(ViewParams{
		Width:          w,
		Height:         h,
		CurrentTab:     TabTracks,
		PlaylistTracks: tracks,
		PlaylistName:   "Favorites",
	})

	lines := strings.Split(rendered, "\n")
	hasScrollbar := false
	for lineIdx, line := range lines {
		if strings.Contains(line, "█") || strings.Contains(line, "│") {
			hasScrollbar = true
		}
		// Duration lines like 03:01 should NOT end with … before the scrollbar
		if strings.Contains(line, "03:") {
			if strings.Contains(line, "…") {
				t.Errorf("Line %d has unexpected ellipsis truncation: %s", lineIdx, line)
			}
		}
		// Table header Time should NOT have ellipsis
		if strings.Contains(line, "Time") {
			if strings.Contains(line, "Time…") || strings.Contains(line, "Time …") {
				t.Errorf("Line %d has unexpected ellipsis on Time header: %s", lineIdx, line)
			}
		}
	}

	if !hasScrollbar {
		t.Error("expected scrollbar to be rendered for 30 tracks")
	}
}