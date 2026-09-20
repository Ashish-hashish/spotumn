package lyrics

import (
	"testing"
)

func TestParseLRC(t *testing.T) {
	raw := `
[00:12.50]Line 1 lyric
[00:15.80]Line 2 lyric
[01:05.123]Line 3 lyric with three decimal places
[02:00.00]Final line
`
	lines := parseLRC(raw)
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	if lines[0].TimeMs != 12500 {
		t.Errorf("line 0 expected 12500ms, got %d", lines[0].TimeMs)
	}
	if lines[0].Text != "Line 1 lyric" {
		t.Errorf("line 0 text unexpected: %s", lines[0].Text)
	}

	if lines[2].TimeMs != 65123 {
		t.Errorf("line 2 expected 65123ms, got %d", lines[2].TimeMs)
	}
}

func TestFindActiveIndex(t *testing.T) {
	lines := []Line{
		{TimeMs: 1000, Text: "First"},
		{TimeMs: 5000, Text: "Second"},
		{TimeMs: 10000, Text: "Third"},
	}

	if idx := FindActiveIndex(lines, 500); idx != 0 {
		t.Errorf("expected 0 before first timestamp, got %d", idx)
	}

	if idx := FindActiveIndex(lines, 3000); idx != 0 {
		t.Errorf("expected 0 for 3000ms, got %d", idx)
	}

	if idx := FindActiveIndex(lines, 5000); idx != 1 {
		t.Errorf("expected 1 for 5000ms, got %d", idx)
	}

	if idx := FindActiveIndex(lines, 7500); idx != 1 {
		t.Errorf("expected 1 for 7500ms, got %d", idx)
	}

	if idx := FindActiveIndex(lines, 15000); idx != 2 {
		t.Errorf("expected 2 for 15000ms, got %d", idx)
	}
}
