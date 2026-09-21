package art

import (
	"image"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// createTestImage creates a solid-color test image
func createTestImage(w, h int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestGraphicTerminalDetection(t *testing.T) {
	origKitty := os.Getenv("KITTY_WINDOW_ID")
	origTerm := os.Getenv("TERM")
	origGhost := os.Getenv("GHOSTTY_RESOURCES_DIR")
	origProg := os.Getenv("TERM_PROGRAM")
	defer func() {
		os.Setenv("KITTY_WINDOW_ID", origKitty)
		os.Setenv("TERM", origTerm)
		os.Setenv("GHOSTTY_RESOURCES_DIR", origGhost)
		os.Setenv("TERM_PROGRAM", origProg)
	}()

	os.Unsetenv("KITTY_WINDOW_ID")
	os.Setenv("TERM", "dumb")
	os.Unsetenv("GHOSTTY_RESOURCES_DIR")
	os.Unsetenv("TERM_PROGRAM")

	isG, fmtStr := isGraphicTerminal()
	if isG || fmtStr != "symbols" {
		t.Errorf("expected false, symbols for dumb terminal, got %v, %s", isG, fmtStr)
	}

	os.Setenv("KITTY_WINDOW_ID", "1")
	isG, fmtStr = isGraphicTerminal()
	if !isG || fmtStr != "kitty" {
		t.Errorf("expected true, kitty when KITTY_WINDOW_ID is set, got %v, %s", isG, fmtStr)
	}
	os.Unsetenv("KITTY_WINDOW_ID")

	os.Setenv("GHOSTTY_RESOURCES_DIR", "/usr/share/ghostty")
	isG, fmtStr = isGraphicTerminal()
	if !isG || fmtStr != "kitty" {
		t.Errorf("expected true, kitty when GHOSTTY_RESOURCES_DIR is set, got %v, %s", isG, fmtStr)
	}
	os.Unsetenv("GHOSTTY_RESOURCES_DIR")

	os.Setenv("TERM_PROGRAM", "WezTerm")
	isG, fmtStr = isGraphicTerminal()
	if !isG || fmtStr != "kitty" {
		t.Errorf("expected true, kitty when TERM_PROGRAM is WezTerm, got %v, %s", isG, fmtStr)
	}
	os.Unsetenv("TERM_PROGRAM")

	os.Setenv("TERM_PROGRAM", "iTerm.app")
	isG, fmtStr = isGraphicTerminal()
	if !isG || fmtStr != "iterm" {
		t.Errorf("expected true, iterm when TERM_PROGRAM is iTerm.app, got %v, %s", isG, fmtStr)
	}
	os.Unsetenv("TERM_PROGRAM")

	os.Setenv("TERM", "foot")
	isG, fmtStr = isGraphicTerminal()
	if !isG || fmtStr != "sixels" {
		t.Errorf("expected true, sixels when TERM is foot, got %v, %s", isG, fmtStr)
	}
}

func TestToHalfBlocksFallback(t *testing.T) {
	ren := NewRenderer("ansi")
	img := createTestImage(20, 20, color.RGBA{R: 120, G: 200, B: 80, A: 255})

	cols := 20
	rows := 10
	output := ren.toHalfBlocks(img, cols, rows)

	lines := strings.Split(output, "\n")
	if len(lines) != rows {
		t.Fatalf("expected %d rows, got %d", rows, len(lines))
	}

	// Must contain half block character ▀ and 24-bit TrueColor escape sequences
	for idx, line := range lines {
		if !strings.Contains(line, "▀") {
			t.Errorf("row %d expected to contain half block ▀", idx)
		}
		if !strings.Contains(line, "\x1b[38;2;") {
			t.Errorf("row %d expected to contain 24-bit foreground escape", idx)
		}
		w := ansi.StringWidth(line)
		if w != cols {
			t.Errorf("row %d: expected visual width %d, got %d", idx, cols, w)
		}
	}
}

func TestRendererModeSwitching(t *testing.T) {
	rAuto := NewRenderer("auto")
	if rAuto.mode != "auto" {
		t.Errorf("expected auto, got %s", rAuto.mode)
	}

	rImage := NewRenderer("image")
	if !rImage.hasChafa {
		t.Log("chafa not found, skipping hasChafa assertions")
	} else {
		if !rImage.CanShowGraphic() {
			t.Error("expected CanShowGraphic=true for mode image when chafa is installed")
		}
	}

	rAnsi := NewRenderer("ansi")
	if rAnsi.CanShowGraphic() {
		t.Error("expected CanShowGraphic=false for mode ansi")
	}
}
