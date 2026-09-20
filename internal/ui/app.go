package ui

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/zmb3/spotify/v2"
	"spotumn/internal/art"
	"spotumn/internal/backend"
	"spotumn/internal/config"
	"spotumn/internal/lyrics"
)

type TickMsg time.Time
type PlaybackMsg *backend.PlaybackState
type QueueMsg *backend.QueueData
type PlaylistsMsg []backend.Playlist
type TracksMsg struct {
	PlaylistURI  string
	PlaylistName string
	Tracks       []backend.Track
}
type HistoryMsg []backend.Track
type LyricsMsg []lyrics.Line
type ArtMsg struct {
	ANSI string
	Zen  bool
}
type DevicesMsg []spotify.PlayerDevice
type UserMsg struct {
	DisplayName string
	UserID      string
}
type SearchResultsMsg []backend.Track
type ErrorMsg error

type AppModel struct {
	client  *backend.Client
	artRen  *art.Renderer
	lyrProv *lyrics.Provider

	width  int
	height int

	focused    FocusedPane
	currentTab CenterTab

	showLeftSidebar  bool
	showRightSidebar bool
	zenMode          bool
	showHelp         bool
	showDevices      bool
	deviceScanning   bool

	devices     []spotify.PlayerDevice
	deviceIndex int

	username string
	userID   string

	navIndex     int
	centerIndex  int
	queueIndex   int
	lyricsCursor int

	searchFocused bool
	searchQuery   string

	playlists      []backend.Playlist
	pinnedURIs     map[string]bool
	playlistFilter PlaylistFilter
	currentPlURI   string
	currentPlName  string
	playlistTracks []backend.Track

	history []backend.Track

	playback *backend.PlaybackState
	queue    []backend.Track

	lyricsLines        []lyrics.Line
	lyricsManualScroll bool
	artANSI            string
	zenArtANSI         string
	lastArtURL         string
	lastTrackURI       string

	tickCount int
}

func getPinnedPath() string {
	return filepath.Join(config.GetDir(), "pinned.json")
}

func loadPinned() map[string]bool {
	res := make(map[string]bool)
	data, err := os.ReadFile(getPinnedPath())
	if err == nil {
		var list []string
		if json.Unmarshal(data, &list) == nil {
			for _, uri := range list {
				res[uri] = true
			}
		}
	}
	return res
}

func savePinned(pinned map[string]bool) {
	var list []string
	for uri, p := range pinned {
		if p {
			list = append(list, uri)
		}
	}
	data, _ := json.Marshal(list)
	_ = os.WriteFile(getPinnedPath(), data, 0600)
}

func sortPlaylistsWithPinned(playlists []backend.Playlist, pinned map[string]bool) []backend.Playlist {
	var pinnedList []backend.Playlist
	var unpinnedList []backend.Playlist
	for _, pl := range playlists {
		if pinned[pl.URI] {
			pinnedList = append(pinnedList, pl)
		} else {
			unpinnedList = append(unpinnedList, pl)
		}
	}
	return append(pinnedList, unpinnedList...)
}

func NewAppModel(client *backend.Client) *AppModel {
	m := &AppModel{
		client:           client,
		artRen:           art.NewRenderer(),
		lyrProv:          lyrics.NewProvider(),
		width:            100,
		height:           30,
		focused:          PaneCenter,
		currentTab:       TabTracks, // Default home is Tracks!
		showLeftSidebar:  true,
		showRightSidebar: true,
		navIndex:         0,
		centerIndex:      0,
		queueIndex:       0,
		lyricsCursor:     0,
		pinnedURIs:       loadPinned(),
		playlistFilter:   FilterAll,
		username:         "BrightestAutumn",
	}

	// Restore last active playback state on startup
	lastState := client.LoadLastState()
	if lastState != nil {
		m.playback = lastState
		if lastState.CurrentTrack != nil {
			m.lastArtURL = lastState.CurrentTrack.ArtURL
			m.lastTrackURI = lastState.CurrentTrack.URI
		}
	}

	return m
}

func (m *AppModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.RequestBackgroundColor,
		m.doTick(),
		m.fetchUserCmd(),
		m.fetchPlaylistsCmd(),
		m.fetchHistoryCmd(),
		m.fetchPlaybackCmd(),
		m.fetchQueueCmd(),
	}
	if m.lastArtURL != "" {
		cmds = append(cmds, m.fetchArtCmd(m.lastArtURL, 40, 20, false))
		cmds = append(cmds, m.fetchArtCmd(m.lastArtURL, 56, 26, true))
	}
	return tea.Batch(cmds...)
}

func (m *AppModel) doTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		SetTheme(msg.IsDark())
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case TickMsg:
		m.tickCount++
		// Advance local progress if playing for smooth seekbar animation
		if m.playback != nil && m.playback.Playing {
			m.playback.ProgressMs += 500
			if m.playback.DurationMs > 0 && m.playback.ProgressMs > m.playback.DurationMs {
				m.playback.ProgressMs = m.playback.DurationMs
			}
		}

		// Auto-follow lyrics pointer with singing line in real-time
		if !m.lyricsManualScroll && len(m.lyricsLines) > 0 && m.playback != nil {
			activeIdx := lyrics.FindActiveIndex(m.lyricsLines, m.playback.ProgressMs)
			if activeIdx >= 0 {
				m.lyricsCursor = activeIdx
			}
		}

		var cmds []tea.Cmd
		cmds = append(cmds, m.doTick())

		// Poll playback state every 3.5s (7 ticks)
		if m.tickCount%7 == 0 {
			cmds = append(cmds, m.fetchPlaybackCmd())
		}
		// Poll queue every 12s (24 ticks)
		if m.tickCount%24 == 0 {
			cmds = append(cmds, m.fetchQueueCmd())
		}
		// Live auto-rescan devices every 1.5s (3 ticks) while devices modal is open
		if m.showDevices && m.tickCount%3 == 0 {
			cmds = append(cmds, m.fetchDevicesCmd())
		}

		return m, tea.Batch(cmds...)

	case UserMsg:
		if msg.DisplayName != "" {
			m.username = msg.DisplayName
		}
		m.userID = msg.UserID
		return m, nil

	case PlaybackMsg:
		if msg != nil {
			m.playback = msg
			m.client.SaveLastState(msg)
			if msg.CurrentTrack != nil {
				if msg.CurrentTrack.URI != m.lastTrackURI {
					m.lastTrackURI = msg.CurrentTrack.URI
					m.lyricsCursor = 0
					m.lyricsManualScroll = false
					var cmds []tea.Cmd
					cmds = append(cmds, m.fetchQueueCmd())
					cmds = append(cmds, m.fetchLyricsCmd(msg.CurrentTrack.Name, msg.CurrentTrack.Artist, msg.CurrentTrack.DurationMs/1000))
					if msg.CurrentTrack.ArtURL != m.lastArtURL && msg.CurrentTrack.ArtURL != "" {
						m.lastArtURL = msg.CurrentTrack.ArtURL
						cmds = append(cmds, m.fetchArtCmd(msg.CurrentTrack.ArtURL, 40, 20, false))
						cmds = append(cmds, m.fetchArtCmd(msg.CurrentTrack.ArtURL, 56, 26, true))
					}
					return m, tea.Batch(cmds...)
				}
			}
		}
		return m, nil

	case QueueMsg:
		if msg != nil {
			m.queue = msg.Items
		}
		return m, nil

	case PlaylistsMsg:
		m.playlists = sortPlaylistsWithPinned(msg, m.pinnedURIs)
		// Automatically display contents of topmost playlist on the default tracks home page
		if len(m.playlists) > 0 && len(m.playlistTracks) == 0 {
			top := m.playlists[0]
			m.currentPlURI = top.URI
			m.currentPlName = top.Name
			return m, m.fetchPlaylistTracksCmd(top.ID, top.URI, top.Name)
		}
		return m, nil

	case HistoryMsg:
		m.history = msg
		return m, nil

	case TracksMsg:
		m.currentPlURI = msg.PlaylistURI
		if msg.PlaylistName != "" {
			m.currentPlName = msg.PlaylistName
		}
		m.playlistTracks = msg.Tracks
		m.centerIndex = 0
		m.currentTab = TabTracks
		return m, nil

	case LyricsMsg:
		m.lyricsLines = msg
		m.lyricsManualScroll = false
		if m.playback != nil {
			activeIdx := lyrics.FindActiveIndex(m.lyricsLines, m.playback.ProgressMs)
			if activeIdx >= 0 {
				m.lyricsCursor = activeIdx
			} else {
				m.lyricsCursor = 0
			}
		} else {
			m.lyricsCursor = 0
		}
		return m, nil

	case ArtMsg:
		if msg.Zen {
			m.zenArtANSI = msg.ANSI
		} else {
			m.artANSI = msg.ANSI
		}
		return m, nil

	case DevicesMsg:
		m.deviceScanning = false
		var oldID spotify.ID
		if m.deviceIndex >= 0 && m.deviceIndex < len(m.devices) {
			oldID = m.devices[m.deviceIndex].ID
		}
		m.devices = msg
		found := false
		if oldID != "" {
			for i, d := range m.devices {
				if d.ID == oldID {
					m.deviceIndex = i
					found = true
					break
				}
			}
		}
		if !found {
			m.deviceIndex = 0
			for i, d := range m.devices {
				if d.Active {
					m.deviceIndex = i
					break
				}
			}
		}
		return m, nil

	case SearchResultsMsg:
		m.playlistTracks = msg
		m.currentPlURI = ""
		m.currentPlName = "Search: " + m.searchQuery
		m.centerIndex = 0
		m.currentTab = TabTracks
		m.focused = PaneCenter
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m *AppModel) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Quit spotumn
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// Modal: Help
	if m.showHelp {
		if key == "?" || key == "esc" || key == "enter" {
			m.showHelp = false
		}
		return m, nil
	}

	// Modal: Devices
	if m.showDevices {
		switch key {
		case "esc", "d":
			m.showDevices = false
			m.deviceScanning = false
			return m, nil
		case "r":
			m.deviceScanning = true
			return m, m.fetchDevicesCmd()
		case "up", "k":
			if m.deviceIndex > 0 {
				m.deviceIndex--
			}
			return m, nil
		case "down", "j":
			if m.deviceIndex < len(m.devices)-1 {
				m.deviceIndex++
			}
			return m, nil
		case "enter":
			if m.deviceIndex >= 0 && m.deviceIndex < len(m.devices) {
				devID := m.devices[m.deviceIndex].ID
				m.deviceScanning = true
				return m, func() tea.Msg {
					_ = m.client.TransferPlayback(context.Background(), devID)
					time.Sleep(300 * time.Millisecond)
					devs, _ := m.client.GetDevices(context.Background())
					return DevicesMsg(devs)
				}
			}
		}
		return m, nil
	}

	// Search bar input mode
	if m.searchFocused {
		switch key {
		case "esc":
			m.searchFocused = false
			return m, nil
		case "enter":
			query := strings.TrimSpace(m.searchQuery)
			m.searchFocused = false
			if query != "" {
				return m, m.searchCmd(query)
			}
			return m, nil
		case "backspace":
			if len(m.searchQuery) > 0 {
				r := []rune(m.searchQuery)
				m.searchQuery = string(r[:len(r)-1])
			}
			return m, nil
		case "space":
			m.searchQuery += " "
			return m, nil
		default:
			text := msg.Key().Text
			if text != "" {
				m.searchQuery += text
				return m, nil
			} else if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.searchQuery += key
				return m, nil
			}
		}
		return m, nil
	}

	// Global Keybindings
	switch key {
	case "esc":
		if m.currentTab == TabLyrics {
			m.lyricsManualScroll = false
			if m.playback != nil && len(m.lyricsLines) > 0 {
				activeIdx := lyrics.FindActiveIndex(m.lyricsLines, m.playback.ProgressMs)
				if activeIdx >= 0 {
					m.lyricsCursor = activeIdx
				}
			}
		}
		return m, nil

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "d":
		m.showDevices = !m.showDevices
		if m.showDevices {
			m.deviceScanning = true
			return m, m.fetchDevicesCmd()
		}
		m.deviceScanning = false
		return m, nil

	case "z":
		m.zenMode = !m.zenMode
		if m.zenMode && m.lastArtURL != "" && m.zenArtANSI == "" {
			return m, m.fetchArtCmd(m.lastArtURL, 56, 26, true)
		}
		return m, nil

	case "H", "shift+h":
		// Smart sidebar toggle
		if m.focused == PaneNav {
			m.showLeftSidebar = false
			m.focused = PaneCenter
		} else if m.focused == PaneRight {
			m.showRightSidebar = false
			m.focused = PaneCenter
		} else {
			if !m.showLeftSidebar || !m.showRightSidebar {
				m.showLeftSidebar = true
				m.showRightSidebar = true
			} else {
				m.showRightSidebar = false
			}
		}
		return m, nil

	case "/":
		m.searchFocused = true
		return m, nil

	case "1":
		m.currentTab = TabTracks
		m.centerIndex = 0
		return m, nil

	case "2":
		m.currentTab = TabHistory
		m.centerIndex = 0
		return m, m.fetchHistoryCmd()

	case "3":
		m.currentTab = TabLyrics
		m.lyricsManualScroll = false
		if m.playback != nil && len(m.lyricsLines) > 0 {
			activeIdx := lyrics.FindActiveIndex(m.lyricsLines, m.playback.ProgressMs)
			if activeIdx >= 0 {
				m.lyricsCursor = activeIdx
			}
		}
		return m, nil

	case "[":
		m.cycleFocus(-1)
		return m, nil

	case "]":
		m.cycleFocus(1)
		return m, nil

	case "f", "t":
		if m.focused == PaneNav {
			m.playlistFilter = (m.playlistFilter + 1) % 4
			m.navIndex = 0
			return m, nil
		}

	case "*":
		if m.focused == PaneNav {
			pls := m.filteredPlaylists()
			if len(pls) > 0 && m.navIndex >= 0 && m.navIndex < len(pls) {
				pl := pls[m.navIndex]
				m.pinnedURIs[pl.URI] = !m.pinnedURIs[pl.URI]
				savePinned(m.pinnedURIs)
				m.playlists = sortPlaylistsWithPinned(m.playlists, m.pinnedURIs)
				return m, nil
			}
		}

	case "q":
		// Queue selected track to Spotify queue
		var targetTrack *backend.Track
		if m.focused == PaneCenter {
			if m.currentTab == TabTracks {
				if m.centerIndex >= 0 && m.centerIndex < len(m.playlistTracks) {
					targetTrack = &m.playlistTracks[m.centerIndex]
				}
			} else if m.currentTab == TabHistory {
				if m.centerIndex >= 0 && m.centerIndex < len(m.history) {
					targetTrack = &m.history[m.centerIndex]
				}
			}
		} else if m.focused == PaneRight {
			if m.queueIndex >= 0 && m.queueIndex < len(m.queue) {
				targetTrack = &m.queue[m.queueIndex]
			}
		}
		if targetTrack != nil {
			uri := targetTrack.URI
			return m, func() tea.Msg {
				_ = m.client.QueueSong(context.Background(), uri)
				time.Sleep(200 * time.Millisecond)
				q, _ := m.client.GetQueue(context.Background())
				return QueueMsg(q)
			}
		}
		return m, nil

	case "space":
		isPlaying := false
		if m.playback != nil {
			isPlaying = m.playback.Playing
		}
		return m, func() tea.Msg {
			_ = m.client.PlayPause(context.Background(), isPlaying)
			time.Sleep(150 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case "p":
		return m, func() tea.Msg {
			_ = m.client.Previous(context.Background())
			time.Sleep(200 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case "n":
		return m, func() tea.Msg {
			_ = m.client.Next(context.Background())
			time.Sleep(200 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case "s":
		curShuffle := false
		if m.playback != nil {
			curShuffle = m.playback.Shuffle
		}
		return m, func() tea.Msg {
			_ = m.client.ToggleShuffle(context.Background(), curShuffle)
			time.Sleep(150 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case "r":
		curRepeat := "off"
		if m.playback != nil {
			curRepeat = m.playback.Repeat
		}
		return m, func() tea.Msg {
			_ = m.client.CycleRepeat(context.Background(), curRepeat)
			time.Sleep(150 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case "+", "=":
		vol := 50
		if m.playback != nil {
			vol = m.playback.Volume
		}
		vol += 5
		if vol > 100 {
			vol = 100
		}
		if m.playback != nil {
			m.playback.Volume = vol
		}
		return m, func() tea.Msg {
			_ = m.client.SetVolume(context.Background(), vol)
			return nil
		}

	case "-":
		vol := 50
		if m.playback != nil {
			vol = m.playback.Volume
		}
		vol -= 5
		if vol < 0 {
			vol = 0
		}
		if m.playback != nil {
			m.playback.Volume = vol
		}
		return m, func() tea.Msg {
			_ = m.client.SetVolume(context.Background(), vol)
			return nil
		}

	case "left", "<", ",":
		pos := 0
		if m.playback != nil {
			pos = m.playback.ProgressMs - 5000
			if pos < 0 {
				pos = 0
			}
			m.playback.ProgressMs = pos
		}
		return m, func() tea.Msg {
			_ = m.client.Seek(context.Background(), pos)
			return nil
		}

	case "right", ">", ".":
		pos := 0
		if m.playback != nil {
			pos = m.playback.ProgressMs + 5000
			if m.playback.DurationMs > 0 && pos > m.playback.DurationMs {
				pos = m.playback.DurationMs
			}
			m.playback.ProgressMs = pos
		}
		return m, func() tea.Msg {
			_ = m.client.Seek(context.Background(), pos)
			return nil
		}

	case "j", "down":
		m.moveCursor(1)
		return m, nil

	case "k", "up":
		m.moveCursor(-1)
		return m, nil

	case "enter":
		return m.handleEnter()
	}

	return m, nil
}

func (m *AppModel) filteredPlaylists() []backend.Playlist {
	if m.playlistFilter == FilterAll {
		return m.playlists
	}
	var res []backend.Playlist
	for _, pl := range m.playlists {
		switch m.playlistFilter {
		case FilterByYou:
			if m.userID != "" && (pl.OwnerID == m.userID || strings.EqualFold(pl.OwnerID, m.username)) {
				res = append(res, pl)
			}
		case FilterSpotify:
			if strings.EqualFold(pl.OwnerID, "spotify") {
				res = append(res, pl)
			}
		case FilterSaved:
			// Show playlists saved from other users/artists or pinned
			isOwn := m.userID != "" && (pl.OwnerID == m.userID || strings.EqualFold(pl.OwnerID, m.username))
			if !isOwn || m.pinnedURIs[pl.URI] {
				res = append(res, pl)
			}
		default:
			res = append(res, pl)
		}
	}
	return res
}

func (m *AppModel) cycleFocus(delta int) {
	var panes []FocusedPane
	if m.showLeftSidebar {
		panes = append(panes, PaneNav)
	}
	panes = append(panes, PaneCenter)
	if m.showRightSidebar {
		panes = append(panes, PaneRight)
	}
	panes = append(panes, PanePlayer)

	curIdx := 0
	for i, p := range panes {
		if p == m.focused {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + delta + len(panes)) % len(panes)
	m.focused = panes[nextIdx]
}

func (m *AppModel) moveCursor(delta int) {
	switch m.focused {
	case PaneNav:
		pls := m.filteredPlaylists()
		newIdx := m.navIndex + delta
		if newIdx >= 0 && newIdx < len(pls) {
			m.navIndex = newIdx
		}

	case PaneCenter:
		switch m.currentTab {
		case TabTracks:
			newIdx := m.centerIndex + delta
			if newIdx >= 0 && newIdx < len(m.playlistTracks) {
				m.centerIndex = newIdx
			}
		case TabHistory:
			newIdx := m.centerIndex + delta
			if newIdx >= 0 && newIdx < len(m.history) {
				m.centerIndex = newIdx
			}
		case TabLyrics:
			m.lyricsManualScroll = true
			newIdx := m.lyricsCursor + delta
			if newIdx >= 0 && newIdx < len(m.lyricsLines) {
				m.lyricsCursor = newIdx
			}
		}

	case PaneRight:
		newIdx := m.queueIndex + delta
		if newIdx >= 0 && newIdx < len(m.queue) {
			m.queueIndex = newIdx
		}
	}
}

func (m *AppModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focused {
	case PaneNav:
		pls := m.filteredPlaylists()
		if m.navIndex >= 0 && m.navIndex < len(pls) {
			pl := pls[m.navIndex]
			m.currentPlURI = pl.URI
			m.currentPlName = pl.Name
			m.currentTab = TabTracks
			m.focused = PaneCenter
			return m, m.fetchPlaylistTracksCmd(pl.ID, pl.URI, pl.Name)
		}

	case PaneCenter:
		switch m.currentTab {
		case TabTracks:
			if m.centerIndex >= 0 && m.centerIndex < len(m.playlistTracks) {
				idx := m.centerIndex
				tracks := m.playlistTracks
				plURI := m.currentPlURI
				return m, func() tea.Msg {
					_ = m.client.PlayTrackList(context.Background(), tracks, idx, plURI)
					time.Sleep(200 * time.Millisecond)
					st, _ := m.client.GetPlaybackState(context.Background())
					return PlaybackMsg(st)
				}
			}

		case TabHistory:
			if m.centerIndex >= 0 && m.centerIndex < len(m.history) {
				idx := m.centerIndex
				tracks := m.history
				return m, func() tea.Msg {
					_ = m.client.PlayTrackList(context.Background(), tracks, idx, "")
					time.Sleep(200 * time.Millisecond)
					st, _ := m.client.GetPlaybackState(context.Background())
					return PlaybackMsg(st)
				}
			}

		case TabLyrics:
			if m.lyricsCursor >= 0 && m.lyricsCursor < len(m.lyricsLines) {
				timeMs := m.lyricsLines[m.lyricsCursor].TimeMs
				m.lyricsManualScroll = false
				if m.playback != nil {
					m.playback.ProgressMs = timeMs
				}
				return m, func() tea.Msg {
					_ = m.client.Seek(context.Background(), timeMs)
					return nil
				}
			}

		}

	case PaneRight:
		if m.queueIndex >= 0 && m.queueIndex < len(m.queue) {
			track := m.queue[m.queueIndex]
			return m, func() tea.Msg {
				_ = m.client.PlayTrack(context.Background(), track.URI, "")
				time.Sleep(200 * time.Millisecond)
				st, _ := m.client.GetPlaybackState(context.Background())
				return PlaybackMsg(st)
			}
		}

	case PanePlayer:
		isPlaying := false
		if m.playback != nil {
			isPlaying = m.playback.Playing
		}
		return m, func() tea.Msg {
			_ = m.client.PlayPause(context.Background(), isPlaying)
			time.Sleep(150 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}
	}

	return m, nil
}

func (m *AppModel) View() tea.View {
	rendered := RenderFullUI(ViewParams{
		Width:            m.width,
		Height:           m.height,
		Focused:          m.focused,
		CurrentTab:       m.currentTab,
		ShowLeftSidebar:  m.showLeftSidebar,
		ShowRightSidebar: m.showRightSidebar,
		ZenMode:          m.zenMode,
		ShowHelp:         m.showHelp,
		ShowDevices:      m.showDevices,
		DeviceScanning:   m.deviceScanning,
		Devices:          m.devices,
		DeviceIndex:      m.deviceIndex,
		Username:         m.username,
		NavIndex:         m.navIndex,
		CenterIndex:      m.centerIndex,
		QueueIndex:       m.queueIndex,
		LyricsCursor:     m.lyricsCursor,
		SearchFocused:    m.searchFocused,
		SearchQuery:      m.searchQuery,
		Playlists:        m.filteredPlaylists(),
		PinnedURIs:       m.pinnedURIs,
		PlaylistFilter:   m.playlistFilter,
		PlaylistTracks:   m.playlistTracks,
		PlaylistName:     m.currentPlName,
		History:          m.history,
		Playback:         m.playback,
		Queue:            m.queue,
		LyricsLines:      m.lyricsLines,
		ArtANSI:          m.artANSI,
		ZenArtANSI:       m.zenArtANSI,
	})

	v := tea.NewView(rendered)
	v.AltScreen = true
	v.WindowTitle = "spotumn"
	// Native terminal background (zero background color override)
	return v
}

// Commands
func (m *AppModel) fetchUserCmd() tea.Cmd {
	return func() tea.Msg {
		name, id := m.client.GetCurrentUser(context.Background())
		return UserMsg{DisplayName: name, UserID: id}
	}
}

func (m *AppModel) fetchDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		devs, err := m.client.GetDevices(context.Background())
		if err != nil {
			return DevicesMsg(nil)
		}
		return DevicesMsg(devs)
	}
}

func (m *AppModel) fetchPlaylistsCmd() tea.Cmd {
	return func() tea.Msg {
		playlists, err := m.client.GetAllPlaylists(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}
		return PlaylistsMsg(playlists)
	}
}

func (m *AppModel) fetchHistoryCmd() tea.Cmd {
	return func() tea.Msg {
		recent, _ := m.client.GetRecentlyPlayed(context.Background())
		return HistoryMsg(recent)
	}
}

func (m *AppModel) fetchPlaybackCmd() tea.Cmd {
	return func() tea.Msg {
		st, err := m.client.GetPlaybackState(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}
		return PlaybackMsg(st)
	}
}

func (m *AppModel) fetchQueueCmd() tea.Cmd {
	return func() tea.Msg {
		q, err := m.client.GetQueue(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}
		return QueueMsg(q)
	}
}

func (m *AppModel) fetchPlaylistTracksCmd(plID, plURI, plName string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.client.GetPlaylistTracks(context.Background(), plID)
		if err != nil {
			return ErrorMsg(err)
		}
		return TracksMsg{
			PlaylistURI:  plURI,
			PlaylistName: plName,
			Tracks:       tracks,
		}
	}
}

func (m *AppModel) fetchLyricsCmd(trackName, artistName string, durSec int) tea.Cmd {
	return func() tea.Msg {
		lines, _ := m.lyrProv.FetchSyncedLyrics(trackName, artistName, durSec)
		return LyricsMsg(lines)
	}
}

func (m *AppModel) fetchArtCmd(url string, w, h int, zen bool) tea.Cmd {
	return func() tea.Msg {
		ansi, _ := m.artRen.Render(url, w, h)
		return ArtMsg{ANSI: ansi, Zen: zen}
	}
}

func (m *AppModel) searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		query = strings.TrimSpace(query)
		if query == "" {
			return SearchResultsMsg(nil)
		}
		tracks, _, _ := m.client.Search(context.Background(), query)
		return SearchResultsMsg(tracks)
	}
}
