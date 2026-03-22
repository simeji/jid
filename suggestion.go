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

// jmespathFunctionDescriptions maps function names to brief usage descriptions.
var jmespathFunctionDescriptions = map[string]string{
	"abs":        "abs(@) — absolute value of a number",
	"avg":        "avg(@) — average of an array of numbers",
	"ceil":       "ceil(@) — ceiling of a number",
	"contains":   "contains(@, value) — true if subject contains value",
	"ends_with":  "ends_with(@, suffix) — true if string ends with suffix",
	"floor":      "floor(@) — floor of a number",
	"join":       "join(glue, @) — join array of strings with glue",
	"keys":       "keys(@) — array of keys of an object",
	"length":     "length(@) — length of string, array, or object",
	"map":        "map(&expr, @) — project expr over each element",
	"max":        "max(@) — maximum value in an array",
	"max_by":     "max_by(@, &field) — element with maximum field value",
	"merge":      "merge(@, obj2) — merge two objects",
	"min":        "min(@) — minimum value in an array",
	"min_by":     "min_by(@, &field) — element with minimum field value",
	"not_null":   "not_null(a, b, ...) — first non-null argument",
	"reverse":    "reverse(@) — reverse a string or array",
	"sort":       "sort(@) — sort an array of strings/numbers",
	"sort_by":    "sort_by(@, &field) — sort array of objects by field",
	"starts_with": "starts_with(@, prefix) — true if string starts with prefix",
	"sum":        "sum(@) — sum of an array of numbers",
	"to_array":   "to_array(@) — wrap non-array value in an array",
	"to_number":  "to_number(@) — convert to number",
	"to_string":  "to_string(@) — convert to JSON string",
	"type":       "type(@) — type name: number, string, boolean, array, object, null",
	"values":     "values(@) — array of values of an object",
}

// FunctionDescription returns the usage description for a JMESPath function name.
// name may include a trailing "(". Returns "" if not found.
func FunctionDescription(name string) string {
	key := strings.TrimSuffix(name, "(")
	return jmespathFunctionDescriptions[key]
}

// jmespathFuncTemplates maps function names to their argument template, cursor-back offset,
// and placeholder length. args is the content inside the parentheses. cursorBack is the
// number of runes from the end of "(args)" to place the cursor (0 = after ")").
// placeholderLen is the number of runes starting at the cursor position that form the
// placeholder text (0 = no placeholder).
var jmespathFuncTemplates = map[string]struct {
	args           string
	cursorBack     int
	placeholderLen int
}{
	"abs":         {"@", 0, 0},
	"avg":         {"@", 0, 0},
	"ceil":        {"@", 0, 0},
	"contains":    {"@, ''", 2, 0},        // cursor inside ''
	"ends_with":   {"@, ''", 2, 0},        // cursor inside ''
	"floor":       {"@", 0, 0},
	"join":        {"'', @", 5, 0},        // cursor inside '' (string separator)
	"keys":        {"@", 0, 0},
	"length":      {"@", 0, 0},
	"map":         {"&expr, @", 8, 4},    // placeholder "expr" (cursor after "&")
	"max":         {"@", 0, 0},
	"max_by":      {"@, &field", 6, 5},   // placeholder "field" (cursor after "&")
	"merge":       {"@, obj2", 5, 4},     // placeholder "obj2"
	"min":         {"@", 0, 0},
	"min_by":      {"@, &field", 6, 5},
	"not_null":    {"a, b", 5, 1},        // placeholder "a"
	"reverse":     {"@", 0, 0},
	"sort":        {"@", 0, 0},
	"sort_by":     {"@, &field", 6, 5},
	"starts_with": {"@, ''", 2, 0},        // cursor inside ''
	"sum":         {"@", 0, 0},
	"to_array":    {"@", 0, 0},
	"to_number":   {"@", 0, 0},
	"to_string":   {"@", 0, 0},
	"type":        {"@", 0, 0},
	"values":      {"@", 0, 0},
}

// FunctionTemplate returns the argument template, cursor-back offset, and placeholder length
// for a JMESPath function. name may include a trailing "(".
func FunctionTemplate(name string) (args string, cursorBack int, placeholderLen int) {
	key := strings.TrimSuffix(name, "(")
	if t, ok := jmespathFuncTemplates[key]; ok {
		return t.args, t.cursorBack, t.placeholderLen
	}
	return "@", 0, 0
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

// jmespathFunctionsByType maps the primary input type to the subset of functions
// that accept that type as @. Types not in the map (UNKNOWN) show all functions.
var jmespathFunctionsByType = map[SuggestionDataType][]string{
	ARRAY: {
		"avg", "contains", "join", "length", "map",
		"max", "max_by", "min", "min_by", "not_null",
		"reverse", "sort", "sort_by", "sum",
		"to_array", "to_string", "type",
	},
	MAP: {
		"keys", "length", "merge", "not_null",
		"to_array", "to_string", "type", "values",
	},
	STRING: {
		"contains", "ends_with", "length", "not_null",
		"reverse", "starts_with",
		"to_array", "to_number", "to_string", "type",
	},
	NUMBER: {
		"abs", "ceil", "floor", "not_null",
		"to_array", "to_string", "type",
	},
	BOOL: {
		"not_null", "to_array", "to_string", "type",
	},
}

// GetFunctionCandidates returns JMESPath function names (with trailing "(")
// that match the given prefix (case-insensitive).
func (s *Suggestion) GetFunctionCandidates(prefix string) []string {
	return s.GetFunctionCandidatesFiltered(prefix, UNKNOWN)
}

// GetFunctionCandidatesFiltered returns function candidates filtered by the
// given data type. UNKNOWN means all functions are shown.
func (s *Suggestion) GetFunctionCandidatesFiltered(prefix string, t SuggestionDataType) []string {
	allowed, hasFilter := jmespathFunctionsByType[t]
	var allowedSet map[string]bool
	if hasFilter {
		allowedSet = make(map[string]bool, len(allowed))
		for _, fn := range allowed {
			allowedSet[fn] = true
		}
	}

	if prefix == "" {
		var out []string
		for _, fn := range jmespathFunctions {
			if !hasFilter || allowedSet[fn] {
				out = append(out, fn+"(")
			}
		}
		return out
	}
	reg, err := regexp.Compile(`(?i)^` + regexp.QuoteMeta(prefix))
	if err != nil {
		return []string{}
	}
	var out []string
	for _, fn := range jmespathFunctions {
		if (!hasFilter || allowedSet[fn]) && reg.MatchString(fn) {
			out = append(out, fn+"(")
		}
	}
	return out
}

// GetFunctionSuggestion returns [completion, suggestion] for function-name
// autocompletion, mirroring the return format of Get().
func (s *Suggestion) GetFunctionSuggestion(prefix string) []string {
	return s.GetFunctionSuggestionFiltered(prefix, UNKNOWN)
}

// GetFunctionSuggestionFiltered returns [completion, suggestion] filtered by
// the given data type.
func (s *Suggestion) GetFunctionSuggestionFiltered(prefix string, t SuggestionDataType) []string {
	candidates := s.GetFunctionCandidatesFiltered(prefix, t)
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
