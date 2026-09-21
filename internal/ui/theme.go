package ui

import (
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Dynamic theme palette supporting Dark/Light terminal background switching
type ThemePalette struct {
	Purple      color.Color // Focused borders, active tabs, titles
	Lavender    color.Color // Secondary headers, focused outlines
	Mint        color.Color // Playing items, seekbar, active lyrics
	Green       color.Color // Success / playing markers
	Peach       color.Color // Accents, warnings
	Yellow      color.Color // Pinned stars (gold)
	Text        color.Color // Primary text
	Subtext     color.Color // Artist, secondary info
	Overlay     color.Color // Muted dividers, time, hints
	BlockText   color.Color // High-contrast text inside colored highlight blocks
	UnfocusedBg color.Color // Subtle surface background for unfocused selection
}

func NewTheme(isDark bool) ThemePalette {
	if isDark {
		return ThemePalette{
			Purple:      lipgloss.Color("#cba6f7"), // Vibrant Mauve / Purple
			Lavender:    lipgloss.Color("#b4befe"), // Soft Lavender
			Mint:        lipgloss.Color("#94e2d5"), // Mint / Teal
			Green:       lipgloss.Color("#a6e3a1"), // Pastel Green
			Peach:       lipgloss.Color("#fab387"), // Warm Peach
			Yellow:      lipgloss.Color("#f9e2af"), // Golden Yellow
			Text:        lipgloss.Color("#cdd6f4"), // High-contrast Light
			Subtext:     lipgloss.Color("#a6adc8"), // Muted Subtext
			Overlay:     lipgloss.Color("#6c7086"), // Faint Overlay
			BlockText:   lipgloss.Color("#11111b"), // Deep dark text for contrast
			UnfocusedBg: lipgloss.Color("#313244"), // Subtle surface block
		}
	}
	return ThemePalette{
		Purple:      lipgloss.Color("#8839ef"), // Deep Royal Purple
		Lavender:    lipgloss.Color("#7287fd"), // Deep Lavender
		Mint:        lipgloss.Color("#179299"), // Deep Teal / Mint
		Green:       lipgloss.Color("#40a02b"), // Forest Green
		Peach:       lipgloss.Color("#fe640b"), // Vivid Orange/Peach
		Yellow:      lipgloss.Color("#df8e1d"), // Deep Gold
		Text:        lipgloss.Color("#4c4f69"), // High-contrast Dark
		Subtext:     lipgloss.Color("#6c6f85"), // Slate Subtext
		Overlay:     lipgloss.Color("#9ca0b0"), // Soft Gray
		BlockText:   lipgloss.Color("#ffffff"), // Pure white text for contrast
		UnfocusedBg: lipgloss.Color("#ccd0da"), // Subtle light surface block
	}
}

var (
	CurrentTheme ThemePalette

	StyleNormal               lipgloss.Style
	StyleFaint                lipgloss.Style
	StyleBold                 lipgloss.Style
	StylePurple               lipgloss.Style
	StyleLavender             lipgloss.Style
	StyleMint                 lipgloss.Style
	StylePeach                lipgloss.Style
	StylePlaying              lipgloss.Style
	StyleActiveFocused        lipgloss.Style
	StyleActiveUnfocused      lipgloss.Style
	StyleActiveFocusedBlock   lipgloss.Style
	StyleActiveUnfocusedBlock lipgloss.Style
	StyleActiveLyricsBlock    lipgloss.Style
)

func init() {
	isDark := true
	if termIn, err := os.Open("/dev/tty"); err == nil {
		defer termIn.Close()
		isDark = lipgloss.HasDarkBackground(termIn, termIn)
	}
	SetTheme(isDark)
}

func SetTheme(isDark bool) {
	CurrentTheme = NewTheme(isDark)
	updateStyles()
}

func updateStyles() {
	StyleNormal = lipgloss.NewStyle().Foreground(CurrentTheme.Text)
	StyleFaint = lipgloss.NewStyle().Foreground(CurrentTheme.Overlay)
	StyleBold = lipgloss.NewStyle().Bold(true)
	StylePurple = lipgloss.NewStyle().Foreground(CurrentTheme.Purple).Bold(true)
	StyleLavender = lipgloss.NewStyle().Foreground(CurrentTheme.Lavender).Bold(true)
	StyleMint = lipgloss.NewStyle().Foreground(CurrentTheme.Mint).Bold(true)
	StylePeach = lipgloss.NewStyle().Foreground(CurrentTheme.Peach)
	StylePlaying = lipgloss.NewStyle().Foreground(CurrentTheme.Mint).Bold(true)
	StyleActiveFocused = lipgloss.NewStyle().Foreground(CurrentTheme.Purple).Bold(true)
	StyleActiveUnfocused = lipgloss.NewStyle().Foreground(CurrentTheme.Lavender).Bold(true)

	// Colored block selection styles with contrasting text
	StyleActiveFocusedBlock = lipgloss.NewStyle().
		Background(CurrentTheme.Purple).
		Foreground(CurrentTheme.BlockText).
		Bold(true)

	StyleActiveUnfocusedBlock = lipgloss.NewStyle().
		Background(CurrentTheme.UnfocusedBg).
		Foreground(CurrentTheme.Text).
		Bold(true)

	StyleActiveLyricsBlock = lipgloss.NewStyle().
		Background(CurrentTheme.Mint).
		Foreground(CurrentTheme.BlockText).
		Bold(true)
}

// RenderPaddedLine renders a row to exact width with the given style
func RenderPaddedLine(rawText string, style lipgloss.Style, width int) string {
	if width <= 0 {
		return ""
	}
	trunc := TruncateString(rawText, width)
	w := ansi.StringWidth(trunc)
	pad := width - w
	if pad < 0 {
		pad = 0
	}
	return style.Render(trunc + strings.Repeat(" ", pad))
}

// PadToWidth fills string with transparent terminal spaces to exact width
func PadToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := ansi.StringWidth(s)
	if w > width {
		s = ansi.Truncate(s, width, "…")
		w = ansi.StringWidth(s)
	}
	pad := width - w
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
}

// TruncateString truncates a string to max width with ellipsis
func TruncateString(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= maxW {
		return s
	}
	return ansi.Truncate(s, maxW, "…")
}

// FormatDuration converts milliseconds to "mm:ss"
func FormatDuration(ms int) string {
	totalSec := ms / 1000
	min := totalSec / 60
	sec := totalSec % 60
	return formatTwoDigits(min) + ":" + formatTwoDigits(sec)
}

func formatTwoDigits(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+(n/10))) + string(rune('0'+(n%10)))
}
