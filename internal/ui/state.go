package ui

type FocusedPane int

const (
	PaneNav FocusedPane = iota
	PaneCenter
	PaneRight
	PanePlayer
)

type CenterTab int

const (
	TabTracks CenterTab = iota
	TabLyrics
	TabHistory
)

func (t CenterTab) Title() string {
	switch t {
	case TabTracks:
		return "Tracks"
	case TabLyrics:
		return "Lyrics"
	case TabHistory:
		return "History"
	default:
		return ""
	}
}
