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
	GetFunctionCandidates(prefix string) []string
	GetFunctionSuggestion(prefix string) []string
}

// jmespathFunctions is the full list of JMESPath built-in functions.
// Each entry is the function name without the trailing "(" so it can be
// used both for completion and for the candidate list.
var jmespathFunctions = []string{
	"abs", "avg", "ceil", "contains", "ends_with", "floor",
	"join", "keys", "length", "map", "max", "max_by",
	"merge", "min", "min_by", "not_null", "reverse",
	"sort", "sort_by", "starts_with", "sum",
	"to_array", "to_number", "to_string", "type", "values",
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

// GetFunctionCandidates returns JMESPath function names (with trailing "(")
// that match the given prefix (case-insensitive).
func (s *Suggestion) GetFunctionCandidates(prefix string) []string {
	if prefix == "" {
		out := make([]string, len(jmespathFunctions))
		for i, fn := range jmespathFunctions {
			out[i] = fn + "("
		}
		return out
	}
	reg, err := regexp.Compile(`(?i)^` + regexp.QuoteMeta(prefix))
	if err != nil {
		return []string{}
	}
	var out []string
	for _, fn := range jmespathFunctions {
		if reg.MatchString(fn) {
			out = append(out, fn+"(")
		}
	}
	return out
}

// GetFunctionSuggestion returns [completion, suggestion] for function-name
// autocompletion, mirroring the return format of Get().
func (s *Suggestion) GetFunctionSuggestion(prefix string) []string {
	candidates := s.GetFunctionCandidates(prefix)
	if len(candidates) == 0 {
		return []string{"", ""}
	}
	// Find longest common prefix among candidates (strip trailing "(" for calc).
	suggestion := candidates[0]
	for _, c := range candidates[1:] {
		// work on names without "("
		a := strings.TrimSuffix(suggestion, "(")
		b := strings.TrimSuffix(c, "(")
		axis := a
		if len(b) < len(a) {
			axis = b
		}
		max := 0
		for i := range axis {
			if i >= len(a) || i >= len(b) || a[i] != b[i] {
				break
			}
			max = i
		}
		if max == 0 && (len(axis) == 0 || a[0] != b[0]) {
			suggestion = ""
			break
		}
		suggestion = a[:max+1] + "("
	}
	if suggestion == "" {
		return []string{"", ""}
	}
	// completion is the remaining characters after what has been typed
	reg, err := regexp.Compile(`(?i)^` + regexp.QuoteMeta(prefix))
	if err != nil {
		return []string{"", ""}
	}
	completion := reg.ReplaceAllString(suggestion, "")
	return []string{completion, suggestion}
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
