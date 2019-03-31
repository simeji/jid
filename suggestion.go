package jid

import (
	"regexp"
	"sort"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
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
	var completion string
	var suggestion string

	if a, err := json.Array(); err == nil {
		if len(a) > 1 {
			kw := regexp.MustCompile(`\[([0-9]+)?\]?`).FindString(keyword)
			if kw == "" {
				return []string{"[", "["}
			} else if kw == "[" {
				return []string{"", "["}
			}
			return []string{strings.Replace(kw+"]", kw, "", -1), kw + "]"}
		}
		return []string{strings.Replace(`[0]`, keyword, "", -1), `[0]`}
	}

	candidateKeys := s.GetCandidateKeys(json, keyword)

	if keyword == "" {
		if l := len(candidateKeys); l > 1 {
			return []string{"", ""}
		} else if l == 1 {
			return []string{candidateKeys[0], candidateKeys[0]}
		}
	}

	for _, key := range candidateKeys {
		// first
		if suggestion == "" && key != "" {
			suggestion = key
		} else {
			axis := suggestion
			if len(suggestion) > len(key) {
				axis = key
			}
			max := 0
			for i, _ := range axis {
				if suggestion[i] != key[i] {
					break
				}
				max = i
			}
			if max == 0 {
				suggestion = ""
				break
			}
			suggestion = suggestion[0 : max+1]
		}
	}
	if reg, err := regexp.Compile("(?i)^" + keyword); err == nil {
		completion = reg.ReplaceAllString(suggestion, "")
	}
	return []string{completion, suggestion}
}

func (s *Suggestion) GetCandidateKeys(json *simplejson.Json, keyword string) []string {
	candidates := []string{}

	if _, err := json.Array(); err == nil {
		return []string{}
	}

	if keyword == "" {
		return getCurrentKeys(json)
	}

	reg, err := regexp.Compile(`(?i)^(\\")?` + keyword + `(\\")?`)
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

	kk := []string{}
	m, err := json.Map()

	if err != nil {
		return kk
	}
	for k := range m {
		kk = append(kk, k)
	}
	sort.Strings(kk)

	keys := []string{}
	for _, k := range kk {
		if strings.Contains(k, ".") {
			var sb strings.Builder
			sb.Grow(len(k) + 4)
			sb.WriteString(`\"`)
			sb.WriteString(k)
			sb.WriteString(`\"`)
			k = sb.String()
		}
		keys = append(keys, k)
	}
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
