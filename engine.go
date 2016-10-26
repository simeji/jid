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
	cursorOffsetX int
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
		cursorOffsetX: 0,
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

		ta := &TerminalDrawAttributes{
			Query:           e.query.StringGet(),
			CursorOffsetX:   e.cursorOffsetX,
			Contents:        contents,
			CandidateIndex:  e.candidateidx,
			ContentsOffsetY: e.contentOffset,
			Complete:        e.complete[0],
			Candidates:      candidates,
		}

		e.term.draw(ta)

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 0:
				e.inputChar(ev.Ch)
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				e.deleteChar()
			case termbox.KeyTab:
				e.tabAction()
			case termbox.KeyArrowLeft, termbox.KeyCtrlB:
				e.moveCursorBackward()
			case termbox.KeyArrowRight, termbox.KeyCtrlF:
				e.moveCursorForward()
			case termbox.KeyHome, termbox.KeyCtrlA:
				e.moveCursorToTop()
			case termbox.KeyEnd, termbox.KeyCtrlE:
				e.moveCursorToEnd()
			case termbox.KeyCtrlK:
				e.scrollToAbove()
			case termbox.KeyCtrlJ:
				e.scrollToBelow()
			case termbox.KeyCtrlL:
				e.toggleKeymode()
			case termbox.KeyCtrlW:
				e.deleteWordBackward()
			case termbox.KeyEsc:
				e.escapeCandidateMode()
			case termbox.KeyEnter:
				if !e.candidatemode {
					return true
				}
				e.confirmCandidate(candidates)
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

func (e *Engine) confirmCandidate(candidates []string) {
	_, _ = e.query.PopKeyword()
	_ = e.query.StringAdd(".")
	q := e.query.StringAdd(candidates[e.candidateidx])
	e.cursorOffsetX = len(q)
	e.queryConfirm = true
}

func (e *Engine) deleteChar() {
	if e.cursorOffsetX > 0 {
		_ = e.query.Delete(e.cursorOffsetX - 1)
		e.cursorOffsetX -= 1
	}
}
func (e *Engine) scrollToBelow() {
	e.contentOffset++
}
func (e *Engine) scrollToAbove() {
	if o := e.contentOffset - 1; o >= 0 {
		e.contentOffset = o
	}
}
func (e *Engine) toggleKeymode() {
	e.keymode = !e.keymode
}
func (e *Engine) deleteWordBackward() {
	if k, _ := e.query.StringPopKeyword(); k != "" && !strings.Contains(k, "[") {
		_ = e.query.StringAdd(".")
	}
	e.cursorOffsetX = len(e.query.Get())
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
	e.cursorOffsetX = len(e.query.Get())
}
func (e *Engine) escapeCandidateMode() {
	e.candidatemode = false
}
func (e *Engine) inputChar(ch rune) {
	b := len(e.query.Get())
	q := e.query.StringInsert(string(ch), e.cursorOffsetX)
	if b < len(q) {
		e.cursorOffsetX += 1
	}
}

func (e *Engine) moveCursorBackward() {
	if e.cursorOffsetX > 0 {
		e.cursorOffsetX -= 1
	}
}
func (e *Engine) moveCursorForward() {
	if len(e.query.Get()) > e.cursorOffsetX {
		e.cursorOffsetX += 1
	}
}
func (e *Engine) moveCursorWordBackwark() {
}
func (e *Engine) moveCursorWordForward() {
}
func (e *Engine) moveCursorToTop() {
	e.cursorOffsetX = 0
}
func (e *Engine) moveCursorToEnd() {
	e.cursorOffsetX = len(e.query.Get())
}
