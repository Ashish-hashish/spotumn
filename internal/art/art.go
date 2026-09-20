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
	"path/filepath"
	"strings"
	"sync"
	"time"

	"spotumn/internal/config"
)

type Renderer struct {
	client    *http.Client
	cacheDir  string
	memCache  map[string]string
	cacheKeys []string
	mu        sync.Mutex
}

func NewRenderer() *Renderer {
	cacheDir := filepath.Join(config.GetDir(), "art_cache")
	_ = os.MkdirAll(cacheDir, 0700)

	return &Renderer{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cacheDir: cacheDir,
		memCache: make(map[string]string),
	}
}

// Render returns ANSI half-block representation of the album art at target width & height
func (r *Renderer) Render(imageURL string, width, height int) (string, error) {
	if imageURL == "" || width <= 0 || height <= 0 {
		return "", nil
	}

	key := fmt.Sprintf("%s:%d:%d", imageURL, width, height)

	r.mu.Lock()
	if val, ok := r.memCache[key]; ok {
		r.mu.Unlock()
		return val, nil
	}
	r.mu.Unlock()

	img, err := r.loadImage(imageURL)
	if err != nil {
		return "", err
	}

	rendered := r.toHalfBlocks(img, width, height)

	r.mu.Lock()
	// Evict oldest if memory cache exceeds 10 items (strict RAM conservation)
	if len(r.cacheKeys) >= 10 {
		oldest := r.cacheKeys[0]
		r.cacheKeys = r.cacheKeys[1:]
		delete(r.memCache, oldest)
	}
	r.memCache[key] = rendered
	r.cacheKeys = append(r.cacheKeys, key)
	r.mu.Unlock()

	return rendered, nil
}

func (r *Renderer) loadImage(imageURL string) (image.Image, error) {
	if !strings.HasPrefix(imageURL, "https://") {
		return nil, fmt.Errorf("insecure or invalid image URL: %s", imageURL)
	}

	h := sha256.Sum256([]byte(imageURL))
	diskPath := filepath.Join(r.cacheDir, hex.EncodeToString(h[:]))

	// Try loading from disk
	if f, err := os.Open(diskPath); err == nil {
		defer f.Close()
		img, _, err := image.Decode(f)
		if err == nil {
			return img, nil
		}
	}

	// Fetch from network with bounded size reader
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Limit to max 4MB to prevent memory exhaustion
	limitReader := io.LimitReader(resp.Body, 4*1024*1024)

	// Save to disk first
	tmpFile, err := os.CreateTemp(r.cacheDir, "art-*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, limitReader); err != nil {
		tmpFile.Close()
		return nil, err
	}
	tmpFile.Close()

	// Atomic rename to final disk cache path
	_ = os.Rename(tmpFile.Name(), diskPath)

	// Decode image
	f, err := os.Open(diskPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// toHalfBlocks converts an image into ANSI half-block characters (▀)
func (r *Renderer) toHalfBlocks(img image.Image, targetW, targetH int) string {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	if imgW == 0 || imgH == 0 {
		return ""
	}

	// Each row of characters renders 2 vertical pixels
	pixelH := targetH * 2
	pixelW := targetW

	var sb strings.Builder

	for y := 0; y < pixelH; y += 2 {
		for x := 0; x < pixelW; x++ {
			// Sample top pixel
			srcX := bounds.Min.X + (x * imgW / pixelW)
			srcYTop := bounds.Min.Y + (y * imgH / pixelH)
			srcYBot := bounds.Min.Y + ((y + 1) * imgH / pixelH)

			rTop, gTop, bTop, _ := img.At(srcX, srcYTop).RGBA()
			rBot, gBot, bBot, _ := img.At(srcX, srcYBot).RGBA()

			// RGBA returns values in [0, 65535], convert to [0, 255]
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
