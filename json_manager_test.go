package jig

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
	result, suggest, err := jm.Get(q)

	assert.Nil(err)
	assert.Equal(`"go"`, result)
	assert.Equal([]string{``, ``}, suggest)

	// data
	data := `{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}}`
	r = bytes.NewBufferString(data)
	jm, _ = NewJsonManager(r)

	// case 2
	q = NewQueryWithString(".abcde")
	result, suggest, err = jm.Get(q)
	assert.Nil(err)
	//assert.Equal(`"2AA2"`, result)
	assert.Equal(`{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}}`, result)
	assert.Equal([]string{``, "abcde"}, suggest)

	// case 3
	q = NewQueryWithString(".abcde_fgh")
	result, suggest, err = jm.Get(q)
	assert.Nil(err)
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, result)

	// case 4
	q = NewQueryWithString(".abcde_fgh.aaa[2]")
	result, suggest, err = jm.Get(q)
	assert.Equal(`[1,2]`, result)

	// case 5
	q = NewQueryWithString(".abcde_fgh.aaa[3]")
	result, suggest, err = jm.Get(q)
	assert.Nil(err)
	assert.Equal(`null`, result)

	// case 6
	q = NewQueryWithString(".abcde_fgh.aa")
	result, suggest, err = jm.Get(q)
	assert.Nil(err)
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, result)
	assert.Equal([]string{`a`, `aaa`}, suggest)

	// case 7
	q = NewQueryWithString(".abcde_fgh.ac")
	result, suggest, err = jm.Get(q)
	assert.Nil(err)
	assert.Equal(`null`, result)
	assert.Equal([]string{``, ``}, suggest)
}

func TestGetPretty(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	jm, _ := NewJsonManager(r)
	q := NewQueryWithString(".name")
	result, _, err := jm.GetPretty(q)

	assert.Nil(err)
	assert.Equal(`"go"`, result)
}

func TestGetItem(t *testing.T) {
	var assert = assert.New(t)

	rr := bytes.NewBufferString("{\"name\":\"go\"}")
	buf, _ := ioutil.ReadAll(rr)
	sj, _ := simplejson.NewJson(buf)

	d, _ := getItem(sj, "name")
	result, _ := d.Encode()
	assert.Equal(string(result), "\"go\"")

	// case 2
	rr = bytes.NewBufferString(`{"name":"go","age":20}`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "age")
	result, _ = d.Encode()
	assert.Equal(string(result), "20")

	// case 3
	rr = bytes.NewBufferString(`{"data":{"name":"go","age":20}}`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "data")
	d2, _ := getItem(d, "name")
	d3, _ := getItem(d, "age")
	result2, _ := d2.Encode()
	result3, _ := d3.Encode()

	assert.Equal(string(result2), `"go"`)
	assert.Equal(string(result3), `20`)

	// case 4
	rr = bytes.NewBufferString(`{"data":[{"name":"test","age":30},{"name":"go","age":20}]}`)
	buf, _ = ioutil.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "data")
	d2, _ = getItem(d, "[1]")
	d3, _ = getItem(d2, "name")
	result, _ = d3.Encode()

	assert.Equal(string(result), `"go"`)
}

func TestGetFilteredData(t *testing.T) {
	var assert = assert.New(t)

	// data
	data := `{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// case 1
	q := NewQueryWithString(".abcde")
	result, s, err := jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Equal(`{"abcde":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}}`, string(d))
	//assert.Equal(`"2AA2"`, string(d))
	assert.Equal([]string{``, `abcde`}, s)

	// case 2
	q = NewQueryWithString(".abcde_fgh")
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, string(d))
	assert.Equal([]string{`.`, `.`}, s)

	// case 3
	q = NewQueryWithString(".abcde_fgh.aaa[2]")
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[1,2]`, string(d))
	assert.Equal([]string{``, `[`}, s)

	// case 4
	q = NewQueryWithString(".abcde_fgh.aaa[3]")
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`null`, string(d))
	assert.Equal([]string{``, `[`}, s)

	// case 5
	q = NewQueryWithString(".abcde_fgh.aaa")
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[123,"cccc",[1,2]]`, string(d))
	assert.Equal([]string{`[`, `[`}, s)

	// case 6
	q = NewQueryWithString(".abcde_fgh.aa")
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"}`, string(d))
	assert.Equal([]string{`a`, `aaa`}, s)

	// case 7
	q = NewQueryWithString(".abcde_fgh.aaa[")
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[123,"cccc",[1,2]]`, string(d))
	assert.Equal([]string{``, `[`}, s)
}

func TestGetFilteredDataWithMatchQuery(t *testing.T) {
	var assert = assert.New(t)

	data := `{"name":[1,2,3], "naming":{"account":"simeji"}, "test":"simeji", "testing":"ok"}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	q := NewQueryWithString(`.name`)
	result, s, err := jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Equal(`[1,2,3]`, string(d))
	assert.Equal([]string{"[", "["}, s)

	q = NewQueryWithString(`.naming`)
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"account":"simeji"}`, string(d))
	assert.Equal([]string{".", "."}, s)

	q = NewQueryWithString(`.test`)
	result, s, err = jm.GetFilteredData(q)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`{"name":[1,2,3],"naming":{"account":"simeji"},"test":"simeji","testing":"ok"}`, string(d))
	assert.Equal([]string{"", "test"}, s)
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
