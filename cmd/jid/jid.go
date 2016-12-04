package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/simeji/jid"
)

func main() {
	content := os.Stdin

	var query bool

	flag.BoolVar(&query, "q", false, "output query")
	flag.Parse()

	e, err := jid.NewEngine(content)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(run(e, query))
}

func run(e jid.EngineInterface, query bool) int {

	result := e.Run()
	if result.GetError() != nil {
		return 2
	}
	if query {
		fmt.Printf("%s", result.GetQueryString())
	} else {
		fmt.Printf("%s", result.GetContent())
	}
	return 0
}
