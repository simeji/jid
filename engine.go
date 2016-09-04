package jig

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"io"
	"strings"
)

const (
	PROMPT              = "[jig]>> "
	DefaultY     int    = 1
	FilterPrompt string = "[Filter]> "
)

type Engine struct {
	manager  *JsonManager
	jq       bool
	pretty   bool
	query    *Query
	complete []string
}

func NewEngine(s io.Reader, q bool, p bool) *Engine {
	j, err := NewJsonManager(s)
	if err != nil {
		return &Engine{}
	}
	e := &Engine{
		manager:  j,
		jq:       q,
		pretty:   p,
		query:    NewQuery([]rune("")),
		complete: []string{"", ""},
	}
	return e
}

func (e Engine) Run() int {

	if !e.render() {
		return 2
	}
	if e.jq {
		fmt.Printf("%s", e.query.StringGet())
	} else if e.pretty {
		s, _, err := e.manager.GetPretty(e.query)
		if err != nil {
			return 1
		}
		fmt.Printf("%s", s)
	} else {
		s, _, err := e.manager.Get(e.query)
		if err != nil {
			return 1
		}
		fmt.Printf("%s", s)
	}
	return 0
}

func (e *Engine) render() bool {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	keymode := false

	var contents []string
	var c string

	for {
		//var flgFilter bool
		if keymode {
			contents = e.manager.GetCandidateKeys(e.query)
		} else {
			c, e.complete, _ = e.manager.GetPretty(e.query)
			contents = strings.Split(c, "\n")
		}
		e.draw(e.query.StringGet(), e.complete[0], contents)
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 0:
				e.inputAction(ev.Ch)
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				e.backspaceAction()
			case termbox.KeyTab:
				e.tabAction()
			case termbox.KeyCtrlK:
				keymode = !keymode
			case termbox.KeySpace:
				e.spaceAction()
			case termbox.KeyCtrlW:
				e.ctrlwAction()
			case termbox.KeyEnter:
				return true
			case termbox.KeyEsc, termbox.KeyCtrlC:
				return false
			default:
			}
		case termbox.EventError:
			panic(ev.Err)
			break
		default:
		}
	}
}

func (e *Engine) spaceAction() {
	_ = e.query.StringAdd(" ")
}
func (e *Engine) backspaceAction() {
	_ = e.query.Delete(1)
}
func (e *Engine) ctrlwAction() {
	_, _ = e.query.PopKeyword()
	_ = e.query.StringAdd(".")
}
func (e *Engine) tabAction() {
	if (e.complete[0] != e.complete[1]) && e.complete[0] != "" {
		_, _ = e.query.PopKeyword()
		_ = e.query.StringAdd(".")
	}
	_ = e.query.StringAdd(e.complete[1])
}
func (e *Engine) inputAction(ch rune) {
	_ = e.query.StringAdd(string(ch))
}

func (e *Engine) draw(query string, complete string, rows []string) {

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	fs := FilterPrompt + query
	cs := complete
	drawln(0, 0, fs+cs, []([]int){[]int{len(fs), len(fs) + len(cs)}})
	termbox.SetCursor(len(fs), 0)

	for idx, row := range rows {
		drawln(0, idx+DefaultY, row, nil)
	}

	termbox.Flush()
}

func drawln(x int, y int, str string, matches [][]int) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault

	var c termbox.Attribute
	for i, s := range str {
		c = color
		for _, match := range matches {
			if i >= match[0] && i < match[1] {
				c = termbox.ColorGreen
			}
		}
		termbox.SetCell(x+i, y, s, c, backgroundColor)
	}
}
