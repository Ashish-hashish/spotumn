package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/zmb3/spotify/v2"
	"spotumn/internal/backend"
	"spotumn/internal/lyrics"
)

type ViewParams struct {
	Width            int
	Height           int
	Focused          FocusedPane
	CurrentTab       CenterTab
	ShowLeftSidebar  bool
	ShowRightSidebar bool
	ZenMode          bool
	ShowHelp         bool
	ShowDevices      bool
	DeviceScanning   bool
	Devices          []spotify.PlayerDevice
	DeviceIndex      int
	Username         string
	NavIndex         int
	CenterIndex      int
	QueueIndex       int
	LyricsCursor     int
	SearchFocused    bool
	SearchQuery      string
	Playlists        []backend.Playlist
	PinnedURIs       map[string]bool
	PlaylistFilter   PlaylistFilter
	PlaylistTracks   []backend.Track
	PlaylistName     string
	History          []backend.Track
	Playback         *backend.PlaybackState
	Queue            []backend.Track
	LyricsLines      []lyrics.Line
	ArtANSI          string
	ZenArtANSI       string
}

// RenderFullUI renders the application running purely on terminal colors with highlighted focused borders
func RenderFullUI(p ViewParams) string {
	if p.Width < 48 || p.Height < 16 {
		msg := "Terminal window too small for spotumn. Please enlarge window."
		return lipgloss.NewStyle().
			Width(p.Width).
			Height(p.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(msg)
	}

	// Modal Help overlay
	if p.ShowHelp {
		return renderHelpOverlay(p)
	}

	// Modal Devices overlay
	if p.ShowDevices {
		return renderDevicesOverlay(p)
	}

	// Zen Mode view
	if p.ZenMode {
		return renderZenMode(p)
	}

	innerW := p.Width - 2
	innerH := p.Height - 2
	playerH := 4
	bodyH := innerH - playerH
	if bodyH < 4 {
		bodyH = 4
	}

	curURI := ""
	curProgress := 0
	var curTrack *backend.Track
	if p.Playback != nil {
		curProgress = p.Playback.ProgressMs
		curTrack = p.Playback.CurrentTrack
		if p.Playback.CurrentTrack != nil {
			curURI = p.Playback.CurrentTrack.URI
		}
	}

	var bodyContent string

	if p.ShowLeftSidebar && p.ShowRightSidebar {
		navW := innerW * 28 / 100
		if navW < 30 {
			navW = 30
		}
		if navW > 45 {
			navW = 45
		}

		rightW := innerW * 34 / 100
		if rightW < 34 {
			rightW = 34
		}
		if rightW > 50 {
			rightW = 50
		}

		centerW := innerW - navW - rightW
		if centerW < 20 {
			centerW = 20
		}

		navView := RenderNav(p.Playlists, p.PinnedURIs, p.PlaylistFilter, p.NavIndex, p.Focused == PaneNav, navW, bodyH)
		centerView := RenderCenter(
			p.CurrentTab,
			p.PlaylistTracks,
			p.History,
			p.PlaylistName,
			curURI,
			p.LyricsLines,
			p.LyricsCursor,
			curProgress,
			p.SearchQuery,
			p.SearchFocused,
			p.CenterIndex,
			p.Focused == PaneCenter,
			centerW,
			bodyH,
		)
		rightView := RenderMergedRight(p.ArtANSI, curTrack, p.Queue, p.QueueIndex, p.Focused == PaneRight, rightW, bodyH)

		bodyContent = lipgloss.JoinHorizontal(lipgloss.Top, navView, centerView, rightView)

	} else if p.ShowLeftSidebar && !p.ShowRightSidebar {
		navW := innerW * 30 / 100
		if navW < 30 {
			navW = 30
		}
		if navW > 45 {
			navW = 45
		}
		centerW := innerW - navW

		navView := RenderNav(p.Playlists, p.PinnedURIs, p.PlaylistFilter, p.NavIndex, p.Focused == PaneNav, navW, bodyH)
		centerView := RenderCenter(
			p.CurrentTab,
			p.PlaylistTracks,
			p.History,
			p.PlaylistName,
			curURI,
			p.LyricsLines,
			p.LyricsCursor,
			curProgress,
			p.SearchQuery,
			p.SearchFocused,
			p.CenterIndex,
			p.Focused == PaneCenter,
			centerW,
			bodyH,
		)

		bodyContent = lipgloss.JoinHorizontal(lipgloss.Top, navView, centerView)

	} else if !p.ShowLeftSidebar && p.ShowRightSidebar {
		rightW := innerW * 36 / 100
		if rightW < 34 {
			rightW = 34
		}
		centerW := innerW - rightW

		centerView := RenderCenter(
			p.CurrentTab,
			p.PlaylistTracks,
			p.History,
			p.PlaylistName,
			curURI,
			p.LyricsLines,
			p.LyricsCursor,
			curProgress,
			p.SearchQuery,
			p.SearchFocused,
			p.CenterIndex,
			p.Focused == PaneCenter,
			centerW,
			bodyH,
		)
		rightView := RenderMergedRight(p.ArtANSI, curTrack, p.Queue, p.QueueIndex, p.Focused == PaneRight, rightW, bodyH)

		bodyContent = lipgloss.JoinHorizontal(lipgloss.Top, centerView, rightView)

	} else {
		bodyContent = RenderCenter(
			p.CurrentTab,
			p.PlaylistTracks,
			p.History,
			p.PlaylistName,
			curURI,
			p.LyricsLines,
			p.LyricsCursor,
			curProgress,
			p.SearchQuery,
			p.SearchFocused,
			p.CenterIndex,
			p.Focused == PaneCenter,
			innerW,
			bodyH,
		)
	}

	playerView := RenderPlayer(p.Playback, p.Focused == PanePlayer, innerW)
	innerCombined := lipgloss.JoinVertical(lipgloss.Left, bodyContent, playerView)

	// Construct overall outer border with green dot before spotumn and person icon before username
	userName := p.Username
	if userName == "" {
		userName = "BrightestAutumn"
	}

	greenDot := lipgloss.NewStyle().Foreground(CurrentTheme.Green).Render("● ")
	titleTag := StyleFaint.Render("─ ") + greenDot + StylePurple.Render("spotumn ")
	personIcon := StyleLavender.Render("👤 ")
	userTag := " " + personIcon + StyleLavender.Render(userName) + StyleFaint.Render(" ─")

	middleDashesLen := innerW - ansi.StringWidth(titleTag) - ansi.StringWidth(userTag)
	if middleDashesLen < 1 {
		middleDashesLen = 1
	}

	cornerTL := StyleFaint.Render("╭")
	cornerTR := StyleFaint.Render("╮")
	cornerBL := StyleFaint.Render("╰")
	cornerBR := StyleFaint.Render("╯")
	sideBar := StyleFaint.Render("│")

	topBorder := cornerTL + titleTag + StyleFaint.Render(strings.Repeat("─", middleDashesLen)) + userTag + cornerTR

	// Format bottom border with keybind shortcuts embedded inside the frame
	rawShortcuts := " [Space] Play ─ [n] Next ─ [p] Prev ─ [[]/[]] Focus ─ [/] Search ─ [q] Queue ─ [?] Help "
	if innerW < 85 {
		rawShortcuts = " [Space] Play ─ [n/p] Prev/Next ─ [[]/[]] Focus ─ [/] Search ─ [?] Help "
	}
	if innerW < 65 {
		rawShortcuts = " [Space] Play ─ [[]/[]] Focus ─ [/] Search ─ [?] Help "
	}
	shortcutsWidth := ansi.StringWidth(rawShortcuts)
	var bottomBorder string
	if innerW > shortcutsWidth+4 {
		remDashes := innerW - shortcutsWidth
		leftD := remDashes / 2
		rightD := remDashes - leftD
		bottomBorder = cornerBL +
			StyleFaint.Render(strings.Repeat("─", leftD)) +
			StyleLavender.Render(rawShortcuts) +
			StyleFaint.Render(strings.Repeat("─", rightD)) +
			cornerBR
	} else {
		bottomBorder = cornerBL + StyleFaint.Render(strings.Repeat("─", innerW)) + cornerBR
	}

	var finalLines []string
	finalLines = append(finalLines, PadToWidth(topBorder, p.Width))

	rawInnerLines := strings.Split(innerCombined, "\n")
	for i := 0; i < innerH; i++ {
		row := ""
		if i < len(rawInnerLines) {
			row = rawInnerLines[i]
		}
		paddedRow := PadToWidth(row, innerW)
		borderedRow := sideBar + paddedRow + sideBar
		finalLines = append(finalLines, PadToWidth(borderedRow, p.Width))
	}

	finalLines = append(finalLines, PadToWidth(bottomBorder, p.Width))

	for len(finalLines) < p.Height {
		finalLines = append(finalLines, PadToWidth("", p.Width))
	}
	if len(finalLines) > p.Height {
		finalLines = finalLines[:p.Height]
	}

	return strings.Join(finalLines, "\n")
}

func renderHelpOverlay(p ViewParams) string {
	modal := RenderKeybindsModal(p.Width, p.Height)

	placed := lipgloss.Place(p.Width, p.Height, lipgloss.Center, lipgloss.Center, modal)

	var lines []string
	for _, l := range strings.Split(placed, "\n") {
		lines = append(lines, PadToWidth(l, p.Width))
	}
	for len(lines) < p.Height {
		lines = append(lines, PadToWidth("", p.Width))
	}
	return strings.Join(lines[:p.Height], "\n")
}

func renderDevicesOverlay(p ViewParams) string {
	modal := RenderDevicesModal(p.Devices, p.DeviceIndex, p.DeviceScanning, p.Width, p.Height)

	placed := lipgloss.Place(p.Width, p.Height, lipgloss.Center, lipgloss.Center, modal)

	var lines []string
	for _, l := range strings.Split(placed, "\n") {
		lines = append(lines, PadToWidth(l, p.Width))
	}
	for len(lines) < p.Height {
		lines = append(lines, PadToWidth("", p.Width))
	}
	return strings.Join(lines[:p.Height], "\n")
}

func renderZenMode(p ViewParams) string {
	innerW := p.Width - 2
	innerH := p.Height - 2

	var lines []string

	lines = append(lines, PadToWidth("", innerW))
	zenHeader := StylePurple.Render("  [ ZEN MODE - Press 'z' to Exit ]")
	lines = append(lines, PadToWidth(zenHeader, innerW))
	lines = append(lines, PadToWidth("", innerW))

	// Centered Album Art (enlarged)
	artANSI := p.ZenArtANSI
	if artANSI == "" {
		artANSI = p.ArtANSI
	}
	if artANSI != "" {
		artRows := strings.Split(artANSI, "\n")
		for _, row := range artRows {
			rowW := ansi.StringWidth(row)
			padLeft := (innerW - rowW) / 2
			if padLeft < 0 {
				padLeft = 0
			}
			centeredRow := strings.Repeat(" ", padLeft) + row
			lines = append(lines, PadToWidth(centeredRow, innerW))
		}
	}

	lines = append(lines, PadToWidth("", innerW))

	// Track and Artist
	if p.Playback != nil && p.Playback.CurrentTrack != nil {
		t := p.Playback.CurrentTrack
		tStyled := StyleBold.Render(t.Name)
		aStyled := StyleLavender.Render(t.Artist)

		lines = append(lines, PadToWidth(lipgloss.PlaceHorizontal(innerW, lipgloss.Center, tStyled), innerW))
		lines = append(lines, PadToWidth(lipgloss.PlaceHorizontal(innerW, lipgloss.Center, aStyled), innerW))
	}

	lines = append(lines, PadToWidth("", innerW))

	// Prominent Active Lyric
	if len(p.LyricsLines) > 0 && p.Playback != nil {
		activeIdx := lyrics.FindActiveIndex(p.LyricsLines, p.Playback.ProgressMs)
		if activeIdx >= 0 && activeIdx < len(p.LyricsLines) {
			activeText := p.LyricsLines[activeIdx].Text
			if activeText == "" {
				activeText = "♪ ♪ ♪"
			}
			bigLyric := StyleMint.Render("❯  " + activeText + "  ❮")
			lines = append(lines, PadToWidth(lipgloss.PlaceHorizontal(innerW, lipgloss.Center, bigLyric), innerW))
		}
	}

	lines = append(lines, PadToWidth("", innerW))

	// Seekbar & Controls in Zen Mode
	if p.Playback != nil {
		elapsed := FormatDuration(p.Playback.ProgressMs)
		total := FormatDuration(p.Playback.DurationMs)
		ratio := 0.0
		if p.Playback.DurationMs > 0 {
			ratio = float64(p.Playback.ProgressMs) / float64(p.Playback.DurationMs)
		}
		seek := renderProgressBar(ratio, 40)
		playIcon := "▶"
		if p.Playback.Playing {
			playIcon = "❚❚"
		}
		ctrls := fmt.Sprintf("%s   [ %s %s %s ]   %s", elapsed, "⏮", playIcon, "⏭", total)
		cStyled := StylePurple.Render(ctrls)
		lines = append(lines, PadToWidth(lipgloss.PlaceHorizontal(innerW, lipgloss.Center, cStyled), innerW))
		lines = append(lines, PadToWidth(lipgloss.PlaceHorizontal(innerW, lipgloss.Center, seek), innerW))
	}

	for len(lines) < innerH {
		lines = append(lines, PadToWidth("", innerW))
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Purple).
		Width(p.Width).
		Height(p.Height)

	rendered := boxStyle.Render(strings.Join(lines[:innerH], "\n"))
	var finalRows []string
	for _, l := range strings.Split(rendered, "\n") {
		finalRows = append(finalRows, PadToWidth(l, p.Width))
	}
	return strings.Join(finalRows, "\n")
}
