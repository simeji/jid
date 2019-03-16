package jid

import (
	"io"
	"strings"

	termbox "github.com/nsf/termbox-go"
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
	manager        *JsonManager
	query          QueryInterface
	queryCursorIdx int
	term           *Terminal
	complete       []string
	keymode        bool
	candidates     []string
	candidatemode  bool
	candidateidx   int
	contentOffset  int
	queryConfirm   bool
	prettyResult   bool
}

type EngineAttribute struct {
	DefaultQuery string
	Monochrome   bool
	PrettyResult bool
}

func NewEngine(s io.Reader, ea *EngineAttribute) (EngineInterface, error) {
	j, err := NewJsonManager(s)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		manager:       j,
		term:          NewTerminal(FilterPrompt, DefaultY, ea.Monochrome),
		query:         NewQuery([]rune(ea.DefaultQuery)),
		complete:      []string{"", ""},
		keymode:       false,
		candidates:    []string{},
		candidatemode: false,
		candidateidx:  0,
		contentOffset: 0,
		queryConfirm:  false,
		prettyResult:  ea.PrettyResult,
	}
	e.queryCursorIdx = e.query.Length()
	return e, nil
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

		if e.query.StringGet() == "" {
			e.query.StringSet(".")
			e.queryCursorIdx = e.query.Length()
		}

		bl := len(contents)
		contents = e.getContents()
		e.setCandidateData()
		e.queryConfirm = false
		if bl != len(contents) {
			e.contentOffset = 0
		}

		ta := &TerminalDrawAttributes{
			Query:           e.query.StringGet(),
			Contents:        contents,
			CandidateIndex:  e.candidateidx,
			ContentsOffsetY: e.contentOffset,
			Complete:        e.complete[0],
			Candidates:      e.candidates,
			CursorOffset:    e.query.IndexOffset(e.queryCursorIdx),
		}
		err = e.term.Draw(ta)
		if err != nil {
			panic(err)
		}

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
			case termbox.KeyCtrlG:
				e.scrollToBottom(len(contents))
			case termbox.KeyCtrlT:
				e.scrollToTop()
			case termbox.KeyCtrlN:
				_, h := termbox.Size()
				e.scrollPageDown(len(contents), h)
			case termbox.KeyCtrlP:
				_, h := termbox.Size()
				e.scrollPageUp(h)
			case termbox.KeyCtrlL:
				e.toggleKeymode()
			case termbox.KeyCtrlU:
				e.deleteLineQuery()
			case termbox.KeyCtrlW:
				e.deleteWordBackward()
			case termbox.KeyEsc:
				e.escapeCandidateMode()
			case termbox.KeyEnter:
				if !e.candidatemode {
					var cc string
					var err error
					if e.prettyResult {
						cc, _, _, err = e.manager.GetPretty(e.query, true)
					} else {
						cc, _, _, err = e.manager.Get(e.query, true)
					}

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
	_ = e.query.StringAdd(e.candidates[e.candidateidx])
	e.queryCursorIdx = e.query.Length()
	e.queryConfirm = true
}

func (e *Engine) deleteChar() {
	if i := e.queryCursorIdx - 1; i > 0 {
		_ = e.query.Delete(i)
		e.queryCursorIdx--
	}

}

func (e *Engine) deleteLineQuery() {
	_ = e.query.StringSet("")
	e.queryCursorIdx = 0
}

func (e *Engine) scrollToBelow() {
	e.contentOffset++
}

func (e *Engine) scrollToAbove() {
	if o := e.contentOffset - 1; o >= 0 {
		e.contentOffset = o
	}
}

func (e *Engine) scrollToBottom(rownum int) {
	e.contentOffset = rownum - 1
}

func (e *Engine) scrollToTop() {
	e.contentOffset = 0
}

func (e *Engine) scrollPageDown(rownum int, height int) {
	co := rownum - 1
	if o := rownum - e.contentOffset; o > height {
		co = e.contentOffset + (height - DefaultY)
	}
	e.contentOffset = co
}

func (e *Engine) scrollPageUp(height int) {
	co := 0
	if o := e.contentOffset - (height - DefaultY); o > 0 {
		co = o
	}
	e.contentOffset = co
}

func (e *Engine) toggleKeymode() {
	e.keymode = !e.keymode
}
func (e *Engine) deleteWordBackward() {
	if k, _ := e.query.StringPopKeyword(); k != "" && !strings.Contains(k, "[") {
		_ = e.query.StringAdd(".")
	}
	e.queryCursorIdx = e.query.Length()
}
func (e *Engine) tabAction() {
	if !e.candidatemode {
		e.candidatemode = true
		if e.complete[0] != e.complete[1] && e.complete[0] != "" {
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
	e.queryCursorIdx = e.query.Length()
}
func (e *Engine) escapeCandidateMode() {
	e.candidatemode = false
}
func (e *Engine) inputChar(ch rune) {
	_ = e.query.Insert([]rune{ch}, e.queryCursorIdx)
	e.queryCursorIdx++
}

func (e *Engine) moveCursorBackward() {
	if i := e.queryCursorIdx - 1; i >= 0 {
		e.queryCursorIdx--
	}
}

func (e *Engine) moveCursorForward() {
	if e.query.Length() > e.queryCursorIdx {
		e.queryCursorIdx++
	}
}

func (e *Engine) moveCursorWordBackwark() {
}
func (e *Engine) moveCursorWordForward() {
}
func (e *Engine) moveCursorToTop() {
	e.queryCursorIdx = 0
}
func (e *Engine) moveCursorToEnd() {
	e.queryCursorIdx = e.query.Length()
}
