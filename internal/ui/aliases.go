// UI type aliases and forwarding helpers for clean package boundaries across components.
package ui

import (
	"time"

	"spotumn/internal/backend"
	"spotumn/internal/media/lyrics"
	"spotumn/internal/ui/modals"
	"spotumn/internal/ui/panes"
	"spotumn/internal/ui/state"
	"spotumn/internal/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type FocusedPane = state.FocusedPane

const (
	PaneNav    = state.PaneNav
	PaneCenter = state.PaneCenter
	PaneRight  = state.PaneRight
	PanePlayer = state.PanePlayer
)

type CenterTab = state.CenterTab

const (
	TabTracks  = state.TabTracks
	TabLyrics  = state.TabLyrics
	TabHistory = state.TabHistory
)

type ZenViewMode = state.ZenViewMode

const (
	ZenViewBoth       = state.ZenViewBoth
	ZenViewArtOnly    = state.ZenViewArtOnly
	ZenViewLyricsOnly = state.ZenViewLyricsOnly
	ZenViewArt        = state.ZenViewArtOnly
	ZenViewLyrics     = state.ZenViewLyricsOnly
)

type PlaylistFilter = state.PlaylistFilter

const (
	FilterAll     = state.FilterAll
	FilterSpotify = state.FilterSpotify
	FilterByYou   = state.FilterByYou
	FilterAlbums  = state.FilterAlbums
	FilterArtists = state.FilterArtists
)

type KeybindItem = state.KeybindItem
type KeybindCategory = state.KeybindCategory
type SettingsState = state.SettingsState
type ViewParams = state.ViewParams
type ContainerHistoryItem = state.ContainerHistoryItem
type containerHistoryItem = state.ContainerHistoryItem

type TickMsg = state.TickMsg
type UserMsg = state.UserMsg
type PlaybackMsg = state.PlaybackMsg
type ErrorMsg = state.ErrorMsg
type DevicesMsg = state.DevicesMsg
type SearchResultsMsg = state.SearchResultsMsg
type QueueMsg = state.QueueMsg
type PlaylistsMsg = state.PlaylistsMsg
type TracksMsg = state.TracksMsg
type AlbumsMsg = state.AlbumsMsg
type ArtistsMsg = state.ArtistsMsg
type HistoryMsg = state.HistoryMsg
type LyricsMsg = state.LyricsMsg
type ArtMsg = state.ArtMsg
type OpenArtistResolvedMsg = state.OpenArtistResolvedMsg
type openArtistResolvedMsg = state.OpenArtistResolvedMsg

type accountActionMsg struct {
	addedIndex  int
	displayName string
	userID      string
	token       *oauth2.Token
	err         error
}

func BgPad(n int) string {
	return theme.BgPad(n)
}

func CenterLine(s string, width int) string {
	return theme.CenterLine(s, width)
}

func CenterOverlay(modal string, width, height int) string {
	return theme.CenterOverlay(modal, width, height)
}

func PadToWidth(s string, width int) string {
	return theme.PadToWidth(s, width)
}

func TruncateString(s string, maxW int) string {
	return theme.TruncateString(s, maxW)
}

func PadPlain(s string, width int) string {
	return theme.PadPlain(s, width)
}

func PanelBox(focused bool, width, height int) lipgloss.Style {
	return theme.PanelBox(focused, width, height)
}

func FormatDuration(ms int) string {
	return theme.FormatDuration(ms)
}

func RenderPaddedLine(rawText string, style lipgloss.Style, width int) string {
	return theme.RenderPaddedLine(rawText, style, width)
}

func StringDisplayWidth(s string) int {
	return theme.StringDisplayWidth(s)
}

func TruncateVisualWidth(s string, maxW int, tail string) string {
	return theme.TruncateVisualWidth(s, maxW, tail)
}

func DeviceTypeGlyph(devType string) string {
	return theme.DeviceTypeGlyph(devType)
}

type ThemeConfig = theme.ThemeConfig
type ThemePalette = theme.ThemePalette

var CurrentTheme = theme.CurrentTheme
var StyleBase = theme.StyleBase

func CatppuccinMocha() theme.ThemeConfig {
	return theme.CatppuccinMocha()
}

func catppuccinMocha() theme.ThemeConfig {
	return theme.CatppuccinMocha()
}

func MergeConfig(base, custom theme.ThemeConfig) theme.ThemeConfig {
	return theme.MergeConfig(base, custom)
}

func mergeConfig(base, custom theme.ThemeConfig) theme.ThemeConfig {
	return theme.MergeConfig(base, custom)
}

type ThemeFile = theme.ThemeFile

func ValidateThemeFile(data []byte) (*theme.ThemeFile, error) {
	return theme.ValidateThemeFile(data)
}

const (
	CacheTargetAll   = state.CacheTargetAll
	CacheTargetArt   = state.CacheTargetArt
	CacheTargetAudio = state.CacheTargetAudio
	CacheTargetState = state.CacheTargetState
)

func ApplyTheme(themeName string, isDark bool) {
	theme.ApplyTheme(themeName, isDark)
	CurrentTheme = theme.CurrentTheme
	StyleBase = theme.StyleBase
}

func SetTheme(isDark bool) {
	theme.SetTheme(isDark)
	CurrentTheme = theme.CurrentTheme
	StyleBase = theme.StyleBase
}

func SetThemeFromConfig(cfg theme.ThemeConfig) {
	theme.SetThemeFromConfig(cfg)
	CurrentTheme = theme.CurrentTheme
	StyleBase = theme.StyleBase
}

func ListAvailableThemes() []string {
	return theme.ListAvailableThemes()
}

func RenderProgressBar(ratio float64, width int) string {
	return theme.RenderProgressBar(ratio, width)
}

func renderProgressBar(ratio float64, width int) string {
	return theme.RenderProgressBar(ratio, width)
}

const SettingItemCount = modals.SettingItemCount

func RenderSettingsModal(s SettingsState, width, height int) string {
	return modals.RenderSettingsModal(s, width, height)
}

func RenderKeybindsModal(items []KeybindItem, selectedIdx int, isEditing bool, width, height int) string {
	return modals.RenderKeybindsModal(items, selectedIdx, isEditing, width, height)
}

func RenderDevicesModal(devices []spotify.PlayerDevice, selectedIdx int, isScanning bool, width, height int) string {
	return modals.RenderDevicesModal(devices, selectedIdx, isScanning, width, height)
}

func NewSettingsState() SettingsState {
	return modals.NewSettingsState()
}

func ClearCacheTarget(target int) error {
	return modals.ClearCacheTarget(target)
}

func ClearCache() error {
	return modals.ClearCache()
}

func RenderNav(playlists []backend.Playlist, pinnedURIs map[string]bool, filter PlaylistFilter, selectedIndex int, focused bool, width, height int) string {
	return panes.RenderNav(playlists, pinnedURIs, filter, selectedIndex, focused, width, height)
}

func RenderNavLines(playlists []backend.Playlist, pinnedURIs map[string]bool, filter PlaylistFilter, selectedIndex int, focused bool, width, height int) []string {
	return panes.RenderNavLines(playlists, pinnedURIs, filter, selectedIndex, focused, width, height)
}

func RenderCenter(
	currentTab CenterTab,
	tracks []backend.Track,
	albums []backend.Playlist,
	artists []backend.Playlist,
	history []backend.Track,
	playlistName string,
	currentPlayingTrackURI string,
	lyricsLines []lyrics.Line,
	lyricsCursor int,
	progressMs int,
	searchQuery string,
	searchFocused bool,
	selectedIndex int,
	focused bool,
	width, height int,
	containerURI ...string,
) string {
	return panes.RenderCenter(
		currentTab,
		tracks,
		albums,
		artists,
		history,
		playlistName,
		currentPlayingTrackURI,
		lyricsLines,
		lyricsCursor,
		progressMs,
		searchQuery,
		searchFocused,
		selectedIndex,
		focused,
		width, height,
		containerURI...,
	)
}

func RenderCenterLines(
	currentTab CenterTab,
	tracks []backend.Track,
	albums []backend.Playlist,
	artists []backend.Playlist,
	history []backend.Track,
	playlistName string,
	currentPlayingTrackURI string,
	lyricsLines []lyrics.Line,
	lyricsCursor int,
	progressMs int,
	searchQuery string,
	searchFocused bool,
	selectedIndex int,
	focused bool,
	width, height int,
	containerURI ...string,
) []string {
	return panes.RenderCenterLines(
		currentTab,
		tracks,
		albums,
		artists,
		history,
		playlistName,
		currentPlayingTrackURI,
		lyricsLines,
		lyricsCursor,
		progressMs,
		searchQuery,
		searchFocused,
		selectedIndex,
		focused,
		width, height,
		containerURI...,
	)
}

func RenderMergedRight(
	artANSI string,
	currentTrack *backend.Track,
	queue []backend.Track,
	queueIndex int,
	focused bool,
	width, height int,
) string {
	return panes.RenderMergedRight(artANSI, currentTrack, queue, queueIndex, focused, width, height)
}

func RenderRightLines(
	artANSI string,
	currentTrack *backend.Track,
	queue []backend.Track,
	queueIndex int,
	focused bool,
	width, height int,
) []string {
	return panes.RenderRightLines(artANSI, currentTrack, queue, queueIndex, focused, width, height)
}

func RenderPlayer(state *backend.PlaybackState, focused bool, width int) string {
	return panes.RenderPlayer(state, focused, width)
}

func RenderPlayerLines(state *backend.PlaybackState, focused bool, width int) []string {
	return panes.RenderPlayerLines(state, focused, width)
}

func RenderMiniSlider(value, width int) string {
	return panes.RenderMiniSlider(value, width)
}

func RenderTabBar(currentTab CenterTab, width int) string {
	return panes.RenderTabBar(currentTab, width)
}

func renderTabBar(currentTab CenterTab, width int) string {
	return panes.RenderTabBar(currentTab, width)
}

func RenderProminentSearchBar(query string, focused bool, width int) []string {
	return panes.RenderProminentSearchBar(query, focused, width)
}

func renderProminentSearchBar(query string, focused bool, width int) []string {
	return panes.RenderProminentSearchBar(query, focused, width)
}

func RenderTrackRow(idx int, t backend.Track, isSelected bool, isPlaying bool, focused bool, titleW, artistW int, width int) string {
	return panes.RenderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, width)
}

func renderTrackRow(idx int, t backend.Track, isSelected bool, isPlaying bool, focused bool, titleW, artistW int, width int) string {
	return panes.RenderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, width)
}

func RenderAlbumRow(idx int, a backend.Playlist, isSelected bool, focused bool, titleW, descW, countW int, width int) string {
	return panes.RenderAlbumRow(idx, a, isSelected, focused, titleW, descW, countW, width)
}

func renderAlbumRow(idx int, a backend.Playlist, isSelected bool, focused bool, titleW, descW, countW int, width int) string {
	return panes.RenderAlbumRow(idx, a, isSelected, focused, titleW, descW, countW, width)
}

func RenderArtistRow(idx int, a backend.Playlist, isSelected bool, focused bool, nameW, typeW, popW int, width int) string {
	return panes.RenderArtistRow(idx, a, isSelected, focused, nameW, typeW, popW, width)
}

func FormatTotalDuration(totalMs int, trackCount int) string {
	return panes.FormatTotalDuration(totalMs, trackCount)
}

func renderArtistRow(idx int, a backend.Playlist, isSelected bool, focused bool, nameW, typeW, popW int, width int) string {
	return panes.RenderArtistRow(idx, a, isSelected, focused, nameW, typeW, popW, width)
}

func RenderTracks(tracks []backend.Track, albums []backend.Playlist, artists []backend.Playlist, playlistName string, currentPlayingURI string, selectedIndex int, focused bool, width, height int, containerURI ...string) []string {
	return panes.RenderTracks(tracks, albums, artists, playlistName, currentPlayingURI, selectedIndex, focused, width, height, containerURI...)
}

func renderTracks(tracks []backend.Track, albums []backend.Playlist, artists []backend.Playlist, playlistName string, currentPlayingURI string, selectedIndex int, focused bool, width, height int, containerURI ...string) []string {
	return panes.RenderTracks(tracks, albums, artists, playlistName, currentPlayingURI, selectedIndex, focused, width, height, containerURI...)
}

func RenderLyrics(lines []lyrics.Line, cursorLine int, progressMs int, focused bool, width, height int) []string {
	return panes.RenderLyrics(lines, cursorLine, progressMs, focused, width, height)
}

func renderLyrics(lines []lyrics.Line, cursorLine int, progressMs int, focused bool, width, height int) []string {
	return panes.RenderLyrics(lines, cursorLine, progressMs, focused, width, height)
}

func NewTestAppModel() *AppModel {
	return &AppModel{
		pinnedURIs:  make(map[string]bool),
		pinnedOrder: make([]string, 0),
	}
}

func (m *AppModel) FilteredPlaylists() []backend.Playlist {
	return m.filteredPlaylists()
}

func (m *AppModel) CycleFilter(delta int) tea.Cmd {
	return m.cycleFilter(delta)
}

func (m *AppModel) HandleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if kp, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleKeyPress(kp)
	}
	return m, nil
}

func (m *AppModel) HandleEnter() (tea.Model, tea.Cmd) {
	return m.handleEnter()
}

func (m *AppModel) HandleSettingsAction() (tea.Model, tea.Cmd) {
	return m.handleSettingsAction()
}

func (m *AppModel) HandleSettingsArrow(delta int) {
	m.handleSettingsArrow(delta)
}

func (m *AppModel) IsLocalActive() bool {
	return m.isLocalActive()
}

func (m *AppModel) SetUserID(v string) *AppModel { m.userID = v; return m }
func (m *AppModel) UserID() string               { return m.userID }

func (m *AppModel) SetUsername(v string) *AppModel { m.username = v; return m }
func (m *AppModel) Username() string               { return m.username }

func (m *AppModel) SetPlaylists(v []backend.Playlist) *AppModel { m.playlists = v; return m }
func (m *AppModel) Playlists() []backend.Playlist               { return m.playlists }

func (m *AppModel) SetAlbums(v []backend.Playlist) *AppModel { m.albums = v; return m }
func (m *AppModel) Albums() []backend.Playlist               { return m.albums }

func (m *AppModel) SetArtists(v []backend.Playlist) *AppModel { m.artists = v; return m }
func (m *AppModel) Artists() []backend.Playlist               { return m.artists }

func (m *AppModel) SetPlaylistFilter(v PlaylistFilter) *AppModel { m.playlistFilter = v; return m }
func (m *AppModel) PlaylistFilter() PlaylistFilter               { return m.playlistFilter }

func (m *AppModel) SetCurrentPlURI(v string) *AppModel { m.currentPlURI = v; return m }
func (m *AppModel) CurrentPlURI() string               { return m.currentPlURI }

func (m *AppModel) SetCurrentPlName(v string) *AppModel { m.currentPlName = v; return m }
func (m *AppModel) CurrentPlName() string               { return m.currentPlName }

func (m *AppModel) SetCenterIndex(v int) *AppModel { m.centerIndex = v; return m }
func (m *AppModel) CenterIndex() int               { return m.centerIndex }

func (m *AppModel) SetNavIndex(v int) *AppModel { m.navIndex = v; return m }
func (m *AppModel) NavIndex() int               { return m.navIndex }

func (m *AppModel) SetNavHistory(v []ContainerHistoryItem) *AppModel { m.navHistory = v; return m }
func (m *AppModel) NavHistory() []ContainerHistoryItem               { return m.navHistory }

func (m *AppModel) SetLyricsCursor(v int) *AppModel { m.lyricsCursor = v; return m }
func (m *AppModel) LyricsCursor() int               { return m.lyricsCursor }

func (m *AppModel) SetLyricsManualScroll(v bool) *AppModel { m.lyricsManualScroll = v; return m }
func (m *AppModel) LyricsManualScroll() bool               { return m.lyricsManualScroll }

func (m *AppModel) SetLyricsPointerMovedAt(v time.Time) *AppModel {
	m.lyricsPointerMovedAt = v
	return m
}
func (m *AppModel) LyricsPointerMovedAt() time.Time { return m.lyricsPointerMovedAt }

func (m *AppModel) SetZenMode(v bool) *AppModel { m.zenMode = v; return m }
func (m *AppModel) ZenMode() bool               { return m.zenMode }

func (m *AppModel) SetArtistAlbums(v []backend.Playlist) *AppModel { m.artistAlbums = v; return m }
func (m *AppModel) ArtistAlbums() []backend.Playlist               { return m.artistAlbums }

func (m *AppModel) SetSearchArtists(v []backend.Playlist) *AppModel { m.searchArtists = v; return m }
func (m *AppModel) SearchArtists() []backend.Playlist               { return m.searchArtists }

func (m *AppModel) SetPlaylistTracks(v []backend.Track) *AppModel { m.playlistTracks = v; return m }
func (m *AppModel) PlaylistTracks() []backend.Track               { return m.playlistTracks }

func (m *AppModel) SetSearchFocused(v bool) *AppModel { m.searchFocused = v; return m }
func (m *AppModel) SearchFocused() bool               { return m.searchFocused }

func (m *AppModel) SetSearchQuery(v string) *AppModel { m.searchQuery = v; return m }
func (m *AppModel) SearchQuery() string               { return m.searchQuery }

func (m *AppModel) SetPinnedOrder(v []string) *AppModel { m.pinnedOrder = v; return m }
func (m *AppModel) PinnedOrder() []string               { return m.pinnedOrder }

func (m *AppModel) SetPinnedURIs(v map[string]bool) *AppModel { m.pinnedURIs = v; return m }
func (m *AppModel) PinnedURIs() map[string]bool               { return m.pinnedURIs }

func (m *AppModel) SetSettingsState(v SettingsState) *AppModel { m.settingsState = v; return m }
func (m *AppModel) SettingsState() SettingsState               { return m.settingsState }

func (m *AppModel) SetShowSettings(v bool) *AppModel { m.showSettings = v; return m }
func (m *AppModel) ShowSettings() bool               { return m.showSettings }

func (m *AppModel) SetDevices(v []spotify.PlayerDevice) *AppModel { m.devices = v; return m }
func (m *AppModel) Devices() []spotify.PlayerDevice               { return m.devices }

func (m *AppModel) SetFocused(v FocusedPane) *AppModel { m.focused = v; return m }
func (m *AppModel) Focused() FocusedPane               { return m.focused }

func (m *AppModel) SetCurrentTab(v CenterTab) *AppModel { m.currentTab = v; return m }
func (m *AppModel) CurrentTab() CenterTab               { return m.currentTab }

func (m *AppModel) SetPlayback(v *backend.PlaybackState) *AppModel { m.playback = v; return m }
func (m *AppModel) Playback() *backend.PlaybackState               { return m.playback }

func (m *AppModel) SetLyricsLines(v []lyrics.Line) *AppModel { m.lyricsLines = v; return m }
func (m *AppModel) LyricsLines() []lyrics.Line               { return m.lyricsLines }

func (m *AppModel) SetLyricsSynced(v bool) *AppModel { m.lyricsSynced = v; return m }
func (m *AppModel) LyricsSynced() bool               { return m.lyricsSynced }

func (m *AppModel) SetDaemon(v *backend.Daemon) *AppModel { m.daemon = v; return m }
func (m *AppModel) Daemon() *backend.Daemon               { return m.daemon }

func (m *AppModel) SetClient(v *backend.Client) *AppModel { m.client = v; return m }
func (m *AppModel) Client() *backend.Client               { return m.client }

func (m *AppModel) SetKeyManager(v *KeyManager) *AppModel { m.keyManager = v; return m }
func (m *AppModel) KeyManager() *KeyManager               { return m.keyManager }

func (m *AppModel) SettingsStatePtr() *SettingsState       { return &m.settingsState }
func (m *AppModel) PinnedURIsPtr() *map[string]bool        { return &m.pinnedURIs }
func (m *AppModel) PinnedOrderPtr() *[]string              { return &m.pinnedOrder }
func (m *AppModel) NavHistoryPtr() *[]ContainerHistoryItem { return &m.navHistory }
func (m *AppModel) ArtistAlbumsPtr() *[]backend.Playlist   { return &m.artistAlbums }
func (m *AppModel) SearchArtistsPtr() *[]backend.Playlist  { return &m.searchArtists }
func (m *AppModel) PlaylistTracksPtr() *[]backend.Track    { return &m.playlistTracks }
