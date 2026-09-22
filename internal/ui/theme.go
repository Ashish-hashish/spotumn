package ui

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"spotumn/internal/config"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"gopkg.in/yaml.v3"
)

// ThemeConfig is the serializable/deserializable theme definition.
// Users can override any subset in ~/.config/spotumn/theme.yml.
type ThemeConfig struct {
	MPrimary          string `yaml:"mPrimary"          json:"mPrimary"`
	MOnPrimary        string `yaml:"mOnPrimary"        json:"mOnPrimary"`
	MSecondary        string `yaml:"mSecondary"        json:"mSecondary"`
	MOnSecondary      string `yaml:"mOnSecondary"      json:"mOnSecondary"`
	MTertiary         string `yaml:"mTertiary"         json:"mTertiary"`
	MOnTertiary       string `yaml:"mOnTertiary"       json:"mOnTertiary"`
	MError            string `yaml:"mError"            json:"mError"`
	MOnError          string `yaml:"mOnError"          json:"mOnError"`
	MSurface          string `yaml:"mSurface"          json:"mSurface"`
	MOnSurface        string `yaml:"mOnSurface"        json:"mOnSurface"`
	MSurfaceVariant   string `yaml:"mSurfaceVariant"   json:"mSurfaceVariant"`
	MOnSurfaceVariant string `yaml:"mOnSurfaceVariant" json:"mOnSurfaceVariant"`
	MOutline          string `yaml:"mOutline"          json:"mOutline"`
	MShadow           string `yaml:"mShadow"           json:"mShadow"`
	MHover            string `yaml:"mHover"            json:"mHover"`
	MOnHover          string `yaml:"mOnHover"          json:"mOnHover"`
	MGold             string `yaml:"mGold"             json:"mGold"`
	MSuccess          string `yaml:"mSuccess"          json:"mSuccess"`
}

// ThemePalette holds resolved color.Color values for the active theme.
type ThemePalette struct {
	Primary          color.Color // Focused borders, active tabs, titles
	OnPrimary        color.Color // Text on primary-colored blocks
	Secondary        color.Color // Secondary headers, lavender accents
	OnSecondary      color.Color // Text on secondary blocks
	Tertiary         color.Color // Playing items, seekbar, active lyrics
	OnTertiary       color.Color // Text on tertiary blocks
	Error            color.Color // Error indicators
	OnError          color.Color // Text on error blocks
	Surface          color.Color // Main background fill
	OnSurface        color.Color // Primary text
	SurfaceVariant   color.Color // Unfocused selection bg
	OnSurfaceVariant color.Color // Subtext, muted text
	Outline          color.Color // Unfocused borders, dividers, faint elements
	Shadow           color.Color // Deep contrast text for colored blocks
	Hover            color.Color // Warm accent (peach) for album descriptions
	OnHover          color.Color // Text on hover blocks
	Gold             color.Color // Pinned star color
	Success          color.Color // Green dot, success indicators
}

// catppuccinMocha returns the built-in Catppuccin Mocha dark theme config.
func catppuccinMocha() ThemeConfig {
	return ThemeConfig{
		MPrimary:          "#cba6f7", // Mauve
		MOnPrimary:        "#11111b", // Crust
		MSecondary:        "#b4befe", // Lavender
		MOnSecondary:      "#11111b",
		MTertiary:         "#94e2d5", // Teal
		MOnTertiary:       "#11111b",
		MError:            "#f38ba8", // Red
		MOnError:          "#11111b",
		MSurface:          "#1e1e2e", // Base
		MOnSurface:        "#cdd6f4", // Text
		MSurfaceVariant:   "#313244", // Surface0
		MOnSurfaceVariant: "#a6adc8", // Subtext0
		MOutline:          "#6c7086", // Overlay0
		MShadow:           "#11111b", // Crust
		MHover:            "#fab387", // Peach
		MOnHover:          "#11111b",
		MGold:             "#f9e2af", // Yellow
		MSuccess:          "#a6e3a1", // Green
	}
}

// catppuccinLatte returns the built-in Catppuccin Latte light theme config.
func catppuccinLatte() ThemeConfig {
	return ThemeConfig{
		MPrimary:          "#8839ef", // Mauve
		MOnPrimary:        "#ffffff",
		MSecondary:        "#7287fd", // Lavender
		MOnSecondary:      "#ffffff",
		MTertiary:         "#179299", // Teal
		MOnTertiary:       "#ffffff",
		MError:            "#d20f39", // Red
		MOnError:          "#ffffff",
		MSurface:          "#eff1f5", // Base
		MOnSurface:        "#4c4f69", // Text
		MSurfaceVariant:   "#ccd0da", // Surface0
		MOnSurfaceVariant: "#6c6f85", // Subtext0
		MOutline:          "#9ca0b0", // Overlay0
		MShadow:           "#dce0e8", // Crust
		MHover:            "#fe640b", // Peach
		MOnHover:          "#ffffff",
		MGold:             "#df8e1d", // Yellow
		MSuccess:          "#40a02b", // Green
	}
}

// resolveConfig converts a ThemeConfig (hex strings) into a ThemePalette (resolved colors).
func resolveConfig(cfg ThemeConfig) ThemePalette {
	return ThemePalette{
		Primary:          lipgloss.Color(cfg.MPrimary),
		OnPrimary:        lipgloss.Color(cfg.MOnPrimary),
		Secondary:        lipgloss.Color(cfg.MSecondary),
		OnSecondary:      lipgloss.Color(cfg.MOnSecondary),
		Tertiary:         lipgloss.Color(cfg.MTertiary),
		OnTertiary:       lipgloss.Color(cfg.MOnTertiary),
		Error:            lipgloss.Color(cfg.MError),
		OnError:          lipgloss.Color(cfg.MOnError),
		Surface:          lipgloss.Color(cfg.MSurface),
		OnSurface:        lipgloss.Color(cfg.MOnSurface),
		SurfaceVariant:   lipgloss.Color(cfg.MSurfaceVariant),
		OnSurfaceVariant: lipgloss.Color(cfg.MOnSurfaceVariant),
		Outline:          lipgloss.Color(cfg.MOutline),
		Shadow:           lipgloss.Color(cfg.MShadow),
		Hover:            lipgloss.Color(cfg.MHover),
		OnHover:          lipgloss.Color(cfg.MOnHover),
		Gold:             lipgloss.Color(cfg.MGold),
		Success:          lipgloss.Color(cfg.MSuccess),
	}
}

// mergeConfig applies non-empty overrides from custom onto base.
func mergeConfig(base, custom ThemeConfig) ThemeConfig {
	if custom.MPrimary != "" {
		base.MPrimary = custom.MPrimary
	}
	if custom.MOnPrimary != "" {
		base.MOnPrimary = custom.MOnPrimary
	}
	if custom.MSecondary != "" {
		base.MSecondary = custom.MSecondary
	}
	if custom.MOnSecondary != "" {
		base.MOnSecondary = custom.MOnSecondary
	}
	if custom.MTertiary != "" {
		base.MTertiary = custom.MTertiary
	}
	if custom.MOnTertiary != "" {
		base.MOnTertiary = custom.MOnTertiary
	}
	if custom.MError != "" {
		base.MError = custom.MError
	}
	if custom.MOnError != "" {
		base.MOnError = custom.MOnError
	}
	if custom.MSurface != "" {
		base.MSurface = custom.MSurface
	}
	if custom.MOnSurface != "" {
		base.MOnSurface = custom.MOnSurface
	}
	if custom.MSurfaceVariant != "" {
		base.MSurfaceVariant = custom.MSurfaceVariant
	}
	if custom.MOnSurfaceVariant != "" {
		base.MOnSurfaceVariant = custom.MOnSurfaceVariant
	}
	if custom.MOutline != "" {
		base.MOutline = custom.MOutline
	}
	if custom.MShadow != "" {
		base.MShadow = custom.MShadow
	}
	if custom.MHover != "" {
		base.MHover = custom.MHover
	}
	if custom.MOnHover != "" {
		base.MOnHover = custom.MOnHover
	}
	if custom.MGold != "" {
		base.MGold = custom.MGold
	}
	if custom.MSuccess != "" {
		base.MSuccess = custom.MSuccess
	}
	return base
}

// LoadThemeFile loads a custom theme YAML from custom path or ~/.config/spotumn/theme.yml.
// Returns zero ThemeConfig if file doesn't exist or can't be parsed.
func LoadThemeFile(customPath ...string) ThemeConfig {
	var themePath string
	if len(customPath) > 0 && customPath[0] != "" {
		themePath = customPath[0]
	} else {
		themePath = filepath.Join(config.GetDir(), "theme.yml")
	}
	data, err := os.ReadFile(themePath)
	if err != nil {
		return ThemeConfig{}
	}
	var custom ThemeConfig
	_ = yaml.Unmarshal(data, &custom)
	return custom
}

var (
	CurrentTheme ThemePalette

	// StyleBase carries the surface background — compose all other styles with it
	StyleBase lipgloss.Style

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

// SetTheme selects the built-in dark/light palette, merges any custom overrides, and rebuilds styles.
func SetTheme(isDark bool) {
	base := catppuccinMocha()
	if !isDark {
		base = catppuccinLatte()
	}
	custom := LoadThemeFile()
	merged := mergeConfig(base, custom)
	CurrentTheme = resolveConfig(merged)
	updateStyles()
}

// SetThemeFromConfig sets the theme with a specific ThemeConfig (for testing).
func SetThemeFromConfig(cfg ThemeConfig) {
	CurrentTheme = resolveConfig(cfg)
	updateStyles()
}

func updateStyles() {
	// Base style: surface background applied to all content
	StyleBase = lipgloss.NewStyle().Background(CurrentTheme.Surface)

	StyleNormal = StyleBase.Foreground(CurrentTheme.OnSurface)
	StyleFaint = StyleBase.Foreground(CurrentTheme.Outline)
	StyleBold = StyleBase.Bold(true).Foreground(CurrentTheme.OnSurface)
	StylePurple = StyleBase.Foreground(CurrentTheme.Primary).Bold(true)
	StyleLavender = StyleBase.Foreground(CurrentTheme.Secondary).Bold(true)
	StyleMint = StyleBase.Foreground(CurrentTheme.Tertiary).Bold(true)
	StylePeach = StyleBase.Foreground(CurrentTheme.Hover)
	StylePlaying = StyleBase.Foreground(CurrentTheme.Tertiary).Bold(true)
	StyleActiveFocused = StyleBase.Foreground(CurrentTheme.Primary).Bold(true)
	StyleActiveUnfocused = StyleBase.Foreground(CurrentTheme.Secondary).Bold(true)

	// Colored block selection styles with contrasting text
	StyleActiveFocusedBlock = lipgloss.NewStyle().
		Background(CurrentTheme.Primary).
		Foreground(CurrentTheme.OnPrimary).
		Bold(true)

	StyleActiveUnfocusedBlock = lipgloss.NewStyle().
		Background(CurrentTheme.SurfaceVariant).
		Foreground(CurrentTheme.OnSurface).
		Bold(true)

	StyleActiveLyricsBlock = lipgloss.NewStyle().
		Background(CurrentTheme.Tertiary).
		Foreground(CurrentTheme.OnTertiary).
		Bold(true)
}

// BgPad returns a string of spaces with the surface background applied.
func BgPad(n int) string {
	if n <= 0 {
		return ""
	}
	return StyleBase.Render(strings.Repeat(" ", n))
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

// PanelBox returns a standardized container border style with full surface background.
func PanelBox(focused bool, width, height int) lipgloss.Style {
	st := lipgloss.NewStyle().
		Width(width).
		Background(CurrentTheme.Surface)
	if height > 0 {
		st = st.Height(height)
	}
	if focused {
		return st.Border(lipgloss.ThickBorder()).
			BorderForeground(CurrentTheme.Primary).
			BorderBackground(CurrentTheme.Surface).
			Bold(true)
	}
	return st.Border(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Outline).
		BorderBackground(CurrentTheme.Surface)
}

// PadPlain truncates or pads plain text with unstyled spaces to exact width.
// Use this for building plain text rows or table columns before passing to lipgloss.Style.Render.
func PadPlain(s string, width int) string {
	if width <= 0 {
		return ""
	}
	s = TruncateString(s, width)
	w := ansi.StringWidth(s)
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

// CenterLine centers a string within the given width, padding with surface-colored spaces on both sides.
func CenterLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if s != "" && !strings.Contains(s, "\x1b[") {
		s = StyleNormal.Render(s)
	}
	w := ansi.StringWidth(s)
	if w >= width {
		return TruncateString(s, width)
	}
	padLeft := (width - w) / 2
	padRight := width - w - padLeft
	return BgPad(padLeft) + s + BgPad(padRight)
}

// CenterOverlay centers a modal dialog block on screen and ensures the entire surrounding backdrop
// is painted with the surface background color (no naked terminal spaces).
func CenterOverlay(modal string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(modal, "\n")
	topPad := (height - len(lines)) / 2
	if topPad < 0 {
		topPad = 0
	}
	emptyRow := BgPad(width)
	var rows []string
	for i := 0; i < topPad; i++ {
		rows = append(rows, emptyRow)
	}
	for _, line := range lines {
		rows = append(rows, CenterLine(line, width))
	}
	for len(rows) < height {
		rows = append(rows, emptyRow)
	}
	if len(rows) > height {
		rows = rows[:height]
	}
	return strings.Join(rows, "\n")
}

// PadToWidth fills string with background-colored spaces to exact width.
// If s is plain text (contains no ANSI escapes), it styles it with StyleNormal so the entire row has surface background.
func PadToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if s != "" && !strings.Contains(s, "\x1b[") {
		s = StyleNormal.Render(s)
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
	return s + BgPad(pad)
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
