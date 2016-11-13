package jid

import (
	"github.com/nsf/termbox-go"
	"io"
	"strings"
)

const (
	DefaultY     int    = 1
	FilterPrompt string = "[Filter]> "
)

type EngineInterface interface {
	Run() EngineResultInterface
	GetQuery() QueryInterface
}

type EngineResultInterface interface {
	GetQueryString() string
	GetContent() string
	GetError() error
}

type Engine struct {
	manager       *JsonManager
	query         QueryInterface
	term          *Terminal
	complete      []string
	keymode       bool
	candidates    []string
	candidatemode bool
	candidateidx  int
	contentOffset int
	queryConfirm  bool
	cursorOffsetX int
}

func NewEngine(s io.Reader) EngineInterface {
	j, err := NewJsonManager(s)
	if err != nil {
		return &Engine{}
	}
	e := &Engine{
		manager:       j,
		term:          NewTerminal(FilterPrompt, DefaultY),
		query:         NewQuery([]rune("")),
		complete:      []string{"", ""},
		keymode:       false,
		candidates:    []string{},
		candidatemode: false,
		candidateidx:  0,
		contentOffset: 0,
		queryConfirm:  false,
		cursorOffsetX: 0,
	}
	return e
}

type EngineResult struct {
	content string
	qs      string
	err     error
}

func (er *EngineResult) GetQueryString() string {
	return er.qs
}

func (er *EngineResult) GetContent() string {
	return er.content
}
func (er *EngineResult) GetError() error {
	return er.err
}

func (e *Engine) GetQuery() QueryInterface {
	return e.query
}

func (e *Engine) Run() EngineResultInterface {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	var contents []string

	for {
		contents = e.getContents()
		e.setCandidateData()
		e.queryConfirm = false

		ta := &TerminalDrawAttributes{
			Query:           e.query.StringGet(),
			CursorOffsetX:   e.cursorOffsetX,
			Contents:        contents,
			CandidateIndex:  e.candidateidx,
			ContentsOffsetY: e.contentOffset,
			Complete:        e.complete[0],
			Candidates:      e.candidates,
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
					cc, _, _, err := e.manager.Get(e.query, true)
					return &EngineResult{
						content: cc,
						qs:      e.query.StringGet(),
						err:     err,
					}
				}
				e.confirmCandidate()
			case termbox.KeyCtrlC:
				return &EngineResult{}
			default:
			}
		case termbox.EventError:
			panic(ev.Err)
			break
		default:
		}
	}
}

func (e *Engine) getContents() []string {
	var c string
	var contents []string
	c, e.complete, e.candidates, _ = e.manager.GetPretty(e.query, e.queryConfirm)
	if e.keymode {
		contents = e.candidates
	} else {
		contents = strings.Split(c, "\n")
	}
	return contents
}

func (e *Engine) setCandidateData() {
	if l := len(e.candidates); e.complete[0] == "" && l > 1 {
		if e.candidateidx >= l {
			e.candidateidx = 0
		}
	} else {
		e.candidatemode = false
	}
	if !e.candidatemode {
		e.candidateidx = 0
		e.candidates = []string{}
	}
}

func (e *Engine) confirmCandidate() {
	_, _ = e.query.PopKeyword()
	_ = e.query.StringAdd(".")
	q := e.query.StringAdd(e.candidates[e.candidateidx])
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
		if e.query.StringGet() == "" {
			_ = e.query.StringAdd(".")
		} else if e.complete[0] != e.complete[1] && e.complete[0] != "" {
			if k, _ := e.query.StringPopKeyword(); !strings.Contains(k, "[") {
				_ = e.query.StringAdd(".")
			}
			_ = e.query.StringAdd(e.complete[1])
		} else {
			_ = e.query.StringAdd(e.complete[0])
		}
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
