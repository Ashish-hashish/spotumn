package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"spotumn/internal/backend"
)

type PlaylistFilter int

const (
	FilterAll PlaylistFilter = iota
	FilterSpotify
	FilterByYou
	FilterAlbums
	FilterArtists
)

func (f PlaylistFilter) String() string {
	switch f {
	case FilterSpotify:
		return "By Spotify"
	case FilterByYou:
		return "By You"
	case FilterAlbums:
		return "Albums"
	case FilterArtists:
		return "Artists"
	default:
		return "ALL"
	}
}

// RenderNavLines generates the lines for the left navigation pane
func RenderNavLines(playlists []backend.Playlist, pinnedURIs map[string]bool, filter PlaylistFilter, selectedIndex int, focused bool, width, height int) []string {
	if width < 10 {
		width = 10
	}
	if height < 3 {
		height = 3
	}

	categoryTitle := "Playlists"
	if filter == FilterAlbums {
		categoryTitle = "Albums"
	} else if filter == FilterArtists {
		categoryTitle = "Artists"
	}

	filterName := filter.String()
	prefixDot := ""
	if focused {
		prefixDot = "● "
	}
	titleText := fmt.Sprintf(" %s%s [%s] ", prefixDot, categoryTitle, filterName)
	title := StylePurple.Render(titleText)
	hints := StyleFaint.Render(" [f] filter  [*] pin ")

	lines := []string{
		alignTwoItems(title, hints, width),
		PadToWidth(StyleFaint.Render(strings.Repeat("─", width)), width),
	}

	availableRows := height - len(lines)
	if availableRows < 1 {
		availableRows = 1
	}

	scrollOffset := 0
	if selectedIndex >= availableRows {
		scrollOffset = selectedIndex - availableRows + 1
	}

	for i := 0; i < availableRows; i++ {
		idx := scrollOffset + i
		if idx >= len(playlists) {
			lines = append(lines, PadToWidth("", width))
			continue
		}

		pl := playlists[idx]
		isSelected := idx == selectedIndex

		prefix := "  "
		if isSelected {
			prefix = "❯ "
		}

		// Bold crisp gold star for pinned playlists
		starPrefix := ""
		if pinnedURIs != nil && pinnedURIs[pl.URI] {
			starPrefix = "★ "
		}

		countStr := ""
		displayName := pl.Name
		if filter == FilterAlbums {
			if pl.OwnerID != "" && width >= 34 {
				displayName = fmt.Sprintf("%s ─ %s", pl.Name, pl.OwnerID)
			}
			countStr = fmt.Sprintf(" (%d)", pl.TrackCount)
		} else if filter == FilterArtists {
			countStr = ""
		} else if pl.TrackCount > 0 {
			countStr = fmt.Sprintf(" (%d)", pl.TrackCount)
		}

		availNameW := width - ansi.StringWidth(prefix) - ansi.StringWidth(starPrefix) - ansi.StringWidth(countStr) - 1
		if availNameW < 4 {
			availNameW = 4
		}

		truncName := TruncateString(displayName, availNameW)
		plainLine := prefix + starPrefix + truncName + countStr

		var lineContent string
		if isSelected && focused {
			lineContent = RenderPaddedLine(plainLine, StyleActiveFocusedBlock, width)
		} else if isSelected {
			lineContent = RenderPaddedLine(plainLine, StyleActiveUnfocusedBlock, width)
		} else if starPrefix != "" {
			starStyled := lipgloss.NewStyle().Foreground(CurrentTheme.Yellow).Bold(true).Render(prefix + starPrefix)
			nameStyled := StyleNormal.Render(truncName)
			cntStyled := StyleFaint.Render(countStr)
			combined := starStyled + nameStyled + cntStyled
			remPad := width - ansi.StringWidth(combined)
			if remPad < 0 {
				remPad = 0
			}
			lineContent = PadToWidth(combined+strings.Repeat(" ", remPad), width)
		} else {
			nameStyled := StyleNormal.Render(prefix + truncName)
			cntStyled := StyleFaint.Render(countStr)
			combined := nameStyled + cntStyled
			remPad := width - ansi.StringWidth(combined)
			if remPad < 0 {
				remPad = 0
			}
			lineContent = PadToWidth(combined+strings.Repeat(" ", remPad), width)
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

// RenderNav renders the left sidebar inside a border that highlights when focused
func RenderNav(playlists []backend.Playlist, pinnedURIs map[string]bool, filter PlaylistFilter, selectedIndex int, focused bool, width, height int) string {
	contentW := width - 2
	if contentW < 10 {
		contentW = 10
	}
	contentH := height - 2
	if contentH < 3 {
		contentH = 3
	}

	lines := RenderNavLines(playlists, pinnedURIs, filter, selectedIndex, focused, contentW, contentH)

	boxStyle := lipgloss.NewStyle().Width(width).Height(height)
	if focused {
		boxStyle = boxStyle.Border(lipgloss.ThickBorder()).BorderForeground(CurrentTheme.Purple).Bold(true)
	} else {
		boxStyle = boxStyle.Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Overlay)
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func alignTwoItems(left, right string, totalW int) string {
	lW := ansi.StringWidth(left)
	rW := ansi.StringWidth(right)
	gap := totalW - (lW + rW)
	if gap < 1 {
		gap = 1
	}
	return PadToWidth(left+strings.Repeat(" ", gap)+right, totalW)
}
