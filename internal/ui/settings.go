package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"spotumn/internal/auth"
	"spotumn/internal/config"
)

// SettingsState tracks user navigation and status inside the Settings modal
type SettingsState struct {
	Index         int
	Status        string
	Themes        []string
	CurrentTheme  string
	Mode          string
	AutoShrink    bool
	CrossfadeSec  int
	Bitrate       int
	Normalisation bool
	Accounts      []auth.Account
	ActiveAccIdx  int
}

// NewSettingsState initializes SettingsState from active config, themes, and accounts
func NewSettingsState() SettingsState {
	cfg := config.Get()
	themes := ListAvailableThemes()
	accMgr := auth.NewAccountManager()
	return SettingsState{
		Index:         0,
		Themes:        themes,
		CurrentTheme:  cfg.Theme,
		Mode:          cfg.AppearanceMode,
		AutoShrink:    cfg.AutoShrinkSidebars,
		CrossfadeSec:  cfg.CrossfadeSec,
		Bitrate:       cfg.Bitrate,
		Normalisation: cfg.Normalisation,
		Accounts:      accMgr.GetAccounts(),
		ActiveAccIdx:  accMgr.GetActiveIndex(),
	}
}

// SettingItemCount returns total number of interactive rows in Settings modal
const SettingItemCount = 9

// ClearCache removes cached album art and state from ~/.cache/spotumn
func ClearCache() error {
	cacheDir := config.GetCacheDir()
	_ = os.RemoveAll(filepath.Join(cacheDir, "art"))
	_ = os.Remove(filepath.Join(cacheDir, "last_state.json"))
	_ = os.MkdirAll(filepath.Join(cacheDir, "art"), 0700)
	return nil
}

// RenderSettingsModal renders the settings modal (100+ chars wide, surface background)
func RenderSettingsModal(state SettingsState, width, height int) string {
	modalW := 102
	if modalW > width-4 {
		modalW = width - 4
	}
	if modalW < 44 {
		modalW = 44
	}

	var sb strings.Builder
	sb.WriteString(PadToWidth("", modalW-2) + "\n")

	// Top navigation badges
	badgeNav := StylePurple.Render("⌜") + StyleBold.Render("↑/↓") + StylePurple.Render("⌟") + BgPad(1) + StyleLavender.Render("Select")
	badgeArrows := StylePurple.Render("⌜") + StyleBold.Render("←/→") + StylePurple.Render("⌟") + BgPad(1) + StyleLavender.Render("Adjust")
	badgeEnter := StylePurple.Render("⌜") + StyleBold.Render("Enter") + StylePurple.Render("⌟") + BgPad(1) + StyleLavender.Render("Action")
	badgeClose := StylePurple.Render("⌜") + StyleBold.Render("Esc") + StylePurple.Render("⌟") + BgPad(1) + StyleLavender.Render("Close")
	tipsLine := "  " + badgeNav + BgPad(3) + badgeArrows + BgPad(3) + badgeEnter + BgPad(3) + badgeClose

	sb.WriteString(PadToWidth(tipsLine, modalW-2) + "\n")
	sb.WriteString(PadToWidth(StyleFaint.Render("  "+strings.Repeat("─", modalW-6)), modalW-2) + "\n")
	sb.WriteString(PadToWidth("", modalW-2) + "\n")

	contentW := modalW - 6
	if contentW < 30 {
		contentW = 30
	}

	rows := []struct {
		title string
		val   string
		desc  string
	}{
		{
			title: "Theme",
			val:   fmt.Sprintf("◀  %s  ▶", state.CurrentTheme),
			desc:  "~/.config/spotumn/themes/ (*.json)",
		},
		{
			title: "Appearance",
			val:   fmt.Sprintf("◀  %s  ▶", strings.ToUpper(state.Mode)),
			desc:  "Dark, Light, or Auto terminal detection",
		},
		{
			title: "Auto-Shrink Panels",
			val:   func() string { if state.AutoShrink { return "[ Enabled ]" } else { return "[ Disabled ]" } }(),
			desc:  "Auto-collapse sidebars when terminal width is tight",
		},
		{
			title: "Crossfade",
			val: func() string {
				if state.CrossfadeSec == 0 {
					return "◀  Off  ▶"
				}
				return fmt.Sprintf("◀  %ds  ▶", state.CrossfadeSec)
			}(),
			desc: "Track overlap duration (0-12s)",
		},
		{
			title: "Audio Quality",
			val: func() string {
				switch state.Bitrate {
				case 96:
					return "◀  Normal (96k)  ▶"
				case 160:
					return "◀  High (160k)  ▶"
				default:
					return "◀  Very High (320k)  ▶"
				}
			}(),
			desc: "Streaming bitrate (96, 160, 320 kbps)",
		},
		{
			title: "Volume Normalisation",
			val: func() string {
				if state.Normalisation {
					return "[ Enabled ]"
				}
				return "[ Disabled ]"
			}(),
			desc: "Equalize loudness across tracks",
		},
		{
			title: "Cache Management",
			val:   "[ Clear Cache ]",
			desc:  "Purge album art & cached state (~/.cache/spotumn/)",
		},
		{
			title: "Spotify Account",
			val: func() string {
				if len(state.Accounts) > 1 && state.ActiveAccIdx >= 0 && state.ActiveAccIdx < len(state.Accounts) {
					return fmt.Sprintf("◀ [%d/%d] %s ▶", state.ActiveAccIdx+1, len(state.Accounts), state.Accounts[state.ActiveAccIdx].DisplayName)
				} else if len(state.Accounts) == 1 {
					return fmt.Sprintf("[%s] + Add", state.Accounts[0].DisplayName)
				}
				return "[ + Add Account ]"
			}(),
			desc: func() string {
				if len(state.Accounts) > 1 {
					return "←/→ switch • Enter: add new account (max 4)"
				}
				return "Enter: add another account via browser"
			}(),
		},
		{
			title: "Session Control",
			val:   "[ Log Out Device ]",
			desc:  "Clear current active session credentials",
		},
	}

	for i, r := range rows {
		isSelected := i == state.Index
		prefix := "  "
		if isSelected {
			prefix = "❯ "
		}

		titleCol := PadPlain(r.title, 24)
		valCol := PadPlain(r.val, 28)
		descCol := TruncateString(r.desc, contentW-54)

		if isSelected {
			rawLine := prefix + titleCol + valCol + descCol
			styled := RenderPaddedLine(rawLine, StyleActiveFocusedBlock, contentW)
			sb.WriteString(PadToWidth("  "+styled, modalW-2) + "\n")
		} else {
			styledTitle := StyleBold.Render(titleCol)
			styledVal := StyleMint.Render(valCol)
			styledDesc := StyleFaint.Render(descCol)
			line := prefix + styledTitle + styledVal + styledDesc
			sb.WriteString(PadToWidth("  "+line, modalW-2) + "\n")
		}
		sb.WriteString(PadToWidth("", modalW-2) + "\n")
	}

	// Status notice banner at bottom if present
	sb.WriteString(PadToWidth(StyleFaint.Render("  "+strings.Repeat("─", modalW-6)), modalW-2) + "\n")
	if state.Status != "" {
		statusText := StyleMint.Render("  " + state.Status)
		sb.WriteString(PadToWidth(statusText, modalW-2) + "\n")
	} else {
		hintText := StyleFaint.Render("  Themes: ~/.config/spotumn/themes/  •  Cache: ~/.cache/spotumn/")
		sb.WriteString(PadToWidth(hintText, modalW-2) + "\n")
	}

	boxStyle := PanelBox(true, modalW, 0)
	return boxStyle.Render(sb.String())
}
