package jig

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewQuery(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("name")
	q := NewQuery(v)

	assert.Equal(*q.query, []rune("name"))
	assert.Equal(*q.complete, []rune(""))
}

func TestQueryGet(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("test")
	q := NewQuery(v)

	assert.Equal(q.Get(), []rune("test"))
}

func TestQuerySet(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("hello")
	q := NewQuery(v)

	assert.Equal(q.Set([]rune("world")), []rune("world"))
}

func TestQueryAdd(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("hello")
	q := NewQuery(v)

	assert.Equal(q.Add([]rune("world")), []rune("helloworld"))
}

func TestQueryClear(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("test")
	q := NewQuery(v)

	assert.Equal(q.Clear(), []rune(""))
}

func TestQueryDelete(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("helloworld")
	q := NewQuery(v)

	assert.Equal(q.Delete(1), []rune("helloworl"))
	assert.Equal(q.Delete(1), []rune("hellowor"))
	assert.Equal(q.Delete(2), []rune("hellow"))
	assert.Equal(q.Delete(8), []rune(""))
}

func TestGetKeywords(t *testing.T) {
	var assert = assert.New(t)

	v := []rune(".test.name")
	q := NewQuery(v)
	assert.Equal(q.GetKeywords(), [][]rune{
		[]rune("test"),
		[]rune("name"),
	})

	v = []rune("test.name")
	q = NewQuery(v)
	assert.Equal(q.GetKeywords(), [][]rune{
		[]rune("test"),
		[]rune("name"),
	})

	v = []rune("test...name")
	q = NewQuery(v)
	assert.Equal(q.GetKeywords(), [][]rune{
		[]rune("test"),
		[]rune("name"),
	})

	v = []rune("hello")
	q = NewQuery(v)
	assert.Equal(q.GetKeywords(), [][]rune{
		[]rune("hello"),
	})

	v = []rune("..hello..")
	q = NewQuery(v)
	assert.Equal(q.GetKeywords(), [][]rune{
		[]rune("hello"),
		[]rune(""),
	})
}

func TestGetLastKeyword(t *testing.T) {
	var assert = assert.New(t)

	v := []rune(".test.name")
	q := NewQuery(v)
	assert.Equal(q.GetLastKeyword(), []rune("name"))

	v = []rune(".test.")
	q = NewQuery(v)
	assert.Equal(q.GetLastKeyword(), []rune(""))

	v = []rune(".test")
	q = NewQuery(v)
	assert.Equal(q.GetLastKeyword(), []rune("test"))
}

func TestPopKeyword(t *testing.T) {
	var assert = assert.New(t)

	v := []rune(".test.name")
	q := NewQuery(v)
	k, query := q.PopKeyword()
	assert.Equal(k, []rune("name"))
	assert.Equal(query, []rune(".test."))
	assert.Equal(q.Get(), []rune(".test."))

	v = []rune("test.name.")
	q = NewQuery(v)
	k, query = q.PopKeyword()
	assert.Equal(k, []rune(""))
	assert.Equal(query, []rune("test.name."))
	assert.Equal(q.Get(), []rune("test.name."))
}

func TestQueryStringGet(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("test")
	q := NewQuery(v)

	assert.Equal(q.StringGet(), "test")
}

func TestQueryStringSet(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("hello")
	q := NewQuery(v)

	assert.Equal(q.StringSet("world"), "world")
}

func TestQueryStringAdd(t *testing.T) {
	var assert = assert.New(t)

	v := []rune("hello")
	q := NewQuery(v)

	assert.Equal(q.StringAdd("world"), "helloworld")
}

func TestStringGetKeywords(t *testing.T) {
	var assert = assert.New(t)

	v := []rune(".test.name")
	q := NewQuery(v)
	assert.Equal(q.StringGetKeywords(), []string{
		"test",
		"name",
	})

	v = []rune("test.name")
	q = NewQuery(v)
	assert.Equal(q.StringGetKeywords(), []string{
		"test",
		"name",
	})
}
