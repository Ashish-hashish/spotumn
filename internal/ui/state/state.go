// Shared UI state types, enums, navigation models, and view layout definitions.
package state

import (
	"time"

	"spotumn/internal/auth"
	"spotumn/internal/backend"
	"spotumn/internal/media/lyrics"

	"github.com/zmb3/spotify/v2"
)

type FocusedPane int

const (
	PaneNav FocusedPane = iota
	PaneCenter
	PaneRight
	PanePlayer
)

type CenterTab int

const (
	TabTracks CenterTab = iota
	TabLyrics
	TabHistory
)

func (t CenterTab) Title() string {
	switch t {
	case TabTracks:
		return "Tracks"
	case TabLyrics:
		return "Lyrics"
	case TabHistory:
		return "History"
	default:
		return ""
	}
}

type ZenViewMode int

const (
	ZenViewBoth ZenViewMode = iota
	ZenViewArtOnly
	ZenViewLyricsOnly
)

func (z ZenViewMode) String() string {
	switch z {
	case ZenViewArtOnly:
		return "Art Only"
	case ZenViewLyricsOnly:
		return "Lyrics Only"
	default:
		return "Art + Lyrics"
	}
}

type PlaylistFilter int

const (
	FilterAll PlaylistFilter = iota
	FilterSpotify
	FilterByYou
	FilterAlbums
	FilterArtists
)

func (f PlaylistFilter) String() string {
	switch f {
	case FilterSpotify:
		return "By Spotify"
	case FilterByYou:
		return "By You"
	case FilterAlbums:
		return "Albums"
	case FilterArtists:
		return "Artists"
	default:
		return "ALL"
	}
}

type KeybindItem struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Key         string   `json:"key"`
	Keys        []string `json:"keys"`
	Desc        string   `json:"desc"`
	DefaultKey  string   `json:"default_key"`
	DefaultKeys []string `json:"default_keys"`
	ReadOnly    bool     `json:"read_only,omitempty"`
}

type KeybindCategory struct {
	Title string
	Items []KeybindItem
}

const (
	CacheTargetAll   = 0
	CacheTargetArt   = 1
	CacheTargetAudio = 2
	CacheTargetState = 3
)

type SettingsState struct {
	Index         int
	Status        string
	Themes        []string
	CurrentTheme  string
	Mode          string
	AutoShrink    bool
	CrossfadeSec  int
	Bitrate       int
	Normalisation bool
	Autoplay      bool
	Accounts      []auth.Account
	ActiveAccIdx  int
	CacheTarget   int
	ConfirmLogout bool
}

type ContainerHistoryItem struct {
	ID          string
	URI         string
	Name        string
	CenterIndex int
	Tracks      []backend.Track
	Albums      []backend.Playlist
	Artists     []backend.Playlist
}

type ViewParams struct {
	Width              int
	Height             int
	Focused            FocusedPane
	CurrentTab         CenterTab
	ShowLeftSidebar    bool
	ShowRightSidebar   bool
	AutoShrinkSidebars bool
	ZenMode            bool
	ZenView            ZenViewMode
	ShowHelp           bool
	HelpIndex          int
	HelpEditing        bool
	KeybindItems       []KeybindItem
	ShowSettings       bool
	SettingsState      SettingsState
	ShowDevices        bool
	DeviceScanning     bool
	Devices            []spotify.PlayerDevice
	DeviceIndex        int
	Username           string
	NavIndex           int
	CenterIndex        int
	QueueIndex         int
	LyricsCursor       int
	SearchFocused      bool
	SearchQuery        string
	Playlists          []backend.Playlist
	PinnedURIs         map[string]bool
	PlaylistFilter     PlaylistFilter
	PlaylistTracks     []backend.Track
	ArtistAlbums       []backend.Playlist
	SearchArtists      []backend.Playlist
	PlaylistName       string
	CurrentPlURI       string
	History            []backend.Track
	Playback           *backend.PlaybackState
	Queue              []backend.Track
	LyricsLines        []lyrics.Line
	ArtANSI            string
	ZenArtANSI         string
}

type TickMsg time.Time
type UserMsg struct{ DisplayName, UserID string }
type PlaybackMsg *backend.PlaybackState
type ErrorMsg error
type DevicesMsg []spotify.PlayerDevice
type SearchResultsMsg struct {
	Tracks  []backend.Track
	Albums  []backend.Playlist
	Artists []backend.Playlist
}
type QueueMsg *backend.QueueData
type PlaylistsMsg []backend.Playlist
type TracksMsg struct {
	PlaylistURI  string
	PlaylistName string
	Tracks       []backend.Track
	Albums       []backend.Playlist
}
type AlbumsMsg []backend.Playlist
type ArtistsMsg []backend.Playlist
type HistoryMsg []backend.Track
type LyricsMsg struct {
	TrackURI string
	Lines    []lyrics.Line
	Synced   bool
	Duration int
}
type ArtMsg struct {
	ANSI     string
	DiskPath string
	IsZen    bool
}
type OpenArtistResolvedMsg struct {
	Artist backend.Playlist
}
