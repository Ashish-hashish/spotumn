package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type KeybindItem struct {
	Key  string
	Desc string
}

type KeybindCategory struct {
	Title string
	Items []KeybindItem
}

var DefaultKeybinds = []KeybindCategory{
	{
		Title: "Playback Controls",
		Items: []KeybindItem{
			{"Space", "Play / Pause playback"},
			{"p / n", "Previous / Next track"},
			{"← / →", "Seek backward / forward 5s (or < / >)"},
			{"+ / -", "Volume up / down (or =)"},
			{"q", "Add highlighted song to Spotify queue"},
			{"d", "Open Connect devices popup (r to rescan)"},
			{"s", "Toggle shuffle mode"},
			{"r", "Cycle repeat mode (off/context/track)"},
		},
	},
	{
		Title: "Navigation & Tabs",
		Items: []KeybindItem{
			{"[ / ]", "Cycle pane focus (Nav -> Center -> Right -> Player)"},
			{"j / k / ↑ / ↓", "Move cursor / navigate items"},
			{"Enter", "Play selected track or seek to lyrics line"},
			{"1 / 2 / 3", "Switch tab: 1:Tracks, 2:History, 3:Lyrics"},
			{"/", "Focus search bar (Esc to exit)"},
			{"f / t", "Cycle playlist category (All/By You/Spotify/Saved)"},
			{"*", "Pin / unpin highlighted playlist (★ prefix)"},
		},
	},
	{
		Title: "Layout & System",
		Items: []KeybindItem{
			{"Shift+H", "Hide focused sidebar (restores both if hidden)"},
			{"z", "Toggle Zen mode (enlarged art + lyrics)"},
			{"?", "Open / close this Keybindings window"},
			{"Esc", "Close modal / reset lyrics scroll / exit search"},
			{"Ctrl+C", "Quit spotumn"},
		},
	},
}

// RenderKeybindsModal renders a centered popup window
func RenderKeybindsModal(width, height int) string {
	modalW := 68
	if modalW > width-4 {
		modalW = width - 4
	}
	if modalW < 40 {
		modalW = 40
	}

	var sb strings.Builder
	sb.WriteString("\n")

	for _, cat := range DefaultKeybinds {
		catHeader := StylePurple.Render("  " + cat.Title)
		sb.WriteString(PadToWidth(catHeader, modalW-2) + "\n")

		for _, item := range cat.Items {
			kStyle := StyleLavender.Render("    " + PadToWidth(item.Key, 16))
			dStyle := StyleNormal.Render(item.Desc)

			line := kStyle + dStyle
			sb.WriteString(PadToWidth(line, modalW-2) + "\n")
		}
		sb.WriteString(PadToWidth("", modalW-2) + "\n")
	}

	footer := StyleFaint.Render("  Press ? or Esc to close")
	sb.WriteString(PadToWidth(footer, modalW-2) + "\n")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(CurrentTheme.Purple).
		Bold(true).
		Width(modalW)

	return boxStyle.Render(sb.String())
}
