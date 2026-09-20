package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"spotumn/internal/backend"
	"spotumn/internal/lyrics"
)

// RenderCenterLines generates content lines for the center pane
func RenderCenterLines(
	currentTab CenterTab,
	tracks []backend.Track,
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
) []string {
	if width < 20 {
		width = 20
	}
	if height < 6 {
		height = 6
	}

	var lines []string

	// 1. Prominent Search Bar at the very top in its own border
	searchBarLines := renderProminentSearchBar(searchQuery, searchFocused, width)
	lines = append(lines, searchBarLines...)

	// 2. Tab Bar: 1 Tracks | 2 History | 3 Lyrics directly below search
	tabBar := renderTabBar(currentTab, width)
	lines = append(lines, tabBar)
	lines = append(lines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

	bodyH := height - len(lines)
	if bodyH < 1 {
		bodyH = 1
	}

	// 3. Tab Content
	var bodyLines []string
	switch currentTab {
	case TabTracks:
		bodyLines = renderTracks(tracks, playlistName, currentPlayingTrackURI, selectedIndex, focused && !searchFocused, width, bodyH)
	case TabHistory:
		bodyLines = renderHistory(history, currentPlayingTrackURI, selectedIndex, focused && !searchFocused, width, bodyH)
	case TabLyrics:
		bodyLines = renderLyrics(lyricsLines, lyricsCursor, progressMs, focused && !searchFocused, width, bodyH)
	}

	lines = append(lines, bodyLines...)

	for len(lines) < height {
		lines = append(lines, PadToWidth("", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return lines
}

// RenderCenter renders the main center pane inside a rounded border that highlights when focused
func RenderCenter(
	currentTab CenterTab,
	tracks []backend.Track,
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
) string {
	contentW := width - 2
	if contentW < 20 {
		contentW = 20
	}
	contentH := height - 2
	if contentH < 6 {
		contentH = 6
	}

	lines := RenderCenterLines(
		currentTab,
		tracks,
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
		contentW,
		contentH,
	)

	boxStyle := lipgloss.NewStyle().Width(width).Height(height)
	if focused {
		boxStyle = boxStyle.Border(lipgloss.ThickBorder()).BorderForeground(CurrentTheme.Purple).Bold(true)
	} else {
		boxStyle = boxStyle.Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Overlay)
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderTabBar(currentTab CenterTab, width int) string {
	tabs := []struct {
		tab   CenterTab
		label string
	}{
		{TabTracks, "1 Tracks"},
		{TabHistory, "2 History"},
		{TabLyrics, "3 Lyrics"},
	}

	var sb strings.Builder
	sb.WriteString(" ")
	for _, t := range tabs {
		if t.tab == currentTab {
			style := StylePurple
			sb.WriteString(style.Render("[ " + t.label + " ]"))
		} else {
			style := StyleFaint
			sb.WriteString(style.Render("  " + t.label + "  "))
		}
		sb.WriteString("  ")
	}

	return PadToWidth(sb.String(), width)
}

func renderProminentSearchBar(query string, focused bool, width int) []string {
	prompt := "  Search: "
	displayQuery := query
	if focused {
		displayQuery += "█"
	} else if displayQuery == "" {
		displayQuery = "Press [/] to search Spotify..."
	}

	barW := width - 4
	if barW < 14 {
		barW = 14
	}

	var promptStyled, textStyled string
	barStyle := lipgloss.NewStyle().Width(barW)
	if focused {
		promptStyled = StylePurple.Render(prompt)
		textStyled = StyleNormal.Render(displayQuery)
		barStyle = barStyle.Border(lipgloss.ThickBorder()).BorderForeground(CurrentTheme.Purple).Bold(true)
	} else {
		promptStyled = StyleFaint.Render(prompt)
		textStyled = StyleFaint.Render(displayQuery)
		barStyle = barStyle.Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Overlay)
	}

	innerContent := promptStyled + textStyled

	var res []string
	for _, l := range strings.Split(barStyle.Render(innerContent), "\n") {
		res = append(res, PadToWidth("  "+l, width))
	}
	return res
}

// renderTrackRow renders a single track row with colored block selection and distinct title/artist colors
func renderTrackRow(idx int, t backend.Track, isSelected bool, isPlaying bool, focused bool, titleW, artistW int, width int) string {
	numCol := fmt.Sprintf("%2d ", idx+1)
	if isPlaying {
		numCol = " ▶ "
	}

	titleCol := TruncateString(t.Name, titleW-1)
	titleCol += strings.Repeat(" ", titleW-ansi.StringWidth(titleCol))

	artistCol := TruncateString(t.Artist, artistW-1)
	artistCol += strings.Repeat(" ", artistW-ansi.StringWidth(artistCol))

	durCol := FormatDuration(t.DurationMs)
	rawRow := numCol + titleCol + artistCol + durCol

	if isSelected && focused {
		return RenderPaddedLine(rawRow, StyleActiveFocusedBlock, width)
	} else if isSelected {
		return RenderPaddedLine(rawRow, StyleActiveUnfocusedBlock, width)
	}

	// Distinct colors: title in bold text, artist in lavender, duration in faint
	var numStyled, titleStyled string
	if isPlaying {
		numStyled = StyleMint.Render(numCol)
		titleStyled = StyleMint.Render(titleCol)
	} else {
		numStyled = StyleFaint.Render(numCol)
		titleStyled = StyleBold.Render(titleCol)
	}
	artistStyled := StyleLavender.Render(artistCol)
	durStyled := StyleFaint.Render(durCol)

	rowContent := numStyled + titleStyled + artistStyled + durStyled
	remPad := width - ansi.StringWidth(rowContent)
	if remPad > 0 {
		rowContent += strings.Repeat(" ", remPad)
	}
	return rowContent
}

// renderTracks renders a clean tracks table with scrolling and colored block row highlighting
func renderTracks(
	tracks []backend.Track,
	playlistName string,
	currentPlayingURI string,
	selectedIndex int,
	focused bool,
	width, height int,
) []string {
	var lines []string

	title := "Tracks"
	if playlistName != "" {
		title = "♫ " + playlistName
	}
	titleHeader := StylePurple.Render("  " + title)
	lines = append(lines, PadToWidth(titleHeader, width))

	if len(tracks) == 0 {
		emptyMsg := StyleFaint.Render("No tracks in this playlist. Select a playlist or press [/] to search.")
		lines = append(lines, PadToWidth("", width))
		lines = append(lines, PadToWidth("  "+emptyMsg, width))
		return lines
	}

	numW := 4
	durW := 6
	remW := width - numW - durW - 4
	if remW < 20 {
		remW = 20
	}
	titleW := remW * 55 / 100
	artistW := remW - titleW

	headerNum := PadToWidth(" #", numW)
	headerTitle := PadToWidth("Title", titleW)
	headerArtist := PadToWidth("Artist", artistW)
	headerDuration := PadToWidth("Time", durW)

	tableHeader := StyleLavender.Render(headerNum + headerTitle + headerArtist + headerDuration)
	lines = append(lines, PadToWidth(tableHeader, width))
	lines = append(lines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

	availRows := height - len(lines)
	if availRows < 1 {
		availRows = 1
	}

	scrollOffset := 0
	if selectedIndex >= availRows {
		scrollOffset = selectedIndex - availRows + 1
	}

	for i := 0; i < availRows; i++ {
		idx := scrollOffset + i
		if idx >= len(tracks) {
			lines = append(lines, PadToWidth("", width))
			continue
		}

		t := tracks[idx]
		isSelected := idx == selectedIndex
		isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""

		lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, width)
		lines = append(lines, lineContent)
	}

	return lines
}

// renderHistory renders the listening history (recently played tracks)
func renderHistory(
	history []backend.Track,
	currentPlayingURI string,
	selectedIndex int,
	focused bool,
	width, height int,
) []string {
	var lines []string

	titleHeader := StylePurple.Render("   Listening History (Recently Played)")
	lines = append(lines, PadToWidth(titleHeader, width))

	if len(history) == 0 {
		emptyMsg := StyleFaint.Render("No listening history available yet. Play some tracks on Spotify!")
		lines = append(lines, PadToWidth("", width))
		lines = append(lines, PadToWidth("  "+emptyMsg, width))
		return lines
	}

	numW := 4
	durW := 6
	remW := width - numW - durW - 4
	if remW < 20 {
		remW = 20
	}
	titleW := remW * 55 / 100
	artistW := remW - titleW

	headerNum := PadToWidth(" #", numW)
	headerTitle := PadToWidth("Title", titleW)
	headerArtist := PadToWidth("Artist", artistW)
	headerDuration := PadToWidth("Time", durW)

	tableHeader := StyleLavender.Render(headerNum + headerTitle + headerArtist + headerDuration)
	lines = append(lines, PadToWidth(tableHeader, width))
	lines = append(lines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

	availRows := height - len(lines)
	if availRows < 1 {
		availRows = 1
	}

	scrollOffset := 0
	if selectedIndex >= availRows {
		scrollOffset = selectedIndex - availRows + 1
	}

	for i := 0; i < availRows; i++ {
		idx := scrollOffset + i
		if idx >= len(history) {
			lines = append(lines, PadToWidth("", width))
			continue
		}

		t := history[idx]
		isSelected := idx == selectedIndex
		isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""

		lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, width)
		lines = append(lines, lineContent)
	}

	return lines
}

// renderLyrics renders synchronized lyrics with colored block active highlight
func renderLyrics(lines []lyrics.Line, cursorLine int, progressMs int, focused bool, width, height int) []string {
	var output []string

	if len(lines) == 0 {
		msg := StyleFaint.Render("No synced lyrics found for this track")
		output = append(output, PadToWidth("", width))
		output = append(output, PadToWidth("  "+msg, width))
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

	halfHeight := height / 2
	startIdx := targetCenter - halfHeight
	if startIdx < 0 {
		startIdx = 0
	}

	for i := 0; i < height; i++ {
		idx := startIdx + i
		if idx >= len(lines) {
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

		var lineRendered string
		if isSinging {
			// Solid Mint colored block for active singing lyric with dark contrasting text
			lineRendered = RenderPaddedLine("  ❯  "+text, StyleActiveLyricsBlock, width)
		} else if isCursor && focused {
			// Subtle Purple colored block for manual cursor navigation
			lineRendered = RenderPaddedLine("  ➜  "+text, StyleActiveFocusedBlock, width)
		} else {
			lineRendered = PadToWidth("     "+StyleFaint.Render(text), width)
		}

		output = append(output, lineRendered)
	}

	return output
}
