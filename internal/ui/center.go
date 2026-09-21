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

	// 2. Tab Bar: 1 Tracks | 2 Lyrics | 3 History directly below search
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
		bodyLines = renderTracks(tracks, albums, artists, playlistName, currentPlayingTrackURI, selectedIndex, focused && !searchFocused, width, bodyH)
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
		{TabLyrics, "2 Lyrics"},
		{TabHistory, "3 History"},
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

// renderAlbumRow renders a single album row in an artist's discography section
func renderAlbumRow(idx int, a backend.Playlist, isSelected bool, focused bool, titleW, descW, countW int, width int) string {
	iconCol := " 💿 "
	titleCol := TruncateString(a.Name, titleW-1)
	titleCol += strings.Repeat(" ", titleW-ansi.StringWidth(titleCol))

	descCol := TruncateString(a.OwnerID, descW-1)
	descCol += strings.Repeat(" ", descW-ansi.StringWidth(descCol))

	countStr := fmt.Sprintf("%d tracks", a.TrackCount)
	if a.TrackCount == 1 {
		countStr = "1 track"
	} else if a.TrackCount == 0 {
		countStr = ""
	}
	countCol := TruncateString(countStr, countW)
	countCol += strings.Repeat(" ", countW-ansi.StringWidth(countCol))

	rawRow := iconCol + titleCol + descCol + countCol

	if isSelected && focused {
		return RenderPaddedLine(rawRow, StyleActiveFocusedBlock, width)
	} else if isSelected {
		return RenderPaddedLine(rawRow, StyleActiveUnfocusedBlock, width)
	}

	iconStyled := StyleLavender.Render(iconCol)
	titleStyled := StyleBold.Render(titleCol)
	descStyled := StylePeach.Render(descCol)
	countStyled := StyleFaint.Render(countCol)

	rowContent := iconStyled + titleStyled + descStyled + countStyled
	remPad := width - ansi.StringWidth(rowContent)
	if remPad > 0 {
		rowContent += strings.Repeat(" ", remPad)
	}
	return rowContent
}

// FormatTotalDuration formats total milliseconds and track count into a clean summary
func FormatTotalDuration(totalMs int, trackCount int) string {
	if trackCount == 0 {
		return ""
	}
	countLabel := fmt.Sprintf("%d tracks", trackCount)
	if trackCount == 1 {
		countLabel = "1 track"
	}
	if totalMs <= 0 {
		return countLabel
	}

	totalSeconds := totalMs / 1000
	hours := totalSeconds / 3600
	mins := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60

	if hours > 0 {
		if mins > 0 {
			return fmt.Sprintf("%s • %d hr %d min", countLabel, hours, mins)
		}
		return fmt.Sprintf("%s • %d hr", countLabel, hours)
	}
	if mins > 0 {
		return fmt.Sprintf("%s • %d min %d sec", countLabel, mins, secs)
	}
	return fmt.Sprintf("%s • %d sec", countLabel, secs)
}

func renderPlaylistHeader(title string, totalMs int, trackCount int, width int) string {
	infoStr := FormatTotalDuration(totalMs, trackCount)
	infoW := ansi.StringWidth(infoStr)

	leftMargin := "  ♫ "
	rightMargin := "  "

	availForTitle := width - infoW - ansi.StringWidth(leftMargin) - ansi.StringWidth(rightMargin) - 2
	if availForTitle < 10 {
		availForTitle = 10
	}

	titleTrunc := TruncateString(title, availForTitle)
	leftStyled := StylePurple.Render(leftMargin + titleTrunc)
	rightStyled := StyleFaint.Render(infoStr + rightMargin)

	usedW := ansi.StringWidth(leftMargin+titleTrunc) + infoW + ansi.StringWidth(rightMargin)
	gap := width - usedW
	if gap < 0 {
		gap = 0
	}

	return leftStyled + strings.Repeat(" ", gap) + rightStyled
}

// renderArtistRow renders a single artist row in search results
func renderArtistRow(idx int, a backend.Playlist, isSelected bool, focused bool, nameW, typeW, popW int, width int) string {
	iconCol := " 󰠃 "
	nameCol := TruncateString(a.Name, nameW-1)
	nameCol += strings.Repeat(" ", nameW-ansi.StringWidth(nameCol))

	typeCol := TruncateString(a.OwnerID, typeW-1)
	typeCol += strings.Repeat(" ", typeW-ansi.StringWidth(typeCol))

	popStr := fmt.Sprintf("%d%% pop", a.TrackCount)
	if a.TrackCount <= 0 {
		popStr = ""
	}
	popCol := TruncateString(popStr, popW)
	popCol += strings.Repeat(" ", popW-ansi.StringWidth(popCol))

	rawRow := iconCol + nameCol + typeCol + popCol

	if isSelected && focused {
		return RenderPaddedLine(rawRow, StyleActiveFocusedBlock, width)
	} else if isSelected {
		return RenderPaddedLine(rawRow, StyleActiveUnfocusedBlock, width)
	}

	iconStyled := StylePeach.Render(iconCol)
	nameStyled := StyleBold.Render(nameCol)
	typeStyled := StyleLavender.Render(typeCol)
	popStyled := StyleFaint.Render(popCol)

	rowContent := iconStyled + nameStyled + typeStyled + popStyled
	remPad := width - ansi.StringWidth(rowContent)
	if remPad > 0 {
		rowContent += strings.Repeat(" ", remPad)
	}
	return rowContent
}

// renderTracks renders a clean tracks table with scrolling, colored block row highlighting, and albums/artists sections
func renderTracks(
	tracks []backend.Track,
	albums []backend.Playlist,
	artists []backend.Playlist,
	playlistName string,
	currentPlayingURI string,
	selectedIndex int,
	focused bool,
	width, height int,
) []string {
	if len(tracks) == 0 && len(albums) == 0 && len(artists) == 0 {
		var lines []string
		title := "Tracks"
		if playlistName != "" {
			title = playlistName
		}
		titleHeader := StylePurple.Render("  ♫ " + title)
		lines = append(lines, PadToWidth(titleHeader, width))
		emptyMsg := StyleFaint.Render("No tracks, albums, or artists found. Select a playlist or press [/] to search.")
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

	var allVisualLines []string
	selectedLineIdx := 0

	totalMs := 0
	for _, t := range tracks {
		totalMs += t.DurationMs
	}

	title := "Tracks"
	if playlistName != "" {
		title = playlistName
	}
	headerLine := renderPlaylistHeader(title, totalMs, len(tracks), width)
	allVisualLines = append(allVisualLines, headerLine)

	if len(albums) > 0 || len(artists) > 0 {
		// Multi-section View: Songs + Albums + Artists
		if len(tracks) > 0 {
			allVisualLines = append(allVisualLines, PadToWidth("", width))
			subHeader := "  ♪ Top Tracks"
			if strings.HasPrefix(playlistName, "Search:") {
				subHeader = "  ♪ Songs"
			}
			allVisualLines = append(allVisualLines, PadToWidth(StyleLavender.Render(subHeader), width))
			headerNum := PadToWidth(" #", numW)
			headerTitle := PadToWidth("Title", titleW)
			headerArtist := PadToWidth("Artist", artistW)
			headerDuration := PadToWidth("Time", durW)
			tableHeader := StyleFaint.Render(headerNum + headerTitle + headerArtist + headerDuration)
			allVisualLines = append(allVisualLines, PadToWidth(tableHeader, width))
			allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

			for idx, t := range tracks {
				if idx == selectedIndex {
					selectedLineIdx = len(allVisualLines)
				}
				isSelected := idx == selectedIndex
				isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""
				lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, width)
				allVisualLines = append(allVisualLines, lineContent)
			}
		}

		// Albums Section
		if len(albums) > 0 {
			iconW := 4
			countW := 12
			remAlbW := width - iconW - countW - 4
			if remAlbW < 20 {
				remAlbW = 20
			}
			albTitleW := remAlbW * 55 / 100
			albDescW := remAlbW - albTitleW

			allVisualLines = append(allVisualLines, PadToWidth("", width))
			albSubHeader := "  💿 Albums & Discography"
			if strings.HasPrefix(playlistName, "Search:") {
				albSubHeader = "  💿 Albums"
			}
			allVisualLines = append(allVisualLines, PadToWidth(StyleLavender.Render(albSubHeader), width))
			albHeader := StyleFaint.Render("    " + PadToWidth("Album", albTitleW) + PadToWidth("Type • Year", albDescW) + PadToWidth("Tracks", countW))
			allVisualLines = append(allVisualLines, PadToWidth(albHeader, width))
			allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

			for aIdx, a := range albums {
				itemIdx := len(tracks) + aIdx
				if itemIdx == selectedIndex {
					selectedLineIdx = len(allVisualLines)
				}
				isSelected := itemIdx == selectedIndex
				lineContent := renderAlbumRow(aIdx, a, isSelected, focused, albTitleW, albDescW, countW, width)
				allVisualLines = append(allVisualLines, lineContent)
			}
		}

		// Artists Section
		if len(artists) > 0 {
			iconW := 4
			popW := 12
			remArtW := width - iconW - popW - 4
			if remArtW < 20 {
				remArtW = 20
			}
			artNameW := remArtW * 60 / 100
			artTypeW := remArtW - artNameW

			allVisualLines = append(allVisualLines, PadToWidth("", width))
			allVisualLines = append(allVisualLines, PadToWidth(StylePeach.Render("  󰠃 Artists"), width))
			artHeader := StyleFaint.Render("    " + PadToWidth("Artist", artNameW) + PadToWidth("Type", artTypeW) + PadToWidth("Popularity", popW))
			allVisualLines = append(allVisualLines, PadToWidth(artHeader, width))
			allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

			for artIdx, art := range artists {
				itemIdx := len(tracks) + len(albums) + artIdx
				if itemIdx == selectedIndex {
					selectedLineIdx = len(allVisualLines)
				}
				isSelected := itemIdx == selectedIndex
				lineContent := renderArtistRow(artIdx, art, isSelected, focused, artNameW, artTypeW, popW, width)
				allVisualLines = append(allVisualLines, lineContent)
			}
		}
	} else {
		// Standard Playlist or Single Album View
		headerNum := PadToWidth(" #", numW)
		headerTitle := PadToWidth("Title", titleW)
		headerArtist := PadToWidth("Artist", artistW)
		headerDuration := PadToWidth("Time", durW)
		tableHeader := StyleLavender.Render(headerNum + headerTitle + headerArtist + headerDuration)
		allVisualLines = append(allVisualLines, PadToWidth(tableHeader, width))
		allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))

		for idx, t := range tracks {
			if idx == selectedIndex {
				selectedLineIdx = len(allVisualLines)
			}
			isSelected := idx == selectedIndex
			isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""
			lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, width)
			allVisualLines = append(allVisualLines, lineContent)
		}
	}

	availRows := height
	if availRows < 1 {
		availRows = 1
	}

	scrollOffset := 0
	if selectedLineIdx >= availRows {
		scrollOffset = selectedLineIdx - availRows + 1
	}

	totalLines := len(allVisualLines)
	thumbH := 1
	thumbStart := 0
	if totalLines > availRows && availRows > 0 {
		thumbH = (availRows * availRows) / totalLines
		if thumbH < 1 {
			thumbH = 1
		}
		maxScroll := totalLines - availRows
		if maxScroll > 0 {
			thumbStart = (scrollOffset * (availRows - thumbH)) / maxScroll
		}
		if thumbStart+thumbH > availRows {
			thumbStart = availRows - thumbH
		}
		if thumbStart < 0 {
			thumbStart = 0
		}
	}

	var lines []string
	for i := 0; i < availRows; i++ {
		lineIdx := scrollOffset + i
		var rawLine string
		if lineIdx < len(allVisualLines) {
			rawLine = allVisualLines[lineIdx]
		}
		if totalLines > availRows {
			var scrollIndicator string
			if i >= thumbStart && i < thumbStart+thumbH {
				scrollIndicator = StylePurple.Render("█")
			} else {
				scrollIndicator = StyleFaint.Render("│")
			}
			trimmed := PadToWidth(rawLine, width-2)
			lines = append(lines, trimmed+" "+scrollIndicator)
		} else {
			lines = append(lines, PadToWidth(rawLine, width))
		}
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
