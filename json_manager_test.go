package jid

import (
	"bytes"
	"io"
	"testing"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
)

func TestNewJson(t *testing.T) {
	var assert = assert.New(t)

	r := bytes.NewBufferString("{\"name\":\"go\"}")
	jm, e := NewJsonManager(r)

	rr := bytes.NewBufferString("{\"name\":\"go\"}")
	buf, _ := io.ReadAll(rr)
	sj, _ := simplejson.NewJson(buf)

	assert.Equal(jm, &JsonManager{
		current:    sj,
		origin:     sj,
		originData: map[string]interface{}{"name": "go"},
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
	buf, _ := io.ReadAll(rr)
	sj, _ := simplejson.NewJson(buf)

	d, _ := getItem(sj, "")
	result, _ := d.Encode()
	assert.Equal(`{"name":"go"}`, string(result))

	d, _ = getItem(sj, "name")
	result, _ = d.Encode()
	assert.Equal(`"go"`, string(result))

	// case 2
	rr = bytes.NewBufferString(`{"name":"go","age":20}`)
	buf, _ = io.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "age")
	result, _ = d.Encode()
	assert.Equal("20", string(result))

	// case 3
	rr = bytes.NewBufferString(`{"data":{"name":"go","age":20}}`)
	buf, _ = io.ReadAll(rr)
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
	buf, _ = io.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "data")
	d2, _ = getItem(d, "[1]")
	d3, _ = getItem(d2, "name")
	result, _ = d3.Encode()

	assert.Equal(`"go"`, string(result))

	// case 5
	rr = bytes.NewBufferString(`[{"name":"go","age":20}]`)
	buf, _ = io.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "")
	result, _ = d.Encode()
	assert.Equal(`[{"age":20,"name":"go"}]`, string(result))

	// case 6
	d, _ = getItem(sj, "[0]")
	result, _ = d.Encode()
	assert.Equal(`{"age":20,"name":"go"}`, string(result))

	// case 7  key contains '.'
	rr = bytes.NewBufferString(`{"na.me":"go","age":20}`)
	buf, _ = io.ReadAll(rr)
	sj, _ = simplejson.NewJson(buf)

	d, _ = getItem(sj, "na.me")
	result, _ = d.Encode()
	assert.Equal(`"go"`, string(result))

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

func TestGetFilteredDataWithContainDots(t *testing.T) {
	var assert = assert.New(t)

	// data
	data := `{"abc.de":"2AA2","abcde_fgh":{"aaa":[123,"cccc",[1,2]],"c":"JJJJ"},"cc":{"a":[3,4]}}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// case 1
	q := NewQueryWithString(`.\"abc.de\"`)
	result, s, c, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Equal(`"2AA2"`, string(d))
	assert.Equal([]string{``, ``}, s)
	assert.Equal([]string{}, c)

	// case 2
	q = NewQueryWithString(`."abc.de"`)
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`null`, string(d))
	assert.Equal([]string{``, ``}, s)
	assert.Equal([]string{}, c)

	// case 3
	q = NewQueryWithString(`.abc.de`)
	result, s, c, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`null`, string(d))
	assert.Equal([]string{"", ""}, s)
	assert.Equal([]string{}, c)
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
	buf, _ := io.ReadAll(r)
	sj, _ := simplejson.NewJson(buf)

	keys := getCurrentKeys(sj)
	assert.Equal([]string{"age", "name", "weight"}, keys)

	r = bytes.NewBufferString(`[2,3,"aa"]`)
	buf, _ = io.ReadAll(r)
	sj, _ = simplejson.NewJson(buf)

	keys = getCurrentKeys(sj)
	assert.Equal([]string{}, keys)
}

func TestGetFilteredDataJMESPath(t *testing.T) {
	var assert = assert.New(t)

	data := `{"users":[{"name":"alice","age":30},{"name":"bob","age":25}],"count":2}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// keys() function — order is not guaranteed by JMESPath
	q := NewQueryWithString(". | keys(@)")
	result, _, _, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Contains(string(d), `"count"`)
	assert.Contains(string(d), `"users"`)

	// length() function
	q = NewQueryWithString(".users | length(@)")
	result, _, _, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`2`, string(d))

	// wildcard projection
	q = NewQueryWithString(".users[*].name")
	result, _, _, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`["alice","bob"]`, string(d))

	// partial pipe (user still typing function name) — should show base result
	q = NewQueryWithString(".users | len")
	result, _, fnCandidates, _ := jm.GetFilteredData(q, false)
	d, _ = result.Encode()
	// should show users array (base expression result)
	assert.Contains(string(d), "alice")
	// function suggestions should include length(
	assert.Contains(fnCandidates, "length(")
}

func TestGetFilteredDataJMESPathWildcard(t *testing.T) {
	var assert = assert.New(t)

	data := `{"users":[{"name":"alice","age":30},{"name":"bob","age":25}],"count":2}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// [*] projection → field candidates from first element
	q := NewQueryWithString(".users[*]")
	result, _, candidates, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Contains(string(d), "alice")
	assert.Contains(candidates, "age")
	assert.Contains(candidates, "name")

	// [*].fieldname (complete field) → array result, no field candidates
	q = NewQueryWithString(".users[*].name")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`["alice","bob"]`, string(d))
	assert.Empty(candidates)

	// trailing dot after [*] → still shows wildcard result + field candidates
	q = NewQueryWithString(".users[*].")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Contains(string(d), "alice")
	assert.Contains(candidates, "age")
	assert.Contains(candidates, "name")

	// partial field after [*] matching some fields → field candidates filtered
	q = NewQueryWithString(".users[*].n")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Contains(string(d), "alice") // shows base wildcard result
	assert.Equal([]string{"name"}, candidates)

	// partial field that matches nothing → fall through to actual result (not wildcard base)
	q = NewQueryWithString(".users[*].zzz")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`[]`, string(d)) // actual JMESPath result, not the wildcard base
	assert.Empty(candidates)

	// [*].field[N] → auto pipe rewrite: treat as "[*].field | [N]"
	// (JMESPath [N] within a projection applies to each element, not the array)
	q = NewQueryWithString(".users[*].name[0]")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`"alice"`, string(d)) // first projected name, not []
	assert.Empty(candidates)

	// [*].objects[N] → auto pipe rewrite → object result with field candidates
	q = NewQueryWithString(".users[*][0]")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Contains(string(d), `"alice"`) // first user object
	assert.Contains(candidates, "name")
	assert.Contains(candidates, "age")

	// [*].field[N].subfield → chained rewrite via pipe: "[*].field | [N].subfield"
	q = NewQueryWithString(".users[*].name[0]")
	result, _, _, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Equal(`"alice"`, string(d))

	// [*].field[N] | func(@) → evalBaseExpr rewrite enables pipe chaining
	q = NewQueryWithString(".users[*][0] | keys(@)")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Contains(string(d), `"age"`)
	assert.Contains(string(d), `"name"`)
	assert.Empty(candidates)

	// [*].field[N]| (pipe after index, function typing mode) → base shows rewritten result
	q = NewQueryWithString(".users[*][0] | ")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Contains(string(d), `"alice"`) // base = first user object (not [])
	assert.Contains(candidates, "keys(")  // MAP-type candidates
	assert.NotContains(candidates, "avg(") // not ARRAY-type
}

func TestGetFilteredDataJMESPathTypeFilter(t *testing.T) {
	var assert = assert.New(t)

	data := `{"items":[1,2,3],"info":{"key":"value"},"text":"hello"}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// pipe after array → ARRAY-typed function candidates only
	q := NewQueryWithString(".items | ")
	_, _, candidates, _ := jm.GetFilteredData(q, false)
	assert.Contains(candidates, "length(")
	assert.Contains(candidates, "sort(")
	assert.NotContains(candidates, "keys(")  // MAP-only
	assert.NotContains(candidates, "abs(")   // NUMBER-only

	// pipe after map → MAP-typed function candidates only
	q = NewQueryWithString(".info | ")
	_, _, candidates, _ = jm.GetFilteredData(q, false)
	assert.Contains(candidates, "keys(")
	assert.Contains(candidates, "values(")
	assert.NotContains(candidates, "avg(")  // ARRAY-only

	// pipe after string → STRING-typed function candidates only
	q = NewQueryWithString(".text | ")
	_, _, candidates, _ = jm.GetFilteredData(q, false)
	assert.Contains(candidates, "contains(")
	assert.Contains(candidates, "starts_with(")
	assert.NotContains(candidates, "sum(")  // ARRAY-only

	// pipe with partial function name → filtered suggestions
	q = NewQueryWithString(".items | so")
	result, _, candidates, _ := jm.GetFilteredData(q, false)
	d, _ := result.Encode()
	assert.Contains(string(d), "1") // shows base (items array)
	assert.Contains(candidates, "sort(")
	assert.Contains(candidates, "sort_by(")
}

func TestGetFilteredDataJMESPathObjectResult(t *testing.T) {
	var assert = assert.New(t)

	data := `{"id":1,"title":"hello","tags":["a","b"]}`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// JMESPath expression that evaluates to an object → field candidates
	q := NewQueryWithString(". | to_array(@)[0]")
	result, _, candidates, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Contains(string(d), `"hello"`)
	assert.Contains(candidates, "id")
	assert.Contains(candidates, "title")
	assert.Contains(candidates, "tags")

	// result with trailing dot after pipe expression
	q = NewQueryWithString(". | to_array(@)[0].")
	result, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	d, _ = result.Encode()
	assert.Contains(string(d), `"hello"`)
	assert.Contains(candidates, "id")
}

func TestAmpFieldPartial(t *testing.T) {
	// basic: "&field)" inside function call
	partial, ok := ampFieldPartial("sort_by(@, &field)")
	assert.True(t, ok)
	assert.Equal(t, "field", partial)

	// partial identifier, no closing paren
	partial, ok = ampFieldPartial("sort_by(@, &na")
	assert.True(t, ok)
	assert.Equal(t, "na", partial)

	// empty partial (just "&)")
	partial, ok = ampFieldPartial("max_by(@, &)")
	assert.True(t, ok)
	assert.Equal(t, "", partial)

	// empty partial (just "&" with no closing paren)
	partial, ok = ampFieldPartial("max_by(@, &")
	assert.True(t, ok)
	assert.Equal(t, "", partial)

	// no "(" → not in function call
	_, ok = ampFieldPartial("&field")
	assert.False(t, ok)

	// "&" before "(" → not after open-paren
	_, ok = ampFieldPartial("&max_by(@, field)")
	assert.False(t, ok)

	// no "&" at all
	_, ok = ampFieldPartial("sort_by(@, name)")
	assert.False(t, ok)

	// identifier with underscore and digits
	partial, ok = ampFieldPartial("sort_by(@, &base_stat2)")
	assert.True(t, ok)
	assert.Equal(t, "base_stat2", partial)
}

func TestAmpFieldCandidates(t *testing.T) {
	data := `[{"name":"alice","age":30},{"name":"bob","age":25}]`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// empty partial → all keys; no inline hint (["",""])
	_, suggest, candidates, err := jm.ampFieldCandidates("@", "")
	assert.Nil(t, err)
	assert.Contains(t, candidates, "name")
	assert.Contains(t, candidates, "age")
	assert.Equal(t, []string{"", ""}, suggest)

	// partial "na" → only "name"; still no inline hint
	_, suggest, candidates, err = jm.ampFieldCandidates("@", "na")
	assert.Nil(t, err)
	assert.Equal(t, []string{"name"}, candidates)
	assert.Equal(t, []string{"", ""}, suggest)

	// placeholder text "field" → no match → fall back to all keys; no inline hint
	_, suggest, candidates, err = jm.ampFieldCandidates("@", "field")
	assert.Nil(t, err)
	assert.Contains(t, candidates, "name")
	assert.Contains(t, candidates, "age")
	assert.Equal(t, []string{"", ""}, suggest)
}

func TestGetFilteredDataJMESPathAmpField(t *testing.T) {
	assert := assert.New(t)
	data := `[{"name":"alice","age":30},{"name":"bob","age":25}]`
	r := bytes.NewBufferString(data)
	jm, _ := NewJsonManager(r)

	// "&field)" placeholder → all field candidates from array elements
	q := NewQueryWithString(". | max_by(@, &field)")
	_, _, candidates, err := jm.GetFilteredData(q, false)
	assert.Nil(err)
	assert.Contains(candidates, "name")
	assert.Contains(candidates, "age")

	// "&na" partial (no closing paren) → "name" candidate
	q = NewQueryWithString(". | max_by(@, &na")
	_, _, candidates, err = jm.GetFilteredData(q, false)
	assert.Nil(err)
	assert.Contains(candidates, "name")

	// confirmed expression → evaluates normally, no amp-field candidates
	q = NewQueryWithString(". | max_by(@, &age)")
	result, _, _, err := jm.GetFilteredData(q, true)
	assert.Nil(err)
	d, _ := result.Encode()
	assert.Contains(string(d), "alice") // age 30 → max

	// confirm=true with a missing field → JMESPath returns an error (cannot compare
	// null values), NOT the base array that the suggestion fallback would return.
	q = NewQueryWithString(". | max_by(@, &missing)")
	_, _, _, err = jm.GetFilteredData(q, true)
	assert.NotNil(err, "confirm=true should propagate the JMESPath error, not silently return base array")
}

func TestIsEmptyJson(t *testing.T) {
	var assert = assert.New(t)
	r := bytes.NewBufferString(`{"name":"go"}`)
	buf, _ := io.ReadAll(r)
	sj, _ := simplejson.NewJson(buf)

	assert.Equal(false, isEmptyJson(sj))
	assert.Equal(true, isEmptyJson(&simplejson.Json{}))
}
