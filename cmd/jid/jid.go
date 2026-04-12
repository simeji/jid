package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/simeji/jid"
)

const VERSION = "1.1.1"

func main() {
	content := os.Stdin

	var qm bool
	var help bool
	var version bool
	var mono bool
	var pretty bool
	qs := "."

	flag.BoolVar(&qm, "q", false, "Output query mode")
	flag.BoolVar(&help, "h", false, "print a help")
	flag.BoolVar(&help, "help", false, "print a help")
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.BoolVar(&mono, "M", false, "monochrome output mode")
	flag.BoolVar(&pretty, "p", false, "pretty print json result")
	flag.Parse()

	if help {
		flag.Usage()
		fmt.Println(getHelpString())
		os.Exit(0)
	}

	if version {
		fmt.Println(fmt.Sprintf("jid version v%s", VERSION))
		os.Exit(0)
	}
	args := flag.Args()
	if len(args) > 0 {
		qs = args[0]
	}

	ea := &jid.EngineAttribute{
		DefaultQuery: qs,
		Monochrome:   mono,
		PrettyResult: pretty,
	}

	e, err := jid.NewEngine(content, ea)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(run(e, qm))
}

func run(e jid.EngineInterface, qm bool) int {

	result := e.Run()
	if result.GetError() != nil {
		return 2
	}
	if qm {
		fmt.Printf("%s", result.GetQueryString())
	} else {
		fmt.Printf("%s", result.GetContent())
	}
	return 0
}

func getHelpString() string {
	return `

============ Load JSON from a file ==============

$ jid < file.json

============ With a JSON filter mode =============

TAB / CTRL-I
  Show available candidates and cycle forward.
  In JMESPath pipe mode, shows field or function candidates.
  Confirms immediately when only one candidate matches.

Shift-TAB
  Cycle candidates backward.
  Outside candidate mode: decrement the last array index.

Enter
  Exit jid and print the current result.
  If a candidate list is open, confirm the selected candidate instead.
  Set exit_on_enter = false in config.toml to disable accidental exit.

CTRL-W
  Delete one segment backward (JMESPath-aware).
  Removes the last field/index/function one step at a time.
  Inside &field) argument: deletes the field name but keeps '&'.

CTRL-U
  Delete whole query.

CTRL-X
  Toggle function description display (visible when function candidates are shown).

CTRL-F / Right Arrow
  Move cursor a character to the right.

CTRL-B / Left Arrow
  Move cursor a character to the left.

CTRL-A
  Move cursor to the first character of the Filter.

CTRL-E
  Move cursor to the end of the Filter.

CTRL-J
  Scroll json buffer 1 line downwards.

CTRL-K
  Scroll json buffer 1 line upwards.

CTRL-G
  Scroll json buffer to bottom.

CTRL-T
  Scroll json buffer to top.

CTRL-N
  Scroll json buffer Page Down.

CTRL-P
  Scroll json buffer Page Up.

CTRL-L
  Toggle view mode: full JSON or keys-only (objects only).

ESC
  Hide the candidate list.

Up Arrow
  Navigate to the previous query in history.

Down Arrow
  Navigate to the next query in history.

============ JMESPath examples =============

.users[*].name             wildcard: extract name from every user
. | keys(@)                pipe + function: list root keys
.users | sort_by(@, &name) sort array of objects by field
.users | length(@)         pipe + function: count elements

`
}
