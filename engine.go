package jig

import (
	//"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	PROMPT              = "[jig]>> "
	DefaultY     int    = 1
	FilterPrompt string = "[Filter]> "
)

var (
	f *[]rune
)

type Engine struct {
	json    *simplejson.Json
	orgJson *simplejson.Json
	query   bool
	pretty  bool
}

func NewEngine(s *os.File, q bool, p bool) *Engine {
	j := parse(s)
	e := &Engine{
		json:    j,
		orgJson: j,
		query:   q,
		pretty:  p,
	}
	return e
}

func (e Engine) Run() int {

	if !e.render(e.json) {
		return 2
	}
	if e.query {
		fmt.Printf("%s", string(*f))
	} else if e.pretty {
		s, err := e.json.EncodePretty()
		if err != nil {
			return 1
		}
		fmt.Printf("%s", string(s))
	} else {
		s, err := e.json.Encode()
		if err != nil {
			return 1
		}
		fmt.Printf("%s", s)
	}
	return 0
}

func parse(content *os.File) *simplejson.Json {
	buf, err := ioutil.ReadAll(content)

	if err != nil {
		log.Fatal(err)
	}

	js, err := simplejson.NewJson(buf)

	if err != nil {
		log.Fatal(err)
	}

	return js
}

// fix:me
func (e *Engine) render(json *simplejson.Json) bool {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	f = &[]rune{}

	contents := e.prettyContents()

	for {
		e.json = e.filterByQuery(string(*f))
		contents = e.prettyContents()
		draw(contents)
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeySpace:
				*f = append(*f, rune(' '))
				draw(contents)
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return false
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if i := len(*f) - 1; i >= 0 {
					slice := *f
					*f = slice[0:i]
				}
				//draw(contents)
			case termbox.KeyEnter:
				return true
			case 0:
				*f = append(*f, rune(ev.Ch))
			default:
			}
		case termbox.EventError:
			panic(ev.Err)
			break
		default:
		}
	}
}

func (e *Engine) prettyContents() []string {
	s, _ := e.json.EncodePretty()
	return strings.Split(string(s), "\n")
}

func (e *Engine) filterByQuery(q string) *simplejson.Json {
	json := e.orgJson
	if len(q) > 0 {
		keywords := strings.Split(q, ".")

		if keywords[0] != "" {
			return json
		}
		keywords = keywords[1:]

		j := json

		re := regexp.MustCompile("\\[[0-9]+\\]")
		delre := regexp.MustCompile("\\[([0-9]+)?")

		for _, keyword := range keywords {
			if len(keyword) == 0 {
				break
			}

			if keyword[:1] == "[" {
				matchIndexes := re.FindAllStringIndex(keyword, -1)
				for _, m := range matchIndexes {
					idx, _ := strconv.Atoi(keyword[m[0]+1 : m[1]-1])
					if tj := j.GetIndex(idx); !isEmptyJson(tj) {
						j = tj
					}
				}
				kw := re.ReplaceAllString(keyword, "")
				if tj := j.Get(kw); !isEmptyJson(tj) {
					j = tj
				}
			} else if keyword[len(keyword)-1:] == "]" {
				matchIndexes := re.FindAllStringIndex(keyword, -1)
				kw := re.ReplaceAllString(keyword, "")
				if tj := j.Get(kw); !isEmptyJson(tj) {
					j = tj
				}
				for _, m := range matchIndexes {
					idx, _ := strconv.Atoi(keyword[m[0]+1 : m[1]-1])
					if tj := j.GetIndex(idx); !isEmptyJson(tj) {
						j = tj
					}
				}
			} else {
				kw := delre.ReplaceAllString(keyword, "")
				if tj := j.Get(kw); !isEmptyJson(tj) {
					j = tj
				}
			}
		}

		switch j.Interface().(type) {
		case nil:
			json = e.orgJson
		default:
			json = j
		}
	}
	return json
}

func isEmptyJson(j *simplejson.Json) bool {
	switch j.Interface().(type) {
	case nil:
		return true
	default:
		return false
	}
}

func draw(rows []string) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	fs := FilterPrompt + string(*f)
	drawln(0, 0, fs)
	termbox.SetCursor(len(fs), 0)

	for idx, row := range rows {
		drawln(0, idx+DefaultY, row)
	}

	termbox.Flush()
}

func drawln(x int, y int, str string) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault

	var c termbox.Attribute
	for i, s := range str {
		c = color
		//for _, match := range matches {
		//if i >= match[0] && i < match[1] {
		//c = termbox.ColorGreen
		//}
		//}
		termbox.SetCell(x+i, y, s, c, backgroundColor)
	}
}
