package jig

import (
	"github.com/nsf/termbox-go"
	"regexp"
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

func (t *Terminal) draw(attr *TerminalDrawAttributes) {

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

	for idx, row := range rows {
		if i := idx - contentOffsetY; i >= 0 {
			t.drawln(0, i+y, row, nil)
		}
	}
	termbox.SetCursor(len(t.prompt)+attr.CursorOffsetX, 0)

	termbox.Flush()
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
