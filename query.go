package jig

import (
	"strings"
)

type Query struct {
	query    *[]rune
	complete *[]rune
}

func NewQuery(query []rune) *Query {
	return &Query{
		query:    &query,
		complete: &[]rune{},
	}
}

func (q *Query) Get() []rune {
	return *q.query
}

func (q *Query) Set(query []rune) []rune {
	q.query = &query
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

	splitQuery := strings.Split(query, ".")
	lastIdx := len(splitQuery) - 1

	keywords := [][]rune{}
	for i, keyword := range splitQuery {
		if keyword != "" || i == lastIdx {
			keywords = append(keywords, []rune(keyword))
		}
	}
	return keywords
}

func (q *Query) GetLastKeyword() []rune {
	keywords := q.GetKeywords()
	return keywords[len(keywords)-1]
}

func (q *Query) PopKeyword() ([]rune, []rune) {
	lastSepIdx := 0
	for i, e := range *q.query {
		if e == '.' {
			lastSepIdx = i
		}
	}
	qq := q.Get()
	keyword := qq[lastSepIdx+1:]
	query := q.Set(qq[0 : lastSepIdx+1])
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
	query := string(*q.query)

	splitQuery := strings.Split(query, ".")
	lastIdx := len(splitQuery) - 1

	keywords := []string{}
	for i, keyword := range splitQuery {
		if keyword != "" || i == lastIdx {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}
