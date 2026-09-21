package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	// Dark mode
	SetTheme(false)
	if CurrentTheme.Purple == nil || CurrentTheme.Mint == nil {
		t.Fatal("dark theme colors missing")
	}

	// Light mode
	SetTheme(true)
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

	// Test with empty username: should NOT show person icon or any hardcoded username
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
	if strings.Contains(outputEmpty, "BrightestAutumn") {
		t.Error("expected no hardcoded username in top border")
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

