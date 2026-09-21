package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"spotumn/internal/config"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type Track struct {
	ID         string
	URI        string
	Name       string
	Artist     string
	ArtistID   string
	Album      string
	DurationMs int
	ArtURL     string
}

type Playlist struct {
	ID         string
	URI        string
	Name       string
	OwnerID    string
	TrackCount int
	ImageURL   string
}

type QueueData struct {
	Current *Track
	Items   []Track
}

type PlaybackState struct {
	Playing      bool
	ProgressMs   int
	DurationMs   int
	Volume       int
	Shuffle      bool
	Repeat       string // "off", "context", "track"
	DeviceName   string
	DeviceID     string
	CurrentTrack *Track
	ContextURI   string
}

type Client struct {
	spClient       *spotify.Client
	mu             sync.RWMutex
	lastState      *PlaybackState
	explicitRemote bool
}

func NewClient(ctx context.Context, ts oauth2.TokenSource) *Client {
	httpClient := oauth2.NewClient(ctx, ts)
	sp := spotify.New(httpClient)
	return &Client{
		spClient: sp,
	}
}

func extractTrack(t *spotify.FullTrack) Track {
	if t == nil {
		return Track{}
	}

	var artists []string
	artistID := ""
	for i, a := range t.Artists {
		artists = append(artists, a.Name)
		if i == 0 {
			artistID = string(a.ID)
		}
	}

	artURL := ""
	if len(t.Album.Images) > 0 {
		artURL = t.Album.Images[0].URL
	}

	return Track{
		ID:         string(t.ID),
		URI:        string(t.URI),
		Name:       t.Name,
		Artist:     strings.Join(artists, ", "),
		ArtistID:   artistID,
		Album:      t.Album.Name,
		DurationMs: int(t.Duration),
		ArtURL:     artURL,
	}
}

// GetPlaybackState fetches the current player state with fallback
func (c *Client) GetPlaybackState(ctx context.Context) (*PlaybackState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, err := c.spClient.PlayerState(ctx)
	if err == nil && state != nil {
		ps := &PlaybackState{
			Playing:    state.Playing,
			ProgressMs: int(state.Progress),
			Volume:     int(state.Device.Volume),
			Shuffle:    state.ShuffleState,
			Repeat:     state.RepeatState,
			DeviceName: state.Device.Name,
			DeviceID:   string(state.Device.ID),
		}
		if state.Item != nil {
			track := extractTrack(state.Item)
			ps.CurrentTrack = &track
			ps.DurationMs = int(state.Item.Duration)
		}
		if state.PlaybackContext.URI != "" {
			ps.ContextURI = string(state.PlaybackContext.URI)
		}
		return ps, nil
	}

	// Fallback to currently playing if full state returned empty/204
	if cp, cpErr := c.spClient.PlayerCurrentlyPlaying(ctx); cpErr == nil && cp != nil && cp.Item != nil {
		track := extractTrack(cp.Item)
		ctxURI := ""
		if cp.PlaybackContext.URI != "" {
			ctxURI = string(cp.PlaybackContext.URI)
		}
		return &PlaybackState{
			Playing:      cp.Playing,
			ProgressMs:   int(cp.Progress),
			DurationMs:   int(cp.Item.Duration),
			Volume:       50,
			CurrentTrack: &track,
			ContextURI:   ctxURI,
		}, nil
	}

	return &PlaybackState{Volume: 50}, nil
}

// GetQueue fetches the real-time playback queue
func (c *Client) GetQueue(ctx context.Context) (*QueueData, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := &QueueData{}
	q, err := c.spClient.GetQueue(ctx)
	if err != nil || q == nil {
		return res, nil
	}

	if q.CurrentlyPlaying.ID != "" {
		cur := extractTrack(&q.CurrentlyPlaying)
		res.Current = &cur
	}

	for _, item := range q.Items {
		if item.ID == "" {
			continue
		}
		t := extractTrack(&item)
		res.Items = append(res.Items, t)
	}

	return res, nil
}

// GetRecommendationsOrTopTracks fetches user's top/saved tracks or popular tracks
func (c *Client) GetRecommendationsOrTopTracks(ctx context.Context) ([]Track, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var tracks []Track

	// 1. Try user's top tracks
	top, err := c.spClient.CurrentUsersTopTracks(ctx, spotify.Limit(30))
	if err == nil && top != nil && len(top.Tracks) > 0 {
		for _, t := range top.Tracks {
			tracks = append(tracks, extractTrack(&t))
		}
		return tracks, nil
	}

	// 2. Try user's saved tracks (Liked Songs)
	saved, err := c.spClient.CurrentUsersTracks(ctx, spotify.Limit(30))
	if err == nil && saved != nil && len(saved.Tracks) > 0 {
		for _, st := range saved.Tracks {
			tracks = append(tracks, extractTrack(&st.FullTrack))
		}
		return tracks, nil
	}

	// 3. Fallback to search popular tracks so the user always has music to listen to
	searchRes, err := c.spClient.Search(ctx, "top hits", spotify.SearchTypeTrack, spotify.Limit(25))
	if err == nil && searchRes.Tracks != nil {
		for _, t := range searchRes.Tracks.Tracks {
			tracks = append(tracks, extractTrack(&t))
		}
	}

	return tracks, nil
}

func extractSimpleTrack(t *spotify.SimpleTrack) Track {
	if t == nil {
		return Track{}
	}
	var artists []string
	artistID := ""
	for i, a := range t.Artists {
		artists = append(artists, a.Name)
		if i == 0 {
			artistID = string(a.ID)
		}
	}
	return Track{
		ID:         string(t.ID),
		URI:        string(t.URI),
		Name:       t.Name,
		Artist:     strings.Join(artists, ", "),
		ArtistID:   artistID,
		DurationMs: int(t.Duration),
	}
}

func extractTrackID(uriOrID string) string {
	if strings.HasPrefix(uriOrID, "spotify:track:") {
		return strings.TrimPrefix(uriOrID, "spotify:track:")
	}
	return uriOrID
}

// GetAllPlaylists paginates through all user playlists
func (c *Client) GetAllPlaylists(ctx context.Context) ([]Playlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var playlists []Playlist
	limit := 50
	offset := 0

	for {
		page, err := c.spClient.CurrentUsersPlaylists(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		for _, p := range page.Playlists {
			img := ""
			if len(p.Images) > 0 {
				img = p.Images[0].URL
			}
			playlists = append(playlists, Playlist{
				ID:         string(p.ID),
				URI:        string(p.URI),
				Name:       p.Name,
				OwnerID:    string(p.Owner.ID),
				TrackCount: int(p.Tracks.Total),
				ImageURL:   img,
			})
		}

		offset += len(page.Playlists)
		if offset >= int(page.Total) || len(page.Playlists) == 0 {
			break
		}
	}

	return playlists, nil
}

// GetPlaylistTracks fetches all tracks from a playlist with fallback
func (c *Client) GetPlaylistTracks(ctx context.Context, playlistID string) ([]Track, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var tracks []Track
	limit := 100
	offset := 0

	for {
		page, err := c.spClient.GetPlaylistItems(ctx, spotify.ID(playlistID), spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			break
		}

		for _, item := range page.Items {
			if item.Track.Track != nil && item.Track.Track.ID != "" {
				tracks = append(tracks, extractTrack(item.Track.Track))
			}
		}

		offset += len(page.Items)
		if offset >= int(page.Total) || len(page.Items) == 0 {
			break
		}
	}

	// Fallback to GetPlaylist if items query returned empty
	if len(tracks) == 0 {
		if pl, err := c.spClient.GetPlaylist(ctx, spotify.ID(playlistID)); err == nil && pl != nil {
			for _, item := range pl.Tracks.Tracks {
				if item.Track.ID != "" {
					tracks = append(tracks, extractTrack(&item.Track))
				}
			}
		}
	}

	return tracks, nil
}

// GetAlbums fetches the user's saved albums from Spotify
func (c *Client) GetAlbums(ctx context.Context) ([]Playlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var albums []Playlist
	limit := 50
	offset := 0

	for {
		page, err := c.spClient.CurrentUsersAlbums(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			break
		}

		for _, a := range page.Albums {
			img := ""
			if len(a.Images) > 0 {
				img = a.Images[0].URL
			}
			artistName := ""
			if len(a.Artists) > 0 {
				artistName = a.Artists[0].Name
			}
			albums = append(albums, Playlist{
				ID:         string(a.ID),
				URI:        string(a.URI),
				Name:       a.Name,
				OwnerID:    artistName,
				TrackCount: int(a.Tracks.Total),
				ImageURL:   img,
			})
		}

		offset += len(page.Albums)
		if offset >= int(page.Total) || len(page.Albums) == 0 {
			break
		}
	}

	return albums, nil
}

// GetArtists fetches user's followed artists, falling back to top artists
func (c *Client) GetArtists(ctx context.Context) ([]Playlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var artists []Playlist

	// 1. Try followed artists first
	cursor, err := c.spClient.CurrentUsersFollowedArtists(ctx, spotify.Limit(50))
	if err == nil && cursor != nil && len(cursor.Artists) > 0 {
		for _, a := range cursor.Artists {
			img := ""
			if len(a.Images) > 0 {
				img = a.Images[0].URL
			}
			artists = append(artists, Playlist{
				ID:         string(a.ID),
				URI:        string(a.URI),
				Name:       a.Name,
				OwnerID:    "Artist",
				TrackCount: int(a.Popularity),
				ImageURL:   img,
			})
		}
		return artists, nil
	}

	// 2. Fallback to top artists
	top, err := c.spClient.CurrentUsersTopArtists(ctx, spotify.Limit(50))
	if err == nil && top != nil {
		for _, a := range top.Artists {
			img := ""
			if len(a.Images) > 0 {
				img = a.Images[0].URL
			}
			artists = append(artists, Playlist{
				ID:         string(a.ID),
				URI:        string(a.URI),
				Name:       a.Name,
				OwnerID:    "Artist",
				TrackCount: int(a.Popularity),
				ImageURL:   img,
			})
		}
	}

	return artists, nil
}

// GetAlbumTracks fetches all tracks belonging to an album
func (c *Client) GetAlbumTracks(ctx context.Context, albumID string) ([]Track, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	album, err := c.spClient.GetAlbum(ctx, spotify.ID(albumID))
	if err == nil && album != nil {
		artURL := ""
		if len(album.Images) > 0 {
			artURL = album.Images[0].URL
		}
		var tracks []Track
		for _, t := range album.Tracks.Tracks {
			artistNames := make([]string, len(t.Artists))
			artistID := ""
			for i, a := range t.Artists {
				artistNames[i] = a.Name
				if i == 0 {
					artistID = string(a.ID)
				}
			}
			tracks = append(tracks, Track{
				ID:         string(t.ID),
				URI:        string(t.URI),
				Name:       t.Name,
				Artist:     strings.Join(artistNames, ", "),
				ArtistID:   artistID,
				Album:      album.Name,
				DurationMs: int(t.Duration),
				ArtURL:     artURL,
			})
		}
		return tracks, nil
	}

	page, err := c.spClient.GetAlbumTracks(ctx, spotify.ID(albumID), spotify.Limit(50))
	if err != nil {
		return nil, err
	}
	var tracks []Track
	for _, t := range page.Tracks {
		artistNames := make([]string, len(t.Artists))
		artistID := ""
		for i, a := range t.Artists {
			artistNames[i] = a.Name
			if i == 0 {
				artistID = string(a.ID)
			}
		}
		tracks = append(tracks, Track{
			ID:         string(t.ID),
			URI:        string(t.URI),
			Name:       t.Name,
			Artist:     strings.Join(artistNames, ", "),
			ArtistID:   artistID,
			DurationMs: int(t.Duration),
		})
	}
	return tracks, nil
}

// GetArtistTracks fetches top tracks for an artist
func (c *Client) GetArtistTracks(ctx context.Context, artistID string) ([]Track, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fts, err := c.spClient.GetArtistsTopTracks(ctx, spotify.ID(artistID), "from_token")
	if err != nil || len(fts) == 0 {
		fts, err = c.spClient.GetArtistsTopTracks(ctx, spotify.ID(artistID), "US")
	}
	if err != nil {
		return nil, err
	}
	var tracks []Track
	for _, t := range fts {
		tracks = append(tracks, extractTrack(&t))
	}
	return tracks, nil
}

// GetArtistAlbums fetches an artist's albums and singles
func (c *Client) GetArtistAlbums(ctx context.Context, artistID string) ([]Playlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.spClient == nil {
		return nil, nil
	}

	types := []spotify.AlbumType{spotify.AlbumTypeAlbum, spotify.AlbumTypeSingle}
	page, err := c.spClient.GetArtistAlbums(ctx, spotify.ID(artistID), types, spotify.Limit(50), spotify.Market("from_token"))
	if err != nil || page == nil || len(page.Albums) == 0 {
		page, err = c.spClient.GetArtistAlbums(ctx, spotify.ID(artistID), types, spotify.Limit(50), spotify.Market("US"))
	}
	if err != nil || page == nil {
		return nil, err
	}

	var rawAlbums []spotify.SimpleAlbum
	rawAlbums = append(rawAlbums, page.Albums...)

	// Fetch up to 2 additional pages (up to 150 albums max for responsiveness)
	for pageCount := 0; pageCount < 2; pageCount++ {
		err := c.spClient.NextPage(ctx, page)
		if err != nil || len(page.Albums) == 0 {
			break
		}
		rawAlbums = append(rawAlbums, page.Albums...)
	}

	var albums []Playlist
	seen := make(map[string]bool)
	for _, a := range rawAlbums {
		cleanName := strings.ToLower(strings.TrimSpace(a.Name))
		if seen[cleanName] {
			continue
		}
		seen[cleanName] = true

		img := ""
		if len(a.Images) > 0 {
			img = a.Images[0].URL
		}
		year := a.ReleaseDate
		if len(year) > 4 {
			year = year[:4]
		}
		typeLabel := "Album"
		grp := a.AlbumGroup
		if grp == "" {
			grp = a.AlbumType
		}
		if grp != "" {
			typeLabel = strings.ToUpper(grp[:1]) + strings.ToLower(grp[1:])
		}
		desc := typeLabel
		if year != "" {
			desc = fmt.Sprintf("%s • %s", typeLabel, year)
		}

		albums = append(albums, Playlist{
			ID:         string(a.ID),
			URI:        string(a.URI),
			Name:       a.Name,
			OwnerID:    desc,
			TrackCount: int(a.TotalTracks),
			ImageURL:   img,
		})
	}
	return albums, nil
}

// GetContainerTracks dynamically routes to album, artist, or playlist tracks depending on URI
func (c *Client) GetContainerTracks(ctx context.Context, id, uri string) ([]Track, error) {
	if strings.HasPrefix(uri, "spotify:album:") {
		return c.GetAlbumTracks(ctx, id)
	}
	if strings.HasPrefix(uri, "spotify:artist:") {
		return c.GetArtistTracks(ctx, id)
	}
	return c.GetPlaylistTracks(ctx, id)
}

// GetArtistByName searches for an artist by name and returns their profile if found
func (c *Client) GetArtistByName(ctx context.Context, name string) (*Playlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.spClient == nil || strings.TrimSpace(name) == "" {
		return nil, nil
	}
	res, err := c.spClient.Search(ctx, name, spotify.SearchTypeArtist, spotify.Limit(1))
	if err != nil || res == nil || res.Artists == nil || len(res.Artists.Artists) == 0 {
		return nil, err
	}
	art := res.Artists.Artists[0]
	imgURL := ""
	if len(art.Images) > 0 {
		imgURL = art.Images[0].URL
	}
	return &Playlist{
		ID:         string(art.ID),
		URI:        string(art.URI),
		Name:       art.Name,
		ImageURL:   imgURL,
		TrackCount: int(art.Popularity),
	}, nil
}

// Search searches for tracks, albums, and artists matching the query
func (c *Client) Search(ctx context.Context, query string) (tracks []Track, albums []Playlist, artists []Playlist, err error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil, nil, nil
	}

	res, err := c.spClient.Search(ctx, query, spotify.SearchTypeTrack|spotify.SearchTypeAlbum|spotify.SearchTypeArtist, spotify.Limit(50))
	if err != nil {
		return nil, nil, nil, err
	}

	if res.Tracks != nil {
		for _, t := range res.Tracks.Tracks {
			tracks = append(tracks, extractTrack(&t))
		}
	}

	if res.Albums != nil {
		for _, a := range res.Albums.Albums {
			img := ""
			if len(a.Images) > 0 {
				img = a.Images[0].URL
			}
			year := a.ReleaseDate
			if len(year) > 4 {
				year = year[:4]
			}
			typeLabel := "Album"
			if a.AlbumType != "" {
				typeLabel = strings.ToUpper(a.AlbumType[:1]) + strings.ToLower(a.AlbumType[1:])
			}
			desc := typeLabel
			if year != "" {
				desc = fmt.Sprintf("%s • %s", typeLabel, year)
			}
			albums = append(albums, Playlist{
				ID:         string(a.ID),
				URI:        string(a.URI),
				Name:       a.Name,
				OwnerID:    desc,
				TrackCount: int(a.TotalTracks),
				ImageURL:   img,
			})
		}
	}

	if res.Artists != nil {
		for _, a := range res.Artists.Artists {
			img := ""
			if len(a.Images) > 0 {
				img = a.Images[0].URL
			}
			artists = append(artists, Playlist{
				ID:         string(a.ID),
				URI:        string(a.URI),
				Name:       a.Name,
				OwnerID:    "Artist",
				TrackCount: int(a.Popularity),
				ImageURL:   img,
			})
		}
	}

	return tracks, albums, artists, nil
}

// PlayTrackList starts playback of a track naturally without wiping queue
func (c *Client) PlayTrackList(ctx context.Context, tracks []Track, startIndex int, contextURI string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := &spotify.PlayOptions{
		DeviceID: c.getTargetDeviceID(ctx),
	}

	if contextURI != "" {
		cURI := spotify.URI(contextURI)
		opts.PlaybackContext = &cURI
		if startIndex >= 0 && startIndex < len(tracks) {
			tURI := spotify.URI(tracks[startIndex].URI)
			opts.PlaybackOffset = &spotify.PlaybackOffset{URI: tURI}
		}
	} else if len(tracks) > 0 && startIndex >= 0 && startIndex < len(tracks) {
		// Natural autoplay: pass selected track and upcoming tracks from search/history
		// so Spotify populates the player's upcoming queue naturally
		var uris []spotify.URI
		endIndex := len(tracks)
		if endIndex-startIndex > 50 {
			endIndex = startIndex + 50
		}
		for i := startIndex; i < endIndex; i++ {
			uris = append(uris, spotify.URI(tracks[i].URI))
		}
		opts.URIs = uris
	}

	return c.spClient.PlayOpt(ctx, opts)
}

// PlayTrack starts playback of a single track with optional context
func (c *Client) PlayTrack(ctx context.Context, trackURI, contextURI string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := &spotify.PlayOptions{
		DeviceID: c.getTargetDeviceID(ctx),
	}

	if contextURI != "" {
		cURI := spotify.URI(contextURI)
		opts.PlaybackContext = &cURI
		if trackURI != "" {
			tURI := spotify.URI(trackURI)
			opts.PlaybackOffset = &spotify.PlaybackOffset{URI: tURI}
		}
	} else if trackURI != "" {
		opts.URIs = []spotify.URI{spotify.URI(trackURI)}
	}

	return c.spClient.PlayOpt(ctx, opts)
}

func (c *Client) hasActiveDevice(ctx context.Context) bool {
	devices, err := c.spClient.PlayerDevices(ctx)
	if err != nil {
		return false
	}
	for _, d := range devices {
		if d.Active {
			return true
		}
	}
	return false
}

func (c *Client) findSpotumnDeviceID(ctx context.Context) spotify.ID {
	devices, err := c.spClient.PlayerDevices(ctx)
	if err != nil {
		return ""
	}
	for _, d := range devices {
		if strings.EqualFold(d.Name, "spotumn") {
			return d.ID
		}
	}
	return ""
}

func (c *Client) getTargetDeviceID(ctx context.Context) *spotify.ID {
	c.mu.RLock()
	remote := c.explicitRemote
	c.mu.RUnlock()

	spotumnID := c.findSpotumnDeviceID(ctx)

	// Default to spotumn for audio playback unless the user explicitly selected a remote device
	if !remote && spotumnID != "" {
		return &spotumnID
	}

	// Remote control mode: let active remote device receive playback
	if c.hasActiveDevice(ctx) {
		return nil
	}

	// Fallback to spotumn or first available device
	if spotumnID != "" {
		return &spotumnID
	}
	devices, err := c.spClient.PlayerDevices(ctx)
	if err == nil && len(devices) > 0 {
		return &devices[0].ID
	}
	return nil
}

// sortDevicesWithSpotumnFirst ensures spotumn is always the topmost device (index 0)
func sortDevicesWithSpotumnFirst(devices []spotify.PlayerDevice) []spotify.PlayerDevice {
	if len(devices) <= 1 {
		return devices
	}
	for i, d := range devices {
		if strings.EqualFold(d.Name, "spotumn") {
			if i > 0 {
				spotumnDev := devices[i]
				copy(devices[1:i+1], devices[0:i])
				devices[0] = spotumnDev
			}
			break
		}
	}
	return devices
}

// GetDevices returns all available Spotify Connect devices with spotumn sorted to the topmost position (index 0)
func (c *Client) GetDevices(ctx context.Context) ([]spotify.PlayerDevice, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	devices, err := c.spClient.PlayerDevices(ctx)
	if err != nil {
		return nil, err
	}
	return sortDevicesWithSpotumnFirst(devices), nil
}

// TransferPlayback transfers active playback to the given device ID
func (c *Client) TransferPlayback(ctx context.Context, deviceID spotify.ID) error {
	spotumnID := c.findSpotumnDeviceID(ctx)
	c.mu.Lock()
	if spotumnID != "" && deviceID == spotumnID {
		c.explicitRemote = false
	} else {
		c.explicitRemote = true
	}
	c.mu.Unlock()

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.TransferPlayback(ctx, deviceID, true)
}

// ToggleDevice toggles playback between spotumn and another available device
func (c *Client) ToggleDevice(ctx context.Context) (string, error) {
	c.mu.RLock()
	devices, err := c.spClient.PlayerDevices(ctx)
	c.mu.RUnlock()
	if err != nil || len(devices) == 0 {
		return "", err
	}

	var spotumnDev *spotify.PlayerDevice
	var otherDev *spotify.PlayerDevice
	for i := range devices {
		if strings.EqualFold(devices[i].Name, "spotumn") {
			spotumnDev = &devices[i]
		} else if otherDev == nil {
			otherDev = &devices[i]
		}
	}

	if spotumnDev != nil && spotumnDev.Active && otherDev != nil {
		_ = c.TransferPlayback(ctx, otherDev.ID)
		return otherDev.Name, nil
	}

	if spotumnDev != nil {
		_ = c.TransferPlayback(ctx, spotumnDev.ID)
		return "spotumn", nil
	}

	return "", nil
}

func (c *Client) Play(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	opts := &spotify.PlayOptions{
		DeviceID: c.getTargetDeviceID(ctx),
	}
	return c.spClient.PlayOpt(ctx, opts)
}

func (c *Client) Pause(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.Pause(ctx)
}

func (c *Client) PlayPause(ctx context.Context, currentlyPlaying bool) error {
	if currentlyPlaying {
		return c.Pause(ctx)
	}
	return c.Play(ctx)
}

func (c *Client) Next(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.Next(ctx)
}

func (c *Client) Previous(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.Previous(ctx)
}

func (c *Client) Seek(ctx context.Context, positionMs int) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.Seek(ctx, positionMs)
}

func (c *Client) SetVolume(ctx context.Context, percent int) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return c.spClient.Volume(ctx, percent)
}

func (c *Client) ToggleShuffle(ctx context.Context, current bool) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.Shuffle(ctx, !current)
}

func (c *Client) CycleRepeat(ctx context.Context, current string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var next string
	switch current {
	case "off":
		next = "context"
	case "context":
		next = "track"
	default:
		next = "off"
	}
	return c.spClient.Repeat(ctx, next)
}

// EnsureActiveDevice looks for an active device, or transfers to spotumn/available device
func (c *Client) EnsureActiveDevice(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	devices, err := c.spClient.PlayerDevices(ctx)
	if err != nil {
		return err
	}

	for _, d := range devices {
		if d.Active {
			return nil // An active device is already playing or ready
		}
	}

	// Try to find spotumn device first
	for _, d := range devices {
		if strings.EqualFold(d.Name, "spotumn") {
			return c.spClient.TransferPlayback(ctx, d.ID, false)
		}
	}

	// Otherwise pick the first available device
	if len(devices) > 0 {
		return c.spClient.TransferPlayback(ctx, devices[0].ID, false)
	}

	return nil
}

// DefaultToSpotumnDevice ensures spotumn is activated as the playback device on startup
func (c *Client) DefaultToSpotumnDevice(ctx context.Context) error {
	c.mu.Lock()
	c.explicitRemote = false
	c.mu.Unlock()

	spotumnID := c.findSpotumnDeviceID(ctx)
	if spotumnID != "" {
		c.mu.RLock()
		defer c.mu.RUnlock()
		return c.spClient.TransferPlayback(ctx, spotumnID, false)
	}
	return nil
}

// GetCurrentUser returns the user's Spotify display name and user ID
func (c *Client) GetCurrentUser(ctx context.Context) (displayName, userID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	user, err := c.spClient.CurrentUser(ctx)
	if err != nil || user == nil {
		return "spotumn", ""
	}
	name := user.DisplayName
	if name == "" {
		name = user.ID
	}
	return name, user.ID
}

// GetTasteRecommendations fetches algorithmic recommendations based on listening trends
func (c *Client) GetTasteRecommendations(ctx context.Context) ([]Track, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var seedIDs []spotify.ID
	top, err := c.spClient.CurrentUsersTopTracks(ctx, spotify.Limit(5))
	if err == nil && top != nil {
		for _, t := range top.Tracks {
			seedIDs = append(seedIDs, t.ID)
		}
	}

	if len(seedIDs) > 0 {
		seeds := spotify.Seeds{Tracks: seedIDs}
		recs, err := c.spClient.GetRecommendations(ctx, seeds, nil, spotify.Limit(24))
		if err == nil && recs != nil && len(recs.Tracks) > 0 {
			var tracks []Track
			for _, st := range recs.Tracks {
				tracks = append(tracks, extractSimpleTrack(&st))
			}
			return tracks, nil
		}
	}

	return c.GetRecommendationsOrTopTracks(ctx)
}

// GetRecentlyPlayed fetches recently played tracks from Spotify
func (c *Client) GetRecentlyPlayed(ctx context.Context) ([]Track, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	recent, err := c.spClient.PlayerRecentlyPlayed(ctx)
	if err != nil || recent == nil {
		return nil, err
	}

	var tracks []Track
	seen := make(map[string]bool)
	for _, item := range recent {
		if item.Track.ID != "" && !seen[string(item.Track.ID)] {
			seen[string(item.Track.ID)] = true
			tracks = append(tracks, extractSimpleTrack(&item.Track))
			if len(tracks) >= 20 {
				break
			}
		}
	}
	return tracks, nil
}

// QueueSong adds a song directly to Spotify's queue
func (c *Client) QueueSong(ctx context.Context, trackURI string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	id := extractTrackID(trackURI)
	if id == "" {
		return nil
	}
	return c.spClient.QueueSong(ctx, spotify.ID(id))
}

// PlayTrackAtPosition starts playback of a track at a specific timestamp (positionMs)
func (c *Client) PlayTrackAtPosition(ctx context.Context, trackURI, contextURI string, positionMs int) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := &spotify.PlayOptions{
		PositionMs: spotify.Numeric(positionMs),
		DeviceID:   c.getTargetDeviceID(ctx),
	}

	if contextURI != "" && !strings.Contains(contextURI, "collection") {
		cURI := spotify.URI(contextURI)
		opts.PlaybackContext = &cURI
		if trackURI != "" {
			tURI := spotify.URI(trackURI)
			opts.PlaybackOffset = &spotify.PlaybackOffset{URI: tURI}
		}
	} else if trackURI != "" {
		opts.URIs = []spotify.URI{spotify.URI(trackURI)}
	}

	err := c.spClient.PlayOpt(ctx, opts)
	if err == nil && positionMs > 0 {
		_ = c.spClient.Seek(ctx, positionMs)
	}
	return err
}

// SaveLastState persists playback state to disk and maintains an in-memory copy
func (c *Client) SaveLastState(ps *PlaybackState) {
	if ps == nil || ps.CurrentTrack == nil {
		return
	}
	c.mu.Lock()
	cp := *ps
	if ps.CurrentTrack != nil {
		ct := *ps.CurrentTrack
		cp.CurrentTrack = &ct
	}
	c.lastState = &cp
	c.mu.Unlock()

	data, err := json.Marshal(&cp)
	if err != nil {
		return
	}
	path := filepath.Join(config.GetDir(), "last_state.json")
	_ = os.WriteFile(path, data, 0600)
}

// GetLastSavedState returns the in-memory last saved state
func (c *Client) GetLastSavedState() *PlaybackState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastState
}

// LoadLastState retrieves the last known playback state from disk
func (c *Client) LoadLastState() *PlaybackState {
	path := filepath.Join(config.GetDir(), "last_state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var ps PlaybackState
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil
	}
	// Always default device back to spotumn on startup
	ps.DeviceName = "spotumn"
	ps.DeviceID = ""
	c.mu.Lock()
	c.lastState = &ps
	c.mu.Unlock()
	return &ps
}
