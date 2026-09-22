package ui

import (
	"fmt"
	"strings"

	"spotumn/internal/backend"
	"spotumn/internal/lyrics"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/zmb3/spotify/v2"
)

type ZenViewMode int

const (
	ZenViewBoth ZenViewMode = iota // Art + Lyrics
	ZenViewArt                     // Art Only
	ZenViewLyrics                  // Lyrics Only
)

func (z ZenViewMode) String() string {
	switch z {
	case ZenViewArt:
		return "Art Only"
	case ZenViewLyrics:
		return "Lyrics Only"
	default:
		return "Art + Lyrics"
	}
}

type ViewParams struct {
	Width            int
	Height           int
	Focused          FocusedPane
	CurrentTab       CenterTab
	ShowLeftSidebar  bool
	ShowRightSidebar bool
	ZenMode          bool
	ZenView          ZenViewMode
	ShowHelp         bool
	HelpIndex        int
	HelpEditing      bool
	KeybindItems     []KeybindItem
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
	ArtistAlbums     []backend.Playlist
	SearchArtists    []backend.Playlist
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
			Background(CurrentTheme.Surface).
			Foreground(CurrentTheme.OnSurface).
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

	var navW, rightW int
	if p.ShowLeftSidebar && p.ShowRightSidebar {
		navW = innerW * 28 / 100
		if navW < 30 {
			navW = 30
		}
		if navW > 45 {
			navW = 45
		}
		rightW = innerW * 34 / 100
		if rightW < 34 {
			rightW = 34
		}
		if rightW > 50 {
			rightW = 50
		}
	} else if p.ShowLeftSidebar {
		navW = innerW * 30 / 100
		if navW < 30 {
			navW = 30
		}
		if navW > 45 {
			navW = 45
		}
	} else if p.ShowRightSidebar {
		rightW = innerW * 36 / 100
		if rightW < 34 {
			rightW = 34
		}
	}

	centerW := innerW - navW - rightW
	if centerW < 20 {
		centerW = 20
	}

	centerView := RenderCenter(
		p.CurrentTab,
		p.PlaylistTracks,
		p.ArtistAlbums,
		p.SearchArtists,
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

	var views []string
	var viewWidths []int
	if navW > 0 {
		views = append(views, RenderNav(p.Playlists, p.PinnedURIs, p.PlaylistFilter, p.NavIndex, p.Focused == PaneNav, navW, bodyH))
		viewWidths = append(viewWidths, navW)
	}
	views = append(views, centerView)
	viewWidths = append(viewWidths, centerW)
	if rightW > 0 {
		views = append(views, RenderMergedRight(p.ArtANSI, curTrack, p.Queue, p.QueueIndex, p.Focused == PaneRight, rightW, bodyH))
		viewWidths = append(viewWidths, rightW)
	}
	bodyContent := joinHorizontalThemed(views, viewWidths, bodyH)

	playerView := RenderPlayer(p.Playback, p.Focused == PanePlayer, innerW)
	innerCombined := bodyContent + "\n" + playerView

	// Construct overall outer border with green dot before spotumn and person icon before username
	userName := p.Username

	greenDot := lipgloss.NewStyle().Foreground(CurrentTheme.Success).Background(CurrentTheme.Surface).Render("● ")
	titleTag := StyleFaint.Render("─ ") + greenDot + StylePurple.Render("spotumn ")
	var userTag string
	if userName != "" {
		personIcon := StyleLavender.Render(" ")
		userTag = BgPad(1) + personIcon + StyleLavender.Render(userName) + StyleFaint.Render(" ─")
	}

	middleDashesLen := innerW - ansi.StringWidth(titleTag) - ansi.StringWidth(userTag)
	if middleDashesLen < 1 {
		middleDashesLen = 1
	}

	cornerTL := StyleFaint.Render("╭")
	cornerTR := StyleFaint.Render("╮")
	sideBar := StyleFaint.Render("│")

	topBorder := cornerTL + titleTag + StyleFaint.Render(strings.Repeat("─", middleDashesLen)) + userTag + cornerTR

	var items []keybindItem
	if innerW >= 85 {
		items = []keybindItem{
			{"Space", "Play"},
			{"n", "Next"},
			{"p", "Prev"},
			{"[", "Focus Left"},
			{"]", "Focus Right"},
			{"/", "Search"},
			{"q", "Queue"},
			{"?", "Help"},
		}
	} else if innerW >= 65 {
		items = []keybindItem{
			{"Space", "Play"},
			{"n/p", "Prev/Next"},
			{"[/]", "Focus"},
			{"/", "Search"},
			{"?", "Help"},
		}
	} else {
		items = []keybindItem{
			{"Space", "Play"},
			{"n/p", "Prev/Next"},
			{"[/]", "Focus"},
			{"?", "Help"},
		}
	}

	bottomBorder := renderKeybindBar(items, innerW)

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

type keybindItem struct {
	key   string
	label string
}

func renderKeybindBar(items []keybindItem, innerW int) string {
	var renderedParts []string
	var rawParts []string
	dashSep := StyleFaint.Render(" ─ ")
	for _, it := range items {
		pill := StylePurple.Render("◖") + StyleBold.Render(it.key) + StylePurple.Render("◗")
		renderedParts = append(renderedParts, pill+BgPad(1)+StyleLavender.Render(it.label))
		rawParts = append(rawParts, "◖"+it.key+"◗ "+it.label)
	}

	rawShortcuts := " " + strings.Join(rawParts, " ─ ") + " "
	shortcutsWidth := ansi.StringWidth(rawShortcuts)

	cornerBL := StyleFaint.Render("╰")
	cornerBR := StyleFaint.Render("╯")

	if innerW > shortcutsWidth+4 {
		remDashes := innerW - shortcutsWidth
		leftD := remDashes / 2
		rightD := remDashes - leftD
		return cornerBL +
			StyleFaint.Render(strings.Repeat("─", leftD)) +
			BgPad(1) + strings.Join(renderedParts, dashSep) + BgPad(1) +
			StyleFaint.Render(strings.Repeat("─", rightD)) +
			cornerBR
	}
	return cornerBL + StyleFaint.Render(strings.Repeat("─", innerW)) + cornerBR
}

func renderHelpOverlay(p ViewParams) string {
	modal := RenderKeybindsModal(p.KeybindItems, p.HelpIndex, p.HelpEditing, p.Width, p.Height)
	return CenterOverlay(modal, p.Width, p.Height)
}

func renderDevicesOverlay(p ViewParams) string {
	modal := RenderDevicesModal(p.Devices, p.DeviceIndex, p.DeviceScanning, p.Width, p.Height)
	return CenterOverlay(modal, p.Width, p.Height)
}

// renderZenLyrics renders centered, vertically-aligned synchronized lyrics for Zen Mode
func renderZenLyrics(lines []lyrics.Line, cursorLine int, progressMs int, width, height int) []string {
	var output []string
	if width < 10 {
		width = 10
	}
	if height < 1 {
		height = 1
	}

	if len(lines) == 0 {
		msg := StyleFaint.Render("No synced lyrics available")
		midRow := height / 2
		for i := 0; i < height; i++ {
			if i == midRow {
				output = append(output, CenterLine(msg, width))
			} else {
				output = append(output, PadToWidth("", width))
			}
		}
		return output
	}

	activeIdx := lyrics.FindActiveIndex(lines, progressMs)
	if activeIdx < 0 {
		activeIdx = 0
	}

	targetCenter := activeIdx
	if cursorLine >= 0 && cursorLine < len(lines) {
		targetCenter = cursorLine
	}

	midRow := height / 2

	for r := 0; r < height; r++ {
		idx := targetCenter - (midRow - r)
		if idx < 0 || idx >= len(lines) {
			output = append(output, PadToWidth("", width))
			continue
		}

		line := lines[idx]
		text := line.Text
		if text == "" {
			text = "♪"
		}

		isSinging := idx == activeIdx
		isCursor := idx == cursorLine

		var styledText string
		if isCursor && isSinging {
			styledText = lipgloss.NewStyle().Foreground(CurrentTheme.Tertiary).Background(CurrentTheme.Surface).Bold(true).Render("❯  " + text + "  ❮")
		} else if isCursor {
			styledText = lipgloss.NewStyle().Foreground(CurrentTheme.Primary).Background(CurrentTheme.Surface).Bold(true).Render("➜  " + text + "  ➜")
		} else if isSinging {
			styledText = lipgloss.NewStyle().Foreground(CurrentTheme.Tertiary).Background(CurrentTheme.Surface).Bold(true).Render("❯  " + text + "  ❮")
		} else {
			styledText = StyleFaint.Render(text)
		}

		output = append(output, CenterLine(styledText, width))
	}

	return output
}

func renderZenMode(p ViewParams) string {
	innerW := p.Width - 2
	innerH := p.Height - 2

	greenDot := lipgloss.NewStyle().Foreground(CurrentTheme.Success).Background(CurrentTheme.Surface).Render("● ")
	titleTag := StyleFaint.Render("─ ") + greenDot + StylePurple.Render("spotumn ") + StyleLavender.Render("◖Zen Mode◗")
	viewModeStr := p.ZenView.String()
	viewBadge := StylePurple.Render("◖") + StyleBold.Render(viewModeStr) + StylePurple.Render("◗")
	viewTag := BgPad(1) + viewBadge + StyleFaint.Render(" ─")

	rawTitle := "─ ● spotumn ◖Zen Mode◗"
	rawView := " ◖" + viewModeStr + "◗ ─"
	middleDashesLen := innerW - ansi.StringWidth(rawTitle) - ansi.StringWidth(rawView)
	if middleDashesLen < 1 {
		middleDashesLen = 1
	}

	cornerTL := StyleFaint.Render("╭")
	cornerTR := StyleFaint.Render("╮")
	sideBar := StyleFaint.Render("│")

	topBorder := cornerTL + titleTag + StyleFaint.Render(strings.Repeat("─", middleDashesLen)) + viewTag + cornerTR

	var items []keybindItem
	if innerW >= 70 {
		items = []keybindItem{
			{"Space", "Play"},
			{"n/p", "Prev/Next"},
			{"j/k", "Lyrics"},
			{"Enter", "Seek"},
			{"v", "View"},
			{"z", "Exit Zen"},
		}
	} else if innerW >= 50 {
		items = []keybindItem{
			{"Space", "Play"},
			{"j/k", "Lyrics"},
			{"v", "View"},
			{"z", "Exit"},
		}
	} else {
		items = []keybindItem{
			{"Space", "Play"},
			{"v", "View"},
			{"z", "Exit"},
		}
	}

	bottomBorder := renderKeybindBar(items, innerW)

	curProgress := 0
	durationMs := 0
	volume := 50
	playing := false
	trackName := "No Track Playing"
	artistName := "spotumn"
	if p.Playback != nil {
		curProgress = p.Playback.ProgressMs
		durationMs = p.Playback.DurationMs
		volume = p.Playback.Volume
		playing = p.Playback.Playing
		if p.Playback.CurrentTrack != nil {
			trackName = p.Playback.CurrentTrack.Name
			artistName = p.Playback.CurrentTrack.Artist
		}
	}
	elapsed := FormatDuration(curProgress)
	total := FormatDuration(durationMs)
	ratio := 0.0
	if durationMs > 0 {
		ratio = float64(curProgress) / float64(durationMs)
	}
	playIcon := "▶"
	if playing {
		playIcon = "❚❚"
	}

	volBar := RenderMiniSlider(volume, 8)
	vStyled := StyleFaint.Render("Vol: [") + volBar + StyleFaint.Render(fmt.Sprintf("] %2d%%", volume))

	artANSI := p.ZenArtANSI
	if artANSI == "" {
		artANSI = p.ArtANSI
	}

	var innerLines []string

	switch p.ZenView {
	case ZenViewBoth:
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
		rightW := innerW - leftW - gapW
		if rightW < 5 {
			rightW = 5
		}

		maxArtRows := innerH - 8
		if maxArtRows < 4 {
			maxArtRows = 4
		}

		var displayedArtRows []string
		if artANSI != "" {
			artRows := strings.Split(artANSI, "\n")
			if len(artRows) > maxArtRows {
				artRows = artRows[:maxArtRows]
			}
			for _, row := range artRows {
				displayedArtRows = append(displayedArtRows, CenterLine(row, leftW))
			}
		}

		// Y-axis vertical centering for artwork and controls stack
		artStackH := len(displayedArtRows) + 7
		topPad := (innerH - artStackH) / 2
		if topPad < 0 {
			topPad = 0
		}

		var leftLines []string
		for i := 0; i < topPad; i++ {
			leftLines = append(leftLines, PadToWidth("", leftW))
		}
		leftLines = append(leftLines, displayedArtRows...)

		leftLines = append(leftLines, PadToWidth("", leftW))
		tStyled := StyleBold.Render(TruncateString(trackName, leftW-4))
		leftLines = append(leftLines, CenterLine(tStyled, leftW))

		aStyled := StyleLavender.Render(TruncateString(artistName, leftW-4))
		leftLines = append(leftLines, CenterLine(aStyled, leftW))

		leftLines = append(leftLines, PadToWidth("", leftW))

		ctrls := fmt.Sprintf("%s   [ ⏮  %s  ⏭ ]   %s", elapsed, playIcon, total)
		if leftW < 36 {
			ctrls = fmt.Sprintf("%s  %s  %s", elapsed, playIcon, total)
		}
		cStyled := StylePurple.Render(ctrls)
		leftLines = append(leftLines, CenterLine(cStyled, leftW))

		barLen := leftW - 10
		if barLen > 36 {
			barLen = 36
		}
		if barLen < 10 {
			barLen = 10
		}
		seek := renderProgressBar(ratio, barLen)
		leftLines = append(leftLines, CenterLine(seek, leftW))
		leftLines = append(leftLines, CenterLine(vStyled, leftW))

		for len(leftLines) < innerH {
			leftLines = append(leftLines, PadToWidth("", leftW))
		}
		if len(leftLines) > innerH {
			leftLines = leftLines[:innerH]
		}

		// Centered lyrics on right side, vertically aligned in Y axis
		rightLines := renderZenLyrics(p.LyricsLines, p.LyricsCursor, curProgress, rightW, innerH)

		// Space separator instead of vertical dividing line
		divider := BgPad(gapW)
		for i := 0; i < innerH; i++ {
			innerLines = append(innerLines, leftLines[i]+divider+rightLines[i])
		}

	case ZenViewArt:
		maxArtRows := innerH - 8
		if maxArtRows < 4 {
			maxArtRows = 4
		}

		var displayedArtRows []string
		if artANSI != "" {
			artRows := strings.Split(artANSI, "\n")
			if len(artRows) > maxArtRows {
				artRows = artRows[:maxArtRows]
			}
			for _, row := range artRows {
				displayedArtRows = append(displayedArtRows, CenterLine(row, innerW))
			}
		}

		// Y-axis vertical centering for centered artwork and controls
		artStackH := len(displayedArtRows) + 7
		topPad := (innerH - artStackH) / 2
		if topPad < 0 {
			topPad = 0
		}

		for i := 0; i < topPad; i++ {
			innerLines = append(innerLines, PadToWidth("", innerW))
		}
		innerLines = append(innerLines, displayedArtRows...)

		innerLines = append(innerLines, PadToWidth("", innerW))
		tStyled := StyleBold.Render(TruncateString(trackName, innerW-4))
		innerLines = append(innerLines, CenterLine(tStyled, innerW))

		aStyled := StyleLavender.Render(TruncateString(artistName, innerW-4))
		innerLines = append(innerLines, CenterLine(aStyled, innerW))

		innerLines = append(innerLines, PadToWidth("", innerW))

		ctrls := fmt.Sprintf("%s   [ ⏮  %s  ⏭ ]   %s", elapsed, playIcon, total)
		cStyled := StylePurple.Render(ctrls)
		innerLines = append(innerLines, CenterLine(cStyled, innerW))

		barLen := 44
		if barLen > innerW-10 {
			barLen = innerW - 10
		}
		if barLen < 12 {
			barLen = 12
		}
		seek := renderProgressBar(ratio, barLen)
		innerLines = append(innerLines, CenterLine(seek, innerW))
		innerLines = append(innerLines, CenterLine(vStyled, innerW))

		for len(innerLines) < innerH {
			innerLines = append(innerLines, PadToWidth("", innerW))
		}
		if len(innerLines) > innerH {
			innerLines = innerLines[:innerH]
		}

	case ZenViewLyrics:
		// Top banner: track and artist
		banner := StyleBold.Render(TruncateString(trackName, innerW/2)) + StyleFaint.Render(" ─ ") + StyleLavender.Render(TruncateString(artistName, innerW/2))
		topRow := CenterLine(banner, innerW)

		// Bottom controls, seekbar, and volume (UNDER the lyrics, never on top of lyrics)
		ctrls := fmt.Sprintf("%s   [ ⏮  %s  ⏭ ]   %s", elapsed, playIcon, total)
		barLen := 40
		if barLen > innerW-10 {
			barLen = innerW - 10
		}
		seek := renderProgressBar(ratio, barLen)
		rowControls := CenterLine(StylePurple.Render(ctrls), innerW)
		rowSeek := CenterLine(seek, innerW)
		rowVol := CenterLine(vStyled, innerW)

		bottomH := 4 // 1 gap + controls + seek + vol
		topH := 2    // 1 pad + banner

		lyricsH := innerH - topH - bottomH
		if lyricsH < 1 {
			lyricsH = 1
		}

		// Centered, vertically aligned synchronized lyrics
		lLines := renderZenLyrics(p.LyricsLines, p.LyricsCursor, curProgress, innerW, lyricsH)

		innerLines = append(innerLines, PadToWidth("", innerW))
		innerLines = append(innerLines, topRow)
		innerLines = append(innerLines, lLines...)
		innerLines = append(innerLines, PadToWidth("", innerW))
		innerLines = append(innerLines, rowControls)
		innerLines = append(innerLines, rowSeek)
		innerLines = append(innerLines, rowVol)

		for len(innerLines) < innerH {
			innerLines = append(innerLines, PadToWidth("", innerW))
		}
		if len(innerLines) > innerH {
			innerLines = innerLines[:innerH]
		}
	}

	var finalLines []string
	finalLines = append(finalLines, PadToWidth(topBorder, p.Width))
	for i := 0; i < innerH; i++ {
		row := ""
		if i < len(innerLines) {
			row = innerLines[i]
		}
		borderedRow := sideBar + PadToWidth(row, innerW) + sideBar
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

func joinHorizontalThemed(views []string, widths []int, height int) string {
	var splitViews [][]string
	for i, v := range views {
		lines := strings.Split(v, "\n")
		w := widths[i]
		for len(lines) < height {
			lines = append(lines, PadToWidth("", w))
		}
		if len(lines) > height {
			lines = lines[:height]
		}
		splitViews = append(splitViews, lines)
	}

	var combined []string
	for r := 0; r < height; r++ {
		var row strings.Builder
		for _, vLines := range splitViews {
			row.WriteString(vLines[r])
		}
		combined = append(combined, row.String())
	}
	return strings.Join(combined, "\n")
}

