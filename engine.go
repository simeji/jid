package jid

import (
	"io"
	"strconv"
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
	// Shift+Tab detection: \x1b[Z arrives as KeyEsc → '[' → 'Z' events
	escapePending      bool
	bracketPending     bool
	showFuncHelp       bool
	// saved state to restore when Esc turns out to be part of Shift+Tab
	savedCandidates    []string
	savedCandidateIdx  int
	savedCandidateMode bool
	// placeholder: ghost text at cursor position after function confirmation
	placeholderStart int // rune index in query; -1 if inactive
	placeholderLen   int
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
		prettyResult:     ea.PrettyResult,
		showFuncHelp:     true,
		placeholderStart: -1,
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

		funcHelp := ""
		if e.showFuncHelp {
			if l := len(e.candidates); l > 0 && strings.HasSuffix(e.candidates[0], "(") {
				selected := e.candidates[e.candidateidx%l]
				funcHelp = FunctionDescription(selected) + "  [Ctrl+X: hide]"
			}
		}
		ta := &TerminalDrawAttributes{
			Query:            e.query.StringGet(),
			Contents:         contents,
			CandidateIndex:   e.candidateidx,
			ContentsOffsetY:  e.contentOffset,
			Complete:         e.complete[0],
			Candidates:       e.candidates,
			CursorOffset:     e.query.IndexOffset(e.queryCursorIdx),
			FuncHelp:         funcHelp,
			PlaceholderStart: e.placeholderStart,
			PlaceholderLen:   e.placeholderLen,
		}
		err = e.term.Draw(ta)
		if err != nil {
			panic(err)
		}

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case 0:
				// Detect Shift+Tab (\x1b[Z) arriving as: KeyEsc → '[' → 'Z'.
				// We insert '[' immediately so normal '[' typing still works;
				// if 'Z' follows we undo the '[' and perform shiftTabAction instead.
				if e.bracketPending && ev.Ch == 'Z' {
					e.bracketPending = false
					e.deleteChar() // remove the '[' we inserted below
					// Restore candidate state saved when Esc arrived
					e.candidates = e.savedCandidates
					e.candidateidx = e.savedCandidateIdx
					e.candidatemode = e.savedCandidateMode
					e.shiftTabAction()
				} else if e.escapePending && ev.Ch == '[' {
					e.escapePending = false
					e.bracketPending = true
					e.inputChar('[') // insert normally; undone above if 'Z' follows
				} else {
					e.escapePending = false
					e.bracketPending = false
					e.inputChar(ev.Ch)
				}
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
				// Save candidate state in case this Esc is the start of Shift+Tab (\x1b[Z)
				e.savedCandidates = e.candidates
				e.savedCandidateIdx = e.candidateidx
				e.savedCandidateMode = e.candidatemode
				e.escapePending = true
				e.bracketPending = false
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
			case termbox.KeyCtrlX:
				if e.candidatemode && len(e.candidates) > 0 &&
					strings.HasSuffix(e.candidates[0], "(") {
					e.showFuncHelp = !e.showFuncHelp
				}
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
	l := len(e.candidates)
	isFuncCandidates := l > 0 && strings.HasSuffix(e.candidates[0], "(")
	// Field candidates returned while in JMESPath pipe mode or after a
	// wildcard projection ([*] / .*) also auto-show without a Tab press.
	qs0 := e.query.StringGet()
	// Match queries that contain a wildcard projection anywhere (e.g. "[*]" at end
	// or followed by ".field"), so candidates auto-show even after the trailing dot.
	isWildcardProjection0 := strings.Contains(qs0, "[*]") || strings.Contains(qs0, ".*")
	isJMESPathFieldCandidates := !isFuncCandidates && l > 0 && (strings.Contains(qs0, "|") || isWildcardProjection0)
	if isFuncCandidates || isJMESPathFieldCandidates {
		// Auto-enter candidate mode so the list appears without a Tab press.
		e.candidatemode = true
		if e.candidateidx >= l {
			e.candidateidx = 0
		}
	} else if e.complete[0] == "" && l > 1 {
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
	selected := e.candidates[e.candidateidx]

	// JMESPath function candidates end with "(". Insert them as a pipe expression
	// so the user gets: <base> | funcname(@)
	if strings.HasSuffix(selected, "(") {
		funcName := strings.TrimSuffix(selected, "(")
		// Strip any trailing "|<partial>" or " | <partial>" typed so far.
		qs := e.query.StringGet()
		if idx := strings.LastIndex(qs, "|"); idx >= 0 {
			_ = e.query.StringSet(strings.TrimRight(qs[:idx], " "))
		}
		args, cursorBack, phLen := FunctionTemplate(funcName)
		_ = e.query.StringAdd(" | ")
		_ = e.query.StringAdd(funcName)
		_ = e.query.StringAdd("(" + args + ")")
		e.queryCursorIdx = e.query.Length() - cursorBack
		if phLen > 0 {
			e.placeholderStart = e.queryCursorIdx
			e.placeholderLen = phLen
		} else {
			e.placeholderStart = -1
			e.placeholderLen = 0
		}
	} else {
		qs := e.query.StringGet()
		if pipeIdx := strings.LastIndex(qs, "|"); pipeIdx >= 0 {
			suffix := strings.TrimLeft(qs[pipeIdx+1:], " ")
			if strings.Contains(suffix, "(") {
				// The pipe suffix is a complete expression (e.g. "to_array(@)[0]").
				// Append .field to the full query rather than stripping back to the base.
				// If query already ends with "." don't add another.
				if strings.HasSuffix(qs, ".") {
					_ = e.query.StringAdd(selected)
				} else {
					_ = e.query.StringAdd("." + selected)
				}
			} else {
				// Simple field name after pipe (e.g. ".[1] | body"); replace with base.field.
				base := strings.TrimRight(qs[:pipeIdx], " ")
				_ = e.query.StringSet(base + "." + selected)
			}
		} else if strings.Contains(qs, "[*]") || strings.Contains(qs, ".*") {
			// Wildcard context: just append .fieldname (no PopKeyword).
			// Covers both "[*]" at end and "[*].something[0]" in mid-path.
			// If query already ends with "." (trailing dot), don't add another.
			if strings.HasSuffix(qs, ".") {
				_ = e.query.StringAdd(selected)
			} else {
				_ = e.query.StringAdd("." + selected)
			}
		} else {
			_, _ = e.query.PopKeyword()
			_ = e.query.StringAdd(".")
			_ = e.query.StringAdd(selected)
		}
		e.queryCursorIdx = e.query.Length()
	}
	e.queryConfirm = true
}

func (e *Engine) deleteChar() {
	e.clearPlaceholder()
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
// removeLastJMESPathSegment removes the last navigation segment from a
// JMESPath expression suffix, working backwards:
//   "to_array(@)[0].id"  → "to_array(@)[0]"  (remove .field)
//   "to_array(@)[0]"     → "to_array(@)"      (remove [index])
//   "to_array(@)"        → ""                 (remove function call)
func removeLastJMESPathSegment(expr string) string {
	if expr == "" {
		return ""
	}
	parenDepth := 0
	bracketDepth := 0
	for i := len(expr) - 1; i >= 0; i-- {
		switch expr[i] {
		case ')':
			parenDepth++
		case '(':
			parenDepth--
		case ']':
			bracketDepth++
		case '[':
			bracketDepth--
			if bracketDepth == 0 && parenDepth == 0 {
				return expr[:i]
			}
		case '.':
			if parenDepth == 0 && bracketDepth == 0 {
				return expr[:i]
			}
		}
	}
	return "" // entire expression is one segment
}

func (e *Engine) deleteWordBackward() {
	qs := e.query.StringGet()
	// JMESPath pipe mode: remove one segment at a time from the suffix.
	if idx := strings.LastIndex(qs, "|"); idx >= 0 {
		suffix := strings.TrimLeft(qs[idx+1:], " ")
		base := strings.TrimRight(qs[:idx], " ")
		if suffix != "" {
			newSuffix := removeLastJMESPathSegment(suffix)
			if newSuffix != "" {
				_ = e.query.StringSet(base + " | " + newSuffix)
			} else {
				// Suffix gone; keep "base | " so next Ctrl+W removes the pipe.
				_ = e.query.StringSet(base + " | ")
			}
		} else {
			// Nothing after pipe: remove the pipe itself.
			_ = e.query.StringSet(base)
		}
		e.queryCursorIdx = e.query.Length()
		return
	}
	if k, _ := e.query.StringPopKeyword(); k != "" && !strings.Contains(k, "[") {
		_ = e.query.StringAdd(".")
	}
	e.queryCursorIdx = e.query.Length()
}
// changeArrayIndex increments (delta=1) or decrements (delta=-1) the last
// array index in the query string, e.g. ".users[0]" → ".users[1]".
// Returns true if the query ended with [N] and was updated.
func (e *Engine) changeArrayIndex(delta int) bool {
	qs := e.query.StringGet()
	end := strings.LastIndex(qs, "]")
	if end < 0 || end != len(qs)-1 {
		return false
	}
	start := strings.LastIndex(qs[:end], "[")
	if start < 0 {
		return false
	}
	idx, err := strconv.Atoi(qs[start+1 : end])
	if err != nil {
		return false
	}
	idx += delta

	// Determine array length from the parent expression for wrap-around.
	maxIdx := -1
	parentQS := qs[:start]
	if parentQS != "" {
		parentQuery := NewQuery([]rune(parentQS))
		if parentJson, _, _, err := e.manager.GetFilteredData(parentQuery, true); err == nil {
			if arr, err := parentJson.Array(); err == nil {
				maxIdx = len(arr) - 1
			}
		}
	}
	if maxIdx >= 0 {
		if idx > maxIdx {
			idx = 0
		} else if idx < 0 {
			idx = maxIdx
		}
	} else {
		if idx < 0 {
			idx = 0
		}
	}
	_ = e.query.StringSet(qs[:start] + "[" + strconv.Itoa(idx) + "]")
	e.queryCursorIdx = e.query.Length()
	return true
}

func (e *Engine) tabAction() {
	if e.candidatemode {
		qs := e.query.StringGet()
		isFuncCandidates := len(e.candidates) > 0 && strings.HasSuffix(e.candidates[0], "(")
		isWildcard := strings.Contains(qs, "[*]") || strings.Contains(qs, ".*")
		isJMESPathField := len(e.candidates) > 0 && !isFuncCandidates && (strings.Contains(qs, "|") || isWildcard)
		if (isFuncCandidates || isJMESPathField) && len(e.candidates) == 1 {
			e.confirmCandidate()
			return
		}
		e.candidateidx = (e.candidateidx + 1) % len(e.candidates)
		e.queryCursorIdx = e.query.Length()
		return
	}
	// Array index increment: query ends with [N], works in all contexts.
	if e.changeArrayIndex(1) {
		return
	}
	// Query ends with "[*": close the wildcard projection.
	if strings.HasSuffix(e.query.StringGet(), "[*") {
		_ = e.query.StringAdd("]")
		e.queryCursorIdx = e.query.Length()
		return
	}
	// Query ends with "[": user explicitly started a subscript → complete to [0].
	// Place cursor right after "[" (before the digits) so the user can type a number.
	if strings.HasSuffix(e.query.StringGet(), "[") {
		_ = e.query.StringAdd("0]")
		e.queryCursorIdx = e.query.Length() - 2
		return
	}
	// complete[1] starting with "[" means the result is an array → add [0] or [N].
	// Multi-element arrays return complete[1]="["; single-element returns "[0]" etc.
	// Works in pipe context too. Function typing mode returns ["",""] so never triggers.
	c1 := e.complete[1]
	if len(c1) >= 1 && c1[0] == '[' {
		add := c1
		if add == "[" {
			add = "[0]"
		}
		_ = e.query.StringAdd(add)
		e.queryCursorIdx = e.query.Length() - 2
		return
	}
	// Do not run field completion (StringPopKeyword) when a pipe is present —
	// that would destroy the pipe expression.
	if strings.Contains(e.query.StringGet(), "|") {
		return
	}
	// Original field completion logic
	e.candidatemode = true
	if e.complete[0] != e.complete[1] && e.complete[0] != "" {
		if k, _ := e.query.StringPopKeyword(); !strings.Contains(k, "[") {
			_ = e.query.StringAdd(".")
		}
		_ = e.query.StringAdd(e.complete[1])
	} else {
		_ = e.query.StringAdd(e.complete[0])
	}
	e.queryCursorIdx = e.query.Length()
}

func (e *Engine) shiftTabAction() {
	if e.candidatemode && len(e.candidates) > 0 {
		if e.candidateidx <= 0 {
			e.candidateidx = len(e.candidates) - 1
		} else {
			e.candidateidx--
		}
		e.queryCursorIdx = e.query.Length()
		return
	}
	e.changeArrayIndex(-1)
}
func (e *Engine) escapeCandidateMode() {
	e.candidatemode = false
}
func (e *Engine) clearPlaceholder() {
	e.placeholderStart = -1
	e.placeholderLen = 0
}

func (e *Engine) inputChar(ch rune) {
	if e.placeholderLen > 0 && e.queryCursorIdx == e.placeholderStart {
		// Replace placeholder text with the typed character
		for i := 0; i < e.placeholderLen; i++ {
			_ = e.query.Delete(e.placeholderStart)
		}
		_ = e.query.Insert([]rune{ch}, e.placeholderStart)
		e.queryCursorIdx = e.placeholderStart + 1
		e.clearPlaceholder()
		return
	}
	e.clearPlaceholder()
	before := e.query.Length()
	_ = e.query.Insert([]rune{ch}, e.queryCursorIdx)
	if e.query.Length() > before {
		e.queryCursorIdx++
	}
}

func (e *Engine) moveCursorBackward() {
	e.clearPlaceholder()
	if i := e.queryCursorIdx - 1; i >= 0 {
		e.queryCursorIdx--
	}
}

func (e *Engine) moveCursorForward() {
	e.clearPlaceholder()
	if e.query.Length() > e.queryCursorIdx {
		e.queryCursorIdx++
	}
}

func (e *Engine) moveCursorWordBackwark() {
}
func (e *Engine) moveCursorWordForward() {
}
func (e *Engine) moveCursorToTop() {
	e.clearPlaceholder()
	e.queryCursorIdx = 0
}
func (e *Engine) moveCursorToEnd() {
	e.clearPlaceholder()
	e.queryCursorIdx = e.query.Length()
}
