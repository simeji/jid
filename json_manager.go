package jig

import (
	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"regexp"
	"sort"
	"strconv"
)

type JsonManager struct {
	current *simplejson.Json
	origin  *simplejson.Json
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
		origin:  j,
		current: j,
	}

	return json, nil
}

func (jm *JsonManager) Get(q QueryInterface) (string, error) {
	json, _ := jm.GetFilteredData(q)

	data, enc_err := json.Encode()

	if enc_err != nil {
		return "", errors.Wrap(enc_err, "failure json encode")
	}

	return string(data), nil
}

func (jm *JsonManager) GetPretty(q QueryInterface) (string, error) {
	json, _ := jm.GetFilteredData(q)
	s, err := json.EncodePretty()
	if err != nil {
		return "", errors.Wrap(err, "failure json encode")
	}
	return string(s), nil
}

func (jm *JsonManager) GetFilteredData(q QueryInterface) (*simplejson.Json, error) {
	json := jm.origin

	lastKeyword, _ := q.StringPopKeyword()
	for _, keyword := range q.StringGetKeywords() {
		json, _ = getItem(json, keyword)
	}
	if j, _ := getItem(json, lastKeyword); !isEmptyJson(j) {
		json = j
	} else if b, _ := getCandidateKeyItem(json, lastKeyword); !b {
		json = j
	}
	return json, nil
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

func getCandidateKeyItem(json *simplejson.Json, s string) (bool, string) {
	re := regexp.MustCompile(`\[([0-9]+)\]`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 0 {
		index, _ := strconv.Atoi(matches[1])
		a, err := json.Array()
		if err != nil {
			return false, ""
		}
		if len(a)-1 < index {
			return false, ""
		}
	} else {
		reg := regexp.MustCompile("(?i)^" + s)
		var candidate string
		result := false
		for _, key := range getCurrentKeys(json) {
			if reg.MatchString(key) {
				result = true
				if candidate == "" {
					candidate = key
				} else {
					axis := candidate
					if len(candidate) > len(key) {
						axis = key
					}
					max := 0
					for i, _ := range axis {
						if candidate[i] == key[i] {
							max = i
						}
					}
					candidate = candidate[0 : max+1]
				}
			}
		}
		candidate = reg.ReplaceAllString(candidate, "")
		return result, candidate
	}
	return true, ""
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

func isEmptyJson(j *simplejson.Json) bool {
	switch j.Interface().(type) {
	case nil:
		return true
	default:
		return false
	}
}
