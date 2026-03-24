package jid

import (
	"strings"

	termbox "github.com/nsf/termbox-go"
)

// stringToKey maps human-readable key strings to termbox.Key values.
var stringToKey = map[string]termbox.Key{
	// Ctrl combinations
	"ctrl+a": termbox.KeyCtrlA,
	"ctrl+b": termbox.KeyCtrlB,
	"ctrl+c": termbox.KeyCtrlC,
	"ctrl+d": termbox.KeyCtrlD,
	"ctrl+e": termbox.KeyCtrlE,
	"ctrl+f": termbox.KeyCtrlF,
	"ctrl+g": termbox.KeyCtrlG,
	"ctrl+h": termbox.KeyCtrlH,
	"ctrl+i": termbox.KeyTab,
	"ctrl+j": termbox.KeyCtrlJ,
	"ctrl+k": termbox.KeyCtrlK,
	"ctrl+l": termbox.KeyCtrlL,
	"ctrl+m": termbox.KeyEnter,
	"ctrl+n": termbox.KeyCtrlN,
	"ctrl+o": termbox.KeyCtrlO,
	"ctrl+p": termbox.KeyCtrlP,
	"ctrl+q": termbox.KeyCtrlQ,
	"ctrl+r": termbox.KeyCtrlR,
	"ctrl+s": termbox.KeyCtrlS,
	"ctrl+t": termbox.KeyCtrlT,
	"ctrl+u": termbox.KeyCtrlU,
	"ctrl+v": termbox.KeyCtrlV,
	"ctrl+w": termbox.KeyCtrlW,
	"ctrl+x": termbox.KeyCtrlX,
	"ctrl+y": termbox.KeyCtrlY,
	"ctrl+z": termbox.KeyCtrlZ,
	// Arrow keys
	"up":    termbox.KeyArrowUp,
	"down":  termbox.KeyArrowDown,
	"left":  termbox.KeyArrowLeft,
	"right": termbox.KeyArrowRight,
	// Special keys
	"tab":       termbox.KeyTab,
	"enter":     termbox.KeyEnter,
	"esc":       termbox.KeyEsc,
	"backspace": termbox.KeyBackspace2,
	"home":      termbox.KeyHome,
	"end":       termbox.KeyEnd,
	"pgup":      termbox.KeyPgup,
	"pgdn":      termbox.KeyPgdn,
	"delete":    termbox.KeyDelete,
	"insert":    termbox.KeyInsert,
	// Function keys
	"f1":  termbox.KeyF1,
	"f2":  termbox.KeyF2,
	"f3":  termbox.KeyF3,
	"f4":  termbox.KeyF4,
	"f5":  termbox.KeyF5,
	"f6":  termbox.KeyF6,
	"f7":  termbox.KeyF7,
	"f8":  termbox.KeyF8,
	"f9":  termbox.KeyF9,
	"f10": termbox.KeyF10,
	"f11": termbox.KeyF11,
	"f12": termbox.KeyF12,
}

// ParseKey converts a key string (e.g. "ctrl+j", "up", "f5") to a termbox.Key.
// Returns 0, false if the string is not recognized.
func ParseKey(s string) (termbox.Key, bool) {
	k, ok := stringToKey[strings.ToLower(strings.TrimSpace(s))]
	return k, ok
}

// resolveKey parses a configured key string, falling back to defaultStr on failure.
func resolveKey(configured, defaultStr string) termbox.Key {
	if k, ok := ParseKey(configured); ok {
		return k
	}
	k, _ := ParseKey(defaultStr)
	return k
}
