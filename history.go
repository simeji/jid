package jid

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

const historyMaxSize = 1000

type History struct {
	entries []string
	idx     int
	maxSize int
	path    string
}

func historyFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.Getenv("HOME")
	}
	return filepath.Join(dir, "jid", "history")
}

func NewHistory() *History {
	h := &History{
		entries: []string{},
		maxSize: historyMaxSize,
		path:    historyFilePath(),
	}
	_ = h.load()
	h.idx = len(h.entries)
	return h
}

// Add adds a query to history with deduplication (moves existing entry to end).
func (h *History) Add(query string) {
	if query == "" || query == "." {
		return
	}
	for i, e := range h.entries {
		if e == query {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}
	h.entries = append(h.entries, query)
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
	h.idx = len(h.entries)
}

// AtEnd returns true when the navigation cursor is at the newest position.
func (h *History) AtEnd() bool {
	return h.idx >= len(h.entries)
}

// Prev moves to the previous (older) entry and returns it.
func (h *History) Prev() (string, bool) {
	if len(h.entries) == 0 || h.idx == 0 {
		return "", false
	}
	h.idx--
	return h.entries[h.idx], true
}

// Next moves to the next (newer) entry and returns it.
// Returns "", false when moving past the newest entry (back to current input).
func (h *History) Next() (string, bool) {
	if h.idx >= len(h.entries) {
		return "", false
	}
	h.idx++
	if h.idx == len(h.entries) {
		return "", false
	}
	return h.entries[h.idx], true
}

// ResetIdx resets navigation to the newest position.
func (h *History) ResetIdx() {
	h.idx = len(h.entries)
}

func (h *History) load() error {
	f, err := os.Open(h.path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}
	if len(h.entries) > h.maxSize {
		h.entries = h.entries[len(h.entries)-h.maxSize:]
	}
	return scanner.Err()
}

func (h *History) Save() error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
		return err
	}
	f, err := os.Create(h.path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, e := range h.entries {
		fmt.Fprintln(w, e)
	}
	return w.Flush()
}
