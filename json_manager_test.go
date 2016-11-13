package jid

import (
	"bytes"
	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestNewJson(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	jm, e := NewJsonManager(r)

	rr := bytes.NewBufferString("{\"name\":\"go\"}")
	buf, _ := ioutil.ReadAll(rr)
	sj, _ := simplejson.NewJson(buf)

	assert.Equal(jm, &JsonManager{
		current:    sj,
		origin:     sj,
		suggestion: NewSuggestion(),
	})
	assert.Nil(e)

	assert.Equal("go", jm.current.Get("name").MustString())
}

func TestNewJsonWithError(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"")
	jm, e := NewJsonManager(r)

	assert.Nil(jm)
	assert.Regexp("invalid json format", e.Error())
}

func TestGet(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	jm, _ := NewJsonManager(r)
	q := NewQueryWithString(".name")
	result, suggest, candidateKeys, err := jm.Get(q, false)

	assert.Nil(err)
	assert.Equal(`"go"`, result)
	assert.Equal([]string{``, ``}, suggest)
	assert.Equal([]string{}, candidateKeys)

	// data
	data := `{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}}`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	// case 2
	q = NewQueryWithString(".abcde")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Nil(err)
	//assert.Equal(`"2AA2"`, result)
	assert.Equal(`{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}}`, result)
	assert.Equal([]string{``, "abcde"}, suggest)

	// case 3
	q = NewQueryWithString(".abcde_fgh")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Nil(err)
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, result)
	assert.Equal([]string{``, ``}, suggest)

	// case 4
	q = NewQueryWithString(".abcde_fgh.aaa[2]")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Equal(`[1,2]`, result)

	// case 5
	q = NewQueryWithString(".abcde_fgh.aaa[3]")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Nil(err)
	assert.Equal(`null`, result)

	// case 6
	q = NewQueryWithString(".abcde_fgh.aa")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Nil(err)
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, result)
	assert.Equal([]string{`a`, `aaa`}, suggest)

	// case 7
	q = NewQueryWithString(".abcde_fgh.ac")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Nil(err)
	assert.Equal(`null`, result)
	assert.Equal([]string{``, ``}, suggest)

	// data
	data = `{"abc":"2AA2","def":{"aaa":"bbb"}}`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	// case 2
	q = NewQueryWithString(".def")
	result, suggest, candidateKeys, err = jm.Get(q, false)
	assert.Nil(err)
	assert.Equal(`{"aaa":"bbb"}`, result)
	assert.Equal([]string{``, ``}, suggest)
}

func TestGetPretty(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	jm, _ := NewJsonManager(r)
	q := NewQueryWithString(".name")
	result, _, _, err := jm.GetPretty(q, true)

	assert.Nil(err)
	assert.Equal(`"go"`, result)
}

func TestGetItem(t *testing.T) {
	var assert = assert.New(t)

	rr := bytes.NewBufferString(`{"name":"go"}`)
	buf, _ := ioutil.ReadAll(rr)
	sj, _ := simplejson.NewJson(buf)

	d, _ := getItem(sj, "")
	result, _ := d.Encode()
	assert.Equal(`{"name":"go"}`, string(result))

	d, _ = getItem(sj, "name")
	result, _ = d.Encode()
	assert.Equal(`"go"`, string(result))

	// case 2
	rr = bytes.NewBufferString(`{"name":"go","age":20}`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "age")
	result, _ = d.Encode()
	assert.Equal("20", string(result))

	// case 3
	rr = bytes.NewBufferString(`{"data":{"name":"go","age":20}}`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "data")
	d2, _ := getItem(d, "name")
	d3, _ := getItem(d, "age")
	result2, _ := d2.Encode()
	result3, _ := d3.Encode()

	assert.Equal(`"go"`, string(result2))
	assert.Equal(`20`, string(result3))

	// case 4
	rr = bytes.NewBufferString(`{"data":[{"name":"test","age":30},{"name":"go","age":20}]}`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "data")
	d2, _ = getItem(d, "[1]")
	d3, _ = getItem(d2, "name")
	result, _ = d3.Encode()

	assert.Equal(`"go"`, string(result))

	// case 5
	rr = bytes.NewBufferString(`[{"name":"go","age":20}]`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "")
	result, _ = d.Encode()
	assert.Equal(`[{"age":20,"name":"go"}]`, string(result))

	// case 6
	d, _ = getItem(sj, "[0]")
	result, _ = d.Encode()
	assert.Equal(`{"age":20,"name":"go"}`, string(result))

}

func TestGetFilteredData(t *testing.T) {
	var assert = assert.New(t)

	// data
	data := `{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"},"cc":{"a":[3,4]}}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// case 1
	q := NewQueryWithString(".abcde")
	result, s, c, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Equal(`{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"},"cc":{"a":[3,4]}}`, string(d))
	//assert.Equal(`"2AA2"`, string(d))
	assert.Equal([]string{``, `abcde`}, s)
	assert.Equal([]string{"abcde", "abcde_fgh"}, c)

	// case 2
	q = NewQueryWithString(".abcde_fgh")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, string(d))
	assert.Equal([]string{``, ``}, s)
	assert.Equal([]string{}, c)

	// case 3
	q = NewQueryWithString(".abcde_fgh.aaa[2]")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[1,2]`, string(d))
	assert.Equal([]string{`[`, `[`}, s)

	// case 4
	q = NewQueryWithString(".abcde_fgh.aaa[3]")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`null`, string(d))
	assert.Equal([]string{``, ``}, s)

	// case 5
	q = NewQueryWithString(".abcde_fgh.aaa")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[123,"cccc",[1,2]]`, string(d))
	assert.Equal([]string{`[`, `[`}, s)

	// case 6
	q = NewQueryWithString(".abcde_fgh.aa")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, string(d))
	assert.Equal([]string{`a`, `aaa`}, s)

	// case 7
	q = NewQueryWithString(".abcde_fgh.aaa[")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[123,"cccc",[1,2]]`, string(d))
	assert.Equal([]string{``, `[`}, s)

	// case 8
	q = NewQueryWithString(".")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"},"cc":{"a":[3,4]}}`, string(d))
	assert.Equal([]string{``, ``}, s)

	// case 9
	q = NewQueryWithString(".cc.")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"a":[3,4]}`, string(d))
	assert.Equal([]string{`a`, `a`}, s)
	assert.Equal([]string{"a"}, c)

	// case 2-1
	data = `{"arraytest":[{"aaa":123,"aab":234},[1,2]]}`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	q = NewQueryWithString(".arraytest[0]")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"aaa":123,"aab":234}`, string(d))
	assert.Equal([]string{``, ``}, s)
	assert.Equal([]string{}, c)

	// case 3-1
	data = `{"aa":"abcde","bb":{"foo":"bar"}}`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	q = NewQueryWithString(".bb")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"foo":"bar"}`, string(d))
	assert.Equal([]string{``, ``}, s)
	assert.Equal([]string{}, c)

	// case 4-1
	data = `[{"name": "simeji"},{"name": "simeji2"}]`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	q = NewQueryWithString("")
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[{"name":"simeji"},{"name":"simeji2"}]`, string(d))
	assert.Equal([]string{``, ``}, s)
	assert.Equal([]string{}, c)

	// case 5-1
	data = `{"PrivateName":"simei", "PrivateAlias": "simeji2"}`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	q = NewQueryWithString(".Private")
	result, s, c, err = jm.GetFilteredData(q, false)
	d, _ = result.Encode()
	assert.Equal([]string{``, `Private`}, s)
	assert.Equal([]string{"PrivateAlias", "PrivateName"}, c)

}

func TestGetFilteredDataWithMatchQuery(t *testing.T) {
	var assert = assert.New(t)

	data := `{"name":[1,2,3], "naming":{"account":"simeji"}, "test":"simeji", "testing":"ok"}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	q := NewQueryWithString(`.name`)
	result, s, c, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Equal(`[1,2,3]`, string(d))
	assert.Equal([]string{"[", "["}, s)
	assert.Equal([]string{}, c)

	q = NewQueryWithString(`.naming`)
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"account":"simeji"}`, string(d))
	assert.Equal([]string{"", ""}, s)
	assert.Equal([]string{}, c)

	q = NewQueryWithString(`.naming.`)
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"account":"simeji"}`, string(d))
	assert.Equal([]string{"account", "account"}, s)
	assert.Equal([]string{"account"}, c)

	q = NewQueryWithString(`.test`)
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"name":[1,2,3],"naming":{"account":"simeji"},"test":"simeji","testing":"ok"}`, string(d))
	assert.Equal([]string{"", "test"}, s)
	assert.Equal([]string{"test", "testing"}, c)
}

func TestGetCandidateKeys(t *testing.T) {
	var assert = assert.New(t)
	data := `{"name":[1,2,3], "naming":{"account":"simeji"}, "test":"simeji", "testing":"ok"}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	q := NewQueryWithString(`.n`)

	keys := jm.GetCandidateKeys(q)
	assert.Equal([]string{"name", "naming"}, keys)

	q = NewQueryWithString(`.`)
	keys = jm.GetCandidateKeys(q)
	assert.Equal([]string{"name", "naming", "test", "testing"}, keys)

	q = NewQueryWithString(`.test`)
	keys = jm.GetCandidateKeys(q)
	assert.Equal([]string{"test", "testing"}, keys)

	q = NewQueryWithString(`.testi`)
	keys = jm.GetCandidateKeys(q)
	assert.Equal([]string{"testing"}, keys)

	q = NewQueryWithString(`.testia`)
	keys = jm.GetCandidateKeys(q)
	assert.Equal([]string{}, keys)
}

func TestGetCurrentKeys(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go","age":20,"weight":60}`)
	buf, _ := ioutil.ReadAll(r)
	sj, _ := simplejson.NewJson(buf)

	keys := getCurrentKeys(sj)
	assert.Equal([]string{"age", "name", "weight"}, keys)

	r = bytes.NewBufferString(`[2,3,"aa"]`)
	buf, _ = ioutil.ReadAll(r)
	sj, _ = simplejson.NewJson(buf)

	keys = getCurrentKeys(sj)
	assert.Equal([]string{}, keys)
}

func TestIsEmptyJson(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	buf, _ := ioutil.ReadAll(r)
	sj, _ := simplejson.NewJson(buf)

	assert.Equal(false, isEmptyJson(sj))
	assert.Equal(true, isEmptyJson(&simplejson.Json{}))
}
