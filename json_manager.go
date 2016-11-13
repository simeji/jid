package jid

import (
	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	//"strings"
)

type JsonManager struct {
	current    *simplejson.Json
	origin     *simplejson.Json
	suggestion *Suggestion
}

func NewJsonManager(reader io.Reader) (*JsonManager, error) {
	buf, err := ioutil.ReadAll(reader)

	if err != nil {
		return nil, errors.Wrap(err, "invalid data")
	}

	j, err2 := simplejson.NewJson(buf)

	if err2 != nil {
		return nil, errors.Wrap(err2, "invalid json format")
	}

	json := &JsonManager{
		origin:     j,
		current:    j,
		suggestion: NewSuggestion(),
	}

	return json, nil
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

func (jm *JsonManager) GetFilteredData(q QueryInterface, confirm bool) (*simplejson.Json, []string, []string, error) {
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
