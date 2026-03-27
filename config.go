package jid

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds all jid configuration.
type Config struct {
	History     HistoryConfig     `toml:"history"`
	Keybindings KeybindingsConfig `toml:"keybindings"`
}

// HistoryConfig controls query history behaviour.
type HistoryConfig struct {
	Path    string `toml:"path"`
	MaxSize int    `toml:"max_size"`
}

// KeybindingsConfig maps action names to key strings (e.g. "ctrl+j", "up").
type KeybindingsConfig struct {
	HistoryPrev    string `toml:"history_prev"`
	HistoryNext    string `toml:"history_next"`
	CandidateNext  string `toml:"candidate_next"`  // cycle forward (default: tab)
	CandidatePrev  string `toml:"candidate_prev"`  // cycle backward (additional key; Shift+Tab always works)
	ScrollDown     string `toml:"scroll_down"`
	ScrollUp       string `toml:"scroll_up"`
	ScrollToBottom string `toml:"scroll_to_bottom"`
	ScrollToTop    string `toml:"scroll_to_top"`
	ScrollPageDown string `toml:"scroll_page_down"`
	ScrollPageUp   string `toml:"scroll_page_up"`
	ToggleKeymode  string `toml:"toggle_keymode"`
	DeleteLine     string `toml:"delete_line"`
	DeleteWord     string `toml:"delete_word"`
	CursorLeft     string `toml:"cursor_left"`
	CursorRight    string `toml:"cursor_right"`
	CursorToStart  string `toml:"cursor_to_start"`
	CursorToEnd    string `toml:"cursor_to_end"`
	ToggleFuncHelp string `toml:"toggle_func_help"`
}

func defaultConfig() Config {
	return Config{
		History: HistoryConfig{
			Path:    "",
			MaxSize: historyMaxSize,
		},
		Keybindings: KeybindingsConfig{
			CandidateNext:  "tab",
			CandidatePrev:  "",
			HistoryPrev:    "up",
			HistoryNext:    "down",
			ScrollDown:     "ctrl+j",
			ScrollUp:       "ctrl+k",
			ScrollToBottom: "ctrl+g",
			ScrollToTop:    "ctrl+t",
			ScrollPageDown: "ctrl+n",
			ScrollPageUp:   "ctrl+p",
			ToggleKeymode:  "ctrl+l",
			DeleteLine:     "ctrl+u",
			DeleteWord:     "ctrl+w",
			CursorLeft:     "ctrl+b",
			CursorRight:    "ctrl+f",
			CursorToStart:  "ctrl+a",
			CursorToEnd:    "ctrl+e",
			ToggleFuncHelp: "ctrl+x",
		},
	}
}

func configFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.Getenv("HOME")
	}
	return filepath.Join(dir, "jid", "config.toml")
}

// HistoryPath returns the resolved history file path.
// Uses the configured path if set, otherwise falls back to the OS default.
func (c *Config) HistoryPath() string {
	if c.History.Path != "" {
		return expandPath(c.History.Path)
	}
	return historyFilePath()
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// LoadConfig reads ~/.config/jid/config.toml (or OS equivalent).
// Missing fields fall back to defaults; missing file returns defaults silently.
func LoadConfig() Config {
	return loadConfigFromPath(configFilePath())
}

func loadConfigFromPath(path string) Config {
	cfg := defaultConfig()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg
	}
	var fileCfg Config
	if _, err := toml.DecodeFile(path, &fileCfg); err != nil {
		return cfg
	}
	// Merge: only override fields that were explicitly set
	if fileCfg.History.Path != "" {
		cfg.History.Path = fileCfg.History.Path
	}
	if fileCfg.History.MaxSize > 0 {
		cfg.History.MaxSize = fileCfg.History.MaxSize
	}
	mergeKeybindings(&cfg.Keybindings, fileCfg.Keybindings)
	return cfg
}

func mergeKeybindings(dst *KeybindingsConfig, src KeybindingsConfig) {
	if src.CandidateNext != "" {
		dst.CandidateNext = src.CandidateNext
	}
	if src.CandidatePrev != "" {
		dst.CandidatePrev = src.CandidatePrev
	}
	if src.HistoryPrev != "" {
		dst.HistoryPrev = src.HistoryPrev
	}
	if src.HistoryNext != "" {
		dst.HistoryNext = src.HistoryNext
	}
	if src.ScrollDown != "" {
		dst.ScrollDown = src.ScrollDown
	}
	if src.ScrollUp != "" {
		dst.ScrollUp = src.ScrollUp
	}
	if src.ScrollToBottom != "" {
		dst.ScrollToBottom = src.ScrollToBottom
	}
	if src.ScrollToTop != "" {
		dst.ScrollToTop = src.ScrollToTop
	}
	if src.ScrollPageDown != "" {
		dst.ScrollPageDown = src.ScrollPageDown
	}
	if src.ScrollPageUp != "" {
		dst.ScrollPageUp = src.ScrollPageUp
	}
	if src.ToggleKeymode != "" {
		dst.ToggleKeymode = src.ToggleKeymode
	}
	if src.DeleteLine != "" {
		dst.DeleteLine = src.DeleteLine
	}
	if src.DeleteWord != "" {
		dst.DeleteWord = src.DeleteWord
	}
	if src.CursorLeft != "" {
		dst.CursorLeft = src.CursorLeft
	}
	if src.CursorRight != "" {
		dst.CursorRight = src.CursorRight
	}
	if src.CursorToStart != "" {
		dst.CursorToStart = src.CursorToStart
	}
	if src.CursorToEnd != "" {
		dst.CursorToEnd = src.CursorToEnd
	}
	if src.ToggleFuncHelp != "" {
		dst.ToggleFuncHelp = src.ToggleFuncHelp
	}
}
