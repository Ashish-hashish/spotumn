package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"spotumn/internal/backend"
)

// RenderPlayerLines generates the 3 content rows for the bottom player
func RenderPlayerLines(state *backend.PlaybackState, focused bool, width int) []string {
	if width < 30 {
		width = 30
	}

	isPlaying := false
	progressMs := 0
	durationMs := 0
	volume := 50
	shuffle := false
	repeat := "off"
	trackName := "No playback active"
	artistName := ""
	deviceName := ""

	if state != nil {
		isPlaying = state.Playing
		progressMs = state.ProgressMs
		durationMs = state.DurationMs
		volume = state.Volume
		shuffle = state.Shuffle
		repeat = state.Repeat
		if state.DeviceName != "" && !strings.EqualFold(state.DeviceName, "spotumn") {
			deviceName = state.DeviceName
		}
		if state.CurrentTrack != nil {
			trackName = state.CurrentTrack.Name
			artistName = state.CurrentTrack.Artist
			if durationMs == 0 {
				durationMs = state.CurrentTrack.DurationMs
			}
		}
	}

	var lines []string

	// Row 1: Track Info (Left) | Controls & Modes (Center) | Volume Slider & Device (Right)
	leftText := " ♫ " + trackName
	if artistName != "" {
		leftText += " - " + artistName
	}
	leftStyled := StyleBold.Render(TruncateString(leftText, width/3))

	playIcon := StyleMint.Render("▶")
	if isPlaying {
		playIcon = StyleMint.Render("❚❚")
	}
	shuffIcon := "off"
	if shuffle {
		shuffIcon = "on"
	}
	repIcon := "off"
	if repeat != "off" && repeat != "" {
		repIcon = repeat
	}
	ctrls := fmt.Sprintf("⏮   %s   ⏭    [S:%s] [R:%s]", playIcon, shuffIcon, repIcon)
	ctrlsStyled := StylePurple.Render(ctrls)

	volBar := renderMiniSlider(volume, 8)
	rightInfo := ""
	if deviceName != "" {
		rightInfo = fmt.Sprintf("[%s] ", TruncateString(deviceName, 12))
	}
	rightInfo += fmt.Sprintf("Vol: [%s] %2d%% ", volBar, volume)
	rightStyled := StyleFaint.Render(rightInfo)

	lines = append(lines, alignRow(leftStyled, ctrlsStyled, rightStyled, width))

	// Row 2: Full-Width Seekbar
	elapsedStr := FormatDuration(progressMs)
	totalStr := FormatDuration(durationMs)
	if durationMs == 0 {
		totalStr = "--:--"
	}

	seekW := width - len(elapsedStr) - len(totalStr) - 4
	if seekW < 10 {
		seekW = 10
	}
	seekRatio := 0.0
	if durationMs > 0 {
		seekRatio = float64(progressMs) / float64(durationMs)
		if seekRatio > 1.0 {
			seekRatio = 1.0
		}
	}
	seekBar := renderProgressBar(seekRatio, seekW)
	seekLine := fmt.Sprintf(" %s %s %s ", elapsedStr, seekBar, totalStr)
	lines = append(lines, PadToWidth(seekLine, width))

	return lines
}

// RenderPlayer renders the bottom player with volume slider inside a border that highlights when focused
func RenderPlayer(state *backend.PlaybackState, focused bool, width int) string {
	contentW := width - 2
	if contentW < 30 {
		contentW = 30
	}

	lines := RenderPlayerLines(state, focused, contentW)

	boxStyle := lipgloss.NewStyle().Width(width).Height(4)
	if focused {
		boxStyle = boxStyle.Border(lipgloss.ThickBorder()).BorderForeground(CurrentTheme.Purple).Bold(true)
	} else {
		boxStyle = boxStyle.Border(lipgloss.RoundedBorder()).BorderForeground(CurrentTheme.Overlay)
	}

	return boxStyle.Render(strings.Join(lines, "\n"))
}

func renderMiniSlider(value, width int) string {
	if width <= 0 {
		return ""
	}
	ratio := float64(value) / 100.0
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filledChars := int(ratio * float64(width))
	emptyChars := width - filledChars

	filled := lipgloss.NewStyle().Foreground(CurrentTheme.Purple).Bold(true).Render(strings.Repeat("━", filledChars))
	empty := StyleFaint.Render(strings.Repeat("─", emptyChars))

	return filled + empty
}

func renderProgressBar(ratio float64, width int) string {
	if width <= 0 {
		return ""
	}
	filledChars := int(ratio * float64(width))
	if filledChars > width {
		filledChars = width
	}
	emptyChars := width - filledChars

	filled := lipgloss.NewStyle().Foreground(CurrentTheme.Mint).Bold(true).Render(strings.Repeat("━", filledChars))
	thumb := lipgloss.NewStyle().Foreground(CurrentTheme.Mint).Bold(true).Render("●")

	if emptyChars > 0 {
		empty := StyleFaint.Render(strings.Repeat("─", emptyChars-1))
		return filled + thumb + empty
	}
	return filled
}

func alignRow(left, center, right string, totalW int) string {
	leftW := ansi.StringWidth(left)
	centerW := ansi.StringWidth(center)
	rightW := ansi.StringWidth(right)

	centerPos := (totalW - centerW) / 2
	if centerPos < leftW+1 {
		centerPos = leftW + 1
	}

	leftPad := centerPos - leftW
	if leftPad < 1 {
		leftPad = 1
	}

	rightPad := totalW - (centerPos + centerW + rightW)
	if rightPad < 1 {
		rightPad = 1
	}

	mid := left + strings.Repeat(" ", leftPad) + center + strings.Repeat(" ", rightPad) + right
	return PadToWidth(mid, totalW)
}
