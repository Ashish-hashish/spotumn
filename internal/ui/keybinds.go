package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"spotumn/internal/config"
)

const (
	ActionPlayPause     = "play_pause"
	ActionPrevTrack     = "prev"
	ActionNextTrack     = "next"
	ActionSeekBack      = "seek_back"
	ActionSeekFwd       = "seek_fwd"
	ActionSeekBackBig   = "seek_back_big"
	ActionSeekFwdBig    = "seek_fwd_big"
	ActionVolumeUp      = "vol_up"
	ActionVolumeDown    = "vol_down"
	ActionVolumeUpBig   = "vol_up_big"
	ActionVolumeDownBig = "vol_down_big"
	ActionQueueTrack    = "queue"
	ActionDevices       = "devices"
	ActionShuffle       = "shuffle"
	ActionRepeat        = "repeat"
	ActionFocusPrev     = "focus_prev"
	ActionFocusNext     = "focus_next"
	ActionCursorUp      = "cursor_up"
	ActionCursorDown    = "cursor_down"
	ActionSelect        = "select"
	ActionOpenArtist    = "open_artist"
	ActionBack          = "back"
	ActionTabTracks     = "tab_tracks"
	ActionTabLyrics     = "tab_lyrics"
	ActionTabHistory    = "tab_history"
	ActionSearch        = "search"
	ActionFilter        = "filter"
	ActionPin           = "pin"
	ActionToggleSidebar = "sidebar"
	ActionZenMode       = "zen_mode"
	ActionZenLayout     = "zen_layout"
	ActionHelp          = "help"
	ActionEscape        = "escape"
	ActionQuit          = "quit"
)

type KeybindItem struct {
	ID          string   `json:"id"`
	Category    string   `json:"category"`
	Key         string   `json:"key"`
	Keys        []string `json:"keys"`
	Desc        string   `json:"desc"`
	DefaultKey  string   `json:"default_key"`
	DefaultKeys []string `json:"default_keys"`
}

type KeybindCategory struct {
	Title string
	Items []KeybindItem
}

func GetDefaultKeybindItems() []KeybindItem {
	return []KeybindItem{
		// Playback Controls
		{ID: ActionPlayPause, Category: "Playback Controls", Key: "Space", Keys: []string{"space"}, Desc: "Play / Pause playback", DefaultKey: "Space", DefaultKeys: []string{"space"}},
		{ID: ActionPrevTrack, Category: "Playback Controls", Key: "p", Keys: []string{"p"}, Desc: "Previous track", DefaultKey: "p", DefaultKeys: []string{"p"}},
		{ID: ActionNextTrack, Category: "Playback Controls", Key: "n", Keys: []string{"n"}, Desc: "Next track", DefaultKey: "n", DefaultKeys: []string{"n"}},
		{ID: ActionSeekBack, Category: "Playback Controls", Key: "←", Keys: []string{"left", "<", ","}, Desc: "Seek backward 5s (or Shift+← for 10s)", DefaultKey: "←", DefaultKeys: []string{"left", "<", ","}},
		{ID: ActionSeekFwd, Category: "Playback Controls", Key: "→", Keys: []string{"right", ">", "."}, Desc: "Seek forward 5s (or Shift+→ for 10s)", DefaultKey: "→", DefaultKeys: []string{"right", ">", "."}},
		{ID: ActionSeekBackBig, Category: "Playback Controls", Key: "Shift+←", Keys: []string{"shift+left"}, Desc: "Seek backward 10s", DefaultKey: "Shift+←", DefaultKeys: []string{"shift+left"}},
		{ID: ActionSeekFwdBig, Category: "Playback Controls", Key: "Shift+→", Keys: []string{"shift+right"}, Desc: "Seek forward 10s", DefaultKey: "Shift+→", DefaultKeys: []string{"shift+right"}},
		{ID: ActionVolumeUp, Category: "Playback Controls", Key: "=", Keys: []string{"="}, Desc: "Volume up 5 points (or + for 10)", DefaultKey: "=", DefaultKeys: []string{"="}},
		{ID: ActionVolumeDown, Category: "Playback Controls", Key: "-", Keys: []string{"-"}, Desc: "Volume down 5 points (or _ for 10)", DefaultKey: "-", DefaultKeys: []string{"-"}},
		{ID: ActionVolumeUpBig, Category: "Playback Controls", Key: "+", Keys: []string{"+", "shift+=", "shift++"}, Desc: "Volume up 10 points (+ or Shift+=)", DefaultKey: "+", DefaultKeys: []string{"+", "shift+=", "shift++"}},
		{ID: ActionVolumeDownBig, Category: "Playback Controls", Key: "_", Keys: []string{"_", "shift+-", "shift+_"}, Desc: "Volume down 10 points (_ or Shift+-)", DefaultKey: "_", DefaultKeys: []string{"_", "shift+-", "shift+_"}},
		{ID: ActionQueueTrack, Category: "Playback Controls", Key: "q", Keys: []string{"q"}, Desc: "Add highlighted song to Spotify queue", DefaultKey: "q", DefaultKeys: []string{"q"}},
		{ID: ActionDevices, Category: "Playback Controls", Key: "d", Keys: []string{"d"}, Desc: "Open Connect devices popup (r to rescan)", DefaultKey: "d", DefaultKeys: []string{"d"}},
		{ID: ActionShuffle, Category: "Playback Controls", Key: "s", Keys: []string{"s"}, Desc: "Toggle shuffle mode (󰒝 on / 󰒞 off)", DefaultKey: "s", DefaultKeys: []string{"s"}},
		{ID: ActionRepeat, Category: "Playback Controls", Key: "r", Keys: []string{"r"}, Desc: "Cycle repeat mode (󰑗 off / 󰑖 all / 󰑘 once)", DefaultKey: "r", DefaultKeys: []string{"r"}},

		// Navigation & Tabs
		{ID: ActionFocusPrev, Category: "Navigation & Tabs", Key: "[", Keys: []string{"["}, Desc: "Cycle pane focus backward", DefaultKey: "[", DefaultKeys: []string{"["}},
		{ID: ActionFocusNext, Category: "Navigation & Tabs", Key: "]", Keys: []string{"]"}, Desc: "Cycle pane focus forward", DefaultKey: "]", DefaultKeys: []string{"]"}},
		{ID: ActionCursorUp, Category: "Navigation & Tabs", Key: "k", Keys: []string{"up", "k"}, Desc: "Move cursor / navigate up (or ↑)", DefaultKey: "k", DefaultKeys: []string{"up", "k"}},
		{ID: ActionCursorDown, Category: "Navigation & Tabs", Key: "j", Keys: []string{"down", "j"}, Desc: "Move cursor / navigate down (or ↓)", DefaultKey: "j", DefaultKeys: []string{"down", "j"}},
		{ID: ActionSelect, Category: "Navigation & Tabs", Key: "Enter", Keys: []string{"enter"}, Desc: "Play track, open album, or seek lyrics", DefaultKey: "Enter", DefaultKeys: []string{"enter"}},
		{ID: ActionOpenArtist, Category: "Navigation & Tabs", Key: "a", Keys: []string{"a"}, Desc: "Open artist page of currently playing track", DefaultKey: "a", DefaultKeys: []string{"a"}},
		{ID: ActionBack, Category: "Navigation & Tabs", Key: "b", Keys: []string{"backspace", "b"}, Desc: "Go back to previous artist / container (or Backspace)", DefaultKey: "b", DefaultKeys: []string{"backspace", "b"}},
		{ID: ActionTabTracks, Category: "Navigation & Tabs", Key: "1", Keys: []string{"1"}, Desc: "Switch tab: 1:Tracks", DefaultKey: "1", DefaultKeys: []string{"1"}},
		{ID: ActionTabLyrics, Category: "Navigation & Tabs", Key: "2", Keys: []string{"2"}, Desc: "Switch tab: 2:Lyrics", DefaultKey: "2", DefaultKeys: []string{"2"}},
		{ID: ActionTabHistory, Category: "Navigation & Tabs", Key: "3", Keys: []string{"3"}, Desc: "Switch tab: 3:History", DefaultKey: "3", DefaultKeys: []string{"3"}},
		{ID: ActionSearch, Category: "Navigation & Tabs", Key: "/", Keys: []string{"/"}, Desc: "Focus search bar (Esc to exit)", DefaultKey: "/", DefaultKeys: []string{"/"}},
		{ID: ActionFilter, Category: "Navigation & Tabs", Key: "f", Keys: []string{"f", "t"}, Desc: "Cycle library filters", DefaultKey: "f", DefaultKeys: []string{"f", "t"}},
		{ID: ActionPin, Category: "Navigation & Tabs", Key: "*", Keys: []string{"*"}, Desc: "Pin / unpin highlighted playlist (★ prefix)", DefaultKey: "*", DefaultKeys: []string{"*"}},

		// Layout & System
		{ID: ActionToggleSidebar, Category: "Layout & System", Key: "Shift+H", Keys: []string{"shift+h", "H"}, Desc: "Hide focused sidebar (restores both if hidden)", DefaultKey: "Shift+H", DefaultKeys: []string{"shift+h", "H"}},
		{ID: ActionZenMode, Category: "Layout & System", Key: "z", Keys: []string{"z"}, Desc: "Toggle Zen mode (fullscreen art & lyrics)", DefaultKey: "z", DefaultKeys: []string{"z"}},
		{ID: ActionZenLayout, Category: "Layout & System", Key: "v", Keys: []string{"v"}, Desc: "Cycle Zen layout (Art+Lyrics / Art / Lyrics)", DefaultKey: "v", DefaultKeys: []string{"v"}},
		{ID: ActionHelp, Category: "Layout & System", Key: "?", Keys: []string{"?"}, Desc: "Open / Close Keybindings window", DefaultKey: "?", DefaultKeys: []string{"?"}},
		{ID: ActionEscape, Category: "Layout & System", Key: "Esc", Keys: []string{"esc"}, Desc: "Close modal / reset lyrics scroll / exit search", DefaultKey: "Esc", DefaultKeys: []string{"esc"}},
		{ID: ActionQuit, Category: "Layout & System", Key: "Ctrl+C", Keys: []string{"ctrl+c"}, Desc: "Quit spotumn", DefaultKey: "Ctrl+C", DefaultKeys: []string{"ctrl+c"}},
	}
}

type KeyManager struct {
	Items []KeybindItem
	file  string
}

func NewKeyManager() *KeyManager {
	km := &KeyManager{
		Items: GetDefaultKeybindItems(),
		file:  filepath.Join(config.GetDir(), "keybinds.json"),
	}
	km.Load()
	return km
}

func (km *KeyManager) Load() {
	data, err := os.ReadFile(km.file)
	if err != nil {
		return
	}
	var custom map[string]string
	if err := json.Unmarshal(data, &custom); err != nil {
		return
	}
	for i := range km.Items {
		if k, ok := custom[km.Items[i].ID]; ok && strings.TrimSpace(k) != "" {
			km.Items[i].Key = formatKeyDisplay(k)
			km.Items[i].Keys = []string{strings.ToLower(k)}
		}
	}
}

func (km *KeyManager) Save() error {
	custom := make(map[string]string)
	for _, it := range km.Items {
		if it.Key != it.DefaultKey {
			custom[it.ID] = it.Key
		}
	}
	data, err := json.MarshalIndent(custom, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(km.file, data, 0600)
}

func (km *KeyManager) SetKey(idx int, key string) {
	if idx < 0 || idx >= len(km.Items) {
		return
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	km.Items[idx].Key = formatKeyDisplay(key)
	km.Items[idx].Keys = []string{strings.ToLower(key)}
}

func (km *KeyManager) ResetItem(idx int) {
	if idx >= 0 && idx < len(km.Items) {
		km.Items[idx].Key = km.Items[idx].DefaultKey
		km.Items[idx].Keys = make([]string, len(km.Items[idx].DefaultKeys))
		copy(km.Items[idx].Keys, km.Items[idx].DefaultKeys)
	}
}

func (km *KeyManager) ResetAll() {
	for i := range km.Items {
		km.ResetItem(i)
	}
}

func (km *KeyManager) IsKeyUsed(key string) bool {
	if km == nil {
		return false
	}
	norm := strings.ToLower(strings.TrimSpace(key))
	for _, it := range km.Items {
		for _, k := range it.Keys {
			if strings.ToLower(k) == norm {
				return true
			}
		}
	}
	return false
}

func (km *KeyManager) Action(pressedKey string) string {
	if km == nil {
		return ""
	}
	norm := strings.ToLower(pressedKey)
	for _, it := range km.Items {
		for _, k := range it.Keys {
			if strings.ToLower(k) == norm {
				return it.ID
			}
		}
	}
	return ""
}

func formatKeyDisplay(key string) string {
	switch strings.ToLower(key) {
	case " ", "space":
		return "Space"
	case "enter":
		return "Enter"
	case "esc", "escape":
		return "Esc"
	case "tab":
		return "Tab"
	case "backspace":
		return "Backspace"
	case "up":
		return "↑"
	case "down":
		return "↓"
	case "left":
		return "←"
	case "right":
		return "→"
	default:
		if strings.HasPrefix(strings.ToLower(key), "ctrl+") {
			return "Ctrl+" + strings.ToUpper(key[5:])
		}
		if strings.HasPrefix(strings.ToLower(key), "alt+") {
			return "Alt+" + strings.ToUpper(key[4:])
		}
		if strings.HasPrefix(strings.ToLower(key), "shift+") {
			return "Shift+" + strings.ToUpper(key[6:])
		}
		return key
	}
}

// RenderKeybindsModal renders an interactive and navigable keybindings window
func RenderKeybindsModal(items []KeybindItem, selectedIdx int, isEditing bool, width, height int) string {
	if len(items) == 0 {
		items = GetDefaultKeybindItems()
	}

	modalW := 100
	if modalW > width-4 {
		modalW = width - 4
	}
	if modalW < 44 {
		modalW = 44
	}

	var sb strings.Builder
	sb.WriteString(PadToWidth("", modalW-2) + "\n")

	// Friendly volume and seek cooldown note at the top
	sb.WriteString(PadToWidth(StyleMint.Render("  ℹ  Tip:"), modalW-2) + "\n")
	sb.WriteString(PadToWidth(StyleNormal.Render("     Spotify limits how fast volume requests can be sent over the internet."), modalW-2) + "\n")
	sb.WriteString(PadToWidth(StyleFaint.Render("     Pressing volume/seek rapidly has a brief cooldown (~200ms) so Spotify doesn't rate-limit you."), modalW-2) + "\n")
	sb.WriteString(PadToWidth(StyleFaint.Render("  "+strings.Repeat("─", modalW-6)), modalW-2) + "\n")
	sb.WriteString(PadToWidth("", modalW-2) + "\n")

	// Calculate visible viewport height for keybind cards
	availH := height - 18
	if availH < 6 {
		availH = 6
	}
	if availH > 18 {
		availH = 18
	}

	startIdx := 0
	if selectedIdx >= availH {
		startIdx = selectedIdx - availH + 1
	}
	endIdx := startIdx + availH
	if endIdx > len(items) {
		endIdx = len(items)
		if endIdx-availH >= 0 {
			startIdx = endIdx - availH
		} else {
			startIdx = 0
		}
	}

	totalItems := len(items)
	viewH := endIdx - startIdx
	thumbH := 1
	thumbStart := 0
	if totalItems > viewH && viewH > 0 {
		thumbH = (viewH * viewH) / totalItems
		if thumbH < 1 {
			thumbH = 1
		}
		maxScroll := totalItems - viewH
		if maxScroll > 0 {
			thumbStart = (startIdx * (viewH - thumbH)) / maxScroll
		}
		if thumbStart+thumbH > viewH {
			thumbStart = viewH - thumbH
		}
		if thumbStart < 0 {
			thumbStart = 0
		}
	}

	contentW := modalW - 5
	if contentW < 30 {
		contentW = 30
	}

	for i := startIdx; i < endIdx; i++ {
		r := i - startIdx
		item := items[i]
		isSelected := i == selectedIdx

		prefix := "  "
		if isSelected {
			prefix = "❯ "
		}

		keyText := item.Key
		if isSelected && isEditing {
			keyText = "[Press key...]"
		}

		keyPill := "◖" + keyText + "◗"
		if item.Key != item.DefaultKey && !(isSelected && isEditing) {
			keyPill = "◖" + keyText + "◗*"
		}

		pillFormatted := PadPlain(keyPill, 16)
		descFormatted := TruncateString(item.Desc, contentW-20)

		// Vertical scrollbar indicator
		var scrollIndicator string
		if totalItems > viewH {
			if r >= thumbStart && r < thumbStart+thumbH {
				scrollIndicator = StylePurple.Render("█")
			} else {
				scrollIndicator = StyleFaint.Render("│")
			}
		} else {
			scrollIndicator = BgPad(1)
		}

		if isSelected {
			rawLine := prefix + pillFormatted + " " + descFormatted
			lineContent := RenderPaddedLine(rawLine, StyleActiveFocusedBlock, contentW)
			sb.WriteString(PadToWidth(lineContent+BgPad(1)+scrollIndicator, modalW-2) + "\n")
		} else {
			var styledPill string
			if item.Key != item.DefaultKey {
				styledPill = StyleMint.Render(pillFormatted)
			} else {
				styledPill = StyleLavender.Render(pillFormatted)
			}
			styledPrefix := StyleNormal.Render(prefix)
			styledLine := styledPrefix + styledPill + BgPad(1) + StyleNormal.Render(descFormatted)
			lineContent := PadToWidth(styledLine, contentW)
			sb.WriteString(PadToWidth(lineContent+BgPad(1)+scrollIndicator, modalW-2) + "\n")
		}
	}

	sb.WriteString(PadToWidth("", modalW-2) + "\n" + PadToWidth(StyleFaint.Render("  "+strings.Repeat("─", modalW-6)), modalW-2) + "\n")

	var footer string
	if isEditing {
		footer = BgPad(2) + StyleMint.Render("⌨  Press any key to assign...") + BgPad(3) + StyleFaint.Render("[Esc] Cancel")
	} else {
		footer = BgPad(2) + StyleLavender.Render("[↑/↓/j/k]") + StyleFaint.Render(" Nav   ") +
			StylePurple.Render("[Enter]") + StyleFaint.Render(" Edit   ") +
			StyleMint.Render("[0]") + StyleFaint.Render(" Reset   ") +
			StyleLavender.Render("[?/Esc]") + StyleFaint.Render(" Close")
	}
	sb.WriteString(PadToWidth(footer, modalW-2) + "\n")

	boxStyle := PanelBox(true, modalW, 0)
	return boxStyle.Render(sb.String())
}
