package backend

import (
	"context"
	"encoding/json"
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
}

type Client struct {
	spClient *spotify.Client
	mu       sync.RWMutex
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
	for _, a := range t.Artists {
		artists = append(artists, a.Name)
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
		return ps, nil
	}

	// Fallback to currently playing if full state returned empty/204
	if cp, cpErr := c.spClient.PlayerCurrentlyPlaying(ctx); cpErr == nil && cp != nil && cp.Item != nil {
		track := extractTrack(cp.Item)
		return &PlaybackState{
			Playing:      cp.Playing,
			ProgressMs:   int(cp.Progress),
			DurationMs:   int(cp.Item.Duration),
			Volume:       50,
			CurrentTrack: &track,
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

	// Deduplicate repeated consecutive tracks
	lastURI := ""
	if res.Current != nil {
		lastURI = res.Current.URI
	}

	for _, item := range q.Items {
		if item.ID == "" {
			continue
		}
		t := extractTrack(&item)
		// Filter out duplicate consecutive tracks (e.g. current song repeated in queue)
		if t.URI == lastURI {
			continue
		}
		lastURI = t.URI
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
	for _, a := range t.Artists {
		artists = append(artists, a.Name)
	}
	return Track{
		ID:         string(t.ID),
		URI:        string(t.URI),
		Name:       t.Name,
		Artist:     strings.Join(artists, ", "),
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

// Search searches for tracks and playlists matching the query
func (c *Client) Search(ctx context.Context, query string) ([]Track, []Playlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil, nil
	}

	res, err := c.spClient.Search(ctx, query, spotify.SearchTypeTrack|spotify.SearchTypePlaylist, spotify.Limit(25))
	if err != nil {
		return nil, nil, err
	}

	var tracks []Track
	if res.Tracks != nil {
		for _, t := range res.Tracks.Tracks {
			tracks = append(tracks, extractTrack(&t))
		}
	}

	var playlists []Playlist
	if res.Playlists != nil {
		for _, p := range res.Playlists.Playlists {
			img := ""
			if len(p.Images) > 0 {
				img = p.Images[0].URL
			}
			playlists = append(playlists, Playlist{
				ID:         string(p.ID),
				URI:        string(p.URI),
				Name:       p.Name,
				TrackCount: int(p.Tracks.Total),
				ImageURL:   img,
			})
		}
	}

	return tracks, playlists, nil
}

// PlayTrackList starts playback of a track naturally without wiping queue
func (c *Client) PlayTrackList(ctx context.Context, tracks []Track, startIndex int, contextURI string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	opts := &spotify.PlayOptions{}

	// If no device is currently active, route to spotumn
	if !c.hasActiveDevice(ctx) {
		if devID := c.findSpotumnDeviceID(ctx); devID != "" {
			opts.DeviceID = &devID
		}
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

	opts := &spotify.PlayOptions{}
	if !c.hasActiveDevice(ctx) {
		if devID := c.findSpotumnDeviceID(ctx); devID != "" {
			opts.DeviceID = &devID
		}
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

// GetDevices returns all available Spotify Connect devices
func (c *Client) GetDevices(ctx context.Context) ([]spotify.PlayerDevice, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.PlayerDevices(ctx)
}

// TransferPlayback transfers active playback to the given device ID
func (c *Client) TransferPlayback(ctx context.Context, deviceID spotify.ID) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.TransferPlayback(ctx, deviceID, true)
}

// ToggleDevice toggles playback between spotumn and another available device
func (c *Client) ToggleDevice(ctx context.Context) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	devices, err := c.spClient.PlayerDevices(ctx)
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
		_ = c.spClient.TransferPlayback(ctx, otherDev.ID, true)
		return otherDev.Name, nil
	}

	if spotumnDev != nil {
		_ = c.spClient.TransferPlayback(ctx, spotumnDev.ID, true)
		return "spotumn", nil
	}

	return "", nil
}

func (c *Client) Play(ctx context.Context) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.spClient.Play(ctx)
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

// SaveLastState persists playback state to disk
func (c *Client) SaveLastState(ps *PlaybackState) {
	if ps == nil || ps.CurrentTrack == nil {
		return
	}
	data, err := json.Marshal(ps)
	if err != nil {
		return
	}
	path := filepath.Join(config.GetDir(), "last_state.json")
	_ = os.WriteFile(path, data, 0600)
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
	ps.Playing = false
	return &ps
}
