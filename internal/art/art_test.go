package art

import (
	"image"
	"image/color"
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

func TestRendererInitialization(t *testing.T) {
	r := NewRenderer("auto")
	if r.mode != "auto" {
		t.Errorf("expected auto mode, got %s", r.mode)
	}

	if BuildVariant != "full" && BuildVariant != "minimal" {
		t.Errorf("unexpected BuildVariant %s", BuildVariant)
	}

	if BuildVariant == "minimal" && HasChafaSupport {
		t.Error("minimal build should not have chafa support")
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
