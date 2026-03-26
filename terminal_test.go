package jid

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
	"github.com/stretchr/testify/assert"
)

// makeCells converts a string to a slice of termbox.Cell with default colors.
func makeCells(s string) []termbox.Cell {
	runes := []rune(s)
	cells := make([]termbox.Cell, len(runes))
	for i, ch := range runes {
		cells[i] = termbox.Cell{Ch: ch, Fg: termbox.ColorDefault, Bg: termbox.ColorDefault}
	}
	return cells
}

func TestHighlightCandidateKeyFound(t *testing.T) {
	// Row: `  "name": "alice"`
	// "name" starts at rune index 2, length 6 (including quotes).
	row := `  "name": "alice"`
	cells := makeCells(row)
	result := highlightCandidateKey(cells, "name")

	// Positions 2–7 (`"name"`) should be highlighted.
	for i, c := range result {
		if i >= 2 && i <= 7 {
			assert.Equal(t, termbox.ColorYellow, c.Bg, "index %d should have yellow bg", i)
			assert.Equal(t, termbox.ColorBlack, c.Fg, "index %d should have black fg", i)
		} else {
			assert.Equal(t, termbox.ColorDefault, c.Bg, "index %d should have default bg", i)
		}
	}
}

func TestHighlightCandidateKeyNotFound(t *testing.T) {
	cells := makeCells(`  "name": "alice"`)
	result := highlightCandidateKey(cells, "missing")
	// Slice is returned unmodified (same contents).
	assert.Equal(t, cells, result)
}

func TestHighlightCandidateKeyValueNotHighlighted(t *testing.T) {
	// "name" appears only as a string value, not followed by ":"
	cells := makeCells(`  "url": "name"`)
	result := highlightCandidateKey(cells, "name")
	for _, c := range result {
		assert.Equal(t, termbox.ColorDefault, c.Bg, "value occurrence must not be highlighted")
	}
}

func TestHighlightCandidateKeyWithSpaceBeforeColon(t *testing.T) {
	// Some formatters emit `"key" : value`
	cells := makeCells(`  "age" : 30`)
	result := highlightCandidateKey(cells, "age")
	// "age" at positions 2–6 (5 chars including quotes)
	for i, c := range result {
		if i >= 2 && i <= 6 {
			assert.Equal(t, termbox.ColorYellow, c.Bg, "index %d should be highlighted", i)
		} else {
			assert.Equal(t, termbox.ColorDefault, c.Bg, "index %d should not be highlighted", i)
		}
	}
}

func TestHighlightCandidateKeyOriginalUnmodified(t *testing.T) {
	// Original slice must not be mutated.
	cells := makeCells(`  "id": 1`)
	orig := make([]termbox.Cell, len(cells))
	copy(orig, cells)
	highlightCandidateKey(cells, "id")
	assert.Equal(t, orig, cells, "original cells slice must not be mutated")
}
