package ui

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"spotumn/internal/art"
	"spotumn/internal/backend"
	"spotumn/internal/config"
	"spotumn/internal/lyrics"

	tea "charm.land/bubbletea/v2"
	"github.com/zmb3/spotify/v2"
)

type TickMsg time.Time
type PlaybackMsg *backend.PlaybackState
type QueueMsg *backend.QueueData
type PlaylistsMsg []backend.Playlist
type containerHistoryItem struct {
	ID          string
	URI         string
	Name        string
	CenterIndex int
	Tracks      []backend.Track
	Albums      []backend.Playlist
	Artists     []backend.Playlist
}

type TracksMsg struct {
	PlaylistURI  string
	PlaylistName string
	Tracks       []backend.Track
	Albums       []backend.Playlist
}
type HistoryMsg []backend.Track
type LyricsMsg []lyrics.Line
type ArtMsg struct {
	ANSI     string
	DiskPath string
	Width    int
	Height   int
	Zen      bool
}
type DevicesMsg []spotify.PlayerDevice
type UserMsg struct {
	DisplayName string
	UserID      string
}
type SearchResultsMsg struct {
	Tracks  []backend.Track
	Albums  []backend.Playlist
	Artists []backend.Playlist
}
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
	zenView          ZenViewMode
	showHelp         bool
	helpIndex        int
	helpEditing      bool
	keyManager       *KeyManager
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
	albums         []backend.Playlist
	artists        []backend.Playlist
	pinnedURIs     map[string]bool
	playlistFilter PlaylistFilter
	currentPlURI   string
	currentPlName  string
	playlistTracks []backend.Track
	artistAlbums   []backend.Playlist
	searchArtists  []backend.Playlist
	navHistory     []containerHistoryItem

	history []backend.Track

	playback *backend.PlaybackState
	queue    []backend.Track

	lyricsLines          []lyrics.Line
	lyricsManualScroll   bool
	lyricsPointerMovedAt time.Time
	artANSI              string
	zenArtANSI           string
	lastArtURL           string
	lastDiskPath         string
	lastTrackURI         string

	tickCount int

	volumeTarget int
	seekTarget   int
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

func NewAppModel(client *backend.Client, cfg ...*config.Config) *AppModel {
	artMode := "auto"
	if len(cfg) > 0 && cfg[0] != nil && cfg[0].ArtRenderer != "" {
		artMode = cfg[0].ArtRenderer
	} else if c, err := config.Load(); err == nil && c.ArtRenderer != "" {
		artMode = c.ArtRenderer
	}

	m := &AppModel{
		client:           client,
		artRen:           art.NewRenderer(artMode),
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
		username:         "",
		keyManager:       NewKeyManager(),
	}

	// Restore last active playback state on startup in paused state
	lastState := client.LoadLastState()
	if lastState != nil {
		lastState.Playing = false // Opens paused, requiring manual play
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
		m.fetchAlbumsCmd(),
		m.fetchArtistsCmd(),
		m.fetchHistoryCmd(),
		m.fetchPlaybackCmd(),
		m.fetchQueueCmd(),
		m.defaultDeviceCmd(),
	}
	if m.lastArtURL != "" {
		cmds = append(cmds, m.fetchArtCmd(m.lastArtURL, 40, 20, false))
		cmds = append(cmds, m.fetchArtCmd(m.lastArtURL, 56, 26, true))
	}
	if m.playback != nil && m.playback.CurrentTrack != nil {
		cmds = append(cmds, m.fetchLyricsCmd(m.playback.CurrentTrack.Name, m.playback.CurrentTrack.Artist, m.playback.DurationMs/1000))
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
			if m.tickCount%4 == 0 {
				m.client.SaveLastState(m.playback)
			}
		}

		// Auto-revert lyrics pointer to live sync if left untouched for 3s
		if m.lyricsManualScroll && len(m.lyricsLines) > 0 && m.playback != nil {
			if time.Since(m.lyricsPointerMovedAt) >= 3*time.Second {
				m.lyricsManualScroll = false
				activeIdx := lyrics.FindActiveIndex(m.lyricsLines, m.playback.ProgressMs)
				if activeIdx >= 0 {
					m.lyricsCursor = activeIdx
				}
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
		} else if msg.UserID != "" {
			m.username = msg.UserID
		}
		m.userID = msg.UserID
		return m, nil

	case PlaybackMsg:
		if msg != nil {
			if msg.CurrentTrack != nil {
				m.playback = msg
				m.client.SaveLastState(msg)
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
			} else if m.playback != nil && m.playback.CurrentTrack != nil {
				// Keep loaded track & exact timestamp when Spotify returns empty state on startup
				m.playback.Playing = false
			} else {
				m.playback = msg
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

	case AlbumsMsg:
		m.albums = sortPlaylistsWithPinned(msg, m.pinnedURIs)
		if m.playlistFilter == FilterAlbums && len(m.albums) > 0 && len(m.playlistTracks) == 0 {
			top := m.albums[0]
			m.currentPlURI = top.URI
			m.currentPlName = top.Name
			return m, m.fetchPlaylistTracksCmd(top.ID, top.URI, top.Name)
		}
		return m, nil

	case ArtistsMsg:
		m.artists = sortPlaylistsWithPinned(msg, m.pinnedURIs)
		if m.playlistFilter == FilterArtists && len(m.artists) > 0 && len(m.playlistTracks) == 0 {
			top := m.artists[0]
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
		m.artistAlbums = msg.Albums
		m.centerIndex = 0
		m.currentTab = TabTracks
		return m, nil

	case openArtistResolvedMsg:
		if msg.Artist.ID != "" {
			m.currentPlURI = msg.Artist.URI
			m.currentPlName = msg.Artist.Name
			m.centerIndex = 0
			m.searchArtists = nil
			return m, m.fetchPlaylistTracksCmd(msg.Artist.ID, msg.Artist.URI, msg.Artist.Name)
		}
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
		if msg.DiskPath != "" {
			m.lastDiskPath = msg.DiskPath
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
		m.playlistTracks = msg.Tracks
		m.artistAlbums = msg.Albums
		m.searchArtists = msg.Artists
		m.currentPlURI = "search:" + m.searchQuery
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

func (m *AppModel) getKeyManager() *KeyManager {
	if m.keyManager == nil {
		m.keyManager = NewKeyManager()
	}
	return m.keyManager
}

func (m *AppModel) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	km := m.getKeyManager()
	act := km.Action(key)

	// Quit spotumn
	if act == ActionQuit || key == "ctrl+c" {
		if m.playback != nil && m.playback.CurrentTrack != nil {
			m.client.SaveLastState(m.playback)
		}
		return m, tea.Quit
	}

	// Modal: Help / Keybinds
	if m.showHelp {
		if m.helpEditing {
			if key == "esc" {
				m.helpEditing = false
				return m, nil
			}
			km.SetKey(m.helpIndex, key)
			_ = km.Save()
			m.helpEditing = false
			return m, nil
		}

		switch key {
		case "?", "esc":
			m.showHelp = false
			return m, nil
		case "up", "k":
			if m.helpIndex > 0 {
				m.helpIndex--
			}
			return m, nil
		case "down", "j":
			if m.helpIndex < len(km.Items)-1 {
				m.helpIndex++
			}
			return m, nil
		case "enter":
			m.helpEditing = true
			return m, nil
		case "0":
			// Reset to default if '0' is not used by an active keybind
			if !km.IsKeyUsed("0") {
				if km.Items[m.helpIndex].Key != km.Items[m.helpIndex].DefaultKey {
					km.ResetItem(m.helpIndex)
				} else {
					km.ResetAll()
				}
				_ = km.Save()
			}
			return m, nil
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

	// Global Keybindings dispatched through custom actions or fallback
	switch act {
	case ActionOpenArtist:
		if m.playback == nil || m.playback.CurrentTrack == nil || m.playback.CurrentTrack.Name == "" {
			return m, nil
		}
		track := m.playback.CurrentTrack
		artistID := track.ArtistID
		artistName := track.Artist
		if artistID == "" && artistName == "" {
			return m, nil
		}

		if m.zenMode {
			m.zenMode = false
		}
		if m.showHelp {
			m.showHelp = false
		}
		if m.showDevices {
			m.showDevices = false
		}

		currID := ""
		if parts := strings.Split(m.currentPlURI, ":"); len(parts) >= 3 {
			currID = parts[2]
		}
		if currID != artistID {
			m.navHistory = append(m.navHistory, containerHistoryItem{
				ID:          currID,
				URI:         m.currentPlURI,
				Name:        m.currentPlName,
				CenterIndex: m.centerIndex,
				Tracks:      m.playlistTracks,
				Albums:      m.artistAlbums,
				Artists:     m.searchArtists,
			})
		}

		m.currentTab = TabTracks
		m.focused = PaneCenter
		m.centerIndex = 0
		m.searchArtists = nil
		m.searchQuery = ""
		m.searchFocused = false

		if artistID != "" {
			artURI := "spotify:artist:" + artistID
			m.currentPlURI = artURI
			m.currentPlName = artistName
			return m, m.fetchPlaylistTracksCmd(artistID, artURI, artistName)
		} else {
			return m, func() tea.Msg {
				art, err := m.client.GetArtistByName(context.Background(), artistName)
				if err != nil || art == nil {
					return nil
				}
				return openArtistResolvedMsg{Artist: *art}
			}
		}

	case ActionBack:
		if len(m.navHistory) > 0 {
			prev := m.navHistory[len(m.navHistory)-1]
			m.navHistory = m.navHistory[:len(m.navHistory)-1]
			m.currentPlURI = prev.URI
			m.currentPlName = prev.Name
			m.centerIndex = prev.CenterIndex
			if strings.HasPrefix(prev.URI, "search:") {
				m.playlistTracks = prev.Tracks
				m.artistAlbums = prev.Albums
				m.searchArtists = prev.Artists
				return m, nil
			}
			return m, m.fetchPlaylistTracksCmd(prev.ID, prev.URI, prev.Name)
		}
		return m, nil

	case ActionEscape:
		if m.showHelp {
			m.showHelp = false
			m.helpEditing = false
			return m, nil
		}
		if m.showDevices {
			m.showDevices = false
			m.deviceScanning = false
			return m, nil
		}
		if m.zenMode {
			m.zenMode = false
			m.lyricsManualScroll = false
			return m, nil
		}
		if len(m.navHistory) > 0 {
			prev := m.navHistory[len(m.navHistory)-1]
			m.navHistory = m.navHistory[:len(m.navHistory)-1]
			m.currentPlURI = prev.URI
			m.currentPlName = prev.Name
			m.centerIndex = prev.CenterIndex
			if strings.HasPrefix(prev.URI, "search:") {
				m.playlistTracks = prev.Tracks
				m.artistAlbums = prev.Albums
				m.searchArtists = prev.Artists
				return m, nil
			}
			return m, m.fetchPlaylistTracksCmd(prev.ID, prev.URI, prev.Name)
		}
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

	case ActionHelp:
		m.showHelp = !m.showHelp
		return m, nil

	case ActionDevices:
		m.showDevices = !m.showDevices
		if m.showDevices {
			m.deviceScanning = true
			return m, m.fetchDevicesCmd()
		}
		m.deviceScanning = false
		return m, nil

	case ActionZenLayout:
		if m.zenMode {
			m.zenView = (m.zenView + 1) % 3
			if m.zenView != ZenViewLyrics && m.lastArtURL != "" {
				_, _, w, h := m.getZenArtGeometry()
				if w > 0 && h > 0 {
					return m, m.fetchArtCmd(m.lastArtURL, w, h, true)
				}
			}
			return m, nil
		}

	case ActionZenMode:
		m.zenMode = !m.zenMode
		if m.zenMode && m.zenView != ZenViewLyrics && m.lastArtURL != "" {
			_, _, w, h := m.getZenArtGeometry()
			if w > 0 && h > 0 {
				return m, m.fetchArtCmd(m.lastArtURL, w, h, true)
			}
		}
		return m, nil

	case ActionToggleSidebar:
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

	case ActionSearch:
		m.searchFocused = true
		return m, nil

	case ActionTabTracks:
		m.currentTab = TabTracks
		m.centerIndex = 0
		return m, nil

	case ActionTabLyrics:
		m.currentTab = TabLyrics
		m.lyricsManualScroll = false
		if m.playback != nil && len(m.lyricsLines) > 0 {
			activeIdx := lyrics.FindActiveIndex(m.lyricsLines, m.playback.ProgressMs)
			if activeIdx >= 0 {
				m.lyricsCursor = activeIdx
			}
		}
		return m, nil

	case ActionTabHistory:
		m.currentTab = TabHistory
		m.centerIndex = 0
		return m, m.fetchHistoryCmd()

	case ActionFocusPrev:
		m.cycleFocus(-1)
		return m, nil

	case ActionFocusNext:
		m.cycleFocus(1)
		return m, nil

	case ActionFilter:
		if m.focused == PaneNav {
			m.playlistFilter = (m.playlistFilter + 1) % 5
			m.navIndex = 0
			m.navHistory = nil
			items := m.filteredPlaylists()
			if len(items) > 0 {
				top := items[0]
				m.currentPlURI = top.URI
				m.currentPlName = top.Name
				return m, m.fetchPlaylistTracksCmd(top.ID, top.URI, top.Name)
			}
			return m, nil
		}

	case ActionPin:
		if m.focused == PaneNav {
			pls := m.filteredPlaylists()
			if len(pls) > 0 && m.navIndex >= 0 && m.navIndex < len(pls) {
				pl := pls[m.navIndex]
				m.pinnedURIs[pl.URI] = !m.pinnedURIs[pl.URI]
				savePinned(m.pinnedURIs)
				m.playlists = sortPlaylistsWithPinned(m.playlists, m.pinnedURIs)
				m.albums = sortPlaylistsWithPinned(m.albums, m.pinnedURIs)
				m.artists = sortPlaylistsWithPinned(m.artists, m.pinnedURIs)
				return m, nil
			}
		}

	case ActionQueueTrack:
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

	case ActionPlayPause:
		isPlaying := false
		if m.playback != nil {
			isPlaying = m.playback.Playing
			m.playback.Playing = !isPlaying
			m.client.SaveLastState(m.playback)
		}
		trackURI := ""
		contextURI := ""
		progressMs := 0
		if m.playback != nil && m.playback.CurrentTrack != nil {
			trackURI = m.playback.CurrentTrack.URI
			contextURI = m.playback.ContextURI
			progressMs = m.playback.ProgressMs
		}
		return m, func() tea.Msg {
			if isPlaying {
				_ = m.client.Pause(context.Background())
			} else {
				err := m.client.Play(context.Background())
				if err != nil && trackURI != "" {
					_ = m.client.PlayTrackAtPosition(context.Background(), trackURI, contextURI, progressMs)
				}
			}
			time.Sleep(150 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			if st != nil && st.CurrentTrack != nil {
				return PlaybackMsg(st)
			}
			return nil
		}

	case ActionPrevTrack:
		return m, func() tea.Msg {
			_ = m.client.Previous(context.Background())
			time.Sleep(200 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case ActionNextTrack:
		return m, func() tea.Msg {
			_ = m.client.Next(context.Background())
			time.Sleep(200 * time.Millisecond)
			st, _ := m.client.GetPlaybackState(context.Background())
			return PlaybackMsg(st)
		}

	case ActionShuffle:
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

	case ActionRepeat:
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

	case ActionVolumeUp, ActionVolumeDown, ActionVolumeUpBig, ActionVolumeDownBig:
		delta := 5
		if act == ActionVolumeDown {
			delta = -5
		} else if act == ActionVolumeUpBig {
			delta = 10
		} else if act == ActionVolumeDownBig {
			delta = -10
		}
		vol := 50
		if m.playback != nil {
			vol = m.playback.Volume
		}
		vol += delta
		if vol > 100 {
			vol = 100
		}
		if vol < 0 {
			vol = 0
		}
		if m.playback != nil {
			m.playback.Volume = vol
			m.client.SaveLastState(m.playback)
		}
		m.volumeTarget = vol
		target := vol
		return m, func() tea.Msg {
			time.Sleep(200 * time.Millisecond)
			if m.volumeTarget == target {
				_ = m.client.SetVolume(context.Background(), target)
			}
			return nil
		}

	case ActionSeekBack, ActionSeekFwd, ActionSeekBackBig, ActionSeekFwdBig:
		deltaMs := 5000
		if act == ActionSeekBack {
			deltaMs = -5000
		} else if act == ActionSeekBackBig {
			deltaMs = -10000
		} else if act == ActionSeekFwdBig {
			deltaMs = 10000
		}
		pos := 0
		if m.playback != nil {
			pos = m.playback.ProgressMs + deltaMs
			if pos < 0 {
				pos = 0
			}
			if m.playback.DurationMs > 0 && pos > m.playback.DurationMs {
				pos = m.playback.DurationMs
			}
			m.playback.ProgressMs = pos
			m.client.SaveLastState(m.playback)
		}
		m.seekTarget = pos
		target := pos
		return m, func() tea.Msg {
			time.Sleep(200 * time.Millisecond)
			if m.seekTarget == target {
				_ = m.client.Seek(context.Background(), target)
			}
			return nil
		}

	case ActionCursorDown:
		m.moveCursor(1)
		return m, nil

	case ActionCursorUp:
		m.moveCursor(-1)
		return m, nil

	case ActionSelect:
		return m.handleEnter()
	}

	// Fallbacks for standard terminal navigation keys
	switch key {
	case "up":
		m.moveCursor(-1)
		return m, nil
	case "down":
		m.moveCursor(1)
		return m, nil
	case "enter":
		return m.handleEnter()
	case "esc":
		if m.zenMode || m.currentTab == TabLyrics {
			m.lyricsManualScroll = false
		}
		return m, nil
	}

	return m, nil
}

func (m *AppModel) filteredPlaylists() []backend.Playlist {
	switch m.playlistFilter {
	case FilterAll:
		return m.playlists
	case FilterSpotify:
		var res []backend.Playlist
		for _, pl := range m.playlists {
			if strings.EqualFold(pl.OwnerID, "spotify") {
				res = append(res, pl)
			}
		}
		return res
	case FilterByYou:
		var res []backend.Playlist
		for _, pl := range m.playlists {
			isOwn := (m.userID != "" && pl.OwnerID == m.userID) || (m.username != "" && strings.EqualFold(pl.OwnerID, m.username))
			if isOwn {
				res = append(res, pl)
			}
		}
		return res
	case FilterAlbums:
		return m.albums
	case FilterArtists:
		return m.artists
	default:
		return m.playlists
	}
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
	if m.zenMode {
		m.lyricsManualScroll = true
		m.lyricsPointerMovedAt = time.Now()
		newIdx := m.lyricsCursor + delta
		if newIdx >= 0 && newIdx < len(m.lyricsLines) {
			m.lyricsCursor = newIdx
		}
		return
	}

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
			totalItems := len(m.playlistTracks) + len(m.artistAlbums) + len(m.searchArtists)
			newIdx := m.centerIndex + delta
			if newIdx >= 0 && newIdx < totalItems {
				m.centerIndex = newIdx
			}
		case TabHistory:
			newIdx := m.centerIndex + delta
			if newIdx >= 0 && newIdx < len(m.history) {
				m.centerIndex = newIdx
			}
		case TabLyrics:
			m.lyricsManualScroll = true
			m.lyricsPointerMovedAt = time.Now()
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
	if m.zenMode {
		if m.lyricsCursor >= 0 && m.lyricsCursor < len(m.lyricsLines) {
			timeMs := m.lyricsLines[m.lyricsCursor].TimeMs
			m.lyricsManualScroll = false
			if m.playback != nil {
				m.playback.ProgressMs = timeMs
				m.client.SaveLastState(m.playback)
			}
			return m, func() tea.Msg {
				_ = m.client.Seek(context.Background(), timeMs)
				return nil
			}
		}
		return m, nil
	}

	switch m.focused {
	case PaneNav:
		pls := m.filteredPlaylists()
		if m.navIndex >= 0 && m.navIndex < len(pls) {
			pl := pls[m.navIndex]
			m.navHistory = nil
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
			} else if m.centerIndex >= len(m.playlistTracks) && m.centerIndex < len(m.playlistTracks)+len(m.artistAlbums) {
				albumIdx := m.centerIndex - len(m.playlistTracks)
				album := m.artistAlbums[albumIdx]

				currID := ""
				if parts := strings.Split(m.currentPlURI, ":"); len(parts) >= 3 {
					currID = parts[2]
				}
				m.navHistory = append(m.navHistory, containerHistoryItem{
					ID:          currID,
					URI:         m.currentPlURI,
					Name:        m.currentPlName,
					CenterIndex: m.centerIndex,
					Tracks:      m.playlistTracks,
					Albums:      m.artistAlbums,
					Artists:     m.searchArtists,
				})

				m.currentPlURI = album.URI
				m.currentPlName = album.Name
				m.centerIndex = 0
				m.searchArtists = nil
				return m, m.fetchPlaylistTracksCmd(album.ID, album.URI, album.Name)
			} else if m.centerIndex >= len(m.playlistTracks)+len(m.artistAlbums) && m.centerIndex < len(m.playlistTracks)+len(m.artistAlbums)+len(m.searchArtists) {
				artIdx := m.centerIndex - len(m.playlistTracks) - len(m.artistAlbums)
				artist := m.searchArtists[artIdx]

				currID := ""
				if parts := strings.Split(m.currentPlURI, ":"); len(parts) >= 3 {
					currID = parts[2]
				}
				m.navHistory = append(m.navHistory, containerHistoryItem{
					ID:          currID,
					URI:         m.currentPlURI,
					Name:        m.currentPlName,
					CenterIndex: m.centerIndex,
					Tracks:      m.playlistTracks,
					Albums:      m.artistAlbums,
					Artists:     m.searchArtists,
				})

				m.currentPlURI = artist.URI
				m.currentPlName = artist.Name
				m.centerIndex = 0
				m.searchArtists = nil
				return m, m.fetchPlaylistTracksCmd(artist.ID, artist.URI, artist.Name)
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
					m.client.SaveLastState(m.playback)
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
		ZenView:          m.zenView,
		ShowHelp:         m.showHelp,
		HelpIndex:        m.helpIndex,
		HelpEditing:      m.helpEditing,
		KeybindItems:     m.getKeyManager().Items,
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
		ArtistAlbums:     m.artistAlbums,
		SearchArtists:    m.searchArtists,
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
	v.BackgroundColor = CurrentTheme.Surface
	return v
}

func (m *AppModel) GetPlaybackState() *backend.PlaybackState {
	return m.playback
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

type AlbumsMsg []backend.Playlist
type ArtistsMsg []backend.Playlist

type openArtistResolvedMsg struct {
	Artist backend.Playlist
}

func (m *AppModel) fetchAlbumsCmd() tea.Cmd {
	return func() tea.Msg {
		albums, err := m.client.GetAlbums(context.Background())
		if err != nil {
			return AlbumsMsg(nil)
		}
		return AlbumsMsg(albums)
	}
}

func (m *AppModel) fetchArtistsCmd() tea.Cmd {
	return func() tea.Msg {
		artists, err := m.client.GetArtists(context.Background())
		if err != nil {
			return ArtistsMsg(nil)
		}
		return ArtistsMsg(artists)
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
		if plID == "" && strings.Contains(plURI, ":") {
			parts := strings.Split(plURI, ":")
			if len(parts) >= 3 {
				plID = parts[2]
			}
		}

		if strings.HasPrefix(plURI, "spotify:artist:") {
			tracks, err := m.client.GetArtistTracks(context.Background(), plID)
			if err != nil {
				return ErrorMsg(err)
			}
			albums, _ := m.client.GetArtistAlbums(context.Background(), plID)
			return TracksMsg{
				PlaylistURI:  plURI,
				PlaylistName: plName,
				Tracks:       tracks,
				Albums:       albums,
			}
		}

		tracks, err := m.client.GetContainerTracks(context.Background(), plID, plURI)
		if err != nil {
			return ErrorMsg(err)
		}
		return TracksMsg{
			PlaylistURI:  plURI,
			PlaylistName: plName,
			Tracks:       tracks,
			Albums:       nil,
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
		ansiStr, diskPath, _ := m.artRen.Render(url, w, h)
		return ArtMsg{
			ANSI:     ansiStr,
			DiskPath: diskPath,
			Width:    w,
			Height:   h,
			Zen:      zen,
		}
	}
}

func (m *AppModel) getZenArtGeometry() (row, col, w, h int) {
	if m.zenView == ZenViewLyrics {
		return 0, 0, 0, 0
	}

	innerW := m.width - 2
	innerH := m.height - 2
	maxH := innerH - 8
	if maxH < 4 {
		return 0, 0, 0, 0
	}

	if m.zenView == ZenViewBoth {
		gapW := 2
		leftW := (innerW - gapW) * 48 / 100
		if leftW < 30 {
			leftW = (innerW - gapW) / 2
		}
		if leftW > innerW-25 {
			leftW = innerW - 25
		}
		if leftW < 10 {
			leftW = innerW / 2
		}
		maxW := leftW - 4
		if maxW < 8 {
			return 0, 0, 0, 0
		}
		targetW := maxH * 2
		if targetW > maxW {
			targetW = maxW
		}
		if targetW%2 != 0 {
			targetW--
		}
		if targetW > 56 {
			targetW = 56
		}
		targetH := targetW / 2
		if targetH > maxH {
			targetH = maxH
		}

		artStackH := targetH + 7
		topPad := (innerH - artStackH) / 2
		if topPad < 0 {
			topPad = 0
		}
		row = 2 + topPad
		col = 2 + (leftW-targetW)/2
		return row, col, targetW, targetH
	}

	// ZenViewArt
	maxW := innerW - 8
	if maxW < 8 {
		return 0, 0, 0, 0
	}
	targetW := maxH * 2
	if targetW > maxW {
		targetW = maxW
	}
	if targetW%2 != 0 {
		targetW--
	}
	if targetW > 64 {
		targetW = 64
	}
	targetH := targetW / 2
	if targetH > maxH {
		targetH = maxH
	}

	artStackH := targetH + 7
	topPad := (innerH - artStackH) / 2
	if topPad < 0 {
		topPad = 0
	}
	row = 2 + topPad
	col = (m.width-targetW)/2 + 1
	return row, col, targetW, targetH
}

func (m *AppModel) searchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		query = strings.TrimSpace(query)
		if query == "" {
			return SearchResultsMsg{}
		}
		tracks, albums, artists, _ := m.client.Search(context.Background(), query)
		return SearchResultsMsg{
			Tracks:  tracks,
			Albums:  albums,
			Artists: artists,
		}
	}
}

func (m *AppModel) defaultDeviceCmd() tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < 6; i++ {
			if err := m.client.DefaultToSpotumnDevice(context.Background()); err == nil {
				st, _ := m.client.GetPlaybackState(context.Background())
				if st != nil && strings.EqualFold(st.DeviceName, "spotumn") {
					return PlaybackMsg(st)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
		return nil
	}
}
