package jid

import (
	"testing"

	termbox "github.com/nsf/termbox-go"
	"github.com/stretchr/testify/assert"
)

func TestParseKey(t *testing.T) {
	tests := []struct {
		input    string
		expected termbox.Key
		ok       bool
	}{
		{"ctrl+j", termbox.KeyCtrlJ, true},
		{"ctrl+k", termbox.KeyCtrlK, true},
		{"ctrl+a", termbox.KeyCtrlA, true},
		{"ctrl+z", termbox.KeyCtrlZ, true},
		{"up", termbox.KeyArrowUp, true},
		{"down", termbox.KeyArrowDown, true},
		{"left", termbox.KeyArrowLeft, true},
		{"right", termbox.KeyArrowRight, true},
		{"tab", termbox.KeyTab, true},
		{"enter", termbox.KeyEnter, true},
		{"esc", termbox.KeyEsc, true},
		{"backspace", termbox.KeyBackspace2, true},
		{"home", termbox.KeyHome, true},
		{"end", termbox.KeyEnd, true},
		{"pgup", termbox.KeyPgup, true},
		{"pgdn", termbox.KeyPgdn, true},
		{"f1", termbox.KeyF1, true},
		{"f12", termbox.KeyF12, true},
		// case insensitive
		{"CTRL+J", termbox.KeyCtrlJ, true},
		{"UP", termbox.KeyArrowUp, true},
		// with whitespace
		{"  ctrl+j  ", termbox.KeyCtrlJ, true},
		// invalid
		{"invalid", 0, false},
		{"ctrl+1", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		k, ok := ParseKey(tt.input)
		assert.Equal(t, tt.ok, ok, "ParseKey(%q) ok", tt.input)
		if tt.ok {
			assert.Equal(t, tt.expected, k, "ParseKey(%q) key", tt.input)
		}
	}
}

func TestResolveKey(t *testing.T) {
	// valid configured key
	assert.Equal(t, termbox.KeyCtrlJ, resolveKey("ctrl+j", "ctrl+k"))

	// invalid configured key → fall back to default
	assert.Equal(t, termbox.KeyCtrlK, resolveKey("invalid", "ctrl+k"))

	// empty configured key → fall back to default
	assert.Equal(t, termbox.KeyArrowUp, resolveKey("", "up"))
}
