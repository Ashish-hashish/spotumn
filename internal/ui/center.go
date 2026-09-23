package ui

import (
	"fmt"
	"strings"

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
	containerURI ...string,
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
		var curURI string
		if len(containerURI) > 0 {
			curURI = containerURI[0]
		}
		bodyLines = renderTracks(tracks, albums, artists, playlistName, currentPlayingTrackURI, selectedIndex, focused && !searchFocused, width, bodyH, curURI)
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
	containerURI ...string,
) string {
	contentW := width - 2
	contentH := height - 2
	if contentW < 10 {
		contentW = 10
	}
	if contentH < 3 {
		contentH = 3
	}

	var curURI string
	if len(containerURI) > 0 {
		curURI = containerURI[0]
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
		curURI,
	)

	boxStyle := PanelBox(focused, width, height)
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
	sb.WriteString(BgPad(1))
	for _, t := range tabs {
		if t.tab == currentTab {
			style := StylePurple
			sb.WriteString(style.Render("[ " + t.label + " ]"))
		} else {
			style := StyleFaint
			sb.WriteString(style.Render("  " + t.label + "  "))
		}
		sb.WriteString(BgPad(2))
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

	maxQueryW := barW - 2 - ansi.StringWidth(prompt)
	if maxQueryW > 4 {
		displayQuery = TruncateString(displayQuery, maxQueryW)
	}

	var promptStyled, textStyled string
	if focused {
		promptStyled = StylePurple.Render(prompt)
		textStyled = StyleNormal.Render(displayQuery)
	} else {
		promptStyled = StyleFaint.Render(prompt)
		textStyled = StyleFaint.Render(displayQuery)
	}

	innerContent := PadToWidth(promptStyled+textStyled, barW-2)
	barStyle := PanelBox(focused, barW, 0)

	var res []string
	for _, l := range strings.Split(barStyle.Render(innerContent), "\n") {
		res = append(res, PadToWidth(BgPad(2)+l, width))
	}
	if len(res) > 3 {
		res = res[:3]
	}
	return res
}

// renderTrackRow renders a single track row with colored block selection and distinct title/artist colors
func renderTrackRow(idx int, t backend.Track, isSelected bool, isPlaying bool, focused bool, titleW, artistW int, width int) string {
	numCol := fmt.Sprintf("%2d ", idx+1)
	if isPlaying {
		numCol = " ▶ "
	}

	titleCol := PadPlain(t.Name, titleW)
	artistCol := PadPlain(t.Artist, artistW)
	durCol := PadPlain(FormatDuration(t.DurationMs), 6)
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
	return PadToWidth(rowContent, width)
}

// renderAlbumRow renders a single album row in an artist's discography section
func renderAlbumRow(idx int, a backend.Playlist, isSelected bool, focused bool, titleW, descW, countW int, width int) string {
	numCol := PadPlain(fmt.Sprintf("%2d", idx+1), 4)
	iconCol := " 💿 "
	titleCol := PadPlain(a.Name, titleW)
	descCol := PadPlain(a.OwnerID, descW)

	countStr := fmt.Sprintf("%d tracks", a.TrackCount)
	if a.TrackCount == 1 {
		countStr = "1 track"
	} else if a.TrackCount == 0 {
		countStr = ""
	}
	countCol := PadPlain(countStr, countW)

	rawRow := numCol + iconCol + titleCol + descCol + countCol

	if isSelected && focused {
		return RenderPaddedLine(rawRow, StyleActiveFocusedBlock, width)
	} else if isSelected {
		return RenderPaddedLine(rawRow, StyleActiveUnfocusedBlock, width)
	}

	numStyled := StyleFaint.Render(numCol)
	iconStyled := StyleLavender.Render(iconCol)
	titleStyled := StyleBold.Render(titleCol)
	descStyled := StylePeach.Render(descCol)
	countStyled := StyleFaint.Render(countCol)

	rowContent := numStyled + iconStyled + titleStyled + descStyled + countStyled
	return PadToWidth(rowContent, width)
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

	return leftStyled + BgPad(gap) + rightStyled
}

// renderArtistRow renders a single artist row in search results
func renderArtistRow(idx int, a backend.Playlist, isSelected bool, focused bool, nameW, typeW, popW int, width int) string {
	numCol := PadPlain(fmt.Sprintf("%2d", idx+1), 4)
	iconCol := " 󰠃 "
	nameCol := PadPlain(a.Name, nameW)
	typeCol := PadPlain(a.OwnerID, typeW)

	popStr := fmt.Sprintf("%d%% pop", a.TrackCount)
	if a.TrackCount <= 0 {
		popStr = ""
	}
	popCol := PadPlain(popStr, popW)

	rawRow := numCol + iconCol + nameCol + typeCol + popCol

	if isSelected && focused {
		return RenderPaddedLine(rawRow, StyleActiveFocusedBlock, width)
	} else if isSelected {
		return RenderPaddedLine(rawRow, StyleActiveUnfocusedBlock, width)
	}

	numStyled := StyleFaint.Render(numCol)
	iconStyled := StylePeach.Render(iconCol)
	nameStyled := StyleBold.Render(nameCol)
	typeStyled := StyleLavender.Render(typeCol)
	popStyled := StyleFaint.Render(popCol)

	rowContent := numStyled + iconStyled + nameStyled + typeStyled + popStyled
	return PadToWidth(rowContent, width)
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
	containerURI ...string,
) []string {
	var curContainerURI string
	if len(containerURI) > 0 {
		curContainerURI = containerURI[0]
	}

	if len(tracks) == 0 && len(albums) == 0 && len(artists) == 0 {
		var lines []string
		title := "Tracks"
		if playlistName != "" {
			title = playlistName
		}
		titleHeader := StylePurple.Render("  ♫ " + title)
		lines = append(lines, PadToWidth(titleHeader, width))
		emptyMsg := StyleFaint.Render("No tracks, albums, or artists found. Select a playlist or press [/] to search.")
		lines = append(lines, PadToWidth(BgPad(2)+emptyMsg, width))
		return lines
	}

	availRows := height
	if availRows < 1 {
		availRows = 1
	}

	totalEstimatedLines := 1
	if len(albums) > 0 || len(artists) > 0 {
		if len(tracks) > 0 {
			totalEstimatedLines += 4 + len(tracks)
		}
		if len(albums) > 0 {
			totalEstimatedLines += 4 + len(albums)
		}
		if len(artists) > 0 {
			totalEstimatedLines += 4 + len(artists)
		}
	} else {
		totalEstimatedLines += 2 + len(tracks)
	}

	needsScroll := totalEstimatedLines > availRows
	tableW := width
	if needsScroll {
		tableW = width - 2
	}
	if tableW < 14 {
		tableW = 14
	}

	numW := 4
	durW := 6
	remW := tableW - numW - durW - 2
	if remW < 12 {
		remW = 12
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
	headerLine := renderPlaylistHeader(title, totalMs, len(tracks), tableW)
	allVisualLines = append(allVisualLines, headerLine)

	if len(albums) > 0 || len(artists) > 0 {
		// Multi-section View: Songs + Albums + Artists
		if len(tracks) > 0 {
			allVisualLines = append(allVisualLines, PadToWidth("", tableW))
			subHeader := "  ♪ Top Tracks"
			isAlbumView := strings.HasPrefix(curContainerURI, "spotify:album:") ||
				(len(tracks) > 0 && tracks[0].Artist != "" && !strings.EqualFold(strings.TrimSpace(playlistName), strings.TrimSpace(tracks[0].Artist)) && !strings.HasPrefix(curContainerURI, "spotify:artist:"))
			if strings.HasPrefix(playlistName, "Search:") {
				subHeader = "  ♪ Songs"
			} else if isAlbumView {
				subHeader = "  ♪ Tracks"
			}
			allVisualLines = append(allVisualLines, PadToWidth(StyleLavender.Render(subHeader), tableW))
			headerNum := PadPlain(" #", numW)
			headerTitle := PadPlain("Title", titleW)
			headerArtist := PadPlain("Artist", artistW)
			headerDuration := PadPlain("Time", durW)
			tableHeader := StyleFaint.Render(headerNum + headerTitle + headerArtist + headerDuration)
			allVisualLines = append(allVisualLines, PadToWidth(tableHeader, tableW))
			allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", tableW)), tableW))

			for idx, t := range tracks {
				if idx == selectedIndex {
					selectedLineIdx = len(allVisualLines)
				}
				isSelected := idx == selectedIndex
				isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""
				lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, tableW)
				allVisualLines = append(allVisualLines, lineContent)
			}
		}

		// Albums Section
		if len(albums) > 0 {
			numW := 4
			iconW := 4
			countW := 12
			remAlbW := tableW - numW - iconW - countW - 2
			if remAlbW < 12 {
				remAlbW = 12
			}
			albTitleW := remAlbW * 55 / 100
			albDescW := remAlbW - albTitleW

			isAlbumView := strings.HasPrefix(curContainerURI, "spotify:album:") ||
				(len(tracks) > 0 && tracks[0].Artist != "" && !strings.EqualFold(strings.TrimSpace(playlistName), strings.TrimSpace(tracks[0].Artist)) && !strings.HasPrefix(curContainerURI, "spotify:artist:"))

			allVisualLines = append(allVisualLines, PadToWidth("", tableW))
			albSubHeader := "  💿 Albums & Discography"
			if strings.HasPrefix(playlistName, "Search:") {
				albSubHeader = "  💿 Albums"
			} else if isAlbumView {
				artistName := ""
				if len(tracks) > 0 && tracks[0].Artist != "" {
					artistName = tracks[0].Artist
				}
				if artistName != "" {
					albSubHeader = fmt.Sprintf("  💿 More from %s", artistName)
				} else {
					albSubHeader = "  💿 More from this Artist"
				}
			}
			allVisualLines = append(allVisualLines, PadToWidth(StyleLavender.Render(albSubHeader), tableW))
			albHeader := StyleFaint.Render(PadPlain(" #", numW) + "    " + PadPlain("Album", albTitleW) + PadPlain("Type • Year", albDescW) + PadPlain("Tracks", countW))
			allVisualLines = append(allVisualLines, PadToWidth(albHeader, tableW))
			allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", tableW)), tableW))

			for aIdx, a := range albums {
				itemIdx := len(tracks) + aIdx
				if itemIdx == selectedIndex {
					selectedLineIdx = len(allVisualLines)
				}
				isSelected := itemIdx == selectedIndex
				lineContent := renderAlbumRow(aIdx, a, isSelected, focused, albTitleW, albDescW, countW, tableW)
				allVisualLines = append(allVisualLines, lineContent)
			}
		}

		// Artists Section
		if len(artists) > 0 {
			numW := 4
			iconW := 4
			popW := 12
			remArtW := tableW - numW - iconW - popW - 2
			if remArtW < 12 {
				remArtW = 12
			}
			artNameW := remArtW * 60 / 100
			artTypeW := remArtW - artNameW

			allVisualLines = append(allVisualLines, PadToWidth("", tableW))
			allVisualLines = append(allVisualLines, PadToWidth(StylePeach.Render("  󰠃 Artists"), tableW))
			artHeader := StyleFaint.Render(PadPlain(" #", numW) + "    " + PadPlain("Artist", artNameW) + PadPlain("Type", artTypeW) + PadPlain("Popularity", popW))
			allVisualLines = append(allVisualLines, PadToWidth(artHeader, tableW))
			allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", tableW)), tableW))

			for artIdx, art := range artists {
				itemIdx := len(tracks) + len(albums) + artIdx
				if itemIdx == selectedIndex {
					selectedLineIdx = len(allVisualLines)
				}
				isSelected := itemIdx == selectedIndex
				lineContent := renderArtistRow(artIdx, art, isSelected, focused, artNameW, artTypeW, popW, tableW)
				allVisualLines = append(allVisualLines, lineContent)
			}
		}
	} else {
		// Standard Playlist or Single Album View
		headerNum := PadPlain(" #", numW)
		headerTitle := PadPlain("Title", titleW)
		headerArtist := PadPlain("Artist", artistW)
		headerDuration := PadPlain("Time", durW)
		tableHeader := StyleLavender.Render(headerNum + headerTitle + headerArtist + headerDuration)
		allVisualLines = append(allVisualLines, PadToWidth(tableHeader, tableW))
		allVisualLines = append(allVisualLines, PadToWidth(StyleFaint.Render(strings.Repeat("─", tableW)), tableW))

		for idx, t := range tracks {
			if idx == selectedIndex {
				selectedLineIdx = len(allVisualLines)
			}
			isSelected := idx == selectedIndex
			isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""
			lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, tableW)
			allVisualLines = append(allVisualLines, lineContent)
		}
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
			trimmed := PadToWidth(rawLine, tableW)
			lines = append(lines, trimmed+BgPad(1)+scrollIndicator)
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
		lines = append(lines, PadToWidth(BgPad(2)+emptyMsg, width))
		return lines
	}

	availRows := height - len(lines) - 2
	if availRows < 1 {
		availRows = 1
	}

	needsScroll := len(history) > availRows
	tableW := width
	if needsScroll {
		tableW = width - 2
	}
	if tableW < 14 {
		tableW = 14
	}

	numW := 4
	durW := 6
	remW := tableW - numW - durW - 2
	if remW < 12 {
		remW = 12
	}
	titleW := remW * 55 / 100
	artistW := remW - titleW

	headerNum := PadPlain(" #", numW)
	headerTitle := PadPlain("Title", titleW)
	headerArtist := PadPlain("Artist", artistW)
	headerDuration := PadPlain("Time", durW)

	tableHeader := StyleLavender.Render(headerNum + headerTitle + headerArtist + headerDuration)
	if needsScroll {
		lines = append(lines, PadToWidth(tableHeader, tableW)+BgPad(1)+StyleFaint.Render("│"))
		lines = append(lines, PadToWidth(StyleFaint.Render(strings.Repeat("─", tableW)), tableW)+BgPad(1)+StyleFaint.Render("│"))
	} else {
		lines = append(lines, PadToWidth(tableHeader, width))
		lines = append(lines, PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width))
	}

	scrollOffset := 0
	if selectedIndex >= availRows {
		scrollOffset = selectedIndex - availRows + 1
	}

	totalItems := len(history)
	thumbH := 1
	thumbStart := 0
	if totalItems > availRows && availRows > 0 {
		thumbH = (availRows * availRows) / totalItems
		if thumbH < 1 {
			thumbH = 1
		}
		maxScroll := totalItems - availRows
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

	for i := 0; i < availRows; i++ {
		idx := scrollOffset + i
		if idx >= len(history) {
			if needsScroll {
				lines = append(lines, PadToWidth("", tableW)+BgPad(1)+StyleFaint.Render("│"))
			} else {
				lines = append(lines, PadToWidth("", width))
			}
			continue
		}

		t := history[idx]
		isSelected := idx == selectedIndex
		isPlaying := t.URI == currentPlayingURI && currentPlayingURI != ""

		lineContent := renderTrackRow(idx, t, isSelected, isPlaying, focused, titleW, artistW, tableW)
		if needsScroll {
			var scrollIndicator string
			if i >= thumbStart && i < thumbStart+thumbH {
				scrollIndicator = StylePurple.Render("█")
			} else {
				scrollIndicator = StyleFaint.Render("│")
			}
			lines = append(lines, lineContent+BgPad(1)+scrollIndicator)
		} else {
			lines = append(lines, lineContent)
		}
	}

	return lines
}

// renderLyrics renders synchronized lyrics or unsynced plain lyrics with pointer navigation
func renderLyrics(lines []lyrics.Line, cursorLine int, progressMs int, focused bool, width, height int) []string {
	var output []string

	if len(lines) == 0 {
		msg := StyleFaint.Render("No lyrics found for this track")
		output = append(output, PadToWidth("", width))
		output = append(output, PadToWidth(BgPad(2)+msg, width))
		return output
	}

	isSynced := false
	for _, l := range lines {
		if l.TimeMs > 0 {
			isSynced = true
			break
		}
	}

	if !isSynced {
		output = append(output, PadToWidth(BgPad(2)+StyleFaint.Render("[Lyrics not synced]"), width))
		height--
		if height < 1 {
			height = 1
		}

		targetCenter := cursorLine
		if targetCenter < 0 {
			targetCenter = 0
		}
		if targetCenter >= len(lines) {
			targetCenter = len(lines) - 1
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

			maxLyricW := width - 6
			if maxLyricW < 4 {
				maxLyricW = 4
			}
			text = TruncateString(text, maxLyricW)

			isCursor := idx == cursorLine
			var lineRendered string
			if isCursor {
				lineRendered = RenderPaddedLine("  ❯  "+text, StyleActiveFocusedBlock, width)
			} else {
				lineRendered = PadToWidth(BgPad(5)+StyleFaint.Render(text), width)
			}
			output = append(output, lineRendered)
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

		maxLyricW := width - 6
		if maxLyricW < 4 {
			maxLyricW = 4
		}
		text = TruncateString(text, maxLyricW)

		isSinging := idx == activeIdx
		isCursor := idx == cursorLine

		var lineRendered string
		if isSinging {
			lineRendered = RenderPaddedLine("  ❯  "+text, StyleActiveLyricsBlock, width)
		} else if isCursor && focused {
			lineRendered = RenderPaddedLine("  ➜  "+text, StyleActiveFocusedBlock, width)
		} else {
			lineRendered = PadToWidth(BgPad(5)+StyleFaint.Render(text), width)
		}

		output = append(output, lineRendered)
	}

	return output
}
