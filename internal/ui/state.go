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
	TabHistory
	TabLyrics
)

func (t CenterTab) Title() string {
	switch t {
	case TabTracks:
		return "Tracks"
	case TabHistory:
		return "History"
	case TabLyrics:
		return "Lyrics"
	default:
		return ""
	}
}
