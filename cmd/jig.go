package main

import (
	"flag"
	"fmt"
	"github.com/simeji/jig"
	"os"
)

func main() {
	content := os.Stdin

	var query bool

	flag.BoolVar(&query, "q", false, "output query")
	flag.Parse()

	e := jig.NewEngine(content)
	os.Exit(run(e, query))
}

func run(e jig.EngineInterface, query bool) int {

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
