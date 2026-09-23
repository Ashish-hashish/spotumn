package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"spotumn/internal/backend"
)

// RenderRightLines generates lines for the right sidebar (Now Playing + Live Queue)
func RenderRightLines(
	artANSI string,
	currentTrack *backend.Track,
	queue []backend.Track,
	queueIndex int,
	focused bool,
	width, height int,
) []string {
	if width < 16 {
		width = 16
	}
	if height < 6 {
		height = 6
	}

	var lines []string

	// Header
	infoTitle := " Now Playing "
	if focused && queueIndex < 0 {
		infoTitle = " ● Now Playing "
	}
	lines = append(lines, PadToWidth(StylePurple.Render(infoTitle), width))

	// Art lines (enlarged, constrained to sidebar width without ANSI bleed)
	if artANSI != "" {
		for _, row := range strings.Split(artANSI, "\n") {
			row = strings.TrimRight(row, "\r")
			if row == "" {
				continue
			}
			if strings.Contains(row, "\x1b[") {
				if !strings.HasSuffix(row, "\x1b[0m") {
					row += "\x1b[0m"
				}
				w := ansi.StringWidth(row)
				if w > width {
					row = ansi.Truncate(row, width, "") + "\x1b[0m"
				}
			}
			lines = append(lines, CenterLine(row, width))
		}
	}

	// Track Info
	if currentTrack != nil && currentTrack.Name != "" {
		trackName := TruncateString(currentTrack.Name, width-4)
		artistName := TruncateString(currentTrack.Artist, width-4)
		albumName := TruncateString(currentTrack.Album, width-4)

		lines = append(lines, PadToWidth(BgPad(2)+StyleBold.Render(trackName), width))
		lines = append(lines, PadToWidth(BgPad(2)+StyleLavender.Render(artistName), width))

		if albumName != "" {
			lines = append(lines, PadToWidth(BgPad(2)+StyleFaint.Render(albumName), width))
		}
	} else {
		lines = append(lines, PadToWidth(BgPad(2)+StyleFaint.Render("No track playing"), width))
	}

	// Divider before queue with right-aligned keybind badge
	lines = append(lines, PadToWidth("", width))
	prefix := "── UP NEXT "
	badge := BgPad(1) + StylePurple.Render("⌜") + StyleBold.Render("q") + StylePurple.Render("⌟") + BgPad(1) + StyleLavender.Render("Queue") + StyleFaint.Render(" ──")
	rawBadge := " ⌜q⌟ Queue ──"
	neededDashes := width - ansi.StringWidth(prefix) - ansi.StringWidth(rawBadge)
	var queueHeaderLine string
	if neededDashes > 0 {
		queueHeaderLine = StyleFaint.Render(prefix+strings.Repeat("─", neededDashes)) + badge
	} else {
		queueHeaderLine = StyleFaint.Render(TruncateString(prefix, width))
	}
	lines = append(lines, PadToWidth(queueHeaderLine, width))

	headerLinesCount := len(lines)
	availableQueueRows := height - headerLinesCount
	if availableQueueRows < 1 {
		availableQueueRows = 1
	}

	scrollOffset := 0
	if queueIndex >= availableQueueRows {
		scrollOffset = queueIndex - availableQueueRows + 1
	}

	for i := 0; i < availableQueueRows; i++ {
		qIdx := scrollOffset + i
		if qIdx >= len(queue) {
			lines = append(lines, PadToWidth("", width))
			continue
		}

		item := queue[qIdx]
		isSelected := qIdx == queueIndex

		numPrefix := fmt.Sprintf("%d. ", qIdx+1)
		availTitleW := width - len(numPrefix) - 4
		if availTitleW < 6 {
			availTitleW = 6
		}

		display := numPrefix + item.Name
		if item.Artist != "" {
			display += " - " + item.Artist
		}
		truncDisplay := TruncateString(display, width-4)

		rawDisplay := "  " + truncDisplay
		if isSelected {
			rawDisplay = "❯ " + truncDisplay
		}

		var lineContent string
		if isSelected && focused {
			lineContent = RenderPaddedLine(rawDisplay, StyleActiveFocusedBlock, width)
		} else if isSelected {
			lineContent = RenderPaddedLine(rawDisplay, StyleActiveUnfocusedBlock, width)
		} else {
			// Unselected queue row: distinct colors for title and artist
			numStyled := StyleFaint.Render("  " + numPrefix)
			remW := width - 4 - ansi.StringWidth(numPrefix)
			if remW < 4 {
				remW = 4
			}
			var contentStyled string
			if item.Artist != "" {
				namePart := TruncateString(item.Name, remW*60/100)
				artistPart := TruncateString(item.Artist, remW-ansi.StringWidth(namePart)-3)
				contentStyled = StyleBold.Render(namePart) + StyleFaint.Render(" - ") + StyleLavender.Render(artistPart)
			} else {
				contentStyled = StyleBold.Render(TruncateString(item.Name, remW))
			}
			rowText := numStyled + contentStyled
			remPad := width - ansi.StringWidth(rowText)
			if remPad > 0 {
				rowText += BgPad(remPad)
			}
			lineContent = rowText
		}

		lines = append(lines, lineContent)
	}

	for len(lines) < height {
		lines = append(lines, PadToWidth("", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return lines
}

// RenderMergedRight renders the right pane inside a border that highlights when focused
func RenderMergedRight(
	artANSI string,
	currentTrack *backend.Track,
	queue []backend.Track,
	queueIndex int,
	focused bool,
	width, height int,
) string {
	contentW := width - 2
	if contentW < 16 {
		contentW = 16
	}
	contentH := height - 2
	if contentH < 6 {
		contentH = 6
	}

	lines := RenderRightLines(artANSI, currentTrack, queue, queueIndex, focused, contentW, contentH)
	boxStyle := PanelBox(focused, width, height)
	return boxStyle.Render(strings.Join(lines, "\n"))
}
