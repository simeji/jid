package jig

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"io"
	"strings"
)

const (
	DefaultY     int    = 1
	FilterPrompt string = "[Filter]> "
)

type Engine struct {
	manager       *JsonManager
	jq            bool
	pretty        bool
	query         *Query
	term          *Terminal
	complete      []string
	keymode       bool
	candidatemode bool
	candidateidx  int
	contentOffset int
	queryConfirm  bool
}

func NewEngine(s io.Reader, q bool, p bool) *Engine {
	j, err := NewJsonManager(s)
	if err != nil {
		return &Engine{}
	}
	e := &Engine{
		manager:       j,
		jq:            q,
		pretty:        p,
		term:          NewTerminal(FilterPrompt, DefaultY),
		query:         NewQuery([]rune("")),
		complete:      []string{"", ""},
		keymode:       false,
		candidatemode: false,
		candidateidx:  0,
		contentOffset: 0,
		queryConfirm:  false,
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
		s, _, _, err := e.manager.GetPretty(e.query, true)
		if err != nil {
			return 1
		}
		fmt.Printf("%s", s)
	} else {
		s, _, _, err := e.manager.Get(e.query, true)
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

	var contents []string
	var candidates []string
	var c string

	for {
		c, e.complete, candidates, _ = e.manager.GetPretty(e.query, e.queryConfirm)
		e.queryConfirm = false
		if e.keymode {
			contents = candidates
		} else {
			contents = strings.Split(c, "\n")
		}
		if l := len(candidates); e.complete[0] == "" && l > 1 {
			//e.candidatemode = true
			if e.candidateidx >= l {
				e.candidateidx = 0
			}
		} else {
			e.candidatemode = false
		}
		if !e.candidatemode {
			e.candidateidx = 0
			candidates = []string{}
		}

		e.term.draw(e.query.StringGet(), e.complete[0], contents, candidates, e.candidateidx, e.contentOffset)

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
				e.ctrlkAction()
			case termbox.KeyCtrlJ:
				e.ctrljAction()
			case termbox.KeyCtrlL:
				e.ctrllAction()
			case termbox.KeySpace:
				e.spaceAction()
			case termbox.KeyCtrlW:
				e.ctrlwAction()
			case termbox.KeyEsc:
				e.escAction()
			case termbox.KeyEnter:
				if !e.candidatemode {
					return true
				}
				_, _ = e.query.PopKeyword()
				_ = e.query.StringAdd(".")
				_ = e.query.StringAdd(candidates[e.candidateidx])
				e.queryConfirm = true

			case termbox.KeyCtrlC:
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
func (e *Engine) ctrljAction() {
	e.contentOffset++
}
func (e *Engine) ctrlkAction() {
	if o := e.contentOffset - 1; o >= 0 {
		e.contentOffset = o
	}
}
func (e *Engine) ctrllAction() {
	e.keymode = !e.keymode
}
func (e *Engine) ctrlwAction() {
	if k, _ := e.query.StringPopKeyword(); k != "" && !strings.Contains(k, "[") {
		_ = e.query.StringAdd(".")
	}
}
func (e *Engine) tabAction() {
	if !e.candidatemode {
		e.candidatemode = true
		if e.complete[0] != e.complete[1] && e.complete[0] != "" {
			if k, _ := e.query.StringPopKeyword(); !strings.Contains(k, "[") {
				_ = e.query.StringAdd(".")
			}
		}
		if e.query.StringGet() == "" {
			_ = e.query.StringAdd(".")
		}
		_ = e.query.StringAdd(e.complete[1])
	} else {
		e.candidateidx = e.candidateidx + 1
	}
}
func (e *Engine) escAction() {
	e.candidatemode = false
}
func (e *Engine) inputAction(ch rune) {
	_ = e.query.StringAdd(string(ch))
}
