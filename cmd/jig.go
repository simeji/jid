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
	var pretty bool

	flag.BoolVar(&query, "q", false, "output query")
	flag.BoolVar(&pretty, "p", false, "pretty print")
	flag.Parse()

	e := jig.NewEngine(content)
	os.Exit(run(e, query, pretty))
}

func run(e jig.EngineInterface, query bool, pretty bool) int {

	result := e.Run()
	if result.GetError() != nil {
		return 2
	}
	if query {
		fmt.Printf("%s", result.GetQueryString())
	} else if pretty {
		//s, _, _, err := e.manager.GetPretty(e.query, true)
		//if err != nil {
		//return 1
		//}
		//fmt.Printf("%s", s)
	} else {
		s := result.GetContent()
		fmt.Printf("%s", s)
	}
	return 0
}
