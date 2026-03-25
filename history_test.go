package jid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHistory(t *testing.T) {
	h := NewHistory("", historyMaxSize)
	assert.NotNil(t, h)
	assert.Equal(t, historyMaxSize, h.maxSize)
	assert.True(t, h.AtEnd())
}

func TestHistoryAddBasic(t *testing.T) {
	h := NewHistory("", 100)
	h.Add(".users")
	h.Add(".users[0].name")
	assert.Equal(t, 2, len(h.entries))
	assert.Equal(t, ".users[0].name", h.entries[1])
}

func TestHistoryAddIgnoresEmpty(t *testing.T) {
	h := NewHistory("", 100)
	h.Add("")
	h.Add(".")
	assert.Equal(t, 0, len(h.entries))
}

func TestHistoryAddDeduplication(t *testing.T) {
	h := NewHistory("", 100)
	h.Add(".users")
	h.Add(".name")
	h.Add(".users") // duplicate → moved to end
	assert.Equal(t, 2, len(h.entries))
	assert.Equal(t, ".name", h.entries[0])
	assert.Equal(t, ".users", h.entries[1])
}

func TestHistoryAddMaxSize(t *testing.T) {
	h := NewHistory("", 3)
	h.Add(".a")
	h.Add(".b")
	h.Add(".c")
	h.Add(".d") // exceeds max → oldest trimmed
	assert.Equal(t, 3, len(h.entries))
	assert.Equal(t, ".b", h.entries[0])
	assert.Equal(t, ".d", h.entries[2])
}

func TestHistoryPrevNext(t *testing.T) {
	h := NewHistory("", 100)
	h.Add(".a")
	h.Add(".b")
	h.Add(".c")

	// navigate backward
	entry, ok := h.Prev()
	assert.True(t, ok)
	assert.Equal(t, ".c", entry)

	entry, ok = h.Prev()
	assert.True(t, ok)
	assert.Equal(t, ".b", entry)

	entry, ok = h.Prev()
	assert.True(t, ok)
	assert.Equal(t, ".a", entry)

	// at beginning: Prev returns false
	_, ok = h.Prev()
	assert.False(t, ok)

	// navigate forward
	entry, ok = h.Next()
	assert.True(t, ok)
	assert.Equal(t, ".b", entry)

	entry, ok = h.Next()
	assert.True(t, ok)
	assert.Equal(t, ".c", entry)

	// past end: Next returns false
	_, ok = h.Next()
	assert.False(t, ok)
	assert.True(t, h.AtEnd())
}

func TestHistoryAtEnd(t *testing.T) {
	h := NewHistory("", 100)
	assert.True(t, h.AtEnd())

	h.Add(".users")
	assert.True(t, h.AtEnd())

	h.Prev()
	assert.False(t, h.AtEnd())

	h.ResetIdx()
	assert.True(t, h.AtEnd())
}

func TestHistorySaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")

	h1 := NewHistory(path, 100)
	h1.Add(".users")
	h1.Add(".users[0].name")
	require.NoError(t, h1.Save())

	h2 := NewHistory(path, 100)
	assert.Equal(t, 2, len(h2.entries))
	assert.Equal(t, ".users", h2.entries[0])
	assert.Equal(t, ".users[0].name", h2.entries[1])
}

func TestHistorySaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "history")

	h := NewHistory(path, 100)
	h.Add(".test")
	require.NoError(t, h.Save())

	_, err := os.Stat(path)
	assert.NoError(t, err)
}

func TestHistoryLoadMaxSize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")

	// save 5 entries
	h1 := NewHistory(path, 100)
	for _, q := range []string{".a", ".b", ".c", ".d", ".e"} {
		h1.Add(q)
	}
	require.NoError(t, h1.Save())

	// load with max_size=3 → keeps last 3
	h2 := NewHistory(path, 3)
	assert.Equal(t, 3, len(h2.entries))
	assert.Equal(t, ".c", h2.entries[0])
	assert.Equal(t, ".e", h2.entries[2])
}
