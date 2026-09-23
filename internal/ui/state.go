package ui

import (
	"time"

	"spotumn/internal/backend"
	"spotumn/internal/lyrics"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
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

type containerHistoryItem struct {
	ID          string
	URI         string
	Name        string
	CenterIndex int
	Tracks      []backend.Track
	Albums      []backend.Playlist
	Artists     []backend.Playlist
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
	Lines    []lyrics.Line
	Synced   bool
	Duration int
}
type ArtMsg struct {
	ANSI     string
	DiskPath string
	IsZen    bool
}
type openArtistResolvedMsg struct {
	Artist backend.Playlist
}
type accountActionMsg struct {
	addedIndex  int
	displayName string
	userID      string
	token       *oauth2.Token
	err         error
}
