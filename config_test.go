package jid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	assert.Equal(t, historyMaxSize, cfg.History.MaxSize)
	assert.Equal(t, "", cfg.History.Path)
	assert.Equal(t, "tab", cfg.Keybindings.CandidateNext)
	assert.Equal(t, "", cfg.Keybindings.CandidatePrev)
	assert.Equal(t, "up", cfg.Keybindings.HistoryPrev)
	assert.Equal(t, "down", cfg.Keybindings.HistoryNext)
	assert.Equal(t, "ctrl+j", cfg.Keybindings.ScrollDown)
	assert.Equal(t, "ctrl+k", cfg.Keybindings.ScrollUp)
	assert.Equal(t, "ctrl+x", cfg.Keybindings.ToggleFuncHelp)
}

func TestLoadConfigMissingFile(t *testing.T) {
	// point config path at a nonexistent file via env manipulation
	// We test via a helper that accepts a path
	cfg := loadConfigFromPath("/nonexistent/path/config.toml")
	assert.Equal(t, defaultConfig(), cfg)
}

func TestLoadConfigFull(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[history]
path = "/tmp/my_history"
max_size = 500

[keybindings]
candidate_next  = "f1"
candidate_prev  = "f2"
history_prev    = "ctrl+p"
history_next    = "ctrl+n"
scroll_down     = "ctrl+j"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	cfg := loadConfigFromPath(path)
	assert.Equal(t, "/tmp/my_history", cfg.History.Path)
	assert.Equal(t, 500, cfg.History.MaxSize)
	assert.Equal(t, "f1", cfg.Keybindings.CandidateNext)
	assert.Equal(t, "f2", cfg.Keybindings.CandidatePrev)
	assert.Equal(t, "ctrl+p", cfg.Keybindings.HistoryPrev)
	assert.Equal(t, "ctrl+n", cfg.Keybindings.HistoryNext)
	// unspecified fields keep defaults
	assert.Equal(t, "ctrl+x", cfg.Keybindings.ToggleFuncHelp)
}

func TestLoadConfigPartialOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	content := `
[keybindings]
scroll_down = "ctrl+d"
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	cfg := loadConfigFromPath(path)
	assert.Equal(t, "ctrl+d", cfg.Keybindings.ScrollDown)
	// all other fields keep defaults
	assert.Equal(t, "tab", cfg.Keybindings.CandidateNext)
	assert.Equal(t, "up", cfg.Keybindings.HistoryPrev)
	assert.Equal(t, historyMaxSize, cfg.History.MaxSize)
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte("not valid toml :::"), 0644))

	cfg := loadConfigFromPath(path)
	assert.Equal(t, defaultConfig(), cfg)
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, ".jid_history"), expandPath("~/.jid_history"))
	assert.Equal(t, "/absolute/path", expandPath("/absolute/path"))
	assert.Equal(t, "relative/path", expandPath("relative/path"))
}

func TestHistoryPathFromConfig(t *testing.T) {
	cfg := defaultConfig()

	// no path configured → OS default
	assert.Equal(t, historyFilePath(), cfg.HistoryPath())

	// custom path
	cfg.History.Path = "/custom/history"
	assert.Equal(t, "/custom/history", cfg.HistoryPath())

	// tilde path
	home, _ := os.UserHomeDir()
	cfg.History.Path = "~/.jid_history"
	assert.Equal(t, filepath.Join(home, ".jid_history"), cfg.HistoryPath())
}

func TestMergeKeybindingsEmpty(t *testing.T) {
	dst := defaultConfig().Keybindings
	mergeKeybindings(&dst, KeybindingsConfig{})
	// nothing should change
	assert.Equal(t, defaultConfig().Keybindings, dst)
}
