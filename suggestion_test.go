package jid

import (
	"bytes"
	"io"
	"testing"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
)

func TestNewSuggestion(t *testing.T) {
	var assert = assert.New(t)
	assert.Equal(NewSuggestion(), &Suggestion{})
}

func TestSuggestionGet(t *testing.T) {
	var assert = assert.New(t)
	j := createJson(`{"name":"simeji-github"}`)
	s := NewSuggestion()

	j = createJson(`{"name":"simeji-github", "naming":"simeji", "nickname":"simejisimeji"}`)
	assert.Equal([]string{"m", "nam"}, s.Get(j, "na"))

	j = createJson(`{"abcde":"simeji-github", "abcdef":"simeji", "ab":"simejisimeji"}`)
	assert.Equal([]string{"", ""}, s.Get(j, ""))
	assert.Equal([]string{"b", "ab"}, s.Get(j, "a"))
	assert.Equal([]string{"de", "abcde"}, s.Get(j, "abc"))
	assert.Equal([]string{"", "abcde"}, s.Get(j, "abcde"))

	j = createJson(`["zero"]`)
	assert.Equal([]string{"[0]", "[0]"}, s.Get(j, ""))
	assert.Equal([]string{"0]", "[0]"}, s.Get(j, "["))
	assert.Equal([]string{"]", "[0]"}, s.Get(j, "[0"))

	j = createJson(`["zero", "one"]`)
	assert.Equal([]string{"[", "["}, s.Get(j, ""))
	assert.Equal([]string{"", "["}, s.Get(j, "["))
	assert.Equal([]string{"]", "[0]"}, s.Get(j, "[0"))

	j = createJson(`{"Abcabc":"simeji-github", "Abcdef":"simeji"}`)
	assert.Equal([]string{"bc", "Abc"}, s.Get(j, "a"))
	assert.Equal([]string{"c", "Abc"}, s.Get(j, "ab"))

	j = createJson(`{"RootDeviceNames":"simeji-github", "RootDeviceType":"simeji"}`)
	assert.Equal([]string{"ootDevice", "RootDevice"}, s.Get(j, "r"))
	assert.Equal([]string{"ootDevice", "RootDevice"}, s.Get(j, "R"))
}

func TestSuggestionGetCurrentType(t *testing.T) {
	var assert = assert.New(t)
	s := NewSuggestion()

	j := createJson(`[1,2,3]`)
	assert.Equal(ARRAY, s.GetCurrentType(j))
	j = createJson(`{"name":[1,2,3], "naming":{"account":"simeji"}, "test":"simeji", "testing":"ok"}`)
	assert.Equal(MAP, s.GetCurrentType(j))
	j = createJson(`"name"`)
	assert.Equal(STRING, s.GetCurrentType(j))
	j = createJson("1")
	assert.Equal(UNKNOWN, s.GetCurrentType(j))
}

func TestSuggestionGetCandidateKeys(t *testing.T) {
	var assert = assert.New(t)
	j := createJson(`{"naming":"simeji", "nickname":"simejisimeji", "city":"tokyo", "name":"simeji-github" }`)
	s := NewSuggestion()

	assert.Equal([]string{"city", "name", "naming", "nickname"}, s.GetCandidateKeys(j, ""))
	assert.Equal([]string{"name", "naming", "nickname"}, s.GetCandidateKeys(j, "n"))
	assert.Equal([]string{"name", "naming"}, s.GetCandidateKeys(j, "na"))
	assert.Equal([]string{}, s.GetCandidateKeys(j, "nana"))

	j = createJson(`{"abcde":"simeji-github", "abcdef":"simeji", "ab":"simejisimeji"}`)
	assert.Equal([]string{"abcde", "abcdef"}, s.GetCandidateKeys(j, "abcde"))

	j = createJson(`{"name":"simeji-github"}`)
	assert.Equal([]string{"name"}, s.GetCandidateKeys(j, ""))

	j = createJson(`{"n":"simeji-github"}`)
	assert.Equal([]string{"n"}, s.GetCandidateKeys(j, ""))

	j = createJson(`[1,2,"aa"]`)
	s = NewSuggestion()
	assert.Equal([]string{}, s.GetCandidateKeys(j, "["))
}
func TestSuggestionGetCandidateKeysWithDots(t *testing.T) {
	var assert = assert.New(t)
	j := createJson(`{"nam.ing":"simeji", "nickname":"simejisimeji", "city":"tokyo", "name":"simeji-github" }`)
	s := NewSuggestion()

	assert.Equal([]string{"city", `\"nam.ing\"`, "name", "nickname"}, s.GetCandidateKeys(j, ""))
	assert.Equal([]string{`\"nam.ing\"`, "name", "nickname"}, s.GetCandidateKeys(j, "n"))
}

func TestGetFunctionCandidates(t *testing.T) {
	var assert = assert.New(t)
	s := NewSuggestion()

	// empty prefix returns all built-in functions with trailing "("
	all := s.GetFunctionCandidates("")
	assert.Equal(len(jmespathFunctions), len(all))
	assert.Contains(all, "length(")
	assert.Contains(all, "sort_by(")

	// prefix filter (case-insensitive)
	assert.Equal([]string{"length("}, s.GetFunctionCandidates("le"))
	assert.Equal([]string{"sort(", "sort_by("}, s.GetFunctionCandidates("so"))
	assert.Nil(s.GetFunctionCandidates("zzz"))
}

func TestGetFunctionCandidatesFiltered(t *testing.T) {
	var assert = assert.New(t)
	s := NewSuggestion()

	// UNKNOWN shows everything
	all := s.GetFunctionCandidatesFiltered("", UNKNOWN)
	assert.Equal(len(jmespathFunctions), len(all))

	// ARRAY: array-compatible only
	arr := s.GetFunctionCandidatesFiltered("", ARRAY)
	assert.Contains(arr, "length(")
	assert.Contains(arr, "sort(")
	assert.Contains(arr, "avg(")
	assert.NotContains(arr, "abs(")   // NUMBER-only
	assert.NotContains(arr, "keys(")  // MAP-only

	// MAP: map-compatible only
	mp := s.GetFunctionCandidatesFiltered("", MAP)
	assert.Contains(mp, "keys(")
	assert.Contains(mp, "values(")
	assert.NotContains(mp, "avg(")    // ARRAY-only

	// STRING: string-compatible only
	str := s.GetFunctionCandidatesFiltered("", STRING)
	assert.Contains(str, "contains(")
	assert.Contains(str, "starts_with(")
	assert.NotContains(str, "sum(")   // ARRAY-only

	// NUMBER: number-compatible only
	num := s.GetFunctionCandidatesFiltered("", NUMBER)
	assert.Contains(num, "abs(")
	assert.Contains(num, "ceil(")
	assert.NotContains(num, "keys(")  // MAP-only

	// prefix + type filter
	assert.Equal([]string{"sort(", "sort_by("}, s.GetFunctionCandidatesFiltered("so", ARRAY))
	assert.Empty(s.GetFunctionCandidatesFiltered("so", MAP)) // sort/sort_by not in MAP
}

func TestGetFunctionSuggestion(t *testing.T) {
	var assert = assert.New(t)
	s := NewSuggestion()

	// unique prefix → full suggestion
	assert.Equal([]string{"ngth(", "length("}, s.GetFunctionSuggestion("le"))

	// common prefix of multiple candidates ("sort" and "sort_by" share "sort")
	assert.Equal([]string{"rt(", "sort("}, s.GetFunctionSuggestion("so"))

	// no match
	assert.Equal([]string{"", ""}, s.GetFunctionSuggestion("zzz"))
}

func TestFunctionDescription(t *testing.T) {
	var assert = assert.New(t)

	assert.NotEmpty(FunctionDescription("length("))
	assert.NotEmpty(FunctionDescription("length")) // without trailing "("
	assert.Empty(FunctionDescription("nonexistent"))
}

func TestFunctionTemplate(t *testing.T) {
	var assert = assert.New(t)

	args, cursorBack, phLen := FunctionTemplate("abs(")
	assert.Equal("@", args)
	assert.Equal(0, cursorBack)
	assert.Equal(0, phLen)

	args, cursorBack, phLen = FunctionTemplate("contains(")
	assert.Equal("@, ''", args)
	assert.Equal(2, cursorBack)
	assert.Equal(0, phLen)

	args, cursorBack, phLen = FunctionTemplate("sort_by(")
	assert.Equal("@, &field", args)
	assert.Equal(6, cursorBack)
	assert.Equal(5, phLen)

	// unknown function falls back to "@"
	args, cursorBack, phLen = FunctionTemplate("unknown(")
	assert.Equal("@", args)
	assert.Equal(0, cursorBack)
	assert.Equal(0, phLen)
}

func createJson(s string) *simplejson.Json {
	r := bytes.NewBufferString(s)
	buf, _ := io.ReadAll(r)
	j, _ := simplejson.NewJson(buf)
	return j
}
