package jid

import (
	"regexp"
	"strings"

	runewidth "github.com/mattn/go-runewidth"
)

type QueryInterface interface {
	Get() []rune
	Set(query []rune) []rune
	Insert(query []rune, idx int) []rune
	Add(query []rune) []rune
	Delete(i int) []rune
	Clear() []rune
	Length() int
	IndexOffset(int) int
	GetChar(int) rune
	GetKeywords() [][]rune
	GetLastKeyword() []rune
	PopKeyword() ([]rune, []rune)
	StringGet() string
	StringSet(query string) string
	StringInsert(query string, idx int) string
	StringAdd(query string) string
	StringGetKeywords() []string
	StringGetLastKeyword() string
	StringPopKeyword() (string, []rune)
}

type Query struct {
	query    *[]rune
	complete *[]rune
}

func NewQuery(query []rune) *Query {
	q := &Query{
		query:    &[]rune{},
		complete: &[]rune{},
	}
	_ = q.Set(query)
	return q
}
func NewQueryWithString(query string) *Query {
	return NewQuery([]rune(query))
}

func (q *Query) Get() []rune {
	return *q.query
}

func (q *Query) GetChar(idx int) rune {
	var r rune = 0
	qq := q.Get()
	if l := len(qq); l > idx && idx >= 0 {
		r = qq[idx]
	}
	return r
}

func (q *Query) Length() int {
	return len(q.Get())
}

func (q *Query) IndexOffset(i int) int {
	o := 0
	if l := q.Length(); i >= l {
		o = runewidth.StringWidth(q.StringGet())
	} else if i >= 0 && i < l {
		o = runewidth.StringWidth(string(q.Get()[:i]))
	}
	return o
}

func (q *Query) Set(query []rune) []rune {
	if validate(query) {
		q.query = &query
	}
	return q.Get()
}

func (q *Query) Insert(query []rune, idx int) []rune {
	qq := q.Get()
	if idx == 0 {
		qq = append(query, qq...)
	} else if idx > 0 && len(qq) >= idx {
		_q := make([]rune, idx+len(query)-1)
		copy(_q, qq[:idx])
		qq = append(append(_q, query...), qq[idx:]...)
	}
	return q.Set(qq)
}

func (q *Query) StringInsert(query string, idx int) string {
	return string(q.Insert([]rune(query), idx))
}

func (q *Query) Add(query []rune) []rune {
	return q.Set(append(q.Get(), query...))
}

func (q *Query) Delete(i int) []rune {
	var d []rune
	qq := q.Get()
	lastIdx := len(qq)
	if i < 0 {
		if lastIdx+i >= 0 {
			d = qq[lastIdx+i:]
			qq = qq[0 : lastIdx+i]
		} else {
			d = qq
			qq = qq[0:0]
		}
	} else if i == 0 {
		d = []rune{}
		qq = qq[1:]
	} else if i > 0 && i < lastIdx {
		d = []rune{qq[i]}
		qq = append(qq[:i], qq[i+1:]...)
	}
	_ = q.Set(qq)
	return d
}

func (q *Query) Clear() []rune {
	return q.Set([]rune(""))
}

func (q *Query) GetKeywords() [][]rune {
	qq := *q.query

	if qq == nil || string(qq) == "" {
		return [][]rune{}
	}

	splitQuery := []string{}
	rr := []rune{}
	enclosed := true
	ql := len(*q.query)
	for i := 0; i < ql; i++ {
		r := qq[i]
		if ii := i + 1; r == '\\' && ql > ii && qq[ii] == '"' {
			enclosed = !enclosed
			i++ // skip '"(double quortation)'
			continue
		}
		if enclosed && r == '.' {
			splitQuery = append(splitQuery, string(rr))
			rr = []rune{}
		} else {
			rr = append(rr, r)
		}
	}
	if rr != nil {
		v := []string{string(rr)}
		if !enclosed {
			v = strings.Split(string(rr), ".")
		}
		splitQuery = append(splitQuery, v...)
	}
	lastIdx := len(splitQuery) - 1

	keywords := [][]rune{}
	for i, keyword := range splitQuery {
		if keyword != "" || i == lastIdx {
			re := regexp.MustCompile(`\[[0-9]*\]?`)
			matchIndexes := re.FindAllStringIndex(keyword, -1)
			if len(matchIndexes) < 1 {
				keywords = append(keywords, []rune(keyword))
			} else {
				if matchIndexes[0][0] > 0 {
					keywords = append(keywords, []rune(keyword[0:matchIndexes[0][0]]))
				}
				for _, matchIndex := range matchIndexes {
					k := keyword[matchIndex[0]:matchIndex[1]]
					keywords = append(keywords, []rune(k))
				}
			}
		}
	}
	return keywords
}

func (q *Query) GetLastKeyword() []rune {
	keywords := q.GetKeywords()
	if l := len(keywords); l > 0 {
		return keywords[l-1]
	}
	return []rune("")
}

func (q *Query) StringGetLastKeyword() string {
	return string(q.GetLastKeyword())
}

func (q *Query) PopKeyword() ([]rune, []rune) {
	keyword := q.GetLastKeyword()
	nq := string(keyword)
	qq := q.StringGet()

	for _, r := range keyword {
		if r == '.' {
			nq = `\"` + string(keyword) + `\"`
			break
		}
	}
	re := regexp.MustCompile(`(\.)?(\\")?` + regexp.QuoteMeta(nq) + "$")

	qq = re.ReplaceAllString(qq, "")

	query := q.Set([]rune(qq))
	return keyword, query
}

func (q *Query) StringGet() string {
	return string(q.Get())
}

func (q *Query) StringSet(query string) string {
	return string(q.Set([]rune(query)))
}

func (q *Query) StringAdd(query string) string {
	return string(q.Add([]rune(query)))
}

func (q *Query) StringGetKeywords() []string {
	var keywords []string
	for _, keyword := range q.GetKeywords() {
		keywords = append(keywords, string(keyword))
	}
	return keywords
}

func (q *Query) StringPopKeyword() (string, []rune) {
	keyword, query := q.PopKeyword()
	return string(keyword), query
}

func validate(r []rune) bool {
	s := string(r)
	if s == "" {
		return true
	}
	if regexp.MustCompile(`^[^.]`).MatchString(s) {
		return false
	}
	if regexp.MustCompile(`\.{2,}`).MatchString(s) {
		return false
	}
	if regexp.MustCompile(`\[[0-9]*\][^\.\[]`).MatchString(s) {
		return false
	}
	if regexp.MustCompile(`\[{2,}|\]{2,}`).MatchString(s) {
		return false
	}
	if regexp.MustCompile(`.\.\[`).MatchString(s) {
		return false
	}
	return true
}
