package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
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

	// Art lines (enlarged)
	if artANSI != "" {
		artRows := strings.Split(artANSI, "\n")
		for _, row := range artRows {
			rowW := ansi.StringWidth(row)
			padLeft := (width - rowW) / 2
			if padLeft < 0 {
				padLeft = 0
			}
			centeredRow := strings.Repeat(" ", padLeft) + row
			lines = append(lines, PadToWidth(centeredRow, width))
		}
	}

	// Track Info
	if currentTrack != nil && currentTrack.Name != "" {
		trackName := TruncateString(currentTrack.Name, width-4)
		artistName := TruncateString(currentTrack.Artist, width-4)
		albumName := TruncateString(currentTrack.Album, width-4)

		lines = append(lines, PadToWidth("  "+StyleBold.Render(trackName), width))
		lines = append(lines, PadToWidth("  "+StyleLavender.Render(artistName), width))

		if albumName != "" {
			lines = append(lines, PadToWidth("  "+StyleFaint.Render(albumName), width))
		}
	} else {
		lines = append(lines, PadToWidth("  "+StyleFaint.Render("No track playing"), width))
	}

	// Divider before Queue
	lines = append(lines, PadToWidth("", width))
	queueHeader := "── UP NEXT ──────────────────────────────"
	lines = append(lines, PadToWidth(StyleFaint.Render(TruncateString(queueHeader, width)), width))

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
				rowText += strings.Repeat(" ", remPad)
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

	boxStyle := lipgloss.NewStyle().Width(width).Height(height)
	if focused {
		boxStyle = boxStyle.Border(lipgloss.ThickBorder()).BorderForeground(CurrentTheme.Purple).Bold(true)
	} else {
		boxStyle = boxStyle.Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Overlay)
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}
