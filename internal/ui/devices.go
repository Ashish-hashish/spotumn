package ui

import (
	"fmt"
	"strings"

	"spotumn/internal/backend"

	"github.com/zmb3/spotify/v2"
)

// DeviceTypeGlyph returns a Nerd Font icon based on device type
func DeviceTypeGlyph(devType string) string {
	switch strings.ToLower(devType) {
	case "computer":
		return "󰍹 "
	case "smartphone", "phone":
		return "󰄡 "
	case "speaker":
		return "󰓃 "
	case "cast", "castaudio", "castvideo":
		return "󰋋 "
	default:
		return "󰓃 "
	}
}

// RenderDevicesModal renders an interactive Spotify Connect devices selection window
func RenderDevicesModal(devices []spotify.PlayerDevice, selectedIdx int, isScanning bool, width, height int) string {
	devices = backend.SortDevicesWithSpotumnFirst(devices)
	modalW := 58
	if modalW > width-4 {
		modalW = width - 4
	}
	if modalW < 36 {
		modalW = 36
	}

	var sb strings.Builder
	title := StylePurple.Render("  Connected Devices")
	if isScanning {
		title += BgPad(1) + StyleMint.Render("● Scanning...")
	}
	sb.WriteString(PadToWidth(title, modalW-2) + "\n")
	sb.WriteString(PadToWidth(StyleFaint.Render(strings.Repeat("─", modalW-2)), modalW-2) + "\n")
	sb.WriteString(PadToWidth("", modalW-2) + "\n")

	if len(devices) == 0 {
		emptyMsg := StyleFaint.Render("  No devices found. Launch Spotify or spotumn.")
		if isScanning {
			emptyMsg = StyleMint.Render("  Scanning for Spotify Connect devices...")
		}
		sb.WriteString(PadToWidth(emptyMsg, modalW-2) + "\n")
		sb.WriteString(PadToWidth("", modalW-2) + "\n")
	} else {
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
			glyph := DeviceTypeGlyph(devType)

			line := fmt.Sprintf("%s%s%s (%s)%s", prefix, glyph, d.Name, devType, activeTag)
			truncLine := TruncateString(line, modalW-4)

			if isSelected {
				sb.WriteString(RenderPaddedLine(truncLine, StyleActiveFocusedBlock, modalW-2) + "\n")
			} else if d.Active {
				sb.WriteString(PadToWidth(StyleMint.Render(truncLine), modalW-2) + "\n")
			} else {
				sb.WriteString(PadToWidth(StyleNormal.Render(truncLine), modalW-2) + "\n")
			}
		}
		sb.WriteString(PadToWidth("", modalW-2) + "\n")
	}

	footer := StyleFaint.Render("  ⌜Enter⌟ Select   ⌜r⌟ Rescan   ⌜Esc / d⌟ Close")
	if isScanning {
		footer = StyleFaint.Render("  ⌜Enter⌟ Select   ") + StyleMint.Render("⌜r⌟ Scanning...") + StyleFaint.Render("   ⌜Esc / d⌟ Close")
	}
	sb.WriteString(PadToWidth(footer, modalW-2) + "\n")

	boxStyle := PanelBox(true, modalW, 0)
	return boxStyle.Render(sb.String())
}
