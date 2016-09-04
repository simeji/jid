package jig

import (
	"github.com/bitly/go-simplejson"
	"regexp"
	"sort"
	"strings"
)

type SuggestionInterface interface {
	Get(json *simplejson.Json, keyword string) []string
	GetCandidateKeys(json *simplejson.Json, keyword string) []string
}

type SuggestionDataType int

const (
	UNKNOWN SuggestionDataType = iota
	ARRAY
	MAP
	NUMBER
	STRING
	BOOL
)

type Suggestion struct {
}

func NewSuggestion() *Suggestion {
	return &Suggestion{}
}

func (s *Suggestion) Get(json *simplejson.Json, keyword string) []string {
	var result string
	var suggestion string

	candidateKeys := s.GetCandidateKeys(json, keyword)

	if a, err := json.Array(); err == nil {
		if len(a) > 1 {
			rep := ""
			if len(keyword) > 0 {
				rep = keyword[0:1]
			}
			return []string{strings.Replace(`[`, rep, "", -1), `[`}
		}
		return []string{strings.Replace(`[0]`, keyword, "", -1), `[0]`}
	}
	if keyword == "" {
		if _, err := json.Map(); err == nil {
			return []string{".", "."}
		}
	}

	for _, key := range candidateKeys {
		if suggestion == "" {
			suggestion = key
		} else {
			axis := suggestion
			if len(suggestion) > len(key) {
				axis = key
			}
			max := 0
			for i, _ := range axis {
				if suggestion[i] == key[i] {
					max = i
				}
			}
			suggestion = suggestion[0 : max+1]
		}
	}
	if reg, err := regexp.Compile("(?i)^" + keyword); err == nil {
		result = reg.ReplaceAllString(suggestion, "")
	}
	return []string{result, suggestion}
}

func (s *Suggestion) GetCandidateKeys(json *simplejson.Json, keyword string) []string {
	var candidates []string

	if _, err := json.Array(); err == nil {
		return []string{}
	}

	if keyword == "" {
		return getCurrentKeys(json)
	}

	reg, err := regexp.Compile("(?i)^" + keyword)
	if err != nil {
		return []string{}
	}
	for _, key := range getCurrentKeys(json) {
		if reg.MatchString(key) {
			candidates = append(candidates, key)
		}
	}
	return candidates
}

func getCurrentKeys(json *simplejson.Json) []string {

	keys := []string{}
	m, err := json.Map()

	if err != nil {
		return keys
	}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *Suggestion) GetCurrentType(json *simplejson.Json) SuggestionDataType {
	if _, err := json.Array(); err == nil {
		return ARRAY
	} else if _, err = json.Map(); err == nil {
		return MAP
	} else if _, err = json.String(); err == nil {
		return STRING
	}
	return UNKNOWN
}
