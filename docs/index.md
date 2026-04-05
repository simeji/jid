---
layout: splash
title: "jid — JSON Incremental Digger"
header:
  overlay_color: "#1a1a2e"
  overlay_filter: "0.6"
  actions:
    - label: "GitHub"
      url: "https://github.com/simeji/jid"
    - label: "Download v1.1.1"
      url: "https://github.com/simeji/jid/releases/tag/v1.1.1"
excerpt: >
  Interactively drill down JSON in your terminal.
  JMESPath support, query history, key highlighting, and more.

feature_row:
  - title: "JMESPath Support"
    excerpt: "Use wildcards `[*]`, pipes `|`, and functions like `length(@)`, `keys(@)`, `sort_by(@, &field)` with type-aware completion."
  - title: "Key Highlighting"
    excerpt: "Matching JSON keys highlight in yellow as you type or Tab through candidates. Auto-scrolls to keep the key in view."
  - title: "Query History"
    excerpt: "Navigate past queries with Up/Down arrows. History persists across sessions and is configurable via `config.toml`."

feature_row2:
  - title: "Tab Completion"
    excerpt: "Press Tab to cycle through field candidates. Shift+Tab to go backwards. Supports arrays, objects, wildcards, and pipe expressions."
  - title: "TOML Config"
    excerpt: "Customize keybindings, history path, and behavior via `~/.config/jid/config.toml`. Set `exit_on_enter = false` to avoid accidental exits."
  - title: "Pretty Output"
    excerpt: "Results are pretty-printed with color. Use `-M` for monochrome, `-p` for pretty JSON output, `-q` for query-only mode."
---

<div style="text-align:center; margin: 2em 0;">
  <img src="https://github.com/user-attachments/assets/78b0ee4c-2171-4557-b690-50d503f91190" alt="jid logo" width="140" />
</div>

{% include feature_row %}

---

## Demo

### Drill-down Navigation

![drill-down demo](https://github.com/user-attachments/assets/b37e5a62-e9e4-4ec5-9cc8-e4ca9180a744)

### JMESPath Functions

![jmespath demo](https://github.com/user-attachments/assets/d6cb5cc7-4e66-4b66-b7bb-c4cc50e8b050)

### With Initial Query

![initial query demo](https://github.com/user-attachments/assets/d4ef1067-ccd1-401d-8cc3-eef4eef96109)

---

## Install

```bash
# Homebrew (macOS / Linux)
brew install jid

# Go
go install github.com/simeji/jid/cmd/jid@latest

# Download binary
# https://github.com/simeji/jid/releases
```

---

## Quick Start

```bash
# Pipe JSON into jid
echo '{"users":[{"name":"alice"},{"name":"bob"}]}' | jid

# Load from file
jid < data.json

# Start with an initial query
jid '.users[0]' < data.json

# Query-only output (for use with jq)
jid -q < data.json

# Pretty-print result
jid -p < data.json
```

---

## Key Bindings

| Key | Action |
|:----|:-------|
| `Tab` | Cycle candidates forward / increment array index |
| `Shift+Tab` | Cycle candidates backward / decrement array index |
| `Up` / `Down` | Navigate query history |
| `Enter` | Confirm query and exit |
| `Ctrl+C` | Exit without output |
| `Ctrl+W` | Delete word (JMESPath-aware) |
| `Ctrl+U` | Clear query |
| `Ctrl+J` / `Ctrl+K` | Scroll JSON down / up |
| `Ctrl+N` / `Ctrl+P` | Page down / up |
| `Ctrl+L` | Toggle key-only view |
| `Ctrl+X` | Toggle function description |
| `Esc` | Hide candidate list |

---

## Configuration

Create `~/.config/jid/config.toml` (macOS: `~/Library/Application Support/jid/config.toml`):

```toml
[behavior]
exit_on_enter = false   # prevent accidental exit on Enter

[history]
path = "~/.jid_history"
max_size = 1000

[keybindings]
quit         = "ctrl+q"   # exit key when exit_on_enter = false
candidate_next = "tab"
history_prev = "up"
history_next = "down"
scroll_down  = "ctrl+j"
scroll_up    = "ctrl+k"
```

---

{% include feature_row id="feature_row2" %}

---

<div style="text-align:center; margin-top: 2em;">
  <a href="https://github.com/simeji/jid" class="btn btn--primary btn--large">View on GitHub</a>
  <a href="https://github.com/simeji/jid/releases/tag/v1.1.1" class="btn btn--inverse btn--large">Download v1.1.1</a>
</div>
