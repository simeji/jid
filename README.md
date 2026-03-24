# jid

[![Circle CI](https://circleci.com/gh/simeji/jid/tree/master.svg?style=shield)](https://circleci.com/gh/simeji/jid/tree/master)

Json Incremental Digger

It's a very simple tool.
You can drill down JSON interactively by using filtering queries like [jq](https://stedolan.github.io/jq/).

**Suggestion**, **Auto completion**, and **JMESPath** support provide a comfortable JSON exploration experience.

## Demo

![demo-jid-main](https://github.com/simeji/jid/wiki/images/demo-jid-main-640-colorize.gif)

## Installation

* [With HomeBrew (for macOS)](#with-homebrew-for-macos)
* [With MacPorts (for macOS)](#with-macports-for-macos)
* [With pkg (for FreeBSD)](#with-pkg-for-freebsd)
* [With scoop (for Windows)](#with-scoop-for-windows)
* [Other package management system](#other-package-management-systems)
* [Simply use "jid" command](#simply-use-jid-command)
* [Build](#build)

### With HomeBrew (for macOS)

```
brew install jid
```

### With MacPorts (for macOS)

```
sudo port install jid
```

### With pkg (for FreeBSD)

```
pkg install jid
```

### With scoop (for Windows)

```
scoop install jid
```

### Other package management systems

Jid can install by package management systems of below OS.

[![Packaging status](https://repology.org/badge/vertical-allrepos/jid.svg)](https://repology.org/metapackage/jid/versions)


### Simply use "jid" command

If you simply want to use `jid` command, please download binary from below.

https://github.com/simeji/jid/releases

## Build

```
go install github.com/simeji/jid/cmd/jid@latest
```

## Usage

### Quick start

* [simple json example](#simple-json-example)
* [simple json example2](#simple-json-example2)
* [with initial query](#with-initial-query)
* [with curl](#with-curl)

#### simple json example

Please execute the below command.

```
echo '{"aa":"2AA2","bb":{"aaa":[123,"cccc",[1,2]],"c":321}}'| jid
```

then, jid will be running.

You can dig JSON data incrementally.

When you enter `.bb.aaa[2]`, you will see the following.

```
[Filter]> .bb.aaa[2]
[
  1,
  2
]
```

Then, you press Enter key and output `[1,2]` and exit.

#### simple json example2

This json is used by [demo section](https://github.com/simeji/jid#demo).
```
echo '{"info":{"date":"2016-10-23","version":1.0},"users":[{"name":"simeji","uri":"https://github.com/simeji","id":1},{"name":"simeji2","uri":"https://example.com/simeji","id":2},{"name":"simeji3","uri":"https://example.com/simeji3","id":3}],"userCount":3}}'|jid
```

#### With a initial query

First argument of `jid` is initial query.
(Use JSON same as [Demo](#demo))

![demo-jid-with-query](https://github.com/simeji/jid/wiki/images/demo-jid-with-query-640.gif)

#### with curl

Sample for using [RDAP](https://datatracker.ietf.org/wg/weirds/documents/) data.

```
curl -s http://rdg.afilias.info/rdap/domain/example.info | jid
```

#### Load JSON from a file

```
jid < file.json
```

## Keymaps

|key|description|
|:-----------|:----------|
|`TAB` / `CTRL` + `I` |Show available items and choose them (cycles forward)|
|`Shift` + `TAB` |Cycle candidates backward / decrement array index|
|`CTRL` + `W` |Delete one JMESPath segment backward (e.g. `.id` → `[0]` → `func(@)` → pipe)|
|`CTRL` + `U` |Delete whole query|
|`CTRL` + `X` |Toggle function description display (visible when function candidates are shown)|
|`CTRL` + `F` / Right Arrow (:arrow_right:)|Move cursor a character to the right|
|`CTRL` + `B` / Left Arrow (:arrow_left:)|Move cursor a character to the left|
|`CTRL` + `A`|To the first character of the 'Filter'|
|`CTRL` + `E`|To the end of the 'Filter'|
|`CTRL` + `J`|Scroll json buffer 1 line downwards|
|`CTRL` + `K`|Scroll json buffer 1 line upwards|
|`CTRL` + `G`|Scroll json buffer to bottom|
|`CTRL` + `T`|Scroll json buffer to top|
|`CTRL` + `N`|Scroll json buffer 'Page Down'|
|`CTRL` + `P`|Scroll json buffer 'Page Up'|
|`CTRL` + `L`|Change view mode whole json or keys (only object)|
|`ESC`|Hide a candidate box|
|Up Arrow|Navigate to previous query in history|
|Down Arrow|Navigate to next query in history|

### Option

|option|description|
|:-----------|:----------|
|First argument ($1) | Initial query|
|-h | print a help|
|-help | print a help|
|-version | print the version and exit|
|-q | Output query mode (for jq)|
|-M | monochrome output mode|

## Configuration

jid can be configured via a TOML file located at:

| OS | Path |
|:---|:-----|
| macOS | `~/Library/Application Support/jid/config.toml` |
| Linux | `~/.config/jid/config.toml` |
| Windows | `%AppData%\jid\config.toml` |

### Example config.toml

```toml
[history]
path = "~/.jid_history"  # custom history file path
max_size = 1000           # number of entries to keep

[keybindings]
history_prev    = "up"      # navigate to older query
history_next    = "down"    # navigate to newer query
scroll_down     = "ctrl+j"
scroll_up       = "ctrl+k"
scroll_to_bottom = "ctrl+g"
scroll_to_top   = "ctrl+t"
scroll_page_down = "ctrl+n"
scroll_page_up  = "ctrl+p"
toggle_keymode  = "ctrl+l"
delete_line     = "ctrl+u"
delete_word     = "ctrl+w"
cursor_left     = "ctrl+b"
cursor_right    = "ctrl+f"
cursor_to_start = "ctrl+a"
cursor_to_end   = "ctrl+e"
toggle_func_help = "ctrl+x"
candidate_next  = "tab"       # cycle candidates forward
candidate_prev  = "ctrl+p"    # cycle candidates backward (additional key; Shift+Tab always works)
```

> **Note:** Shift+Tab (`\x1b[Z`) is a fixed terminal escape sequence and always triggers backward cycling regardless of `candidate_prev`.

### Supported key strings

`ctrl+a` … `ctrl+z`, `up`, `down`, `left`, `right`, `tab`, `enter`, `esc`, `backspace`, `home`, `end`, `pgup`, `pgdn`, `delete`, `f1` … `f12`

### Query History

Queries are saved automatically on Enter. The history file path follows the same OS convention as the config file (e.g. `~/Library/Application Support/jid/history` on macOS) unless overridden in `config.toml`.

## JMESPath Support

jid supports [JMESPath](https://jmespath.org/) expressions in addition to the traditional dot-path notation.
JMESPath mode is automatically activated when the query contains pipe (`|`), wildcards (`[*]`), filter expressions (`[?`), or function calls.

### JMESPath Query Examples

```
.                          traditional: show root JSON
.users                     traditional: navigate to users field
.users[0].name             traditional: array index + field access

.users[*].name             wildcard projection: extract name from every user
.users[*].address.city     nested wildcard projection
.users[*].<Tab>            show field candidates from array elements

. | keys(@)                pipe: list root object keys
.users | length(@)         pipe: count users array
.users | sort_by(@, &name) pipe: sort users by name field
.users | reverse(@)        pipe: reverse the array

.[1] | to_array(@)[0].id   chained pipe with indexing
. | to_array(@)[0]         wrap root in array and index

.users[*].name | [0]       project names then index
```

### Wildcard Projection + Array Index

After a wildcard projection like `.game_indices[*].version`, the result is an array.
Use `[N]` to navigate into it — jid automatically rewrites to pipe form internally:

```
.game_indices[*]           → field candidates: game_index, version
.game_indices[*].version   → shows array of version objects; suggests [
.game_indices[*].version[0]           → first version object {name, url}
.game_indices[*].version[0].name      → first version's name
.game_indices[*].version[0] | keys(@) → keys of first version object
.game_indices[*].version[0] | keys(@) | sort(@)  → sorted keys
```

> **Note**: In standard JMESPath, `[*].field[0]` applies `[0]` to each projected
> element rather than the projected array, producing `[]`. jid detects this pattern
> and transparently rewrites it to `[*].field | [0]` so `[0]` indexes the array.

### Function Candidates

When you type `|` after a field, jid shows available JMESPath functions filtered by the type of the preceding expression:

| Input type | Suggested functions |
|:-----------|:-------------------|
| Array | `avg`, `contains`, `join`, `length`, `map`, `max`, `max_by`, `min`, `min_by`, `not_null`, `reverse`, `sort`, `sort_by`, `sum`, `to_array`, `to_string`, `type` |
| Object | `keys`, `length`, `merge`, `not_null`, `to_array`, `to_string`, `type`, `values` |
| String | `contains`, `ends_with`, `length`, `not_null`, `reverse`, `starts_with`, `to_array`, `to_number`, `to_string`, `type` |
| Number | `abs`, `ceil`, `floor`, `not_null`, `to_array`, `to_string`, `type` |

A usage description is shown below the candidate list (toggle with `Ctrl+X`).

### Function Argument Templates

When a function candidate is confirmed, the arguments are automatically filled in and the cursor is placed at the right position:

| Function | Inserted as | Cursor position |
|:---------|:-----------|:----------------|
| `contains` | `contains(@, '')` | inside `''` |
| `ends_with` | `ends_with(@, '')` | inside `''` |
| `starts_with` | `starts_with(@, '')` | inside `''` |
| `join` | `join('', @)` | inside `''` (separator) |
| `sort_by` | `sort_by(@, &field)` | on `field` placeholder |
| `max_by` | `max_by(@, &field)` | on `field` placeholder |
| `map` | `map(&expr, @)` | on `expr` placeholder |

Placeholder text is shown in blue. Typing any character replaces the entire placeholder.

### Wildcard Projection Navigation

After a wildcard expression like `.game_indices[*]`, jid shows the field names of the array elements as candidates:

```
.game_indices[*]           → candidates: game_index, version
.game_indices[*].<Tab>     → same candidates (trailing dot still shows fields)
.game_indices[*].v<Tab>    → filtered: version
.game_indices[*].version   → shows array result; suggests [ for index navigation
.game_indices[*].version[0] → first version object; candidates: name, url
```

### Ctrl+W in JMESPath Mode

`Ctrl+W` removes one segment at a time from the end of a JMESPath expression:

```
.[3] | to_array(@)[0].id  →(Ctrl+W)→  .[3] | to_array(@)[0]
.[3] | to_array(@)[0]     →(Ctrl+W)→  .[3] | to_array(@)
.[3] | to_array(@)        →(Ctrl+W)→  .[3] |
.[3] |                    →(Ctrl+W)→  .[3]
```
