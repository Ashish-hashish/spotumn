package art

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"spotumn/internal/config"
)

type Renderer struct {
	client    *http.Client
	cacheDir  string
	mode      string // "auto", "image", "ansi"
	hasChafa  bool
	memCache  map[string]string
	diskCache map[string]string
	cacheKeys []string
	mu        sync.Mutex
}

func NewRenderer(mode string) *Renderer {
	cacheDir := filepath.Join(config.GetDir(), "art_cache")
	_ = os.MkdirAll(cacheDir, 0700)

	_, err := exec.LookPath("chafa")
	hasChafa := (err == nil)

	if mode == "" {
		mode = "auto"
	}

	return &Renderer{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cacheDir:  cacheDir,
		mode:      strings.ToLower(strings.TrimSpace(mode)),
		hasChafa:  hasChafa,
		memCache:  make(map[string]string),
		diskCache: make(map[string]string),
	}
}

// isGraphicTerminal detects if terminal supports true image graphics (Kitty, iTerm, Sixel)
func isGraphicTerminal() (bool, string) {
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true, "kitty"
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "xterm-kitty" || term == "kitty" {
		return true, "kitty"
	}
	if os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return true, "kitty"
	}
	termProg := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	if termProg == "wezterm" || os.Getenv("WEZTERM_EXECUTABLE") != "" || os.Getenv("WEZTERM_PANE") != "" {
		return true, "kitty"
	}
	if termProg == "iterm.app" {
		return true, "iterm"
	}
	if term == "foot" || term == "mlterm" {
		return true, "sixels"
	}
	return false, "symbols"
}

// CanShowGraphic returns true if terminal and chafa support direct image rendering
func (r *Renderer) CanShowGraphic() bool {
	if r.mode == "ansi" {
		return false
	}
	if !r.hasChafa {
		return false
	}
	if r.mode == "image" {
		return true
	}
	isGraphic, _ := isGraphicTerminal()
	return isGraphic
}

// DrawGraphicArt executes chafa to generate terminal graphics escape codes
// and writes them directly to terminal stdout at the target cell position.
func DrawGraphicArt(diskPath string, row, col, width, height int) error {
	if diskPath == "" || width <= 0 || height <= 0 || row <= 0 || col <= 0 {
		return nil
	}

	isGraphic, format := isGraphicTerminal()
	if !isGraphic {
		return nil
	}

	cmd := exec.Command("chafa", "--probe=off", "-f", format, fmt.Sprintf("--size=%dx%d", width, height), diskPath)
	out, err := cmd.Output()
	if err != nil || len(out) == 0 {
		return err
	}

	s := string(out)
	s = strings.ReplaceAll(s, "\x1b[?25l", "")
	s = strings.ReplaceAll(s, "\x1b[?25h", "")
	s = strings.TrimRight(s, "\r\n")

	// Save cursor, clear previous graphics, move cursor to (row, col), send chafa image, restore cursor
	payload := fmt.Sprintf("\x1b[s\x1b_Ga=d,d=a\x1b\\\x1b[%d;%dH%s\x1b[u", row, col, s)
	_, err = os.Stdout.WriteString(payload)
	return err
}

// ClearGraphics clears any active Kitty graphics from terminal
func ClearGraphics() {
	isGraphic, _ := isGraphicTerminal()
	if isGraphic {
		_, _ = os.Stdout.WriteString("\x1b_Ga=d,d=a\x1b\\")
	}
}

// Render returns the UI layout representation and disk path of the album artwork.
func (r *Renderer) Render(imageURL string, width, height int) (string, string, error) {
	if imageURL == "" || width <= 0 || height <= 0 {
		return "", "", nil
	}

	key := fmt.Sprintf("%s:%d:%d", imageURL, width, height)

	r.mu.Lock()
	if val, ok := r.memCache[key]; ok {
		diskPath := r.diskCache[imageURL]
		r.mu.Unlock()
		return val, diskPath, nil
	}
	r.mu.Unlock()

	diskPath, img, err := r.ensureImage(imageURL)
	if err != nil {
		return "", "", err
	}

	var rendered string

	// 1. If terminal supports high-res graphics (Kitty/Ghostty/WezTerm/iTerm2),
	// reserve exact rectangular space for the graphic.
	if r.CanShowGraphic() {
		var rows []string
		for i := 0; i < height; i++ {
			rows = append(rows, strings.Repeat(" ", width))
		}
		rendered = strings.Join(rows, "\n")
	} else {
		// 2. ANSI Fallback: Use Chafa high-definition 24-bit TrueColor block+sextants
		if r.hasChafa && diskPath != "" {
			if out, err := r.renderWithChafa(diskPath, width, height); err == nil && len(strings.TrimSpace(out)) > 0 {
				rendered = out
			}
		}

		// Universal pure Go fallback: 24-bit Truecolor half-blocks (▀)
		if rendered == "" && img != nil {
			rendered = r.toHalfBlocks(img, width, height)
		}
	}

	r.mu.Lock()
	if len(r.cacheKeys) >= 10 {
		oldest := r.cacheKeys[0]
		r.cacheKeys = r.cacheKeys[1:]
		delete(r.memCache, oldest)
	}
	r.memCache[key] = rendered
	r.diskCache[imageURL] = diskPath
	r.cacheKeys = append(r.cacheKeys, key)
	r.mu.Unlock()

	return rendered, diskPath, nil
}

// renderWithChafa uses the system chafa binary to render 24-bit Truecolor sextant/quad character art
func (r *Renderer) renderWithChafa(diskPath string, width, height int) (string, error) {
	cmd := exec.Command("chafa",
		"--probe=off",
		"-c", "full",
		"--format=symbols",
		"--symbols=block+sextant",
		fmt.Sprintf("--size=%dx%d", width, height),
		diskPath,
	)
	out, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		return cleanChafaOutput(string(out)), nil
	}

	cmd = exec.Command("chafa",
		"--probe=off",
		"-c", "full",
		"--format=symbols",
		"--symbols=quad+half",
		fmt.Sprintf("--size=%dx%d", width, height),
		diskPath,
	)
	out, err = cmd.Output()
	if err != nil {
		return "", err
	}
	return cleanChafaOutput(string(out)), nil
}

func cleanChafaOutput(s string) string {
	s = strings.ReplaceAll(s, "\x1b[?25l", "")
	s = strings.ReplaceAll(s, "\x1b[?25h", "")
	return strings.TrimRight(s, "\r\n")
}

func (r *Renderer) ensureImage(imageURL string) (string, image.Image, error) {
	if !strings.HasPrefix(imageURL, "https://") {
		return "", nil, fmt.Errorf("insecure or invalid image URL: %s", imageURL)
	}

	h := sha256.Sum256([]byte(imageURL))
	diskPath := filepath.Join(r.cacheDir, hex.EncodeToString(h[:]))

	// Try loading from disk
	if f, err := os.Open(diskPath); err == nil {
		defer f.Close()
		img, _, err := image.Decode(f)
		if err == nil {
			return diskPath, img, nil
		}
	}

	// Fetch from network with bounded size reader
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return "", nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	limitReader := io.LimitReader(resp.Body, 4*1024*1024)

	tmpFile, err := os.CreateTemp(r.cacheDir, "art-*")
	if err != nil {
		return "", nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, limitReader); err != nil {
		tmpFile.Close()
		return "", nil, err
	}
	tmpFile.Close()

	_ = os.Rename(tmpFile.Name(), diskPath)

	f, err := os.Open(diskPath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return "", nil, err
	}

	return diskPath, img, nil
}

// toHalfBlocks converts an image into pure Go ANSI half-block characters (▀)
func (r *Renderer) toHalfBlocks(img image.Image, targetW, targetH int) string {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return ""
	}

	pixelH := targetH * 2
	pixelW := targetW

	var sb strings.Builder

	for y := 0; y < pixelH; y += 2 {
		for x := 0; x < pixelW; x++ {
			srcX := bounds.Min.X + (x * imgW / pixelW)
			srcYTop := bounds.Min.Y + (y * imgH / pixelH)
			srcYBot := bounds.Min.Y + ((y + 1) * imgH / pixelH)

			rTop, gTop, bTop, _ := img.At(srcX, srcYTop).RGBA()
			rBot, gBot, bBot, _ := img.At(srcX, srcYBot).RGBA()

			sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m",
				rTop>>8, gTop>>8, bTop>>8,
				rBot>>8, gBot>>8, bBot>>8,
			))
		}
		if y+2 < pixelH {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}
