package jid

import (
	"github.com/nsf/termbox-go"
	"github.com/nwidger/jsoncolor"

	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type Terminal struct {
	defaultY int
	prompt   string
}

type TerminalDrawAttributes struct {
	Query           string
	CursorOffsetX   int
	Contents        []string
	CandidateIndex  int
	ContentsOffsetY int
	Complete        string
	Candidates      []string
}

func NewTerminal(prompt string, defaultY int) *Terminal {
	return &Terminal{
		prompt:   prompt,
		defaultY: defaultY,
	}
}

func (t *Terminal) draw(attr *TerminalDrawAttributes) error {

	query := attr.Query
	complete := attr.Complete
	rows := attr.Contents
	candidates := attr.Candidates
	candidateidx := attr.CandidateIndex
	contentOffsetY := attr.ContentsOffsetY

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	fs := t.prompt + query
	cs := complete
	y := t.defaultY

	t.drawln(0, 0, fs+cs, []([]int){[]int{len(fs), len(fs) + len(cs)}})

	if len(candidates) > 0 {
		y = t.drawCandidates(0, t.defaultY, candidateidx, candidates)
	}

	cellsArr, err := t.rowsToCells(rows)
	if err != nil {
		return err
	}

	for idx, cells := range cellsArr {
		if i := idx - contentOffsetY; i >= 0 {
			t.drawCells(0, i+y, cells)
		}
	}
	termbox.SetCursor(len(t.prompt)+attr.CursorOffsetX, 0)

	termbox.Flush()
	return nil
}

type termboxSprintfFuncer struct {
	fg         termbox.Attribute
	bg         termbox.Attribute
	appendFunc func(string, termbox.Attribute, termbox.Attribute)
}

func (tsf *termboxSprintfFuncer) SprintfFunc() func(format string, a ...interface{}) string {
	return func(format string, a ...interface{}) string {
		str := fmt.Sprintf(format, a...)
		tsf.appendFunc(str, tsf.fg, tsf.bg)
		return str
	}
}

func (t *Terminal) rowsToCells(rows []string) ([][]termbox.Cell, error) {
	var cells [][]termbox.Cell

	appendString := func(str string, fg, bg termbox.Attribute) {
		if cells == nil {
			cells = [][]termbox.Cell{[]termbox.Cell{}}
		}
		for _, s := range str {
			if s == '\n' {
				cells = append(cells, []termbox.Cell{})
				continue
			}
			cells[len(cells)-1] = append(cells[len(cells)-1], termbox.Cell{
				Ch: s,
				Fg: fg,
				Bg: bg,
			})
		}
	}

	formatter := jsoncolor.NewFormatter()

	regular := &termboxSprintfFuncer{
		fg:         termbox.ColorDefault,
		bg:         termbox.ColorDefault,
		appendFunc: appendString,
	}

	bold := &termboxSprintfFuncer{
		fg:         termbox.AttrBold,
		bg:         termbox.ColorDefault,
		appendFunc: appendString,
	}

	blueBold := &termboxSprintfFuncer{
		fg:         termbox.ColorBlue | termbox.AttrBold,
		bg:         termbox.ColorDefault,
		appendFunc: appendString,
	}

	green := &termboxSprintfFuncer{
		fg:         termbox.ColorGreen,
		bg:         termbox.ColorDefault,
		appendFunc: appendString,
	}

	blackBold := &termboxSprintfFuncer{
		fg:         termbox.ColorBlack | termbox.AttrBold,
		bg:         termbox.ColorDefault,
		appendFunc: appendString,
	}

	formatter.SpaceColor = regular
	formatter.CommaColor = bold
	formatter.ColonColor = bold
	formatter.ObjectColor = bold
	formatter.ArrayColor = bold
	formatter.FieldQuoteColor = blueBold
	formatter.FieldColor = blueBold
	formatter.StringQuoteColor = green
	formatter.StringColor = green
	formatter.TrueColor = regular
	formatter.FalseColor = regular
	formatter.NumberColor = regular
	formatter.NullColor = blackBold

	err := formatter.Format(ioutil.Discard, []byte(strings.Join(rows, "\n")))
	if err != nil {
		return nil, err
	}

	return cells, nil
}

func (t *Terminal) drawCells(x int, y int, cells []termbox.Cell) {
	for i, c := range cells {
		termbox.SetCell(x+i, y, c.Ch, c.Fg, c.Bg)
	}
}

func (t *Terminal) drawln(x int, y int, str string, matches [][]int) {
	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault

	var c termbox.Attribute
	for i, s := range str {
		c = color
		for _, match := range matches {
			if i >= match[0]+1 && i < match[1] {
				c = termbox.ColorGreen
			}
		}
		termbox.SetCell(x+i, y, s, c, backgroundColor)
	}
}

func (t *Terminal) drawCandidates(x int, y int, index int, candidates []string) int {
	color := termbox.ColorBlack
	backgroundColor := termbox.ColorWhite

	w, _ := termbox.Size()

	ss := candidates[index]
	re := regexp.MustCompile("[[:space:]]" + ss + "[[:space:]]")

	var rows []string
	var str string
	for _, word := range candidates {
		combine := " "
		if l := len(str); l+len(word)+1 >= w {
			rows = append(rows, str+" ")
			str = ""
		}
		str += combine + word
	}
	rows = append(rows, str+" ")

	for i, row := range rows {
		match := re.FindStringIndex(row)
		var c termbox.Attribute
		for ii, s := range row {
			c = color
			backgroundColor = termbox.ColorMagenta
			if match != nil && ii >= match[0]+1 && ii < match[1]-1 {
				backgroundColor = termbox.ColorWhite
			}
			termbox.SetCell(x+ii, y+i, s, c, backgroundColor)
		}
	}
	return y + len(rows)
}
