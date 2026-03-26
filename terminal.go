package jid

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	runewidth "github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
	"github.com/nwidger/jsoncolor"
)

type Terminal struct {
	defaultY   int
	prompt     string
	formatter  *jsoncolor.Formatter
	monochrome bool
	outputArea *[][]termbox.Cell
}

type TerminalDrawAttributes struct {
	Query             string
	Contents          []string
	CandidateIndex    int
	ContentsOffsetY   int
	Complete          string
	Candidates        []string
	CursorOffset      int
	FuncHelp          string
	PlaceholderStart  int // rune index in Query; -1 if no placeholder
	PlaceholderLen    int
	SelectedCandidate string // field name to highlight in JSON; "" if none
}

func NewTerminal(prompt string, defaultY int, monochrome bool) *Terminal {
	t := &Terminal{
		prompt:     prompt,
		defaultY:   defaultY,
		monochrome: monochrome,
		outputArea: &[][]termbox.Cell{},
		formatter:  nil,
	}
	if !monochrome {
		t.formatter = t.initColorizeFormatter()
	}
	return t
}

func (t *Terminal) Draw(attr *TerminalDrawAttributes) error {

	query := attr.Query
	complete := attr.Complete
	rows := attr.Contents
	candidates := attr.Candidates
	candidateidx := attr.CandidateIndex
	contentOffsetY := attr.ContentsOffsetY

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	y := t.defaultY
	_, h := termbox.Size()

	t.drawFilterLine(query, complete, attr.PlaceholderStart, attr.PlaceholderLen)

	if len(candidates) > 0 {
		y = t.drawCandidates(0, t.defaultY, candidateidx, candidates)
	}
	if attr.FuncHelp != "" {
		col := 0
		for _, ch := range attr.FuncHelp {
			termbox.SetCell(col, y, ch, termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault)
			col++
		}
		y++
	}

	cellsArr, err := t.rowsToCells(rows)
	if err != nil {
		return err
	}

	if attr.SelectedCandidate != "" {
		for i, row := range cellsArr {
			cellsArr[i] = highlightCandidateKey(row, attr.SelectedCandidate)
		}
	}

	for idx, cells := range cellsArr {
		i := idx - contentOffsetY
		if i >= 0 {
			t.drawCells(0, i+y, cells)
		}
		if i > h {
			break
		}
	}

	termbox.SetCursor(len(t.prompt)+attr.CursorOffset, 0)

	termbox.Flush()
	return nil
}

func (t *Terminal) drawFilterLine(qs string, complete string, phStart int, phLen int) error {
	fs := t.prompt + qs
	cs := complete
	str := fs + cs

	color := termbox.ColorDefault
	backgroundColor := termbox.ColorDefault

	// Compute byte offsets of placeholder range in str (prompt + query are ASCII)
	promptBytes := len(t.prompt)
	phByteStart := -1
	phByteEnd := -1
	if phStart >= 0 && phLen > 0 {
		phByteStart = promptBytes + phStart
		phByteEnd = phByteStart + phLen
	}

	var cells []termbox.Cell
	match := []int{len(fs), len(fs + cs)}

	var c termbox.Attribute
	for i, s := range str {
		c = color
		if i >= match[0] && i < match[1] {
			c = termbox.ColorGreen
		}
		if phByteStart >= 0 && i >= phByteStart && i < phByteEnd {
			c = termbox.ColorBlue
		}
		cells = append(cells, termbox.Cell{
			Ch: s,
			Fg: c,
			Bg: backgroundColor,
		})
	}
	t.drawCells(0, 0, cells)
	return nil
}

type termboxSprintfFuncer struct {
	fg         termbox.Attribute
	bg         termbox.Attribute
	outputArea *[][]termbox.Cell
}

func (tsf *termboxSprintfFuncer) SprintfFunc() func(format string, a ...interface{}) string {
	return func(format string, a ...interface{}) string {
		cells := tsf.outputArea
		idx := len(*cells) - 1
		str := fmt.Sprintf(format, a...)
		for _, s := range str {
			if s == '\n' {
				*cells = append(*cells, []termbox.Cell{})
				idx++
				continue
			}
			(*cells)[idx] = append((*cells)[idx], termbox.Cell{
				Ch: s,
				Fg: tsf.fg,
				Bg: tsf.bg,
			})
		}
		return "dummy"
	}
}

func (t *Terminal) initColorizeFormatter() *jsoncolor.Formatter {
	formatter := jsoncolor.NewFormatter()

	regular := &termboxSprintfFuncer{
		fg:         termbox.ColorDefault,
		bg:         termbox.ColorDefault,
		outputArea: t.outputArea,
	}

	bold := &termboxSprintfFuncer{
		fg:         termbox.AttrBold,
		bg:         termbox.ColorDefault,
		outputArea: t.outputArea,
	}

	blueBold := &termboxSprintfFuncer{
		fg:         termbox.ColorBlue | termbox.AttrBold,
		bg:         termbox.ColorDefault,
		outputArea: t.outputArea,
	}

	green := &termboxSprintfFuncer{
		fg:         termbox.ColorGreen,
		bg:         termbox.ColorDefault,
		outputArea: t.outputArea,
	}

	blackBold := &termboxSprintfFuncer{
		fg:         termbox.ColorBlack | termbox.AttrBold,
		bg:         termbox.ColorDefault,
		outputArea: t.outputArea,
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

	return formatter
}

func (t *Terminal) rowsToCells(rows []string) ([][]termbox.Cell, error) {
	*t.outputArea = [][]termbox.Cell{[]termbox.Cell{}}

	var err error

	if t.formatter != nil {
		err = t.formatter.Format(io.Discard, []byte(strings.Join(rows, "\n")))
	}

	cells := *t.outputArea

	if err != nil || t.monochrome {
		cells = [][]termbox.Cell{}
		for _, row := range rows {
			var cls []termbox.Cell
			for _, char := range row {
				cls = append(cls, termbox.Cell{
					Ch: char,
					Fg: termbox.ColorDefault,
					Bg: termbox.ColorDefault,
				})
			}
			cells = append(cells, cls)
		}
	}

	return cells, nil
}

func (t *Terminal) drawCells(x int, y int, cells []termbox.Cell) {
	i := 0
	for _, c := range cells {
		termbox.SetCell(x+i, y, c.Ch, c.Fg, c.Bg)

		w := runewidth.RuneWidth(c.Ch)
		if w == 0 || w == 2 && runewidth.IsAmbiguousWidth(c.Ch) {
			w = 1
		}

		i += w
	}
}

func (t *Terminal) drawFuncHelp(x int, y int, help string) {
	fg := termbox.ColorYellow
	bg := termbox.ColorDefault
	i := 0
	for _, ch := range help {
		termbox.SetCell(x+i, y, ch, fg, bg)
		w := runewidth.RuneWidth(ch)
		if w == 0 || w == 2 && runewidth.IsAmbiguousWidth(ch) {
			w = 1
		}
		i += w
	}
}

// highlightCandidateKey highlights the JSON key matching `key` in a row of cells
// by applying a yellow background. Returns the original slice if the key is not found.
func highlightCandidateKey(cells []termbox.Cell, key string) []termbox.Cell {
	var sb strings.Builder
	for _, c := range cells {
		sb.WriteRune(c.Ch)
	}
	rowStr := sb.String()
	pattern := `"` + key + `"`
	idx := strings.Index(rowStr, pattern)
	if idx < 0 {
		return cells
	}
	// Confirm it is a JSON key (followed by optional whitespace then ':')
	rest := strings.TrimLeft(rowStr[idx+len(pattern):], " \t")
	if !strings.HasPrefix(rest, ":") {
		return cells
	}
	result := make([]termbox.Cell, len(cells))
	copy(result, cells)
	runeStart := len([]rune(rowStr[:idx]))
	patternRuneLen := len([]rune(pattern))
	for i := runeStart; i < runeStart+patternRuneLen && i < len(result); i++ {
		result[i].Fg = termbox.ColorBlack
		result[i].Bg = termbox.ColorYellow
	}
	return result
}

func (t *Terminal) drawCandidates(x int, y int, index int, candidates []string) int {
	color := termbox.ColorBlack
	backgroundColor := termbox.ColorWhite

	w, _ := termbox.Size()

	ss := candidates[index]
	re := regexp.MustCompile("[[:space:]]" + regexp.QuoteMeta(ss) + "[[:space:]]")

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
		ii := 0
		for k, s := range row {
			c = color
			backgroundColor = termbox.ColorMagenta
			if match != nil && k >= match[0]+1 && k < match[1]-1 {
				backgroundColor = termbox.ColorWhite
			}
			termbox.SetCell(x+ii, y+i, s, c, backgroundColor)
			w := runewidth.RuneWidth(s)
			if w == 0 || w == 2 && runewidth.IsAmbiguousWidth(s) {
				w = 1
			}
			ii += w
		}
	}
	return y + len(rows)
}
