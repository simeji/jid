package main

import (
	//"fmt"
	"flag"
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

	j := jig.NewEngine(content, query, pretty)
	os.Exit(j.Run())
}
