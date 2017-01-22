package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/simeji/jid"
)

const VERSION = "0.7.2"

func main() {
	content := os.Stdin

	var qm bool
	var help bool
	var version bool
	var mono bool
	qs := "."

	flag.BoolVar(&qm, "q", false, "Output query mode")
	flag.BoolVar(&help, "h", false, "print a help")
	flag.BoolVar(&help, "help", false, "print a help")
	flag.BoolVar(&version, "version", false, "print the version and exit")
	flag.BoolVar(&mono, "M", false, "monochrome output mode")
	flag.Parse()

	if help {
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
  Show available items and choice them

CTRL-W
  Delete from the cursor to the start of the word

CTRL-U
  Delete whole query

CTRL-F / Right Arrow
  Move cursor a character to the right

CTRL-B / Left Arrow
  Move cursor a character to the left

CTRL-A
  To the first character of the 'Filter'

CTRL-E
  To the end of the 'Filter'

CTRL-J
  Scroll json buffer 1 line downwards

CTRL-K
  Scroll json buffer 1 line upwards

CTRL-G
  Scroll json buffer to bottom

CTRL-T
  Scroll json buffer to top

CTRL-L
  Change view mode whole json or keys (only object)

ESC
  Hide a candidate box

`
}
