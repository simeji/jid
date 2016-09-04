package jig

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

func (jm *JsonManager) Get(q QueryInterface) (string, []string, error) {
	json, suggestion, _ := jm.GetFilteredData(q)

	data, enc_err := json.Encode()

	if enc_err != nil {
		return "", []string{"", ""}, errors.Wrap(enc_err, "failure json encode")
	}

	return string(data), suggestion, nil
}

func (jm *JsonManager) GetPretty(q QueryInterface) (string, []string, error) {
	json, suggestion, _ := jm.GetFilteredData(q)
	s, err := json.EncodePretty()
	if err != nil {
		return "", []string{"", ""}, errors.Wrap(err, "failure json encode")
	}
	return string(s), suggestion, nil
}

func (jm *JsonManager) GetFilteredData(q QueryInterface) (*simplejson.Json, []string, error) {
	json := jm.origin

	lastKeyword := q.StringGetLastKeyword()
	keywords := q.StringGetKeywords()
	idx := 0
	if l := len(keywords); l > 0 {
		idx = l - 1
	}
	for _, keyword := range keywords[0:idx] {
		json, _ = getItem(json, keyword)
	}
	reg := regexp.MustCompile(`\[[0-9]*$`)

	suggest := jm.suggestion.Get(json, lastKeyword)
	candidateKeys := jm.suggestion.GetCandidateKeys(json, lastKeyword)
	if len(reg.FindString(lastKeyword)) < 1 {
		candidateNum := len(candidateKeys)
		if j, _ := getItem(json, lastKeyword); !isEmptyJson(j) && candidateNum == 1 {
			json = j
			suggest = jm.suggestion.Get(json, "")
		} else if candidateNum < 1 {
			json = j
		} else {
		}
	}
	return json, suggest, nil
}

func (jm *JsonManager) GetCandidateKeys(q QueryInterface) []string {
	return jm.suggestion.GetCandidateKeys(jm.current, q.StringGetLastKeyword())
}

func getItem(json *simplejson.Json, s string) (*simplejson.Json, error) {
	var result *simplejson.Json

	re := regexp.MustCompile(`\[([0-9]+)\]`)
	matches := re.FindStringSubmatch(s)

	if len(matches) > 0 {
		index, _ := strconv.Atoi(matches[1])
		result = json.GetIndex(index)
	} else {
		result = json.Get(s)
	}
	return result, nil
}

func isEmptyJson(j *simplejson.Json) bool {
	switch j.Interface().(type) {
	case nil:
		return true
	default:
		return false
	}
}
