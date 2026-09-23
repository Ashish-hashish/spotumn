package ui

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"spotumn/internal/config"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// ThemeConfig is the serializable/deserializable theme palette definition for dark or light mode.
type ThemeConfig struct {
	MPrimary          string `json:"mPrimary"`
	MOnPrimary        string `json:"mOnPrimary"`
	MSecondary        string `json:"mSecondary"`
	MOnSecondary      string `json:"mOnSecondary"`
	MTertiary         string `json:"mTertiary"`
	MOnTertiary       string `json:"mOnTertiary"`
	MError            string `json:"mError"`
	MOnError          string `json:"mOnError"`
	MSurface          string `json:"mSurface"`
	MOnSurface        string `json:"mOnSurface"`
	MSurfaceVariant   string `json:"mSurfaceVariant"`
	MOnSurfaceVariant string `json:"mOnSurfaceVariant"`
	MOutline          string `json:"mOutline"`
	MShadow           string `json:"mShadow"`
	MHover            string `json:"mHover"`
	MOnHover          string `json:"mOnHover"`
	MGold             string `json:"mGold"`
	MSuccess          string `json:"mSuccess"`
}

// ThemeFile represents the dual dark/light JSON theme definition.
type ThemeFile struct {
	Dark  ThemeConfig `json:"dark"`
	Light ThemeConfig `json:"light"`
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
	Hover            color.Color // Warm accent for descriptions
	OnHover          color.Color // Text on hover blocks
	Gold             color.Color // Pinned star color
	Success          color.Color // Green dot, success indicators
}

// BuiltinPresets contains compiled-in themes: spotify, catppuccin, dracula, gruvbox, monochrome, tokyonight.
var BuiltinPresets = map[string]ThemeFile{
	"spotify": {
		Dark: ThemeConfig{
			MPrimary: "#1db954", MOnPrimary: "#000000",
			MSecondary: "#1ed760", MOnSecondary: "#000000",
			MTertiary: "#1db954", MOnTertiary: "#000000",
			MError: "#e91429", MOnError: "#ffffff",
			MSurface: "#121212", MOnSurface: "#ffffff",
			MSurfaceVariant: "#282828", MOnSurfaceVariant: "#b3b3b3",
			MOutline: "#535353", MShadow: "#000000",
			MHover: "#1ed760", MOnHover: "#000000",
			MGold: "#f59b23", MSuccess: "#1db954",
		},
		Light: ThemeConfig{
			MPrimary: "#1db954", MOnPrimary: "#ffffff",
			MSecondary: "#169c46", MOnSecondary: "#ffffff",
			MTertiary: "#1db954", MOnTertiary: "#ffffff",
			MError: "#e91429", MOnError: "#ffffff",
			MSurface: "#ffffff", MOnSurface: "#121212",
			MSurfaceVariant: "#f2f2f2", MOnSurfaceVariant: "#6a6a6a",
			MOutline: "#cccccc", MShadow: "#000000",
			MHover: "#169c46", MOnHover: "#ffffff",
			MGold: "#d97706", MSuccess: "#1db954",
		},
	},
	"catppuccin": {
		Dark: ThemeConfig{
			MPrimary: "#cba6f7", MOnPrimary: "#11111b",
			MSecondary: "#b4befe", MOnSecondary: "#11111b",
			MTertiary: "#94e2d5", MOnTertiary: "#11111b",
			MError: "#f38ba8", MOnError: "#11111b",
			MSurface: "#1e1e2e", MOnSurface: "#cdd6f4",
			MSurfaceVariant: "#313244", MOnSurfaceVariant: "#a6adc8",
			MOutline: "#6c7086", MShadow: "#11111b",
			MHover: "#fab387", MOnHover: "#11111b",
			MGold: "#f9e2af", MSuccess: "#a6e3a1",
		},
		Light: ThemeConfig{
			MPrimary: "#8839ef", MOnPrimary: "#ffffff",
			MSecondary: "#7287fd", MOnSecondary: "#ffffff",
			MTertiary: "#179299", MOnTertiary: "#ffffff",
			MError: "#d20f39", MOnError: "#ffffff",
			MSurface: "#eff1f5", MOnSurface: "#4c4f69",
			MSurfaceVariant: "#ccd0da", MOnSurfaceVariant: "#6c6f85",
			MOutline: "#9ca0b0", MShadow: "#dce0e8",
			MHover: "#fe640b", MOnHover: "#ffffff",
			MGold: "#df8e1d", MSuccess: "#40a02b",
		},
	},
	"dracula": {
		Dark: ThemeConfig{
			MPrimary: "#bd93f9", MOnPrimary: "#282a36",
			MSecondary: "#ff79c6", MOnSecondary: "#282a36",
			MTertiary: "#8be9fd", MOnTertiary: "#282a36",
			MError: "#ff5555", MOnError: "#282a36",
			MSurface: "#282a36", MOnSurface: "#f8f8f2",
			MSurfaceVariant: "#44475a", MOnSurfaceVariant: "#bfbfbf",
			MOutline: "#6272a4", MShadow: "#191a21",
			MHover: "#ffb86c", MOnHover: "#282a36",
			MGold: "#f1fa8c", MSuccess: "#50fa7b",
		},
		Light: ThemeConfig{
			MPrimary: "#7b5ea7", MOnPrimary: "#ffffff",
			MSecondary: "#b83280", MOnSecondary: "#ffffff",
			MTertiary: "#0d768a", MOnTertiary: "#ffffff",
			MError: "#cb1a1a", MOnError: "#ffffff",
			MSurface: "#f8f8f2", MOnSurface: "#282a36",
			MSurfaceVariant: "#e2e4e9", MOnSurfaceVariant: "#6272a4",
			MOutline: "#a0a4b8", MShadow: "#d5d7e0",
			MHover: "#c96a00", MOnHover: "#ffffff",
			MGold: "#b08800", MSuccess: "#2b8a3e",
		},
	},
	"gruvbox": {
		Dark: ThemeConfig{
			MPrimary: "#fabd2f", MOnPrimary: "#282828",
			MSecondary: "#fe8019", MOnSecondary: "#282828",
			MTertiary: "#8ec07c", MOnTertiary: "#282828",
			MError: "#fb4934", MOnError: "#282828",
			MSurface: "#282828", MOnSurface: "#ebdbb2",
			MSurfaceVariant: "#3c3836", MOnSurfaceVariant: "#a89984",
			MOutline: "#665c54", MShadow: "#1d2021",
			MHover: "#d3869b", MOnHover: "#282828",
			MGold: "#fabd2f", MSuccess: "#b8bb26",
		},
		Light: ThemeConfig{
			MPrimary: "#b57614", MOnPrimary: "#fbf1c7",
			MSecondary: "#af3a03", MOnSecondary: "#fbf1c7",
			MTertiary: "#427b58", MOnTertiary: "#fbf1c7",
			MError: "#9d0006", MOnError: "#fbf1c7",
			MSurface: "#fbf1c7", MOnSurface: "#3c3836",
			MSurfaceVariant: "#ebdbb2", MOnSurfaceVariant: "#7c6f64",
			MOutline: "#bdae93", MShadow: "#d5c4a1",
			MHover: "#8f3f71", MOnHover: "#fbf1c7",
			MGold: "#b57614", MSuccess: "#79740e",
		},
	},
	"monochrome": {
		Dark: ThemeConfig{
			MPrimary: "#ffffff", MOnPrimary: "#000000",
			MSecondary: "#ffffff", MOnSecondary: "#000000",
			MTertiary: "#ffffff", MOnTertiary: "#000000",
			MError: "#ffffff", MOnError: "#000000",
			MSurface: "#000000", MOnSurface: "#ffffff",
			MSurfaceVariant: "#222222", MOnSurfaceVariant: "#888888",
			MOutline: "#555555", MShadow: "#000000",
			MHover: "#ffffff", MOnHover: "#000000",
			MGold: "#ffffff", MSuccess: "#ffffff",
		},
		Light: ThemeConfig{
			MPrimary: "#000000", MOnPrimary: "#ffffff",
			MSecondary: "#000000", MOnSecondary: "#ffffff",
			MTertiary: "#000000", MOnTertiary: "#ffffff",
			MError: "#000000", MOnError: "#ffffff",
			MSurface: "#ffffff", MOnSurface: "#000000",
			MSurfaceVariant: "#dddddd", MOnSurfaceVariant: "#777777",
			MOutline: "#aaaaaa", MShadow: "#ffffff",
			MHover: "#000000", MOnHover: "#ffffff",
			MGold: "#000000", MSuccess: "#000000",
		},
	},
	"tokyonight": {
		Dark: ThemeConfig{
			MPrimary: "#7aa2f7", MOnPrimary: "#1a1b26",
			MSecondary: "#bb9af7", MOnSecondary: "#1a1b26",
			MTertiary: "#7dcfff", MOnTertiary: "#1a1b26",
			MError: "#f7768e", MOnError: "#1a1b26",
			MSurface: "#1a1b26", MOnSurface: "#c0caf5",
			MSurfaceVariant: "#24283b", MOnSurfaceVariant: "#787c99",
			MOutline: "#565f89", MShadow: "#16161e",
			MHover: "#ff9e64", MOnHover: "#1a1b26",
			MGold: "#e0af68", MSuccess: "#9ece6a",
		},
		Light: ThemeConfig{
			MPrimary: "#2e7de9", MOnPrimary: "#ffffff",
			MSecondary: "#7847bd", MOnSecondary: "#ffffff",
			MTertiary: "#007197", MOnTertiary: "#ffffff",
			MError: "#8c4351", MOnError: "#ffffff",
			MSurface: "#e1e2e7", MOnSurface: "#3760bf",
			MSurfaceVariant: "#cbccd1", MOnSurfaceVariant: "#6172b0",
			MOutline: "#8990b3", MShadow: "#d0d1d6",
			MHover: "#b15c00", MOnHover: "#ffffff",
			MGold: "#8f5e15", MSuccess: "#387068",
		},
	},
}

// stripJSONComments removes single-line and multi-line comments from JSON data
func stripJSONComments(data []byte) []byte {
	var out []byte
	inString := false
	inSingleComment := false
	inMultiComment := false
	escape := false

	n := len(data)
	for i := 0; i < n; i++ {
		b := data[i]

		if inSingleComment {
			if b == '\n' {
				inSingleComment = false
				out = append(out, b)
			}
			continue
		}

		if inMultiComment {
			if b == '*' && i+1 < n && data[i+1] == '/' {
				inMultiComment = false
				i++
			}
			continue
		}

		if inString {
			out = append(out, b)
			if escape {
				escape = false
			} else if b == '\\' {
				escape = true
			} else if b == '"' {
				inString = false
			}
			continue
		}

		if b == '"' {
			inString = true
			out = append(out, b)
			continue
		}

		if b == '/' && i+1 < n {
			if data[i+1] == '/' {
				inSingleComment = true
				i++
				continue
			} else if data[i+1] == '*' {
				inMultiComment = true
				i++
				continue
			}
		}

		out = append(out, b)
	}

	return out
}

// ValidateThemeFile parses JSON theme data, stripping any comments and handling single or dual modes.
func ValidateThemeFile(data []byte) (*ThemeFile, error) {
	cleaned := stripJSONComments(data)
	var tf ThemeFile
	if err := json.Unmarshal(cleaned, &tf); err != nil {
		var flat ThemeConfig
		if err2 := json.Unmarshal(cleaned, &flat); err2 == nil && flat.MSurface != "" && flat.MPrimary != "" {
			return &ThemeFile{Dark: flat, Light: flat}, nil
		}
		return nil, err
	}
	if tf.Dark.MSurface == "" && tf.Light.MSurface != "" {
		tf.Dark = tf.Light
	} else if tf.Light.MSurface == "" && tf.Dark.MSurface != "" {
		tf.Light = tf.Dark
	}
	if tf.Dark.MSurface == "" || tf.Dark.MPrimary == "" {
		return nil, fmt.Errorf("theme missing required color fields")
	}
	return &tf, nil
}

// ListAvailableThemes returns all built-in presets and valid user themes from ~/.config/spotumn/themes/
func ListAvailableThemes() []string {
	seen := make(map[string]bool)
	var themes []string

	presets := []string{"spotify", "catppuccin", "dracula", "gruvbox", "monochrome", "tokyonight"}
	for _, p := range presets {
		seen[p] = true
		themes = append(themes, p)
	}

	dir := config.GetThemesDir()
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			if _, err := ValidateThemeFile(data); err == nil {
				name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
				if !seen[name] {
					seen[name] = true
					themes = append(themes, name)
				}
			}
		}
	}
	return themes
}

// LoadTheme resolves a theme by name for dark or light mode.
func LoadTheme(themeName string, isDark bool) ThemePalette {
	themeName = strings.ToLower(strings.TrimSpace(themeName))
	if themeName == "" {
		themeName = "spotify"
	}

	// 1. Check user theme folder ~/.config/spotumn/themes/<name>.json
	themePath := filepath.Join(config.GetThemesDir(), themeName+".json")
	if data, err := os.ReadFile(themePath); err == nil {
		if tf, err := ValidateThemeFile(data); err == nil {
			if isDark {
				return resolveConfig(tf.Dark)
			}
			return resolveConfig(tf.Light)
		}
	}

	// 2. Check built-in presets
	if tf, ok := BuiltinPresets[themeName]; ok {
		if isDark {
			return resolveConfig(tf.Dark)
		}
		return resolveConfig(tf.Light)
	}

	// 3. Fallback to default spotify preset
	tf := BuiltinPresets["spotify"]
	if isDark {
		return resolveConfig(tf.Dark)
	}
	return resolveConfig(tf.Light)
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

func catppuccinMocha() ThemeConfig {
	return BuiltinPresets["catppuccin"].Dark
}

func mergeConfig(base, custom ThemeConfig) ThemeConfig {
	if custom.MPrimary != "" { base.MPrimary = custom.MPrimary }
	if custom.MOnPrimary != "" { base.MOnPrimary = custom.MOnPrimary }
	if custom.MSecondary != "" { base.MSecondary = custom.MSecondary }
	if custom.MOnSecondary != "" { base.MOnSecondary = custom.MOnSecondary }
	if custom.MTertiary != "" { base.MTertiary = custom.MTertiary }
	if custom.MOnTertiary != "" { base.MOnTertiary = custom.MOnTertiary }
	if custom.MError != "" { base.MError = custom.MError }
	if custom.MOnError != "" { base.MOnError = custom.MOnError }
	if custom.MSurface != "" { base.MSurface = custom.MSurface }
	if custom.MOnSurface != "" { base.MOnSurface = custom.MOnSurface }
	if custom.MSurfaceVariant != "" { base.MSurfaceVariant = custom.MSurfaceVariant }
	if custom.MOnSurfaceVariant != "" { base.MOnSurfaceVariant = custom.MOnSurfaceVariant }
	if custom.MOutline != "" { base.MOutline = custom.MOutline }
	if custom.MShadow != "" { base.MShadow = custom.MShadow }
	if custom.MHover != "" { base.MHover = custom.MHover }
	if custom.MOnHover != "" { base.MOnHover = custom.MOnHover }
	if custom.MGold != "" { base.MGold = custom.MGold }
	if custom.MSuccess != "" { base.MSuccess = custom.MSuccess }
	return base
}

// LoadThemeFile is kept for test compatibility.
func LoadThemeFile(customPath ...string) ThemeConfig {
	if len(customPath) > 0 && customPath[0] != "" {
		data, err := os.ReadFile(customPath[0])
		if err == nil {
			var tf ThemeFile
			if err := json.Unmarshal(data, &tf); err == nil {
				return tf.Dark
			}
		}
	}
	return BuiltinPresets["spotify"].Dark
}

// ApplyTheme applies a specific theme name and dark/light flag and updates styles.
func ApplyTheme(themeName string, isDark bool) {
	CurrentTheme = LoadTheme(themeName, isDark)
	updateStyles()
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

// SetTheme selects the palette based on config and terminal darkness, and rebuilds styles.
func SetTheme(isDark bool) {
	cfg := config.Get()
	if cfg.AppearanceMode == "dark" {
		isDark = true
	} else if cfg.AppearanceMode == "light" {
		isDark = false
	}
	themeName := cfg.Theme
	if themeName == "" {
		themeName = "spotify"
	}
	ApplyTheme(themeName, isDark)
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

func isComplexCombiningMark(r rune) bool {
	if !unicode.Is(unicode.M, r) {
		return false
	}
	return (r >= 0x0300 && r <= 0x036F) || // Combining Diacritical Marks
		(r >= 0x0590 && r <= 0x08FF) || // Hebrew, Arabic, Syriac, Thaana
		(r >= 0x0900 && r <= 0x0DFF) || // Indic (Devanagari, Bengali, Gurmukhi, Gujarati, Oriya, Tamil, Telugu, Kannada, Malayalam, Sinhala)
		(r >= 0x0E00 && r <= 0x0EFF) || // Thai, Lao
		(r >= 0x0F00 && r <= 0x0FFF) || // Tibetan
		(r >= 0x1000 && r <= 0x109F) || // Myanmar (Burmese)
		(r >= 0x1780 && r <= 0x17FF) || // Khmer
		(r >= 0x1900 && r <= 0x1BFF) || // Tai, Balinese, Sundanese, Batak
		(r >= 0x1CD0 && r <= 0x1CFF) || // Vedic Extensions
		(r >= 0x1DC0 && r <= 0x1DFF) || // Combining Diacritical Marks Supplement
		(r >= 0x20D0 && r <= 0x20FF) || // Combining Diacritical Marks for Symbols
		(r >= 0xA800 && r <= 0xABFF) || // Saurashtra, Kayah Li, Rejang, Javanese, Cham, Tai Viet, Meetei Mayek
		(r >= 0xFE20 && r <= 0xFE2F)    // Combining Half Marks
}

func runeDisplayWidth(r rune) int {
	w := ansi.StringWidth(string(r))
	if w == 0 && isComplexCombiningMark(r) {
		return 1
	}
	if unicode.Is(unicode.Mc, r) && w == 0 {
		return 1
	}
	return w
}

func StringDisplayWidth(s string) int {
	w := ansi.StringWidth(s)
	for _, r := range s {
		if isComplexCombiningMark(r) && ansi.StringWidth(string(r)) == 0 {
			w++
		}
	}
	return w
}

// TruncateVisualWidth truncates a string (with or without ANSI sequences) to exact max visual cells.
func TruncateVisualWidth(s string, maxW int, tail string) string {
	if maxW <= 0 {
		return ""
	}
	if StringDisplayWidth(s) <= maxW {
		return s
	}

	tailW := StringDisplayWidth(tail)
	targetW := maxW - tailW
	if targetW < 0 {
		targetW = 0
	}

	var sb strings.Builder
	curW := 0
	inEsc := false
	hadEsc := false

	for i, r := range s {
		if r == 0x1b {
			inEsc = true
			hadEsc = true
			sb.WriteRune(r)
			continue
		}
		if inEsc {
			sb.WriteRune(r)
			if r == 'm' || r == 'K' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z' && r != 'm') {
				inEsc = false
			}
			continue
		}

		rw := runeDisplayWidth(r)
		if curW+rw > targetW {
			_ = i
			break
		}
		curW += rw
		sb.WriteRune(r)
	}

	sb.WriteString(tail)
	if hadEsc {
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}

// RenderPaddedLine renders a row to exact width with the given style
func RenderPaddedLine(rawText string, style lipgloss.Style, width int) string {
	if width <= 0 {
		return ""
	}
	trunc := TruncateString(rawText, width)
	w := StringDisplayWidth(trunc)
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
	w := StringDisplayWidth(s)
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
	w := StringDisplayWidth(s)
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
	w := StringDisplayWidth(s)
	if w > width {
		s = ansi.Truncate(s, width, "…")
		w = StringDisplayWidth(s)
	}
	pad := width - w
	if pad <= 0 {
		return s
	}
	return s + BgPad(pad)
}

// TruncateString truncates a string to max width with ellipsis using visual cell measurements
func TruncateString(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	if StringDisplayWidth(s) <= maxW {
		return s
	}
	return TruncateVisualWidth(s, maxW, "…")
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
