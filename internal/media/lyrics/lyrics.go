// Synced and plain lyrics provider - fetches and parses time-synchronized lyrics from LRCLIB.
package lyrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Line struct {
	TimeMs int
	Text   string
}

type Provider struct {
	client *http.Client
}

func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

var lrcRegex = regexp.MustCompile(`^\[(\d{2}):(\d{2})\.(\d{2,3})\](.*)$`)

type lrclibItem struct {
	SyncedLyrics string `json:"syncedLyrics"`
	PlainLyrics  string `json:"plainLyrics"`
}

// query lrclib with exact duration first, falling back to fuzzy text search
func (p *Provider) FetchSyncedLyrics(trackName, artistName string, durationSec int) ([]Line, bool, error) {
	trackName = cleanTrackName(trackName)
	if strings.TrimSpace(trackName) == "" {
		return nil, false, nil
	}

	u := fmt.Sprintf("https://lrclib.net/api/get?track_name=%s&artist_name=%s&duration=%d",
		url.QueryEscape(trackName),
		url.QueryEscape(artistName),
		durationSec,
	)

	if lines, synced := p.fetchFromURL(u, false); len(lines) > 0 {
		return lines, synced, nil
	}

	searchURL := fmt.Sprintf("https://lrclib.net/api/search?track_name=%s&artist_name=%s",
		url.QueryEscape(trackName),
		url.QueryEscape(artistName),
	)

	if lines, synced := p.fetchFromURL(searchURL, true); len(lines) > 0 {
		return lines, synced, nil
	}

	generalSearchURL := fmt.Sprintf("https://lrclib.net/api/search?q=%s",
		url.QueryEscape(trackName+" "+artistName),
	)

	lines, synced := p.fetchFromURL(generalSearchURL, true)
	return lines, synced, nil
}

func (p *Provider) fetchFromURL(targetURL string, isArray bool) ([]Line, bool) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, false
	}
	req.Header.Set("User-Agent", "spotumn/1.0.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false
	}

	limitReader := io.LimitReader(resp.Body, 512*1024)

	if isArray {
		var items []lrclibItem
		if err := json.NewDecoder(limitReader).Decode(&items); err != nil {
			return nil, false
		}
		for _, item := range items {
			if item.SyncedLyrics != "" {
				return ParseLRC(item.SyncedLyrics), true
			}
		}
		for _, item := range items {
			if item.PlainLyrics != "" {
				return parsePlain(item.PlainLyrics), false
			}
		}
		return nil, false
	}

	var item lrclibItem
	if err := json.NewDecoder(limitReader).Decode(&item); err != nil {
		return nil, false
	}
	return extractItemLyrics(item)
}

func extractItemLyrics(item lrclibItem) ([]Line, bool) {
	if item.SyncedLyrics != "" {
		return ParseLRC(item.SyncedLyrics), true
	}
	if item.PlainLyrics != "" {
		return parsePlain(item.PlainLyrics), false
	}
	return nil, false
}

// strip suffix metadata like remasters or live tags so lyrics match cleanly
func cleanTrackName(name string) string {
	if idx := strings.Index(name, " - "); idx > 0 {
		name = name[:idx]
	}
	return strings.TrimSpace(name)
}

func parsePlain(plain string) []Line {
	var lines []Line
	for _, l := range strings.Split(plain, "\n") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			lines = append(lines, Line{TimeMs: 0, Text: trimmed})
		}
	}
	return lines
}

// parse standard lrc timestamps into millisecond offsets and sort chronologically
func ParseLRC(raw string) []Line {
	var lines []Line
	for _, rawLine := range strings.Split(raw, "\n") {
		rawLine = strings.TrimSpace(rawLine)
		matches := lrcRegex.FindStringSubmatch(rawLine)
		if len(matches) == 5 {
			min, _ := strconv.Atoi(matches[1])
			sec, _ := strconv.Atoi(matches[2])
			msRaw := matches[3]
			if len(msRaw) == 2 {
				msRaw += "0"
			}
			ms, _ := strconv.Atoi(msRaw)

			totalMs := min*60000 + sec*1000 + ms
			text := strings.TrimSpace(matches[4])
			lines = append(lines, Line{
				TimeMs: totalMs,
				Text:   text,
			})
		}
	}

	sort.Slice(lines, func(i, j int) bool {
		return lines[i].TimeMs < lines[j].TimeMs
	})

	return lines
}

func FindActiveIndex(lines []Line, progressMs int) int {
	if len(lines) == 0 {
		return -1
	}
	active := 0
	for i, l := range lines {
		if l.TimeMs <= progressMs {
			active = i
		} else {
			break
		}
	}
	return active
}
