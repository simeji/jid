package jid

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
	jmespath "github.com/jmespath/go-jmespath"
	"github.com/pkg/errors"
	"io"
)

type JsonManager struct {
	current    *simplejson.Json
	origin     *simplejson.Json
	originData interface{}
	suggestion *Suggestion
}

func NewJsonManager(reader io.Reader) (*JsonManager, error) {
	buf, err := io.ReadAll(reader)

	if err != nil {
		return nil, errors.Wrap(err, "invalid data")
	}

	j, err2 := simplejson.NewJson(buf)

	if err2 != nil {
		return nil, errors.Wrap(err2, "invalid json format")
	}

	var originData interface{}
	if err3 := json.Unmarshal(buf, &originData); err3 != nil {
		return nil, errors.Wrap(err3, "invalid json format")
	}

	jm := &JsonManager{
		origin:     j,
		current:    j,
		originData: originData,
		suggestion: NewSuggestion(),
	}

	return jm, nil
}

func (jm *JsonManager) Get(q QueryInterface, confirm bool) (string, []string, []string, error) {
	json, suggestion, candidates, _ := jm.GetFilteredData(q, confirm)

	data, enc_err := json.Encode()

	if enc_err != nil {
		return "", []string{"", ""}, []string{"", ""}, errors.Wrap(enc_err, "failure json encode")
	}

	return string(data), suggestion, candidates, nil
}

func (jm *JsonManager) GetPretty(q QueryInterface, confirm bool) (string, []string, []string, error) {
	json, suggestion, candidates, _ := jm.GetFilteredData(q, confirm)
	s, err := json.EncodePretty()
	if err != nil {
		return "", []string{"", ""}, []string{"", ""}, errors.Wrap(err, "failure json encode")
	}
	return string(s), suggestion, candidates, nil
}

// isJMESPathQuery returns true when the query contains JMESPath-specific syntax
// that goes beyond the simple dot/bracket path notation jid already supports.
// Specifically: pipe expressions, wildcard projections, filter expressions,
// function calls, multi-select hashes, or bare @ references.
func isJMESPathQuery(qs string) bool {
	inner := strings.TrimPrefix(qs, ".")
	// pipe: " | "
	if strings.Contains(inner, " | ") {
		return true
	}
	// wildcard array projection [*] or .*
	if regexp.MustCompile(`\[\*\]|\.\*`).MatchString(inner) {
		return true
	}
	// filter expression [?
	if strings.Contains(inner, "[?") {
		return true
	}
	// function call: word(
	if regexp.MustCompile(`[a-z_]+\(`).MatchString(inner) {
		return true
	}
	// multi-select hash or bare @ reference
	if strings.ContainsAny(inner, "@{}") {
		return true
	}
	return false
}

// jmespathExprFromQuery converts jid's leading-dot query string to a JMESPath expression.
// "."              -> "@"
// ".foo.bar"       -> "foo.bar"
// ". | keys(@)"    -> "keys(@)"   (root + pipe → just the function)
// ".foo | keys(@)" -> "foo | keys(@)"
func jmespathExprFromQuery(qs string) string {
	expr := strings.TrimPrefix(qs, ".")
	if expr == "" {
		return "@"
	}
	// ". | func" becomes " | func" after TrimPrefix — strip the leading " | "
	if strings.HasPrefix(expr, " | ") {
		return expr[3:]
	}
	return expr
}

// evalJMESPath evaluates a JMESPath expression against the raw JSON data and
// returns the result as a *simplejson.Json.
func (jm *JsonManager) evalJMESPath(expr string) (*simplejson.Json, error) {
	result, err := jmespath.Search(expr, jm.originData)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return simplejson.NewJson(b)
}

// baseExprBeforePipe returns the JMESPath expression that precedes the last
// ` | ` in the query. Returns "@" if the base is the root (".").
// Returns "" if there is no pipe at all.
func baseExprBeforePipe(qs string) (string, bool) {
	inner := strings.TrimPrefix(qs, ".")
	idx := strings.LastIndex(inner, " | ")
	if idx < 0 {
		return "", false
	}
	base := inner[:idx]
	if base == "" {
		return "@", true // root
	}
	return base, true
}

// pipeSuffix returns everything after the last ` | ` in the query.
func pipeSuffix(qs string) string {
	inner := strings.TrimPrefix(qs, ".")
	if idx := strings.LastIndex(inner, " | "); idx >= 0 {
		return inner[idx+3:]
	}
	return ""
}


func (jm *JsonManager) GetFilteredData(q QueryInterface, confirm bool) (*simplejson.Json, []string, []string, error) {
	qs := q.StringGet()

	if isJMESPathQuery(qs) {
		return jm.getFilteredDataJMESPath(qs, confirm)
	}

	return jm.getFilteredDataLegacy(q, confirm)
}

// isFunctionTypingMode returns true when the user appears to be mid-typing a
// JMESPath function name after a pipe: the suffix after " | " contains only
// identifier characters and no opening parenthesis yet.
func isFunctionTypingMode(qs string) bool {
	suffix := pipeSuffix(qs)
	if suffix == "" {
		return false
	}
	// If there's no "(" yet, the user hasn't completed the function call.
	return !strings.Contains(suffix, "(")
}

// getFilteredDataJMESPath handles queries that contain JMESPath-specific syntax.
func (jm *JsonManager) getFilteredDataJMESPath(qs string, confirm bool) (*simplejson.Json, []string, []string, error) {
	expr := jmespathExprFromQuery(qs)

	// If the user is mid-typing a function name after a pipe, don't evaluate
	// the (incomplete) expression. Show the base result and function suggestions.
	if isFunctionTypingMode(qs) {
		baseExpr, hasPipe := baseExprBeforePipe(qs)
		suffix := pipeSuffix(qs)
		if hasPipe {
			var baseResult *simplejson.Json
			if baseExpr == "@" {
				baseResult = jm.origin
			} else {
				var berr error
				baseResult, berr = jm.evalJMESPath(baseExpr)
				if berr != nil {
					baseResult = jm.origin
				}
			}
			fnCandidates := jm.suggestion.GetFunctionCandidates(suffix)
			fnSuggest := jm.suggestion.GetFunctionSuggestion(suffix)
			return baseResult, fnSuggest, fnCandidates, nil
		}
	}

	// Try evaluating the full expression first.
	result, err := jm.evalJMESPath(expr)
	if err == nil {
		// Full expression is valid – no function-prefix suggestions needed.
		return result, []string{"", ""}, []string{}, nil
	}

	// Expression is incomplete (parse error). Check whether the user is typing after a pipe.
	baseExpr, hasPipe := baseExprBeforePipe(qs)
	suffix := pipeSuffix(qs)

	if hasPipe {
		// Show the result of the base expression while the user types the function.
		var baseResult *simplejson.Json
		if baseExpr == "@" {
			baseResult = jm.origin
		} else {
			var berr error
			baseResult, berr = jm.evalJMESPath(baseExpr)
			if berr != nil {
				baseResult = jm.origin
			}
		}
		// Provide function-name candidates matching the typed suffix.
		fnCandidates := jm.suggestion.GetFunctionCandidates(suffix)
		fnSuggest := jm.suggestion.GetFunctionSuggestion(suffix)
		return baseResult, fnSuggest, fnCandidates, nil
	}

	// No pipe yet – the expression itself is partially typed but invalid.
	// Fall back to the origin JSON so the display isn't empty.
	return jm.origin, []string{"", ""}, []string{}, nil
}

// getFilteredDataLegacy is the original keyword-traversal logic, unchanged.
func (jm *JsonManager) getFilteredDataLegacy(q QueryInterface, confirm bool) (*simplejson.Json, []string, []string, error) {
	json := jm.origin

	lastKeyword := q.StringGetLastKeyword()
	keywords := q.StringGetKeywords()

	idx := 0
	if l := len(keywords); l == 0 {
		return json, []string{"", ""}, []string{}, nil
	} else if l > 0 {
		idx = l - 1
	}
	for _, keyword := range keywords[0:idx] {
		json, _ = getItem(json, keyword)
	}
	reg := regexp.MustCompile(`\[[0-9]*$`)

	suggest := jm.suggestion.Get(json, lastKeyword)
	candidateKeys := jm.suggestion.GetCandidateKeys(json, lastKeyword)
	// hash
	if len(reg.FindString(lastKeyword)) < 1 {
		candidateNum := len(candidateKeys)
		if j, exist := getItem(json, lastKeyword); exist && (confirm || candidateNum == 1) {
			json = j
			candidateKeys = []string{}
			if _, err := json.Array(); err == nil {
				suggest = jm.suggestion.Get(json, "")
			} else {
				suggest = []string{"", ""}
			}
		} else if candidateNum < 1 {
			json = j
			suggest = jm.suggestion.Get(json, "")
		}
	}
	return json, suggest, candidateKeys, nil
}

func (jm *JsonManager) GetCandidateKeys(q QueryInterface) []string {
	return jm.suggestion.GetCandidateKeys(jm.current, q.StringGetLastKeyword())
}

func getItem(json *simplejson.Json, s string) (*simplejson.Json, bool) {
	var result *simplejson.Json
	var exist bool

	re := regexp.MustCompile(`\[([0-9]+)\]`)
	matches := re.FindStringSubmatch(s)

	if s == "" {
		return json, false
	}

	// Query include [
	if len(matches) > 0 {
		index, _ := strconv.Atoi(matches[1])
		if a, err := json.Array(); err != nil {
			exist = false
		} else if len(a) < index {
			exist = false
		}
		result = json.GetIndex(index)
	} else {
		result, exist = json.CheckGet(s)
		if result == nil {
			result = &simplejson.Json{}
		}
	}
	return result, exist
}

func isEmptyJson(j *simplejson.Json) bool {
	switch j.Interface().(type) {
	case nil:
		return true
	default:
		return false
	}
}
