package jig

import (
	"regexp"
	"strings"
)

type QueryInterface interface {
	Get() []rune
	Set(query []rune) []rune
	Add(query []rune) []rune
	Delete(i int) []rune
	Clear() []rune
	GetKeywords() [][]rune
	GetLastKeyword() []rune
	PopKeyword() ([]rune, []rune)
	StringGet() string
	StringSet(query string) string
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

func (q *Query) Set(query []rune) []rune {
	if validate(query) {
		q.query = &query
	}
	return q.Get()
}

func (q *Query) Add(query []rune) []rune {
	return q.Set(append(q.Get(), query...))
}

func (q *Query) Delete(i int) []rune {
	qq := q.Get()
	newLastIdx := len(qq) - i
	if newLastIdx < 0 {
		newLastIdx = 0
	}
	return q.Set(qq[0:newLastIdx])
}

func (q *Query) Clear() []rune {
	return q.Set([]rune(""))
}

func (q *Query) GetKeywords() [][]rune {
	query := string(*q.query)

	if query == "" {
		return [][]rune{}
	}

	splitQuery := strings.Split(query, ".")
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
	var keyword []rune
	var lastSepIdx int
	var lastBracketIdx int
	qq := q.Get()
	for i, e := range qq {
		if e == '.' {
			lastSepIdx = i
		} else if e == '[' {
			lastBracketIdx = i
		}
	}

	if lastBracketIdx > lastSepIdx {
		lastSepIdx = lastBracketIdx
	}

	keywords := q.GetKeywords()
	if l := len(keywords); l > 0 {
		keyword = keywords[l-1]
	}
	query := q.Set(qq[0:lastSepIdx])
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
