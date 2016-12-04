package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/simeji/jid"
)

func main() {
	content := os.Stdin

	var qm bool
	qs := "."

	flag.BoolVar(&qm, "q", false, "Output query mode")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		qs = args[0]
	}

	e, err := jid.NewEngine(content, qs)

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
