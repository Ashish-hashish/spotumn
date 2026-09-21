package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/zmb3/spotify/v2"
)

// RenderDevicesModal renders an interactive Spotify Connect devices selection window
func RenderDevicesModal(devices []spotify.PlayerDevice, selectedIdx int, isScanning bool, width, height int) string {
	modalW := 56
	if modalW > width-4 {
		modalW = width - 4
	}
	if modalW < 36 {
		modalW = 36
	}

	var sb strings.Builder
	title := StylePurple.Render("  Connected Devices")
	if isScanning {
		title += " " + StyleMint.Render("● Scanning...")
	}
	sb.WriteString(PadToWidth(title, modalW-2) + "\n")
	sb.WriteString(PadToWidth(StyleFaint.Render(strings.Repeat("─", modalW-2)), modalW-2) + "\n\n")

	if len(devices) == 0 {
		emptyMsg := StyleFaint.Render("  No devices found. Launch Spotify or spotumn.")
		if isScanning {
			emptyMsg = StyleMint.Render("  Scanning for Spotify Connect devices...")
		}
		sb.WriteString(PadToWidth(emptyMsg, modalW-2) + "\n\n")
	} else {
		for i, d := range devices {
			if strings.EqualFold(d.Name, "spotumn") {
				if i > 0 {
					devCopy := make([]spotify.PlayerDevice, len(devices))
					copy(devCopy, devices)
					spotDev := devCopy[i]
					copy(devCopy[1:i+1], devCopy[0:i])
					devCopy[0] = spotDev
					devices = devCopy
				}
				break
			}
		}

		for i, d := range devices {
			isSelected := i == selectedIdx
			prefix := "  "
			if isSelected {
				prefix = "❯ "
			}

			activeTag := ""
			if d.Active {
				activeTag = " [Active]"
			}

			devType := d.Type
			if devType == "" {
				devType = "Speaker"
			}

			line := fmt.Sprintf("%s%s (%s)%s", prefix, d.Name, devType, activeTag)
			truncLine := TruncateString(line, modalW-4)

			if isSelected {
				sb.WriteString(RenderPaddedLine(truncLine, StyleActiveFocusedBlock, modalW-2) + "\n")
			} else if d.Active {
				sb.WriteString(PadToWidth(StyleMint.Render(truncLine), modalW-2) + "\n")
			} else {
				sb.WriteString(PadToWidth(StyleNormal.Render(truncLine), modalW-2) + "\n")
			}
		}
		sb.WriteString("\n")
	}

	footer := StyleFaint.Render("  [Enter] Select   [r] Rescan   [d/Esc] Close")
	if isScanning {
		footer = StyleFaint.Render("  [Enter] Select   ") + StyleMint.Render("[r] Scanning...") + StyleFaint.Render("   [d/Esc] Close")
	}
	sb.WriteString(PadToWidth(footer, modalW-2) + "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(CurrentTheme.Purple).
		Bold(true).
		Width(modalW)

	return boxStyle.Render(sb.String())
}
